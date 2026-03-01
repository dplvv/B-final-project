package site

import (
  "net/http"
  "strconv"
  "strings"
  "time"

  "github.com/go-chi/chi/v5"
  "golang.org/x/crypto/bcrypt"

  "rehab-app/internal/db"
  "rehab-app/internal/middleware"
  "rehab-app/internal/models"
)

func (s *Site) rewardsPage(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  data := s.baseData(r, "Поощрения", "rewards")
  data["Error"] = r.URL.Query().Get("error")
  data["Success"] = r.URL.Query().Get("success")

  var points int
  _ = s.DB.QueryRow(`select coalesce(points_balance, 0) from user_points where user_id = $1`, user.ID).Scan(&points)
  data["Points"] = points

  var rows, err = s.DB.Query(
    `select id, title, description, points_cost, coalesce(category, '')
     from rewards
     where active = true
     order by points_cost`,
  )
  rewards := []rewardView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var rwd rewardView
      _ = rows.Scan(&rwd.ID, &rwd.Title, &rwd.Description, &rwd.PointsCost, &rwd.Category)
      rewards = append(rewards, rwd)
    }
  }
  data["Rewards"] = rewards

  redemptionStatus := map[string]string{}
  rows, err = s.DB.Query(
    `select reward_id, status from reward_redemptions where user_id = $1 and status = 'pending' order by redeemed_at desc`,
    user.ID,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var rewardID, status string
      _ = rows.Scan(&rewardID, &status)
      if _, exists := redemptionStatus[rewardID]; !exists {
        switch status {
        case "approved":
          redemptionStatus[rewardID] = "Одобрено"
        case "rejected":
          redemptionStatus[rewardID] = "Отклонено"
        default:
          redemptionStatus[rewardID] = "Ожидает подтверждения"
        }
      }
    }
  }
  data["RedemptionStatus"] = redemptionStatus

  s.render(w, "rewards", data)
}

func (s *Site) rewardRedeem(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  rewardID := chi.URLParam(r, "id")
  if rewardID == "" {
    http.Redirect(w, r, "/rewards?error=Не%20выбрано%20поощрение", http.StatusSeeOther)
    return
  }

  var cost int
  err := s.DB.QueryRow(`select points_cost from rewards where id = $1 and active = true`, rewardID).Scan(&cost)
  if err != nil {
    http.Redirect(w, r, "/rewards?error=Поощрение%20не%20найдено", http.StatusSeeOther)
    return
  }

  var points int
  _ = s.DB.QueryRow(`select coalesce(points_balance, 0) from user_points where user_id = $1`, user.ID).Scan(&points)
  if points < cost {
    http.Redirect(w, r, "/rewards?error=Недостаточно%20баллов", http.StatusSeeOther)
    return
  }

  _, _ = s.DB.Exec(
    `insert into reward_redemptions (user_id, reward_id, status)
     values ($1, $2, 'pending')`,
    user.ID,
    rewardID,
  )
  _, _ = s.DB.Exec(
    `update user_points set points_balance = greatest(points_balance - $1, 0) where user_id = $2`,
    cost,
    user.ID,
  )

  http.Redirect(w, r, "/rewards?success=Заявка%20на%20поощрение%20отправлена", http.StatusSeeOther)
}

func (s *Site) supportPage(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  data := s.baseData(r, "Поддержка", "support")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  var rows, err = s.DB.Query(
    `select id, subject, message, status, coalesce(response, ''), created_at
     from support_tickets
     where user_id = $1
     order by created_at desc`,
    user.ID,
  )
  tickets := []supportTicketView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var t supportTicketView
      var created time.Time
      _ = rows.Scan(&t.ID, &t.Subject, &t.Message, &t.Status, &t.Response, &created)
      t.CreatedAt = created.Format("02.01.2006 15:04")
      t.Messages = s.loadSupportMessages(t.ID)
      tickets = append(tickets, t)
    }
  }
  data["Tickets"] = tickets
  s.render(w, "support", data)
}

func (s *Site) loadSupportMessages(ticketID string) []supportMessageView {
  rows, err := s.DB.Query(
    `select coalesce(u.name, ''), m.sender_role, m.message, m.created_at
     from support_ticket_messages m
     left join users u on u.id = m.sender_id
     where m.ticket_id = $1
     order by m.created_at`,
    ticketID,
  )
  if err != nil {
    return nil
  }
  defer rows.Close()

  messages := []supportMessageView{}
  for rows.Next() {
    var msg supportMessageView
    var created time.Time
    _ = rows.Scan(&msg.SenderName, &msg.SenderRole, &msg.Message, &created)
    msg.CreatedAt = created.Format("02.01.2006 15:04")
    messages = append(messages, msg)
  }
  return messages
}

