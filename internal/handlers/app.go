package handlers

import (
  "database/sql"

  "rehab-app/internal/config"
  "rehab-app/internal/middleware"
  "rehab-app/internal/web"
)

type App struct {
  DB      *sql.DB
  Renderer *web.Renderer
  Config  config.Config
  Sessions *middleware.SessionManager
}

func NewApp(db *sql.DB, renderer *web.Renderer, cfg config.Config, sessions *middleware.SessionManager) *App {
  return &App{DB: db, Renderer: renderer, Config: cfg, Sessions: sessions}
}
