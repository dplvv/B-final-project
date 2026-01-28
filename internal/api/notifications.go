package api

import (
  "net/http"

  "github.com/go-chi/chi/v5"
)

func (api *API) NotificationsRead(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, _ = api.DB.Exec(
    `update notifications set read_at = now() where id = $1 and user_id = $2`,
    id,
    userID,
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "read"})
}

