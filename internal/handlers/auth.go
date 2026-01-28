package handlers

import (
  "crypto/rand"
  "database/sql"
  "encoding/base64"
  "net/http"
  "time"

  "golang.org/x/crypto/bcrypt"

  "rehab-app/internal/db"
)

func (a *App) LoginPage(w http.ResponseWriter, r *http.Request) {
  data := a.baseData(w, r)
  data["Title"] = "Вход"
  data["UseShell"] = false
  data["ShowNav"] = false
  _ = a.Renderer.Render(w, "login", data)
}

func (a *App) LoginPost(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  employeeID := r.FormValue("employee_id")
  password := r.FormValue("password")
  name := r.FormValue("name")

  if employeeID == "" || password == "" {
    a.setFlash(w, "Заполните все поля")
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var userID string
  var passwordHash string
  err := a.DB.QueryRow(
    "select id, password_hash from users where employee_id = $1",
    employeeID,
  ).Scan(&userID, &passwordHash)
  if err != nil {
    if err == sql.ErrNoRows {
      a.setFlash(w, "Пользователь не найден")
      http.Redirect(w, r, "/login", http.StatusSeeOther)
      return
    }
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
    a.setFlash(w, "Неверный пароль")
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if name != "" {
    _, _ = a.DB.Exec("update users set name = $1 where id = $2", name, userID)
  }

  if err := a.createSession(w, userID); err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  var onboardingComplete bool
  _ = a.DB.QueryRow("select onboarding_complete from user_profiles where user_id = $1", userID).Scan(&onboardingComplete)
  if !onboardingComplete {
    http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) RegisterPage(w http.ResponseWriter, r *http.Request) {
  if !a.Config.AllowSelfRegister {
    http.NotFound(w, r)
    return
  }
  data := a.baseData(w, r)
  data["Title"] = "Регистрация"
  data["UseShell"] = false
  data["ShowNav"] = false
  _ = a.Renderer.Render(w, "register", data)
}

func (a *App) RegisterPost(w http.ResponseWriter, r *http.Request) {
  if !a.Config.AllowSelfRegister {
    http.NotFound(w, r)
    return
  }
  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  name := r.FormValue("name")
  employeeID := r.FormValue("employee_id")
  department := r.FormValue("department")
  password := r.FormValue("password")

  if name == "" || employeeID == "" || password == "" {
    a.setFlash(w, "Заполните все поля")
    http.Redirect(w, r, "/register", http.StatusSeeOther)
    return
  }

  hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
  if err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  var userID string
  err = a.DB.QueryRow(
    `insert into users (name, employee_id, password_hash, role, department)
     values ($1, $2, $3, 'employee', $4)
     returning id`,
    name,
    employeeID,
    string(hash),
    department,
  ).Scan(&userID)
  if err != nil {
    a.setFlash(w, "Не удалось создать пользователя")
    http.Redirect(w, r, "/register", http.StatusSeeOther)
    return
  }

  _ = db.EnsureUserDefaults(a.DB, userID)

  if err := a.createSession(w, userID); err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  http.Redirect(w, r, "/onboarding", http.StatusSeeOther)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
  cookie, err := r.Cookie(a.Config.CookieName)
  if err == nil && cookie.Value != "" {
    _, _ = a.DB.Exec("delete from sessions where token = $1", cookie.Value)
  }

  http.SetCookie(w, &http.Cookie{
    Name:     a.Config.CookieName,
    Value:    "",
    Path:     "/",
    MaxAge:   -1,
    HttpOnly: true,
    Secure:   a.Config.CookieSecure,
    SameSite: http.SameSiteLaxMode,
  })

  http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) createSession(w http.ResponseWriter, userID string) error {
  token, err := randomToken(32)
  if err != nil {
    return err
  }

  expiresAt := time.Now().Add(a.Config.SessionTTL)
  _, err = a.DB.Exec(
    `insert into sessions (user_id, token, expires_at)
     values ($1, $2, $3)`,
    userID,
    token,
    expiresAt,
  )
  if err != nil {
    return err
  }

  http.SetCookie(w, &http.Cookie{
    Name:     a.Config.CookieName,
    Value:    token,
    Path:     "/",
    Expires:  expiresAt,
    HttpOnly: true,
    Secure:   a.Config.CookieSecure,
    SameSite: http.SameSiteLaxMode,
  })

  return nil
}

func randomToken(size int) (string, error) {
  buffer := make([]byte, size)
  if _, err := rand.Read(buffer); err != nil {
    return "", err
  }
  return base64.RawURLEncoding.EncodeToString(buffer), nil
}
