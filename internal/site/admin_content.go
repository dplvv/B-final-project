package site

import (
  "bytes"
  "database/sql"
  "errors"
  "fmt"
  "io"
  "net/http"
  "net/url"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "time"

	"github.com/go-chi/chi/v5"
)

func (s *Site) adminExercises(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Cache-Control", "no-store")
  data := s.baseData(r, "Упражнения", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  rows, err := s.DB.Query(
    `select id, name, description, coalesce(category, ''), coalesce(difficulty, ''),
            coalesce(sets, 0), coalesce(reps, ''), coalesce(rest_seconds, 0),
            coalesce(duration_seconds, 0),
            coalesce(array_to_string(muscle_groups, ','), ''),
            coalesce(array_to_string(equipment, ','), ''),
            coalesce(image_url, '')
     from exercises
     order by name`,
  )
  exercises := []exerciseCard{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var ex exerciseCard
      var musclesRaw string
      var equipmentRaw string
      if err := rows.Scan(&ex.ID, &ex.Name, &ex.Description, &ex.Category, &ex.Difficulty, &ex.Sets, &ex.Reps, &ex.Rest, &ex.Duration, &musclesRaw, &equipmentRaw, &ex.ImageURL); err != nil {
        continue
      }
      ex.MuscleGroups = parseCSV(musclesRaw)
      ex.Equipment = parseCSV(equipmentRaw)
      ex.ImageURL = normalizeImageURL(ex.ImageURL)
      exercises = append(exercises, ex)
    }
  }
  data["Exercises"] = exercises
  s.render(w, "admin_exercises", data)
}

func (s *Site) adminRewards(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Поощрения", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  rows, err := s.DB.Query(
    `select id, title, description, points_cost, coalesce(category, ''), active
     from rewards
     order by points_cost`,
  )
  rewards := []rewardView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var rwd rewardView
      _ = rows.Scan(&rwd.ID, &rwd.Title, &rwd.Description, &rwd.PointsCost, &rwd.Category, &rwd.Active)
      rewards = append(rewards, rwd)
    }
  }
  data["Rewards"] = rewards
  s.render(w, "admin_rewards", data)
}

func (s *Site) adminRewardCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/rewards?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  title := strings.TrimSpace(r.FormValue("title"))
  description := strings.TrimSpace(r.FormValue("description"))
  category := strings.TrimSpace(r.FormValue("category"))
  cost, _ := strconv.Atoi(r.FormValue("points_cost"))
  active := r.FormValue("active") == "on"
  if title == "" || description == "" {
    http.Redirect(w, r, "/admin/rewards?error=Заполните%20название%20и%20описание", http.StatusSeeOther)
    return
  }

  _, err := s.DB.Exec(
    `insert into rewards (title, description, points_cost, category, active)
     values ($1, $2, $3, nullif($4, ''), $5)
     on conflict (title)
     do update set description = excluded.description,
                   points_cost = excluded.points_cost,
                   category = excluded.category,
                   active = excluded.active`,
    title,
    description,
    cost,
    category,
    active,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/rewards?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/rewards?success=Поощрение%20сохранено", http.StatusSeeOther)
}

