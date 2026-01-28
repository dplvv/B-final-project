package handlers

import (
  "net/http"
  "time"

  "rehab-app/internal/middleware"
  "rehab-app/internal/models"
)

func (a *App) baseData(w http.ResponseWriter, r *http.Request) map[string]any {
  data := map[string]any{}
  if user := middleware.UserFromContext(r.Context()); user != nil {
    data["User"] = user
  }
  data["Flash"] = a.popFlash(w, r)
  return data
}

func (a *App) popFlash(w http.ResponseWriter, r *http.Request) string {
  cookie, err := r.Cookie("flash")
  if err != nil {
    return ""
  }
  value := cookie.Value
  a.clearFlash(w)
  return value
}

func (a *App) clearFlash(w http.ResponseWriter) {
  http.SetCookie(w, &http.Cookie{
    Name:     "flash",
    Value:    "",
    Path:     "/",
    Expires:  time.Unix(0, 0),
    MaxAge:   -1,
    HttpOnly: true,
  })
}

func (a *App) setFlash(w http.ResponseWriter, message string) {
  http.SetCookie(w, &http.Cookie{
    Name:     "flash",
    Value:    message,
    Path:     "/",
    MaxAge:   5,
    HttpOnly: true,
  })
}

func (a *App) userFromRequest(r *http.Request) *models.User {
  return middleware.UserFromContext(r.Context())
}
