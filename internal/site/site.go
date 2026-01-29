package site

import (
  "crypto/rand"
  "database/sql"
  "encoding/base64"
  "errors"
  "log"
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/go-chi/chi/v5"
  "golang.org/x/crypto/bcrypt"

  "rehab-app/internal/config"
  "rehab-app/internal/db"
  "rehab-app/internal/middleware"
  "rehab-app/internal/models"
  "rehab-app/internal/web"
)

type Site struct {
  DB       *sql.DB
  Renderer *web.Renderer
  Sessions *middleware.SessionManager
  Config   config.Config
}

type workoutCard struct {
  ID          string
  Name        string
  Description string
  Duration    int
  Difficulty  string
  Category    string
  Exercises   int
}

type exerciseCard struct {
  ID          string
  Name        string
  Description string
  Category    string
  Difficulty  string
  Sets        int
  Reps        string
  Rest        int
  Duration    int
  MuscleGroups []string
  Equipment   []string
  VideoURL    string
}

type programCard struct {
  ID          string
  Name        string
  Description string
  Workouts    int
  Duration    int
  Difficulty  string
  Category    string
}

type sessionExercise struct {
  ID            string
  Name          string
  Description   string
  Sets          int
  Reps          string
  Rest          int
  CompletedSets int
  Completed     bool
}

type sessionView struct {
  ID              string
  WorkoutName     string
  StartedAt       string
  DurationMinutes int
  ProgressPercent int
  CurrentSet      int
  CurrentExercise *sessionExercise
  Exercises       []sessionExercise
  Completed       bool
  Feedback        *sessionFeedbackView
}

type dashboardStats struct {
  Workouts           int
  Minutes            int
  Points             int
  GoalsCompleted     int
  GoalsTotal         int
  GoalsPercent       int
  AchievementPercent int
}

type profileView struct {
  User         models.User
  Age          int
  FitnessLevel string
  Goals        []string
  Points       int
}

func New(dbConn *sql.DB, renderer *web.Renderer, sessions *middleware.SessionManager, cfg config.Config) *Site {
  return &Site{DB: dbConn, Renderer: renderer, Sessions: sessions, Config: cfg}
}

func (s *Site) Router() chi.Router {
  r := chi.NewRouter()
  r.Use(s.Sessions.Load)

  r.Get("/login", s.loginPage)
  r.Post("/login", s.loginSubmit)
  r.Get("/register", s.registerPage)
  r.Post("/register", s.registerSubmit)
  r.Post("/logout", s.logout)

  r.Group(func(pr chi.Router) {
    pr.Use(s.Sessions.RequireAuth)
    pr.Get("/", s.dashboard)
    pr.Get("/questionnaire", s.questionnairePage)
    pr.Post("/questionnaire", s.questionnaireSubmit)
    pr.Get("/program", s.planPage)
    pr.Get("/programs/{id}", s.programDetail)
    pr.Post("/programs/{id}/start", s.programStart)
    pr.Post("/plan/regenerate", s.planRegenerate)
    pr.Post("/plan-workouts/{id}/start", s.planWorkoutStart)
    pr.Post("/plan-workouts/{id}/skip", s.planWorkoutSkip)
    pr.Get("/workouts/{id}", s.workoutDetail)
    pr.Post("/workouts/{id}/start", s.startWorkout)
    pr.Get("/sessions/{id}", s.sessionDetail)
    pr.Post("/sessions/{id}/set", s.sessionCompleteSet)
    pr.Post("/sessions/{id}/complete", s.sessionComplete)
    pr.Post("/sessions/{id}/feedback", s.sessionFeedback)
    pr.Get("/exercises", s.exercises)
    pr.Get("/exercises/{id}", s.exerciseDetail)
    pr.Get("/history", s.planHistory)
    pr.Get("/leaderboard", s.leaderboard)
    pr.Get("/achievements", s.achievementsPage)
    pr.Get("/profile", s.profile)
    pr.Post("/profile", s.profileUpdate)
  })

  return r
}

