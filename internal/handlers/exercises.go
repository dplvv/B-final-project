package handlers

import (
  "net/http"

  "github.com/go-chi/chi/v5"
  "rehab-app/internal/models"
)

func (a *App) ExerciseLibrary(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  category := r.URL.Query().Get("category")
  difficulty := r.URL.Query().Get("difficulty")
  search := r.URL.Query().Get("q")

  query := `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
                   coalesce(sets, 0), coalesce(reps, ''), coalesce(duration_seconds, 0), coalesce(rest_seconds, 0)
            from exercises
            where ($1 = '' or category = $1)
              and ($2 = '' or difficulty = $2)
              and ($3 = '' or name ilike '%' || $3 || '%')
            order by name`

  rows, err := a.DB.Query(query, category, difficulty, search)
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  exercises := []models.Exercise{}
  for rows.Next() {
    var ex models.Exercise
    _ = rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Duration, &ex.Rest)
    exercises = append(exercises, ex)
  }

  data := map[string]any{
    "Exercises": exercises,
    "Category":  category,
    "Difficulty": difficulty,
    "Query":     search,
  }

  a.renderPage(w, r, "exercise_library", "Библиотека упражнений", "", data)
}

func (a *App) ExerciseDetail(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  exerciseID := chi.URLParam(r, "id")
  if exerciseID == "" {
    http.Redirect(w, r, "/exercise-library", http.StatusSeeOther)
    return
  }

  var ex models.Exercise
  err := a.DB.QueryRow(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(duration_seconds, 0), coalesce(rest_seconds, 0),
            muscle_groups, equipment, coalesce(video_url, '')
     from exercises where id = $1`,
    exerciseID,
  ).Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Duration, &ex.Rest, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  data := map[string]any{
    "Exercise": ex,
  }

  a.renderFullPage(w, r, "exercise_detail", "Упражнение", data)
}
