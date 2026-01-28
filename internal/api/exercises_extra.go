package api

import (
  "net/http"

  "github.com/go-chi/chi/v5"
)

func (api *API) ExerciseDetail(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var name, description, category, difficulty, reps, videoURL string
  var sets, duration, rest int
  var muscleGroups []string
  var equipment []string
  err := api.DB.QueryRow(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(duration_seconds, 0), coalesce(rest_seconds, 0),
            muscle_groups, equipment, coalesce(video_url, '')
     from exercises where id = $1`,
    id,
  ).Scan(&id, &name, &description, &category, &difficulty, &sets, &reps, &duration, &rest, &muscleGroups, &equipment, &videoURL)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "exercise": map[string]any{
      "id": id,
      "name": name,
      "description": description,
      "category": category,
      "difficulty": difficulty,
      "sets": sets,
      "reps": reps,
      "duration": duration,
      "rest": rest,
      "muscle_groups": muscleGroups,
      "equipment": equipment,
      "video_url": videoURL,
    },
  })
}

