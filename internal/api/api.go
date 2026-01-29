package api

import (
  "database/sql"
  "encoding/json"
  "net/http"
  "time"

  "github.com/go-chi/chi/v5"

  "rehab-app/internal/config"
)

type API struct {
  DB     *sql.DB
  Config config.Config
}

func New(db *sql.DB, cfg config.Config) *API {
  return &API{DB: db, Config: cfg}
}

func (api *API) Router() chi.Router {
  router := chi.NewRouter()
  router.Post("/auth/login", api.Login)
  router.Post("/auth/register", api.Register)
  router.Post("/auth/forgot", api.ForgotPassword)

  router.Group(func(r chi.Router) {
    r.Use(api.AuthMiddleware)

    r.Get("/auth/me", api.Me)
    r.Post("/auth/logout", api.Logout)

    r.Get("/profile", api.Profile)
    r.Put("/profile", api.ProfileUpdate)
    r.Get("/settings", api.Settings)
    r.Put("/settings", api.SettingsUpdate)

    r.Get("/questionnaire", api.Questionnaire)
    r.Post("/questionnaire", api.QuestionnaireSubmit)
    r.Post("/onboarding/complete", api.OnboardingComplete)

    r.Get("/program", api.Program)
    r.Get("/workouts", api.Workouts)
    r.Get("/workouts/{id}", api.WorkoutDetail)
    r.Post("/workout-sessions", api.StartWorkoutSession)
    r.Get("/workout-sessions/{id}", api.WorkoutSession)
    r.Post("/workout-sessions/{id}/complete-set", api.WorkoutCompleteSet)
    r.Post("/workout-sessions/{id}/complete", api.WorkoutComplete)
    r.Get("/workout-sessions/{id}/summary", api.WorkoutSessionSummary)
    r.Get("/history", api.History)

    r.Get("/exercises", api.Exercises)
    r.Get("/exercises/{id}", api.ExerciseDetail)

    r.Get("/progress", api.Progress)
    r.Get("/statistics", api.Statistics)
    r.Get("/achievements", api.Achievements)

    r.Route("/goals", func(gr chi.Router) {
      gr.Get("/", api.Goals)
      gr.Post("/", api.GoalsCreate)
      gr.Put("/{id}", api.GoalsUpdate)
      gr.Delete("/{id}", api.GoalsDelete)
    })

    r.Get("/notifications", api.Notifications)
    r.Post("/notifications/{id}/read", api.NotificationsRead)

    r.Get("/calendar", api.Calendar)

    r.Get("/recommendations", api.Recommendations)
    r.Get("/recommendations/{id}", api.RecommendationDetail)
    r.Post("/recommendations/{id}/bookmark", api.RecommendationBookmark)
    r.Delete("/recommendations/{id}/bookmark", api.RecommendationBookmarkRemove)
    r.Post("/recommendations/{id}/practice", api.RecommendationPractice)

    r.Get("/videos", api.Videos)

    r.Get("/medical-info", api.MedicalInfo)
    r.Put("/medical-info", api.MedicalInfoUpdate)

    r.Route("/feedback", func(fr chi.Router) {
      fr.Get("/", api.Feedback)
      fr.Post("/", api.FeedbackSubmit)
    })

    r.Route("/support", func(sr chi.Router) {
      sr.Get("/faq", api.SupportFAQ)
      sr.Get("/tickets", api.SupportTickets)
      sr.Post("/tickets", api.SupportCreate)
    })

    r.Route("/community", func(cr chi.Router) {
      cr.Get("/leaderboard", api.Leaderboard)
      cr.Get("/departments", api.Departments)
      cr.Get("/challenges", api.Challenges)
    })

    r.Route("/manager", func(mr chi.Router) {
      mr.Use(api.RequireRole("manager", "admin"))
      mr.Get("/dashboard", api.ManagerDashboard)
      mr.Get("/employees", api.ManagerEmployees)
      mr.Get("/employees/{id}", api.ManagerEmployeeDetail)
      mr.Post("/award", api.ManagerAward)
      mr.Get("/redemptions", api.ManagerRedemptions)
      mr.Post("/redemptions/{id}/approve", api.ManagerApproveRedemption)
      mr.Post("/redemptions/{id}/reject", api.ManagerRejectRedemption)
      mr.Get("/support/tickets", api.ManagerSupportTickets)
      mr.Post("/support/tickets/{id}/respond", api.ManagerSupportRespond)
    })

    r.Route("/admin", func(ar chi.Router) {
      ar.Use(api.RequireRole("admin"))
      ar.Get("/users", api.AdminUsers)
      ar.Post("/users", api.AdminUsersCreate)
      ar.Put("/users/{id}", api.AdminUsersUpdate)
      ar.Post("/users/{id}/reset-password", api.AdminUsersResetPassword)

      ar.Route("/content", func(cr chi.Router) {
        cr.Get("/exercises", api.AdminExercises)
        cr.Post("/exercises", api.AdminExercisesCreate)
        cr.Put("/exercises/{id}", api.AdminExercisesUpdate)
        cr.Delete("/exercises/{id}", api.AdminExercisesDelete)

        cr.Get("/workouts", api.AdminWorkouts)
        cr.Post("/workouts", api.AdminWorkoutsCreate)
        cr.Put("/workouts/{id}", api.AdminWorkoutsUpdate)
        cr.Delete("/workouts/{id}", api.AdminWorkoutsDelete)
        cr.Post("/workouts/{id}/exercises", api.AdminWorkoutsSetExercises)

        cr.Get("/programs", api.AdminPrograms)
        cr.Post("/programs", api.AdminProgramsCreate)
        cr.Put("/programs/{id}", api.AdminProgramsUpdate)
        cr.Delete("/programs/{id}", api.AdminProgramsDelete)
        cr.Get("/programs/{id}/workouts", api.AdminProgramsWorkouts)
        cr.Post("/programs/{id}/workouts", api.AdminProgramsSetWorkouts)

        cr.Get("/recommendations", api.AdminRecommendations)
        cr.Post("/recommendations", api.AdminRecommendationsCreate)
        cr.Put("/recommendations/{id}", api.AdminRecommendationsUpdate)
        cr.Delete("/recommendations/{id}", api.AdminRecommendationsDelete)

        cr.Get("/videos", api.AdminVideos)
        cr.Post("/videos", api.AdminVideosCreate)
        cr.Put("/videos/{id}", api.AdminVideosUpdate)
        cr.Delete("/videos/{id}", api.AdminVideosDelete)

        cr.Get("/rewards", api.AdminRewards)
        cr.Post("/rewards", api.AdminRewardsCreate)
        cr.Put("/rewards/{id}", api.AdminRewardsUpdate)
        cr.Delete("/rewards/{id}", api.AdminRewardsDelete)
      })

      ar.Get("/support/tickets", api.AdminSupportTickets)
      ar.Post("/support/tickets/{id}/respond", api.AdminSupportRespond)

      ar.Get("/redemptions", api.AdminRedemptions)
      ar.Post("/redemptions/{id}/approve", api.AdminApproveRedemption)
      ar.Post("/redemptions/{id}/reject", api.AdminRejectRedemption)
    })
  })

  return router
}

func (api *API) AuthMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    token := tokenFromRequest(r)
    if token == "" {
      writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing token"})
      return
    }

    var userID string
    err := api.DB.QueryRow(
      `select user_id from api_tokens where token = $1 and (expires_at is null or expires_at > now())`,
      token,
    ).Scan(&userID)
    if err != nil {
      writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid token"})
      return
    }

    ctx := r.Context()
    ctx = contextWithUserID(ctx, userID)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  _ = json.NewEncoder(w).Encode(payload)
}

func (api *API) createToken(userID string) (string, error) {
  token, err := randomToken(32)
  if err != nil {
    return "", err
  }
  expiresAt := time.Now().Add(api.Config.APITokenTTL)
  _, err = api.DB.Exec(
    `insert into api_tokens (user_id, token, expires_at)
     values ($1, $2, $3)`,
    userID,
    token,
    expiresAt,
  )
  if err != nil {
    return "", err
  }
  return token, nil
}
