package web

import (
  "fmt"
  "html/template"
  "io"
  "strings"
  "time"
)

type Renderer struct {
  templates *template.Template
}

func NewRenderer() (*Renderer, error) {
  var tmpl *template.Template
  funcMap := template.FuncMap{
    "upper": strings.ToUpper,
    "split": strings.Split,
    "render": func(name string, data any) template.HTML {
      if name == "" || tmpl == nil {
        return ""
      }
      var b strings.Builder
      if err := tmpl.ExecuteTemplate(&b, name, data); err != nil {
        return ""
      }
      return template.HTML(b.String())
    },
    "formatDate": func(value time.Time) string {
      if value.IsZero() {
        return ""
      }
      return value.Format("02.01.2006")
    },
    "add": func(a, b int) int {
      return a + b
    },
    "sub": func(a, b int) int {
      return a - b
    },
    "percent": func(part, total int) int {
      if total == 0 {
        return 0
      }
      return int(float64(part) / float64(total) * 100)
    },
    "seq": func(n int) []int {
      if n <= 0 {
        return []int{}
      }
      out := make([]int, n)
      for i := 0; i < n; i++ {
        out[i] = i
      }
      return out
    },
    "list": func(values ...string) []string {
      return values
    },
    "weekday": func(value time.Time) string {
      days := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
      index := int(value.Weekday())
      if index < 0 || index >= len(days) {
        return ""
      }
      return days[index]
    },
  }

  parsed, err := template.New("base").Funcs(funcMap).ParseFS(
    FS,
    "templates/base.html",
    "templates/partials/*.html",
    "templates/pages/*.html",
  )
  if err != nil {
    return nil, fmt.Errorf("parse templates: %w", err)
  }

  tmpl = parsed
  return &Renderer{templates: tmpl}, nil
}

func (r *Renderer) Render(w io.Writer, name string, data map[string]any) error {
  if data == nil {
    data = map[string]any{}
  }
  data["ContentTemplate"] = "content." + name
  return r.templates.ExecuteTemplate(w, "base", data)
}
