package handlers

import (
  "net/http"
  "time"
)

type calendarDay struct {
  Day        int
  Date       string
  IsWorkout  bool
  IsToday    bool
  IsSelected bool
}

type calendarWorkout struct {
  ID        string
  Name      string
  Date      string
  Duration  int
  Exercises int
  Calories  int
  Completed bool
}

func (a *App) Calendar(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

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
  rows, err := a.DB.Query(
    `select ws.id, w.name, date(ws.started_at), coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.total_exercises, 0), coalesce(ws.calories_burned, 0), ws.completed_at is not null
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1 and ws.started_at >= $2 and ws.started_at < $3`,
    user.ID,
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
    for _, w := range list {
      if w.Completed {
        totalWorkouts++
        totalMinutes += w.Duration
      }
    }
  }

  streak := computeWorkoutStreak(a.DB, user.ID)

  selectedWorkouts := []calendarWorkout{}
  selectedLabel := ""
  if selectedDate != "" {
    selectedWorkouts = workoutsByDate[selectedDate]
    if parsed, err := time.Parse("2006-01-02", selectedDate); err == nil {
      selectedLabel = parsed.Format("02.01.2006")
    }
  }

  recentWorkouts := []calendarWorkout{}
  recentRows, _ := a.DB.Query(
    `select ws.id, w.name, date(ws.started_at), coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.total_exercises, 0), coalesce(ws.calories_burned, 0), ws.completed_at is not null
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1
     order by ws.started_at desc
     limit 5`,
    user.ID,
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

  data := map[string]any{
    "MonthLabel":       monthLabel(currentMonth),
    "MonthParam":       currentMonth.Format("2006-01"),
    "PrevMonth":        currentMonth.AddDate(0, -1, 0).Format("2006-01"),
    "NextMonth":        currentMonth.AddDate(0, 1, 0).Format("2006-01"),
    "StartingDay":      startingDayOfWeek,
    "Days":             days,
    "SelectedDate":     selectedDate,
    "SelectedLabel":    selectedLabel,
    "SelectedWorkouts": selectedWorkouts,
    "RecentWorkouts":   recentWorkouts,
    "TotalWorkouts":    totalWorkouts,
    "CurrentStreak":    streak,
    "TotalHours":       int(float64(totalMinutes) / 60.0 + 0.5),
  }

  a.renderPage(w, r, "calendar", "Календарь", "", data)
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
