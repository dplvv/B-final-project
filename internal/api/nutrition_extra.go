package api

import (
  "database/sql"
  "net/http"
  "time"

  "github.com/go-chi/chi/v5"
)

type waterUpdateRequest struct {
  Date  string `json:"date"`
  Delta int    `json:"delta"`
  Goal  int    `json:"goal"`
}

type mealRequest struct {
  Date     string `json:"date"`
  Title    string `json:"title"`
  Calories int    `json:"calories"`
  MealType string `json:"meal_type"`
}

func (api *API) NutritionWater(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  date := r.URL.Query().Get("date")
  if date == "" {
    date = time.Now().Format("2006-01-02")
  }

  var amount int
  var goal int
  err := api.DB.QueryRow(
    `select amount_ml, goal_ml
     from nutrition_water_logs
     where user_id = $1 and log_date = $2`,
    userID,
    date,
  ).Scan(&amount, &goal)
  if err != nil {
    if err == sql.ErrNoRows {
      amount = 0
      goal = 2500
    } else {
      writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
      return
    }
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "date":  date,
    "amount": amount,
    "goal":  goal,
  })
}

func (api *API) NutritionWaterUpdate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var req waterUpdateRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  date := req.Date
  if date == "" {
    date = time.Now().Format("2006-01-02")
  }
  goal := req.Goal
  if goal == 0 {
    goal = 2500
  }

  var amount int
  var updatedGoal int
  err := api.DB.QueryRow(
    `insert into nutrition_water_logs (user_id, log_date, amount_ml, goal_ml)
     values ($1, $2, greatest($3, 0), $4)
     on conflict (user_id, log_date)
     do update set amount_ml = greatest(0, nutrition_water_logs.amount_ml + $3),
                   goal_ml = case when $4 > 0 then $4 else nutrition_water_logs.goal_ml end,
                   updated_at = now()
     returning amount_ml, goal_ml`,
    userID,
    date,
    req.Delta,
    goal,
  ).Scan(&amount, &updatedGoal)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "date":  date,
    "amount": amount,
    "goal":  updatedGoal,
  })
}

func (api *API) NutritionDiary(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  date := r.URL.Query().Get("date")
  if date == "" {
    date = time.Now().Format("2006-01-02")
  }

  rows, err := api.DB.Query(
    `select id, title, calories, meal_type
     from nutrition_diary
     where user_id = $1 and log_date = $2
     order by created_at desc`,
    userID,
    date,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  meals := []map[string]any{}
  for rows.Next() {
    var id, title, mealType string
    var calories int
    _ = rows.Scan(&id, &title, &calories, &mealType)
    meals = append(meals, map[string]any{
      "id": id,
      "title": title,
      "calories": calories,
      "meal_type": mealType,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{
    "date": date,
    "meals": meals,
  })
}

func (api *API) NutritionDiaryCreate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var req mealRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.Title == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing title"})
    return
  }

  date := req.Date
  if date == "" {
    date = time.Now().Format("2006-01-02")
  }
  mealType := req.MealType
  if mealType == "" {
    mealType = "other"
  }

  var id string
  err := api.DB.QueryRow(
    `insert into nutrition_diary (user_id, log_date, title, calories, meal_type)
     values ($1, $2, $3, $4, $5)
     returning id`,
    userID,
    date,
    req.Title,
    req.Calories,
    mealType,
  ).Scan(&id)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (api *API) NutritionDiaryDelete(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  id := chi.URLParam(r, "id")
  if id == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing id"})
    return
  }

  _, err := api.DB.Exec(
    `delete from nutrition_diary where id = $1 and user_id = $2`,
    id,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}
