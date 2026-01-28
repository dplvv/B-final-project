package api

import "net/http"

type settingsPayload struct {
  Notifications struct {
    Enabled          bool `json:"enabled"`
    WorkoutReminders bool `json:"workout_reminders"`
    AchievementAlerts bool `json:"achievement_alerts"`
    WeeklyReports    bool `json:"weekly_reports"`
    RemindersEnabled bool `json:"reminders_enabled"`
  } `json:"notifications"`
  Preferences struct {
    Language string `json:"language"`
    Theme    string `json:"theme"`
    Units    string `json:"units"`
  } `json:"preferences"`
  Privacy struct {
    ShareProgress    bool `json:"share_progress"`
    ShowInLeaderboard bool `json:"show_in_leaderboard"`
  } `json:"privacy"`
}

func (api *API) Settings(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  settings, err := api.loadSettings(userID)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  writeJSON(w, http.StatusOK, settings)
}

func (api *API) SettingsUpdate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var payload settingsPayload
  if err := decodeJSON(r, &payload); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if payload.Preferences.Language == "" {
    payload.Preferences.Language = "ru"
  }
  if payload.Preferences.Theme == "" {
    payload.Preferences.Theme = "light"
  }
  if payload.Preferences.Units == "" {
    payload.Preferences.Units = "metric"
  }

  _, err := api.DB.Exec(
    `update user_settings
     set notifications_enabled = $1,
         reminders_enabled = $2,
         language = $3,
         theme = $4,
         workout_reminders = $5,
         achievement_alerts = $6,
         weekly_reports = $7,
         share_progress = $8,
         show_in_leaderboard = $9,
         units = $10,
         updated_at = now()
     where user_id = $11`,
    payload.Notifications.Enabled,
    payload.Notifications.RemindersEnabled,
    payload.Preferences.Language,
    payload.Preferences.Theme,
    payload.Notifications.WorkoutReminders,
    payload.Notifications.AchievementAlerts,
    payload.Notifications.WeeklyReports,
    payload.Privacy.ShareProgress,
    payload.Privacy.ShowInLeaderboard,
    payload.Preferences.Units,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  settings, _ := api.loadSettings(userID)
  writeJSON(w, http.StatusOK, settings)
}