func (s *Site) supportSubmit(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/support", http.StatusSeeOther)
    return
  }
  subject := strings.TrimSpace(r.FormValue("subject"))
  message := strings.TrimSpace(r.FormValue("message"))
  if subject == "" || message == "" {
    http.Redirect(w, r, "/support?error=Заполните%20все%20поля", http.StatusSeeOther)
    return
  }

  var ticketID string
  _ = s.DB.QueryRow(
    `insert into support_tickets (user_id, category, subject, message)
     values ($1, 'general', $2, $3)
     returning id`,
    user.ID,
    subject,
    message,
  ).Scan(&ticketID)
  if ticketID != "" {
    _, _ = s.DB.Exec(
      `insert into support_ticket_messages (ticket_id, sender_id, sender_role, message)
       values ($1, $2, 'employee', $3)`,
      ticketID,
      user.ID,
      message,
    )
  }

  http.Redirect(w, r, "/support?success=Обращение%20отправлено", http.StatusSeeOther)
}

func (s *Site) supportMessageSubmit(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  ticketID := chi.URLParam(r, "id")
  if ticketID == "" {
    http.Redirect(w, r, "/support", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/support", http.StatusSeeOther)
    return
  }
  message := strings.TrimSpace(r.FormValue("message"))
  if message == "" {
    http.Redirect(w, r, "/support?error=Заполните%20сообщение", http.StatusSeeOther)
    return
  }

  var ownerID string
  _ = s.DB.QueryRow(`select user_id from support_tickets where id = $1`, ticketID).Scan(&ownerID)
  if ownerID != user.ID {
    http.Error(w, "Доступ запрещён", http.StatusForbidden)
    return
  }

  _, _ = s.DB.Exec(
    `insert into support_ticket_messages (ticket_id, sender_id, sender_role, message)
     values ($1, $2, 'employee', $3)`,
    ticketID,
    user.ID,
    message,
  )
  _, _ = s.DB.Exec(`update support_tickets set status = 'open', updated_at = now() where id = $1`, ticketID)
  http.Redirect(w, r, "/support?success=Сообщение%20добавлено", http.StatusSeeOther)
}

func (s *Site) supportClose(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  ticketID := chi.URLParam(r, "id")
  if ticketID == "" {
    http.Redirect(w, r, "/support", http.StatusSeeOther)
    return
  }

  var ownerID string
  err := s.DB.QueryRow(`select user_id from support_tickets where id = $1`, ticketID).Scan(&ownerID)
  if err != nil {
    http.Redirect(w, r, "/support?error=Обращение%20не%20найдено", http.StatusSeeOther)
    return
  }
  if ownerID != user.ID {
    http.Error(w, "Доступ запрещён", http.StatusForbidden)
    return
  }

  _, _ = s.DB.Exec(`update support_tickets set status = 'closed', updated_at = now() where id = $1`, ticketID)
  _, _ = s.DB.Exec(
    `insert into support_ticket_messages (ticket_id, sender_id, sender_role, message)
     values ($1, $2, 'employee', $3)`,
    ticketID,
    user.ID,
    "Проблема решена",
  )
  http.Redirect(w, r, "/support?success=Обращение%20закрыто", http.StatusSeeOther)
}

func (s *Site) managerDashboard(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  data := s.baseData(r, "Руководитель", "manager")
  period := normalizeTrendPeriod(r.URL.Query().Get("period"))
  rangeFilter := parseTrendDateRange(r, time.Now())
  trendConfig, _, _ := trendPeriodConfigWithRange(period, time.Now(), rangeFilter)
  data["TrendPeriod"] = period
  data["TrendBadge"] = trendConfig.Badge
  data["TrendDateFrom"] = rangeFilter.FromValue
  data["TrendDateTo"] = rangeFilter.ToValue

  department := strings.TrimSpace(user.Department)
  query := `select u.id, u.name, u.employee_id, coalesce(u.department, ''), coalesce(u.position, ''), coalesce(p.points_balance, 0)
            from users u
            left join user_points p on p.user_id = u.id
            where u.role = 'employee'`
  args := []any{}
  if user.Role == "manager" && department != "" {
    query += " and u.department = $1"
    args = append(args, department)
  }
  query += " order by u.name"

  rows, err := s.DB.Query(query, args...)
  employees := []managerEmployeeView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var e managerEmployeeView
      _ = rows.Scan(&e.ID, &e.Name, &e.EmployeeID, &e.Department, &e.Position, &e.Points)
      employees = append(employees, e)
    }
  }

  var totalAchievements int
  _ = s.DB.QueryRow(`select count(*) from achievements`).Scan(&totalAchievements)
  unlockedMap := map[string]int{}
  rows, err = s.DB.Query(
    `select user_id, count(*)
     from user_achievements
     where unlocked = true
     group by user_id`,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var userID string
      var count int
      _ = rows.Scan(&userID, &count)
      unlockedMap[userID] = count
    }
  }

  for i := range employees {
    employees[i].AchievementsTotal = totalAchievements
    employees[i].AchievementsUnlocked = unlockedMap[employees[i].ID]
  }
  data["Employees"] = employees

  redemptionQuery := `select rr.id, u.name, coalesce(u.department, ''), r.title, r.points_cost
                      from reward_redemptions rr
                      join users u on u.id = rr.user_id
                      join rewards r on r.id = rr.reward_id
                      where rr.status = 'pending'`
  args = []any{}
  if user.Role == "manager" && department != "" {
    redemptionQuery += " and u.department = $1"
    args = append(args, department)
  }
  redemptionQuery += " order by rr.redeemed_at desc"

  rows, err = s.DB.Query(redemptionQuery, args...)
  redemptions := []redemptionView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var rv redemptionView
      _ = rows.Scan(&rv.ID, &rv.EmployeeName, &rv.Department, &rv.RewardTitle, &rv.PointsCost)
      redemptions = append(redemptions, rv)
    }
  }
  data["Redemptions"] = redemptions
  data["DepartmentTimeTrend"] = s.loadDepartmentTrendPeriod(department, period, rangeFilter)
  data["DepartmentEmployeeTrend"] = s.loadEmployeeTrend(department, period, rangeFilter)

  s.render(w, "manager", data)
}

