package api

import "net/http"

type medicalPayload struct {
  ChronicDiseases []string `json:"chronic_diseases"`
  Injuries        []string `json:"injuries"`
  Medications     []string `json:"medications"`
  Allergies       []string `json:"allergies"`
  DoctorApproval  bool     `json:"doctor_approval"`
  LastCheckup     string   `json:"last_checkup"`
  Restrictions    []string `json:"restrictions"`
}

func (api *API) MedicalInfo(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var payload medicalPayload
  _ = api.DB.QueryRow(
    `select chronic_diseases, injuries, medications, allergies, doctor_approval,
            coalesce(last_checkup::text, ''), restrictions
     from medical_info where user_id = $1`,
    userID,
  ).Scan(
    &payload.ChronicDiseases,
    &payload.Injuries,
    &payload.Medications,
    &payload.Allergies,
    &payload.DoctorApproval,
    &payload.LastCheckup,
    &payload.Restrictions,
  )

  writeJSON(w, http.StatusOK, payload)
}

func (api *API) MedicalInfoUpdate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var payload medicalPayload
  if err := decodeJSON(r, &payload); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  _, err := api.DB.Exec(
    `update medical_info
     set chronic_diseases = $1,
         injuries = $2,
         medications = $3,
         allergies = $4,
         doctor_approval = $5,
         last_checkup = nullif($6, '')::date,
         restrictions = $7,
         updated_at = now()
     where user_id = $8`,
    payload.ChronicDiseases,
    payload.Injuries,
    payload.Medications,
    payload.Allergies,
    payload.DoctorApproval,
    payload.LastCheckup,
    payload.Restrictions,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusOK, map[string]any{"status": "saved"})
}

