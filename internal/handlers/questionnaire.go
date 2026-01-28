package handlers

import (
  "encoding/json"
  "net/http"
  "strconv"
)

type questionOption struct {
  Value    string
  Label    string
  Selected bool
}

type questionnaireQuestion struct {
  ID      string
  Text    string
  Type    string
  Options []questionOption
}

var questionnaireDefinitions = []struct {
  ID       string
  Question string
  Type     string
  Options  []questionOption
}{
  {
    ID: "activity",
    Question: "Как часто вы занимаетесь физической активностью?",
    Type: "single",
    Options: []questionOption{
      {Value: "never", Label: "Не занимаюсь"},
      {Value: "1-2", Label: "1-2 раза в неделю"},
      {Value: "3-4", Label: "3-4 раза в неделю"},
      {Value: "5+", Label: "5+ раз в неделю"},
    },
  },
  {
    ID: "experience",
    Question: "Ваш опыт спортивных тренировок?",
    Type: "single",
    Options: []questionOption{
      {Value: "beginner", Label: "Новичок (меньше 6 месяцев)"},
      {Value: "intermediate", Label: "Средний (6 мес - 2 года)"},
      {Value: "advanced", Label: "Продвинутый (более 2 лет)"},
    },
  },
  {
    ID: "restrictions",
    Question: "Есть ли у вас ограничения по здоровью?",
    Type: "multiple",
    Options: []questionOption{
      {Value: "back", Label: "Проблемы со спиной"},
      {Value: "joints", Label: "Проблемы с суставами"},
      {Value: "cardio", Label: "Сердечно-сосудистые"},
      {Value: "none", Label: "Нет ограничений"},
    },
  },
  {
    ID: "goals",
    Question: "Какие цели вы хотите достичь?",
    Type: "multiple",
    Options: []questionOption{
      {Value: "rehab", Label: "Реабилитация"},
      {Value: "strength", Label: "Увеличение силы"},
      {Value: "flexibility", Label: "Улучшение гибкости"},
      {Value: "endurance", Label: "Выносливость"},
      {Value: "posture", Label: "Коррекция осанки"},
    },
  },
  {
    ID: "pain",
    Question: "Испытываете ли вы боль или дискомфорт?",
    Type: "multiple",
    Options: []questionOption{
      {Value: "neck", Label: "Шея"},
      {Value: "shoulders", Label: "Плечи"},
      {Value: "back", Label: "Спина"},
      {Value: "knees", Label: "Колени"},
      {Value: "none", Label: "Нет дискомфорта"},
    },
  },
}

func (a *App) Questionnaire(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  step := 0
  if value := r.URL.Query().Get("step"); value != "" {
    if parsed, err := strconv.Atoi(value); err == nil {
      step = parsed
    }
  }
  if step < 0 {
    step = 0
  }
  if step >= len(questionnaireDefinitions) {
    step = len(questionnaireDefinitions) - 1
  }

  answers := loadQuestionnaireAnswers(a, user.ID)
  definition := questionnaireDefinitions[step]
  options := make([]questionOption, 0, len(definition.Options))

  selected := map[string]bool{}
  if answer, ok := answers[definition.ID]; ok {
    switch typed := answer.(type) {
    case string:
      selected[typed] = true
    case []string:
      for _, value := range typed {
        selected[value] = true
      }
    case []any:
      for _, value := range typed {
        if str, ok := value.(string); ok {
          selected[str] = true
        }
      }
    }
  }

  for _, option := range definition.Options {
    option.Selected = selected[option.Value]
    options = append(options, option)
  }

  question := questionnaireQuestion{
    ID:      definition.ID,
    Text:    definition.Question,
    Type:    definition.Type,
    Options: options,
  }

  answered := len(selected) > 0

  data := map[string]any{
    "Question":   question,
    "Step":       step,
    "TotalSteps": len(questionnaireDefinitions),
    "Progress":   int(float64(step+1) / float64(len(questionnaireDefinitions)) * 100),
    "Answered":   answered,
    "CanBack":    step > 0,
  }

  a.renderFullPage(w, r, "questionnaire", "Оценка состояния", data)
}

func (a *App) QuestionnaireSubmit(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  step, _ := strconv.Atoi(r.FormValue("step"))
  if step < 0 {
    step = 0
  }
  if step >= len(questionnaireDefinitions) {
    step = len(questionnaireDefinitions) - 1
  }

  answers := loadQuestionnaireAnswers(a, user.ID)
  question := questionnaireDefinitions[step]
  values := r.Form["answer"]
  if question.Type == "single" {
    if len(values) > 0 {
      answers[question.ID] = values[0]
    }
  } else {
    answers[question.ID] = normalizeSelections(values)
  }

  if err := saveQuestionnaireAnswers(a, user.ID, answers); err != nil {
    http.Error(w, "server error", http.StatusInternalServerError)
    return
  }

  if step >= len(questionnaireDefinitions)-1 {
    fitnessLevel, _ := answers["experience"].(string)
    if fitnessLevel == "" {
      fitnessLevel = "beginner"
    }
    restrictions := toStringSlice(answers["restrictions"])
    goals := toStringSlice(answers["goals"])

    _, _ = a.DB.Exec(
      `update user_profiles
       set fitness_level = $1, restrictions = $2, goals = $3, updated_at = now()
       where user_id = $4`,
      fitnessLevel,
      restrictions,
      goals,
      user.ID,
    )

    a.setFlash(w, "Оценка завершена. Программа сформирована.")
    http.Redirect(w, r, "/program", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/questionnaire?step="+strconv.Itoa(step+1), http.StatusSeeOther)
}

func loadQuestionnaireAnswers(a *App, userID string) map[string]any {
  answers := map[string]any{}
  var raw []byte
  err := a.DB.QueryRow("select answers from questionnaire_responses where user_id = $1", userID).Scan(&raw)
  if err != nil {
    return answers
  }
  _ = json.Unmarshal(raw, &answers)
  return answers
}

func saveQuestionnaireAnswers(a *App, userID string, answers map[string]any) error {
  payload, err := json.Marshal(answers)
  if err != nil {
    return err
  }
  _, err = a.DB.Exec(
    `insert into questionnaire_responses (user_id, answers)
     values ($1, $2)
     on conflict (user_id) do update set answers = excluded.answers, updated_at = now()`,
    userID,
    payload,
  )
  return err
}

func toStringSlice(value any) []string {
  if value == nil {
    return []string{}
  }
  switch typed := value.(type) {
  case []string:
    return typed
  case []any:
    output := []string{}
    for _, item := range typed {
      if str, ok := item.(string); ok {
        output = append(output, str)
      }
    }
    return output
  case string:
    return []string{typed}
  default:
    return []string{}
  }
}

func normalizeSelections(values []string) []string {
  hasNone := false
  for _, value := range values {
    if value == "none" {
      hasNone = true
      break
    }
  }
  if hasNone {
    return []string{"none"}
  }
  return values
}
