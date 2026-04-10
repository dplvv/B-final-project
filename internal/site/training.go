package site

import (
	"database/sql"
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"rehab-app/internal/middleware"
)

type questionnaireData struct {
	Goal           string   `json:"goal"`
	FitnessLevel   string   `json:"fitness_level"`
	DaysPerWeek    int      `json:"days_per_week"`
	SessionMinutes int      `json:"session_minutes"`
	Equipment      []string `json:"equipment"`
	Preferences    string   `json:"preferences"`
}

type planRecord struct {
	ID           string
	Goal         string
	Level        string
	Frequency    int
	Status       string
	PausedReason string
	CreatedAt    time.Time
}

type planWorkoutView struct {
	ID               string
	WorkoutID        string
	Name             string
	Description      string
	Difficulty       string
	Category         string
	Duration         int
	Week             int
	Day              int
	Intensity        int
	ScheduledDate    string
	ScheduledDateISO string
	ScheduledTime    string
	Status           string
	SkipReason       string
	SessionID        string
	StartAllowed     bool
	AvailableFrom    string
}

type planCalendarDay struct {
	ID            string
	WorkoutID     string
	Name          string
	ScheduledDate string
	ScheduledTime string
	Week          int
	Day           int
	Intensity     int
	Status        string
}

type planCalendarWeek struct {
	Week int
	Days []planCalendarDay
}

type planChangeView struct {
	ChangedAt string
	Reason    string
}

type leaderboardRow struct {
	Name         string
	Department   string
	Points       int
	Workouts     int
	Minutes      int
	LastWorkout  string
	AvgTolerance int
}

type achievementView struct {
	Title        string
	Description  string
	Icon         string
	Unlocked     bool
	Progress     int
	Total        int
	PointsReward int
}

type sessionFeedbackView struct {
	PerceivedExertion int
	Tolerance         int
	PainLevel         int
	Wellbeing         int
	Comment           string
}

func (s *Site) questionnairePage(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	w.Header().Set("Cache-Control", "no-store")
	data := s.baseData(r, "Опросник", "questionnaire")
	editMode := r.URL.Query().Get("edit") == "1"
	returnTo := strings.TrimSpace(r.URL.Query().Get("from"))
	if returnTo != "" && !strings.HasPrefix(returnTo, "/") {
		returnTo = ""
	}
	q, _ := s.loadQuestionnaire(user.ID)
	if q.SessionMinutes == 0 {
		q.SessionMinutes = sessionMinutesForLevel(resolveLevel(q.FitnessLevel))
	}
	data["Questionnaire"] = q
	data["Errors"] = map[string]string{}
	data["Restrictions"] = restrictionOptions()
	data["EquipmentOptions"] = equipmentOptions()
	data["PreferenceOptions"] = preferenceOptions()
	data["SelectedPreferences"] = parseCSV(q.Preferences)
	data["QuestionnaireComplete"] = !s.needsQuestionnaire(user.ID)
	data["EditMode"] = editMode
	data["ReturnTo"] = returnTo

	data["SelectedRestrictions"] = s.loadRestrictions(user.ID)
	data["DoctorApproval"] = s.loadDoctorApproval(user.ID)

	s.render(w, "questionnaire", data)
}

func (s *Site) questionnaireSubmit(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	previous, _ := s.loadQuestionnaire(user.ID)
	prevRestrictions := s.loadRestrictions(user.ID)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}

	days, _ := strconv.Atoi(r.FormValue("days_per_week"))
	equipment := r.Form["equipment"]
	if len(equipment) == 1 && strings.TrimSpace(equipment[0]) == "" {
		equipment = []string{}
	}
	equipment = normalizeEquipmentSelection(equipment)
	selectedPreferences := normalizePreferenceSelection(r.Form["preferences"])
	preferences := strings.Join(selectedPreferences, ", ")

	q := questionnaireData{
		Goal:           strings.TrimSpace(r.FormValue("goal")),
		FitnessLevel:   strings.TrimSpace(r.FormValue("fitness_level")),
		DaysPerWeek:    days,
		SessionMinutes: 0,
		Equipment:      equipment,
		Preferences:    preferences,
	}
	q.SessionMinutes = sessionMinutesForLevel(resolveLevel(q.FitnessLevel))

	errors := validateQuestionnaire(q)
	if len(errors) > 0 {
		data := s.baseData(r, "Опросник", "questionnaire")
		data["Questionnaire"] = q
		data["Errors"] = errors
		data["Restrictions"] = restrictionOptions()
		data["EquipmentOptions"] = equipmentOptions()
		data["PreferenceOptions"] = preferenceOptions()
		data["SelectedPreferences"] = parseCSV(q.Preferences)
		data["SelectedRestrictions"] = r.Form["restrictions"]
		data["DoctorApproval"] = s.loadDoctorApproval(user.ID)
		s.render(w, "questionnaire", data)
		return
	}

	if err := s.saveQuestionnaire(user.ID, q); err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	goals := []string{}
	if q.Goal != "" {
		goals = append(goals, q.Goal)
	}
	goalsCSV := strings.Join(goals, ",")
	if _, err := s.DB.Exec(
		`insert into user_profiles (user_id, fitness_level, goals, updated_at)
     values (
       $1,
       nullif($2, ''),
       case when trim($3) = '' then '{}'::text[] else string_to_array($3, ',') end,
       now()
     )
     on conflict (user_id)
     do update set fitness_level = excluded.fitness_level,
                   goals = excluded.goals,
                   updated_at = now()`,
		user.ID,
		q.FitnessLevel,
		goalsCSV,
	); err != nil {
		http.Error(w, "Ошибка сохранения параметров профиля", http.StatusInternalServerError)
		return
	}
	_, _ = s.DB.Exec(
		`update training_plans
     set goal = $1, updated_at = now()
     where user_id = $2 and status in ('active', 'paused')`,
		q.Goal,
		user.ID,
	)

	restrictions := normalizeRestrictionSelection(r.Form["restrictions"])
	restrictionsCSV := strings.Join(restrictions, ",")
	if _, err := s.DB.Exec(
		`insert into user_profiles (user_id, restrictions, updated_at)
     values (
       $1,
       case when trim($2) = '' then '{}'::text[] else string_to_array($2, ',') end,
       now()
     )
     on conflict (user_id)
     do update set restrictions = case when trim($2) = '' then '{}'::text[] else string_to_array($2, ',') end,
                   updated_at = now()`,
		user.ID,
		restrictionsCSV,
	); err != nil {
		http.Error(w, "Ошибка сохранения ограничений", http.StatusInternalServerError)
		return
	}

	if questionnaireChanged(previous, q, prevRestrictions, restrictions) {
		if plan, err := s.getActivePlan(user.ID); err == nil && plan != nil {
			if !adminPauseLockedForRole(user.Role, plan.Status, plan.PausedReason) {
				_, _ = s.DB.Exec(`update training_plans set status = 'archived', updated_at = now() where id = $1`, plan.ID)
				_, _ = s.ensurePlan(user.ID)
			}
		} else {
			_, _ = s.ensurePlan(user.ID)
		}
	}

	returnTo := strings.TrimSpace(r.FormValue("return_to"))
	if returnTo != "" && strings.HasPrefix(returnTo, "/") {
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/program", http.StatusSeeOther)
}

func (s *Site) planPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	needsQuestionnaire := s.needsQuestionnaire(user.ID)
	setupIncomplete := needsQuestionnaire

	data := s.baseData(r, "План", "program")
	data["SetupNeedsQuestionnaire"] = needsQuestionnaire
	data["SetupIncomplete"] = setupIncomplete
	data["Programs"] = s.fetchPrograms()
	if r.URL.Query().Get("doctor") == "1" {
		data["DoctorGate"] = true
	}
	if setupIncomplete {
		s.render(w, "program", data)
		return
	}

	plan, err := s.ensurePlan(user.ID)
	if err != nil || plan == nil {
		data["PlanError"] = "Не удалось сформировать план. Проверьте анкету."
		data["PlanWorkouts"] = []planWorkoutView{}
		s.render(w, "program", data)
		return
	}

	q, _ := s.loadQuestionnaire(user.ID)
	restrictions := s.loadRestrictions(user.ID)
	var doctorApproval bool
	_ = s.DB.QueryRow(`select doctor_approval from user_profiles where user_id = $1`, user.ID).Scan(&doctorApproval)

	var lastChange planChangeView
	var changedAt time.Time
	var reason string
	if err := s.DB.QueryRow(
		`select changed_at, reason from training_plan_changes where plan_id = $1 order by changed_at desc limit 1`,
		plan.ID,
	).Scan(&changedAt, &reason); err == nil {
		lastChange = planChangeView{ChangedAt: changedAt.Format("02.01.2006 15:04"), Reason: reason}
		data["PlanLastChange"] = lastChange
	}

	workouts := s.fetchPlanWorkouts(plan.ID)
	nextWorkout, _ := s.fetchNextPlanWorkout(user.ID)
	data["Plan"] = plan
	data["PlanWorkouts"] = workouts
	data["PlanCalendar"] = buildPlanCalendar(workouts)
	data["NextWorkout"] = nextWorkout
	data["PlanPaused"] = plan.Status == "paused"
	data["PlanPausedReason"] = plan.PausedReason
	data["PlanAdminLocked"] = adminPauseLockedForRole(user.Role, plan.Status, plan.PausedReason)
	data["PlanLaunchBlocked"] = planLaunchBlockedForRole(user.Role, plan.Status, plan.PausedReason)
	data["Questionnaire"] = q
	data["Restrictions"] = restrictions
	data["DoctorApproval"] = doctorApproval
	s.render(w, "program", data)
}

