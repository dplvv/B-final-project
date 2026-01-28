package api

import (
  "fmt"
  "net/http"
  "time"

  "github.com/go-chi/chi/v5"
)

type awardRequest struct {
  EmployeeID string `json:"employee_id"`
  Points     int    `json:"points"`
  Reason     string `json:"reason"`
}

type supportResponseRequest struct {
  Response string `json:"response"`
  Status   string `json:"status"`
}

func (api *API) ManagerDashboard(w http.ResponseWriter, r *http.Request) {
  var employeesCount int
  var totalWorkouts int
  var totalMinutes int
  _ = api.DB.QueryRow(
    `select coalesce(count(*), 0) from users where role = 'employee'`,
  ).Scan(&employeesCount)
  _ = api.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where completed_at is not null`,
  ).Scan(&totalWorkouts, &totalMinutes)

  employees := []map[string]any{}
  rows, _ := api.DB.Query(
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
      var id, name, department string
      var workouts, minutes, points int
      _ = rows.Scan(&id, &name, &department, &workouts, &minutes, &points)
      employees = append(employees, map[string]any{
        "id": id,
        "name": name,
        "department": department,
        "workouts": workouts,
        "hours": float64(minutes) / 60.0,
        "points": points,
      })
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "employees_count": employeesCount,
    "total_workouts": totalWorkouts,
    "total_hours": float64(totalMinutes) / 60.0,
    "employees": employees,
  })
}

func (api *API) ManagerEmployees(w http.ResponseWriter, r *http.Request) {
  rows, _ := api.DB.Query(
    `select u.id, u.name, coalesce(u.department, ''),
            coalesce(count(ws.id), 0), coalesce(sum(ws.duration_minutes), 0), coalesce(up.points_balance, 0)
     from users u
     left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
     left join user_points up on up.user_id = u.id
     where u.role = 'employee'
     group by u.id, up.points_balance
     order by u.name`,
  )

  employees := []map[string]any{}
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var id, name, department string
      var workouts, minutes, points int
      _ = rows.Scan(&id, &name, &department, &workouts, &minutes, &points)
      employees = append(employees, map[string]any{
        "id": id,
        "name": name,
        "department": department,
        "workouts": workouts,
        "hours": float64(minutes) / 60.0,
        "points": points,
      })
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{"employees": employees})
}

func (api *API) ManagerEmployeeDetail(w http.ResponseWriter, r *http.Request) {
  employeeID := chi.URLParam(r, "id")
  if employeeID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var name, employeeNumber, role, department, position string
  err := api.DB.QueryRow(
    `select id, name, employee_id, role, coalesce(department, ''), coalesce(position, '')
     from users where id = $1`,
    employeeID,
  ).Scan(&employeeID, &name, &employeeNumber, &role, &department, &position)
  if err != nil {
    writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
    return
  }

  var workoutsCount int
  var minutes int
  var points int
  _ = api.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions where user_id = $1 and completed_at is not null`,
    employeeID,
  ).Scan(&workoutsCount, &minutes)
  _ = api.DB.QueryRow("select points_balance from user_points where user_id = $1", employeeID).Scan(&points)

  writeJSON(w, http.StatusOK, map[string]any{
    "employee": map[string]any{
      "id": employeeID,
      "name": name,
      "employee_id": employeeNumber,
      "role": role,
      "department": department,
      "position": position,
    },
    "workouts": workoutsCount,
    "hours": float64(minutes) / 60.0,
    "points": points,
  })
}

func (api *API) ManagerAward(w http.ResponseWriter, r *http.Request) {
  managerID := userIDFromContext(r.Context())

  var req awardRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.EmployeeID == "" || req.Points == 0 {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  _, _ = api.DB.Exec(
    `insert into incentive_awards (user_id, points, reason, awarded_by)
     values ($1, $2, $3, $4)`,
    req.EmployeeID,
    req.Points,
    req.Reason,
    managerID,
  )

  _, _ = api.DB.Exec(
    `update user_points
     set points_balance = points_balance + $1, points_total = points_total + $1, updated_at = now()
     where user_id = $2`,
    req.Points,
    req.EmployeeID,
  )

  _, _ = api.DB.Exec(
    `insert into notifications (user_id, title, message, type)
     values ($1, $2, $3, $4)`,
    req.EmployeeID,
    "Начислены баллы",
    "Вам начислено "+fmt.Sprint(req.Points)+" баллов за достижение",
    "success",
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "awarded"})
}

func (api *API) ManagerRedemptions(w http.ResponseWriter, r *http.Request) {
  rows, _ := api.DB.Query(
    `select rr.id, rr.status, rr.redeemed_at, r.title, r.points_cost, u.name
     from reward_redemptions rr
     join rewards r on r.id = rr.reward_id
     join users u on u.id = rr.user_id
     order by rr.redeemed_at desc`,
  )

  redemptions := []map[string]any{}
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var id, status, title, userName string
      var redeemedAt time.Time
      var points int
      _ = rows.Scan(&id, &status, &redeemedAt, &title, &points, &userName)
      redemptions = append(redemptions, map[string]any{
        "id": id,
        "status": status,
        "redeemed_at": redeemedAt.Format("2006-01-02"),
        "reward": title,
        "points": points,
        "user": userName,
      })
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{"redemptions": redemptions})
}

func (api *API) ManagerApproveRedemption(w http.ResponseWriter, r *http.Request) {
  managerID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, _ = api.DB.Exec(
    `update reward_redemptions set status = 'approved', approved_by = $1 where id = $2`,
    managerID,
    id,
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "approved"})
}

func (api *API) ManagerRejectRedemption(w http.ResponseWriter, r *http.Request) {
  managerID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var userID string
  var points int
  _ = api.DB.QueryRow(
    `select rr.user_id, r.points_cost
     from reward_redemptions rr
     join rewards r on r.id = rr.reward_id
     where rr.id = $1`,
    id,
  ).Scan(&userID, &points)

  _, _ = api.DB.Exec(
    `update reward_redemptions set status = 'rejected', approved_by = $1 where id = $2`,
    managerID,
    id,
  )

  if userID != "" && points > 0 {
    _, _ = api.DB.Exec(
      `update user_points
       set points_balance = points_balance + $1, updated_at = now()
       where user_id = $2`,
      points,
      userID,
    )
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "rejected"})
}

func (api *API) ManagerSupportTickets(w http.ResponseWriter, r *http.Request) {
  rows, _ := api.DB.Query(
    `select st.id, st.category, st.subject, st.status, st.created_at, u.name
     from support_tickets st
     join users u on u.id = st.user_id
     order by st.created_at desc`,
  )

  tickets := []map[string]any{}
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var id, category, subject, status, userName string
      var createdAt time.Time
      _ = rows.Scan(&id, &category, &subject, &status, &createdAt, &userName)
      tickets = append(tickets, map[string]any{
        "id": id,
        "category": category,
        "subject": subject,
        "status": status,
        "created_at": createdAt.Format("2006-01-02"),
        "user": userName,
      })
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{"tickets": tickets})
}

func (api *API) ManagerSupportRespond(w http.ResponseWriter, r *http.Request) {
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  var req supportResponseRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  status := req.Status
  if status == "" {
    status = "resolved"
  }

  _, _ = api.DB.Exec(
    `update support_tickets set response = $1, status = $2, updated_at = now()
     where id = $3`,
    req.Response,
    status,
    id,
  )

  writeJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}
