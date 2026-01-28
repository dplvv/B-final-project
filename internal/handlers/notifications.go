package handlers

import (
  "net/http"
  "time"

  "rehab-app/internal/models"
)

func (a *App) Notifications(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  rows, err := a.DB.Query(
    `select id, title, message, type, created_at, read_at
     from notifications
     where user_id = $1
     order by created_at desc`,
    user.ID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  notifications := []models.Notification{}
  for rows.Next() {
    var item models.Notification
    var created time.Time
    var readAt *time.Time
    _ = rows.Scan(&item.ID, &item.Title, &item.Message, &item.Type, &created, &readAt)
    item.Created = created.Format("02.01.2006")
    item.Read = readAt != nil
    notifications = append(notifications, item)
  }

  data := map[string]any{
    "Notifications": notifications,
  }

  a.renderPage(w, r, "notifications", "Уведомления", "", data)
}

func (a *App) NotificationsRead(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  id := r.FormValue("id")
  if id == "" {
    http.Redirect(w, r, "/notifications", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec("update notifications set read_at = now() where id = $1 and user_id = $2", id, user.ID)
  http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}
