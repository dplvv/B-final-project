package api

import (
  "net/http"
  "time"
)

func (api *API) Achievements(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  workoutsCount := 0
  _ = api.DB.QueryRow("select coalesce(count(*), 0) from workout_sessions where user_id = $1 and completed_at is not null", userID).Scan(&workoutsCount)
  streak := computeWorkoutStreak(api.DB, userID)
  monthlyCount := 0
  _ = api.DB.QueryRow(
    `select coalesce(count(*), 0)
     from workout_sessions
     where user_id = $1 and completed_at >= now() - interval '30 days'`,
    userID,
  ).Scan(&monthlyCount)

  _ = ensureUserAchievements(api, userID, workoutsCount, streak, monthlyCount)

  rows, err := api.DB.Query(
    `select a.id, a.title, a.description, a.icon,
            coalesce(ua.unlocked, false), ua.unlocked_at, coalesce(ua.progress, 0), coalesce(ua.total, 0)
     from achievements a
     left join user_achievements ua on ua.achievement_id = a.id and ua.user_id = $1
     order by a.created_at`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  achievements := []map[string]any{}
  for rows.Next() {
    var id, title, description, icon string
    var unlocked bool
    var unlockedAt *time.Time
    var progress, total int
    _ = rows.Scan(&id, &title, &description, &icon, &unlocked, &unlockedAt, &progress, &total)
    unlockedDate := ""
    if unlockedAt != nil {
      unlockedDate = unlockedAt.Format("2006-01-02")
    }
    achievements = append(achievements, map[string]any{
      "id": id,
      "title": title,
      "description": description,
      "icon": icon,
      "unlocked": unlocked,
      "unlocked_date": unlockedDate,
      "progress": progress,
      "total": total,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"achievements": achievements})
}

func ensureUserAchievements(api *API, userID string, workoutsCount int, streak int, monthlyCount int) error {
  rows, err := api.DB.Query("select id, title from achievements")
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

    _, _ = api.DB.Exec(
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