func (s *Site) managerEmployee(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  employeeID := chi.URLParam(r, "id")
  if employeeID == "" {
    http.NotFound(w, r)
    return
  }

  var employee managerEmployeeView
  err := s.DB.QueryRow(
    `select id, name, employee_id, coalesce(department, ''), coalesce(position, '')
     from users where id = $1 and role = 'employee'`,
    employeeID,
  ).Scan(&employee.ID, &employee.Name, &employee.EmployeeID, &employee.Department, &employee.Position)
  if err != nil {
    http.NotFound(w, r)
    return
  }
  if user.Role == "manager" && user.Department != "" && !strings.EqualFold(user.Department, employee.Department) {
    http.Error(w, "Доступ запрещён", http.StatusForbidden)
    return
  }

  var points int
  _ = s.DB.QueryRow(`select coalesce(points_balance, 0) from user_points where user_id = $1`, employee.ID).Scan(&points)

  rows, err := s.DB.Query(
    `select a.title, a.description, a.icon, a.points_reward,
            coalesce(ua.unlocked, false), coalesce(ua.progress, 0), coalesce(ua.total, 0)
     from achievements a
     left join user_achievements ua on ua.achievement_id = a.id and ua.user_id = $1
     order by a.title`,
    employee.ID,
  )
  achievements := []achievementView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var v achievementView
      _ = rows.Scan(&v.Title, &v.Description, &v.Icon, &v.PointsReward, &v.Unlocked, &v.Progress, &v.Total)
      achievements = append(achievements, v)
    }
  }

  workoutHistory := []managerWorkoutView{}
  rows, err = s.DB.Query(
    `select w.name, ws.completed_at, coalesce(ws.duration_minutes, 0), coalesce(ws.calories_burned, 0)
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1 and ws.completed_at is not null
     order by ws.completed_at desc
     limit 12`,
    employee.ID,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var v managerWorkoutView
      var completedAt time.Time
      _ = rows.Scan(&v.Name, &completedAt, &v.Duration, &v.Calories)
      v.CompletedAt = completedAt.Format("02.01.2006 15:04")
      workoutHistory = append(workoutHistory, v)
    }
  }

  data := s.baseData(r, employee.Name, "manager")
  data["Employee"] = employee
  data["Points"] = points
  data["Achievements"] = achievements
  data["Workouts"] = workoutHistory
  s.render(w, "manager_employee", data)
}

