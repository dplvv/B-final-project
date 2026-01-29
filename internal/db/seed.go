package db

import (
  "database/sql"
  "errors"
  "fmt"
  "time"

  "golang.org/x/crypto/bcrypt"
)

type seedUser struct {
  Name       string
  EmployeeID string
  Role       string
  Department string
  Password   string
}

func Seed(db *sql.DB) error {
  users := []seedUser{
    {Name: "Иван Петров", EmployeeID: "10001", Role: "employee", Department: "Инженерный отдел", Password: "password"},
    {Name: "Мария Соколова", EmployeeID: "20001", Role: "manager", Department: "HR", Password: "password"},
    {Name: "Администратор", EmployeeID: "90000", Role: "admin", Department: "ИТ", Password: "password"},
  }

  userIDs := map[string]string{}
  for _, user := range users {
    id, err := ensureUser(db, user)
    if err != nil {
      return err
    }
    userIDs[user.EmployeeID] = id
    if err := EnsureUserDefaults(db, id); err != nil {
      return err
    }
  }

  if err := seedExercises(db); err != nil {
    return err
  }
  if err := seedWorkouts(db); err != nil {
    return err
  }
  if err := seedWorkoutExercises(db); err != nil {
    return err
  }
  if err := seedPrograms(db); err != nil {
    return err
  }
  if err := seedAchievements(db); err != nil {
    return err
  }
  if err := seedRecommendations(db); err != nil {
    return err
  }
  if err := seedVideos(db); err != nil {
    return err
  }
  if err := seedRewards(db); err != nil {
    return err
  }

  if employeeID, ok := userIDs["10001"]; ok {
    _ = seedSampleSessions(db, employeeID)
  }

  return nil
}

