package api

import (
  "net/http"
)

type profileUpdateRequest struct {
  Name         *string  `json:"name"`
  Department   *string  `json:"department"`
  Position     *string  `json:"position"`
  Age          *int     `json:"age"`
  FitnessLevel *string  `json:"fitness_level"`
  Restrictions *[]string `json:"restrictions"`
  Goals        *[]string `json:"goals"`
}

type questionnairePayload struct {
  FitnessLevel string            `json:"fitness_level"`
  Restrictions []string          `json:"restrictions"`
  Goals        []string          `json:"goals"`
  Answers      map[string]any    `json:"answers"`
}

func (api *API) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var payload profileUpdateRequest
  if err := decodeJSON(r, &payload); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if payload.Name != nil || payload.Department != nil || payload.Position != nil {
    name := ""
    department := ""
    position := ""
    if payload.Name != nil {
      name = *payload.Name
    }
    if payload.Department != nil {
      department = *payload.Department
    }
    if payload.Position != nil {
      position = *payload.Position
    }

    if _, err := api.DB.Exec(
      `update users set
         name = coalesce(nullif($1, ''), name),
         department = case when $2 <> '' then $2 else department end,
         position = case when $3 <> '' then $3 else position end,
         updated_at = now()
       where id = $4`,
      name,
      department,
      position,
      userID,
    ); err != nil {
      writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
      return
    }
  }

  if payload.Age != nil || payload.FitnessLevel != nil || payload.Restrictions != nil || payload.Goals != nil {
    _, err := api.DB.Exec(
      `update user_profiles
       set age = coalesce($1, age),
           fitness_level = coalesce(nullif($2, ''), fitness_level),
           restrictions = coalesce($3, restrictions),
           goals = coalesce($4, goals),
           updated_at = now()
       where user_id = $5`,
      payload.Age,
      payload.FitnessLevel,
      payload.Restrictions,
      payload.Goals,
      userID,
    )
    if err != nil {
      writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
      return
    }
  }

  api.Profile(w, r)
}

func (api *API) OnboardingComplete(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  _, _ = api.DB.Exec(
    `update user_profiles
     set onboarding_complete = true, updated_at = now()
     where user_id = $1`,
    userID,
  )
  writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (api *API) Questionnaire(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())
  var answers map[string]any
  _ = api.DB.QueryRow(
    `select answers from questionnaire_responses where user_id = $1`,
    userID,
  ).Scan(&answers)

  writeJSON(w, http.StatusOK, map[string]any{
    "answers": answers,
  })
}

func (api *API) QuestionnaireSubmit(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var payload questionnairePayload
  if err := decodeJSON(r, &payload); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if payload.Answers == nil {
    payload.Answers = map[string]any{}
  }

  _, err := api.DB.Exec(
    `insert into questionnaire_responses (user_id, answers)
     values ($1, $2)
     on conflict (user_id)
     do update set answers = excluded.answers, updated_at = now()`,
    userID,
    payload.Answers,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  _, err = api.DB.Exec(
    `update user_profiles
     set fitness_level = $1,
         restrictions = $2,
         goals = $3,
         updated_at = now()
     where user_id = $4`,
    payload.FitnessLevel,
    payload.Restrictions,
    payload.Goals,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "saved"})
}