func (s *Site) managerAward(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  employeeID := chi.URLParam(r, "id")
  if employeeID == "" {
    http.NotFound(w, r)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/manager", http.StatusSeeOther)
    return
  }
  points, _ := strconv.Atoi(r.FormValue("points"))
  if points <= 0 {
    http.Redirect(w, r, "/manager/employees/"+employeeID, http.StatusSeeOther)
    return
  }
  reason := strings.TrimSpace(r.FormValue("reason"))

  if user.Role == "manager" && user.Department != "" {
    var dept string
    _ = s.DB.QueryRow(`select coalesce(department, '') from users where id = $1`, employeeID).Scan(&dept)
    if !strings.EqualFold(dept, user.Department) {
      http.Error(w, "Доступ запрещён", http.StatusForbidden)
      return
    }
  }

  _, _ = s.DB.Exec(
    `insert into incentive_awards (user_id, points, reason, awarded_by)
     values ($1, $2, $3, $4)`,
    employeeID,
    points,
    reason,
    user.ID,
  )
  _, _ = s.DB.Exec(
    `insert into user_points (user_id, points_balance, points_total)
     values ($1, $2, $2)
     on conflict (user_id)
     do update set points_balance = user_points.points_balance + $2,
                   points_total = user_points.points_total + $2`,
    employeeID,
    points,
  )

  http.Redirect(w, r, "/manager/employees/"+employeeID, http.StatusSeeOther)
}

func (s *Site) managerRedemptionApprove(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  redemptionID := chi.URLParam(r, "id")
  if redemptionID == "" {
    http.Redirect(w, r, "/manager", http.StatusSeeOther)
    return
  }

  if !s.managerRedemptionAllowed(user, redemptionID) {
    http.Error(w, "Доступ запрещён", http.StatusForbidden)
    return
  }

  _, _ = s.DB.Exec(
    `update reward_redemptions set status = 'approved', approved_by = $1, handled_at = now() where id = $2`,
    user.ID,
    redemptionID,
  )
  http.Redirect(w, r, "/manager", http.StatusSeeOther)
}

func (s *Site) managerRedemptionReject(w http.ResponseWriter, r *http.Request) {
  user := middleware.UserFromContext(r.Context())
  redemptionID := chi.URLParam(r, "id")
  if redemptionID == "" {
    http.Redirect(w, r, "/manager", http.StatusSeeOther)
    return
  }

  var userID string
  var cost int
  err := s.DB.QueryRow(
    `select rr.user_id, r.points_cost
     from reward_redemptions rr
     join rewards r on r.id = rr.reward_id
     where rr.id = $1`,
    redemptionID,
  ).Scan(&userID, &cost)
  if err != nil {
    http.Redirect(w, r, "/manager", http.StatusSeeOther)
    return
  }

  if user.Role == "manager" && user.Department != "" {
    var dept string
    _ = s.DB.QueryRow(`select coalesce(department, '') from users where id = $1`, userID).Scan(&dept)
    if !strings.EqualFold(dept, user.Department) {
      http.Error(w, "Доступ запрещён", http.StatusForbidden)
      return
    }
  }

  _, _ = s.DB.Exec(
    `update reward_redemptions set status = 'rejected', approved_by = $1, handled_at = now() where id = $2`,
    user.ID,
    redemptionID,
  )
  _, _ = s.DB.Exec(
    `update user_points set points_balance = points_balance + $1 where user_id = $2`,
    cost,
    userID,
  )

  http.Redirect(w, r, "/manager", http.StatusSeeOther)
}

