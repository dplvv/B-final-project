package handlers

import "net/http"

func (a *App) Settings(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var notifications bool
  var reminders bool
  var language string
  var theme string
  _ = a.DB.QueryRow(
    `select notifications_enabled, reminders_enabled, language, theme
     from user_settings where user_id = $1`,
    user.ID,
  ).Scan(&notifications, &reminders, &language, &theme)

  data := map[string]any{
    "Notifications": notifications,
    "Reminders":     reminders,
    "Language":      language,
    "Theme":         theme,
  }

  a.renderPage(w, r, "settings", "Настройки", "", data)
}

func (a *App) SettingsUpdate(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  notifications := r.FormValue("notifications") == "on"
  reminders := r.FormValue("reminders") == "on"
  language := r.FormValue("language")
  theme := r.FormValue("theme")

  _, _ = a.DB.Exec(
    `update user_settings
     set notifications_enabled = $1, reminders_enabled = $2, language = $3, theme = $4, updated_at = now()
     where user_id = $5`,
    notifications,
    reminders,
    language,
    theme,
    user.ID,
  )

  a.setFlash(w, "Настройки сохранены")
  http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
