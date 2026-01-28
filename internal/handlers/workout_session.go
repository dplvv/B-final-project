package handlers

import (
  "database/sql"
  "net/http"

  "github.com/go-chi/chi/v5"
)

type sessionExercise struct {
  ID            string
  Name          string
  Description   string
  Sets          int
  Reps          string
  RestSeconds   int
  CompletedSets int
  Completed     bool
}

type sessionView struct {
  SessionID      string
  WorkoutName    string
  WorkoutDuration int
  Exercises      []sessionExercise
  CurrentExercise sessionExercise
  CurrentIndex   int
  CurrentSet     int
  TotalSets      int
  CompletedSets  int
}

func (a *App) WorkoutSession(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  var workoutName string
  var workoutDuration int
  var ownerID string
  err := a.DB.QueryRow(
    `select w.name, w.duration_minutes, ws.user_id
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.id = $1`,
    sessionID,
  ).Scan(&workoutName, &workoutDuration, &ownerID)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  if ownerID != user.ID {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }

  rows, err := a.DB.Query(
    `select wse.id, e.name, e.description,
            coalesce(we.sets, e.sets, 1),
            coalesce(we.reps, e.reps, '10'),
            coalesce(we.rest_seconds, e.rest_seconds, 30),
            wse.completed_sets,
            wse.completed
     from workout_session_exercises wse
     join exercises e on e.id = wse.exercise_id
     left join workout_exercises we on we.exercise_id = e.id
       and we.workout_id = (select workout_id from workout_sessions where id = $1)
     where wse.session_id = $1
     order by wse.sort_order`,
    sessionID,
  )
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }
  defer rows.Close()

  exercises := []sessionExercise{}
  for rows.Next() {
    var ex sessionExercise
    if err := rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Sets, &ex.Reps, &ex.RestSeconds, &ex.CompletedSets, &ex.Completed); err != nil {
      http.Error(w, "server error", http.StatusInternalServerError)
      return
    }
    exercises = append(exercises, ex)
  }

  if len(exercises) == 0 {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  currentIndex := -1
  for i, ex := range exercises {
    if !ex.Completed {
      currentIndex = i
      break
    }
  }
  if currentIndex == -1 {
    http.Redirect(w, r, "/workout-complete?id="+sessionID, http.StatusSeeOther)
    return
  }

  totalSets := 0
  completedSets := 0
  for _, ex := range exercises {
    totalSets += ex.Sets
    completedSets += ex.CompletedSets
    if ex.Completed {
      completedSets += ex.Sets - ex.CompletedSets
    }
  }

  current := exercises[currentIndex]
  currentSet := current.CompletedSets + 1

  progressPercent := 0
  if totalSets > 0 {
    progressPercent = int(float64(completedSets) / float64(totalSets) * 100)
  }

  data := map[string]any{
    "SessionID":       sessionID,
    "WorkoutName":     workoutName,
    "WorkoutDuration": workoutDuration,
    "CurrentExercise": current,
    "CurrentIndex":    currentIndex,
    "TotalExercises":  len(exercises),
    "CurrentSet":      currentSet,
    "TotalSets":       totalSets,
    "CompletedSets":   completedSets,
    "ProgressPercent": progressPercent,
  }

  a.renderFullPage(w, r, "workout_session", "Тренировка", data)
}

func (a *App) WorkoutCompleteSet(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  var ownerID string
  err := a.DB.QueryRow("select user_id from workout_sessions where id = $1", sessionID).Scan(&ownerID)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  if ownerID != user.ID {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }

  var exID string
  var sets int
  var completedSets int
  err = a.DB.QueryRow(
    `select wse.id, coalesce(we.sets, e.sets, 1), wse.completed_sets
     from workout_session_exercises wse
     join exercises e on e.id = wse.exercise_id
     left join workout_exercises we on we.exercise_id = e.id
       and we.workout_id = (select workout_id from workout_sessions where id = $1)
     where wse.session_id = $1 and wse.completed = false
     order by wse.sort_order
     limit 1`,
    sessionID,
  ).Scan(&exID, &sets, &completedSets)
  if err != nil {
    if err == sql.ErrNoRows {
      http.Redirect(w, r, "/workout-complete?id="+sessionID, http.StatusSeeOther)
      return
    }
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  completedSets++
  completed := completedSets >= sets

  _, _ = a.DB.Exec(
    `update workout_session_exercises
     set completed_sets = $1, completed = $2
     where id = $3`,
    completedSets,
    completed,
    exID,
  )

  if completed {
    _, _ = a.DB.Exec(
      `update workout_sessions
       set completed_exercises = completed_exercises + 1
       where id = $1`,
      sessionID,
    )
  }

  var totalExercises int
  var completedExercises int
  _ = a.DB.QueryRow(
    `select total_exercises, completed_exercises from workout_sessions where id = $1`,
    sessionID,
  ).Scan(&totalExercises, &completedExercises)

  if totalExercises > 0 && completedExercises >= totalExercises {
    _ = completeWorkoutSession(a, sessionID)
    http.Redirect(w, r, "/workout-complete?id="+sessionID, http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/workout-sessions/"+sessionID, http.StatusSeeOther)
}

func completeWorkoutSession(a *App, sessionID string) error {
  _, err := a.DB.Exec(
    `update workout_sessions
     set completed_at = now(), duration_minutes = coalesce(duration_minutes, 30), calories_burned = coalesce(calories_burned, 250)
     where id = $1`,
    sessionID,
  )
  if err != nil {
    return err
  }

  var userID string
  _ = a.DB.QueryRow("select user_id from workout_sessions where id = $1", sessionID).Scan(&userID)
  if userID != "" {
    points := 10
    _, _ = a.DB.Exec(
      `update user_points
       set points_balance = points_balance + $1, points_total = points_total + $1, updated_at = now()
       where user_id = $2`,
      points,
      userID,
    )
    _, _ = a.DB.Exec(
      `insert into notifications (user_id, title, message, type)
       values ($1, $2, $3, $4)`,
      userID,
      "Начислены баллы",
      "Вы получили 10 баллов за завершение тренировки",
      "success",
    )
  }

  return nil
}
