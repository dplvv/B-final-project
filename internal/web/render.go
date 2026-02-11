package web

import (
  "fmt"
  "html/template"
  "io"
  "net/url"
  "strings"
  "time"
)

type Renderer struct {
  base *template.Template
}

func NewRenderer() (*Renderer, error) {
  funcMap := template.FuncMap{
    "eq": func(a, b any) bool { return a == b },
    "join": strings.Join,
    "urlpath": func(value string) string {
      return url.PathEscape(value)
    },
    "urlquery": func(value string) string {
      return url.QueryEscape(value)
    },
    "contains": func(list []string, value string) bool {
      for _, item := range list {
        if strings.EqualFold(item, value) {
          return true
        }
      }
      return false
    },
    "add": func(a, b int) int {
      return a + b
    },
    "percent": func(progress, total int) int {
      if total <= 0 {
        return 0
      }
      value := int(float64(progress) / float64(total) * 100)
      if value < 0 {
        return 0
      }
      if value > 100 {
        return 100
      }
      return value
    },
    "formatDate": func(t time.Time) string {
      if t.IsZero() {
        return ""
      }
      return t.Format("02.01.2006")
    },
    "formatDateTime": func(t time.Time) string {
      if t.IsZero() {
        return ""
      }
      return t.Format("02.01.2006 15:04")
    },
  }

  tmpl, err := template.New("base").Funcs(funcMap).ParseFS(
    FS,
    "templates/base.html",
    "templates/partials/*.html",
  )
  if err != nil {
    return nil, fmt.Errorf("parse templates: %w", err)
  }

  return &Renderer{base: tmpl}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data any) error {
  tmpl, err := r.base.Clone()
  if err != nil {
    return fmt.Errorf("clone templates: %w", err)
  }
  _, err = tmpl.ParseFS(FS, "templates/pages/"+name+".html")
  if err != nil {
    return fmt.Errorf("parse page template: %w", err)
  }
  return tmpl.ExecuteTemplate(w, name, data)
}
