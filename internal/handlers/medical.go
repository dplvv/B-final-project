package handlers

import (
  "net/http"
  "strings"
)

func (a *App) MedicalInfo(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  var chronic []string
  var injuries []string
  var meds []string
  var allergies []string
  var restrictions []string
  var doctorApproval bool
  var lastCheckup string

  _ = a.DB.QueryRow(
    `select chronic_diseases, injuries, medications, allergies, restrictions, doctor_approval, coalesce(last_checkup::text, '')
     from medical_info where user_id = $1`,
    user.ID,
  ).Scan(&chronic, &injuries, &meds, &allergies, &restrictions, &doctorApproval, &lastCheckup)

  data := map[string]any{
    "Chronic": chronic,
    "Injuries": injuries,
    "Medications": meds,
    "Allergies": allergies,
    "Restrictions": restrictions,
    "DoctorApproval": doctorApproval,
    "LastCheckup": lastCheckup,
  }

  a.renderPage(w, r, "medical_info", "Медицинская информация", "", data)
}

func (a *App) MedicalInfoUpdate(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  chronic := parseList(r.FormValue("chronic"))
  injuries := parseList(r.FormValue("injuries"))
  meds := parseList(r.FormValue("medications"))
  allergies := parseList(r.FormValue("allergies"))
  restrictions := parseList(r.FormValue("restrictions"))
  doctorApproval := r.FormValue("doctor_approval") == "on"
  lastCheckup := r.FormValue("last_checkup")

  _, _ = a.DB.Exec(
    `update medical_info
     set chronic_diseases = $1, injuries = $2, medications = $3, allergies = $4, restrictions = $5,
         doctor_approval = $6, last_checkup = nullif($7, '')::date, updated_at = now()
     where user_id = $8`,
    chronic,
    injuries,
    meds,
    allergies,
    restrictions,
    doctorApproval,
    lastCheckup,
    user.ID,
  )

  a.setFlash(w, "Медицинская информация обновлена")
  http.Redirect(w, r, "/medical-info", http.StatusSeeOther)
}

func parseList(value string) []string {
  items := []string{}
  for _, part := range strings.Split(value, "\n") {
    trimmed := strings.TrimSpace(part)
    if trimmed != "" {
      items = append(items, trimmed)
    }
  }
  if len(items) == 0 {
    for _, part := range strings.Split(value, ",") {
      trimmed := strings.TrimSpace(part)
      if trimmed != "" {
        items = append(items, trimmed)
      }
    }
  }
  return items
}
