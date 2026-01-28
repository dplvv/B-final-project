package handlers

import (
  "net/http"
  "time"
)

type historyItem struct {
  ID        string
  WorkoutID string
  Name      string
  Date      time.Time
  Duration  int
  Completed bool
  Exercises int
  Calories  int
  Rating    float64
}

type historyGroup struct {
  Label string
  Items []historyItem
}

func (a *App) History(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  filter := r.URL.Query().Get("filter")
  if filter == "" {
    filter = "all"
  }

  rows, err := a.DB.Query(
    `select ws.id, ws.workout_id, w.name, ws.started_at, coalesce(ws.duration_minutes, w.duration_minutes),
            ws.completed_at is not null, coalesce(ws.total_exercises, 0), coalesce(ws.calories_burned, 0),
            coalesce((select avg(rating)::float from feedback f where f.workout_session_id = ws.id), 0)
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1
     order by ws.started_at desc`,
    user.ID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  items := []historyItem{}
  for rows.Next() {
    var item historyItem
    _ = rows.Scan(&item.ID, &item.WorkoutID, &item.Name, &item.Date, &item.Duration, &item.Completed, &item.Exercises, &item.Calories, &item.Rating)
    if filter == "completed" && !item.Completed {
      continue
    }
    if filter == "skipped" && item.Completed {
      continue
    }
    items = append(items, item)
  }

  grouped := map[string][]historyItem{}
  order := []string{}
  for _, item := range items {
    label := monthYearLabel(item.Date)
    if _, ok := grouped[label]; !ok {
      grouped[label] = []historyItem{}
      order = append(order, label)
    }
    grouped[label] = append(grouped[label], item)
  }

  groups := []historyGroup{}
  for _, label := range order {
    groups = append(groups, historyGroup{Label: label, Items: grouped[label]})
  }

  completedCount := 0
  totalDuration := 0
  totalCalories := 0
  ratingSum := 0.0
  ratingCount := 0
  for _, item := range items {
    if item.Completed {
      completedCount++
      totalDuration += item.Duration
      totalCalories += item.Calories
      if item.Rating > 0 {
        ratingSum += item.Rating
        ratingCount++
      }
    }
  }

  avgRating := 0.0
  if ratingCount > 0 {
    avgRating = ratingSum / float64(ratingCount)
  }

  data := map[string]any{
    "Filter":         filter,
    "Groups":         groups,
    "CompletedCount": completedCount,
    "TotalDuration":  totalDuration,
    "TotalCalories":  totalCalories,
    "AverageRating":  avgRating,
    "TotalCount":     len(items),
  }

  a.renderPage(w, r, "history", "История тренировок", "", data)
}

func monthYearLabel(date time.Time) string {
  months := []string{"январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"}
  month := int(date.Month()) - 1
  if month < 0 || month >= len(months) {
    return date.Format("01.2006")
  }
  return months[month] + " " + date.Format("2006")
}
