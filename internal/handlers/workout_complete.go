package handlers

import (
  "net/http"
)

func (a *App) WorkoutComplete(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  sessionID := r.URL.Query().Get("id")
  if sessionID == "" {
    http.Redirect(w, r, "/progress", http.StatusSeeOther)
    return
  }

  var workoutName string
  var duration int
  var totalExercises int
  var completedExercises int
  var calories int
  var ownerID string
  err := a.DB.QueryRow(
    `select w.name, coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.total_exercises, 0), coalesce(ws.completed_exercises, 0), coalesce(ws.calories_burned, 0),
            ws.user_id
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.id = $1`,
    sessionID,
  ).Scan(&workoutName, &duration, &totalExercises, &completedExercises, &calories, &ownerID)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  if ownerID != user.ID {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }

  completionRate := 0
  if totalExercises > 0 {
    completionRate = int(float64(completedExercises) / float64(totalExercises) * 100)
  }

  data := map[string]any{
    "WorkoutName":       workoutName,
    "Duration":          duration,
    "TotalExercises":    totalExercises,
    "CompletedExercises": completedExercises,
    "CaloriesBurned":    calories,
    "CompletionRate":    completionRate,
  }

  a.renderFullPage(w, r, "workout_complete", "Тренировка завершена", data)
}
