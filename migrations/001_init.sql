-- +migrate Up
create extension if not exists "uuid-ossp";

create table if not exists users (
  id uuid primary key default uuid_generate_v4(),
  name text not null,
  employee_id text not null unique,
  password_hash text not null,
  role text not null default 'employee',
  department text,
  position text,
  password_temp boolean not null default false,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists user_profiles (
  user_id uuid primary key references users(id) on delete cascade,
  age int,
  fitness_level text,
  goals text[] not null default '{}',
  updated_at timestamptz not null default now()
);

create table if not exists questionnaire_responses (
  user_id uuid primary key references users(id) on delete cascade,
  answers jsonb not null default '{}',
  updated_at timestamptz not null default now()
);

create table if not exists medical_info (
  user_id uuid primary key references users(id) on delete cascade,
  doctor_approval boolean not null default false,
  restrictions text[] not null default '{}',
  updated_at timestamptz not null default now()
);

create table if not exists exercises (
  id uuid primary key default uuid_generate_v4(),
  name text not null unique,
  description text not null,
  category text,
  difficulty text,
  sets int,
  reps text,
  duration_seconds int,
  rest_seconds int,
  muscle_groups text[] not null default '{}',
  equipment text[] not null default '{}',
  image_url text,
  created_at timestamptz not null default now()
);

create table if not exists workouts (
  id uuid primary key default uuid_generate_v4(),
  name text not null unique,
  description text not null,
  duration_minutes int not null,
  difficulty text not null,
  category text,
  created_at timestamptz not null default now()
);

create table if not exists workout_exercises (
  workout_id uuid references workouts(id) on delete cascade,
  exercise_id uuid references exercises(id) on delete cascade,
  sort_order int not null,
  sets int,
  reps text,
  duration_seconds int,
  rest_seconds int,
  primary key (workout_id, exercise_id)
);

create table if not exists programs (
  id uuid primary key default uuid_generate_v4(),
  name text not null,
  description text not null,
  active boolean not null default true,
  muscle_groups text[] not null default '{}',
  created_at timestamptz not null default now()
);

create table if not exists program_workouts (
  program_id uuid references programs(id) on delete cascade,
  workout_id uuid references workouts(id) on delete cascade,
  sort_order int not null,
  primary key (program_id, workout_id)
);

create table if not exists user_programs (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  program_id uuid references programs(id) on delete cascade,
  start_date date not null,
  active boolean not null default true
);

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
  skip_reason text
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

create table if not exists workout_sessions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  workout_id uuid references workouts(id) on delete cascade,
  started_at timestamptz not null default now(),
  completed_at timestamptz,
  last_set_completed_at timestamptz,
  duration_minutes int,
  total_exercises int,
  completed_exercises int,
  calories_burned int,
  plan_workout_id uuid references training_plan_workouts(id) on delete set null
);

alter table training_plan_workouts
  add column if not exists session_id uuid references workout_sessions(id) on delete set null;

create table if not exists workout_session_exercises (
  id uuid primary key default uuid_generate_v4(),
  session_id uuid references workout_sessions(id) on delete cascade,
  exercise_id uuid references exercises(id) on delete cascade,
  sort_order int not null,
  completed_sets int not null default 0,
  notes text,
  completed boolean not null default false
);

create table if not exists workout_session_sets (
  id uuid primary key default uuid_generate_v4(),
  session_id uuid references workout_sessions(id) on delete cascade,
  session_exercise_id uuid references workout_session_exercises(id) on delete cascade,
  set_index int not null,
  started_at timestamptz not null,
  completed_at timestamptz not null,
  duration_seconds int not null
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

create table if not exists achievements (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  description text not null,
  icon text not null,
  points_reward int not null default 0,
  metric text,
  target int,
  created_at timestamptz not null default now()
);

create table if not exists user_achievements (
  user_id uuid references users(id) on delete cascade,
  achievement_id uuid references achievements(id) on delete cascade,
  unlocked boolean not null default false,
  unlocked_at timestamptz,
  progress int not null default 0,
  total int not null default 0,
  primary key (user_id, achievement_id)
);

create table if not exists user_points (
  user_id uuid primary key references users(id) on delete cascade,
  points_balance int not null default 0,
  points_total int not null default 0,
  updated_at timestamptz not null default now()
);

create table if not exists rewards (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  description text not null,
  points_cost int not null,
  category text,
  active boolean not null default true
);

create table if not exists reward_redemptions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  reward_id uuid references rewards(id) on delete cascade,
  status text not null default 'pending',
  redeemed_at timestamptz not null default now(),
  handled_at timestamptz,
  approved_by uuid references users(id) on delete set null
);

create table if not exists incentive_awards (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  points int not null,
  reason text,
  awarded_by uuid references users(id) on delete set null,
  awarded_at timestamptz not null default now()
);

create table if not exists support_tickets (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  category text not null,
  subject text not null,
  message text not null,
  status text not null default 'open',
  response text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists support_ticket_messages (
  id uuid primary key default uuid_generate_v4(),
  ticket_id uuid references support_tickets(id) on delete cascade,
  sender_id uuid references users(id) on delete set null,
  sender_role text not null,
  message text not null,
  created_at timestamptz not null default now()
);

alter table exercises add column if not exists image_url text;
alter table exercises drop column if exists video_url;
alter table workout_sessions add column if not exists last_set_completed_at timestamptz;
alter table achievements add column if not exists metric text;
alter table achievements add column if not exists target int;

create table if not exists password_reset_requests (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  status text not null default 'open',
  created_at timestamptz not null default now(),
  handled_at timestamptz,
  handled_by uuid references users(id) on delete set null
);

create table if not exists sessions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  token text not null unique,
  expires_at timestamptz not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_sessions_token on sessions(token);
create index if not exists idx_workout_sessions_user on workout_sessions(user_id);
create index if not exists idx_workout_session_sets_session on workout_session_sets(session_id);
create index if not exists idx_workout_session_sets_exercise on workout_session_sets(session_exercise_id);
create index if not exists idx_plan_user on training_plans(user_id);
create index if not exists idx_plan_workouts_plan on training_plan_workouts(plan_id);
create index if not exists idx_plan_workouts_status on training_plan_workouts(status);
create index if not exists idx_plan_changes_plan on training_plan_changes(plan_id);
create index if not exists idx_support_messages_ticket on support_ticket_messages(ticket_id);
create index if not exists idx_password_reset_status on password_reset_requests(status);

-- +migrate Down
-- (intentionally left blank)
