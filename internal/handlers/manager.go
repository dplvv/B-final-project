package handlers

import (
  "net/http"

  "github.com/go-chi/chi/v5"
  "rehab-app/internal/models"
)

func (a *App) ManagerDashboard(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil || (user.Role != "manager" && user.Role != "admin") {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }

  var employeesCount int
  var totalWorkouts int
  var totalMinutes int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0) from users where role = 'employee'`,
  ).Scan(&employeesCount)
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where completed_at is not null`,
  ).Scan(&totalWorkouts, &totalMinutes)

  employees := []models.EmployeeStats{}
  rows, _ := a.DB.Query(
    `select u.id, u.name, coalesce(u.department, ''),
            coalesce(count(ws.id), 0), coalesce(sum(ws.duration_minutes), 0), coalesce(up.points_balance, 0)
     from users u
     left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
     left join user_points up on up.user_id = u.id
     where u.role = 'employee'
     group by u.id, up.points_balance
     order by count(ws.id) desc
     limit 20`,
  )
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var stat models.EmployeeStats
      var minutes int
      _ = rows.Scan(&stat.UserID, &stat.Name, &stat.Department, &stat.WorkoutsCount, &minutes, &stat.Points)
      stat.HoursTotal = float64(minutes) / 60.0
      employees = append(employees, stat)
    }
  }

  data := map[string]any{
    "Employees": employees,
    "EmployeesCount": employeesCount,
    "TotalWorkouts": totalWorkouts,
    "TotalHours": float64(totalMinutes) / 60.0,
  }

  a.renderPage(w, r, "manager_dashboard", "Панель руководителя", "", data)
}

func (a *App) ManagerEmployee(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil || (user.Role != "manager" && user.Role != "admin") {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }

  employeeID := chi.URLParam(r, "id")
  if employeeID == "" {
    http.Redirect(w, r, "/manager", http.StatusSeeOther)
    return
  }

  var employee models.User
  err := a.DB.QueryRow(
    `select id, name, employee_id, role, coalesce(department, ''), coalesce(position, '')
     from users where id = $1`,
    employeeID,
  ).Scan(&employee.ID, &employee.Name, &employee.EmployeeID, &employee.Role, &employee.Department, &employee.Position)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  var workoutsCount int
  var minutes int
  var points int
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where user_id = $1 and completed_at is not null`,
    employeeID,
  ).Scan(&workoutsCount, &minutes)
  _ = a.DB.QueryRow("select points_balance from user_points where user_id = $1", employeeID).Scan(&points)

  data := map[string]any{
    "Employee": employee,
    "WorkoutsCount": workoutsCount,
    "HoursTotal": float64(minutes) / 60.0,
    "Points": points,
  }

  a.renderPage(w, r, "manager_employee", "Сотрудник", "", data)
}

func (a *App) ManagerAward(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil || (user.Role != "manager" && user.Role != "admin") {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  employeeID := r.FormValue("employee_id")
  points := r.FormValue("points")
  reason := r.FormValue("reason")
  if employeeID == "" || points == "" {
    a.setFlash(w, "Заполните все поля")
    http.Redirect(w, r, "/manager", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec(
    `insert into incentive_awards (user_id, points, reason, awarded_by)
     values ($1, $2, $3, $4)`,
    employeeID,
    points,
    reason,
    user.ID,
  )

  _, _ = a.DB.Exec(
    `update user_points
     set points_balance = points_balance + $1, points_total = points_total + $1, updated_at = now()
     where user_id = $2`,
    points,
    employeeID,
  )

  a.setFlash(w, "Баллы начислены")
  http.Redirect(w, r, "/manager", http.StatusSeeOther)
}