func (s *Site) baseData(r *http.Request, title, active string) map[string]any {
  return map[string]any{
    "Title":  title,
    "Active": active,
    "User":   middleware.UserFromContext(r.Context()),
  }
}

func (s *Site) render(w http.ResponseWriter, name string, data map[string]any) {
  if err := s.Renderer.Render(w, name, data); err != nil {
    log.Printf("render %s: %v", name, err)
    http.Error(w, "template error", http.StatusInternalServerError)
  }
}

func (s *Site) loginPage(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Вход", "")
  data["HideNav"] = true
  data["Error"] = r.URL.Query().Get("error")
  data["AllowRegister"] = s.Config.AllowSelfRegister
  s.render(w, "login", data)
}

func (s *Site) loginSubmit(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/login?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }

  employeeID := strings.TrimSpace(r.FormValue("employee_id"))
  password := r.FormValue("password")
  if employeeID == "" || password == "" {
    http.Redirect(w, r, "/login?error=Заполните%20все%20поля", http.StatusSeeOther)
    return
  }

  var userID string
  var hash string
  err := s.DB.QueryRow(
    `select id, password_hash from users where employee_id = $1`,
    employeeID,
  ).Scan(&userID, &hash)
  if err != nil {
    http.Redirect(w, r, "/login?error=Неверные%20данные", http.StatusSeeOther)
    return
  }

  if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
    http.Redirect(w, r, "/login?error=Неверные%20данные", http.StatusSeeOther)
    return
  }

  _ = db.EnsureUserDefaults(s.DB, userID)

  if err := s.createSession(w, userID); err != nil {
    http.Redirect(w, r, "/login?error=Ошибка%20сессии", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) registerPage(w http.ResponseWriter, r *http.Request) {
  if !s.Config.AllowSelfRegister {
    http.Error(w, "registration disabled", http.StatusForbidden)
    return
  }
  data := s.baseData(r, "Регистрация", "")
  data["HideNav"] = true
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "register", data)
}

