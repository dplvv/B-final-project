package site

import (
  "net/http"
  "strconv"
  "strings"

  "github.com/go-chi/chi/v5"
)

func (s *Site) adminExercises(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Упражнения", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  rows, err := s.DB.Query(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(rest_seconds, 0),
            coalesce(duration_seconds, 0), coalesce(muscle_groups, '{}'), coalesce(equipment, '{}'), coalesce(video_url, '')
     from exercises
     order by name`,
  )
  exercises := []exerciseCard{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var ex exerciseCard
      _ = rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &ex.MuscleGroups, &ex.Equipment, &ex.VideoURL)
      ex.VideoURL = normalizeVideoURL(ex.VideoURL)
      exercises = append(exercises, ex)
    }
  }
  data["Exercises"] = exercises
  s.render(w, "admin_exercises", data)
}

func (s *Site) adminExerciseCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  name := strings.TrimSpace(r.FormValue("name"))
  description := strings.TrimSpace(r.FormValue("description"))
  category := strings.TrimSpace(r.FormValue("category"))
  difficulty := strings.TrimSpace(r.FormValue("difficulty"))
  sets, _ := strconv.Atoi(r.FormValue("sets"))
  reps := strings.TrimSpace(r.FormValue("reps"))
  rest, _ := strconv.Atoi(r.FormValue("rest_seconds"))
  muscles := parseCSV(r.FormValue("muscle_groups"))
  equipment := parseCSV(r.FormValue("equipment"))
  video := normalizeVideoURL(r.FormValue("video_url"))
  if name == "" || description == "" {
    http.Redirect(w, r, "/admin/exercises?error=Заполните%20название%20и%20описание", http.StatusSeeOther)
    return
  }

  _, err := s.DB.Exec(
    `insert into exercises (name, description, category, difficulty, sets, reps, rest_seconds, muscle_groups, equipment, video_url)
     values ($1, $2, nullif($3, ''), nullif($4, ''), $5, nullif($6, ''), $7, $8, $9, nullif($10, ''))
     on conflict (name)
     do update set description = excluded.description,
                   category = excluded.category,
                   difficulty = excluded.difficulty,
                   sets = excluded.sets,
                   reps = excluded.reps,
                   rest_seconds = excluded.rest_seconds,
                   muscle_groups = excluded.muscle_groups,
                   equipment = excluded.equipment,
                   video_url = excluded.video_url`,
    name,
    description,
    category,
    difficulty,
    sets,
    reps,
    rest,
    muscles,
    equipment,
    video,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/admin/exercises?success=Упражнение%20сохранено", http.StatusSeeOther)
}

func (s *Site) adminExerciseUpdate(w http.ResponseWriter, r *http.Request) {
  exerciseID := chi.URLParam(r, "id")
  if exerciseID == "" {
    http.Redirect(w, r, "/admin/exercises?error=Не%20найдено%20упражнение", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  name := strings.TrimSpace(r.FormValue("name"))
  description := strings.TrimSpace(r.FormValue("description"))
  category := strings.TrimSpace(r.FormValue("category"))
  difficulty := strings.TrimSpace(r.FormValue("difficulty"))
  sets, _ := strconv.Atoi(r.FormValue("sets"))
  reps := strings.TrimSpace(r.FormValue("reps"))
  rest, _ := strconv.Atoi(r.FormValue("rest_seconds"))
  muscles := parseCSV(r.FormValue("muscle_groups"))
  equipment := parseCSV(r.FormValue("equipment"))
  video := normalizeVideoURL(r.FormValue("video_url"))

  _, err := s.DB.Exec(
    `update exercises
     set name = coalesce(nullif($1, ''), name),
         description = coalesce(nullif($2, ''), description),
         category = nullif($3, ''),
         difficulty = nullif($4, ''),
         sets = $5,
         reps = nullif($6, ''),
         rest_seconds = $7,
         muscle_groups = $8,
         equipment = $9,
         video_url = nullif($10, ''),
         created_at = created_at
     where id = $11`,
    name,
    description,
    category,
    difficulty,
    sets,
    reps,
    rest,
    muscles,
    equipment,
    video,
    exerciseID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/exercises?success=Упражнение%20обновлено", http.StatusSeeOther)
}

func (s *Site) adminWorkouts(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Тренировки", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  rows, err := s.DB.Query(
    `select id, name, description, duration_minutes, difficulty, coalesce(category, '')
     from workouts
     order by name`,
  )
  workouts := []workoutCard{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var wCard workoutCard
      _ = rows.Scan(&wCard.ID, &wCard.Name, &wCard.Description, &wCard.Duration, &wCard.Difficulty, &wCard.Category)
      workouts = append(workouts, wCard)
    }
  }
  data["Workouts"] = workouts
  s.render(w, "admin_workouts", data)
}

func (s *Site) adminWorkoutCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/workouts?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  name := strings.TrimSpace(r.FormValue("name"))
  description := strings.TrimSpace(r.FormValue("description"))
  duration, _ := strconv.Atoi(r.FormValue("duration_minutes"))
  difficulty := strings.TrimSpace(r.FormValue("difficulty"))
  category := strings.TrimSpace(r.FormValue("category"))
  if name == "" || description == "" {
    http.Redirect(w, r, "/admin/workouts?error=Заполните%20название%20и%20описание", http.StatusSeeOther)
    return
  }
  if duration <= 0 {
    duration = 20
  }

  _, err := s.DB.Exec(
    `insert into workouts (name, description, duration_minutes, difficulty, category)
     values ($1, $2, $3, nullif($4, ''), nullif($5, ''))
     on conflict (name)
     do update set description = excluded.description,
                   duration_minutes = excluded.duration_minutes,
                   difficulty = excluded.difficulty,
                   category = excluded.category`,
    name,
    description,
    duration,
    difficulty,
    category,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/workouts?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/workouts?success=Тренировка%20сохранена", http.StatusSeeOther)
}

func (s *Site) adminWorkoutUpdate(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    http.Redirect(w, r, "/admin/workouts?error=Не%20найдена%20тренировка", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/workouts?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  name := strings.TrimSpace(r.FormValue("name"))
  description := strings.TrimSpace(r.FormValue("description"))
  duration, _ := strconv.Atoi(r.FormValue("duration_minutes"))
  difficulty := strings.TrimSpace(r.FormValue("difficulty"))
  category := strings.TrimSpace(r.FormValue("category"))
  if duration <= 0 {
    duration = 20
  }
  _, err := s.DB.Exec(
    `update workouts
     set name = coalesce(nullif($1, ''), name),
         description = coalesce(nullif($2, ''), description),
         duration_minutes = $3,
         difficulty = nullif($4, ''),
         category = nullif($5, '')
     where id = $6`,
    name,
    description,
    duration,
    difficulty,
    category,
    workoutID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/workouts?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/workouts?success=Тренировка%20обновлена", http.StatusSeeOther)
}

func (s *Site) adminWorkoutDetail(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    http.NotFound(w, r)
    return
  }
  var workout workoutCard
  err := s.DB.QueryRow(
    `select id, name, description, duration_minutes, difficulty, coalesce(category, '')
     from workouts where id = $1`,
    workoutID,
  ).Scan(&workout.ID, &workout.Name, &workout.Description, &workout.Duration, &workout.Difficulty, &workout.Category)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  rows, err := s.DB.Query(
    `select e.id, e.name, coalesce(we.sort_order, 1),
            coalesce(we.sets, e.sets, 1), coalesce(we.reps, e.reps, '10'), coalesce(we.rest_seconds, e.rest_seconds, 30)
     from workout_exercises we
     join exercises e on e.id = we.exercise_id
     where we.workout_id = $1
     order by we.sort_order`,
    workoutID,
  )
  type workoutExerciseView struct {
    ID    string
    Name  string
    Order int
    Sets  int
    Reps  string
    Rest  int
  }
  list := []workoutExerciseView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var v workoutExerciseView
      _ = rows.Scan(&v.ID, &v.Name, &v.Order, &v.Sets, &v.Reps, &v.Rest)
      list = append(list, v)
    }
  }

  exRows, err := s.DB.Query(`select id, name from exercises order by name`)
  exercises := []exerciseCard{}
  if err == nil {
    defer exRows.Close()
    for exRows.Next() {
      var ex exerciseCard
      _ = exRows.Scan(&ex.ID, &ex.Name)
      exercises = append(exercises, ex)
    }
  }

  data := s.baseData(r, workout.Name, "admin")
  data["Workout"] = workout
  data["WorkoutExercises"] = list
  data["Exercises"] = exercises
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "admin_workout_detail", data)
}

func (s *Site) adminWorkoutExerciseAdd(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    http.Redirect(w, r, "/admin/workouts", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/workouts/"+workoutID+"?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  exerciseID := r.FormValue("exercise_id")
  order, _ := strconv.Atoi(r.FormValue("sort_order"))
  sets, _ := strconv.Atoi(r.FormValue("sets"))
  reps := strings.TrimSpace(r.FormValue("reps"))
  rest, _ := strconv.Atoi(r.FormValue("rest_seconds"))
  if exerciseID == "" {
    http.Redirect(w, r, "/admin/workouts/"+workoutID+"?error=Выберите%20упражнение", http.StatusSeeOther)
    return
  }
  if order <= 0 {
    order = 1
  }

  _, err := s.DB.Exec(
    `insert into workout_exercises (workout_id, exercise_id, sort_order, sets, reps, rest_seconds)
     values ($1, $2, $3, nullif($4, 0), nullif($5, ''), nullif($6, 0))
     on conflict (workout_id, exercise_id)
     do update set sort_order = excluded.sort_order,
                   sets = excluded.sets,
                   reps = excluded.reps,
                   rest_seconds = excluded.rest_seconds`,
    workoutID,
    exerciseID,
    order,
    sets,
    reps,
    rest,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/workouts/"+workoutID+"?error=Не%20удалось%20добавить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/workouts/"+workoutID+"?success=Упражнение%20добавлено", http.StatusSeeOther)
}

func (s *Site) adminWorkoutExerciseRemove(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  exerciseID := chi.URLParam(r, "exerciseId")
  if workoutID == "" || exerciseID == "" {
    http.Redirect(w, r, "/admin/workouts", http.StatusSeeOther)
    return
  }
  _, _ = s.DB.Exec(`delete from workout_exercises where workout_id = $1 and exercise_id = $2`, workoutID, exerciseID)
  http.Redirect(w, r, "/admin/workouts/"+workoutID+"?success=Упражнение%20удалено", http.StatusSeeOther)
}

func (s *Site) adminPrograms(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Программы", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  rows, err := s.DB.Query(
    `select id, name, description, coalesce(muscle_groups, '{}')
     from programs
     order by created_at desc`,
  )
  programs := []programCard{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var p programCard
      _ = rows.Scan(&p.ID, &p.Name, &p.Description, &p.MuscleGroups)
      programs = append(programs, p)
    }
  }

  data["Programs"] = programs
  s.render(w, "admin_programs", data)
}

func (s *Site) adminProgramCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/programs?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  name := strings.TrimSpace(r.FormValue("name"))
  description := strings.TrimSpace(r.FormValue("description"))
  muscles := parseCSV(r.FormValue("muscle_groups"))
  if name == "" || description == "" {
    http.Redirect(w, r, "/admin/programs?error=Заполните%20название%20и%20описание", http.StatusSeeOther)
    return
  }
  _, err := s.DB.Exec(
    `insert into programs (name, description, muscle_groups)
     values ($1, $2, $3)
     on conflict (name)
     do update set description = excluded.description,
                   muscle_groups = excluded.muscle_groups`,
    name,
    description,
    muscles,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/programs?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/programs?success=Программа%20сохранена", http.StatusSeeOther)
}

func (s *Site) adminProgramUpdate(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  if programID == "" {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/programs?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  name := strings.TrimSpace(r.FormValue("name"))
  description := strings.TrimSpace(r.FormValue("description"))
  muscles := parseCSV(r.FormValue("muscle_groups"))
  _, err := s.DB.Exec(
    `update programs
     set name = coalesce(nullif($1, ''), name),
         description = coalesce(nullif($2, ''), description),
         muscle_groups = $3
     where id = $4`,
    name,
    description,
    muscles,
    programID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/programs?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/programs?success=Программа%20обновлена", http.StatusSeeOther)
}

func (s *Site) adminProgramDetail(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  if programID == "" {
    http.NotFound(w, r)
    return
  }
  var program programCard
  err := s.DB.QueryRow(
    `select id, name, description, coalesce(muscle_groups, '{}')
     from programs where id = $1`,
    programID,
  ).Scan(&program.ID, &program.Name, &program.Description, &program.MuscleGroups)
  if err != nil {
    http.NotFound(w, r)
    return
  }

  rows, err := s.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, ''), pw.sort_order
     from program_workouts pw
     join workouts w on w.id = pw.workout_id
     where pw.program_id = $1
     order by pw.sort_order`,
    programID,
  )
  type programWorkoutView struct {
    ID       string
    Name     string
    Description string
    Duration int
    Difficulty string
    Category string
    Order    int
  }
  list := []programWorkoutView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var v programWorkoutView
      _ = rows.Scan(&v.ID, &v.Name, &v.Description, &v.Duration, &v.Difficulty, &v.Category, &v.Order)
      list = append(list, v)
    }
  }

  wRows, err := s.DB.Query(`select id, name from workouts order by name`)
  workouts := []workoutCard{}
  if err == nil {
    defer wRows.Close()
    for wRows.Next() {
      var wCard workoutCard
      _ = wRows.Scan(&wCard.ID, &wCard.Name)
      workouts = append(workouts, wCard)
    }
  }

  data := s.baseData(r, program.Name, "admin")
  data["Program"] = program
  data["ProgramWorkouts"] = list
  data["Workouts"] = workouts
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "admin_program_detail", data)
}

