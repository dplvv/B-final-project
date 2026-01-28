package api

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "time"

  "golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
  EmployeeID string `json:"employee_id"`
  Password   string `json:"password"`
}

func (api *API) Login(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
    return
  }

  var req loginRequest
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  var userID string
  var hash string
  var name string
  var role string
  var department string
  var position string
  err := api.DB.QueryRow(
    `select id, password_hash, name, role, coalesce(department, ''), coalesce(position, '')
     from users where employee_id = $1`,
    req.EmployeeID,
  ).Scan(&userID, &hash, &name, &role, &department, &position)
  if err != nil {
    if err == sql.ErrNoRows {
      writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
      return
    }
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
    writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
    return
  }

  token, err := api.createToken(userID)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  var fitnessLevel string
  var onboardingComplete bool
  _ = api.DB.QueryRow(
    `select coalesce(fitness_level, ''), onboarding_complete
     from user_profiles where user_id = $1`,
    userID,
  ).Scan(&fitnessLevel, &onboardingComplete)

  writeJSON(w, http.StatusOK, map[string]any{
    "token": token,
    "user": map[string]any{
      "id": userID,
      "name": name,
      "employee_id": req.EmployeeID,
      "role": role,
      "department": department,
      "position": position,
      "fitness_level": fitnessLevel,
      "onboarding_complete": onboardingComplete,
    },
  })
}

func (api *API) Profile(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  var name, employeeID, role, department, position string
  _ = api.DB.QueryRow(
    `select name, employee_id, role, coalesce(department, ''), coalesce(position, '')
     from users where id = $1`,
    userID,
  ).Scan(&name, &employeeID, &role, &department, &position)

  var age int
  var fitnessLevel string
  var restrictions []string
  var goals []string
  var onboardingComplete bool
  _ = api.DB.QueryRow(
    `select coalesce(age, 0), coalesce(fitness_level, ''), restrictions, goals, onboarding_complete
     from user_profiles where user_id = $1`,
    userID,
  ).Scan(&age, &fitnessLevel, &restrictions, &goals, &onboardingComplete)

  var points int
  _ = api.DB.QueryRow("select points_balance from user_points where user_id = $1", userID).Scan(&points)

  writeJSON(w, http.StatusOK, map[string]any{
    "user": map[string]any{
      "id": userID,
      "name": name,
      "employee_id": employeeID,
      "role": role,
      "department": department,
      "position": position,
      "points": points,
    },
    "profile": map[string]any{
      "age": age,
      "fitness_level": fitnessLevel,
      "restrictions": restrictions,
      "goals": goals,
      "onboarding_complete": onboardingComplete,
    },
  })
}

func (api *API) Workouts(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, '')
     from workouts w
     order by w.created_at`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  workouts := []map[string]any{}
  for rows.Next() {
    var id, name, description, difficulty, category string
    var duration int
    _ = rows.Scan(&id, &name, &description, &duration, &difficulty, &category)
    workouts = append(workouts, map[string]any{
      "id": id,
      "name": name,
      "description": description,
      "duration": duration,
      "difficulty": difficulty,
      "category": category,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"workouts": workouts})
}

func (api *API) Exercises(w http.ResponseWriter, r *http.Request) {
  query := r.URL.Query().Get("q")
  category := r.URL.Query().Get("category")
  difficulty := r.URL.Query().Get("difficulty")

  rows, err := api.DB.Query(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(duration_seconds, 0), coalesce(rest_seconds, 0),
            muscle_groups, equipment, coalesce(video_url, '')
     from exercises
     where ($1 = '' or name ilike '%' || $1 || '%')
       and ($2 = '' or category ilike '%' || $2 || '%')
       and ($3 = '' or difficulty ilike '%' || $3 || '%')
     order by name`,
    query,
    category,
    difficulty,
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

func (api *API) Progress(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  var workouts int
  var minutes int
  _ = api.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where user_id = $1 and completed_at is not null`,
    userID,
  ).Scan(&workouts, &minutes)

  streak := computeWorkoutStreak(api.DB, userID)
  achievements := 0
  _ = api.DB.QueryRow(
    `select coalesce(count(*), 0)
     from user_achievements
     where user_id = $1 and unlocked = true`,
    userID,
  ).Scan(&achievements)

  writeJSON(w, http.StatusOK, map[string]any{
    "workouts": workouts,
    "hours": float64(minutes) / 60.0,
    "streak": streak,
    "achievements": achievements,
  })
}

func (api *API) Notifications(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  rows, err := api.DB.Query(
    `select id, title, message, type, created_at, read_at
     from notifications where user_id = $1
     order by created_at desc`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  notifications := []map[string]any{}
  for rows.Next() {
    var id, title, message, ntype string
    var created time.Time
    var readAt *time.Time
    _ = rows.Scan(&id, &title, &message, &ntype, &created, &readAt)
    notifications = append(notifications, map[string]any{
      "id": id,
      "title": title,
      "message": message,
      "type": ntype,
      "created_at": created.Format(time.RFC3339),
      "read": readAt != nil,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"notifications": notifications})
}

func (api *API) Leaderboard(w http.ResponseWriter, r *http.Request) {
  rows, err := api.DB.Query(
    `select u.id, u.name, coalesce(u.department, ''), coalesce(count(ws.id), 0), coalesce(sum(ws.duration_minutes), 0), coalesce(up.points_balance, 0)
     from users u
     left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
     left join user_points up on up.user_id = u.id
     left join user_settings us on us.user_id = u.id
     where u.role = 'employee' and coalesce(us.show_in_leaderboard, true) = true
     group by u.id, up.points_balance
     order by up.points_balance desc
     limit 10`,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  leaderboard := []map[string]any{}
  for rows.Next() {
    var id, name, department string
    var workouts, minutes, points int
    _ = rows.Scan(&id, &name, &department, &workouts, &minutes, &points)
    leaderboard = append(leaderboard, map[string]any{
      "id": id,
      "name": name,
      "department": department,
      "workouts": workouts,
      "hours": float64(minutes) / 60.0,
      "points": points,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"leaderboard": leaderboard})
}
