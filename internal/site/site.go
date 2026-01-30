package site

import (
  "crypto/rand"
  "database/sql"
  "encoding/base64"
  "errors"
  "log"
  "net/http"
  "net/url"
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
  MuscleGroups []string
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
  AchievementPercent int
}

type profileView struct {
  User         models.User
  Age          int
  FitnessLevel string
  Goals        []string
  Points       int
}

type supportTicketView struct {
  ID        string
  Subject   string
  Message   string
  Status    string
  CreatedAt string
  Response  string
  EmployeeName string
  EmployeeID   string
}

type rewardView struct {
  ID          string
  Title       string
  Description string
  PointsCost  int
  Category    string
}

type managerEmployeeView struct {
  ID                 string
  Name               string
  EmployeeID         string
  Department         string
  Position           string
  Role               string
  DoctorApproval     bool
  Points             int
  AchievementsTotal  int
  AchievementsUnlocked int
}

type redemptionView struct {
  ID          string
  EmployeeName string
  Department  string
  RewardTitle string
  PointsCost  int
}

type feedbackAdminView struct {
  EmployeeName     string
  WorkoutName      string
  PerceivedExertion int
  Tolerance         int
  PainLevel         int
  Wellbeing         int
  Comment           string
  CreatedAt         string
}

type adminPlanView struct {
  UserID     string
  Name       string
  EmployeeID string
  PlanID     string
  Goal       string
  Level      string
  Status     string
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
  r.Get("/password-reset", s.passwordResetPage)
  r.Post("/password-reset", s.passwordResetSubmit)
  r.Post("/logout", s.logout)

  r.Group(func(pr chi.Router) {
    pr.Use(s.Sessions.RequireAuth)
    pr.Use(s.requirePasswordChange)
    pr.Get("/", s.dashboard)
    pr.Get("/change-password", s.changePasswordPage)
    pr.Post("/change-password", s.changePasswordSubmit)
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
    pr.Get("/rewards", s.rewardsPage)
    pr.Post("/rewards/{id}/redeem", s.rewardRedeem)
    pr.Get("/support", s.supportPage)
    pr.Post("/support", s.supportSubmit)
    pr.Get("/profile", s.profile)
    pr.Post("/profile", s.profileUpdate)

    pr.Route("/manager", func(mr chi.Router) {
      mr.Use(s.requireRoles("manager"))
      mr.Get("/", s.managerDashboard)
      mr.Get("/employees/{id}", s.managerEmployee)
      mr.Post("/employees/{id}/award", s.managerAward)
      mr.Post("/redemptions/{id}/approve", s.managerRedemptionApprove)
      mr.Post("/redemptions/{id}/reject", s.managerRedemptionReject)
    })

    pr.Route("/admin", func(ar chi.Router) {
      ar.Use(s.requireRoles("admin"))
      ar.Get("/", s.adminDashboard)
      ar.Get("/exercises", s.adminExercises)
      ar.Post("/exercises", s.adminExerciseCreate)
      ar.Post("/exercises/{id}/update", s.adminExerciseUpdate)
      ar.Get("/workouts", s.adminWorkouts)
      ar.Post("/workouts", s.adminWorkoutCreate)
      ar.Post("/workouts/{id}/update", s.adminWorkoutUpdate)
      ar.Get("/workouts/{id}", s.adminWorkoutDetail)
      ar.Post("/workouts/{id}/exercises/add", s.adminWorkoutExerciseAdd)
      ar.Post("/workouts/{id}/exercises/{exerciseId}/remove", s.adminWorkoutExerciseRemove)
      ar.Get("/programs", s.adminPrograms)
      ar.Post("/programs", s.adminProgramCreate)
      ar.Post("/programs/{id}/update", s.adminProgramUpdate)
      ar.Get("/programs/{id}", s.adminProgramDetail)
      ar.Post("/programs/{id}/workouts/add", s.adminProgramWorkoutAdd)
      ar.Post("/programs/{id}/workouts/{workoutId}/remove", s.adminProgramWorkoutRemove)
      ar.Get("/plans", s.adminPlans)
      ar.Get("/plans/{id}", s.adminPlanDetail)
      ar.Post("/plans/{id}/regenerate", s.adminPlanRegenerate)
      ar.Post("/plans/{id}/pause", s.adminPlanPause)
      ar.Post("/plans/{id}/resume", s.adminPlanResume)
      ar.Post("/plans/{id}/workouts/{planWorkoutId}/replace", s.adminPlanWorkoutReplace)
      ar.Post("/users/create", s.adminUserCreate)
      ar.Post("/users/{id}/update", s.adminUserUpdate)
      ar.Post("/users/{id}/reset-password", s.adminUserResetPassword)
      ar.Get("/feedback", s.adminFeedback)
      ar.Get("/support", s.adminSupport)
      ar.Post("/support/{id}/respond", s.adminSupportRespond)
      ar.Post("/password-requests/{id}/resolve", s.adminPasswordRequestResolve)
    })
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

func (s *Site) requireRoles(roles ...string) func(http.Handler) http.Handler {
  allowed := map[string]bool{}
  for _, role := range roles {
    allowed[role] = true
  }
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      user := middleware.UserFromContext(r.Context())
      if user == nil || !allowed[user.Role] {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
      }
      next.ServeHTTP(w, r)
    })
  }
}