func (s *Site) adminProgramWorkoutAdd(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  if programID == "" {
    http.Redirect(w, r, "/admin/programs", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/programs/"+programID+"?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  workoutID := r.FormValue("workout_id")
  order, _ := strconv.Atoi(r.FormValue("sort_order"))
  if workoutID == "" {
    http.Redirect(w, r, "/admin/programs/"+programID+"?error=Выберите%20тренировку", http.StatusSeeOther)
    return
  }
  if order <= 0 {
    order = 1
  }
  _, err := s.DB.Exec(
    `insert into program_workouts (program_id, workout_id, sort_order)
     values ($1, $2, $3)
     on conflict (program_id, workout_id)
     do update set sort_order = excluded.sort_order`,
    programID,
    workoutID,
    order,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/programs/"+programID+"?error=Не%20удалось%20добавить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/programs/"+programID+"?success=Тренировка%20добавлена", http.StatusSeeOther)
}

func (s *Site) adminProgramWorkoutRemove(w http.ResponseWriter, r *http.Request) {
  programID := chi.URLParam(r, "id")
  workoutID := chi.URLParam(r, "workoutId")
  if programID == "" || workoutID == "" {
    http.Redirect(w, r, "/admin/programs", http.StatusSeeOther)
    return
  }
  _, _ = s.DB.Exec(`delete from program_workouts where program_id = $1 and workout_id = $2`, programID, workoutID)
  http.Redirect(w, r, "/admin/programs/"+programID+"?success=Тренировка%20удалена", http.StatusSeeOther)
}

func (s *Site) adminPlans(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Планы сотрудников", "admin")
  rows, err := s.DB.Query(
    `select u.id, u.name, u.employee_id,
            coalesce(tp.id::text, ''), coalesce(tp.goal, ''), coalesce(tp.level, ''), coalesce(tp.status, '')
     from users u
     left join training_plans tp on tp.user_id = u.id and tp.status in ('active', 'paused')
     order by u.name`,
  )
  plans := []adminPlanView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var v adminPlanView
      _ = rows.Scan(&v.UserID, &v.Name, &v.EmployeeID, &v.PlanID, &v.Goal, &v.Level, &v.Status)
      plans = append(plans, v)
    }
  }
  data["Plans"] = plans
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "admin_plans", data)
}

