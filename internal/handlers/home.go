package handlers

import (
  "database/sql"
  "net/http"

  "rehab-app/internal/models"
)

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  profile := models.UserProfile{}
  _ = a.DB.QueryRow(
    `select coalesce(age, 0), coalesce(fitness_level, ''), restrictions, goals, onboarding_complete
     from user_profiles where user_id = $1`,
    user.ID,
  ).Scan(&profile.Age, &profile.FitnessLevel, &profile.Restrictions, &profile.Goals, &profile.OnboardingComplete)

  needsAssessment := profile.FitnessLevel == ""

  var workoutsCount int
  var totalMinutes int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where user_id = $1 and completed_at is not null`,
    user.ID,
  ).Scan(&workoutsCount, &totalMinutes)

  var achievementsCount int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0) from user_achievements where user_id = $1 and unlocked = true`,
    user.ID,
  ).Scan(&achievementsCount)

  var nextWorkout models.Workout
  err := a.DB.QueryRow(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, w.category
     from workouts w
     order by w.created_at
     limit 1`,
  ).Scan(&nextWorkout.ID, &nextWorkout.Name, &nextWorkout.Description, &nextWorkout.Duration, &nextWorkout.Difficulty, &nextWorkout.Category)
  if err == sql.ErrNoRows {
    nextWorkout = models.Workout{}
  }

  recommendations := []models.Recommendation{}
  rows, _ := a.DB.Query(
    `select id, title, body, coalesce(category, '')
     from recommendations
     order by created_at desc
     limit 3`,
  )
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var rec models.Recommendation
      _ = rows.Scan(&rec.ID, &rec.Title, &rec.Body, &rec.Category)
      recommendations = append(recommendations, rec)
    }
  }

  data := map[string]any{
    "User":            user,
    "NeedsAssessment": needsAssessment,
    "WorkoutsCount":   workoutsCount,
    "HoursTotal":      float64(totalMinutes) / 60.0,
    "AchievementsCount": achievementsCount,
    "NextWorkout":     nextWorkout,
    "Recommendations": recommendations,
  }

  a.renderPage(w, r, "home", "Главная", "home", data)
}
