package api

import (
  "database/sql"
  "net/http"
  "time"

  "github.com/go-chi/chi/v5"
)

type startSessionRequest struct {
  WorkoutID string `json:"workout_id"`
}

func (api *API) Program(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  rows, err := api.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, ''),
            coalesce(count(we.exercise_id), 0) as exercise_count,
            exists (select 1 from workout_sessions ws where ws.user_id = $1 and ws.workout_id = w.id and ws.completed_at is not null) as completed
     from workouts w
     left join workout_exercises we on we.workout_id = w.id
     group by w.id
     order by w.created_at`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  workouts := []map[string]any{}
  completedCount := 0
  totalDuration := 0
  index := 0
  for rows.Next() {
    var id, name, description, difficulty, category string
    var duration int
    var exercisesCount int
    var completed bool
    if err := rows.Scan(&id, &name, &description, &duration, &difficulty, &category, &exercisesCount, &completed); err != nil {
      writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
      return
    }

    if completed {
      completedCount++
      totalDuration += duration
    }

    recommendedDate := time.Now().AddDate(0, 0, index).Format("2006-01-02")
    index++

    workouts = append(workouts, map[string]any{
      "id": id,
      "name": name,
      "description": description,
      "duration": duration,
      "difficulty": difficulty,
      "category": category,
      "exercises_count": exercisesCount,
      "completed": completed,
      "recommended_date": recommendedDate,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "workouts": workouts,
    "completed_count": completedCount,
    "total_duration": totalDuration,
  })
}

func (api *API) WorkoutDetail(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var name, description, difficulty, category string
  var duration int
  err := api.DB.QueryRow(
    `select name, description, duration_minutes, difficulty, coalesce(category, '')
     from workouts where id = $1`,
    workoutID,
  ).Scan(&name, &description, &duration, &difficulty, &category)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }

  rows, err := api.DB.Query(
    `select e.id, e.name, e.description,
            coalesce(we.sets, e.sets, 1),
            coalesce(we.reps, e.reps, '10'),
            coalesce(we.duration_seconds, e.duration_seconds, 0),
            coalesce(we.rest_seconds, e.rest_seconds, 30),
            coalesce(e.video_url, ''),
            coalesce(e.category, ''),
            coalesce(e.difficulty, '')
     from workout_exercises we
     join exercises e on e.id = we.exercise_id
     where we.workout_id = $1
     order by we.sort_order`,
    workoutID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  exercises := []map[string]any{}
  for rows.Next() {
    var exID, exName, exDescription, exReps, videoURL, exCategory, exDifficulty string
    var exSets, exDuration, exRest int
    _ = rows.Scan(&exID, &exName, &exDescription, &exSets, &exReps, &exDuration, &exRest, &videoURL, &exCategory, &exDifficulty)
    exercises = append(exercises, map[string]any{
      "id": exID,
      "name": exName,
      "description": exDescription,
      "sets": exSets,
      "reps": exReps,
      "duration": exDuration,
      "rest": exRest,
      "video_url": videoURL,
      "category": exCategory,
      "difficulty": exDifficulty,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "workout": map[string]any{
      "id": workoutID,
      "name": name,
      "description": description,
      "duration": duration,
      "difficulty": difficulty,
      "category": category,
      "exercises": exercises,
    },
  })
}

func (api *API) StartWorkoutSession(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var req startSessionRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.WorkoutID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing workout_id"})
    return
  }

  var exercisesCount int
  _ = api.DB.QueryRow("select count(*) from workout_exercises where workout_id = $1", req.WorkoutID).Scan(&exercisesCount)

  var sessionID string
  err := api.DB.QueryRow(
    `insert into workout_sessions (user_id, workout_id, total_exercises, completed_exercises)
     values ($1, $2, $3, 0)
     returning id`,
    userID,
    req.WorkoutID,
    exercisesCount,
  ).Scan(&sessionID)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  rows, err := api.DB.Query(
    `select exercise_id, sort_order
     from workout_exercises
     where workout_id = $1
     order by sort_order`,
    req.WorkoutID,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var exerciseID string
      var order int
      _ = rows.Scan(&exerciseID, &order)
      _, _ = api.DB.Exec(
        `insert into workout_session_exercises (session_id, exercise_id, sort_order)
         values ($1, $2, $3)`,
        sessionID,
        exerciseID,
        order,
      )
    }
  }

  writeJSON(w, http.StatusCreated, map[string]any{"session_id": sessionID})
}

func (api *API) WorkoutSession(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var workoutID string
  var workoutName string
  var workoutDuration int
  var ownerID string
  err := api.DB.QueryRow(
    `select ws.workout_id, w.name, w.duration_minutes, ws.user_id
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.id = $1`,
    sessionID,
  ).Scan(&workoutID, &workoutName, &workoutDuration, &ownerID)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }
  if ownerID != userID {
    writeJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
    return
  }

  rows, err := api.DB.Query(
    `select wse.id, e.name, e.description,
            coalesce(we.sets, e.sets, 1),
            coalesce(we.reps, e.reps, '10'),
            coalesce(we.rest_seconds, e.rest_seconds, 30),
            wse.completed_sets,
            wse.completed
     from workout_session_exercises wse
     join exercises e on e.id = wse.exercise_id
     left join workout_exercises we on we.exercise_id = e.id and we.workout_id = $1
     where wse.session_id = $2
     order by wse.sort_order`,
    workoutID,
    sessionID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  exercises := []map[string]any{}
  currentIndex := -1
  totalSets := 0
  completedSets := 0
  for rows.Next() {
    var exID, name, description string
    var sets, restSeconds, completed int
    var reps string
    var isCompleted bool
    if err := rows.Scan(&exID, &name, &description, &sets, &reps, &restSeconds, &completed, &isCompleted); err != nil {
      writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
      return
    }

    if currentIndex == -1 && !isCompleted {
      currentIndex = len(exercises)
    }

    totalSets += sets
    completedSets += completed
    if isCompleted {
      completedSets += sets - completed
    }

    exercises = append(exercises, map[string]any{
      "id": exID,
      "name": name,
      "description": description,
      "sets": sets,
      "reps": reps,
      "rest": restSeconds,
      "completed_sets": completed,
      "completed": isCompleted,
    })
  }

  if len(exercises) == 0 {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "no exercises"})
    return
  }

  if currentIndex == -1 {
    writeJSON(w, http.StatusOK, map[string]any{
      "session_id": sessionID,
      "workout": map[string]any{
        "id": workoutID,
        "name": workoutName,
        "duration": workoutDuration,
      },
      "status": "completed",
    })
    return
  }

  current := exercises[currentIndex]
  currentSet := current["completed_sets"].(int) + 1

  progressPercent := 0
  if totalSets > 0 {
    progressPercent = int(float64(completedSets) / float64(totalSets) * 100)
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "session_id": sessionID,
    "workout": map[string]any{
      "id": workoutID,
      "name": workoutName,
      "duration": workoutDuration,
    },
    "exercises": exercises,
    "current_exercise": current,
    "current_index": currentIndex,
    "total_exercises": len(exercises),
    "current_set": currentSet,
    "total_sets": totalSets,
    "completed_sets": completedSets,
    "progress_percent": progressPercent,
    "status": "in_progress",
  })
}

func (api *API) WorkoutCompleteSet(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var ownerID string
  err := api.DB.QueryRow("select user_id from workout_sessions where id = $1", sessionID).Scan(&ownerID)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }
  if ownerID != userID {
    writeJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
    return
  }

  var exID string
  var sets int
  var completedSets int
  err = api.DB.QueryRow(
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
      _ = api.completeWorkoutSession(sessionID)
      writeJSON(w, http.StatusOK, map[string]any{"status": "completed"})
      return
    }
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  completedSets++
  completed := completedSets >= sets

  _, _ = api.DB.Exec(
    `update workout_session_exercises
     set completed_sets = $1, completed = $2
     where id = $3`,
    completedSets,
    completed,
    exID,
  )

  if completed {
    _, _ = api.DB.Exec(
      `update workout_sessions
       set completed_exercises = completed_exercises + 1
       where id = $1`,
      sessionID,
    )
  }

  var totalExercises int
  var completedExercises int
  _ = api.DB.QueryRow(
    `select total_exercises, completed_exercises from workout_sessions where id = $1`,
    sessionID,
  ).Scan(&totalExercises, &completedExercises)

  if totalExercises > 0 && completedExercises >= totalExercises {
    _ = api.completeWorkoutSession(sessionID)
    writeJSON(w, http.StatusOK, map[string]any{"status": "completed"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) WorkoutComplete(w http.ResponseWriter, r *http.Request) {
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }
  _ = api.completeWorkoutSession(sessionID)
  writeJSON(w, http.StatusOK, map[string]any{"status": "completed"})
}

func (api *API) WorkoutSessionSummary(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var ownerID string
  var workoutName string
  var duration int
  var totalExercises int
  var completedExercises int
  var calories int
  err := api.DB.QueryRow(
    `select ws.user_id, w.name, coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.total_exercises, 0), coalesce(ws.completed_exercises, 0), coalesce(ws.calories_burned, 0)
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.id = $1`,
    sessionID,
  ).Scan(&ownerID, &workoutName, &duration, &totalExercises, &completedExercises, &calories)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }
  if ownerID != userID {
    writeJSON(w, http.StatusForbidden, map[string]any{"error": "forbidden"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "session_id": sessionID,
    "workout_name": workoutName,
    "duration": duration,
    "total_exercises": totalExercises,
    "completed_exercises": completedExercises,
    "calories": calories,
  })
}

func (api *API) History(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  rows, err := api.DB.Query(
    `select ws.id, w.id, w.name, ws.started_at, coalesce(ws.duration_minutes, w.duration_minutes),
            coalesce(ws.completed_exercises, 0), coalesce(ws.total_exercises, 0),
            ws.completed_at is not null, coalesce(ws.calories_burned, 0), coalesce(f.rating, 0)
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     left join (
       select workout_session_id, max(rating) as rating
       from feedback
       group by workout_session_id
     ) f on f.workout_session_id = ws.id
     where ws.user_id = $1
     order by ws.started_at desc`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  sessions := []map[string]any{}
  for rows.Next() {
    var id, workoutID, workoutName string
    var startedAt time.Time
    var duration int
    var completedExercises int
    var totalExercises int
    var completed bool
    var calories int
    var rating int
    _ = rows.Scan(&id, &workoutID, &workoutName, &startedAt, &duration, &completedExercises, &totalExercises, &completed, &calories, &rating)
    sessions = append(sessions, map[string]any{
      "id": id,
      "workout_id": workoutID,
      "workout_name": workoutName,
      "date": startedAt.Format("2006-01-02"),
      "duration": duration,
      "completed_exercises": completedExercises,
      "total_exercises": totalExercises,
      "completed": completed,
      "calories": calories,
      "rating": rating,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"history": sessions})
}

func (api *API) completeWorkoutSession(sessionID string) error {
  _, err := api.DB.Exec(
    `update workout_sessions
     set completed_at = now(), duration_minutes = coalesce(duration_minutes, 30), calories_burned = coalesce(calories_burned, 250)
     where id = $1`,
    sessionID,
  )
  if err != nil {
    return err
  }

  var userID string
  _ = api.DB.QueryRow("select user_id from workout_sessions where id = $1", sessionID).Scan(&userID)
  if userID != "" {
    points := 10
    _, _ = api.DB.Exec(
      `update user_points
       set points_balance = points_balance + $1, points_total = points_total + $1, updated_at = now()
       where user_id = $2`,
      points,
      userID,
    )
    _, _ = api.DB.Exec(
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