func (s *Site) planRegenerate(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if s.ensureOnboarding(w, r, user.ID) {
		return
	}

	if plan, err := s.getActivePlan(user.ID); err == nil && plan != nil {
		if adminPauseLockedForRole(user.Role, plan.Status, plan.PausedReason) {
			http.Redirect(w, r, "/program?paused=1", http.StatusSeeOther)
			return
		}
		_, _ = s.DB.Exec(`update training_plans set status = 'archived', updated_at = now() where id = $1`, plan.ID)
	}

	_, _ = s.ensurePlan(user.ID)
	http.Redirect(w, r, "/program", http.StatusSeeOther)
}

func (s *Site) planWorkoutStart(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	s.expireOverduePlanWorkouts(user.ID, time.Now())
	scheduleDate, scheduleTime := scheduleDateAndTime(time.Now())
	if !s.requireDoctorApproval(w, r, user.ID) {
		return
	}
	planWorkoutID := chi.URLParam(r, "id")
	if planWorkoutID == "" {
		http.NotFound(w, r)
		return
	}

	var workoutID string
	var planID string
	var status string
	var sessionID sql.NullString
	var startAllowed bool
	err := s.DB.QueryRow(
		`select workout_id, plan_id, status, session_id,
            case
              when scheduled_date is null then true
              when scheduled_date <> $2::date then false
              when scheduled_time is null then true
              else scheduled_time <= $3::time
            end
     from training_plan_workouts
     where id = $1`,
		planWorkoutID,
		scheduleDate,
		scheduleTime,
	).Scan(&workoutID, &planID, &status, &sessionID, &startAllowed)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if !s.planOwnedByUser(planID, user.ID) {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}
	if status != "pending" && status != "in_progress" {
		http.Redirect(w, r, "/program", http.StatusSeeOther)
		return
	}
	if !startAllowed {
		http.Redirect(w, r, "/program", http.StatusSeeOther)
		return
	}

	var planStatus string
	var pausedReason string
	_ = s.DB.QueryRow(
		`select status, coalesce(paused_reason, '') from training_plans where id = $1`,
		planID,
	).Scan(&planStatus, &pausedReason)
	if planLaunchBlockedForRole(user.Role, planStatus, pausedReason) {
		http.Redirect(w, r, "/program?paused=1", http.StatusSeeOther)
		return
	}

	if status == "completed" && sessionID.Valid {
		http.Redirect(w, r, "/sessions/"+sessionID.String, http.StatusSeeOther)
		return
	}

	if sessionID.Valid {
		http.Redirect(w, r, "/sessions/"+sessionID.String, http.StatusSeeOther)
		return
	}

	sessionIDValue, err := s.createWorkoutSession(user.ID, workoutID, planWorkoutID)
	if err != nil {
		http.Redirect(w, r, "/program", http.StatusSeeOther)
		return
	}

	_, _ = s.DB.Exec(`update training_plan_workouts set status = 'in_progress' where id = $1`, planWorkoutID)
	http.Redirect(w, r, "/sessions/"+sessionIDValue, http.StatusSeeOther)
}

func (s *Site) planWorkoutSkip(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	planWorkoutID := chi.URLParam(r, "id")
	if planWorkoutID == "" {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/program", http.StatusSeeOther)
		return
	}

	reason := strings.TrimSpace(r.FormValue("skip_reason"))
	if reason == "" {
		reason = "Пропуск без указания причины"
	}

	var planID string
	var status string
	err := s.DB.QueryRow(`select plan_id, status from training_plan_workouts where id = $1`, planWorkoutID).Scan(&planID, &status)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if !s.planOwnedByUser(planID, user.ID) {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}
	if status == "completed" || status == "in_progress" {
		http.Redirect(w, r, "/program", http.StatusSeeOther)
		return
	}

	before := s.planSnapshot(planID)
	_, _ = s.DB.Exec(
		`update training_plan_workouts set status = 'skipped', skip_reason = $1 where id = $2`,
		reason,
		planWorkoutID,
	)
	after := s.planSnapshot(planID)
	s.logPlanChange(user.ID, planID, "skip", "Пропуск тренировки", before, after)
	s.applyAdaptation(user.ID, planID, "skip")

	http.Redirect(w, r, "/program", http.StatusSeeOther)
}

func (s *Site) sessionFeedback(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.NotFound(w, r)
		return
	}

	var ownerID string
	var completedAt sql.NullTime
	err := s.DB.QueryRow(
		`select user_id, completed_at from workout_sessions where id = $1`,
		sessionID,
	).Scan(&ownerID, &completedAt)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if ownerID != user.ID {
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
		return
	}
	if !completedAt.Valid {
		http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
		return
	}

	perceived, _ := strconv.Atoi(r.FormValue("perceived_exertion"))
	tolerance, _ := strconv.Atoi(r.FormValue("tolerance"))
	pain, _ := strconv.Atoi(r.FormValue("pain_level"))
	wellbeing, _ := strconv.Atoi(r.FormValue("wellbeing"))
	comment := strings.TrimSpace(r.FormValue("comment"))

	_, _ = s.DB.Exec(
		`insert into workout_session_feedback (session_id, user_id, perceived_exertion, tolerance, pain_level, wellbeing, comment)
     values ($1, $2, $3, $4, $5, $6, $7)
     on conflict (session_id)
     do update set perceived_exertion = excluded.perceived_exertion,
                   tolerance = excluded.tolerance,
                   pain_level = excluded.pain_level,
                   wellbeing = excluded.wellbeing,
                   comment = excluded.comment`,
		sessionID,
		user.ID,
		perceived,
		tolerance,
		pain,
		wellbeing,
		comment,
	)

	planID := s.planIDBySession(sessionID)
	if planID == "" {
		planID = s.attachSessionToPlan(user.ID, sessionID)
	}
	if planID != "" {
		s.applyAdaptation(user.ID, planID, "feedback")
	}

	http.Redirect(w, r, "/sessions/"+sessionID, http.StatusSeeOther)
}

