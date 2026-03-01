package db

import "database/sql"

// EnsureUserDefaults creates minimal related records required by the app flow.
func EnsureUserDefaults(database *sql.DB, userID string) error {
	_, _ = database.Exec("insert into user_profiles (user_id) values ($1) on conflict do nothing", userID)
	_, _ = database.Exec("insert into medical_info (user_id) values ($1) on conflict do nothing", userID)
	_, _ = database.Exec("insert into user_points (user_id) values ($1) on conflict do nothing", userID)
	return nil
}