func ensureUser(db *sql.DB, user seedUser) (string, error) {
  var id string
  err := db.QueryRow("select id from users where employee_id = $1", user.EmployeeID).Scan(&id)
  if err == nil {
    return id, nil
  }
  if !errors.Is(err, sql.ErrNoRows) {
    return "", fmt.Errorf("lookup user: %w", err)
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
  if err != nil {
    return "", fmt.Errorf("hash password: %w", err)
  }

  err = db.QueryRow(
    `insert into users (name, employee_id, password_hash, role, department)
     values ($1, $2, $3, $4, $5)
     returning id`,
    user.Name,
    user.EmployeeID,
    string(hash),
    user.Role,
    user.Department,
  ).Scan(&id)
  if err != nil {
    return "", fmt.Errorf("insert user: %w", err)
  }

  return id, nil
}

func EnsureUserDefaults(db *sql.DB, userID string) error {
  _, _ = db.Exec("insert into user_profiles (user_id) values ($1) on conflict do nothing", userID)
  _, _ = db.Exec("insert into user_settings (user_id) values ($1) on conflict do nothing", userID)
  _, _ = db.Exec("insert into medical_info (user_id) values ($1) on conflict do nothing", userID)
  _, _ = db.Exec("insert into user_points (user_id) values ($1) on conflict do nothing", userID)
  return nil
}

func seedExercises(db *sql.DB) error {
  exercises := []struct {
    Name        string
    Description string
    Category    string
    Difficulty  string
    Sets        int
    Reps        string
    Rest        int
    Muscles     []string
    Equipment   []string
    VideoURL    string
  }{
    {"Планка", "Укрепляет мышцы кора и спины.", "Кор", "Средняя", 3, "30-45 сек", 45, []string{"Пресс", "Спина"}, []string{"Коврик"}, "https://www.youtube.com/embed/pSHjTRCQxIw"},
    {"Супермен", "Лежа на животе, поднимайте руки и ноги.", "Спина", "Легкая", 3, "12-15", 30, []string{"Спина"}, []string{"Коврик"}, "https://www.youtube.com/embed/cc6UVRS7PW4"},
    {"Ягодичный мост", "Поднимайте таз, удерживая пресс.", "Ноги", "Легкая", 3, "12-15", 30, []string{"Ягодицы"}, []string{"Коврик"}, "https://www.youtube.com/embed/m2Zx-57cSok"},
    {"Разведение рук с эспандером", "Контролируемое разведение рук.", "Плечи", "Средняя", 3, "12-15", 30, []string{"Плечи"}, []string{"Эспандер"}, "https://www.youtube.com/embed/3Vv2t0z3tQY"},
    {"Приседания к стулу", "Приседайте до касания стула.", "Ноги", "Легкая", 3, "10-12", 45, []string{"Ноги"}, []string{"Стул"}, "https://www.youtube.com/embed/YaXPRqUwItQ"},
    {"Растяжка груди", "Растяните грудные мышцы у стены.", "Растяжка", "Легкая", 2, "20-30 сек", 20, []string{"Грудь"}, []string{"Стена"}, "https://www.youtube.com/embed/tJt4hQ9x30E"},
    {"Кошка-корова", "Мобилизация позвоночника в спокойном темпе.", "Мобилизация", "Легкая", 2, "8-10", 20, []string{"Спина"}, []string{"Коврик"}, "https://www.youtube.com/embed/kqnua4rHVVA"},
    {"Повороты корпуса сидя", "Мягкое вращение грудного отдела.", "Мобилизация", "Легкая", 2, "10-12", 20, []string{"Спина", "Кор"}, []string{"Стул"}, "https://www.youtube.com/embed/0BhfKxK1uK8"},
    {"Ходьба на месте", "Легкий разогрев, поддерживает пульс.", "Кардио", "Легкая", 2, "60 сек", 30, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/0f0kQqLZcW4"},
    {"Выпады назад", "Укрепление ног и баланса.", "Ноги", "Средняя", 3, "10-12", 45, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/3D2WQF6kQDc"},
    {"Растяжка задней поверхности бедра", "Мягкая растяжка сидя.", "Растяжка", "Легкая", 2, "20-30 сек", 20, []string{"Ноги"}, []string{"Коврик"}, "https://www.youtube.com/embed/2ZlK7VqkX1Y"},
    {"Подъемы на носки", "Укрепление голеней.", "Ноги", "Легкая", 3, "12-15", 30, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/-M4-G8p8fmc"},
  }

  for _, ex := range exercises {
    _, err := db.Exec(
      `insert into exercises (name, description, category, difficulty, sets, reps, rest_seconds, muscle_groups, equipment, video_url)
       values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
       on conflict do nothing`,
      ex.Name,
      ex.Description,
      ex.Category,
      ex.Difficulty,
      ex.Sets,
      ex.Reps,
      ex.Rest,
      ex.Muscles,
      ex.Equipment,
      ex.VideoURL,
    )
    if err != nil {
      return fmt.Errorf("seed exercises: %w", err)
    }
  }

  return nil
}

func seedWorkouts(db *sql.DB) error {
  workouts := []struct {
    Name        string
    Description string
    Duration    int
    Difficulty  string
    Category    string
  }{
    {"Базовая реабилитация", "Укрепление мышц спины и коррекция осанки", 30, "Средняя", "Реабилитация"},
    {"Мягкая мобилизация", "Легкая разминка для суставов", 20, "Легкая", "Мобилизация"},
    {"Стабилизация корпуса", "Тренировка мышц кора", 25, "Средняя", "Кор"},
    {"Офисная разминка", "Снятие напряжения после рабочего дня", 15, "Легкая", "Мобилизация"},
    {"Разгрузка спины", "Упражнения для спины и осанки", 25, "Легкая", "Спина"},
    {"Легкое кардио", "Мягкая кардио-нагрузка и разогрев", 20, "Легкая", "Кардио"},
    {"Нижняя часть тела", "Укрепление ног и баланса", 30, "Средняя", "Ноги"},
    {"Гибкость и растяжка", "Расслабление и гибкость всего тела", 25, "Легкая", "Растяжка"},
  }

  for _, w := range workouts {
    _, err := db.Exec(
      `insert into workouts (name, description, duration_minutes, difficulty, category)
       values ($1, $2, $3, $4, $5)
       on conflict do nothing`,
      w.Name,
      w.Description,
      w.Duration,
      w.Difficulty,
      w.Category,
    )
    if err != nil {
      return fmt.Errorf("seed workouts: %w", err)
    }
  }

  return nil
}

func seedWorkoutExercises(db *sql.DB) error {
  exerciseIDs := map[string]string{}
  rows, err := db.Query(`select id, name from exercises`)
  if err != nil {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    var id string
    var name string
    if err := rows.Scan(&id, &name); err != nil {
      return err
    }
    exerciseIDs[name] = id
  }

  workoutIDs := map[string]string{}
  workoutRows, err := db.Query(`select id, name from workouts`)
  if err != nil {
    return err
  }
  defer workoutRows.Close()
  for workoutRows.Next() {
    var id string
    var name string
    if err := workoutRows.Scan(&id, &name); err != nil {
      return err
    }
    workoutIDs[name] = id
  }

  type item struct {
    workout string
    exercise string
    order int
  }

  items := []item{
    {workout: "Базовая реабилитация", exercise: "Планка", order: 1},
    {workout: "Базовая реабилитация", exercise: "Супермен", order: 2},
    {workout: "Базовая реабилитация", exercise: "Ягодичный мост", order: 3},
    {workout: "Мягкая мобилизация", exercise: "Растяжка груди", order: 1},
    {workout: "Мягкая мобилизация", exercise: "Приседания к стулу", order: 2},
    {workout: "Стабилизация корпуса", exercise: "Планка", order: 1},
    {workout: "Стабилизация корпуса", exercise: "Разведение рук с эспандером", order: 2},
    {workout: "Стабилизация корпуса", exercise: "Супермен", order: 3},
    {workout: "Офисная разминка", exercise: "Кошка-корова", order: 1},
    {workout: "Офисная разминка", exercise: "Повороты корпуса сидя", order: 2},
    {workout: "Офисная разминка", exercise: "Растяжка груди", order: 3},
    {workout: "Разгрузка спины", exercise: "Супермен", order: 1},
    {workout: "Разгрузка спины", exercise: "Кошка-корова", order: 2},
    {workout: "Разгрузка спины", exercise: "Планка", order: 3},
    {workout: "Легкое кардио", exercise: "Ходьба на месте", order: 1},
    {workout: "Легкое кардио", exercise: "Подъемы на носки", order: 2},
    {workout: "Нижняя часть тела", exercise: "Приседания к стулу", order: 1},
    {workout: "Нижняя часть тела", exercise: "Выпады назад", order: 2},
    {workout: "Нижняя часть тела", exercise: "Ягодичный мост", order: 3},
    {workout: "Гибкость и растяжка", exercise: "Растяжка груди", order: 1},
    {workout: "Гибкость и растяжка", exercise: "Растяжка задней поверхности бедра", order: 2},
  }

  for _, item := range items {
    workoutID := workoutIDs[item.workout]
    exerciseID := exerciseIDs[item.exercise]
    if workoutID == "" || exerciseID == "" {
      continue
    }
    _, _ = db.Exec(
      `insert into workout_exercises (workout_id, exercise_id, sort_order)
       values ($1, $2, $3)
       on conflict do nothing`,
      workoutID,
      exerciseID,
      item.order,
    )
  }

  return nil
}

func seedPrograms(db *sql.DB) error {
  programs := []struct {
    Name        string
    Description string
    Workouts    []string
  }{
    {
      Name:        "Персональная программа",
      Description: "План на 4 недели с постепенным увеличением нагрузки",
      Workouts:    []string{"Базовая реабилитация", "Стабилизация корпуса", "Разгрузка спины"},
    },
    {
      Name:        "Офисная мобилизация",
      Description: "Короткие тренировки для восстановления после работы",
      Workouts:    []string{"Офисная разминка", "Мягкая мобилизация", "Гибкость и растяжка"},
    },
    {
      Name:        "Легкое кардио",
      Description: "Мягкая кардио-нагрузка без перегрузки суставов",
      Workouts:    []string{"Легкое кардио", "Офисная разминка", "Гибкость и растяжка"},
    },
  }

  workoutIDs := map[string]string{}
  rows, err := db.Query(`select id, name from workouts`)
  if err != nil {
    return err
  }
  defer rows.Close()
  for rows.Next() {
    var id string
    var name string
    if err := rows.Scan(&id, &name); err != nil {
      return err
    }
    workoutIDs[name] = id
  }

  for _, p := range programs {
    var programID string
    err := db.QueryRow("select id from programs where name = $1", p.Name).Scan(&programID)
    if err != nil {
      if !errors.Is(err, sql.ErrNoRows) {
        return err
      }
      err = db.QueryRow(
        `insert into programs (name, description) values ($1, $2) returning id`,
        p.Name,
        p.Description,
      ).Scan(&programID)
      if err != nil {
        return fmt.Errorf("insert program: %w", err)
      }
    }

    order := 1
    for _, w := range p.Workouts {
      workoutID := workoutIDs[w]
      if workoutID == "" {
        continue
      }
      _, _ = db.Exec(
        `insert into program_workouts (program_id, workout_id, sort_order)
         values ($1, $2, $3)
         on conflict do nothing`,
        programID,
        workoutID,
        order,
      )
      order++
    }
  }

  return nil
}

func seedAchievements(db *sql.DB) error {
  achievements := []struct {
    Title       string
    Description string
    Icon        string
  }{
    {"Первый шаг", "Завершите первую тренировку", "spark"},
    {"Серия", "5 тренировок подряд", "flame"},
    {"Настойчивость", "10 тренировок за месяц", "shield"},
  }

  for _, a := range achievements {
    _, err := db.Exec(
      `insert into achievements (title, description, icon)
       values ($1, $2, $3)
       on conflict do nothing`,
      a.Title,
      a.Description,
      a.Icon,
    )
    if err != nil {
      return fmt.Errorf("seed achievements: %w", err)
    }
  }

  return nil
}

func seedRecommendations(db *sql.DB) error {
  recommendations := []struct {
    Title string
    Body  string
    Category string
    Icon string
    Excerpt string
    ReadTime int
  }{
    {"Правильная техника выполнения планки", "Планка является базовым упражнением для развития силы мышц кора и стабилизаторов спины. Удерживайте корпус в нейтральном положении и контролируйте дыхание.", "Техника", "🏋️", "Планка - одно из самых эффективных упражнений для укрепления кора", 5},
    {"Важность разминки перед тренировкой", "Разминка подготавливает мышцы и суставы к нагрузке и снижает риск травм. Начните с легкой кардио активности и динамических движений.", "Безопасность", "🤸", "Правильная разминка снижает риск травм и повышает эффективность", 4},
    {"Управление стрессом через движение", "Регулярные упражнения снижают уровень кортизола и улучшают настроение. Даже короткая прогулка повышает уровень энергии.", "Психология", "🧘", "Как физическая активность помогает справляться со стрессом", 6},
    {"Профилактика болей в спине", "Сидячий образ жизни является основной причиной болей в спине. Регулярно выполняйте упражнения на укрепление мышц спины.", "Здоровье", "🦴", "Упражнения для укрепления спины и улучшения осанки", 8},
    {"Дыхательные техники при упражнениях", "Правильное дыхание повышает эффективность упражнений. Выдыхайте на усилии и не задерживайте дыхание.", "Техника", "💨", "Как правильно дышать во время различных упражнений", 5},
  }

  for _, r := range recommendations {
    _, err := db.Exec(
      `insert into recommendations (title, body, category, icon, excerpt, read_time)
       values ($1, $2, $3, $4, $5, $6)
       on conflict do nothing`,
      r.Title,
      r.Body,
      r.Category,
      r.Icon,
      r.Excerpt,
      r.ReadTime,
    )
    if err != nil {
      return fmt.Errorf("seed recommendations: %w", err)
    }
  }

  return nil
}

func seedVideos(db *sql.DB) error {
  videos := []struct {
    Title       string
    Description string
    Duration    int
    Category    string
    Difficulty  string
    URL         string
  }{
    {"Разминка для шеи", "Мягкая разминка для офисных сотрудников.", 8, "Разминка", "Легкая", "https://example.com/video1"},
    {"Укрепление спины", "Базовые упражнения для спины.", 12, "Спина", "Средняя", "https://example.com/video2"},
    {"Стретчинг после работы", "Расслабление и восстановление.", 10, "Растяжка", "Легкая", "https://example.com/video3"},
  }

  for _, v := range videos {
    _, err := db.Exec(
      `insert into video_tutorials (title, description, duration_minutes, category, difficulty, url)
       values ($1, $2, $3, $4, $5, $6)
       on conflict do nothing`,
      v.Title,
      v.Description,
      v.Duration,
      v.Category,
      v.Difficulty,
      v.URL,
    )
    if err != nil {
      return fmt.Errorf("seed videos: %w", err)
    }
  }

  return nil
}

func seedRewards(db *sql.DB) error {
  rewards := []struct {
    Title       string
    Description string
    Cost        int
    Category    string
  }{
    {"Дополнительный выходной", "Бонус за регулярные тренировки.", 300, "Премирование"},
    {"Мерч компании", "Фирменная футболка.", 120, "Мотивация"},
    {"Сертификат спортмагазина", "Скидка на спортивные товары.", 200, "Бонус"},
  }

  for _, reward := range rewards {
    _, err := db.Exec(
      `insert into rewards (title, description, points_cost, category)
       values ($1, $2, $3, $4)
       on conflict do nothing`,
      reward.Title,
      reward.Description,
      reward.Cost,
      reward.Category,
    )
    if err != nil {
      return fmt.Errorf("seed rewards: %w", err)
    }
  }

  return nil
}

func seedSampleSessions(db *sql.DB, userID string) error {
  var workoutID string
  err := db.QueryRow("select id from workouts order by created_at limit 1").Scan(&workoutID)
  if err != nil {
    return err
  }

  for i := 0; i < 3; i++ {
    started := time.Now().AddDate(0, 0, -7+i*2)
    completed := started.Add(35 * time.Minute)
    _, err := db.Exec(
      `insert into workout_sessions (user_id, workout_id, started_at, completed_at, duration_minutes, total_exercises, completed_exercises, calories_burned)
       values ($1, $2, $3, $4, $5, $6, $7, $8)`,
      userID,
      workoutID,
      started,
      completed,
      30,
      6,
      6,
      240,
    )
    if err != nil {
      return err
    }
  }

  _, _ = db.Exec("update user_points set points_balance = 180, points_total = 180 where user_id = $1", userID)
  _, _ = db.Exec(
    `update user_profiles
     set fitness_level = 'beginner', restrictions = '{back}', goals = '{rehab}'
     where user_id = $1`,
    userID,
  )
  return nil
}
