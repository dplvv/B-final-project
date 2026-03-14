#!/usr/bin/env bash
set -euo pipefail

SRC_DB="${1:-rehab_app}"
ER_DB="${2:-rehab_er}"
DB_USER="${DB_USER:-rehab}"

TABLES=(
  public.users
  public.user_profiles
  public.user_points
  public.achievements
  public.user_achievements
  public.rewards
  public.reward_redemptions
  public.exercises
  public.workouts
  public.workout_exercises
  public.programs
  public.program_workouts
  public.training_plans
  public.training_plan_workouts
  public.training_plan_changes
  public.workout_sessions
  public.workout_session_exercises
  public.workout_session_sets
  public.workout_session_feedback
  public.support_tickets
  public.support_ticket_messages
)

dump_args=(--schema-only --no-owner --no-privileges)
for table in "${TABLES[@]}"; do
  dump_args+=(-t "$table")
done

docker compose exec -T db psql -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "drop database if exists $ER_DB"
docker compose exec -T db psql -U "$DB_USER" -d postgres -v ON_ERROR_STOP=1 -c "create database $ER_DB"
docker compose exec -T db psql -U "$DB_USER" -d "$ER_DB" -v ON_ERROR_STOP=1 -c 'create extension if not exists "uuid-ossp"'

docker compose exec -T db pg_dump -U "$DB_USER" -d "$SRC_DB" "${dump_args[@]}" \
  | docker compose exec -T db psql -U "$DB_USER" -d "$ER_DB" -v ON_ERROR_STOP=1

echo "ER database '$ER_DB' recreated from '$SRC_DB' with ${#TABLES[@]} business tables."
