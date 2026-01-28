package handlers

import (
  "net/http"
  "time"

  "rehab-app/internal/models"
)

type redemptionItem struct {
  RewardTitle string
  Status      string
  Date        string
}

func (a *App) Motivation(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var points int
  _ = a.DB.QueryRow("select points_balance from user_points where user_id = $1", user.ID).Scan(&points)

  rewards := []models.Reward{}
  rewardRows, _ := a.DB.Query(
    `select id, title, description, points_cost, coalesce(category, '')
     from rewards
     where active = true
     order by points_cost`,
  )
  if rewardRows != nil {
    defer rewardRows.Close()
    for rewardRows.Next() {
      var reward models.Reward
      _ = rewardRows.Scan(&reward.ID, &reward.Title, &reward.Description, &reward.PointsCost, &reward.Category)
      rewards = append(rewards, reward)
    }
  }

  leaderboard := []models.EmployeeStats{}
  rows, _ := a.DB.Query(
    `select u.id, u.name, coalesce(u.department, ''),
            coalesce(count(ws.id), 0), coalesce(sum(ws.duration_minutes), 0), coalesce(up.points_balance, 0)
     from users u
     left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
     left join user_points up on up.user_id = u.id
     group by u.id, up.points_balance
     order by up.points_balance desc
     limit 5`,
  )
  if rows != nil {
    defer rows.Close()
    for rows.Next() {
      var stat models.EmployeeStats
      var minutes int
      _ = rows.Scan(&stat.UserID, &stat.Name, &stat.Department, &stat.WorkoutsCount, &minutes, &stat.Points)
      stat.HoursTotal = float64(minutes) / 60.0
      leaderboard = append(leaderboard, stat)
    }
  }

  redemptions := []redemptionItem{}
  redemptionRows, _ := a.DB.Query(
    `select r.title, rr.status, rr.redeemed_at
     from reward_redemptions rr
     join rewards r on r.id = rr.reward_id
     where rr.user_id = $1
     order by rr.redeemed_at desc
     limit 5`,
    user.ID,
  )
  if redemptionRows != nil {
    defer redemptionRows.Close()
    for redemptionRows.Next() {
      var item redemptionItem
      var date time.Time
      _ = redemptionRows.Scan(&item.RewardTitle, &item.Status, &date)
      item.Date = date.Format("02.01.2006")
      redemptions = append(redemptions, item)
    }
  }

  data := map[string]any{
    "Points": points,
    "Rewards": rewards,
    "Leaderboard": leaderboard,
    "Redemptions": redemptions,
  }

  a.renderPage(w, r, "motivation", "Мотивация", "", data)
}

func (a *App) RedeemReward(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  rewardID := r.FormValue("reward_id")
  if rewardID == "" {
    http.Redirect(w, r, "/motivation", http.StatusSeeOther)
    return
  }

  var cost int
  _ = a.DB.QueryRow("select points_cost from rewards where id = $1", rewardID).Scan(&cost)

  var balance int
  _ = a.DB.QueryRow("select points_balance from user_points where user_id = $1", user.ID).Scan(&balance)
  if balance < cost {
    a.setFlash(w, "Недостаточно баллов")
    http.Redirect(w, r, "/motivation", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec(
    `insert into reward_redemptions (user_id, reward_id)
     values ($1, $2)`,
    user.ID,
    rewardID,
  )

  _, _ = a.DB.Exec(
    `update user_points
     set points_balance = points_balance - $1, updated_at = now()
     where user_id = $2`,
    cost,
    user.ID,
  )

  a.setFlash(w, "Заявка на награду отправлена")
  http.Redirect(w, r, "/motivation", http.StatusSeeOther)
}
