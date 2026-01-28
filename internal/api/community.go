package api

import "net/http"

func (api *API) Departments(w http.ResponseWriter, r *http.Request) {
  rows, _ := api.DB.Query(
    `select coalesce(department, 'Не указан'), count(*) as members,
            coalesce(avg(workout_count), 0)
     from (
       select u.id, u.department, count(ws.id) as workout_count
       from users u
       left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
       left join user_settings us on us.user_id = u.id
       where u.role = 'employee' and coalesce(us.share_progress, true) = true
       group by u.id, u.department
     ) stats
     group by department
     order by members desc`,
  )

  departments := []map[string]any{}
  colors := []string{"bg-blue-500", "bg-purple-500", "bg-pink-500", "bg-green-500", "bg-orange-500"}
  if rows != nil {
    defer rows.Close()
    index := 0
    for rows.Next() {
      var name string
      var members int
      var avgWorkouts int
      _ = rows.Scan(&name, &members, &avgWorkouts)
      departments = append(departments, map[string]any{
        "name": name,
        "members": members,
        "avg_workouts": avgWorkouts,
        "color_class": colors[index%len(colors)],
      })
      index++
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{"departments": departments})
}

func (api *API) Challenges(w http.ResponseWriter, r *http.Request) {
  challenges := []map[string]any{
    {"id": "1", "title": "Декабрьский марафон", "description": "Завершите 20 тренировок в декабре", "participants": 156, "days_left": 20, "reward": "500 баллов", "progress": 12, "total": 20, "icon": "🏃"},
    {"id": "2", "title": "Командный дух", "description": "Ваш отдел должен выполнить 100 тренировок", "participants": 45, "days_left": 15, "reward": "300 баллов", "progress": 67, "total": 100, "icon": "👥"},
    {"id": "3", "title": "Утренняя зарядка", "description": "10 тренировок до 9:00", "participants": 89, "days_left": 30, "reward": "200 баллов", "progress": 3, "total": 10, "icon": "🌅"},
  }

  writeJSON(w, http.StatusOK, map[string]any{"challenges": challenges})
}

