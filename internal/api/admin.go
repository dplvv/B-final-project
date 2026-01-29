package api

import (
  "crypto/rand"
  "encoding/base64"
  "net/http"

  "github.com/go-chi/chi/v5"
  "golang.org/x/crypto/bcrypt"

  "rehab-app/internal/db"
)

type adminUserRequest struct {
  Name       string `json:"name"`
  EmployeeID string `json:"employee_id"`
  Department string `json:"department"`
  Position   string `json:"position"`
  Role       string `json:"role"`
  Password   string `json:"password"`
}

type exercisePayload struct {
  Name         string   `json:"name"`
  Description  string   `json:"description"`
  Category     string   `json:"category"`
  Difficulty   string   `json:"difficulty"`
  Sets         int      `json:"sets"`
  Reps         string   `json:"reps"`
  Duration     int      `json:"duration_seconds"`
  RestSeconds  int      `json:"rest_seconds"`
  MuscleGroups []string `json:"muscle_groups"`
  Equipment    []string `json:"equipment"`
  VideoURL     string   `json:"video_url"`
}

type workoutPayload struct {
  Name        string `json:"name"`
  Description string `json:"description"`
  Duration    int    `json:"duration_minutes"`
  Difficulty  string `json:"difficulty"`
  Category    string `json:"category"`
}

type workoutExercisePayload struct {
  ExerciseID string `json:"exercise_id"`
  SortOrder  int    `json:"sort_order"`
  Sets       int    `json:"sets"`
  Reps       string `json:"reps"`
  Duration   int    `json:"duration_seconds"`
  Rest       int    `json:"rest_seconds"`
}

type programPayload struct {
  Name        string `json:"name"`
  Description string `json:"description"`
  Active      bool   `json:"active"`
}

type programWorkoutPayload struct {
  WorkoutID string `json:"workout_id"`
  SortOrder int    `json:"sort_order"`
}

type recommendationPayload struct {
  Title    string `json:"title"`
  Body     string `json:"body"`
  Category string `json:"category"`
  Icon     string `json:"icon"`
  Excerpt  string `json:"excerpt"`
  ReadTime int    `json:"read_time"`
}

type videoPayload struct {
  Title       string `json:"title"`
  Description string `json:"description"`
  Duration    int    `json:"duration_minutes"`
  Category    string `json:"category"`
  Difficulty  string `json:"difficulty"`
  URL         string `json:"url"`
}

type rewardPayload struct {
  Title       string `json:"title"`
  Description string `json:"description"`
  PointsCost  int    `json:"points_cost"`
  Category    string `json:"category"`
  Active      bool   `json:"active"`
}