func (s *Site) requirePasswordChange(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/change-password" || r.URL.Path == "/logout" {
      next.ServeHTTP(w, r)
      return
    }
    user := middleware.UserFromContext(r.Context())
    if user == nil {
      next.ServeHTTP(w, r)
      return
    }
    var temp bool
    _ = s.DB.QueryRow(`select password_temp from users where id = $1`, user.ID).Scan(&temp)
    if temp {
      http.Redirect(w, r, "/change-password", http.StatusSeeOther)
      return
    }
    next.ServeHTTP(w, r)
  })
}

func (s *Site) requireDoctorApproval(w http.ResponseWriter, r *http.Request, userID string) bool {
  if !s.loadDoctorApproval(userID) {
    http.Redirect(w, r, "/program?doctor=1", http.StatusSeeOther)
    return false
  }
  return true
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

  var tempPassword bool
  _ = s.DB.QueryRow(`select password_temp from users where id = $1`, userID).Scan(&tempPassword)

  if err := s.createSession(w, userID); err != nil {
    http.Redirect(w, r, "/login?error=Ошибка%20сессии", http.StatusSeeOther)
    return
  }

  if tempPassword {
    http.Redirect(w, r, "/change-password", http.StatusSeeOther)
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
    http.Redirect(w, r, "/register?error=ID-сотрудника%20уже%20занят", http.StatusSeeOther)
    return
  }

  _ = db.EnsureUserDefaults(s.DB, userID)
  if err := s.createSession(w, userID); err != nil {
    http.Redirect(w, r, "/register?error=Ошибка%20сессии", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) passwordResetPage(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Сброс пароля", "")
  data["HideNav"] = true
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "password_reset", data)
}

func (s *Site) passwordResetSubmit(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/password-reset?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  employeeID := strings.TrimSpace(r.FormValue("employee_id"))
  if employeeID == "" {
    http.Redirect(w, r, "/password-reset?error=Введите%20ID-сотрудника", http.StatusSeeOther)
    return
  }

  var userID string
  err := s.DB.QueryRow(`select id from users where employee_id = $1`, employeeID).Scan(&userID)
  if err == nil {
    var existing string
    err = s.DB.QueryRow(
      `select id from password_reset_requests where user_id = $1 and status = 'open'`,
      userID,
    ).Scan(&existing)
    if err != nil {
      _, _ = s.DB.Exec(
        `insert into password_reset_requests (user_id) values ($1)`,
        userID,
      )
    }
  }

  http.Redirect(w, r, "/password-reset?success=Заявка%20отправлена.%20Администратор%20свяжется%20с%20вами", http.StatusSeeOther)
}

func (s *Site) changePasswordPage(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Смена пароля", "")
  data["HideNav"] = true
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "change_password", data)
}

