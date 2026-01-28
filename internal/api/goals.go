package api

import (
  "net/http"

  "github.com/go-chi/chi/v5"
)

type goalRequest struct {
  Title       string `json:"title"`
  Description string `json:"description"`
  TargetDate  string `json:"target_date"`
  Category    string `json:"category"`
  Progress    int    `json:"progress"`
}

func (api *API) Goals(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  rows, err := api.DB.Query(
    `select id, title, description, coalesce(target_date::text, ''), progress, category
     from goals where user_id = $1
     order by created_at desc`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  goals := []map[string]any{}
  for rows.Next() {
    var id, title, description, targetDate, category string
    var progress int
    _ = rows.Scan(&id, &title, &description, &targetDate, &progress, &category)
    goals = append(goals, map[string]any{
      "id": id,
      "title": title,
      "description": description,
      "target_date": targetDate,
      "progress": progress,
      "category": category,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"goals": goals})
}

func (api *API) GoalsCreate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var req goalRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Title == "" || req.Category == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  _, err := api.DB.Exec(
    `insert into goals (user_id, title, description, target_date, category, progress)
     values ($1, $2, $3, nullif($4, '')::date, $5, $6)`,
    userID,
    req.Title,
    req.Description,
    req.TargetDate,
    req.Category,
    req.Progress,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"status": "created"})
}

func (api *API) GoalsUpdate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  goalID := chi.URLParam(r, "id")
  if goalID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req goalRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update goals
     set title = coalesce(nullif($1, ''), title),
         description = coalesce($2, description),
         target_date = nullif($3, '')::date,
         category = coalesce(nullif($4, ''), category),
         progress = $5
     where id = $6 and user_id = $7`,
    req.Title,
    req.Description,
    req.TargetDate,
    req.Category,
    req.Progress,
    goalID,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) GoalsDelete(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  goalID := chi.URLParam(r, "id")
  if goalID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from goals where id = $1 and user_id = $2", goalID, userID)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

