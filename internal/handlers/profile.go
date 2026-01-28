package handlers

import (
  "database/sql"
  "net/http"
  "time"
)

func (a *App) Profile(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var age int
  var fitnessLevel string
  var restrictions []string
  var goals []string
  _ = a.DB.QueryRow(
    `select coalesce(age, 0), coalesce(fitness_level, ''), restrictions, goals
     from user_profiles where user_id = $1`,
    user.ID,
  ).Scan(&age, &fitnessLevel, &restrictions, &goals)

  var workoutsCount int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0) from workout_sessions where user_id = $1 and completed_at is not null`,
    user.ID,
  ).Scan(&workoutsCount)

  var achievementsCount int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0) from user_achievements where user_id = $1 and unlocked = true`,
    user.ID,
  ).Scan(&achievementsCount)

  streak := computeWorkoutStreak(a.DB, user.ID)

  data := map[string]any{
    "User":             user,
    "Age":              age,
    "FitnessLevel":     fitnessLevel,
    "Restrictions":     mapRestrictions(restrictions),
    "Goals":            mapGoals(goals),
    "WorkoutsCount":    workoutsCount,
    "AchievementsCount": achievementsCount,
    "Streak":           streak,
  }

  a.renderPage(w, r, "profile", "Профиль", "profile", data)
}

func (a *App) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  name := r.FormValue("name")
  department := r.FormValue("department")
  position := r.FormValue("position")
  age := r.FormValue("age")

  if name != "" {
    _, _ = a.DB.Exec("update users set name = $1, department = $2, position = $3 where id = $4", name, department, position, user.ID)
  }
  if age != "" {
    _, _ = a.DB.Exec("update user_profiles set age = $1, updated_at = now() where user_id = $2", age, user.ID)
  }

  a.setFlash(w, "Профиль обновлен")
  http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func computeWorkoutStreak(db *sql.DB, userID string) int {
  rows, err := db.Query(
    `select date(completed_at)
     from workout_sessions
     where user_id = $1 and completed_at is not null
     order by date(completed_at) desc`,
    userID,
  )
  if err != nil {
    return 0
  }
  defer rows.Close()

  streak := 0
  var lastDate time.Time
  for rows.Next() {
    var date time.Time
    if err := rows.Scan(&date); err != nil {
      return streak
    }
    if streak == 0 {
      streak = 1
      lastDate = date
      continue
    }
    if lastDate.AddDate(0, 0, -1).Equal(date) {
      streak++
      lastDate = date
    } else {
      break
    }
  }

  return streak
}

func mapRestrictions(values []string) []string {
  labels := map[string]string{
    "back": "Проблемы со спиной",
    "joints": "Проблемы с суставами",
    "cardio": "Сердечно-сосудистые",
    "none": "Нет ограничений",
  }
  result := []string{}
  for _, value := range values {
    if label, ok := labels[value]; ok {
      result = append(result, "🔹 "+label)
    } else if value != "" {
      result = append(result, value)
    }
  }
  return result
}

func mapGoals(values []string) []string {
  labels := map[string]string{
    "rehab": "Реабилитация",
    "strength": "Увеличение силы",
    "flexibility": "Улучшение гибкости",
    "endurance": "Выносливость",
    "posture": "Коррекция осанки",
  }
  result := []string{}
  for _, value := range values {
    if label, ok := labels[value]; ok {
      result = append(result, label)
    } else if value != "" {
      result = append(result, value)
    }
  }
  return result
}