func (s *Site) registerSubmit(w http.ResponseWriter, r *http.Request) {
  if !s.Config.AllowSelfRegister {
    http.Error(w, "registration disabled", http.StatusForbidden)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/register?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }

  name := strings.TrimSpace(r.FormValue("name"))
  employeeID := strings.TrimSpace(r.FormValue("employee_id"))
  department := strings.TrimSpace(r.FormValue("department"))
  position := strings.TrimSpace(r.FormValue("position"))
  password := r.FormValue("password")

  if name == "" || employeeID == "" || password == "" {
    http.Redirect(w, r, "/register?error=Заполните%20обязательные%20поля", http.StatusSeeOther)
    return
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    http.Redirect(w, r, "/register?error=Ошибка%20пароля", http.StatusSeeOther)
    return
  }

  var userID string
  err = s.DB.QueryRow(
    `insert into users (name, employee_id, password_hash, role, department, position)
     values ($1, $2, $3, 'employee', $4, $5)
     returning id`,
    name,
    employeeID,
    string(hash),
    nullIfEmpty(department),
    nullIfEmpty(position),
  ).Scan(&userID)
  if err != nil {
    http.Redirect(w, r, "/register?error=Табельный%20номер%20уже%20занят", http.StatusSeeOther)
    return
  }

  _ = db.EnsureUserDefaults(s.DB, userID)
  if err := s.createSession(w, userID); err != nil {
    http.Redirect(w, r, "/register?error=Ошибка%20сессии", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) logout(w http.ResponseWriter, r *http.Request) {
  cookie, err := r.Cookie(s.Config.CookieName)
  if err == nil && cookie.Value != "" {
    _, _ = s.DB.Exec("delete from sessions where token = $1", cookie.Value)
  }

  http.SetCookie(w, &http.Cookie{
    Name:     s.Config.CookieName,
    Value:    "",
    Path:     "/",
    Expires:  time.Unix(0, 0),
    MaxAge:   -1,
    HttpOnly: true,
    Secure:   s.Config.CookieSecure,
    SameSite: http.SameSiteLaxMode,
  })

  http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Site) dashboard(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  data := s.baseData(r, "Главная", "dashboard")
  needsQuestionnaire := s.needsQuestionnaire(user.ID)
  setupIncomplete := needsQuestionnaire
  setupSteps := 0
  if needsQuestionnaire {
    setupSteps++
  }
  data["SetupNeedsQuestionnaire"] = needsQuestionnaire
  data["SetupIncomplete"] = setupIncomplete
  data["SetupSteps"] = setupSteps

  stats := dashboardStats{}
  _ = s.DB.QueryRow(
    `select count(*) from workout_sessions where user_id = $1 and completed_at is not null`,
    user.ID,
  ).Scan(&stats.Workouts)
  _ = s.DB.QueryRow(
    `select coalesce(sum(duration_minutes), 0) from workout_sessions where user_id = $1 and completed_at is not null`,
    user.ID,
  ).Scan(&stats.Minutes)
  _ = s.DB.QueryRow(
    `select coalesce(points_balance, 0) from user_points where user_id = $1`,
    user.ID,
  ).Scan(&stats.Points)

  var goalsTotal int
  var goalsCompleted int
  _ = s.DB.QueryRow(
    `select count(*) from goals where user_id = $1`,
    user.ID,
  ).Scan(&goalsTotal)
  _ = s.DB.QueryRow(
    `select count(*) from goals where user_id = $1 and progress >= 100`,
    user.ID,
  ).Scan(&goalsCompleted)
  stats.GoalsTotal = goalsTotal
  stats.GoalsCompleted = goalsCompleted
  if goalsTotal > 0 {
    stats.GoalsPercent = int(float64(goalsCompleted) / float64(goalsTotal) * 100)
  }

  var achievementsTotal int
  var achievementsUnlocked int
  _ = s.DB.QueryRow(`select count(*) from user_achievements where user_id = $1`, user.ID).Scan(&achievementsTotal)
  _ = s.DB.QueryRow(`select count(*) from user_achievements where user_id = $1 and unlocked = true`, user.ID).Scan(&achievementsUnlocked)
  if achievementsTotal > 0 {
    stats.AchievementPercent = int(float64(achievementsUnlocked) / float64(achievementsTotal) * 100)
  }

  var nextWorkout *planWorkoutView
  if !setupIncomplete {
    plan, planErr := s.ensurePlan(user.ID)
    if plan != nil {
      data["PlanPaused"] = plan.Status == "paused"
      data["PlanPausedReason"] = plan.PausedReason
      data["PlanLevel"] = plan.Level
      data["PlanFrequency"] = plan.Frequency
    }
    if planErr != nil {
      data["PlanError"] = "Не удалось сформировать план. Проверьте анкету."
    }

    nextWorkout, _ = s.fetchNextPlanWorkout(user.ID)
  }

  data["Stats"] = stats
  data["NextWorkout"] = nextWorkout
  s.render(w, "dashboard", data)
}

func (s *Site) workoutDetail(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    http.NotFound(w, r)
    return
  }

  var workout workoutCard
  err := s.DB.QueryRow(
    `select id, name, description, duration_minutes, difficulty, coalesce(category, '')
     from workouts where id = $1`,
    workoutID,
  ).Scan(&workout.ID, &workout.Name, &workout.Description, &workout.Duration, &workout.Difficulty, &workout.Category)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  _ = s.ensureWorkoutExercises(workoutID)

  exercises := []exerciseCard{}
  rows, err := s.DB.Query(
    `select e.id, e.name, e.description, coalesce(e.category, ''), coalesce(we.sets, e.sets, 1),
            coalesce(we.reps, e.reps, '10'), coalesce(we.rest_seconds, e.rest_seconds, 30),
            coalesce(e.duration_seconds, 0), coalesce(e.muscle_groups, '{}'), coalesce(e.equipment, '{}'), coalesce(e.video_url, '')
     from workout_exercises we
     join exercises e on e.id = we.exercise_id
     where we.workout_id = $1
     order by we.sort_order`,
    workoutID,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var ex exerciseCard
      _ = rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
      exercises = append(exercises, ex)
    }
  }

  data := s.baseData(r, workout.Name, "program")
  data["Workout"] = workout
  data["Exercises"] = exercises
  data["WorkoutEquipment"] = uniqueStrings(exercises, func(ex exerciseCard) []string { return ex.Equipment })
  data["WorkoutMuscles"] = uniqueStrings(exercises, func(ex exerciseCard) []string { return ex.MuscleGroups })
  data["WorkoutWarmup"], data["WorkoutMain"], data["WorkoutCooldown"] = splitWorkoutDuration(workout.Duration)
  data["WorkoutGuidance"] = workoutGuidance(workout.Difficulty)
  s.render(w, "workout", data)
}

