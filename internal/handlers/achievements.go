package handlers

import (
  "net/http"

  "rehab-app/internal/models"
)

func (a *App) Achievements(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  workoutsCount := 0
  _ = a.DB.QueryRow("select coalesce(count(*), 0) from workout_sessions where user_id = $1 and completed_at is not null", user.ID).Scan(&workoutsCount)
  streak := computeWorkoutStreak(a.DB, user.ID)
  monthlyCount := 0
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0)
     from workout_sessions
     where user_id = $1 and completed_at >= now() - interval '30 days'`,
    user.ID,
  ).Scan(&monthlyCount)

  _ = ensureUserAchievements(a, user.ID, workoutsCount, streak, monthlyCount)

  rows, err := a.DB.Query(
    `select a.id, a.title, a.description, a.icon,
            coalesce(ua.unlocked, false), ua.unlocked_at, coalesce(ua.progress, 0), coalesce(ua.total, 0)
     from achievements a
     left join user_achievements ua on ua.achievement_id = a.id and ua.user_id = $1
     order by a.created_at`,
    user.ID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  achievements := []models.Achievement{}
  for rows.Next() {
    var ach models.Achievement
    _ = rows.Scan(&ach.ID, &ach.Title, &ach.Description, &ach.Icon, &ach.Unlocked, &ach.UnlockedAt, &ach.Progress, &ach.Total)
    achievements = append(achievements, ach)
  }

  data := map[string]any{
    "Achievements": achievements,
  }

  a.renderPage(w, r, "achievements", "Достижения", "", data)
}

func ensureUserAchievements(a *App, userID string, workoutsCount int, streak int, monthlyCount int) error {
  rows, err := a.DB.Query("select id, title from achievements")
  if err != nil {
    return err
  }
  defer rows.Close()

  for rows.Next() {
    var id string
    var title string
    if err := rows.Scan(&id, &title); err != nil {
      continue
    }

    total := 1
    progress := 0
    unlocked := false

    switch title {
    case "Первый шаг":
      total = 1
      progress = workoutsCount
    case "Серия":
      total = 5
      progress = streak
    case "Настойчивость":
      total = 10
      progress = monthlyCount
    default:
      total = 1
      progress = workoutsCount
    }

    if progress >= total {
      unlocked = true
    }

    _, _ = a.DB.Exec(
      `insert into user_achievements (user_id, achievement_id, unlocked, unlocked_at, progress, total)
       values ($1, $2, $3, case when $3 then now() else null end, $4, $5)
       on conflict (user_id, achievement_id)
       do update set unlocked = excluded.unlocked, unlocked_at = excluded.unlocked_at, progress = excluded.progress, total = excluded.total`,
      userID,
      id,
      unlocked,
      progress,
      total,
    )
  }

  return nil
}