func (s *Site) loadNotifications(userID string) []planChangeView {
	type historyEntry struct {
		When   time.Time
		Reason string
	}
	entries := []historyEntry{}
	clearedAt := time.Unix(0, 0).UTC()
	_ = s.DB.QueryRow(
		`select coalesce(notifications_cleared_at, to_timestamp(0))
     from user_profiles
     where user_id = $1`,
		userID,
	).Scan(&clearedAt)

	role := ""
	department := ""
	_ = s.DB.QueryRow(
		`select role, coalesce(department, '')
     from users
     where id = $1`,
		userID,
	).Scan(&role, &department)

	var rows, err = s.DB.Query(
		`select changed_at, reason
     from training_plan_changes
     where user_id = $1 and changed_at > $2
     order by changed_at desc`,
		userID,
		clearedAt,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var changedAt time.Time
			var reason string
			_ = rows.Scan(&changedAt, &reason)
			entries = append(entries, historyEntry{When: changedAt, Reason: reason})
		}
	}

	rows, err = s.DB.Query(
		`select rr.status, r.title, coalesce(rr.handled_at, rr.redeemed_at)
     from reward_redemptions rr
     join rewards r on r.id = rr.reward_id
     where rr.user_id = $1
       and rr.status in ('approved', 'rejected')
       and coalesce(rr.handled_at, rr.redeemed_at) > $2
     order by coalesce(rr.handled_at, rr.redeemed_at) desc`,
		userID,
		clearedAt,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var title string
			var when time.Time
			_ = rows.Scan(&status, &title, &when)
			reason := "Поощрение «" + title + "»"
			if status == "approved" {
				reason += " одобрено"
			} else {
				reason += " отклонено"
			}
			entries = append(entries, historyEntry{When: when, Reason: reason})
		}
	}

	rows, err = s.DB.Query(
		`select stm.created_at, coalesce(st.subject, '')
     from support_ticket_messages stm
     join support_tickets st on st.id = stm.ticket_id
     where st.user_id = $1
       and stm.sender_role = 'admin'
       and stm.created_at > $2
     order by stm.created_at desc`,
		userID,
		clearedAt,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var createdAt time.Time
			var subject string
			_ = rows.Scan(&createdAt, &subject)
			subject = strings.TrimSpace(subject)
			reason := "Получен ответ по обращению в поддержку"
			if subject != "" {
				reason += ": «" + subject + "»"
			}
			entries = append(entries, historyEntry{When: createdAt, Reason: reason})
		}
	}

	rows, err = s.DB.Query(
		`select created_at, coalesce(source, ''), points, coalesce(reason, '')
     from user_point_events
     where user_id = $1
       and created_at > $2
     order by created_at desc`,
		userID,
		clearedAt,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var createdAt time.Time
			var source string
			var points int
			var reasonRaw string
			_ = rows.Scan(&createdAt, &source, &points, &reasonRaw)

			pointsText := strconv.Itoa(points)
			if points > 0 {
				pointsText = "+" + pointsText
			}
			reason := "Изменение баллов: " + pointsText
			switch strings.TrimSpace(source) {
			case "manager_award":
				reason = "Руководитель начислил " + pointsText + " баллов"
			case "achievement":
				reason = "Начислено " + pointsText + " баллов за достижение"
			case "workout_completion":
				reason = "Начислено " + pointsText + " баллов за завершение тренировки"
			case "reward_rejection_refund":
				reason = "Возвращено " + pointsText + " баллов по заявке на поощрение"
			}
			reasonRaw = strings.TrimSpace(reasonRaw)
			if reasonRaw != "" {
				reason += ": " + reasonRaw
			}
			entries = append(entries, historyEntry{When: createdAt, Reason: reason})
		}
	}

	reminderSettings := s.loadReminderSettings(userID)
	if reminderSettings.Enabled {
		now := time.Now()
		location := now.Location()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
		tomorrow := today.AddDate(0, 0, 1)
		inTwoDays := today.AddDate(0, 0, 2)

		selectedWeekdays := map[int]bool{}
		for _, dayRaw := range reminderSettings.Weekdays {
			day, parseErr := strconv.Atoi(strings.TrimSpace(dayRaw))
			if parseErr == nil && day >= 1 && day <= 7 {
				selectedWeekdays[day] = true
			}
		}
		if len(selectedWeekdays) == 0 {
			selectedWeekdays[1] = true
			selectedWeekdays[2] = true
			selectedWeekdays[3] = true
			selectedWeekdays[4] = true
			selectedWeekdays[5] = true
		}

		reminderHour := 9
		reminderMinute := 0
		if parsedTime, parseErr := time.Parse("15:04", reminderSettings.ReminderTime); parseErr == nil {
			reminderHour = parsedTime.Hour()
			reminderMinute = parsedTime.Minute()
		}

		rows, err = s.DB.Query(
			`select pw.scheduled_date,
              coalesce(to_char(pw.scheduled_time, 'HH24:MI'), ''),
              w.name,
              pw.intensity
       from training_plan_workouts pw
       join training_plans tp on tp.id = pw.plan_id
       join workouts w on w.id = pw.workout_id
       where tp.user_id = $1
         and tp.status = 'active'
         and pw.status in ('pending', 'in_progress')
         and pw.scheduled_date is not null
         and pw.scheduled_date >= $2
         and pw.scheduled_date <= $3
       order by pw.scheduled_date, coalesce(pw.scheduled_time, time '23:59'), pw.week, pw.day`,
			userID,
			today,
			inTwoDays,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var scheduledDate time.Time
				var scheduledTime string
				var workoutName string
				var intensity int
				_ = rows.Scan(&scheduledDate, &scheduledTime, &workoutName, &intensity)

				scheduledLocal := scheduledDate.In(location)
				scheduledDay := time.Date(scheduledLocal.Year(), scheduledLocal.Month(), scheduledLocal.Day(), 0, 0, 0, 0, location)
				weekday := int(scheduledDay.Weekday())
				if weekday == 0 {
					weekday = 7
				}
				if !selectedWeekdays[weekday] {
					continue
				}

				reminderMoment := time.Date(
					scheduledDay.Year(),
					scheduledDay.Month(),
					scheduledDay.Day(),
					reminderHour,
					reminderMinute,
					0,
					0,
					location,
				)
				scheduledTime = strings.TrimSpace(scheduledTime)
				if scheduledTime != "" {
					if parsedScheduledTime, parseErr := time.Parse("15:04", scheduledTime); parseErr == nil {
						workoutMoment := time.Date(
							scheduledDay.Year(),
							scheduledDay.Month(),
							scheduledDay.Day(),
							parsedScheduledTime.Hour(),
							parsedScheduledTime.Minute(),
							0,
							0,
							location,
						)
						// Приоритет у выбранного пользователем времени напоминания.
						// Если оно наступает после старта тренировки, смещаем напоминание
						// за 30 минут до тренировки, чтобы оно не приходило "задним числом".
						if !reminderMoment.Before(workoutMoment) {
							adjusted := workoutMoment.Add(-30 * time.Minute)
							if adjusted.Before(scheduledDay) {
								adjusted = workoutMoment
							}
							reminderMoment = adjusted
						}
					}
				}

				if !reminderMoment.After(clearedAt) {
					continue
				}

				dayLabel := scheduledDay.Format("02.01")
				switch {
				case scheduledDay.Equal(today):
					dayLabel = "сегодня"
				case scheduledDay.Equal(tomorrow):
					dayLabel = "завтра"
				}

				reason := "Напоминание: " + dayLabel + " тренировка «" + workoutName + "»"
				if scheduledTime != "" {
					reason += " в " + scheduledTime
				}
				if intensity > 0 {
					reason += " (интенсивность " + strconv.Itoa(intensity) + ")"
				}
				entries = append(entries, historyEntry{When: reminderMoment, Reason: reason})
			}
		}
	}

	if role == "admin" {
		rows, err = s.DB.Query(
			`select st.created_at, coalesce(u.name, ''), coalesce(st.subject, '')
       from support_tickets st
       left join users u on u.id = st.user_id
       where st.status = 'open'
         and st.created_at > $1
       order by st.created_at desc`,
			clearedAt,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var createdAt time.Time
				var employeeName string
				var subject string
				_ = rows.Scan(&createdAt, &employeeName, &subject)
				employeeName = strings.TrimSpace(employeeName)
				subject = strings.TrimSpace(subject)
				reason := "Новое обращение в поддержку"
				if employeeName != "" {
					reason += " от " + employeeName
				}
				if subject != "" {
					reason += ": «" + subject + "»"
				}
				entries = append(entries, historyEntry{When: createdAt, Reason: reason})
			}
		}

		rows, err = s.DB.Query(
			`select pr.created_at, coalesce(u.name, ''), coalesce(u.employee_id, '')
       from password_reset_requests pr
       join users u on u.id = pr.user_id
       where pr.status = 'open'
         and pr.created_at > $1
       order by pr.created_at desc`,
			clearedAt,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var createdAt time.Time
				var employeeName string
				var employeeID string
				_ = rows.Scan(&createdAt, &employeeName, &employeeID)
				employeeName = strings.TrimSpace(employeeName)
				employeeID = strings.TrimSpace(employeeID)
				reason := "Запрос на сброс пароля"
				if employeeName != "" {
					reason += ": " + employeeName
				}
				if employeeID != "" {
					reason += " (табельный № " + employeeID + ")"
				}
				entries = append(entries, historyEntry{When: createdAt, Reason: reason})
			}
		}
	}

	if role == "manager" {
		rows, err = s.DB.Query(
			`select rr.redeemed_at, coalesce(u.name, ''), coalesce(u.employee_id, ''), coalesce(r.title, '')
       from reward_redemptions rr
       join users u on u.id = rr.user_id
       join rewards r on r.id = rr.reward_id
       where rr.status = 'pending'
         and rr.redeemed_at > $1
         and ($2 = '' or coalesce(u.department, '') = $2)
       order by rr.redeemed_at desc`,
			clearedAt,
			department,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var redeemedAt time.Time
				var employeeName string
				var employeeID string
				var rewardTitle string
				_ = rows.Scan(&redeemedAt, &employeeName, &employeeID, &rewardTitle)
				employeeName = strings.TrimSpace(employeeName)
				employeeID = strings.TrimSpace(employeeID)
				rewardTitle = strings.TrimSpace(rewardTitle)
				reason := "Новая заявка на поощрение"
				if employeeName != "" {
					reason += ": " + employeeName
				}
				if employeeID != "" {
					reason += " (табельный № " + employeeID + ")"
				}
				if rewardTitle != "" {
					reason += " — «" + rewardTitle + "»"
				}
				entries = append(entries, historyEntry{When: redeemedAt, Reason: reason})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].When.After(entries[j].When)
	})

	notifications := make([]planChangeView, 0, len(entries))
	for _, entry := range entries {
		notifications = append(notifications, planChangeView{
			ChangedAt: entry.When.Format("02.01.2006 15:04"),
			Reason:    entry.Reason,
		})
	}
	return notifications
}

func (s *Site) leaderboard(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(
		`select u.name, coalesce(u.department, ''),
            coalesce(p.points_total, 0) as points,
            coalesce(count(ws.id), 0) as workouts,
            coalesce(sum(ws.duration_minutes), 0) as minutes,
            coalesce(max(ws.completed_at), to_timestamp(0)) as last_workout,
            coalesce(f.avg_tolerance, 0) as avg_tolerance
     from users u
     left join user_points p on p.user_id = u.id
     left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
     left join (
       select user_id, avg(coalesce(tolerance, 0)) as avg_tolerance
       from workout_session_feedback
       group by user_id
     ) f on f.user_id = u.id
     group by u.id, p.points_total, f.avg_tolerance
     order by points desc, workouts desc, minutes desc, u.name`,
	)
	list := []leaderboardRow{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var row leaderboardRow
			var last time.Time
			var avgTolerance float64
			_ = rows.Scan(&row.Name, &row.Department, &row.Points, &row.Workouts, &row.Minutes, &last, &avgTolerance)
			row.AvgTolerance = int(avgTolerance + 0.5)
			if !last.IsZero() && last.Unix() > 0 {
				row.LastWorkout = last.Format("02.01.2006")
			}
			list = append(list, row)
		}
	}

	data := s.baseData(r, "Рейтинг", "leaderboard")
	data["Leaderboard"] = list
	s.render(w, "leaderboard", data)
}

