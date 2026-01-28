-- +migrate Up
create table if not exists nutrition_water_logs (
  user_id uuid references users(id) on delete cascade,
  log_date date not null,
  amount_ml int not null default 0,
  goal_ml int not null default 2500,
  updated_at timestamptz not null default now(),
  primary key (user_id, log_date)
);

create table if not exists nutrition_diary (
  id uuid primary key default uuid_generate_v4(),
  user_id uuid references users(id) on delete cascade,
  log_date date not null,
  title text not null,
  calories int not null default 0,
  meal_type text not null default 'other',
  created_at timestamptz not null default now()
);

create index if not exists idx_nutrition_water_user_date on nutrition_water_logs(user_id, log_date);
create index if not exists idx_nutrition_diary_user_date on nutrition_diary(user_id, log_date);

-- +migrate Down
-- (intentionally left blank for now)