func (s *Site) adminPlanDetail(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.NotFound(w, r)
    return
  }
  var userName string
  var employeeID string
  _ = s.DB.QueryRow(`select name, employee_id from users where id = $1`, userID).Scan(&userName, &employeeID)

  plan, _ := s.getActivePlan(userID)
  workouts := []planWorkoutView{}
  if plan != nil {
    workouts = s.fetchPlanWorkouts(plan.ID)
  }

  wRows, err := s.DB.Query(`select id, name, description, duration_minutes, difficulty, coalesce(category, '') from workouts order by name`)
  allWorkouts := []workoutCard{}
  if err == nil {
    defer wRows.Close()
    for wRows.Next() {
      var wCard workoutCard
      _ = wRows.Scan(&wCard.ID, &wCard.Name, &wCard.Description, &wCard.Duration, &wCard.Difficulty, &wCard.Category)
      allWorkouts = append(allWorkouts, wCard)
    }
  }

  data := s.baseData(r, "План сотрудника", "admin")
  data["UserID"] = userID
  data["EmployeeName"] = userName
  data["EmployeeID"] = employeeID
  data["Plan"] = plan
  data["PlanWorkouts"] = workouts
  data["Workouts"] = allWorkouts
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")
  s.render(w, "admin_plan_detail", data)
}