func (s *Site) achievementsPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	s.updateAchievements(user.ID)
	rows, err := s.DB.Query(
		`select a.title, a.description, a.icon, a.points_reward,
            coalesce(ua.unlocked, false), coalesce(ua.progress, 0), coalesce(ua.total, 0)
     from achievements a
     left join user_achievements ua on ua.achievement_id = a.id and ua.user_id = $1
     order by coalesce(ua.unlocked, false), a.title`,
		user.ID,
	)
	views := []achievementView{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var v achievementView
			_ = rows.Scan(&v.Title, &v.Description, &v.Icon, &v.PointsReward, &v.Unlocked, &v.Progress, &v.Total)
			views = append(views, v)
		}
	}

	data := s.baseData(r, "Достижения", "achievements")
	data["Achievements"] = views
	s.render(w, "achievements", data)
}

func (s *Site) ensureOnboarding(w http.ResponseWriter, r *http.Request, userID string) bool {
	if s.needsQuestionnaire(userID) {
		http.Redirect(w, r, "/questionnaire", http.StatusSeeOther)
		return true
	}
	return false
}

func (s *Site) needsQuestionnaire(userID string) bool {
	q, err := s.loadQuestionnaire(userID)
	if err != nil {
		return true
	}
	if q.Goal == "" || q.FitnessLevel == "" || q.DaysPerWeek == 0 {
		return true
	}
	return false
}

func validateQuestionnaire(q questionnaireData) map[string]string {
	errors := map[string]string{}
	if q.Goal == "" {
		errors["goal"] = "Выберите цель"
	}
	if q.FitnessLevel == "" {
		errors["fitness_level"] = "Укажите уровень"
	}
	if q.DaysPerWeek < 1 || q.DaysPerWeek > 7 {
		errors["days_per_week"] = "Частота 1-7"
	}
	return errors
}

func (s *Site) loadQuestionnaire(userID string) (questionnaireData, error) {
	var raw []byte
	err := s.DB.QueryRow(`select answers from user_profiles where user_id = $1`, userID).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return questionnaireData{}, nil
		}
		return questionnaireData{}, err
	}
	if len(raw) == 0 {
		return questionnaireData{}, nil
	}
	var q questionnaireData
	_ = json.Unmarshal(raw, &q)
	return q, nil
}

func (s *Site) saveQuestionnaire(userID string, q questionnaireData) error {
	payload, err := json.Marshal(q)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(
		`insert into user_profiles (user_id, answers, updated_at)
     values ($1, $2, now())
     on conflict (user_id)
     do update set answers = excluded.answers, updated_at = now()`,
		userID,
		payload,
	)
	return err
}

