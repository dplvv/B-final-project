package handlers

import "net/http"

func (a *App) renderPage(w http.ResponseWriter, r *http.Request, name string, title string, current string, data map[string]any) {
  base := a.baseData(w, r)
  for key, value := range data {
    base[key] = value
  }
  base["Title"] = title
  base["UseShell"] = true
  base["ShowNav"] = true
  base["Current"] = current
  _ = a.Renderer.Render(w, name, base)
}

func (a *App) renderFullPage(w http.ResponseWriter, r *http.Request, name string, title string, data map[string]any) {
  base := a.baseData(w, r)
  for key, value := range data {
    base[key] = value
  }
  base["Title"] = title
  base["UseShell"] = false
  base["ShowNav"] = false
  _ = a.Renderer.Render(w, name, base)
}
