package handlers

import (
  "net/http"
  "time"

  "rehab-app/internal/models"
)

type weeklyDay struct {
  Day       string
  Completed bool
  Duration  int
}

type monthlyStat struct {
  Month    string
  Workouts int
  Hours    float64
}

func (a *App) Progress(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var totalWorkouts int
  var totalMinutes int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where user_id = $1 and completed_at is not null`,
    user.ID,
  ).Scan(&totalWorkouts, &totalMinutes)

  streak := computeWorkoutStreak(a.DB, user.ID)

  weekly := buildWeeklyData(a, user.ID)

  monthly := buildMonthlyStats(a, user.ID)

  achievements := []models.Achievement{}
  rows, _ := a.DB.Query(
    `select a.id, a.title, a.description, a.icon, coalesce(ua.unlocked, false), ua.unlocked_at,
            coalesce(ua.progress, 0), coalesce(ua.total, 0)
     from achievements a
     left join user_achievements ua on ua.achievement_id = a.id and ua.user_id = $1
     order by a.created_at`,
    user.ID,
  )
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var ach models.Achievement
      _ = rows.Scan(&ach.ID, &ach.Title, &ach.Description, &ach.Icon, &ach.Unlocked, &ach.UnlockedAt, &ach.Progress, &ach.Total)
      achievements = append(achievements, ach)
    }
  }

  completedAchievements := 0
  for _, ach := range achievements {
    if ach.Unlocked {
      completedAchievements++
    }
  }

  completedThisWeek := 0
  for _, day := range weekly {
    if day.Completed {
      completedThisWeek++
    }
  }

  data := map[string]any{
    "TotalWorkouts": totalWorkouts,
    "TotalHours":    int(float64(totalMinutes)/60.0 + 0.5),
    "CurrentStreak": streak,
    "WeeklyGoal":    4,
    "CompletedThisWeek": completedThisWeek,
    "WeeklyData":    weekly,
    "MonthlyStats":  monthly,
    "Achievements":  achievements,
    "CompletedAchievements": completedAchievements,
  }

  a.renderPage(w, r, "progress", "Прогресс", "progress", data)
}

func buildWeeklyData(a *App, userID string) []weeklyDay {
  now := time.Now()
  start := now.AddDate(0, 0, -6)

  data := []weeklyDay{}
  days := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}

  for i := 0; i < 7; i++ {
    date := start.AddDate(0, 0, i)
    var count int
    var minutes int
    _ = a.DB.QueryRow(
      `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
       from workout_sessions
       where user_id = $1 and date(completed_at) = $2`,
      userID,
      date.Format("2006-01-02"),
    ).Scan(&count, &minutes)

    data = append(data, weeklyDay{
      Day:       days[int(date.Weekday())],
      Completed: count > 0,
      Duration:  minutes,
    })
  }

  return data
}

func buildMonthlyStats(a *App, userID string) []monthlyStat {
  now := time.Now()
  stats := []monthlyStat{}
  months := []string{"Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"}

  for i := 2; i >= 0; i-- {
    date := now.AddDate(0, -i, 0)
    monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
    var workouts int
    var minutes int
    _ = a.DB.QueryRow(
      `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
       from workout_sessions
       where user_id = $1 and completed_at >= $2 and completed_at < $3`,
      userID,
      monthStart,
      monthStart.AddDate(0, 1, 0),
    ).Scan(&workouts, &minutes)

    stats = append(stats, monthlyStat{
      Month:    months[int(date.Month())-1],
      Workouts: workouts,
      Hours:    float64(minutes) / 60.0,
    })
  }

  return stats
}