func (s *Site) getActivePlan(userID string) (*planRecord, error) {
	var plan planRecord
	err := s.DB.QueryRow(
		`select id, goal, level, frequency, status, coalesce(paused_reason, ''), created_at
     from training_plans
     where user_id = $1 and status in ('active', 'paused')
     order by created_at desc
     limit 1`,
		userID,
	).Scan(&plan.ID, &plan.Goal, &plan.Level, &plan.Frequency, &plan.Status, &plan.PausedReason, &plan.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (s *Site) ensurePlan(userID string) (*planRecord, error) {
	plan, err := s.getActivePlan(userID)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		return plan, nil
	}

	created, err := s.generatePlan(userID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Site) generatePlan(userID string) (*planRecord, error) {
	q, err := s.loadQuestionnaire(userID)
	if err != nil {
		return nil, err
	}
	restrictions := s.loadRestrictions(userID)
	doctorApproval := s.loadDoctorApproval(userID)

	level := resolvePlanLevel(q.FitnessLevel, doctorApproval)
	frequency := normalizePlanFrequency(q.DaysPerWeek)

	goalCategories := categoriesForGoal(q.Goal)
	preferenceCategories := categoriesFromPreferences(q.Preferences)
	categories := mergeCategories(goalCategories, preferenceCategories)
	availableEquipment, noEquipmentOnly := resolveAvailableEquipment(q.Equipment, q.Preferences)

	workouts := s.selectPlanWorkouts(level, frequency, categories, restrictions, availableEquipment, noEquipmentOnly)
	targetMinutes := targetPlanMinutes(level, doctorApproval)
	if filtered := filterWorkoutsByDuration(workouts, targetMinutes, 10); len(filtered) > 0 {
		workouts = filtered
	}
	rand.Seed(time.Now().UnixNano())
	workouts = shuffleWorkouts(workouts)

	planID := ""
	status := "active"
	pausedReason := ""
	if len(workouts) == 0 {
		status = "paused"
		pausedReason = "Нет подходящих тренировок по ограничениям. Нужна консультация."
	}
	err = s.DB.QueryRow(
		`insert into training_plans (user_id, goal, level, frequency, status, paused_reason)
     values ($1, $2, $3, $4, $5, $6)
     returning id`,
		userID,
		q.Goal,
		level,
		frequency,
		status,
		nullIfEmpty(pausedReason),
	).Scan(&planID)
	if err != nil {
		return nil, err
	}

	if len(workouts) == 0 {
		s.logPlanChange(userID, planID, "no_workouts", "Нет подходящих тренировок по ограничениям", nil, s.planSnapshot(planID))
		return &planRecord{
			ID:           planID,
			Goal:         q.Goal,
			Level:        level,
			Frequency:    frequency,
			Status:       status,
			PausedReason: pausedReason,
			CreatedAt:    time.Now(),
		}, nil
	}

	start := nextWeekStart(time.Now())
	weeks := 4
	offsets := weeklyOffsets(frequency)
	if len(offsets) == 0 {
		offsets = []int{0}
	}

	for week := 1; week <= weeks; week++ {
		weekList := rotateWorkouts(workouts, week-1)
		for idx, offset := range offsets {
			workout := weekList[idx%len(weekList)]
			scheduled := start.AddDate(0, 0, (week-1)*7+offset)
			scheduledTime := defaultPlanWorkoutTime(offset)
			_, _ = s.DB.Exec(
				`insert into training_plan_workouts (plan_id, workout_id, week, day, scheduled_date, scheduled_time, intensity)
         values ($1, $2, $3, $4, $5, $6::time, 1)`,
				planID,
				workout.ID,
				week,
				idx+1,
				scheduled,
				scheduledTime,
			)
		}
	}

	s.logPlanChange(userID, planID, "initial", "Первичный подбор программы", nil, s.planSnapshot(planID))

	return &planRecord{
		ID:        planID,
		Goal:      q.Goal,
		Level:     level,
		Frequency: frequency,
		Status:    status,
		CreatedAt: time.Now(),
	}, nil
}

func normalizePlanFrequency(daysPerWeek int) int {
	frequency := daysPerWeek
	if frequency <= 0 {
		frequency = 3
	}
	if frequency > 7 {
		frequency = 7
	}
	return frequency
}

func resolvePlanLevel(fitnessLevel string, doctorApproval bool) string {
	level := resolveLevel(fitnessLevel)
	if !doctorApproval {
		return "Легкая"
	}
	return level
}

func targetPlanMinutes(level string, doctorApproval bool) int {
	targetMinutes := sessionMinutesForLevel(level)
	if !doctorApproval && targetMinutes > 30 {
		return 30
	}
	return targetMinutes
}

func resolveAvailableEquipment(questionnaireEquipment []string, preferences string) ([]string, bool) {
	availableEquipment := normalizeEquipmentSelection(questionnaireEquipment)
	noEquipmentOnly := containsEquipmentValue(availableEquipment, "Без инвентаря")
	if noEquipmentOnly {
		return []string{}, true
	}
	if len(availableEquipment) == 0 && !prefersNoEquipment(preferences) {
		return []string{"Коврик"}, false
	}
	return availableEquipment, false
}

func (s *Site) selectPlanWorkouts(level string, frequency int, categories, restrictions, availableEquipment []string, noEquipmentOnly bool) []workoutCard {
	workouts := s.fetchWorkouts(level, categories, restrictions, availableEquipment)
	minWorkouts := frequency
	if minWorkouts < 2 {
		minWorkouts = 2
	}

	if len(workouts) < minWorkouts && len(categories) > 0 {
		workouts = mergeWorkoutCards(workouts, s.fetchWorkouts(level, []string{}, restrictions, availableEquipment))
	}

	if len(workouts) < minWorkouts && !noEquipmentOnly {
		relaxedEquipment := mergeEquipmentOptions(availableEquipment, []string{"Коврик", "Стул"})
		workouts = mergeWorkoutCards(workouts, s.fetchWorkouts(level, categories, restrictions, relaxedEquipment))
		if len(workouts) < minWorkouts && len(categories) > 0 {
			workouts = mergeWorkoutCards(workouts, s.fetchWorkouts(level, []string{}, restrictions, relaxedEquipment))
		}
	}

	return workouts
}

func (s *Site) fetchPlanWorkouts(planID string) []planWorkoutView {
	scheduleDate, scheduleTime := scheduleDateAndTime(time.Now())
	rows, err := s.DB.Query(
		`select pw.id, w.id, w.name, w.description, w.difficulty, coalesce(w.category, ''), w.duration_minutes,
            pw.week, pw.day, pw.scheduled_date, coalesce(to_char(pw.scheduled_time, 'HH24:MI'), ''),
            pw.intensity, pw.status, coalesce(pw.skip_reason, ''), coalesce(pw.session_id::text, ''),
            case
              when pw.scheduled_date is null then true
              when pw.scheduled_date <> $2::date then false
              when pw.scheduled_time is null then true
              else pw.scheduled_time <= $3::time
            end
     from training_plan_workouts pw
     join workouts w on w.id = pw.workout_id
     where pw.plan_id = $1
     order by pw.week, pw.day`,
		planID,
		scheduleDate,
		scheduleTime,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	list := []planWorkoutView{}
	for rows.Next() {
		var v planWorkoutView
		var date sql.NullTime
		var scheduledTime string
		_ = rows.Scan(&v.ID, &v.WorkoutID, &v.Name, &v.Description, &v.Difficulty, &v.Category, &v.Duration,
			&v.Week, &v.Day, &date, &scheduledTime, &v.Intensity, &v.Status, &v.SkipReason, &v.SessionID, &v.StartAllowed)
		if date.Valid {
			v.ScheduledDate = date.Time.Format("02.01.2006")
			v.ScheduledDateISO = date.Time.Format("2006-01-02")
		}
		v.ScheduledTime = strings.TrimSpace(scheduledTime)
		if !v.StartAllowed && v.ScheduledTime != "" {
			v.AvailableFrom = "Доступно с " + v.ScheduledTime
		}
		list = append(list, v)
	}
	return list
}

func buildPlanCalendar(items []planWorkoutView) []planCalendarWeek {
	weeks := map[int][]planCalendarDay{}
	order := []int{}
	for _, item := range items {
		if _, ok := weeks[item.Week]; !ok {
			order = append(order, item.Week)
		}
		weeks[item.Week] = append(weeks[item.Week], planCalendarDay{
			ID:            item.ID,
			WorkoutID:     item.WorkoutID,
			Name:          item.Name,
			ScheduledDate: item.ScheduledDate,
			ScheduledTime: item.ScheduledTime,
			Week:          item.Week,
			Day:           item.Day,
			Intensity:     item.Intensity,
			Status:        item.Status,
		})
	}
	sort.Ints(order)
	calendar := []planCalendarWeek{}
	for _, week := range order {
		days := weeks[week]
		sort.Slice(days, func(i, j int) bool { return days[i].Day < days[j].Day })
		calendar = append(calendar, planCalendarWeek{
			Week: week,
			Days: days,
		})
	}
	return calendar
}

func (s *Site) fetchNextPlanWorkout(userID string) (*planWorkoutView, error) {
	scheduleDate, scheduleTime := scheduleDateAndTime(time.Now())
	var v planWorkoutView
	var date sql.NullTime
	err := s.DB.QueryRow(
		`select pw.id, w.id, w.name, w.description, w.difficulty, coalesce(w.category, ''), w.duration_minutes,
            pw.week, pw.day, pw.scheduled_date, coalesce(to_char(pw.scheduled_time, 'HH24:MI'), ''),
            pw.intensity, pw.status, coalesce(pw.skip_reason, ''), coalesce(pw.session_id::text, ''),
            case
              when pw.scheduled_date is null then true
              when pw.scheduled_date <> $2::date then false
              when pw.scheduled_time is null then true
              else pw.scheduled_time <= $3::time
            end
     from training_plan_workouts pw
     join training_plans tp on tp.id = pw.plan_id
     join workouts w on w.id = pw.workout_id
     where tp.user_id = $1 and tp.status in ('active', 'paused')
       and pw.status in ('pending', 'in_progress')
       and (pw.scheduled_date is null or pw.scheduled_date >= $2::date)
     order by pw.scheduled_date nulls last, coalesce(pw.scheduled_time, time '23:59'), pw.week, pw.day
     limit 1`,
		userID,
		scheduleDate,
		scheduleTime,
	).Scan(&v.ID, &v.WorkoutID, &v.Name, &v.Description, &v.Difficulty, &v.Category, &v.Duration,
		&v.Week, &v.Day, &date, &v.ScheduledTime, &v.Intensity, &v.Status, &v.SkipReason, &v.SessionID, &v.StartAllowed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if date.Valid {
		v.ScheduledDate = date.Time.Format("02.01.2006")
		v.ScheduledDateISO = date.Time.Format("2006-01-02")
	}
	v.ScheduledTime = strings.TrimSpace(v.ScheduledTime)
	if !v.StartAllowed && v.ScheduledTime != "" {
		v.AvailableFrom = "Доступно с " + v.ScheduledTime
	}
	return &v, nil
}

func (s *Site) expireOverduePlanWorkouts(userID string, now time.Time) int {
	if strings.TrimSpace(userID) == "" {
		return 0
	}

	cutoff := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	type overdueWorkout struct {
		PlanID        string
		PlanWorkoutID string
		SessionID     string
	}
	overdue := []overdueWorkout{}

	rows, err := s.DB.Query(
		`select pw.plan_id, pw.id, coalesce(pw.session_id::text, '')
     from training_plan_workouts pw
     join training_plans tp on tp.id = pw.plan_id
     where tp.user_id = $1
       and tp.status = 'active'
       and pw.status in ('pending', 'in_progress')
       and pw.scheduled_date is not null
       and pw.scheduled_date < $2
     order by pw.plan_id, pw.scheduled_date, pw.week, pw.day`,
		userID,
		cutoff,
	)
	if err != nil {
		return 0
	}
	defer rows.Close()

	for rows.Next() {
		var item overdueWorkout
		_ = rows.Scan(&item.PlanID, &item.PlanWorkoutID, &item.SessionID)
		overdue = append(overdue, item)
	}
	if len(overdue) == 0 {
		return 0
	}

	reason := "Автопропуск: тренировка не выполнена в назначенный день"
	beforeSnapshots := map[string]json.RawMessage{}
	skippedByPlan := map[string]int{}
	totalSkipped := 0

	for _, item := range overdue {
		if _, exists := beforeSnapshots[item.PlanID]; !exists {
			beforeSnapshots[item.PlanID] = s.planSnapshot(item.PlanID)
		}
		if item.SessionID != "" {
			_, _ = s.DB.Exec(
				`delete from workout_sessions
         where id = $1 and completed_at is null`,
				item.SessionID,
			)
		}
		result, _ := s.DB.Exec(
			`update training_plan_workouts
       set status = 'skipped',
           skip_reason = $1,
           session_id = null
       where id = $2 and status in ('pending', 'in_progress')`,
			reason,
			item.PlanWorkoutID,
		)
		if affectedRows(result) == 0 {
			continue
		}
		skippedByPlan[item.PlanID]++
		totalSkipped++
	}

	for planID, skipped := range skippedByPlan {
		if skipped <= 0 {
			continue
		}
		after := s.planSnapshot(planID)
		changeReason := "Автопропуск: " + strconv.Itoa(skipped) + " трен. не выполнены в назначенный день"
		s.logPlanChange(userID, planID, "overdue_miss", changeReason, beforeSnapshots[planID], after)
	}

	return totalSkipped
}

func (s *Site) fetchWorkouts(level string, categories []string, restrictions []string, equipment []string) []workoutCard {
	rows, err := s.DB.Query(
		`select w.id, w.name, w.description, w.duration_minutes, w.difficulty, coalesce(w.category, ''), count(we.exercise_id)
     from workouts w
     left join workout_exercises we on we.workout_id = w.id
     group by w.id
     order by w.created_at`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	allowedLevels := difficultyAllowed(level)
	output := []workoutCard{}
	for rows.Next() {
		var card workoutCard
		_ = rows.Scan(&card.ID, &card.Name, &card.Description, &card.Duration, &card.Difficulty, &card.Category, &card.Exercises)
		if len(allowedLevels) > 0 && !allowedLevels[card.Difficulty] {
			continue
		}
		if len(categories) > 0 && card.Category != "" {
			matched := false
			for _, c := range categories {
				if strings.EqualFold(card.Category, c) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if !s.workoutAllowed(card.ID, restrictions, equipment) {
			continue
		}
		output = append(output, card)
	}
	return output
}

func (s *Site) workoutAllowed(workoutID string, restrictions []string, equipment []string) bool {
	restrictedCategories := restrictionCategories(restrictions)

	rows, err := s.DB.Query(
		`select coalesce(category, ''), coalesce(array_to_string(equipment, ','), '') from exercises e
     join workout_exercises we on we.exercise_id = e.id
     where we.workout_id = $1`,
		workoutID,
	)
	if err != nil {
		return true
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var equipRaw string
		_ = rows.Scan(&category, &equipRaw)
		equip := parseCSV(equipRaw)
		if category != "" {
			if restrictedCategories[category] {
				return false
			}
		}
		if len(equip) > 0 {
			if !isSubset(equip, equipment) {
				return false
			}
		}
	}

	return true
}

func (s *Site) applyAdaptation(userID, planID, trigger string) {
	plan, err := s.getActivePlan(userID)
	if err != nil || plan == nil || plan.ID != planID {
		return
	}

	var skipped int
	_ = s.DB.QueryRow(
		`select count(*) from training_plan_workouts
     where plan_id = $1 and status = 'skipped' and scheduled_date >= current_date - interval '7 days'`,
		planID,
	).Scan(&skipped)
	if trigger == "skip" {
		if skipped < 2 {
			return
		}
		before := s.planSnapshot(planID)
		result, _ := s.DB.Exec(
			`update training_plan_workouts
       set scheduled_date = scheduled_date + interval '7 days'
       where plan_id = $1 and status = 'pending' and scheduled_date is not null`,
			planID,
		)
		if affectedRows(result) == 0 {
			return
		}
		after := s.planSnapshot(planID)
		s.logPlanChange(userID, planID, "missed", "Есть пропуски: план перераспределён", before, after)
		return
	}
	if trigger != "feedback" {
		return
	}

	type feedbackSnapshot struct {
		Pain      int
		Tolerance int
		Exertion  int
		Wellbeing int
	}

	samples := 0
	sumPain := 0
	sumTolerance := 0
	sumExertion := 0
	sumWellbeing := 0
	last := feedbackSnapshot{}

	rows, err := s.DB.Query(
		`select coalesce(pain_level, 0), coalesce(tolerance, 0), coalesce(perceived_exertion, 0), coalesce(wellbeing, 0)
     from workout_session_feedback f
     join workout_sessions ws on ws.id = f.session_id
     join training_plan_workouts tpw on tpw.id = ws.plan_workout_id
     where f.user_id = $1 and tpw.plan_id = $2
     order by ws.completed_at desc nulls last, f.created_at desc
     limit 4`,
		userID,
		planID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pain, tolerance, exertion, wellbeing int
			_ = rows.Scan(&pain, &tolerance, &exertion, &wellbeing)
			if samples == 0 {
				last = feedbackSnapshot{Pain: pain, Tolerance: tolerance, Exertion: exertion, Wellbeing: wellbeing}
			}
			samples++
			sumPain += pain
			sumTolerance += tolerance
			sumExertion += exertion
			sumWellbeing += wellbeing
		}
	}

	if samples == 0 {
		return
	}

	avgPain := float64(sumPain) / float64(samples)
	avgTolerance := float64(sumTolerance) / float64(samples)
	avgExertion := float64(sumExertion) / float64(samples)
	avgWellbeing := float64(sumWellbeing) / float64(samples)

	reasonCode := ""
	reason := ""
	updateQuery := ""
	lastExcellent := last.Tolerance >= 4 && last.Wellbeing >= 4 && last.Pain <= 1 && last.Exertion <= 3
	lastCritical := last.Pain >= 4 || last.Tolerance <= 2 || last.Wellbeing <= 2 || last.Exertion >= 5
	avgWarning := avgPain >= 3.5 || (avgPain >= 3 && avgWellbeing <= 2.5)
	avgRegression := avgTolerance <= 2.5 || avgWellbeing <= 2.5 || (avgExertion >= 4 && avgTolerance <= 3.2)

	if lastExcellent && skipped == 0 && avgPain <= 3 && avgExertion <= 4.5 {
		reasonCode = "progression"
		reason = "Хорошая переносимость: повышена интенсивность"
		updateQuery = `update training_plan_workouts
                   set intensity = intensity + 1
                   where plan_id = $1 and status = 'pending' and intensity < 3`
	} else if last.Pain >= 4 || (!lastExcellent && avgWarning) {
		reasonCode = "warning"
		reason = "Отмечен дискомфорт: снижена интенсивность и рекомендована консультация"
		updateQuery = `update training_plan_workouts
                   set intensity = intensity - 1
                   where plan_id = $1 and status = 'pending' and intensity > 1`
	} else if lastCritical || (!lastExcellent && avgRegression) {
		reasonCode = "regression"
		reason = "Низкая переносимость нагрузки: снижена интенсивность"
		updateQuery = `update training_plan_workouts
                   set intensity = intensity - 1
                   where plan_id = $1 and status = 'pending' and intensity > 1`
	}

	if reasonCode == "" {
		return
	}
	before := s.planSnapshot(planID)
	var pendingCount int
	var minIntensity int
	var maxIntensity int
	_ = s.DB.QueryRow(
		`select count(*), coalesce(min(intensity), 0), coalesce(max(intensity), 0)
     from training_plan_workouts
     where plan_id = $1 and status = 'pending'`,
		planID,
	).Scan(&pendingCount, &minIntensity, &maxIntensity)

	result, _ := s.DB.Exec(updateQuery, planID)
	if affectedRows(result) == 0 {
		noEffectReason := reason
		if pendingCount == 0 {
			noEffectReason = "Оценка сохранена: в плане нет будущих тренировок для корректировки"
		} else if reasonCode == "progression" && maxIntensity >= 3 {
			noEffectReason = "Хорошая переносимость: интенсивность уже максимальная"
		} else if (reasonCode == "regression" || reasonCode == "warning") && minIntensity <= 1 {
			if reasonCode == "warning" {
				noEffectReason = "Отмечен дискомфорт: интенсивность уже минимальная, рекомендована консультация"
			} else {
				noEffectReason = "Низкая переносимость нагрузки: интенсивность уже минимальная"
			}
		} else {
			noEffectReason = reason + " (изменения не требуются)"
		}
		s.logPlanChange(userID, planID, reasonCode, noEffectReason, before, before)
		return
	}
	after := s.planSnapshot(planID)
	s.logPlanChange(userID, planID, reasonCode, reason, before, after)
}

func affectedRows(result sql.Result) int64 {
	if result == nil {
		return 0
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0
	}
	return rows
}

func (s *Site) planSnapshot(planID string) json.RawMessage {
	rows, err := s.DB.Query(
		`select workout_id, week, day, intensity, status, coalesce(to_char(scheduled_time, 'HH24:MI'), '')
     from training_plan_workouts
     where plan_id = $1
     order by week, day`,
		planID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	type snapshotItem struct {
		WorkoutID     string `json:"workout_id"`
		Week          int    `json:"week"`
		Day           int    `json:"day"`
		Intensity     int    `json:"intensity"`
		Status        string `json:"status"`
		ScheduledTime string `json:"scheduled_time,omitempty"`
	}

	items := []snapshotItem{}
	for rows.Next() {
		var item snapshotItem
		_ = rows.Scan(&item.WorkoutID, &item.Week, &item.Day, &item.Intensity, &item.Status, &item.ScheduledTime)
		items = append(items, item)
	}

	payload, _ := json.Marshal(items)
	return payload
}

func (s *Site) logPlanChange(userID, planID, code, reason string, before, after json.RawMessage) {
	_, _ = s.DB.Exec(
		`insert into training_plan_changes (plan_id, user_id, reason_code, reason, before_plan, after_plan)
     values ($1, $2, $3, $4, $5, $6)`,
		planID,
		userID,
		code,
		reason,
		before,
		after,
	)
}

func resolveLevel(fitness string) string {
	level := strings.TrimSpace(strings.ToLower(fitness))
	switch level {
	case "низкий", "легкая", "легкий":
		level = "Легкая"
	case "средний", "средняя":
		level = "Средняя"
	case "высокий", "продвинутая":
		level = "Продвинутая"
	default:
		level = "Легкая"
	}

	return level
}

func sessionMinutesForLevel(level string) int {
	switch level {
	case "Средняя":
		return 30
	case "Продвинутая":
		return 40
	default:
		return 20
	}
}

func categoriesForGoal(goal string) []string {
	switch strings.ToLower(goal) {
	case "восстановление":
		return []string{"Реабилитация"}
	case "подвижность", "мобилизация":
		return []string{"Мобилизация", "Растяжка"}
	case "сила":
		return []string{"Кор", "Спина", "Ноги"}
	case "выносливость":
		return []string{"Кардио"}
	default:
		return []string{}
	}
}

func categoriesFromPreferences(preferences string) []string {
	text := strings.ToLower(preferences)
	mapping := map[string]string{
		"растяж": "Растяжка",
		"мобил":  "Мобилизация",
		"кардио": "Кардио",
		"спина":  "Спина",
		"кор":    "Кор",
		"ног":    "Ноги",
		"плеч":   "Плечи",
		"баланс": "Кор",
		"осанк":  "Спина",
		"стабил": "Кор",
	}
	out := []string{}
	seen := map[string]bool{}
	for _, item := range parseCSV(preferences) {
		cleaned := strings.TrimSpace(item)
		if cleaned == "" {
			continue
		}
		key := strings.ToLower(cleaned)
		if category, ok := mapping[key]; ok {
			if !seen[category] {
				seen[category] = true
				out = append(out, category)
			}
			continue
		}
		for matchKey, category := range mapping {
			if strings.Contains(key, matchKey) && !seen[category] {
				seen[category] = true
				out = append(out, category)
			}
		}
	}
	for key, category := range mapping {
		if strings.Contains(text, key) && !seen[category] {
			seen[category] = true
			out = append(out, category)
		}
	}
	return out
}

func prefersNoEquipment(preferences string) bool {
	text := strings.ToLower(preferences)
	if strings.Contains(text, "без инвентар") || strings.Contains(text, "без оборудования") || strings.Contains(text, "без снар") {
		return true
	}
	return false
}

func mergeCategories(sets ...[]string) []string {
	seen := map[string]bool{}
	merged := []string{}
	for _, list := range sets {
		for _, item := range list {
			if item == "" {
				continue
			}
			key := strings.ToLower(item)
			if seen[key] {
				continue
			}
			seen[key] = true
			merged = append(merged, item)
		}
	}
	return merged
}

func questionnaireChanged(prev, next questionnaireData, prevRestrictions, nextRestrictions []string) bool {
	if prev.Goal != next.Goal || prev.FitnessLevel != next.FitnessLevel || prev.DaysPerWeek != next.DaysPerWeek {
		return true
	}
	if prev.Preferences != next.Preferences {
		return true
	}
	if !sameStringSet(prev.Equipment, next.Equipment) {
		return true
	}
	if !sameStringSet(prevRestrictions, nextRestrictions) {
		return true
	}
	return false
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, item := range a {
		seen[strings.ToLower(strings.TrimSpace(item))]++
	}
	for _, item := range b {
		key := strings.ToLower(strings.TrimSpace(item))
		if seen[key] == 0 {
			return false
		}
		seen[key]--
	}
	for _, v := range seen {
		if v != 0 {
			return false
		}
	}
	return true
}

func filterWorkoutsByDuration(list []workoutCard, target, tolerance int) []workoutCard {
	if target <= 0 || tolerance < 0 {
		return list
	}
	filtered := make([]workoutCard, 0, len(list))
	for _, item := range list {
		if item.Duration == 0 {
			continue
		}
		delta := item.Duration - target
		if delta < 0 {
			delta = -delta
		}
		if delta <= tolerance {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func shuffleWorkouts(list []workoutCard) []workoutCard {
	if len(list) == 0 {
		return list
	}
	out := make([]workoutCard, len(list))
	copy(out, list)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func rotateWorkouts(list []workoutCard, offset int) []workoutCard {
	if len(list) == 0 {
		return list
	}
	shift := offset % len(list)
	if shift == 0 {
		return list
	}
	out := append([]workoutCard{}, list[shift:]...)
	out = append(out, list[:shift]...)
	return out
}

func difficultyAllowed(level string) map[string]bool {
	allowed := map[string]bool{}
	switch level {
	case "Средняя":
		allowed["Легкая"] = true
		allowed["Средняя"] = true
	case "Продвинутая":
		allowed["Легкая"] = true
		allowed["Средняя"] = true
		allowed["Сложная"] = true
	default:
		allowed["Легкая"] = true
	}
	return allowed
}

func restrictionOptions() []string {
	return []string{"Колени", "Спина", "Плечи", "Сердце", "Растяжка"}
}

func equipmentOptions() []string {
	return []string{
		"Без инвентаря",
		"Коврик",
		"Резинка",
		"Гантели",
		"Стул",
		"Фитбол",
	}
}

func preferenceOptions() []string {
	return []string{
		"Растяжка",
		"Мобилизация",
		"Кардио",
		"Кор",
		"Спина",
		"Ноги",
		"Плечи",
	}
}

func normalizeEquipmentSelection(list []string) []string {
	cleaned := []string{}
	seen := map[string]bool{}
	for _, item := range list {
		canonical := canonicalEquipmentName(item)
		if canonical == "" {
			continue
		}
		lower := strings.ToLower(canonical)
		if seen[lower] {
			continue
		}
		seen[lower] = true
		cleaned = append(cleaned, canonical)
	}
	if containsEquipmentValue(cleaned, "Без инвентаря") {
		return []string{"Без инвентаря"}
	}
	return cleaned
}

func containsEquipmentValue(list []string, target string) bool {
	targetCanonical := strings.ToLower(canonicalEquipmentName(target))
	if targetCanonical == "" {
		return false
	}
	for _, item := range list {
		value := strings.ToLower(canonicalEquipmentName(item))
		if value == "" {
			continue
		}
		if value == targetCanonical {
			return true
		}
	}
	return false
}

func mergeEquipmentOptions(base []string, extra []string) []string {
	merged := append([]string{}, base...)
	for _, item := range extra {
		merged = append(merged, item)
	}
	return normalizeEquipmentSelection(merged)
}

func mergeWorkoutCards(primary, secondary []workoutCard) []workoutCard {
	if len(secondary) == 0 {
		return primary
	}
	merged := append([]workoutCard{}, primary...)
	seen := map[string]bool{}
	for _, item := range merged {
		seen[item.ID] = true
	}
	for _, item := range secondary {
		if seen[item.ID] {
			continue
		}
		seen[item.ID] = true
		merged = append(merged, item)
	}
	return merged
}

func normalizeSelection(list []string, options []string) []string {
	allowed := map[string]string{}
	for _, item := range options {
		allowed[strings.ToLower(strings.TrimSpace(item))] = item
	}

	cleaned := []string{}
	seen := map[string]bool{}
	for _, item := range list {
		key := strings.ToLower(strings.TrimSpace(item))
		if key == "" {
			continue
		}
		value, ok := allowed[key]
		if !ok {
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func normalizePreferenceSelection(list []string) []string {
	return normalizeSelection(list, preferenceOptions())
}

func normalizeRestrictionSelection(list []string) []string {
	return normalizeSelection(list, restrictionOptions())
}

func isAdminPauseReason(reason string) bool {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(normalized, "администратор")
}

func isSickLeavePauseReason(reason string) bool {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(normalized, "больнич")
}

func adminPauseLockedForRole(role, status, pausedReason string) bool {
	if status != "paused" {
		return false
	}
	if isSickLeavePauseReason(pausedReason) {
		return true
	}
	if !isAdminPauseReason(pausedReason) {
		return false
	}
	return !isAdminRole(role)
}

func planLaunchBlockedForRole(role, status, pausedReason string) bool {
	if status != "paused" {
		return false
	}
	if isSickLeavePauseReason(pausedReason) {
		return true
	}
	return !(isAdminRole(role) && isAdminPauseReason(pausedReason))
}

func isAdminRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "admin")
}

func weeklyOffsets(frequency int) []int {
	switch frequency {
	case 1:
		return []int{2}
	case 2:
		return []int{1, 4}
	case 3:
		return []int{0, 2, 4}
	case 4:
		return []int{0, 2, 4, 6}
	case 5:
		return []int{0, 1, 3, 5, 6}
	case 6:
		return []int{0, 1, 2, 4, 5, 6}
	case 7:
		return []int{0, 1, 2, 3, 4, 5, 6}
	default:
		if frequency <= 0 {
			return []int{}
		}
		offsets := []int{}
		step := float64(7) / float64(frequency)
		used := map[int]bool{}
		for i := 0; i < frequency; i++ {
			raw := int(math.Round(float64(i) * step))
			if raw < 0 {
				raw = 0
			}
			if raw > 6 {
				raw = 6
			}
			for used[raw] && raw < 6 {
				raw++
			}
			used[raw] = true
			offsets = append(offsets, raw)
		}
		sort.Ints(offsets)
		return offsets
	}
}

func defaultPlanWorkoutTime(weekdayOffset int) string {
	if weekdayOffset >= 5 {
		return "10:00"
	}
	switch weekdayOffset {
	case 0, 1:
		return "08:30"
	case 2, 3:
		return "18:30"
	default:
		return "09:00"
	}
}

func restrictionCategories(restrictions []string) map[string]bool {
	out := map[string]bool{}
	mapping := map[string][]string{
		"Колени":   {"Ноги"},
		"Спина":    {"Спина"},
		"Плечи":    {"Плечи"},
		"Сердце":   {"Кардио"},
		"Растяжка": {"Растяжка"},
	}
	for _, r := range restrictions {
		if cats, ok := mapping[r]; ok {
			for _, c := range cats {
				out[c] = true
			}
		}
	}
	return out
}

func (s *Site) loadRestrictions(userID string) []string {
	var raw string
	_ = s.DB.QueryRow(
		`select coalesce(array_to_string(restrictions, ','), '')
     from user_profiles
     where user_id = $1`,
		userID,
	).Scan(&raw)
	return parseCSV(raw)
}

func (s *Site) loadDoctorApproval(userID string) bool {
	var approval bool
	_ = s.DB.QueryRow(`select doctor_approval from user_profiles where user_id = $1`, userID).Scan(&approval)
	return approval
}

func isSubset(need, have []string) bool {
	noEquipmentOnly := containsEquipmentValue(have, "Без инвентаря")
	allowed := map[string]bool{}
	for _, item := range have {
		canonical := canonicalEquipmentName(item)
		if canonical == "" {
			continue
		}
		if canonical == "Без инвентаря" {
			continue
		}
		allowed[strings.ToLower(canonical)] = true
	}
	for _, item := range need {
		canonical := canonicalEquipmentName(item)
		if canonical == "" {
			// Окружение вроде стены/пола и неизвестные метки не блокируют подбор.
			continue
		}
		if canonical == "Коврик" {
			// Коврик считаем базово доступным.
			continue
		}
		if noEquipmentOnly {
			return false
		}
		if !allowed[strings.ToLower(canonical)] {
			return false
		}
	}
	return true
}

func canonicalEquipmentName(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return ""
	}

	switch {
	case strings.Contains(v, "без инвентар"), strings.Contains(v, "без оборуд"), strings.Contains(v, "без снар"):
		return "Без инвентаря"
	case strings.Contains(v, "коврик"):
		return "Коврик"
	case strings.Contains(v, "резин"), strings.Contains(v, "эспандер"):
		return "Резинка"
	case strings.Contains(v, "гантел"):
		return "Гантели"
	case strings.Contains(v, "стул"):
		return "Стул"
	case strings.Contains(v, "фитбол"), strings.Contains(v, "мяч"):
		return "Фитбол"
	case strings.Contains(v, "стена"), strings.Contains(v, "пол"), strings.Contains(v, "пространство"):
		// Базовые условия считаем доступными по умолчанию.
		return ""
	default:
		// Неизвестный инвентарь не должен ломать подбор.
		return ""
	}
}

func nextWeekStart(now time.Time) time.Time {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	daysUntilMonday := 8 - weekday
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, daysUntilMonday)
	return start
}

func (s *Site) planOwnedByUser(planID, userID string) bool {
	var owner string
	err := s.DB.QueryRow(`select user_id from training_plans where id = $1`, planID).Scan(&owner)
	return err == nil && owner == userID
}

func (s *Site) planIDBySession(sessionID string) string {
	var planID string
	_ = s.DB.QueryRow(
		`select tpw.plan_id
     from workout_sessions ws
     join training_plan_workouts tpw on tpw.id = ws.plan_workout_id
     where ws.id = $1`,
		sessionID,
	).Scan(&planID)
	return planID
}

func (s *Site) attachSessionToPlan(userID, sessionID string) string {
	var workoutID string
	if err := s.DB.QueryRow(
		`select workout_id
     from workout_sessions
     where id = $1 and user_id = $2`,
		sessionID,
		userID,
	).Scan(&workoutID); err != nil {
		return ""
	}

	var planWorkoutID string
	var planID string
	var linkedSessionID string
	scheduleDate, scheduleTime := scheduleDateAndTime(time.Now())
	err := s.DB.QueryRow(
		`select pw.id, pw.plan_id, coalesce(pw.session_id::text, '')
     from training_plan_workouts pw
     join training_plans tp on tp.id = pw.plan_id
     where tp.user_id = $1
       and tp.status in ('active', 'paused')
       and pw.workout_id = $2
       and pw.status in ('pending', 'in_progress')
       and (
         pw.scheduled_date is null
         or (
           pw.scheduled_date = $3::date
           and (pw.scheduled_time is null or pw.scheduled_time <= $4::time)
         )
       )
     order by case when pw.status = 'in_progress' then 0 else 1 end,
              pw.scheduled_date nulls last, coalesce(pw.scheduled_time, time '23:59'), pw.week, pw.day
     limit 1`,
		userID,
		workoutID,
		scheduleDate,
		scheduleTime,
	).Scan(&planWorkoutID, &planID, &linkedSessionID)
	if err != nil {
		return ""
	}

	if linkedSessionID != "" && linkedSessionID != sessionID {
		return ""
	}

	_, _ = s.DB.Exec(
		`update workout_sessions
     set plan_workout_id = $1
     where id = $2`,
		planWorkoutID,
		sessionID,
	)
	_, _ = s.DB.Exec(
		`update training_plan_workouts
     set session_id = $1,
         status = 'completed'
     where id = $2`,
		sessionID,
		planWorkoutID,
	)
	return planID
}

func (s *Site) updateAchievements(userID string) {
	rows, err := s.DB.Query(`select id, title, points_reward, coalesce(metric, ''), coalesce(target, 0) from achievements`)
	if err != nil {
		return
	}
	defer rows.Close()

	type ach struct {
		ID           string
		Title        string
		PointsReward int
		Metric       string
		Target       int
	}
	list := []ach{}
	for rows.Next() {
		var a ach
		_ = rows.Scan(&a.ID, &a.Title, &a.PointsReward, &a.Metric, &a.Target)
		list = append(list, a)
	}

	existing := map[string]bool{}
	exRows, err := s.DB.Query(`select achievement_id, unlocked from user_achievements where user_id = $1`, userID)
	if err == nil {
		defer exRows.Close()
		for exRows.Next() {
			var id string
			var unlocked bool
			_ = exRows.Scan(&id, &unlocked)
			existing[id] = unlocked
		}
	}

	var total int
	_ = s.DB.QueryRow(`select count(*) from workout_sessions where user_id = $1 and completed_at is not null`, userID).Scan(&total)
	var last7 int
	_ = s.DB.QueryRow(`select count(*) from workout_sessions where user_id = $1 and completed_at >= now() - interval '7 days'`, userID).Scan(&last7)
	var last30 int
	_ = s.DB.QueryRow(`select count(*) from workout_sessions where user_id = $1 and completed_at >= now() - interval '30 days'`, userID).Scan(&last30)
	var minutesTotal int
	_ = s.DB.QueryRow(
		`select coalesce(sum(duration_minutes), 0)
     from workout_sessions
     where user_id = $1 and completed_at is not null`,
		userID,
	).Scan(&minutesTotal)
	var minutesMonth int
	_ = s.DB.QueryRow(
		`select coalesce(sum(duration_minutes), 0)
     from workout_sessions
     where user_id = $1 and completed_at >= now() - interval '30 days'`,
		userID,
	).Scan(&minutesMonth)

	streak := s.computeStreak(userID)

	for _, a := range list {
		progress := 0
		target := a.Target
		metric := strings.ToLower(strings.TrimSpace(a.Metric))
		if target <= 0 {
			target = 1
		}
		switch metric {
		case "total":
			progress = total
		case "week":
			progress = last7
		case "streak":
			progress = streak
		case "month":
			progress = last30
		case "minutes_total":
			progress = minutesTotal
		case "minutes_month":
			progress = minutesMonth
		default:
			switch a.Title {
			case "Первый шаг":
				progress = total
				target = 1
			case "Первые три":
				progress = total
				target = 3
			case "Пять тренировок":
				progress = total
				target = 5
			case "Серия":
				progress = streak
				target = 5
			case "Неделя подряд":
				progress = streak
				target = 7
			case "Железная воля":
				progress = streak
				target = 10
			case "Активные 2 недели":
				progress = streak
				target = 14
			case "Настойчивость":
				progress = last30
				target = 10
			case "12 тренировок за месяц":
				progress = last30
				target = 12
			case "Регулярность":
				progress = last30
				target = 8
			case "Месяц активности":
				progress = last30
				target = 20
			case "15 тренировок":
				progress = total
				target = 15
			case "Марафон":
				progress = total
				target = 25
			case "Активная неделя I":
				progress = last7
				target = 3
			case "Активная неделя II":
				progress = last7
				target = 5
			case "Минуты восстановления I":
				progress = minutesTotal
				target = 150
			case "Минуты восстановления II":
				progress = minutesTotal
				target = 600
			default:
				progress = total
				target = 1
			}
		}

		unlocked := progress >= target
		if progress > target {
			progress = target
		}

		_, _ = s.DB.Exec(
			`insert into user_achievements (user_id, achievement_id, unlocked, unlocked_at, progress, total)
       values ($1, $2, $3, case when $3 then now() else null end, $4, $5)
       on conflict (user_id, achievement_id)
       do update set unlocked = excluded.unlocked,
                     unlocked_at = case
                                     when user_achievements.unlocked then user_achievements.unlocked_at
                                     when excluded.unlocked then now()
                                     else user_achievements.unlocked_at
                                   end,
                     progress = excluded.progress,
                     total = excluded.total`,
			userID,
			a.ID,
			unlocked,
			progress,
			target,
		)

		if unlocked && !existing[a.ID] && a.PointsReward > 0 {
			_, _ = s.DB.Exec(
				`insert into user_points (user_id, points_balance, points_total)
         values ($1, $2, $2)
         on conflict (user_id)
         do update set points_balance = user_points.points_balance + $2,
                       points_total = user_points.points_total + $2`,
				userID,
				a.PointsReward,
			)
			_, _ = s.DB.Exec(
				`insert into user_point_events (user_id, source, source_id, points, reason)
         values ($1, 'achievement', $2::uuid, $3, $4)`,
				userID,
				a.ID,
				a.PointsReward,
				"Достижение «"+a.Title+"»",
			)
		}
	}
}

func (s *Site) computeStreak(userID string) int {
	rows, err := s.DB.Query(
		`select completed_at from workout_sessions
     where user_id = $1 and completed_at is not null
     order by completed_at desc`,
		userID,
	)
	if err != nil {
		return 0
	}
	defer rows.Close()

	streak := 0
	var last time.Time
	for rows.Next() {
		var completed time.Time
		_ = rows.Scan(&completed)
		if streak == 0 {
			streak = 1
			last = completed
			continue
		}
		if completed.After(last.AddDate(0, 0, -2)) {
			streak++
			last = completed
			continue
		}
		break
	}
	return streak
}

func (s *Site) createWorkoutSession(userID, workoutID, planWorkoutID string) (string, error) {
	_ = s.ensureWorkoutExercises(workoutID)
	var exercisesCount int
	_ = s.DB.QueryRow("select count(*) from workout_exercises where workout_id = $1", workoutID).Scan(&exercisesCount)

	var sessionID string
	err := s.DB.QueryRow(
		`insert into workout_sessions (user_id, workout_id, total_exercises, completed_exercises, plan_workout_id)
     values ($1, $2, $3, 0, $4)
     returning id`,
		userID,
		workoutID,
		exercisesCount,
		nullIfEmpty(planWorkoutID),
	).Scan(&sessionID)
	if err != nil {
		return "", err
	}

	rows, err := s.DB.Query(
		`select exercise_id, sort_order from workout_exercises where workout_id = $1 order by sort_order`,
		workoutID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var exerciseID string
			var order int
			_ = rows.Scan(&exerciseID, &order)
			_, _ = s.DB.Exec(
				`insert into workout_session_exercises (session_id, exercise_id, sort_order)
         values ($1, $2, $3)`,
				sessionID,
				exerciseID,
				order,
			)
		}
	}

	if planWorkoutID != "" {
		_, _ = s.DB.Exec(`update training_plan_workouts set session_id = $1 where id = $2`, sessionID, planWorkoutID)
	}

	return sessionID, nil
}
