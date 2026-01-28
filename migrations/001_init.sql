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
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists user_profiles (
  user_id uuid primary key references users(id) on delete cascade,
  age int,
  fitness_level text,
  restrictions text[] not null default '{}',
  goals text[] not null default '{}',
  onboarding_complete boolean not null default false,
  updated_at timestamptz not null default now()
);

create table if not exists questionnaire_responses (
  user_id uuid primary key references users(id) on delete cascade,
  answers jsonb not null default '{}',
  updated_at timestamptz not null default now()
);

create table if not exists user_settings (
  user_id uuid primary key references users(id) on delete cascade,
  notifications_enabled boolean not null default true,
  reminders_enabled boolean not null default true,
  language text not null default 'ru',
  theme text not null default 'light',
  updated_at timestamptz not null default now()
);

create table if not exists medical_info (
  user_id uuid primary key references users(id) on delete cascade,
  chronic_diseases text[] not null default '{}',
  injuries text[] not null default '{}',
  medications text[] not null default '{}',
  allergies text[] not null default '{}',
  doctor_approval boolean not null default false,
  last_checkup date,
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
  video_url text,
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

create table if not exists workout_sessions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  workout_id uuid references workouts(id) on delete cascade,
  started_at timestamptz not null default now(),
  completed_at timestamptz,
  duration_minutes int,
  total_exercises int,
  completed_exercises int,
  calories_burned int
);

create table if not exists workout_session_exercises (
  id uuid primary key default uuid_generate_v4(),
  session_id uuid references workout_sessions(id) on delete cascade,
  exercise_id uuid references exercises(id) on delete cascade,
  sort_order int not null,
  completed_sets int not null default 0,
  notes text,
  completed boolean not null default false
);

create table if not exists achievements (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  description text not null,
  icon text not null,
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

create table if not exists goals (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  title text not null,
  description text not null,
  target_date date,
  progress int not null default 0,
  category text not null,
  created_at timestamptz not null default now()
);

create table if not exists notifications (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  title text not null,
  message text not null,
  type text not null,
  created_at timestamptz not null default now(),
  read_at timestamptz
);

create table if not exists recommendations (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  body text not null,
  category text,
  icon text,
  excerpt text,
  read_time int,
  created_at timestamptz not null default now()
);

create table if not exists calendar_events (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  title text not null,
  event_date date not null,
  event_type text not null,
  metadata jsonb not null default '{}'
);

create table if not exists feedback (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  workout_session_id uuid references workout_sessions(id) on delete set null,
  rating int not null,
  comment text,
  created_at timestamptz not null default now()
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

create table if not exists community_posts (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  title text not null,
  body text not null,
  created_at timestamptz not null default now(),
  likes_count int not null default 0
);

create table if not exists community_comments (
  id uuid primary key default uuid_generate_v4(),
  post_id uuid references community_posts(id) on delete cascade,
  user_id uuid references users(id) on delete cascade,
  body text not null,
  created_at timestamptz not null default now()
);

create table if not exists video_tutorials (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  description text not null,
  duration_minutes int,
  category text,
  difficulty text,
  url text,
  created_at timestamptz not null default now()
);

create table if not exists nutrition_items (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  description text not null,
  calories int,
  category text,
  created_at timestamptz not null default now()
);

create table if not exists rewards (
  id uuid primary key default uuid_generate_v4(),
  title text not null unique,
  description text not null,
  points_cost int not null,
  category text,
  active boolean not null default true
);

create table if not exists user_points (
  user_id uuid primary key references users(id) on delete cascade,
  points_balance int not null default 0,
  points_total int not null default 0,
  updated_at timestamptz not null default now()
);

create table if not exists reward_redemptions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  reward_id uuid references rewards(id) on delete cascade,
  status text not null default 'pending',
  redeemed_at timestamptz not null default now(),
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

create table if not exists sessions (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  token text not null unique,
  expires_at timestamptz not null,
  created_at timestamptz not null default now()
);

create table if not exists api_tokens (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  token text not null unique,
  expires_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_workout_sessions_user on workout_sessions(user_id);
create index if not exists idx_notifications_user on notifications(user_id);
create index if not exists idx_goals_user on goals(user_id);
create index if not exists idx_calendar_user on calendar_events(user_id);
create index if not exists idx_sessions_token on sessions(token);
create index if not exists idx_api_tokens_token on api_tokens(token);

-- +migrate Down
-- (intentionally left blank for now)