func (s *Site) adminDashboard(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Администрирование", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")
  period := normalizeTrendPeriod(r.URL.Query().Get("period"))
  rangeFilter := parseTrendDateRange(r, time.Now())
  trendConfig, trendStart, trendEnd := trendPeriodConfigWithRange(period, time.Now(), rangeFilter)
  data["TrendPeriod"] = period
  data["TrendBadge"] = trendConfig.Badge
  data["TrendDateFrom"] = rangeFilter.FromValue
  data["TrendDateTo"] = rangeFilter.ToValue

  var usersCount int
  _ = s.DB.QueryRow(`select count(*) from users`).Scan(&usersCount)
  var programs int
  var workouts int
  _ = s.DB.QueryRow(`select count(*) from programs`).Scan(&programs)
  _ = s.DB.QueryRow(`select count(*) from workouts`).Scan(&workouts)
  var pendingResets int
  var pendingSupport int
  var pendingRewards int
  _ = s.DB.QueryRow(`select count(*) from password_reset_requests where status = 'open'`).Scan(&pendingResets)
  _ = s.DB.QueryRow(`select count(*) from support_tickets where status = 'open'`).Scan(&pendingSupport)
  _ = s.DB.QueryRow(`select count(*) from reward_redemptions where status = 'pending'`).Scan(&pendingRewards)
  data["Stats"] = map[string]int{
    "Users":   usersCount,
    "Content": programs + workouts,
    "Pending": pendingResets + pendingSupport + pendingRewards,
  }
  data["AdminTimeTrend"] = s.loadDepartmentTrendPeriod("", period, rangeFilter)
  data["AdminEmployeeTrend"] = s.loadEmployeeTrend("", period, rangeFilter)

  topWorkouts := []analyticsItem{}
  rows, err := s.DB.Query(
    `select w.name, count(ws.id)
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.completed_at is not null
     group by w.name
     order by count(ws.id) desc
     limit 5`,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var item analyticsItem
      _ = rows.Scan(&item.Name, &item.Count)
      topWorkouts = append(topWorkouts, item)
    }
  }

  topExercises := []analyticsItem{}
  rows, err = s.DB.Query(
    `select e.name, count(wse.id)
     from workout_session_exercises wse
     join workout_sessions ws on ws.id = wse.session_id
     join exercises e on e.id = wse.exercise_id
     where ws.completed_at is not null
     group by e.name
     order by count(wse.id) desc
     limit 5`,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var item analyticsItem
      _ = rows.Scan(&item.Name, &item.Count)
      topExercises = append(topExercises, item)
    }
  }
  data["TopWorkouts"] = topWorkouts
  data["TopExercises"] = topExercises

  programAnalytics := []programAnalyticsView{}
  rows, err = s.DB.Query(
    `select p.name, count(pw.workout_id), coalesce(sum(w.duration_minutes), 0)
     from programs p
     left join program_workouts pw on pw.program_id = p.id
     left join workouts w on w.id = pw.workout_id
     group by p.id, p.name
     order by count(pw.workout_id) desc, p.name
     limit 6`,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var item programAnalyticsView
      _ = rows.Scan(&item.Name, &item.Workouts, &item.Minutes)
      programAnalytics = append(programAnalytics, item)
    }
  }
  data["ProgramAnalytics"] = programAnalytics

  intensityStats := []intensityStat{}
  rows, err = s.DB.Query(
    `select intensity, count(*)
     from training_plan_workouts
     group by intensity
     order by intensity`,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var stat intensityStat
      _ = rows.Scan(&stat.Level, &stat.Count)
      intensityStats = append(intensityStats, stat)
    }
  }
  data["IntensityStats"] = intensityStats
  data["DepartmentStats"] = s.loadDepartmentStats(trendStart, trendEnd)

  rows, err = s.DB.Query(
    `select u.id, u.name, u.employee_id, u.role, coalesce(u.department, ''), coalesce(u.position, ''),
            coalesce(to_char(up.birth_date, 'YYYY-MM-DD'), ''),
            coalesce(extract(year from age(current_date, up.birth_date))::int, up.age, 0),
            coalesce(mi.doctor_approval, false)
     from users u
     left join user_profiles up on up.user_id = u.id
     left join medical_info mi on mi.user_id = u.id
     order by u.created_at desc`,
  )
  users := []managerEmployeeView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var u managerEmployeeView
      _ = rows.Scan(&u.ID, &u.Name, &u.EmployeeID, &u.Role, &u.Department, &u.Position, &u.BirthDate, &u.Age, &u.DoctorApproval)
      users = append(users, u)
    }
  }
  data["Users"] = users

  rows, err = s.DB.Query(
    `select pr.id, u.name, u.employee_id, pr.created_at
     from password_reset_requests pr
     join users u on u.id = pr.user_id
     where pr.status = 'open'
     order by pr.created_at desc`,
  )
  requests := []supportTicketView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var req supportTicketView
      var created time.Time
      _ = rows.Scan(&req.ID, &req.EmployeeName, &req.EmployeeID, &created)
      req.CreatedAt = created.Format("02.01.2006 15:04")
      requests = append(requests, req)
    }
  }
  data["PasswordRequests"] = requests

  s.render(w, "admin", data)
}

