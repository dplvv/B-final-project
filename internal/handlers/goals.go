package handlers

import (
  "net/http"
  "time"

  "rehab-app/internal/models"
)

func (a *App) Goals(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  rows, err := a.DB.Query(
    `select id, title, description, coalesce(target_date::text, ''), progress, category
     from goals where user_id = $1
     order by created_at desc`,
    user.ID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  goals := []models.Goal{}
  for rows.Next() {
    var goal models.Goal
    _ = rows.Scan(&goal.ID, &goal.Title, &goal.Description, &goal.TargetDate, &goal.Progress, &goal.Category)
    goals = append(goals, goal)
  }

  data := map[string]any{
    "Goals": goals,
    "Today": time.Now().Format("2006-01-02"),
  }

  a.renderPage(w, r, "goals", "Цели", "", data)
}

func (a *App) GoalsCreate(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  title := r.FormValue("title")
  description := r.FormValue("description")
  target := r.FormValue("target_date")
  category := r.FormValue("category")

  if title == "" || category == "" {
    a.setFlash(w, "Заполните обязательные поля")
    http.Redirect(w, r, "/goals", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec(
    `insert into goals (user_id, title, description, target_date, category)
     values ($1, $2, $3, nullif($4, '')::date, $5)`,
    user.ID,
    title,
    description,
    target,
    category,
  )

  http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (a *App) GoalsUpdateProgress(w http.ResponseWriter, r *http.Request) {
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
  progress := r.FormValue("progress")
  if id == "" {
    http.Redirect(w, r, "/goals", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec("update goals set progress = $1 where id = $2 and user_id = $3", progress, id, user.ID)
  http.Redirect(w, r, "/goals", http.StatusSeeOther)
}
