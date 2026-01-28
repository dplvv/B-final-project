package api

import (
  "database/sql"
  "time"
)

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