func (s *Site) adminUserCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }

  name := strings.TrimSpace(r.FormValue("name"))
  employeeID := strings.TrimSpace(r.FormValue("employee_id"))
  role := strings.TrimSpace(r.FormValue("role"))
  department := strings.TrimSpace(r.FormValue("department"))
  position := strings.TrimSpace(r.FormValue("position"))
  birthDateRaw := strings.TrimSpace(r.FormValue("birth_date"))
  tempPassword := r.FormValue("temp_password")
  if name == "" || employeeID == "" || tempPassword == "" {
    http.Redirect(w, r, "/admin?error=Заполните%20все%20поля", http.StatusSeeOther)
    return
  }
  if role == "" {
    role = "employee"
  }

  var birthDate time.Time
  if birthDateRaw != "" {
    parsed, err := time.Parse("2006-01-02", birthDateRaw)
    if err != nil {
      http.Redirect(w, r, "/admin?error=Некорректная%20дата%20рождения", http.StatusSeeOther)
      return
    }
    birthDate = parsed
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
  if err != nil {
    http.Redirect(w, r, "/admin?error=Ошибка%20пароля", http.StatusSeeOther)
    return
  }

  var userID string
  err = s.DB.QueryRow(
    `insert into users (name, employee_id, password_hash, role, department, position, password_temp)
     values ($1, $2, $3, $4, nullif($5, ''), nullif($6, ''), true)
     returning id`,
    name,
    employeeID,
    string(hash),
    role,
    department,
    position,
  ).Scan(&userID)
  if err != nil {
    http.Redirect(w, r, "/admin?error=ID%20уже%20занят", http.StatusSeeOther)
    return
  }

  _ = db.EnsureUserDefaults(s.DB, userID)
  if !birthDate.IsZero() {
    age := ageFromBirthDate(birthDate, time.Now())
    _, _ = s.DB.Exec(
      `update user_profiles
       set birth_date = $1,
           age = $2,
           updated_at = now()
       where user_id = $3`,
      birthDate,
      age,
      userID,
    )
  }
  http.Redirect(w, r, "/admin?success=Пользователь%20создан", http.StatusSeeOther)
}

func (s *Site) adminUserUpdate(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.Redirect(w, r, "/admin?error=Не%20найден%20пользователь", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }

  name := strings.TrimSpace(r.FormValue("name"))
  employeeID := strings.TrimSpace(r.FormValue("employee_id"))
  department := strings.TrimSpace(r.FormValue("department"))
  position := strings.TrimSpace(r.FormValue("position"))
  role := strings.TrimSpace(r.FormValue("role"))
  birthDateRaw := strings.TrimSpace(r.FormValue("birth_date"))
  doctorApproval := r.FormValue("doctor_approval") == "on"

  _, err := s.DB.Exec(
    `update users set
       name = case when $1 <> '' then $1 else name end,
       employee_id = case when $2 <> '' then $2 else employee_id end,
       department = case when $3 <> '' then $3 else department end,
       position = case when $4 <> '' then $4 else position end,
       role = case when $5 <> '' then $5 else role end,
       updated_at = now()
     where id = $6`,
    name,
    employeeID,
    department,
    position,
    role,
    userID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin?error=Не%20удалось%20обновить%20пользователя", http.StatusSeeOther)
    return
  }
  _, _ = s.DB.Exec(
    `insert into medical_info (user_id, doctor_approval, updated_at)
     values ($2, $1, now())
     on conflict (user_id)
     do update set doctor_approval = excluded.doctor_approval, updated_at = now()`,
    doctorApproval,
    userID,
  )

  if birthDateRaw != "" {
    birthDate, err := time.Parse("2006-01-02", birthDateRaw)
    if err != nil {
      http.Redirect(w, r, "/admin?error=Некорректная%20дата%20рождения", http.StatusSeeOther)
      return
    }
    age := ageFromBirthDate(birthDate, time.Now())
    _, _ = s.DB.Exec(
      `insert into user_profiles (user_id, birth_date, age, updated_at)
       values ($1, $2, $3, now())
       on conflict (user_id)
       do update set birth_date = excluded.birth_date,
                     age = excluded.age,
                     updated_at = now()`,
      userID,
      birthDate,
      age,
    )
  }

  http.Redirect(w, r, "/admin?success=Данные%20обновлены", http.StatusSeeOther)
}

