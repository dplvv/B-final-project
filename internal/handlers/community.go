package handlers

import "net/http"

type leaderboardEntry struct {
  UserID     string
  Name       string
  Department string
  Avatar     string
  Workouts   int
  Hours      float64
  Streak     int
  Points     int
}

type departmentStat struct {
  Name        string
  Members     int
  AvgWorkouts int
  ColorClass  string
}

type challenge struct {
  Title       string
  Description string
  Participants int
  DaysLeft    int
  Reward      string
  Progress    int
  Total       int
  Icon        string
}

func (a *App) Community(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  tab := r.URL.Query().Get("tab")
  if tab == "" {
    tab = "leaderboard"
  }

  leaderboard := []leaderboardEntry{}
  rows, err := a.DB.Query(
    `select u.id, u.name, coalesce(u.department, ''),
            coalesce(count(ws.id), 0) as workouts,
            coalesce(sum(ws.duration_minutes), 0) as minutes,
            coalesce(up.points_balance, 0) as points
     from users u
     left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
     left join user_points up on up.user_id = u.id
     group by u.id, up.points_balance
     order by points desc, workouts desc
     limit 10`,
  )
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var entry leaderboardEntry
      var minutes int
      _ = rows.Scan(&entry.UserID, &entry.Name, &entry.Department, &entry.Workouts, &minutes, &entry.Points)
      entry.Hours = float64(minutes) / 60.0
      entry.Avatar = "👤"
      entry.Streak = computeWorkoutStreak(a.DB, entry.UserID)
      leaderboard = append(leaderboard, entry)
    }
  }

  currentRank := 0
  var currentEntry leaderboardEntry
  for i, entry := range leaderboard {
    if entry.UserID == user.ID {
      currentRank = i + 1
      currentEntry = entry
      break
    }
  }
  if currentEntry.UserID == "" {
    var workouts int
    var minutes int
    var points int
    _ = a.DB.QueryRow(
      `select coalesce(count(ws.id), 0), coalesce(sum(ws.duration_minutes), 0), coalesce(up.points_balance, 0)
       from users u
       left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
       left join user_points up on up.user_id = u.id
       where u.id = $1
       group by up.points_balance`,
      user.ID,
    ).Scan(&workouts, &minutes, &points)
    currentEntry = leaderboardEntry{
      UserID: user.ID,
      Name: user.Name,
      Department: user.Department,
      Avatar: "👤",
      Workouts: workouts,
      Hours: float64(minutes) / 60.0,
      Streak: computeWorkoutStreak(a.DB, user.ID),
      Points: points,
    }
  }

  if currentRank == 0 {
    _ = a.DB.QueryRow(
      `select rank from (
         select user_id, rank() over (order by points_balance desc) as rank
         from user_points
       ) ranked where user_id = $1`,
      user.ID,
    ).Scan(&currentRank)
  }

  departments := []departmentStat{}
  depRows, _ := a.DB.Query(
    `select coalesce(department, 'Не указан'), count(*) as members,
            coalesce(avg(workout_count), 0)
     from (
       select u.id, u.department, count(ws.id) as workout_count
       from users u
       left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
       group by u.id, u.department
     ) stats
     group by department
     order by members desc`,
  )
  colors := []string{"bg-blue-500", "bg-purple-500", "bg-pink-500", "bg-green-500", "bg-orange-500"}
  if depRows != nil {
    defer depRows.Close()
    index := 0
    for depRows.Next() {
      var stat departmentStat
      _ = depRows.Scan(&stat.Name, &stat.Members, &stat.AvgWorkouts)
      stat.ColorClass = colors[index%len(colors)]
      departments = append(departments, stat)
      index++
    }
  }

  challenges := []challenge{
    {Title: "Декабрьский марафон", Description: "Завершите 20 тренировок в декабре", Participants: 156, DaysLeft: 20, Reward: "500 баллов", Progress: 12, Total: 20, Icon: "🏃"},
    {Title: "Командный дух", Description: "Ваш отдел должен выполнить 100 тренировок", Participants: 45, DaysLeft: 15, Reward: "300 баллов", Progress: 67, Total: 100, Icon: "👥"},
    {Title: "Утренняя зарядка", Description: "10 тренировок до 9:00", Participants: 89, DaysLeft: 30, Reward: "200 баллов", Progress: 3, Total: 10, Icon: "🌅"},
  }

  data := map[string]any{
    "Tab":         tab,
    "Leaderboard": leaderboard,
    "CurrentUser": currentEntry,
    "CurrentRank": currentRank,
    "Departments": departments,
    "Challenges":  challenges,
  }

  a.renderPage(w, r, "community", "Сообщество", "", data)
}
