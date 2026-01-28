package api

import (
  "database/sql"
  "errors"
  "net/http"
  "strings"

  "golang.org/x/crypto/bcrypt"

  "rehab-app/internal/db"
)

type registerRequest struct {
  Name       string `json:"name"`
  EmployeeID string `json:"employee_id"`
  Department string `json:"department"`
  Position   string `json:"position"`
  Password   string `json:"password"`
}

type forgotPasswordRequest struct {
  EmployeeID string `json:"employee_id"`
  Message    string `json:"message"`
}

func (api *API) Register(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
    return
  }
  if !api.Config.AllowSelfRegister {
    writeJSON(w, http.StatusForbidden, map[string]any{"error": "registration disabled"})
    return
  }

  var req registerRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if req.Name == "" || req.EmployeeID == "" || req.Password == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  var userID string
  err = api.DB.QueryRow(
    `insert into users (name, employee_id, password_hash, role, department, position)
     values ($1, $2, $3, 'employee', $4, $5)
     returning id`,
    req.Name,
    req.EmployeeID,
    string(hash),
    req.Department,
    req.Position,
  ).Scan(&userID)
  if err != nil {
    if isUniqueViolation(err) {
      writeJSON(w, http.StatusConflict, map[string]any{"error": "employee_id already exists"})
      return
    }
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  _ = db.EnsureUserDefaults(api.DB, userID)

  token, err := api.createToken(userID)
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{
    "token": token,
    "user": map[string]any{
      "id": userID,
      "name": req.Name,
      "employee_id": req.EmployeeID,
      "role": "employee",
      "department": req.Department,
      "position": req.Position,
      "fitness_level": "",
      "onboarding_complete": false,
    },
  })
}

func (api *API) Me(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  if userID == "" {
    writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "missing user"})
    return
  }

  var name, employeeID, role, department, position string
  err := api.DB.QueryRow(
    `select name, employee_id, role, coalesce(department, ''), coalesce(position, '')
     from users where id = $1`,
    userID,
  ).Scan(&name, &employeeID, &role, &department, &position)
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "user not found"})
      return
    }
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  var age int
  var fitnessLevel string
  var restrictions []string
  var goals []string
  var onboardingComplete bool
  _ = api.DB.QueryRow(
    `select coalesce(age, 0), coalesce(fitness_level, ''), restrictions, goals, onboarding_complete
     from user_profiles where user_id = $1`,
    userID,
  ).Scan(&age, &fitnessLevel, &restrictions, &goals, &onboardingComplete)

  settings, _ := api.loadSettings(userID)

  writeJSON(w, http.StatusOK, map[string]any{
    "user": map[string]any{
      "id": userID,
      "name": name,
      "employee_id": employeeID,
      "role": role,
      "department": department,
      "position": position,
    },
    "profile": map[string]any{
      "age": age,
      "fitness_level": fitnessLevel,
      "restrictions": restrictions,
      "goals": goals,
      "onboarding_complete": onboardingComplete,
    },
    "settings": settings,
  })
}

func (api *API) Logout(w http.ResponseWriter, r *http.Request) {
  token := tokenFromRequest(r)
  if token != "" {
    _, _ = api.DB.Exec("delete from api_tokens where token = $1", token)
  }
  writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (api *API) ForgotPassword(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
    return
  }

  var req forgotPasswordRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }
  if req.EmployeeID == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing employee_id"})
    return
  }

  var userID string
  err := api.DB.QueryRow("select id from users where employee_id = $1", req.EmployeeID).Scan(&userID)
  if err == nil && userID != "" {
    message := req.Message
    if message == "" {
      message = "Запрос на сброс пароля"
    }
    _, _ = api.DB.Exec(
      `insert into support_tickets (user_id, category, subject, message)
       values ($1, $2, $3, $4)`,
      userID,
      "Аккаунт",
      "Сброс пароля",
      message,
    )
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "requested"})
}

func tokenFromRequest(r *http.Request) string {
  token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
  if token == "" {
    token = strings.TrimSpace(r.URL.Query().Get("token"))
  }
  return token
}
