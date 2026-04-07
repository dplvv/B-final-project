package site

import (
	"database/sql"
	"errors"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"
)

func normalizeCorporateEmail(value string) (string, error) {
	email := strings.TrimSpace(strings.ToLower(value))
	if email == "" {
		return "", nil
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return "", err
	}
	normalized := strings.TrimSpace(strings.ToLower(parsed.Address))
	if normalized == "" || strings.Count(normalized, "@") != 1 {
		return "", errors.New("invalid email")
	}
	return normalized, nil
}

func (s *Site) loadTodayPlanProgress(userID string) profilePlanProgressView {
	view := profilePlanProgressView{}
	if strings.TrimSpace(userID) == "" {
		return view
	}
	scheduleDate, scheduleTime := scheduleDateAndTime(time.Now())

	var scheduledDate sql.NullTime
	err := s.DB.QueryRow(
		`select pw.id,
            w.name,
            pw.status,
            coalesce(pw.session_id::text, ''),
            pw.scheduled_date,
            coalesce(to_char(pw.scheduled_time, 'HH24:MI'), ''),
            case
              when pw.scheduled_date is null then true
              when pw.scheduled_date <> $2::date then false
              when pw.scheduled_time is null then true
              else pw.scheduled_time <= $3::time
            end as start_allowed
     from training_plan_workouts pw
     join training_plans tp on tp.id = pw.plan_id
     join workouts w on w.id = pw.workout_id
     where tp.user_id = $1
       and tp.status in ('active', 'paused')
       and pw.scheduled_date = $2::date
     order by case pw.status
                when 'in_progress' then 0
                when 'pending' then 1
                when 'completed' then 2
                when 'skipped' then 3
                else 4
              end,
              coalesce(pw.scheduled_time, time '23:59'),
              pw.week,
              pw.day
     limit 1`,
		userID,
		scheduleDate,
		scheduleTime,
	).Scan(
		&view.PlanWorkoutID,
		&view.WorkoutName,
		&view.Status,
		&view.SessionID,
		&scheduledDate,
		&view.ScheduledTime,
		&view.StartAllowed,
	)
	if err != nil {
		return view
	}

	view.HasWorkout = true
	view.StatusLabel = statusLabelRU(view.Status)
	if scheduledDate.Valid {
		view.ScheduledDate = scheduledDate.Time.Format("02.01.2006")
	}
	if !view.StartAllowed && view.ScheduledTime != "" {
		view.AvailableFrom = "Доступно с " + view.ScheduledTime
	}

	switch view.Status {
	case "completed":
		view.Progress = 100
	case "skipped":
		view.Progress = 0
	default:
		view.Progress = 0
	}

	if view.SessionID != "" {
		var completedExercises int
		var totalExercises int
		var completedAt sql.NullTime
		_ = s.DB.QueryRow(
			`select coalesce(completed_exercises, 0),
              coalesce(total_exercises, 0),
              completed_at
       from workout_sessions
       where id = $1`,
			view.SessionID,
		).Scan(&completedExercises, &totalExercises, &completedAt)

		if completedAt.Valid {
			view.Progress = 100
			if view.Status == "pending" || view.Status == "in_progress" {
				view.Status = "completed"
			}
		} else if totalExercises > 0 {
			percent := int(float64(completedExercises) / float64(totalExercises) * 100)
			if percent < 0 {
				percent = 0
			}
			if percent > 99 {
				percent = 99
			}
			if view.Status == "in_progress" && percent == 0 {
				percent = 10
			}
			view.Progress = percent
		}
		view.StatusLabel = statusLabelRU(view.Status)
	}

	return view
}

func (s *Site) loadLeaderboardPlace(userID string) (int, int) {
	place, total := s.loadLeaderboardPlaceByRole(userID, "employee")
	if place > 0 {
		return place, total
	}
	return s.loadLeaderboardPlaceByRole(userID, "")
}

func (s *Site) loadLeaderboardPlaceByRole(userID, role string) (int, int) {
	var place int
	var total int
	query := `with ranked as (
              select u.id,
                     rank() over (
                       order by coalesce(p.points_total, 0) desc,
                                coalesce(count(ws.id), 0) desc,
                                coalesce(sum(ws.duration_minutes), 0) desc,
                                u.name
                     ) as place,
                     count(*) over() as total
              from users u
              left join user_points p on p.user_id = u.id
              left join workout_sessions ws on ws.user_id = u.id and ws.completed_at is not null
              where ($2 = '' or u.role = $2)
              group by u.id, u.name, p.points_total
            )
            select place, total
            from ranked
            where id = $1`
	err := s.DB.QueryRow(query, userID, role).Scan(&place, &total)
	if err != nil {
		return 0, 0
	}
	return place, total
}

func (s *Site) loadReminderSettings(userID string) reminderSettingsView {
	settings := reminderSettingsView{
		Enabled:      true,
		ReminderTime: "09:00",
		Weekdays:     []string{"1", "2", "3", "4", "5"},
	}
	if strings.TrimSpace(userID) == "" {
		return settings
	}

	var weekdaysRaw string
	err := s.DB.QueryRow(
		`select enabled,
            coalesce(to_char(reminder_time, 'HH24:MI'), '09:00'),
            coalesce(array_to_string(weekdays, ','), '1,2,3,4,5')
     from user_reminder_settings
     where user_id = $1`,
		userID,
	).Scan(&settings.Enabled, &settings.ReminderTime, &weekdaysRaw)
	if err != nil {
		return settings
	}

	parsed := parseCSV(weekdaysRaw)
	if len(parsed) > 0 {
		settings.Weekdays = parsed
	}
	return settings
}

func reminderWeekdayOptions() []reminderWeekdayOption {
	return []reminderWeekdayOption{
		{Value: "1", Label: "Пн"},
		{Value: "2", Label: "Вт"},
		{Value: "3", Label: "Ср"},
		{Value: "4", Label: "Чт"},
		{Value: "5", Label: "Пт"},
		{Value: "6", Label: "Сб"},
		{Value: "7", Label: "Вс"},
	}
}

func normalizeReminderWeekdays(values []string) []int {
	if len(values) == 0 {
		return nil
	}
	unique := map[int]bool{}
	output := []int{}
	for _, raw := range values {
		day, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			continue
		}
		if day < 1 || day > 7 {
			continue
		}
		if unique[day] {
			continue
		}
		unique[day] = true
		output = append(output, day)
	}
	sort.Ints(output)
	return output
}

func intsToCSV(values []int) string {
	if len(values) == 0 {
		return ""
	}
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, strconv.Itoa(value))
	}
	return strings.Join(items, ",")
}

func statusLabelRU(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "active":
		return "Активно"
	case "paused":
		return "На паузе"
	case "archived":
		return "Архив"
	case "pending":
		return "Не начата"
	case "in_progress":
		return "В процессе"
	case "completed":
		return "Завершена"
	case "skipped":
		return "Пропущена"
	case "open":
		return "Открыто"
	case "closed":
		return "Закрыто"
	case "resolved":
		return "Решено"
	case "approved":
		return "Одобрено"
	case "rejected":
		return "Отклонено"
	default:
		return value
	}
}
