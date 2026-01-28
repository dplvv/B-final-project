package handlers

import (
  "net/http"
  "time"
)

type feedbackItem struct {
  ID       string
  Workout  string
  Rating   int
  Comment  string
  Created  string
}

func (a *App) Feedback(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  sessionID := r.URL.Query().Get("session")

  rows, err := a.DB.Query(
    `select f.id, w.name, f.rating, coalesce(f.comment, ''), f.created_at
     from feedback f
     left join workout_sessions ws on ws.id = f.workout_session_id
     left join workouts w on w.id = ws.workout_id
     where f.user_id = $1
     order by f.created_at desc
     limit 5`,
    user.ID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  feedbacks := []feedbackItem{}
  for rows.Next() {
    var item feedbackItem
    var created time.Time
    _ = rows.Scan(&item.ID, &item.Workout, &item.Rating, &item.Comment, &created)
    item.Created = created.Format("02.01.2006")
    feedbacks = append(feedbacks, item)
  }

  data := map[string]any{
    "Feedbacks": feedbacks,
    "SessionID": sessionID,
  }

  a.renderPage(w, r, "feedback", "Отзыв", "", data)
}

func (a *App) FeedbackSubmit(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  sessionID := r.FormValue("session")
  rating := r.FormValue("rating")
  comment := r.FormValue("comment")

  if rating == "" {
    a.setFlash(w, "Укажите оценку")
    http.Redirect(w, r, "/feedback", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec(
    `insert into feedback (user_id, workout_session_id, rating, comment)
     values ($1, nullif($2, '')::uuid, $3, $4)`,
    user.ID,
    sessionID,
    rating,
    comment,
  )

  a.setFlash(w, "Спасибо за отзыв!")
  http.Redirect(w, r, "/feedback", http.StatusSeeOther)
}