func (s *Site) adminRewardUpdate(w http.ResponseWriter, r *http.Request) {
  rewardID := chi.URLParam(r, "id")
  if rewardID == "" {
    http.Redirect(w, r, "/admin/rewards?error=Не%20найдено%20поощрение", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/rewards?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  title := strings.TrimSpace(r.FormValue("title"))
  description := strings.TrimSpace(r.FormValue("description"))
  category := strings.TrimSpace(r.FormValue("category"))
  cost, _ := strconv.Atoi(r.FormValue("points_cost"))
  active := r.FormValue("active") == "on"

  _, err := s.DB.Exec(
    `update rewards
     set title = coalesce(nullif($1, ''), title),
         description = coalesce(nullif($2, ''), description),
         points_cost = $3,
         category = nullif($4, ''),
         active = $5
     where id = $6`,
    title,
    description,
    cost,
    category,
    active,
    rewardID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/rewards?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/rewards?success=Поощрение%20обновлено", http.StatusSeeOther)
}

func (s *Site) adminRewardDelete(w http.ResponseWriter, r *http.Request) {
  rewardID := chi.URLParam(r, "id")
  if rewardID == "" {
    http.Redirect(w, r, "/admin/rewards?error=Не%20найдено%20поощрение", http.StatusSeeOther)
    return
  }
  _, err := s.DB.Exec(`delete from rewards where id = $1`, rewardID)
  if err != nil {
    http.Redirect(w, r, "/admin/rewards?error=Не%20удалось%20удалить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/rewards?success=Поощрение%20удалено", http.StatusSeeOther)
}

func (s *Site) adminAchievements(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Достижения", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  type achievementAdminView struct {
    ID           string
    Title        string
    Description  string
    Icon         string
    PointsReward int
    Metric       string
    Target       int
  }

  rows, err := s.DB.Query(
    `select id, title, description, icon, points_reward, coalesce(metric, ''), coalesce(target, 0)
     from achievements
     order by created_at desc`,
  )
  achievements := []achievementAdminView{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var a achievementAdminView
      _ = rows.Scan(&a.ID, &a.Title, &a.Description, &a.Icon, &a.PointsReward, &a.Metric, &a.Target)
      achievements = append(achievements, a)
    }
  }
  data["Achievements"] = achievements
  s.render(w, "admin_achievements", data)
}

func (s *Site) adminAchievementCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/achievements?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  title := strings.TrimSpace(r.FormValue("title"))
  description := strings.TrimSpace(r.FormValue("description"))
  icon := strings.TrimSpace(r.FormValue("icon"))
  metric := strings.TrimSpace(r.FormValue("metric"))
  target, _ := strconv.Atoi(r.FormValue("target"))
  points, _ := strconv.Atoi(r.FormValue("points_reward"))
  if title == "" || description == "" || icon == "" || target <= 0 {
    http.Redirect(w, r, "/admin/achievements?error=Заполните%20все%20поля", http.StatusSeeOther)
    return
  }

  _, err := s.DB.Exec(
    `insert into achievements (title, description, icon, points_reward, metric, target)
     values ($1, $2, $3, $4, $5, $6)
     on conflict (title)
     do update set description = excluded.description,
                   icon = excluded.icon,
                   points_reward = excluded.points_reward,
                   metric = excluded.metric,
                   target = excluded.target`,
    title,
    description,
    icon,
    points,
    metric,
    target,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/achievements?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/achievements?success=Достижение%20сохранено", http.StatusSeeOther)
}

func (s *Site) adminAchievementUpdate(w http.ResponseWriter, r *http.Request) {
  achievementID := chi.URLParam(r, "id")
  if achievementID == "" {
    http.Redirect(w, r, "/admin/achievements?error=Не%20найдено%20достижение", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/achievements?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }
  title := strings.TrimSpace(r.FormValue("title"))
  description := strings.TrimSpace(r.FormValue("description"))
  icon := strings.TrimSpace(r.FormValue("icon"))
  metric := strings.TrimSpace(r.FormValue("metric"))
  target, _ := strconv.Atoi(r.FormValue("target"))
  points, _ := strconv.Atoi(r.FormValue("points_reward"))

  _, err := s.DB.Exec(
    `update achievements
     set title = coalesce(nullif($1, ''), title),
         description = coalesce(nullif($2, ''), description),
         icon = coalesce(nullif($3, ''), icon),
         points_reward = $4,
         metric = nullif($5, ''),
         target = $6
     where id = $7`,
    title,
    description,
    icon,
    points,
    metric,
    target,
    achievementID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/achievements?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/achievements?success=Достижение%20обновлено", http.StatusSeeOther)
}

func (s *Site) adminAchievementDelete(w http.ResponseWriter, r *http.Request) {
  achievementID := chi.URLParam(r, "id")
  if achievementID == "" {
    http.Redirect(w, r, "/admin/achievements?error=Не%20найдено%20достижение", http.StatusSeeOther)
    return
  }
  _, err := s.DB.Exec(`delete from achievements where id = $1`, achievementID)
  if err != nil {
    http.Redirect(w, r, "/admin/achievements?error=Не%20удалось%20удалить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/achievements?success=Достижение%20удалено", http.StatusSeeOther)
}

func (s *Site) adminExerciseCreate(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseMultipartForm(8 << 20); err != nil {
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
  imageURL, err := s.saveExerciseImage(r)
  if err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Некорректный%20файл%20изображения", http.StatusSeeOther)
    return
  }
  if name == "" || description == "" {
    http.Redirect(w, r, "/admin/exercises?error=Заполните%20название%20и%20описание", http.StatusSeeOther)
    return
  }

  _, err = s.DB.Exec(
    `insert into exercises (name, description, category, difficulty, sets, reps, rest_seconds, muscle_groups, equipment, image_url)
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
                   image_url = excluded.image_url`,
    name,
    description,
    category,
    difficulty,
    sets,
    reps,
    rest,
    muscles,
    equipment,
    imageURL,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/admin/exercises?success=Упражнение%20сохранено&ts="+strconv.FormatInt(time.Now().Unix(), 10), http.StatusSeeOther)
}

func (s *Site) adminExerciseUpdate(w http.ResponseWriter, r *http.Request) {
  exerciseID := chi.URLParam(r, "id")
  if exerciseID == "" {
    http.Redirect(w, r, "/admin/exercises?error=Не%20найдено%20упражнение", http.StatusSeeOther)
    return
  }
  if err := r.ParseMultipartForm(8 << 20); err != nil {
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
  imageURL, err := s.saveExerciseImage(r)
  if err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Некорректный%20файл%20изображения", http.StatusSeeOther)
    return
  }
  if imageURL == "" {
    _ = s.DB.QueryRow(`select coalesce(image_url, '') from exercises where id = $1`, exerciseID).Scan(&imageURL)
    imageURL = normalizeImageURL(imageURL)
  }

  _, err = s.DB.Exec(
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
         image_url = nullif($10, ''),
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
    imageURL,
    exerciseID,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/exercises?success=Упражнение%20обновлено&ts="+strconv.FormatInt(time.Now().Unix(), 10), http.StatusSeeOther)
}

func (s *Site) adminExerciseImageDelete(w http.ResponseWriter, r *http.Request) {
  exerciseID := chi.URLParam(r, "id")
  if exerciseID == "" {
    http.Redirect(w, r, "/admin/exercises?error=Не%20найдено%20упражнение", http.StatusSeeOther)
    return
  }

  var imageURL string
  if err := s.DB.QueryRow(`select coalesce(image_url, '') from exercises where id = $1`, exerciseID).Scan(&imageURL); err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20найти%20фото", http.StatusSeeOther)
    return
  }
  imageURL = normalizeImageURL(imageURL)

  if imageURL != "" && strings.HasPrefix(imageURL, "/uploads/") {
    filename := filepath.Base(imageURL)
    if filename != "" && filename != "." && filename != "/" {
      _ = os.Remove(filepath.Join("uploads", filename))
    }
  }

  if _, err := s.DB.Exec(`update exercises set image_url = null where id = $1`, exerciseID); err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20удалить%20фото", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/admin/exercises?success=Фото%20удалено", http.StatusSeeOther)
}

func (s *Site) adminExerciseDelete(w http.ResponseWriter, r *http.Request) {
  exerciseID := chi.URLParam(r, "id")
  if exerciseID == "" {
    http.Redirect(w, r, "/admin/exercises?error=Не%20найдено%20упражнение", http.StatusSeeOther)
    return
  }

  if _, err := s.DB.Exec(`delete from exercises where id = $1`, exerciseID); err != nil {
    http.Redirect(w, r, "/admin/exercises?error=Не%20удалось%20удалить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/exercises?success=Упражнение%20удалено", http.StatusSeeOther)
}

func (s *Site) saveExerciseImage(r *http.Request) (string, error) {
  file, header, err := r.FormFile("image_file")
  if err != nil {
    if errors.Is(err, http.ErrMissingFile) {
      return "", nil
    }
    return "", err
  }
  defer file.Close()

  headerBuf := make([]byte, 512)
  n, _ := file.Read(headerBuf)
  contentType := http.DetectContentType(headerBuf[:n])
  ext := imageExtension(contentType)
  if ext == "" && header != nil {
    ext = strings.ToLower(filepath.Ext(header.Filename))
    if !isAllowedImageExt(ext) {
      ext = ""
    }
  }
  if ext == "" {
    return "", fmt.Errorf("unsupported image type")
  }

  if err := os.MkdirAll("uploads", 0o755); err != nil {
    return "", err
  }

  token, err := randomToken(8)
  if err != nil {
    return "", err
  }
  filename := "exercise_" + token + ext
  dstPath := filepath.Join("uploads", filename)

  reader := io.MultiReader(bytes.NewReader(headerBuf[:n]), file)
  out, err := os.Create(dstPath)
  if err != nil {
    return "", err
  }
  defer out.Close()

  if _, err := io.Copy(out, reader); err != nil {
    return "", err
  }

  return "/uploads/" + filename, nil
}

func imageExtension(contentType string) string {
  switch strings.ToLower(contentType) {
  case "image/jpeg":
    return ".jpg"
  case "image/png":
    return ".png"
  case "image/webp":
    return ".webp"
  case "image/gif":
    return ".gif"
  default:
    return ""
  }
}

func isAllowedImageExt(ext string) bool {
  switch strings.ToLower(ext) {
  case ".jpg", ".jpeg", ".png", ".webp", ".gif":
    return true
  default:
    return false
  }
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

func (s *Site) adminWorkoutDelete(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  if workoutID == "" {
    http.Redirect(w, r, "/admin/workouts?error=Не%20найдена%20тренировка", http.StatusSeeOther)
    return
  }

  if _, err := s.DB.Exec(`delete from workouts where id = $1`, workoutID); err != nil {
    http.Redirect(w, r, "/admin/workouts?error=Не%20удалось%20удалить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/workouts?success=Тренировка%20удалена", http.StatusSeeOther)
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
    _ = s.DB.QueryRow(
      `select coalesce(max(sort_order), 0) + 1 from workout_exercises where workout_id = $1`,
      workoutID,
    ).Scan(&order)
    if order <= 0 {
      order = 1
    }
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
  s.normalizeWorkoutOrder(workoutID)
  http.Redirect(w, r, "/admin/workouts/"+workoutID+"?success=Упражнение%20добавлено", http.StatusSeeOther)
}

func (s *Site) adminWorkoutExerciseUpdate(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  exerciseID := chi.URLParam(r, "exerciseId")
  if workoutID == "" || exerciseID == "" {
    http.Redirect(w, r, "/admin/workouts", http.StatusSeeOther)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Redirect(w, r, "/admin/workouts/"+workoutID+"?error=Некорректные%20данные", http.StatusSeeOther)
    return
  }

  newExerciseID := strings.TrimSpace(r.FormValue("exercise_id"))
  if newExerciseID == "" {
    newExerciseID = exerciseID
  }
  order, _ := strconv.Atoi(r.FormValue("sort_order"))
  if order <= 0 {
    _ = s.DB.QueryRow(
      `select coalesce(sort_order, 1) from workout_exercises where workout_id = $1 and exercise_id = $2`,
      workoutID,
      exerciseID,
    ).Scan(&order)
    if order <= 0 {
      order = 1
    }
  }
  sets, _ := strconv.Atoi(r.FormValue("sets"))
  reps := strings.TrimSpace(r.FormValue("reps"))
  rest, _ := strconv.Atoi(r.FormValue("rest_seconds"))

  if newExerciseID != exerciseID {
    _, _ = s.DB.Exec(`delete from workout_exercises where workout_id = $1 and exercise_id = $2`, workoutID, exerciseID)
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
    newExerciseID,
    order,
    sets,
    reps,
    rest,
  )
  if err != nil {
    http.Redirect(w, r, "/admin/workouts/"+workoutID+"?error=Не%20удалось%20обновить", http.StatusSeeOther)
    return
  }
  s.normalizeWorkoutOrder(workoutID)
  http.Redirect(w, r, "/admin/workouts/"+workoutID+"?success=Упражнение%20обновлено", http.StatusSeeOther)
}

func (s *Site) adminWorkoutExerciseRemove(w http.ResponseWriter, r *http.Request) {
  workoutID := chi.URLParam(r, "id")
  exerciseID := chi.URLParam(r, "exerciseId")
  if workoutID == "" || exerciseID == "" {
    http.Redirect(w, r, "/admin/workouts", http.StatusSeeOther)
    return
  }
  _, _ = s.DB.Exec(`delete from workout_exercises where workout_id = $1 and exercise_id = $2`, workoutID, exerciseID)
  s.normalizeWorkoutOrder(workoutID)
  http.Redirect(w, r, "/admin/workouts/"+workoutID+"?success=Упражнение%20удалено", http.StatusSeeOther)
}

func (s *Site) normalizeWorkoutOrder(workoutID string) {
  rows, err := s.DB.Query(
    `select exercise_id
     from workout_exercises
     where workout_id = $1
     order by sort_order, exercise_id`,
    workoutID,
  )
  if err != nil {
    return
  }
  defer rows.Close()

  order := 1
  for rows.Next() {
    var exerciseID string
    _ = rows.Scan(&exerciseID)
    _, _ = s.DB.Exec(
      `update workout_exercises
       set sort_order = $1
       where workout_id = $2 and exercise_id = $3`,
      order,
      workoutID,
      exerciseID,
    )
    order++
  }
}

func (s *Site) adminPrograms(w http.ResponseWriter, r *http.Request) {
  data := s.baseData(r, "Программы", "admin")
  data["Success"] = r.URL.Query().Get("success")
  data["Error"] = r.URL.Query().Get("error")

  rows, err := s.DB.Query(
    `select id, name, description,
            coalesce(replace(replace(replace(muscle_groups::text, '{', ''), '}', ''), '"', ''), '')
     from programs
     order by created_at desc`,
  )
  programs := []programCard{}
  if err == nil {
    defer rows.Close()
    for rows.Next() {
      var p programCard
      var musclesRaw string
      if err := rows.Scan(&p.ID, &p.Name, &p.Description, &musclesRaw); err != nil {
        continue
      }
      p.MuscleGroups = parseCSV(musclesRaw)
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
  var existingProgramID string
  err := s.DB.QueryRow(`select id from programs where lower(name) = lower($1) limit 1`, name).Scan(&existingProgramID)
  if err == nil {
    _, err = s.DB.Exec(
      `update programs
       set description = $1,
           muscle_groups = $2,
           active = true
       where id = $3`,
      description,
      muscles,
      existingProgramID,
    )
  } else if errors.Is(err, sql.ErrNoRows) {
    _, err = s.DB.Exec(
      `insert into programs (name, description, muscle_groups, active)
       values ($1, $2, $3, true)`,
      name,
      description,
      muscles,
    )
  }
  if err != nil {
    http.Redirect(w, r, "/admin/programs?error=Не%20удалось%20сохранить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/programs?success=Программа%20сохранена", http.StatusSeeOther)
}

func (s *Site) resolveProgramID(rawParam string) string {
  rawParam = strings.TrimSpace(rawParam)
  if rawParam == "" {
    return ""
  }

  var resolvedID string
  err := s.DB.QueryRow(`select id from programs where id = $1`, rawParam).Scan(&resolvedID)
  if err == nil {
    return resolvedID
  }

  programID := normalizeResourceID(rawParam)
  if programID != rawParam {
    err = s.DB.QueryRow(`select id from programs where id = $1`, programID).Scan(&resolvedID)
    if err == nil {
      return resolvedID
    }
  }

  nameParam := rawParam
  if decoded, decodeErr := url.PathUnescape(nameParam); decodeErr == nil {
    nameParam = decoded
  }
  nameParam = strings.TrimSpace(nameParam)
  if nameParam == "" {
    return ""
  }

  _ = s.DB.QueryRow(`select id from programs where lower(name) = lower($1)`, nameParam).Scan(&resolvedID)
  return resolvedID
}

func (s *Site) resolveProgramIDFromRequest(r *http.Request) string {
  if r == nil {
    return ""
  }

  if resolved := s.resolveProgramID(chi.URLParam(r, "id")); resolved != "" {
    return resolved
  }
  if resolved := s.resolveProgramID(strings.TrimSpace(r.URL.Query().Get("program_id"))); resolved != "" {
    return resolved
  }

  _ = r.ParseForm()
  if resolved := s.resolveProgramID(strings.TrimSpace(r.FormValue("program_id"))); resolved != "" {
    return resolved
  }

  ref := strings.TrimSpace(r.Referer())
  if ref == "" {
    return ""
  }
  parsed, err := url.Parse(ref)
  if err != nil {
    return ""
  }
  parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
  for i := 0; i+1 < len(parts); i++ {
    if parts[i] == "programs" {
      if resolved := s.resolveProgramID(parts[i+1]); resolved != "" {
        return resolved
      }
      break
    }
  }
  return ""
}

func (s *Site) adminProgramUpdate(w http.ResponseWriter, r *http.Request) {
  programID := s.resolveProgramIDFromRequest(r)
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
         muscle_groups = $3,
         active = true
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

func (s *Site) adminProgramDelete(w http.ResponseWriter, r *http.Request) {
  programID := s.resolveProgramIDFromRequest(r)
  if programID == "" {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
    return
  }

  if _, err := s.DB.Exec(`delete from programs where id = $1`, programID); err != nil {
    http.Redirect(w, r, "/admin/programs?error=Не%20удалось%20удалить", http.StatusSeeOther)
    return
  }
  http.Redirect(w, r, "/admin/programs?success=Программа%20удалена", http.StatusSeeOther)
}

func (s *Site) adminProgramDetail(w http.ResponseWriter, r *http.Request) {
  rawParam := strings.TrimSpace(chi.URLParam(r, "id"))
  if rawParam == "" {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
    return
  }
  programID := s.resolveProgramID(rawParam)
  if programID == "" {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
    return
  }
  var program programCard
  var musclesRaw string
  err := s.DB.QueryRow(
    `select id, name, description,
            coalesce(replace(replace(replace(muscle_groups::text, '{', ''), '}', ''), '"', ''), '')
     from programs where id = $1`,
    programID,
  ).Scan(&program.ID, &program.Name, &program.Description, &musclesRaw)
  program.MuscleGroups = parseCSV(musclesRaw)
  if err != nil {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
    return
  }

  rows, err := s.DB.Query(
    `select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, ''), pw.sort_order
     from program_workouts pw
     join workouts w on w.id = pw.workout_id
     where pw.program_id = $1
     order by pw.sort_order`,
    program.ID,
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
  programID := s.resolveProgramIDFromRequest(r)
  if programID == "" {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
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
  programID := s.resolveProgramIDFromRequest(r)
  workoutID := chi.URLParam(r, "workoutId")
  if programID == "" || workoutID == "" {
    http.Redirect(w, r, "/admin/programs?error=Не%20найдена%20программа", http.StatusSeeOther)
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

func adminPlanPath(userID string) string {
  base := "/admin/plans"
  if strings.TrimSpace(userID) == "" {
    return base
  }
  return base + "/" + userID
}

func adminPlanRedirect(w http.ResponseWriter, r *http.Request, userID, kind, message string) {
  destination := adminPlanPath(userID)
  if strings.TrimSpace(kind) != "" && strings.TrimSpace(message) != "" {
    query := url.Values{}
    query.Set(kind, message)
    destination += "?" + query.Encode()
  }
  http.Redirect(w, r, destination, http.StatusSeeOther)
}

func (s *Site) adminPlanRegenerate(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    adminPlanRedirect(w, r, "", "", "")
    return
  }
  if plan, err := s.getActivePlan(userID); err == nil && plan != nil {
    _, _ = s.DB.Exec(`update training_plans set status = 'archived', updated_at = now() where id = $1`, plan.ID)
  }
  _, _ = s.ensurePlan(userID)
  adminPlanRedirect(w, r, userID, "success", "План пересобран")
}

func (s *Site) adminPlanPause(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    adminPlanRedirect(w, r, "", "", "")
    return
  }
  if plan, err := s.getActivePlan(userID); err == nil && plan != nil {
    _, _ = s.DB.Exec(`update training_plans set status = 'paused', paused_reason = 'Приостановлено администратором', updated_at = now() where id = $1`, plan.ID)
  }
  adminPlanRedirect(w, r, userID, "success", "План приостановлен")
}

func (s *Site) adminPlanPauseSick(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    adminPlanRedirect(w, r, "", "", "")
    return
  }
  if err := r.ParseForm(); err != nil {
    adminPlanRedirect(w, r, userID, "error", "Некорректные данные")
    return
  }

  fromRaw := strings.TrimSpace(r.FormValue("sick_from"))
  toRaw := strings.TrimSpace(r.FormValue("sick_to"))
  comment := strings.TrimSpace(r.FormValue("comment"))
  if fromRaw == "" || toRaw == "" {
    adminPlanRedirect(w, r, userID, "error", "Укажите период больничного")
    return
  }

  fromDate, err := time.Parse("2006-01-02", fromRaw)
  if err != nil {
    adminPlanRedirect(w, r, userID, "error", "Некорректная дата начала")
    return
  }
  toDate, err := time.Parse("2006-01-02", toRaw)
  if err != nil {
    adminPlanRedirect(w, r, userID, "error", "Некорректная дата окончания")
    return
  }
  if toDate.Before(fromDate) {
    adminPlanRedirect(w, r, userID, "error", "Дата окончания раньше даты начала")
    return
  }

  plan, err := s.getActivePlan(userID)
  if err != nil || plan == nil {
    adminPlanRedirect(w, r, userID, "error", "План не найден")
    return
  }

  daysCount := int(toDate.Sub(fromDate).Hours()/24) + 1
  if daysCount < 1 {
    daysCount = 1
  }

  before := s.planSnapshot(plan.ID)
  result, _ := s.DB.Exec(
    `update training_plan_workouts
     set scheduled_date = scheduled_date + ($1::int * interval '1 day')
     where plan_id = $2
       and status = 'pending'
       and scheduled_date is not null
       and scheduled_date >= $3`,
    daysCount,
    plan.ID,
    fromDate,
  )

  pauseReason := fmt.Sprintf(
    "Приостановлено администратором по больничному листу: %s — %s",
    fromDate.Format("02.01.2006"),
    toDate.Format("02.01.2006"),
  )
  if comment != "" {
    pauseReason += ". " + comment
  }
  _, _ = s.DB.Exec(
    `update training_plans
     set status = 'paused',
         paused_reason = $1,
         updated_at = now()
     where id = $2`,
    pauseReason,
    plan.ID,
  )

  shifted := int(affectedRows(result))
  after := s.planSnapshot(plan.ID)
  changeReason := fmt.Sprintf(
    "Больничный %s — %s: сдвиг тренировок на %d дн. (перенесено: %d)",
    fromDate.Format("02.01.2006"),
    toDate.Format("02.01.2006"),
    daysCount,
    shifted,
  )
  s.logPlanChange(userID, plan.ID, "sick_leave", changeReason, before, after)

  adminPlanRedirect(w, r, userID, "success", "План приостановлен по больничному")
}

func (s *Site) adminPlanResume(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    adminPlanRedirect(w, r, "", "", "")
    return
  }
  if plan, err := s.getActivePlan(userID); err == nil && plan != nil {
    _, _ = s.DB.Exec(`update training_plans set status = 'active', paused_reason = null, updated_at = now() where id = $1`, plan.ID)
  }
  adminPlanRedirect(w, r, userID, "success", "План возобновлен")
}

func (s *Site) adminPlanDelete(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  if userID == "" {
    adminPlanRedirect(w, r, "", "error", "Не найден сотрудник")
    return
  }

  if _, err := s.DB.Exec(`delete from training_plans where user_id = $1`, userID); err != nil {
    adminPlanRedirect(w, r, "", "error", "Не удалось удалить план")
    return
  }
  adminPlanRedirect(w, r, "", "success", "Планы сотрудника удалены")
}

func (s *Site) adminPlanWorkoutReplace(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  planWorkoutID := chi.URLParam(r, "planWorkoutId")
  if userID == "" || planWorkoutID == "" {
    adminPlanRedirect(w, r, "", "", "")
    return
  }
  if err := r.ParseForm(); err != nil {
    adminPlanRedirect(w, r, userID, "error", "Некорректные данные")
    return
  }
  workoutID := r.FormValue("workout_id")
  if workoutID == "" {
    adminPlanRedirect(w, r, userID, "error", "Выберите тренировку")
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
    adminPlanRedirect(w, r, userID, "error", "Не удалось обновить")
    return
  }
  adminPlanRedirect(w, r, userID, "success", "Тренировка заменена")
}

func (s *Site) adminPlanWorkoutReschedule(w http.ResponseWriter, r *http.Request) {
  userID := chi.URLParam(r, "id")
  planWorkoutID := chi.URLParam(r, "planWorkoutId")
  if userID == "" || planWorkoutID == "" {
    adminPlanRedirect(w, r, "", "", "")
    return
  }
  if err := r.ParseForm(); err != nil {
    adminPlanRedirect(w, r, userID, "error", "Некорректные данные")
    return
  }

  scheduledDateRaw := strings.TrimSpace(r.FormValue("scheduled_date"))
  if scheduledDateRaw == "" {
    adminPlanRedirect(w, r, userID, "error", "Укажите дату тренировки")
    return
  }
  scheduledDate, err := time.Parse("2006-01-02", scheduledDateRaw)
  if err != nil {
    adminPlanRedirect(w, r, userID, "error", "Некорректная дата тренировки")
    return
  }

  var planID string
  var status string
  err = s.DB.QueryRow(
    `select pw.plan_id, pw.status
     from training_plan_workouts pw
     join training_plans tp on tp.id = pw.plan_id
     where pw.id = $1 and tp.user_id = $2`,
    planWorkoutID,
    userID,
  ).Scan(&planID, &status)
  if err != nil {
    adminPlanRedirect(w, r, userID, "error", "Тренировка плана не найдена")
    return
  }
  if status != "pending" {
    adminPlanRedirect(w, r, userID, "error", "Можно изменять дату только запланированной тренировки")
    return
  }

  before := s.planSnapshot(planID)
  result, err := s.DB.Exec(
    `update training_plan_workouts
     set scheduled_date = $1
     where id = $2 and status = 'pending'`,
    scheduledDate,
    planWorkoutID,
  )
  if err != nil || affectedRows(result) == 0 {
    adminPlanRedirect(w, r, userID, "error", "Не удалось изменить дату тренировки")
    return
  }

  after := s.planSnapshot(planID)
  s.logPlanChange(
    userID,
    planID,
    "reschedule",
    "Администратор изменил дату тренировки на "+scheduledDate.Format("02.01.2006"),
    before,
    after,
  )
  adminPlanRedirect(w, r, userID, "success", "Дата тренировки обновлена")
}
