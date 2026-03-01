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

## Первый администратор (без seed)
Модуль не заполняет БД начальными данными. Чтобы получить доступ в админ-панель:
1) Зарегистрируйте обычного пользователя через `/register`.
2) Назначьте ему роль администратора в БД:
   `update users set role = 'admin' where employee_id = '<ID_сотрудника>';`

## Основные страницы
- `/login`, `/register`
- `/` (дашборд), `/program`, `/exercises`, `/profile`
- `/workouts/{id}`, `/sessions/{id}`
