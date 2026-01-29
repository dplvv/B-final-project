-- +migrate Up

create table if not exists training_plans (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  goal text not null,
  level text not null,
  frequency int not null,
  status text not null default 'active',
  paused_reason text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists training_plan_workouts (
  id uuid primary key default uuid_generate_v4(),
  plan_id uuid references training_plans(id) on delete cascade,
  workout_id uuid references workouts(id) on delete cascade,
  week int not null,
  day int not null,
  scheduled_date date,
  intensity int not null default 1,
  status text not null default 'pending',
  skip_reason text,
  session_id uuid references workout_sessions(id) on delete set null
);

create table if not exists training_plan_changes (
  id uuid primary key default uuid_generate_v4(),
  plan_id uuid references training_plans(id) on delete cascade,
  user_id uuid references users(id) on delete cascade,
  changed_at timestamptz not null default now(),
  reason_code text not null,
  reason text not null,
  before_plan jsonb,
  after_plan jsonb
);

create table if not exists workout_session_feedback (
  session_id uuid primary key references workout_sessions(id) on delete cascade,
  user_id uuid references users(id) on delete cascade,
  perceived_exertion int,
  tolerance int,
  pain_level int,
  wellbeing int,
  comment text,
  created_at timestamptz not null default now()
);

alter table workout_sessions
  add column if not exists plan_workout_id uuid references training_plan_workouts(id) on delete set null;

create index if not exists idx_plan_user on training_plans(user_id);
create index if not exists idx_plan_workouts_plan on training_plan_workouts(plan_id);
create index if not exists idx_plan_workouts_status on training_plan_workouts(status);
create index if not exists idx_plan_changes_plan on training_plan_changes(plan_id);

-- +migrate Down
-- (not implemented)