func (s *Site) startWorkout(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  if s.ensureOnboarding(w, r, user.ID) {
    return
  }
  if plan, _ := s.getActivePlan(user.ID); plan != nil && plan.Status == "paused" {
    http.Redirect(w, r, "/program?paused=1", http.StatusSeeOther)
    return
  }
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    http.NotFound(w, r)
    return
  }

  sessionID, err := s.createWorkoutSession(user.ID, workoutID, "")
  if err != nil {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
}

func (s *Site) sessionDetail(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    http.NotFound(w, r)
    return
  }

  view, err := s.buildSessionView(user.ID, sessionID)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  data := s.baseData(r, "Сессия", "program")
  data["Session"] = view
  s.render(w, "session", data)
}

func (s *Site) sessionCompleteSet(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    http.NotFound(w, r)
    return
  }

  _ = s.completeSet(user.ID, sessionID)
  http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
}

func (s *Site) sessionComplete(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  sessionID := chi.URLParam(r, "id")
  if sessionID == "" {
    http.NotFound(w, r)
    return
  }

  _ = s.completeWorkoutSession(user.ID, sessionID)
  http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
}

func (s *Site) exercises(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Упражнения", "exercises")
  query := strings.TrimSpace(r.URL.Query().Get("q"))
  difficulty := strings.TrimSpace(r.URL.Query().Get("difficulty"))

  rows, err := s.DB.Query(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(rest_seconds, 0),
            coalesce(duration_seconds, 0), coalesce(muscle_groups, '{}'), coalesce(equipment, '{}'), coalesce(video_url, '')
     from exercises
     where ($1 = '' or name ilike '%' || $1 || '%')
       and ($2 = '' or difficulty ilike '%' || $2 || '%')
     order by name`,
    query,
    difficulty,
  )
  exercises := []exerciseCard{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var ex exerciseCard
      _ = rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
      exercises = append(exercises, ex)
    }
  }

  data["Exercises"] = exercises
  data["Query"] = query
  data["Difficulty"] = difficulty
  s.render(w, "exercises", data)
}

func (s *Site) profile(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  view := profileView{User: *user}

  _ = s.DB.QueryRow(
    `select coalesce(age, 0), coalesce(fitness_level, ''), goals
     from user_profiles where user_id = $1`,
    user.ID,
  ).Scan(&view.Age, &view.FitnessLevel, &view.Goals)
  _ = s.DB.QueryRow(
    `select coalesce(points_balance, 0) from user_points where user_id = $1`,
    user.ID,
  ).Scan(&view.Points)

  data := s.baseData(r, "Профиль", "profile")
  data["Profile"] = view
  data["Success"] = r.URL.Query().Get("success")
  q, _ := s.loadQuestionnaire(user.ID)
  data["Questionnaire"] = q
  data["QuestionnaireComplete"] = !s.needsQuestionnaire(user.ID)
  var restrictions []string
  var doctorApproval bool
  _ = s.DB.QueryRow(`select restrictions, doctor_approval from medical_info where user_id = $1`, user.ID).Scan(&restrictions, &doctorApproval)
  data["SelectedRestrictions"] = restrictions
  data["DoctorApproval"] = doctorApproval
  s.render(w, "profile", data)
}

func (s *Site) profileUpdate(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/profile", http.StatusSeeOther)
    return
  }

  name := strings.TrimSpace(r.FormValue("name"))
  department := strings.TrimSpace(r.FormValue("department"))
  position := strings.TrimSpace(r.FormValue("position"))
  ageStr := strings.TrimSpace(r.FormValue("age"))
  fitness := strings.TrimSpace(r.FormValue("fitness_level"))
  goalsStr := strings.TrimSpace(r.FormValue("goals"))

  if name != "" || department != "" || position != "" {
    _, _ = s.DB.Exec(
      `update users set
         name = coalesce(nullif($1, ''), name),
         department = case when $2 <> '' then $2 else department end,
         position = case when $3 <> '' then $3 else position end,
         updated_at = now()
       where id = $4`,
      name,
      department,
      position,
      user.ID,
    )
  }

  var ageValue *int
  if ageStr != "" {
    if parsed, err := strconv.Atoi(ageStr); err == nil {
      ageValue = &parsed
    }
  }

  goals := parseCSV(goalsStr)
  _, _ = s.DB.Exec(
    `update user_profiles
     set age = coalesce($1, age),
         fitness_level = coalesce(nullif($2, ''), fitness_level),
         goals = coalesce($3, goals),
         updated_at = now()
     where user_id = $4`,
    ageValue,
    nullIfEmpty(fitness),
    goals,
    user.ID,
  )

  http.Redirect(w, r, "/profile?success=Данные%20обновлены", http.StatusSeeOther)
}

func (s *Site) exerciseDetail(w http.ResponseWriter, r *http.Request) {
  exerciseID := chi.URLParam(r, "id")
  if exerciseID == "" {
    http.NotFound(w, r)
    return
  }

  var ex exerciseCard
  err := s.DB.QueryRow(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(rest_seconds, 0),
            coalesce(duration_seconds, 0), coalesce(muscle_groups, '{}'), coalesce(equipment, '{}'), coalesce(video_url, '')
     from exercises where id = $1`,
    exerciseID,
  ).Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  data := s.baseData(r, ex.Name, "exercises")
  data["Exercise"] = ex
  s.render(w, "exercise", data)
}

