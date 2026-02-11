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
    {"Планка", "Укрепляет мышцы кора и спины. Держите нейтральное положение и активируйте пресс.", "Кор", "Средняя", 3, "30-45 сек", 45, []string{"Пресс", "Спина"}, []string{"Коврик"}, "https://www.youtube.com/embed/pSHjTRCQxIw"},
    {"Супермен", "Лежа на животе, поднимайте руки и ноги. Подъем плавный, без рывков.", "Спина", "Легкая", 3, "12-15", 30, []string{"Спина"}, []string{"Коврик"}, "https://www.youtube.com/embed/cc6UVRS7PW4"},
    {"Ягодичный мост", "Поднимайте таз, удерживая пресс. Сведите лопатки и не переразгибайте поясницу.", "Ноги", "Легкая", 3, "12-15", 30, []string{"Ягодицы"}, []string{"Коврик"}, "https://www.youtube.com/embed/m2Zx-57cSok"},
    {"Разведение рук с эспандером", "Контролируемое разведение рук. Лопатки сведены, движение плавное.", "Плечи", "Средняя", 3, "12-15", 30, []string{"Плечи"}, []string{"Эспандер"}, "https://www.youtube.com/embed/3Vv2t0z3tQY"},
    {"Приседания к стулу", "Приседайте до касания стула. Колени смотрят по линии стоп.", "Ноги", "Легкая", 3, "10-12", 45, []string{"Ноги"}, []string{"Стул"}, "https://www.youtube.com/embed/YaXPRqUwItQ"},
    {"Растяжка груди", "Растяните грудные мышцы у стены. Дыхание спокойное, без боли.", "Растяжка", "Легкая", 2, "20-30 сек", 20, []string{"Грудь"}, []string{"Стена"}, "https://www.youtube.com/embed/tJt4hQ9x30E"},
    {"Кошка-корова", "Мобилизация позвоночника в спокойном темпе. Движение синхронизируйте с дыханием.", "Мобилизация", "Легкая", 2, "8-10", 20, []string{"Спина"}, []string{"Коврик"}, "https://www.youtube.com/embed/kqnua4rHVVA"},
    {"Повороты корпуса сидя", "Мягкое вращение грудного отдела. Не скручивайте поясницу.", "Мобилизация", "Легкая", 2, "10-12", 20, []string{"Спина", "Кор"}, []string{"Стул"}, "https://www.youtube.com/embed/0BhfKxK1uK8"},
    {"Ходьба на месте", "Легкий разогрев, поддерживает пульс. Следите за ровным дыханием.", "Кардио", "Легкая", 2, "60 сек", 30, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/0f0kQqLZcW4"},
    {"Выпады назад", "Укрепление ног и баланса. Спина ровная, шаг контролируемый.", "Ноги", "Средняя", 3, "10-12", 45, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/3D2WQF6kQDc"},
    {"Растяжка задней поверхности бедра", "Мягкая растяжка сидя. Спина прямая, наклон от таза.", "Растяжка", "Легкая", 2, "20-30 сек", 20, []string{"Ноги"}, []string{"Коврик"}, "https://www.youtube.com/embed/2ZlK7VqkX1Y"},
    {"Подъемы на носки", "Укрепление голеней. Движение медленное, пятки контролируемые.", "Ноги", "Легкая", 3, "12-15", 30, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/-M4-G8p8fmc"},
    {"Боковая планка", "Удержание корпуса на боку для укрепления косых мышц. Не заваливайте таз.", "Кор", "Средняя", 3, "20-30 сек", 40, []string{"Пресс", "Боковые мышцы"}, []string{"Коврик"}, "https://www.youtube.com/embed/K2VljzCC16g"},
    {"Скручивания", "Классическое упражнение для пресса с контролем дыхания. Подбородок слегка прижат.", "Кор", "Легкая", 3, "12-15", 30, []string{"Пресс"}, []string{"Коврик"}, "https://www.youtube.com/embed/Xyd_fa5zoEU"},
    {"Мертвый жук", "Стабилизация корпуса с попеременным движением рук и ног. Поясница прижата.", "Кор", "Легкая", 3, "10-12", 30, []string{"Пресс", "Кор"}, []string{"Коврик"}, "https://www.youtube.com/embed/4jHRx3D9uj0"},
    {"Тяга резинки к поясу", "Тяга эспандера к поясу для мышц спины. Лопатки сводите в конце.", "Спина", "Средняя", 3, "12-15", 40, []string{"Спина", "Лопатки"}, []string{"Резинка"}, "https://www.youtube.com/embed/RC6uov9XpsM"},
    {"Сведение лопаток стоя", "Сведение лопаток для улучшения осанки. Не поднимайте плечи.", "Спина", "Легкая", 3, "12-15", 30, []string{"Лопатки", "Спина"}, []string{}, "https://www.youtube.com/embed/0pJ1QF1wSXA"},
    {"Подъемы рук с гантелями", "Контролируемый подъем рук в стороны. Амплитуда до уровня плеч.", "Плечи", "Средняя", 3, "10-12", 40, []string{"Плечи"}, []string{"Гантели"}, "https://www.youtube.com/embed/q5sNYB1gR5I"},
    {"Растяжка плеч у стены", "Мягкая растяжка плечевого пояса у стены. Дышите ровно.", "Растяжка", "Легкая", 2, "20-30 сек", 20, []string{"Плечи"}, []string{"Стена"}, "https://www.youtube.com/embed/1b7g8KZVafk"},
    {"Мобилизация грудного отдела на ролле", "Разгибание грудного отдела на массажном ролле. Не давите на поясницу.", "Мобилизация", "Легкая", 2, "8-10", 20, []string{"Грудной отдел", "Спина"}, []string{"Ролл"}, "https://www.youtube.com/embed/p6Q3F5R9D9M"},
    {"Баланс на одной ноге", "Укрепление баланса и стабилизаторов. Фокус на устойчивости и ровной осанке.", "Мобилизация", "Легкая", 2, "30 сек", 20, []string{"Ноги", "Кор"}, []string{}, "https://www.youtube.com/embed/4K9Uq2i5Tz0"},
    {"Подъем коленей стоя", "Подъем коленей для мягкой кардио-нагрузки. Поддерживайте темп.", "Кардио", "Легкая", 3, "40 сек", 30, []string{"Ноги"}, []string{}, "https://www.youtube.com/embed/2M4dL1F7t5U"},
    {"Ягодичный мост с резинкой", "Мост с акцентом на ягодицы и стабилизацию. Сохраняйте нейтральную спину.", "Ноги", "Средняя", 3, "12-15", 40, []string{"Ягодицы", "Ноги"}, []string{"Резинка", "Коврик"}, "https://www.youtube.com/embed/TVhL41hFrkY"},
    {"Шаги на платформу", "Подъемы на платформу для тонуса и кардио. Полная опора всей стопой.", "Кардио", "Средняя", 3, "40 сек", 40, []string{"Ноги"}, []string{"Степ-платформа"}, "https://www.youtube.com/embed/7g6XQfGZ4qg"},
    {"Отведение ноги назад с резинкой", "Укрепление ягодиц с эспандером. Корпус устойчив.", "Ноги", "Средняя", 3, "12-15", 40, []string{"Ягодицы"}, []string{"Резинка"}, "https://www.youtube.com/embed/3l4BH7Tg0sA"},
    {"Тяга гантели в наклоне", "Укрепление спины и широчайших. Спина ровная, локоть вдоль корпуса.", "Спина", "Средняя", 3, "10-12", 40, []string{"Спина"}, []string{"Гантель", "Стул"}, "https://www.youtube.com/embed/VtE6pG4hQ4Q"},
    {"Растяжка икр у стены", "Растяжка икроножных мышц с опорой. Пятка на полу.", "Растяжка", "Легкая", 2, "20-30 сек", 20, []string{"Икры"}, []string{"Стена"}, "https://www.youtube.com/embed/8YsXKc7LDU8"},
    {"Динамический шаг в сторону", "Легкая динамика для тазобедренных суставов. Движения плавные.", "Мобилизация", "Легкая", 2, "10-12", 20, []string{"Ноги", "Тазобедренные"}, []string{}, "https://www.youtube.com/embed/9T0VY6sT1Tg"},
  }

  for _, ex := range exercises {
    _, err := db.Exec(
      `insert into exercises (name, description, category, difficulty, sets, reps, rest_seconds, muscle_groups, equipment, video_url)
       values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
       on conflict (name) do nothing`,
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
    {"Базовая реабилитация", "Укрепление мышц спины и коррекция осанки. Мягкий темп, акцент на безопасную технику.", 30, "Средняя", "Реабилитация"},
    {"Мягкая мобилизация", "Легкая разминка для суставов и позвоночника с плавными амплитудами.", 20, "Легкая", "Мобилизация"},
    {"Стабилизация корпуса", "Тренировка мышц кора для устойчивости и контроля положения тела.", 25, "Средняя", "Кор"},
    {"Офисная разминка", "Снятие напряжения после рабочего дня: мягкая мобилизация и растяжка.", 15, "Легкая", "Мобилизация"},
    {"Разгрузка спины", "Упражнения для спины и осанки с акцентом на контроль дыхания.", 25, "Легкая", "Спина"},
    {"Легкое кардио", "Мягкая кардио-нагрузка и разогрев без перегрузки суставов.", 20, "Легкая", "Кардио"},
    {"Нижняя часть тела", "Укрепление ног и баланса, развитие устойчивости и опоры.", 30, "Средняя", "Ноги"},
    {"Гибкость и растяжка", "Расслабление и гибкость всего тела, восстановление подвижности.", 25, "Легкая", "Растяжка"},
    {"Баланс и устойчивость", "Развитие стабильности и баланса корпуса, тренировка контроля.", 20, "Легкая", "Кор"},
    {"Плечевой пояс", "Укрепление плеч и снятие напряжения, работа с лопатками.", 25, "Средняя", "Плечи"},
    {"Силовая осанка", "Укрепление мышц спины и лопаток, улучшение вертикали корпуса.", 30, "Средняя", "Спина"},
    {"Мобилизация грудного отдела", "Мягкая мобилизация грудного отдела для улучшения подвижности.", 20, "Легкая", "Мобилизация"},
    {"Кардио и тонус", "Динамическая сессия для энергии и выносливости с контролем пульса.", 25, "Средняя", "Кардио"},
    {"Укрепление ног", "Укрепление ягодиц и стабилизаторов, развитие опоры.", 30, "Средняя", "Ноги"},
    {"Кор и пресс", "Фокус на мышцах кора и пресса, укрепление центра тела.", 20, "Легкая", "Кор"},
  }

  for _, w := range workouts {
    _, err := db.Exec(
      `insert into workouts (name, description, duration_minutes, difficulty, category)
       values ($1, $2, $3, $4, $5)
       on conflict (name) do nothing`,
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
    {workout: "Баланс и устойчивость", exercise: "Баланс на одной ноге", order: 1},
    {workout: "Баланс и устойчивость", exercise: "Боковая планка", order: 2},
    {workout: "Баланс и устойчивость", exercise: "Мертвый жук", order: 3},
    {workout: "Плечевой пояс", exercise: "Разведение рук с эспандером", order: 1},
    {workout: "Плечевой пояс", exercise: "Подъемы рук с гантелями", order: 2},
    {workout: "Плечевой пояс", exercise: "Растяжка плеч у стены", order: 3},
    {workout: "Силовая осанка", exercise: "Тяга резинки к поясу", order: 1},
    {workout: "Силовая осанка", exercise: "Сведение лопаток стоя", order: 2},
    {workout: "Силовая осанка", exercise: "Супермен", order: 3},
    {workout: "Мобилизация грудного отдела", exercise: "Мобилизация грудного отдела на ролле", order: 1},
    {workout: "Мобилизация грудного отдела", exercise: "Повороты корпуса сидя", order: 2},
    {workout: "Мобилизация грудного отдела", exercise: "Кошка-корова", order: 3},
    {workout: "Кардио и тонус", exercise: "Ходьба на месте", order: 1},
    {workout: "Кардио и тонус", exercise: "Подъем коленей стоя", order: 2},
    {workout: "Кардио и тонус", exercise: "Шаги на платформу", order: 3},
    {workout: "Укрепление ног", exercise: "Приседания к стулу", order: 1},
    {workout: "Укрепление ног", exercise: "Отведение ноги назад с резинкой", order: 2},
    {workout: "Укрепление ног", exercise: "Ягодичный мост с резинкой", order: 3},
    {workout: "Кор и пресс", exercise: "Планка", order: 1},
    {workout: "Кор и пресс", exercise: "Скручивания", order: 2},
    {workout: "Кор и пресс", exercise: "Боковая планка", order: 3},
  }

  workoutHasExercises := map[string]bool{}
  for _, item := range items {
    workoutID := workoutIDs[item.workout]
    exerciseID := exerciseIDs[item.exercise]
    if workoutID == "" || exerciseID == "" {
      continue
    }
    if _, checked := workoutHasExercises[workoutID]; !checked {
      var count int
      _ = db.QueryRow(`select count(*) from workout_exercises where workout_id = $1`, workoutID).Scan(&count)
      workoutHasExercises[workoutID] = count > 0
    }
    if workoutHasExercises[workoutID] {
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
    Muscles     []string
  }{
    {
      Name:        "Персональная программа",
      Description: "План на 4 недели с постепенным увеличением нагрузки, акцентом на стабилизацию и безопасную мобильность.",
      Workouts:    []string{"Базовая реабилитация", "Стабилизация корпуса", "Разгрузка спины"},
      Muscles:     []string{"Спина", "Кор"},
    },
    {
      Name:        "Офисная мобилизация",
      Description: "Короткие комплексы для снятия напряжения, улучшения осанки и подвижности суставов.",
      Workouts:    []string{"Офисная разминка", "Мягкая мобилизация", "Гибкость и растяжка"},
      Muscles:     []string{"Спина", "Плечи"},
    },
    {
      Name:        "Легкое кардио",
      Description: "Мягкая кардио-нагрузка без перегрузки суставов, с обязательной растяжкой.",
      Workouts:    []string{"Легкое кардио", "Офисная разминка", "Гибкость и растяжка"},
      Muscles:     []string{"Ноги", "Кардио"},
    },
    {
      Name:        "Сильная спина",
      Description: "Фокус на мышцах спины и лопаточного пояса, улучшение осанки и устойчивости.",
      Workouts:    []string{"Разгрузка спины", "Стабилизация корпуса", "Базовая реабилитация"},
      Muscles:     []string{"Спина", "Кор", "Плечи"},
    },
    {
      Name:        "Гибкость и баланс",
      Description: "Развитие подвижности и устойчивости, мягкие упражнения на баланс.",
      Workouts:    []string{"Гибкость и растяжка", "Мягкая мобилизация", "Офисная разминка"},
      Muscles:     []string{"Растяжка", "Спина", "Ноги"},
    },
    {
      Name:        "Ноги и устойчивость",
      Description: "Укрепление нижней части тела с безопасными упражнениями на баланс.",
      Workouts:    []string{"Нижняя часть тела", "Легкое кардио", "Гибкость и растяжка"},
      Muscles:     []string{"Ноги", "Кор"},
    },
    {
      Name:        "Мобилизация плеч",
      Description: "Снятие напряжения в плечах и грудном отделе, контроль амплитуды движений.",
      Workouts:    []string{"Офисная разминка", "Мягкая мобилизация", "Разгрузка спины"},
      Muscles:     []string{"Плечи", "Спина"},
    },
    {
      Name:        "Кардио и энергия",
      Description: "Комбинация кардио и растяжки для повышения тонуса и выносливости.",
      Workouts:    []string{"Кардио и тонус", "Легкое кардио", "Гибкость и растяжка"},
      Muscles:     []string{"Ноги", "Кардио"},
    },
    {
      Name:        "Осанка и спина",
      Description: "Комплекс для укрепления спины и улучшения осанки.",
      Workouts:    []string{"Силовая осанка", "Разгрузка спины", "Стабилизация корпуса"},
      Muscles:     []string{"Спина", "Кор"},
    },
    {
      Name:        "Баланс и контроль",
      Description: "Упражнения на устойчивость и контроль движений.",
      Workouts:    []string{"Баланс и устойчивость", "Мягкая мобилизация", "Гибкость и растяжка"},
      Muscles:     []string{"Кор", "Ноги"},
    },
    {
      Name:        "Плечи без боли",
      Description: "Снятие напряжения в плечах и безопасное укрепление.",
      Workouts:    []string{"Плечевой пояс", "Мобилизация плеч", "Офисная разминка"},
      Muscles:     []string{"Плечи", "Спина"},
    },
    {
      Name:        "Ноги в тонусе",
      Description: "Укрепление нижней части тела и устойчивости.",
      Workouts:    []string{"Укрепление ног", "Нижняя часть тела", "Гибкость и растяжка"},
      Muscles:     []string{"Ноги", "Кор"},
    },
    {
      Name:        "Кор и стабильность",
      Description: "Укрепление мышц кора, улучшение контроля и стабилизации корпуса.",
      Workouts:    []string{"Кор и пресс", "Стабилизация корпуса", "Баланс и устойчивость"},
      Muscles:     []string{"Кор", "Пресс"},
    },
    {
      Name:        "Мягкая растяжка",
      Description: "Спокойная программа для восстановления подвижности и расслабления.",
      Workouts:    []string{"Гибкость и растяжка", "Мягкая мобилизация", "Мобилизация грудного отдела"},
      Muscles:     []string{"Растяжка", "Спина"},
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
    inserted := false
    err := db.QueryRow("select id from programs where name = $1", p.Name).Scan(&programID)
    if err != nil {
      if !errors.Is(err, sql.ErrNoRows) {
        return err
      }
      err = db.QueryRow(
        `insert into programs (name, description, muscle_groups) values ($1, $2, $3) returning id`,
        p.Name,
        p.Description,
        p.Muscles,
      ).Scan(&programID)
      if err != nil {
        return fmt.Errorf("insert program: %w", err)
      }
      inserted = true
    }

    if !inserted {
      continue
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
    Points      int
  }{
    {"Первый шаг", "Завершите первую тренировку", "👣", 20},
    {"Первые три", "Завершите 3 тренировки", "⭐", 30},
    {"Серия", "5 тренировок подряд", "🔥", 40},
    {"Железная воля", "10 тренировок подряд", "⚡", 70},
    {"Настойчивость", "10 тренировок за месяц", "🛡️", 60},
    {"Регулярность", "8 тренировок за месяц", "📅", 50},
    {"Месяц активности", "20 тренировок за месяц", "🏅", 120},
    {"Марафон", "25 тренировок всего", "🏆", 100},
  }

  for _, a := range achievements {
    _, err := db.Exec(
      `insert into achievements (title, description, icon, points_reward)
       values ($1, $2, $3, $4)
       on conflict (title) do nothing`,
      a.Title,
      a.Description,
      a.Icon,
      a.Points,
    )
    if err != nil {
      return fmt.Errorf("seed achievements: %w", err)
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
    {"Сертификат кафе", "Небольшой подарок за активность.", 150, "Бонус"},
    {"Сеанс массажа", "Восстановление после тренировок.", 250, "Здоровье"},
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
     set fitness_level = 'Легкая', goals = '{Восстановление}'
     where user_id = $1`,
    userID,
  )
  _, _ = db.Exec(
    `update medical_info
     set restrictions = '{Спина}'
     where user_id = $1`,
    userID,
  )
  return nil
}
