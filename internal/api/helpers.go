package api

import (
  "encoding/json"
  "errors"
  "net/http"

  "github.com/jackc/pgconn"
)

func decodeJSON(r *http.Request, dst any) error {
  dec := json.NewDecoder(r.Body)
  dec.DisallowUnknownFields()
  return dec.Decode(dst)
}

func isUniqueViolation(err error) bool {
  var pgErr *pgconn.PgError
  if errors.As(err, &pgErr) {
    return pgErr.Code == "23505"
  }
  return false
}

func (api *API) loadSettings(userID string) (map[string]any, error) {
  var notificationsEnabled bool
  var remindersEnabled bool
  var language string
  var theme string
  var workoutReminders bool
  var achievementAlerts bool
  var weeklyReports bool
  var shareProgress bool
  var showInLeaderboard bool
  var units string

  err := api.DB.QueryRow(
    `select notifications_enabled, reminders_enabled, language, theme,
            workout_reminders, achievement_alerts, weekly_reports,
            share_progress, show_in_leaderboard, units
     from user_settings where user_id = $1`,
    userID,
  ).Scan(
    &notificationsEnabled,
    &remindersEnabled,
    &language,
    &theme,
    &workoutReminders,
    &achievementAlerts,
    &weeklyReports,
    &shareProgress,
    &showInLeaderboard,
    &units,
  )
  if err != nil {
    return map[string]any{}, err
  }

  return map[string]any{
    "notifications": map[string]any{
      "enabled": notificationsEnabled,
      "workout_reminders": workoutReminders,
      "achievement_alerts": achievementAlerts,
      "weekly_reports": weeklyReports,
      "reminders_enabled": remindersEnabled,
    },
    "preferences": map[string]any{
      "language": language,
      "theme": theme,
      "units": units,
    },
    "privacy": map[string]any{
      "share_progress": shareProgress,
      "show_in_leaderboard": showInLeaderboard,
    },
  }, nil
}

