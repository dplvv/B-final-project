# Мобильное приложение для реабилитации

Полностью Go-приложение: сервер, шаблоны и стили. Интерфейс рендерится на сервере, данные хранятся в PostgreSQL.

## Requirements
- Go 1.22+
- PostgreSQL 14+

## Quick start
1) Запустите PostgreSQL (при желании через Docker):
   `docker-compose up -d`

2) Установите переменные окружения (см. `.env.example`):
   `export DATABASE_URL=postgres://rehab:rehab@localhost:5432/rehab_app?sslmode=disable`

3) Запустите сервер:
   `go run .`

Приложение будет доступно по адресу `http://localhost:8080`.

## Seed accounts (если `SEED_DATA=true`)
- Employee: `10001` / `password`
- Manager: `20001` / `password`
- Admin: `90000` / `password`

## Основные страницы
- `/login`, `/register`
- `/` (дашборд), `/program`, `/exercises`, `/profile`
- `/workouts/{id}`, `/sessions/{id}`

## ER-диаграмма (чистая БД)
Для построения ER-диаграммы без служебных таблиц используйте отдельную БД `rehab_er`:

`./scripts/refresh_er_db.sh rehab_app rehab_er`

Скрипт пересоздаёт `rehab_er` из `rehab_app` только с бизнес-таблицами и связями.