func (s *Site) changePasswordSubmit(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/change-password?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  password := r.FormValue("password")
  confirm := r.FormValue("password_confirm")
  if len(password) < 8 {
    http.Redirect(w, r, "/change-password?error=Минимум%208%20символов", http.StatusSeeOther)
    return
  }
  if password != confirm {
    http.Redirect(w, r, "/change-password?error=Пароли%20не%20совпадают", http.StatusSeeOther)
    return
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    http.Redirect(w, r, "/change-password?error=Ошибка%20пароля", http.StatusSeeOther)
    return
  }
  _, _ = s.DB.Exec(
    `update users set password_hash = $1, password_temp = false, updated_at = now() where id = $2`,
    string(hash),
    user.ID,
  )
  http.Redirect(w, r, "/?password_changed=1", http.StatusSeeOther)
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

  var achievementsTotal int
  var achievementsUnlocked int
  _ = s.DB.QueryRow(`select count(*) from user_achievements where user_id = $1`, user.ID).Scan(&achievementsTotal)
  _ = s.DB.QueryRow(`select count(*) from user_achievements where user_id = $1 and unlocked = true`, user.ID).Scan(&achievementsUnlocked)
  if achievementsTotal > 0 {
    stats.AchievementPercent = int(float64(achievementsUnlocked) / float64(achievementsTotal) * 100)
  }

  var nextWorkout *planWorkoutView
  var goal string
  if !setupIncomplete {
    plan, planErr := s.ensurePlan(user.ID)
    if plan != nil {
      data["PlanPaused"] = plan.Status == "paused"
      data["PlanPausedReason"] = plan.PausedReason
      data["PlanLevel"] = plan.Level
      data["PlanFrequency"] = plan.Frequency
      goal = plan.Goal
    }
    if planErr != nil {
      data["PlanError"] = "Не удалось сформировать план. Проверьте анкету."
    }

    nextWorkout, _ = s.fetchNextPlanWorkout(user.ID)
  }
  if goal == "" {
    if q, err := s.loadQuestionnaire(user.ID); err == nil {
      goal = q.Goal
    }
  }

  data["Stats"] = stats
  data["NextWorkout"] = nextWorkout
  data["DoctorApproved"] = s.loadDoctorApproval(user.ID)
  data["Goal"] = goal
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
      ex.VideoURL = normalizeVideoURL(ex.VideoURL)
      exercises = append(exercises, ex)
    }
  }

  data := s.baseData(r, workout.Name, "program")
  data["DoctorApproved"] = s.loadDoctorApproval(middleware.UserFromContext(r.Context()).ID)
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
  if !s.requireDoctorApproval(w, r, user.ID) {
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
  data["Error"] = r.URL.Query().Get("error")

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
      ex.ID = normalizeResourceID(ex.ID)
      ex.VideoURL = normalizeVideoURL(ex.VideoURL)
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
  w.Header().Set("Cache-Control", "no-store")
  _ = db.EnsureUserDefaults(s.DB, user.ID)
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
  data["CanEditProfile"] = user.Role == "admin"
  q, _ := s.loadQuestionnaire(user.ID)
  if q.SessionMinutes == 0 {
    q.SessionMinutes = sessionMinutesForLevel(resolveLevel(q.FitnessLevel))
  }
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
  if user.Role != "admin" {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }
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
  rawParam := strings.TrimSpace(chi.URLParam(r, "id"))
  exerciseID := normalizeResourceID(rawParam)
  if exerciseID == "" {
    http.Redirect(w, r, "/exercises?error=Упражнение%20не%20найдено", http.StatusSeeOther)
    return
  }

  var ex exerciseCard
  err := s.DB.QueryRow(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(rest_seconds, 0),
            coalesce(duration_seconds, 0), coalesce(muscle_groups, '{}'), coalesce(equipment, '{}'), coalesce(video_url, '')
     from exercises
     where replace(lower(id::text), '-', '') = replace($1, '-', '')`,
    exerciseID,
  ).Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
  if err != nil {
    err = s.DB.QueryRow(
      `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
              coalesce(sets, 0), coalesce(reps, ''), coalesce(rest_seconds, 0),
              coalesce(duration_seconds, 0), coalesce(muscle_groups, '{}'), coalesce(equipment, '{}'), coalesce(video_url, '')
       from exercises where lower(name) = lower($1)`,
      rawParam,
    ).Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
    if err != nil {
      http.Redirect(w, r, "/exercises?error=Упражнение%20не%20найдено", http.StatusSeeOther)
      return
    }
  }
  ex.ID = normalizeResourceID(ex.ID)
  ex.VideoURL = normalizeVideoURL(ex.VideoURL)

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
    `select p.id, p.name, p.description, coalesce(p.muscle_groups, '{}')
     from programs p where p.id = $1`,
    programID,
  ).Scan(&program.ID, &program.Name, &program.Description, &program.MuscleGroups)
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
  if !s.requireDoctorApproval(w, r, user.ID) {
    return
  }
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
    `select p.id, p.name, p.description, coalesce(p.muscle_groups, '{}')
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
    if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.MuscleGroups); err != nil {
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

func normalizeVideoURL(value string) string {
  trimmed := strings.TrimSpace(value)
  if trimmed == "" {
    return ""
  }
  if strings.Contains(trimmed, "<iframe") {
    if src := extractIframeSrc(trimmed); src != "" {
      trimmed = src
    }
  }
  if strings.Contains(trimmed, "youtube.com/embed/") {
    return trimmed
  }
  if strings.HasPrefix(trimmed, "youtu.be/") || strings.HasPrefix(trimmed, "youtube.com/") || strings.HasPrefix(trimmed, "www.youtube.com/") {
    trimmed = "https://" + trimmed
  }
  parsed, err := url.Parse(trimmed)
  if err != nil {
    return trimmed
  }
  host := strings.ToLower(parsed.Host)
  path := strings.Trim(parsed.Path, "/")
  if strings.Contains(host, "youtu.be") {
    if path != "" {
      id := strings.Split(path, "/")[0]
      if id != "" {
        return "https://www.youtube.com/embed/" + id
      }
    }
    return trimmed
  }
  if strings.Contains(host, "youtube.com") {
    if strings.HasPrefix(path, "embed/") {
      return trimmed
    }
    if strings.HasPrefix(path, "shorts/") {
      id := strings.TrimPrefix(path, "shorts/")
      if id != "" {
        return "https://www.youtube.com/embed/" + id
      }
    }
    if strings.HasPrefix(path, "live/") {
      id := strings.TrimPrefix(path, "live/")
      if id != "" {
        return "https://www.youtube.com/embed/" + id
      }
    }
    if strings.HasSuffix(path, "watch") || path == "watch" {
      vid := parsed.Query().Get("v")
      if vid != "" {
        return "https://www.youtube.com/embed/" + vid
      }
    }
  }
  return trimmed
}

func extractIframeSrc(value string) string {
  srcIndex := strings.Index(strings.ToLower(value), "src=")
  if srcIndex == -1 {
    return ""
  }
  rest := value[srcIndex+4:]
  if len(rest) == 0 {
    return ""
  }
  quote := rest[0]
  if quote != '"' && quote != '\'' {
    return ""
  }
  end := strings.IndexByte(rest[1:], quote)
  if end == -1 {
    return ""
  }
  return rest[1 : 1+end]
}

func normalizeResourceID(value string) string {
  trimmed := strings.TrimSpace(value)
  trimmed = strings.Trim(trimmed, "{}")
  trimmed = strings.ToLower(trimmed)
  return trimmed
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
