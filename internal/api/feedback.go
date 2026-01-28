package api

import (
  "net/http"
  "time"
)

type feedbackRequest struct {
  SessionID string `json:"session_id"`
  Rating    int    `json:"rating"`
  Comment   string `json:"comment"`
}

func (api *API) Feedback(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  rows, err := api.DB.Query(
    `select f.id, w.name, f.rating, coalesce(f.comment, ''), f.created_at
     from feedback f
     left join workout_sessions ws on ws.id = f.workout_session_id
     left join workouts w on w.id = ws.workout_id
     where f.user_id = $1
     order by f.created_at desc
     limit 20`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  feedbacks := []map[string]any{}
  for rows.Next() {
    var id, workout, comment string
    var rating int
    var created time.Time
    _ = rows.Scan(&id, &workout, &rating, &comment, &created)
    feedbacks = append(feedbacks, map[string]any{
      "id": id,
      "workout": workout,
      "rating": rating,
      "comment": comment,
      "created": created.Format("2006-01-02"),
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"feedbacks": feedbacks})
}

func (api *API) FeedbackSubmit(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var req feedbackRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if req.Rating == 0 {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "rating required"})
    return
  }

  _, err := api.DB.Exec(
    `insert into feedback (user_id, workout_session_id, rating, comment)
     values ($1, nullif($2, '')::uuid, $3, $4)`,
    userID,
    req.SessionID,
    req.Rating,
    req.Comment,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"status": "created"})
}

