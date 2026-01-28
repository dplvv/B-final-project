package api

import (
  "net/http"
  "time"

  "github.com/go-chi/chi/v5"
)

type practiceRequest struct {
  Date string `json:"date"`
}

func (api *API) Recommendations(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  category := r.URL.Query().Get("category")

  rows, err := api.DB.Query(
    `select r.id, r.title, r.category, coalesce(r.read_time, 5), coalesce(r.icon, '📘'), coalesce(r.excerpt, ''), r.body,
            exists (select 1 from recommendation_bookmarks rb where rb.user_id = $2 and rb.recommendation_id = r.id) as bookmarked
     from recommendations r
     where ($1 = '' or r.category = $1)
     order by r.created_at desc`,
    category,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  articles := []map[string]any{}
  categories := []string{"Все"}
  seen := map[string]bool{}
  for rows.Next() {
    var id, title, cat, icon, excerpt, body string
    var readTime int
    var bookmarked bool
    _ = rows.Scan(&id, &title, &cat, &readTime, &icon, &excerpt, &body, &bookmarked)
    articles = append(articles, map[string]any{
      "id": id,
      "title": title,
      "category": cat,
      "read_time": readTime,
      "icon": icon,
      "excerpt": excerpt,
      "body": body,
      "bookmarked": bookmarked,
    })
    if cat != "" && !seen[cat] {
      categories = append(categories, cat)
      seen[cat] = true
    }
  }

  tips := []map[string]any{
    {"icon": "💧", "title": "Пейте воду", "description": "Употребляйте 2-2.5 литра воды в день"},
    {"icon": "😴", "title": "Высыпайтесь", "description": "7-8 часов сна необходимы для восстановления"},
    {"icon": "🎯", "title": "Ставьте цели", "description": "Конкретные цели помогают сохранять мотивацию"},
    {"icon": "📊", "title": "Отслеживайте прогресс", "description": "Записывайте результаты каждой тренировки"},
  }

  videoRows, _ := api.DB.Query(
    `select id, title, coalesce(duration_minutes, 5)
     from video_tutorials
     order by created_at desc
     limit 3`,
  )
  videos := []map[string]any{}
  if videoRows != nil {
    defer videoRows.Close()
    for videoRows.Next() {
      var id, title string
      var duration int
      _ = videoRows.Scan(&id, &title, &duration)
      videos = append(videos, map[string]any{
        "id": id,
        "title": title,
        "duration": duration,
      })
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "articles": articles,
    "category": category,
    "categories": categories,
    "tips": tips,
    "videos": videos,
  })
}

func (api *API) RecommendationDetail(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var title, cat, icon, excerpt, body string
  var readTime int
  err := api.DB.QueryRow(
    `select id, title, category, coalesce(read_time, 5), coalesce(icon, '📘'), coalesce(excerpt, ''), body
     from recommendations where id = $1`,
    id,
  ).Scan(&id, &title, &cat, &readTime, &icon, &excerpt, &body)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }

  bookmarked := false
  _ = api.DB.QueryRow(
    `select exists (select 1 from recommendation_bookmarks where user_id = $1 and recommendation_id = $2)`,
    userID,
    id,
  ).Scan(&bookmarked)

  writeJSON(w, http.StatusOK, map[string]any{
    "article": map[string]any{
      "id": id,
      "title": title,
      "category": cat,
      "read_time": readTime,
      "icon": icon,
      "excerpt": excerpt,
      "body": body,
      "bookmarked": bookmarked,
    },
  })
}

func (api *API) RecommendationBookmark(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, _ = api.DB.Exec(
    `insert into recommendation_bookmarks (user_id, recommendation_id)
     values ($1, $2)
     on conflict do nothing`,
    userID,
    id,
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "bookmarked"})
}

func (api *API) RecommendationBookmarkRemove(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, _ = api.DB.Exec(
    `delete from recommendation_bookmarks where user_id = $1 and recommendation_id = $2`,
    userID,
    id,
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "removed"})
}

func (api *API) RecommendationPractice(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req practiceRequest
  _ = decodeJSON(r, &req)

  date := time.Now()
  if req.Date != "" {
    if parsed, err := time.Parse("2006-01-02", req.Date); err == nil {
      date = parsed
    }
  }

  var title string
  _ = api.DB.QueryRow("select title from recommendations where id = $1", id).Scan(&title)

  _, _ = api.DB.Exec(
    `insert into calendar_events (user_id, title, event_date, event_type, metadata)
     values ($1, $2, $3, 'practice', jsonb_build_object('recommendation_id', $4))`,
    userID,
    "Практика: "+title,
    date.Format("2006-01-02"),
    id,
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "scheduled"})
}

