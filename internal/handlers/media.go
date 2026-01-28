package handlers

import (
  "net/http"

  "rehab-app/internal/models"
)

func (a *App) VideoTutorials(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  category := r.URL.Query().Get("category")
  query := r.URL.Query().Get("q")

  rows, err := a.DB.Query(
    `select id, title, description, coalesce(duration_minutes, 0), coalesce(category, ''), coalesce(difficulty, ''), coalesce(url, '')
     from video_tutorials
     where ($1 = '' or category = $1)
       and ($2 = '' or title ilike '%' || $2 || '%')
     order by created_at desc`,
    category,
    query,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  videos := []models.VideoTutorial{}
  for rows.Next() {
    var video models.VideoTutorial
    _ = rows.Scan(&video.ID, &video.Title, &video.Description, &video.Duration, &video.Category, &video.Difficulty, &video.URL)
    videos = append(videos, video)
  }

  data := map[string]any{
    "Videos": videos,
    "Category": category,
    "Query": query,
  }

  a.renderPage(w, r, "video_tutorials", "Видео уроки", "", data)
}

func (a *App) Nutrition(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  category := r.URL.Query().Get("category")

  rows, err := a.DB.Query(
    `select id, title, description, coalesce(calories, 0), coalesce(category, '')
     from nutrition_items
     where ($1 = '' or category = $1)
     order by created_at desc`,
    category,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  items := []models.NutritionItem{}
  for rows.Next() {
    var item models.NutritionItem
    _ = rows.Scan(&item.ID, &item.Title, &item.Description, &item.Calories, &item.Category)
    items = append(items, item)
  }

  data := map[string]any{
    "Items": items,
    "Category": category,
  }

  a.renderPage(w, r, "nutrition", "Питание", "", data)
}