func (api *API) AdminUsers(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select u.id, u.name, u.employee_id, u.role, coalesce(u.department, ''), coalesce(u.position, ''),
            coalesce(up.points_balance, 0)
     from users u
     left join user_points up on up.user_id = u.id
     order by u.created_at desc`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  users := []map[string]any{}
  for rows.Next() {
    var id, name, employeeID, role, department, position string
    var points int
    _ = rows.Scan(&id, &name, &employeeID, &role, &department, &position, &points)
    users = append(users, map[string]any{
      "id": id,
      "name": name,
      "employee_id": employeeID,
      "role": role,
      "department": department,
      "position": position,
      "points": points,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

func (api *API) AdminUsersCreate(w http.ResponseWriter, r *http.Request) {
  var req adminUserRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Name == "" || req.EmployeeID == "" || req.Password == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }
  if req.Role == "" {
    req.Role = "employee"
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  var id string
  err = api.DB.QueryRow(
    `insert into users (name, employee_id, password_hash, role, department, position)
     values ($1, $2, $3, $4, $5, $6)
     returning id`,
    req.Name,
    req.EmployeeID,
    string(hash),
    req.Role,
    req.Department,
    req.Position,
  ).Scan(&id)
  if err != nil {
    if isUniqueViolation(err) {
      writeJSON(w, http.StatusConflict, map[string]any{"error": "employee_id already exists"})
      return
    }
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  _ = db.EnsureUserDefaults(api.DB, id)

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminUsersUpdate(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req adminUserRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if req.Role == "" {
    req.Role = "employee"
  }

  _, err := api.DB.Exec(
    `update users
     set name = coalesce(nullif($1, ''), name),
         department = case when $2 <> '' then $2 else department end,
         position = case when $3 <> '' then $3 else position end,
         role = $4,
         updated_at = now()
     where id = $5`,
    req.Name,
    req.Department,
    req.Position,
    req.Role,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  if req.Password != "" {
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err == nil {
      _, _ = api.DB.Exec("update users set password_hash = $1 where id = $2", string(hash), userID)
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminUsersResetPassword(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req adminUserRequest
  _ = decodeJSON(r, &req)
  password := req.Password
  if password == "" {
    password = randomPassword(10)
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  _, err = api.DB.Exec("update users set password_hash = $1 where id = $2", string(hash), userID)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"password": password})
}

func (api *API) AdminExercises(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(duration_seconds, 0), coalesce(rest_seconds, 0),
            muscle_groups, equipment, coalesce(video_url, '')
     from exercises
     order by name`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  exercises := []map[string]any{}
  for rows.Next() {
    var id, name, description, category, difficulty, reps, videoURL string
    var sets, duration, rest int
    var muscleGroups []string
    var equipment []string
    _ = rows.Scan(&id, &name, &description, &category, &difficulty, &sets, &reps, &duration, &rest, &muscleGroups, &equipment, &videoURL)
    exercises = append(exercises, map[string]any{
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
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"exercises": exercises})
}

func (api *API) AdminExercisesCreate(w http.ResponseWriter, r *http.Request) {
  var req exercisePayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Name == "" || req.Description == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  var id string
  err := api.DB.QueryRow(
    `insert into exercises (name, description, category, difficulty, sets, reps, duration_seconds, rest_seconds, muscle_groups, equipment, video_url)
     values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
     returning id`,
    req.Name,
    req.Description,
    req.Category,
    req.Difficulty,
    req.Sets,
    req.Reps,
    req.Duration,
    req.RestSeconds,
    req.MuscleGroups,
    req.Equipment,
    req.VideoURL,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminExercisesUpdate(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req exercisePayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update exercises
     set name = coalesce(nullif($1, ''), name),
         description = coalesce(nullif($2, ''), description),
         category = $3,
         difficulty = $4,
         sets = $5,
         reps = $6,
         duration_seconds = $7,
         rest_seconds = $8,
         muscle_groups = $9,
         equipment = $10,
         video_url = $11
     where id = $12`,
    req.Name,
    req.Description,
    req.Category,
    req.Difficulty,
    req.Sets,
    req.Reps,
    req.Duration,
    req.RestSeconds,
    req.MuscleGroups,
    req.Equipment,
    req.VideoURL,
    id,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminExercisesDelete(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from exercises where id = $1", id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (api *API) AdminWorkouts(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, ''),
            coalesce(count(we.exercise_id), 0)
     from workouts w
     left join workout_exercises we on we.workout_id = w.id
     group by w.id
     order by w.created_at desc`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  workouts := []map[string]any{}
  for rows.Next() {
    var id, name, description, difficulty, category string
    var duration, exercises int
    _ = rows.Scan(&id, &name, &description, &duration, &difficulty, &category, &exercises)
    workouts = append(workouts, map[string]any{
      "id": id,
      "name": name,
      "description": description,
      "duration": duration,
      "difficulty": difficulty,
      "category": category,
      "exercises": exercises,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"workouts": workouts})
}

func (api *API) AdminWorkoutsCreate(w http.ResponseWriter, r *http.Request) {
  var req workoutPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Name == "" || req.Description == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  var id string
  err := api.DB.QueryRow(
    `insert into workouts (name, description, duration_minutes, difficulty, category)
     values ($1, $2, $3, $4, $5)
     returning id`,
    req.Name,
    req.Description,
    req.Duration,
    req.Difficulty,
    req.Category,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminWorkoutsUpdate(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req workoutPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update workouts
     set name = coalesce(nullif($1, ''), name),
         description = coalesce(nullif($2, ''), description),
         duration_minutes = $3,
         difficulty = $4,
         category = $5
     where id = $6`,
    req.Name,
    req.Description,
    req.Duration,
    req.Difficulty,
    req.Category,
    id,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminWorkoutsDelete(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from workouts where id = $1", id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (api *API) AdminWorkoutsSetExercises(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var items []workoutExercisePayload
  if err := decodeJSON(r, &items); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, _ = api.DB.Exec("delete from workout_exercises where workout_id = $1", workoutID)
  for _, item := range items {
    if item.ExerciseID == "" {
      continue
    }
    _, _ = api.DB.Exec(
      `insert into workout_exercises (workout_id, exercise_id, sort_order, sets, reps, duration_seconds, rest_seconds)
       values ($1, $2, $3, nullif($4, 0), nullif($5, ''), nullif($6, 0), nullif($7, 0))`,
      workoutID,
      item.ExerciseID,
      item.SortOrder,
      item.Sets,
      item.Reps,
      item.Duration,
      item.Rest,
    )
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminPrograms(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select p.id, p.name, p.description, p.active, coalesce(count(pw.workout_id), 0)
     from programs p
     left join program_workouts pw on pw.program_id = p.id
     group by p.id
     order by p.created_at desc`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  programs := []map[string]any{}
  for rows.Next() {
    var id, name, description string
    var active bool
    var workouts int
    _ = rows.Scan(&id, &name, &description, &active, &workouts)
    programs = append(programs, map[string]any{
      "id": id,
      "name": name,
      "description": description,
      "active": active,
      "workouts": workouts,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"programs": programs})
}

func (api *API) AdminProgramsCreate(w http.ResponseWriter, r *http.Request) {
  var req programPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Name == "" || req.Description == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  var id string
  err := api.DB.QueryRow(
    `insert into programs (name, description, active)
     values ($1, $2, $3)
     returning id`,
    req.Name,
    req.Description,
    req.Active,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminProgramsUpdate(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req programPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update programs
     set name = coalesce(nullif($1, ''), name),
         description = coalesce(nullif($2, ''), description),
         active = $3
     where id = $4`,
    req.Name,
    req.Description,
    req.Active,
    id,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminProgramsDelete(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from programs where id = $1", id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (api *API) AdminProgramsWorkouts(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  if programID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  rows, err := api.DB.Query(
    `select pw.workout_id, pw.sort_order, w.name
     from program_workouts pw
     join workouts w on w.id = pw.workout_id
     where pw.program_id = $1
     order by pw.sort_order`,
    programID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  workouts := []map[string]any{}
  for rows.Next() {
    var workoutID, name string
    var sortOrder int
    _ = rows.Scan(&workoutID, &sortOrder, &name)
    workouts = append(workouts, map[string]any{
      "workout_id": workoutID,
      "sort_order": sortOrder,
      "name": name,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"workouts": workouts})
}

func (api *API) AdminProgramsSetWorkouts(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  if programID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var items []programWorkoutPayload
  if err := decodeJSON(r, &items); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, _ = api.DB.Exec("delete from program_workouts where program_id = $1", programID)
  for _, item := range items {
    if item.WorkoutID == "" {
      continue
    }
    _, _ = api.DB.Exec(
      `insert into program_workouts (program_id, workout_id, sort_order)
       values ($1, $2, $3)`,
      programID,
      item.WorkoutID,
      item.SortOrder,
    )
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminRecommendations(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select id, title, category, coalesce(read_time, 5), coalesce(icon, ''), coalesce(excerpt, ''), body
     from recommendations
     order by created_at desc`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  items := []map[string]any{}
  for rows.Next() {
    var id, title, category, icon, excerpt, body string
    var readTime int
    _ = rows.Scan(&id, &title, &category, &readTime, &icon, &excerpt, &body)
    items = append(items, map[string]any{
      "id": id,
      "title": title,
      "category": category,
      "read_time": readTime,
      "icon": icon,
      "excerpt": excerpt,
      "body": body,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"recommendations": items})
}

func (api *API) AdminRecommendationsCreate(w http.ResponseWriter, r *http.Request) {
  var req recommendationPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Title == "" || req.Body == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  var id string
  err := api.DB.QueryRow(
    `insert into recommendations (title, body, category, icon, excerpt, read_time)
     values ($1, $2, $3, $4, $5, $6)
     returning id`,
    req.Title,
    req.Body,
    req.Category,
    req.Icon,
    req.Excerpt,
    req.ReadTime,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminRecommendationsUpdate(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req recommendationPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update recommendations
     set title = coalesce(nullif($1, ''), title),
         body = coalesce(nullif($2, ''), body),
         category = $3,
         icon = $4,
         excerpt = $5,
         read_time = $6
     where id = $7`,
    req.Title,
    req.Body,
    req.Category,
    req.Icon,
    req.Excerpt,
    req.ReadTime,
    id,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminRecommendationsDelete(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from recommendations where id = $1", id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (api *API) AdminVideos(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select id, title, description, coalesce(duration_minutes, 0), coalesce(category, ''), coalesce(difficulty, ''), coalesce(url, '')
     from video_tutorials
     order by created_at desc`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  items := []map[string]any{}
  for rows.Next() {
    var id, title, description, category, difficulty, url string
    var duration int
    _ = rows.Scan(&id, &title, &description, &duration, &category, &difficulty, &url)
    items = append(items, map[string]any{
      "id": id,
      "title": title,
      "description": description,
      "duration": duration,
      "category": category,
      "difficulty": difficulty,
      "url": url,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"videos": items})
}

func (api *API) AdminVideosCreate(w http.ResponseWriter, r *http.Request) {
  var req videoPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Title == "" || req.Description == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  var id string
  err := api.DB.QueryRow(
    `insert into video_tutorials (title, description, duration_minutes, category, difficulty, url)
     values ($1, $2, $3, $4, $5, $6)
     returning id`,
    req.Title,
    req.Description,
    req.Duration,
    req.Category,
    req.Difficulty,
    req.URL,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminVideosUpdate(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req videoPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update video_tutorials
     set title = coalesce(nullif($1, ''), title),
         description = coalesce(nullif($2, ''), description),
         duration_minutes = $3,
         category = $4,
         difficulty = $5,
         url = $6
     where id = $7`,
    req.Title,
    req.Description,
    req.Duration,
    req.Category,
    req.Difficulty,
    req.URL,
    id,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminVideosDelete(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from video_tutorials where id = $1", id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}


func (api *API) AdminRewards(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select id, title, description, points_cost, coalesce(category, ''), active
     from rewards
     order by points_cost`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  items := []map[string]any{}
  for rows.Next() {
    var id, title, description, category string
    var points int
    var active bool
    _ = rows.Scan(&id, &title, &description, &points, &category, &active)
    items = append(items, map[string]any{
      "id": id,
      "title": title,
      "description": description,
      "points_cost": points,
      "category": category,
      "active": active,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"rewards": items})
}

func (api *API) AdminRewardsCreate(w http.ResponseWriter, r *http.Request) {
  var req rewardPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Title == "" || req.Description == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  var id string
  err := api.DB.QueryRow(
    `insert into rewards (title, description, points_cost, category, active)
     values ($1, $2, $3, $4, $5)
     returning id`,
    req.Title,
    req.Description,
    req.PointsCost,
    req.Category,
    req.Active,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) AdminRewardsUpdate(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req rewardPayload
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update rewards
     set title = coalesce(nullif($1, ''), title),
         description = coalesce(nullif($2, ''), description),
         points_cost = $3,
         category = $4,
         active = $5
     where id = $6`,
    req.Title,
    req.Description,
    req.PointsCost,
    req.Category,
    req.Active,
    id,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (api *API) AdminRewardsDelete(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec("delete from rewards where id = $1", id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (api *API) AdminSupportTickets(w http.ResponseWriter, r *http.Request) {
  api.ManagerSupportTickets(w, r)
}

func (api *API) AdminSupportRespond(w http.ResponseWriter, r *http.Request) {
  api.ManagerSupportRespond(w, r)
}

func (api *API) AdminRedemptions(w http.ResponseWriter, r *http.Request) {
  api.ManagerRedemptions(w, r)
}

func (api *API) AdminApproveRedemption(w http.ResponseWriter, r *http.Request) {
  api.ManagerApproveRedemption(w, r)
}

func (api *API) AdminRejectRedemption(w http.ResponseWriter, r *http.Request) {
  api.ManagerRejectRedemption(w, r)
}

func randomPassword(size int) string {
  buf := make([]byte, size)
  if _, err := rand.Read(buf); err != nil {
    return "password"
  }
  return base64.RawURLEncoding.EncodeToString(buf)[:size]
}