func (s *Site) adminUserDelete(w http.ResponseWriter, r *http.Request) {
  admin := middleware.UserFromContext(r.Context())
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.Redirect(w, r, "/admin?error=Не%20найден%20пользователь", http.StatusSeeOther)
    return
  }
  if admin != nil && admin.ID == userID {
    http.Redirect(w, r, "/admin?error=Нельзя%20удалить%20текущего%20пользователя", http.StatusSeeOther)
    return
  }

  var role string
  err := s.DB.QueryRow(`select role from users where id = $1`, userID).Scan(&role)
  if err != nil {
    http.Redirect(w, r, "/admin?error=Пользователь%20не%20найден", http.StatusSeeOther)
    return
  }
  if strings.EqualFold(role, "admin") {
    var admins int
    _ = s.DB.QueryRow(`select count(*) from users where role = 'admin'`).Scan(&admins)
    if admins <= 1 {
      http.Redirect(w, r, "/admin?error=Нельзя%20удалить%20последнего%20администратора", http.StatusSeeOther)
      return
    }
  }

  if _, err := s.DB.Exec(`delete from users where id = $1`, userID); err != nil {
    http.Redirect(w, r, "/admin?error=Не%20удалось%20удалить%20пользователя", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin?success=Пользователь%20удален", http.StatusSeeOther)
}

func (s *Site) adminUserResetPassword(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.Redirect(w, r, "/admin?error=Не%20найден%20пользователь", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  tempPassword := r.FormValue("temp_password")
  if tempPassword == "" {
    http.Redirect(w, r, "/admin?error=Введите%20временный%20пароль", http.StatusSeeOther)
    return
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
  if err != nil {
    http.Redirect(w, r, "/admin?error=Ошибка%20пароля", http.StatusSeeOther)
    return
  }
  _, _ = s.DB.Exec(
    `update users set password_hash = $1, password_temp = true, updated_at = now() where id = $2`,
    string(hash),
    userID,
  )
  http.Redirect(w, r, "/admin?success=Пароль%20сброшен", http.StatusSeeOther)
}

func (s *Site) adminPasswordRequestResolve(w http.ResponseWriter, r *http.Request) {
  admin := middleware.UserFromContext(r.Context())
  requestID := chi.URLParam(r, "id")
  if requestID == "" {
    http.Redirect(w, r, "/admin?error=Не%20найден%20запрос", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  tempPassword := r.FormValue("temp_password")
  if tempPassword == "" {
    http.Redirect(w, r, "/admin?error=Введите%20временный%20пароль", http.StatusSeeOther)
    return
  }

  var userID string
  err := s.DB.QueryRow(`select user_id from password_reset_requests where id = $1 and status = 'open'`, requestID).Scan(&userID)
  if err != nil {
    http.Redirect(w, r, "/admin?error=Запрос%20не%20найден", http.StatusSeeOther)
    return
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
  if err != nil {
    http.Redirect(w, r, "/admin?error=Ошибка%20пароля", http.StatusSeeOther)
    return
  }

  _, _ = s.DB.Exec(
    `update users set password_hash = $1, password_temp = true, updated_at = now() where id = $2`,
    string(hash),
    userID,
  )
  _, _ = s.DB.Exec(
    `update password_reset_requests
     set status = 'resolved', handled_at = now(), handled_by = $1
     where id = $2`,
    admin.ID,
    requestID,
  )

  http.Redirect(w, r, "/admin?success=Пароль%20сброшен", http.StatusSeeOther)
}

func (s *Site) adminFeedback(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Отзывы", "admin")
  rows, err := s.DB.Query(
    `select u.name, w.name, f.perceived_exertion, f.tolerance, f.pain_level, f.wellbeing, coalesce(f.comment, ''), f.created_at
     from workout_session_feedback f
     join users u on u.id = f.user_id
     join workout_sessions ws on ws.id = f.session_id
     join workouts w on w.id = ws.workout_id
     order by f.created_at desc`,
  )
  list := []feedbackAdminView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var v feedbackAdminView
      var created time.Time
      _ = rows.Scan(&v.EmployeeName, &v.WorkoutName, &v.PerceivedExertion, &v.Tolerance, &v.PainLevel, &v.Wellbeing, &v.Comment, &created)
      v.CreatedAt = created.Format("02.01.2006 15:04")
      list = append(list, v)
    }
  }
  data["Feedbacks"] = list
  s.render(w, "admin_feedback", data)
}

func (s *Site) adminSupport(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Обращения", "admin")
  rows, err := s.DB.Query(
    `select t.id, u.name, t.subject, t.message, t.status, coalesce(t.response, ''), t.created_at
     from support_tickets t
     join users u on u.id = t.user_id
     order by t.created_at desc`,
  )
  tickets := []supportTicketView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var t supportTicketView
      var created time.Time
      _ = rows.Scan(&t.ID, &t.EmployeeName, &t.Subject, &t.Message, &t.Status, &t.Response, &created)
      t.CreatedAt = created.Format("02.01.2006 15:04")
      t.Messages = s.loadSupportMessages(t.ID)
      tickets = append(tickets, t)
    }
  }
  data["Tickets"] = tickets
  s.render(w, "admin_support", data)
}

func (s *Site) adminSupportRespond(w http.ResponseWriter, r *http.Request) {
  admin := middleware.UserFromContext(r.Context())
  ticketID := chi.URLParam(r, "id")
  if ticketID == "" {
    http.Redirect(w, r, "/admin/support", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/support", http.StatusSeeOther)
    return
  }
  response := strings.TrimSpace(r.FormValue("response"))
  if response == "" {
    http.Redirect(w, r, "/admin/support", http.StatusSeeOther)
    return
  }

  if admin != nil {
    _, _ = s.DB.Exec(
    `insert into support_ticket_messages (ticket_id, sender_id, sender_role, message)
     values ($1, $2, 'admin', $3)`,
    ticketID,
    admin.ID,
    response,
  )
  }
  _, _ = s.DB.Exec(
    `update support_tickets
     set response = $1,
         status = case when status = 'closed' then 'closed' else 'open' end,
         updated_at = now()
     where id = $2`,
    response,
    ticketID,
  )
  http.Redirect(w, r, "/admin/support", http.StatusSeeOther)
}

func (s *Site) adminSupportClose(w http.ResponseWriter, r *http.Request) {
  admin := middleware.UserFromContext(r.Context())
  if admin == nil || admin.Role != "admin" {
    http.Error(w, "Доступ запрещён", http.StatusForbidden)
    return
  }
  ticketID := chi.URLParam(r, "id")
  if ticketID == "" {
    http.Redirect(w, r, "/admin/support", http.StatusSeeOther)
    return
  }

  _, _ = s.DB.Exec(
    `update support_tickets
     set status = 'closed', updated_at = now()
     where id = $1`,
    ticketID,
  )
  _, _ = s.DB.Exec(
    `insert into support_ticket_messages (ticket_id, sender_id, sender_role, message)
     values ($1, $2, 'admin', $3)`,
    ticketID,
    admin.ID,
    "Проблема решена",
  )
  http.Redirect(w, r, "/admin/support", http.StatusSeeOther)
}

func (s *Site) managerRedemptionAllowed(user *models.User, redemptionID string) bool {
  if user == nil {
    return false
  }
  if user.Role != "manager" || user.Department == "" {
    return false
  }

  var dept string
  err := s.DB.QueryRow(
    `select coalesce(u.department, '')
     from reward_redemptions rr
     join users u on u.id = rr.user_id
     where rr.id = $1`,
    redemptionID,
  ).Scan(&dept)
  if err != nil {
    return false
  }
  return strings.EqualFold(dept, user.Department)
}

func (s *Site) loadDepartmentStats(start, end time.Time) []departmentStatView {
  rows, err := s.DB.Query(
    `with employee_base as (
       select id,
              coalesce(nullif(trim(department), ''), 'Без отдела') as department
       from users
       where role = 'employee'
     ),
     workout_stat as (
       select ws.user_id,
              count(*) as workouts,
              coalesce(sum(ws.duration_minutes), 0) as minutes
       from workout_sessions ws
       where ws.completed_at is not null
         and ws.completed_at >= $1
         and ws.completed_at < $2
       group by ws.user_id
     ),
     feedback_stat as (
       select f.user_id,
              avg(coalesce(f.tolerance, 0)) as tolerance
       from workout_session_feedback f
       where f.created_at >= $1
         and f.created_at < $2
       group by f.user_id
     )
     select eb.department,
            count(*) as employees,
            coalesce(sum(ws.workouts), 0) as workouts,
            coalesce(sum(ws.minutes), 0) as minutes,
            coalesce(avg(fs.tolerance), 0) as avg_tolerance
     from employee_base eb
     left join workout_stat ws on ws.user_id = eb.id
     left join feedback_stat fs on fs.user_id = eb.id
     group by eb.department
     order by workouts desc, eb.department`,
    start,
    end,
  )
  if err != nil {
    return nil
  }
  defer rows.Close()

  stats := []departmentStatView{}
  for rows.Next() {
    var item departmentStatView
    var avgTolerance float64
    _ = rows.Scan(&item.Department, &item.Employees, &item.Workouts, &item.Minutes, &avgTolerance)
    item.AvgTolerance = int(avgTolerance + 0.5)
    if item.AvgTolerance < 0 {
      item.AvgTolerance = 0
    }
    if item.AvgTolerance > 5 {
      item.AvgTolerance = 5
    }
    stats = append(stats, item)
  }
  return stats
}

func ageFromBirthDate(birthDate time.Time, now time.Time) int {
  if birthDate.IsZero() {
    return 0
  }
  years := now.Year() - birthDate.Year()
  birthdayThisYear := time.Date(now.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, now.Location())
  if now.Before(birthdayThisYear) {
    years--
  }
  if years < 0 {
    return 0
  }
  return years
}
