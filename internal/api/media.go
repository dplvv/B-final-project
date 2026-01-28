package api

import "net/http"

func (api *API) Videos(w http.ResponseWriter, r *http.Request) {
  category := r.URL.Query().Get("category")
  query := r.URL.Query().Get("q")

  rows, err := api.DB.Query(
    `select id, title, description, coalesce(duration_minutes, 0), coalesce(category, ''), coalesce(difficulty, ''), coalesce(url, '')
     from video_tutorials
     where ($1 = '' or category = $1)
       and ($2 = '' or title ilike '%' || $2 || '%')
     order by created_at desc`,
    category,
    query,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  videos := []map[string]any{}
  for rows.Next() {
    var id, title, description, cat, difficulty, url string
    var duration int
    _ = rows.Scan(&id, &title, &description, &duration, &cat, &difficulty, &url)
    videos = append(videos, map[string]any{
      "id": id,
      "title": title,
      "description": description,
      "duration": duration,
      "category": cat,
      "difficulty": difficulty,
      "url": url,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "videos": videos,
    "category": category,
    "query": query,
  })
}

func (api *API) Nutrition(w http.ResponseWriter, r *http.Request) {
  category := r.URL.Query().Get("category")

  rows, err := api.DB.Query(
    `select id, title, description, coalesce(calories, 0), coalesce(category, '')
     from nutrition_items
     where ($1 = '' or category = $1)
     order by created_at desc`,
    category,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  items := []map[string]any{}
  for rows.Next() {
    var id, title, description, cat string
    var calories int
    _ = rows.Scan(&id, &title, &description, &calories, &cat)
    items = append(items, map[string]any{
      "id": id,
      "title": title,
      "description": description,
      "calories": calories,
      "category": cat,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "items": items,
    "category": category,
  })
}

