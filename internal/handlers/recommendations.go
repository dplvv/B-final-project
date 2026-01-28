package handlers

import (
  "net/http"

  "rehab-app/internal/models"
)

type recommendationArticle struct {
  ID       string
  Title    string
  Category string
  ReadTime int
  Icon     string
  Excerpt  string
  Body     string
}

type quickTip struct {
  Icon string
  Title string
  Description string
}

func (a *App) Recommendations(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  category := r.URL.Query().Get("category")
  selectedID := r.URL.Query().Get("article")

  rows, err := a.DB.Query(
    `select id, title, category, coalesce(read_time, 5), coalesce(icon, '📘'), coalesce(excerpt, ''), body
     from recommendations
     where ($1 = '' or category = $1)
     order by created_at desc`,
    category,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  articles := []recommendationArticle{}
  var selected recommendationArticle
  for rows.Next() {
    var item recommendationArticle
    _ = rows.Scan(&item.ID, &item.Title, &item.Category, &item.ReadTime, &item.Icon, &item.Excerpt, &item.Body)
    articles = append(articles, item)
    if item.ID == selectedID {
      selected = item
    }
  }

  tips := []quickTip{
    {Icon: "💧", Title: "Пейте воду", Description: "Употребляйте 2-2.5 литра воды в день"},
    {Icon: "😴", Title: "Высыпайтесь", Description: "7-8 часов сна необходимы для восстановления"},
    {Icon: "🎯", Title: "Ставьте цели", Description: "Конкретные цели помогают сохранять мотивацию"},
    {Icon: "📊", Title: "Отслеживайте прогресс", Description: "Записывайте результаты каждой тренировки"},
  }

  videoRows, _ := a.DB.Query(
    `select id, title, coalesce(duration_minutes, 5)
     from video_tutorials
     order by created_at desc
     limit 3`,
  )
  videos := []models.VideoTutorial{}
  if videoRows != nil {
    defer videoRows.Close()
    for videoRows.Next() {
      var video models.VideoTutorial
      _ = videoRows.Scan(&video.ID, &video.Title, &video.Duration)
      videos = append(videos, video)
    }
  }

  categories := []string{"Все"}
  seen := map[string]bool{}
  for _, article := range articles {
    if article.Category == "" || seen[article.Category] {
      continue
    }
    categories = append(categories, article.Category)
    seen[article.Category] = true
  }

  data := map[string]any{
    "Articles":    articles,
    "Selected":    selected,
    "HasSelected": selected.ID != "",
    "Category":    category,
    "Categories":  categories,
    "Tips":        tips,
    "Videos":      videos,
  }

  a.renderPage(w, r, "recommendations", "Рекомендации", "recommendations", data)
}