func (s *Site) programDetail(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  if programID == "" {
    http.NotFound(w, r)
    return
  }

  var program programCard
  err := s.DB.QueryRow(
    `select p.id, p.name, p.description
     from programs p where p.id = $1`,
    programID,
  ).Scan(&program.ID, &program.Name, &program.Description)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  workouts := []workoutCard{}
  rows, err := s.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, '')
     from program_workouts pw
     join workouts w on w.id = pw.workout_id
     where pw.program_id = $1
     order by pw.sort_order`,
    programID,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var wCard workoutCard
      _ = rows.Scan(&wCard.ID, &wCard.Name, &wCard.Description, &wCard.Duration, &wCard.Difficulty, &wCard.Category)
      workouts = append(workouts, wCard)
      program.Duration += wCard.Duration
    }
  }
  program.Workouts = len(workouts)

  data := s.baseData(r, program.Name, "program")
  data["Program"] = program
  data["ProgramWorkouts"] = workouts
  s.render(w, "program_detail", data)
}

func (s *Site) programStart(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  programID := chi.URLParam(r, "id")
  if programID == "" {
    http.NotFound(w, r)
    return
  }

  var workoutID string
  err := s.DB.QueryRow(
    `select workout_id from program_workouts where program_id = $1 order by sort_order limit 1`,
    programID,
  ).Scan(&workoutID)
  if err != nil {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  _, _ = s.DB.Exec(
    `insert into user_programs (user_id, program_id, start_date, active)
     select $1, $2, current_date, true
     where not exists (
       select 1 from user_programs where user_id = $1 and program_id = $2 and active = true
     )`,
    user.ID,
    programID,
  )

  sessionID, err := s.createWorkoutSession(user.ID, workoutID, "")
  if err != nil {
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
}
func (s *Site) createSession(w http.ResponseWriter, userID string) error {
  token, err := randomToken(32)
  if err != nil {
    return err
  }

  expires := time.Now().Add(s.Config.SessionTTL)
  _, err = s.DB.Exec(
    `insert into sessions (user_id, token, expires_at)
     values ($1, $2, $3)`,
    userID,
    token,
    expires,
  )
  if err != nil {
    return err
  }

  http.SetCookie(w, &http.Cookie{
    Name:     s.Config.CookieName,
    Value:    token,
    Path:     "/",
    Expires:  expires,
    HttpOnly: true,
    Secure:   s.Config.CookieSecure,
    SameSite: http.SameSiteLaxMode,
  })

  return nil
}

func (s *Site) ensureWorkoutExercises(workoutID string) error {
  var count int
  if err := s.DB.QueryRow("select count(*) from workout_exercises where workout_id = $1", workoutID).Scan(&count); err != nil {
    return err
  }
  if count > 0 {
    return nil
  }

  var category string
  var difficulty string
  _ = s.DB.QueryRow(
    `select coalesce(category, ''), coalesce(difficulty, '') from workouts where id = $1`,
    workoutID,
  ).Scan(&category, &difficulty)

  candidates := s.pickExercises(category, difficulty, 5)
  if len(candidates) == 0 && category != "" {
    candidates = s.pickExercises(category, "", 5)
  }
  if len(candidates) == 0 && difficulty != "" {
    candidates = s.pickExercises("", difficulty, 5)
  }
  if len(candidates) == 0 {
    candidates = s.pickExercises("", "", 5)
  }

  for i, id := range candidates {
    _, _ = s.DB.Exec(
      `insert into workout_exercises (workout_id, exercise_id, sort_order)
       values ($1, $2, $3)
       on conflict do nothing`,
      workoutID,
      id,
      i+1,
    )
  }
  return nil
}

func (s *Site) pickExercises(category, difficulty string, limit int) []string {
  rows, err := s.DB.Query(
    `select id
     from exercises
     where ($1 = '' or category = $1)
       and ($2 = '' or difficulty = $2)
     order by created_at
     limit $3`,
    category,
    difficulty,
    limit,
  )
  if err != nil {
    return nil
  }
  defer rows.Close()

  ids := []string{}
  for rows.Next() {
    var id string
    _ = rows.Scan(&id)
    ids = append(ids, id)
  }
  return ids
}

func (s *Site) fetchPrograms() []programCard {
  rows, err := s.DB.Query(
    `select p.id, p.name, p.description
     from programs p
     where p.active = true
     order by p.created_at`,
  )
  if err != nil {
    return nil
  }
  defer rows.Close()

  programs := []programCard{}
  for rows.Next() {
    var p programCard
    if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
      continue
    }
    var workouts int
    var duration int
    _ = s.DB.QueryRow(
      `select count(*), coalesce(sum(w.duration_minutes), 0)
       from program_workouts pw
       join workouts w on w.id = pw.workout_id
       where pw.program_id = $1`,
      p.ID,
    ).Scan(&workouts, &duration)
    p.Workouts = workouts
    p.Duration = duration
    programs = append(programs, p)
  }
  return programs
}

func (s *Site) buildSessionView(userID, sessionID string) (*sessionView, error) {
  var workoutID string
  var workoutName string
  var workoutDuration int
  var ownerID string
  var startedAt time.Time
  var completedAt sql.NullTime

  err := s.DB.QueryRow(
    `select ws.workout_id, w.name, w.duration_minutes, ws.user_id, ws.started_at, ws.completed_at
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.id = $1`,
    sessionID,
  ).Scan(&workoutID, &workoutName, &workoutDuration, &ownerID, &startedAt, &completedAt)
  if err != nil {
    return nil, err
  }
  if ownerID != userID {
    return nil, errors.New("forbidden")
  }

  rows, err := s.DB.Query(
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
    return nil, err
  }
  defer rows.Close()

  exercises := []sessionExercise{}
  currentIndex := -1
  totalSets := 0
  completedSets := 0

  for rows.Next() {
    var ex sessionExercise
    if err := rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Sets, &ex.Reps, &ex.Rest, &ex.CompletedSets, &ex.Completed); err != nil {
      return nil, err
    }

    if currentIndex == -1 && !ex.Completed {
      currentIndex = len(exercises)
    }

    totalSets += ex.Sets
    completedSets += ex.CompletedSets
    if ex.Completed {
      completedSets += ex.Sets - ex.CompletedSets
    }

    exercises = append(exercises, ex)
  }

  view := &sessionView{
    ID:          sessionID,
    WorkoutName: workoutName,
    StartedAt:   startedAt.Format("02.01.2006 15:04"),
    DurationMinutes: workoutDuration,
    Exercises:   exercises,
  }

  var feedback sessionFeedbackView
  err = s.DB.QueryRow(
    `select coalesce(perceived_exertion, 0), coalesce(tolerance, 0), coalesce(pain_level, 0), coalesce(wellbeing, 0), coalesce(comment, '')
     from workout_session_feedback where session_id = $1`,
    sessionID,
  ).Scan(&feedback.PerceivedExertion, &feedback.Tolerance, &feedback.PainLevel, &feedback.Wellbeing, &feedback.Comment)
  if err == nil {
    view.Feedback = &feedback
  }

  if completedAt.Valid {
    view.Completed = true
    view.ProgressPercent = 100
    return view, nil
  }

  if currentIndex == -1 {
    view.Completed = true
    view.ProgressPercent = 100
    return view, nil
  }

  progress := 0
  if totalSets > 0 {
    progress = int(float64(completedSets) / float64(totalSets) * 100)
  }

  current := exercises[currentIndex]
  view.CurrentExercise = &current
  view.CurrentSet = current.CompletedSets + 1
  view.ProgressPercent = progress
  return view, nil
}

func (s *Site) completeSet(userID, sessionID string) error {
  var ownerID string
  err := s.DB.QueryRow("select user_id from workout_sessions where id = $1", sessionID).Scan(&ownerID)
  if err != nil {
    return err
  }
  if ownerID != userID {
    return errors.New("forbidden")
  }

  var exID string
  var sets int
  var completedSets int
  err = s.DB.QueryRow(
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
    if errors.Is(err, sql.ErrNoRows) {
      return s.completeWorkoutSession(userID, sessionID)
    }
    return err
  }

  completedSets++
  completed := completedSets >= sets
  _, _ = s.DB.Exec(
    `update workout_session_exercises set completed_sets = $1, completed = $2 where id = $3`,
    completedSets,
    completed,
    exID,
  )

  if completed {
    _, _ = s.DB.Exec(
      `update workout_sessions set completed_exercises = completed_exercises + 1 where id = $1`,
      sessionID,
    )
  }

  var totalExercises int
  var completedExercises int
  _ = s.DB.QueryRow(
    `select total_exercises, completed_exercises from workout_sessions where id = $1`,
    sessionID,
  ).Scan(&totalExercises, &completedExercises)

  if totalExercises > 0 && completedExercises >= totalExercises {
    return s.completeWorkoutSession(userID, sessionID)
  }

  return nil
}

func (s *Site) completeWorkoutSession(userID, sessionID string) error {
  var ownerID string
  var completedAt sql.NullTime
  _ = s.DB.QueryRow("select user_id, completed_at from workout_sessions where id = $1", sessionID).Scan(&ownerID, &completedAt)
  if ownerID != userID {
    return errors.New("forbidden")
  }
  if completedAt.Valid {
    return nil
  }

  _, err := s.DB.Exec(
    `update workout_sessions
     set completed_at = now(), duration_minutes = coalesce(duration_minutes, 30), calories_burned = coalesce(calories_burned, 250)
     where id = $1`,
    sessionID,
  )
  if err != nil {
    return err
  }

  var planWorkoutID sql.NullString
  _ = s.DB.QueryRow(
    `select plan_workout_id from workout_sessions where id = $1`,
    sessionID,
  ).Scan(&planWorkoutID)
  if planWorkoutID.Valid {
    _, _ = s.DB.Exec(`update training_plan_workouts set status = 'completed' where id = $1`, planWorkoutID.String)
  }

  points := 10
  _, _ = s.DB.Exec(
    `update user_points
     set points_balance = points_balance + $1, points_total = points_total + $1, updated_at = now()
     where user_id = $2`,
    points,
    userID,
  )
  _, _ = s.DB.Exec(
    `insert into notifications (user_id, title, message, type)
     values ($1, $2, $3, $4)`,
    userID,
    "Начислены баллы",
    "Вы получили 10 баллов за завершение тренировки",
    "success",
  )

  s.updateAchievements(userID)
  return nil
}

func randomToken(length int) (string, error) {
  buf := make([]byte, length)
  if _, err := rand.Read(buf); err != nil {
    return "", err
  }
  return base64.RawURLEncoding.EncodeToString(buf), nil
}

func nullIfEmpty(value string) any {
  if strings.TrimSpace(value) == "" {
    return nil
  }
  return value
}

func parseCSV(value string) []string {
  if strings.TrimSpace(value) == "" {
    return []string{}
  }
  raw := strings.Split(value, ",")
  out := make([]string, 0, len(raw))
  for _, item := range raw {
    cleaned := strings.TrimSpace(item)
    if cleaned != "" {
      out = append(out, cleaned)
    }
  }
  return out
}

func uniqueStrings(exercises []exerciseCard, selector func(exerciseCard) []string) []string {
  seen := map[string]bool{}
  out := []string{}
  for _, ex := range exercises {
    for _, value := range selector(ex) {
      cleaned := strings.TrimSpace(value)
      if cleaned == "" {
        continue
      }
      if !seen[cleaned] {
        seen[cleaned] = true
        out = append(out, cleaned)
      }
    }
  }
  return out
}

func splitWorkoutDuration(total int) (int, int, int) {
  if total <= 0 {
    total = 30
  }
  warmup := total / 6
  if warmup < 5 {
    warmup = 5
  }
  if warmup > 10 {
    warmup = 10
  }
  cooldown := warmup
  main := total - warmup - cooldown
  if main < 10 {
    main = 10
    remaining := total - main
    if remaining < 0 {
      remaining = 0
    }
    warmup = remaining / 2
    cooldown = remaining - warmup
  }
  if main < 0 {
    main = 0
  }
  return warmup, main, cooldown
}

func workoutGuidance(difficulty string) []string {
  switch strings.ToLower(strings.TrimSpace(difficulty)) {
  case "средняя":
    return []string{
      "Сохраняйте ровный темп, контролируйте дыхание.",
      "Если появляется боль выше 3/5 — остановитесь и отметьте в дневнике.",
      "После завершения заполните обратную связь для адаптации плана.",
    }
  case "сложная":
    return []string{
      "Работайте в технике, избегайте рывков.",
      "При дискомфорте снизьте интенсивность и зафиксируйте симптомы.",
      "После тренировки заполните обратную связь для безопасной корректировки.",
    }
  default:
    return []string{
      "Двигайтесь плавно, без резких ускорений.",
      "Если почувствовали боль или ухудшение самочувствия — остановитесь.",
      "Заполните обратную связь после тренировки для корректировки плана.",
    }
  }
}
