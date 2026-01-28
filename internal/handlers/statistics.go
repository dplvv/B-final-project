package handlers

import (
  "fmt"
  "net/http"
  "strings"
  "time"
)

type weeklyStat struct {
  Day      string
  Workouts int
  Duration int
}

type monthlyTrend struct {
  Month    string
  Workouts int
}

type categoryStat struct {
  Category   string
  Percentage int
  Color      string
  ColorClass string
}

type recordStat struct {
  Title string
  Value string
  IconClass string
  BgClass string
}

func (a *App) Statistics(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  weekly := buildWeeklyStats(a, user.ID)
  monthly := buildMonthlyTrends(a, user.ID)
  categories := buildCategoryStats(a, user.ID)
  records := buildRecordStats(a, user.ID)

  maxWeekly := 0
  for _, item := range weekly {
    if item.Duration > maxWeekly {
      maxWeekly = item.Duration
    }
  }
  if maxWeekly == 0 {
    maxWeekly = 1
  }

  maxMonthly := 0
  for _, item := range monthly {
    if item.Workouts > maxMonthly {
      maxMonthly = item.Workouts
    }
  }
  if maxMonthly == 0 {
    maxMonthly = 1
  }

  points := []string{}
  if len(monthly) > 1 {
    for i, item := range monthly {
      x := float64(i) / float64(len(monthly)-1) * 300.0
      y := 100.0 - (float64(item.Workouts)/float64(maxMonthly))*80.0
      points = append(points, fmt.Sprintf("%.0f,%.0f", x, y))
    }
  }

  totalWorkouts := 0
  totalMinutes := 0
  _ = a.DB.QueryRow(
    `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
     from workout_sessions
     where user_id = $1 and completed_at is not null`,
    user.ID,
  ).Scan(&totalWorkouts, &totalMinutes)

  data := map[string]any{
    "WeeklyData": weekly,
    "MonthlyData": monthly,
    "CategoryData": categories,
    "Records": records,
    "TotalWorkouts": totalWorkouts,
    "TotalHours": float64(totalMinutes) / 60.0,
    "MaxWeekly": maxWeekly,
    "MaxMonthly": maxMonthly,
    "MonthlyPoints": strings.Join(points, " "),
  }

  a.renderFullPage(w, r, "statistics", "Статистика", data)
}

func buildWeeklyStats(a *App, userID string) []weeklyStat {
  now := time.Now()
  start := now.AddDate(0, 0, -6)
  days := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
  stats := []weeklyStat{}

  for i := 0; i < 7; i++ {
    date := start.AddDate(0, 0, i)
    var workouts int
    var minutes int
    _ = a.DB.QueryRow(
      `select coalesce(count(*), 0), coalesce(sum(duration_minutes), 0)
       from workout_sessions
       where user_id = $1 and date(completed_at) = $2`,
      userID,
      date.Format("2006-01-02"),
    ).Scan(&workouts, &minutes)

    stats = append(stats, weeklyStat{
      Day:      days[int(date.Weekday())],
      Workouts: workouts,
      Duration: minutes,
    })
  }

  return stats
}

func buildMonthlyTrends(a *App, userID string) []monthlyTrend {
  now := time.Now()
  months := []string{"Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"}
  trends := []monthlyTrend{}

  for i := 5; i >= 0; i-- {
    date := now.AddDate(0, -i, 0)
    monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)
    var workouts int
    _ = a.DB.QueryRow(
      `select coalesce(count(*), 0)
       from workout_sessions
       where user_id = $1 and completed_at >= $2 and completed_at < $3`,
      userID,
      monthStart,
      monthStart.AddDate(0, 1, 0),
    ).Scan(&workouts)

    trends = append(trends, monthlyTrend{
      Month:    months[int(date.Month())-1],
      Workouts: workouts,
    })
  }

  return trends
}

func buildCategoryStats(a *App, userID string) []categoryStat {
  rows, err := a.DB.Query(
    `select coalesce(w.category, 'Другое'), count(*)
     from workout_sessions ws
     join workouts w on w.id = ws.workout_id
     where ws.user_id = $1 and ws.completed_at is not null
     group by w.category`,
    userID,
  )
  if err != nil {
    return []categoryStat{}
  }
  defer rows.Close()

  total := 0
  type raw struct {
    category string
    count int
  }
  rawRows := []raw{}
  for rows.Next() {
    var item raw
    _ = rows.Scan(&item.category, &item.count)
    rawRows = append(rawRows, item)
    total += item.count
  }

  colors := []string{"blue", "purple", "green", "orange"}
  classes := []string{"bg-blue-500", "bg-purple-500", "bg-green-500", "bg-orange-500"}
  stats := []categoryStat{}
  for i, item := range rawRows {
    percent := 0
    if total > 0 {
      percent = int(float64(item.count) / float64(total) * 100)
    }
    stats = append(stats, categoryStat{
      Category: item.category,
      Percentage: percent,
      Color: colors[i%len(colors)],
      ColorClass: classes[i%len(classes)],
    })
  }

  return stats
}

func buildRecordStats(a *App, userID string) []recordStat {
  streak := computeWorkoutStreak(a.DB, userID)
  var maxDuration int
  var totalCalories int
  _ = a.DB.QueryRow(
    `select coalesce(max(duration_minutes), 0), coalesce(sum(calories_burned), 0)
     from workout_sessions
     where user_id = $1 and completed_at is not null`,
    userID,
  ).Scan(&maxDuration, &totalCalories)

  records := []recordStat{
    {Title: "Самая длинная серия", Value: formatDays(streak), IconClass: "text-yellow-800", BgClass: "bg-yellow-50"},
    {Title: "Самая долгая тренировка", Value: formatMinutes(maxDuration), IconClass: "text-purple-800", BgClass: "bg-purple-50"},
    {Title: "Всего калорий", Value: formatCalories(totalCalories), IconClass: "text-green-800", BgClass: "bg-green-50"},
  }

  return records
}

func formatDays(value int) string {
  if value == 0 {
    return "0 дней"
  }
  return fmt.Sprintf("%d дней подряд", value)
}

func formatMinutes(value int) string {
  if value == 0 {
    return "0 минут"
  }
  return fmt.Sprintf("%d минут", value)
}

func formatCalories(value int) string {
  if value == 0 {
    return "0 ккал"
  }
  return fmt.Sprintf("%d ккал", value)
}
