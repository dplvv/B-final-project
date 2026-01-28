-- +migrate Up
alter table user_settings
  add column if not exists workout_reminders boolean not null default true;

alter table user_settings
  add column if not exists achievement_alerts boolean not null default true;

alter table user_settings
  add column if not exists weekly_reports boolean not null default false;

alter table user_settings
  add column if not exists share_progress boolean not null default true;

alter table user_settings
  add column if not exists show_in_leaderboard boolean not null default true;

alter table user_settings
  add column if not exists units text not null default 'metric';

create table if not exists recommendation_bookmarks (
  user_id uuid references users(id) on delete cascade,
  recommendation_id uuid references recommendations(id) on delete cascade,
  created_at timestamptz not null default now(),
  primary key (user_id, recommendation_id)
);

-- +migrate Down
-- (intentionally left blank)
