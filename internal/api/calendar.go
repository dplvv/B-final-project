package api

import (
  "net/http"
  "time"
)

type calendarDay struct {
  Day        int    `json:"day"`
  Date       string `json:"date"`
  IsWorkout  bool   `json:"is_workout"`
  IsToday    bool   `json:"is_today"`
  IsSelected bool   `json:"is_selected"`
}

type calendarWorkout struct {
  ID        string `json:"id"`
  Name      string `json:"name"`
  Date      string `json:"date"`
  Duration  int    `json:"duration"`
  Exercises int    `json:"exercises"`
  Calories  int    `json:"calories"`
  Completed bool   `json:"completed"`
}

func (api *API) Calendar(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  now := time.Now()
  monthParam := r.URL.Query().Get("month")
  currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
  if monthParam != "" {
    if parsed, err := time.Parse("2006-01", monthParam); err == nil {
      currentMonth = time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.Local)
    }
  }

  selectedDate := r.URL.Query().Get("date")

  year := currentMonth.Year()
  month := currentMonth.Month()
  firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
  lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)
  daysInMonth := lastDay.Day()
  startingDayOfWeek := int(firstDay.Weekday())

  workoutsByDate := map[string][]calendarWorkout{}
  nextMonth := firstDay.AddDate(0, 1, 0)
  rows, err := api.DB.Query(
    `select ws.id, w.name, date(ws.started_at), coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.total_exercises, 0), coalesce(ws.calories_burned, 0), ws.completed_at is not null
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1 and ws.started_at >= $2 and ws.started_at < $3`,
    userID,
    firstDay,
    nextMonth,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var workout calendarWorkout
      var date time.Time
      _ = rows.Scan(&workout.ID, &workout.Name, &date, &workout.Duration, &workout.Exercises, &workout.Calories, &workout.Completed)
      workout.Date = date.Format("2006-01-02")
      workoutsByDate[workout.Date] = append(workoutsByDate[workout.Date], workout)
    }
  }

  days := []calendarDay{}
  for i := 1; i <= daysInMonth; i++ {
    date := time.Date(year, month, i, 0, 0, 0, 0, time.Local)
    dateStr := date.Format("2006-01-02")
    _, hasWorkout := workoutsByDate[dateStr]
    day := calendarDay{
      Day:        i,
      Date:       dateStr,
      IsWorkout:  hasWorkout,
      IsToday:    dateStr == now.Format("2006-01-02"),
      IsSelected: selectedDate == dateStr,
    }
    days = append(days, day)
  }

  totalWorkouts := 0
  totalMinutes := 0
  for _, list := range workoutsByDate {
    for _, witem := range list {
      if witem.Completed {
        totalWorkouts++
        totalMinutes += witem.Duration
      }
    }
  }

  streak := computeWorkoutStreak(api.DB, userID)

  selectedWorkouts := []calendarWorkout{}
  selectedLabel := ""
  if selectedDate != "" {
    selectedWorkouts = workoutsByDate[selectedDate]
    if parsed, err := time.Parse("2006-01-02", selectedDate); err == nil {
      selectedLabel = parsed.Format("02.01.2006")
    }
  }

  recentWorkouts := []calendarWorkout{}
  recentRows, _ := api.DB.Query(
    `select ws.id, w.name, date(ws.started_at), coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.total_exercises, 0), coalesce(ws.calories_burned, 0), ws.completed_at is not null
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1
     order by ws.started_at desc
     limit 5`,
    userID,
  )
  if recentRows != nil {
    defer recentRows.Close()
    for recentRows.Next() {
      var workout calendarWorkout
      var date time.Time
      _ = recentRows.Scan(&workout.ID, &workout.Name, &date, &workout.Duration, &workout.Exercises, &workout.Calories, &workout.Completed)
      workout.Date = date.Format("2006-01-02")
      recentWorkouts = append(recentWorkouts, workout)
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "month_label": monthLabel(currentMonth),
    "month_param": currentMonth.Format("2006-01"),
    "prev_month": currentMonth.AddDate(0, -1, 0).Format("2006-01"),
    "next_month": currentMonth.AddDate(0, 1, 0).Format("2006-01"),
    "starting_day": startingDayOfWeek,
    "days": days,
    "selected_date": selectedDate,
    "selected_label": selectedLabel,
    "selected_workouts": selectedWorkouts,
    "recent_workouts": recentWorkouts,
    "total_workouts": totalWorkouts,
    "current_streak": streak,
    "total_hours": int(float64(totalMinutes)/60.0 + 0.5),
  })
}

func monthLabel(date time.Time) string {
  months := []string{
    "январь", "февраль", "март", "апрель", "май", "июнь",
    "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь",
  }
  month := int(date.Month()) - 1
  if month < 0 || month >= len(months) {
    return date.Format("01.2006")
  }
  return months[month] + " " + date.Format("2006")
}

