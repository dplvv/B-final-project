package handlers

import (
  "net/http"
  "time"

  "rehab-app/internal/models"
)

type programWorkout struct {
  models.Workout
  ExercisesCount int
  Completed      bool
  RecommendedDate string
}

func (a *App) Program(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var fitnessLevel string
  _ = a.DB.QueryRow("select coalesce(fitness_level, '') from user_profiles where user_id = $1", user.ID).Scan(&fitnessLevel)

  rows, err := a.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, ''),
            coalesce(count(we.exercise_id), 0) as exercise_count,
            exists (select 1 from workout_sessions ws where ws.user_id = $1 and ws.workout_id = w.id and ws.completed_at is not null) as completed
     from workouts w
     left join workout_exercises we on we.workout_id = w.id
     group by w.id
     order by w.created_at`,
    user.ID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  workouts := []programWorkout{}
  completedCount := 0
  totalDuration := 0
  index := 0
  for rows.Next() {
    var workout programWorkout
    err := rows.Scan(
      &workout.ID,
      &workout.Name,
      &workout.Description,
      &workout.Duration,
      &workout.Difficulty,
      &workout.Category,
      &workout.ExercisesCount,
      &workout.Completed,
    )
    if err != nil {
      http.Error(w, "server error", http.StatusInternalServerError)
      return
    }

    if workout.Completed {
      completedCount++
      totalDuration += workout.Duration
    }

    recommendedDate := time.Now().AddDate(0, 0, index).Format("02.01.2006")
    workout.RecommendedDate = recommendedDate
    index++

    workouts = append(workouts, workout)
  }

  data := map[string]any{
    "FitnessLevel":  fitnessLevel,
    "Workouts":      workouts,
    "CompletedCount": completedCount,
    "TotalDuration": totalDuration,
  }

  a.renderPage(w, r, "program", "Программа тренировок", "program", data)
}

func (a *App) StartWorkout(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  workoutID := r.URL.Query().Get("id")
  if workoutID == "" {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  var exercisesCount int
  _ = a.DB.QueryRow("select count(*) from workout_exercises where workout_id = $1", workoutID).Scan(&exercisesCount)

  var sessionID string
  err := a.DB.QueryRow(
    `insert into workout_sessions (user_id, workout_id, total_exercises, completed_exercises)
     values ($1, $2, $3, 0)
     returning id`,
    user.ID,
    workoutID,
    exercisesCount,
  ).Scan(&sessionID)
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  rows, err := a.DB.Query(
    `select exercise_id, sort_order
     from workout_exercises
     where workout_id = $1
     order by sort_order`,
    workoutID,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var exerciseID string
      var order int
      _ = rows.Scan(&exerciseID, &order)
      _, _ = a.DB.Exec(
        `insert into workout_session_exercises (session_id, exercise_id, sort_order)
         values ($1, $2, $3)`,
        sessionID,
        exerciseID,
        order,
      )
    }
  }

  http.Redirect(w, r, "/workout-sessions/"+sessionID, http.StatusSeeOther)
}