func (s *Site) adminPlanRegenerate(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.Redirect(w, r, "/admin/plans", http.StatusSeeOther)
    return
  }
  if plan, err := s.getActivePlan(userID); err == nil && plan != nil {
    _, _ = s.DB.Exec(`update training_plans set status = 'archived', updated_at = now() where id = $1`, plan.ID)
  }
  _, _ = s.ensurePlan(userID)
  http.Redirect(w, r, "/admin/plans/"+userID+"?success=План%20пересобран", http.StatusSeeOther)
}

func (s *Site) adminPlanPause(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.Redirect(w, r, "/admin/plans", http.StatusSeeOther)
    return
  }
  if plan, err := s.getActivePlan(userID); err == nil && plan != nil {
    _, _ = s.DB.Exec(`update training_plans set status = 'paused', paused_reason = 'Приостановлено администратором', updated_at = now() where id = $1`, plan.ID)
  }
  http.Redirect(w, r, "/admin/plans/"+userID+"?success=План%20приостановлен", http.StatusSeeOther)
}

func (s *Site) adminPlanResume(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    http.Redirect(w, r, "/admin/plans", http.StatusSeeOther)
    return
  }
  if plan, err := s.getActivePlan(userID); err == nil && plan != nil {
    _, _ = s.DB.Exec(`update training_plans set status = 'active', paused_reason = null, updated_at = now() where id = $1`, plan.ID)
  }
  http.Redirect(w, r, "/admin/plans/"+userID+"?success=План%20возобновлен", http.StatusSeeOther)
}

func (s *Site) adminPlanWorkoutReplace(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  planWorkoutID := chi.URLParam(r, "planWorkoutId")
  if userID == "" || planWorkoutID == "" {
    http.Redirect(w, r, "/admin/plans", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/plans/"+userID+"?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  workoutID := r.FormValue("workout_id")
  if workoutID == "" {
    http.Redirect(w, r, "/admin/plans/"+userID+"?error=Выберите%20тренировку", http.StatusSeeOther)
    return
  }
  _, err := s.DB.Exec(
    `update training_plan_workouts
     set workout_id = $1, status = 'pending', session_id = null
     where id = $2`,
    workoutID,
    planWorkoutID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/plans/"+userID+"?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/plans/"+userID+"?success=Тренировка%20заменена", http.StatusSeeOther)
}
