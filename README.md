
  # Мобильное приложение для реабилитации

  This is a code bundle for Мобильное приложение для реабилитации. The original project is available at https://www.figma.com/design/56fdQBGrkHjTEkRmJZV4OM/%D0%9C%D0%BE%D0%B1%D0%B8%D0%BB%D1%8C%D0%BD%D0%BE%D0%B5-%D0%BF%D1%80%D0%B8%D0%BB%D0%BE%D0%B6%D0%B5%D0%BD%D0%B8%D0%B5-%D0%B4%D0%BB%D1%8F-%D1%80%D0%B5%D0%B0%D0%B1%D0%B8%D0%BB%D0%B8%D1%82%D0%B0%D1%86%D0%B8%D0%B8.

  ## Go backend + SSR frontend

  This version is a pure Go app with server-rendered HTML and PostgreSQL storage.

  ### Requirements
  - Go 1.22+
  - PostgreSQL 14+

  ### Quick start
  1) Start PostgreSQL (optional with Docker):
     `docker-compose up -d`

  2) Set environment variables (see `.env.example`):
     `export DATABASE_URL=postgres://rehab:rehab@localhost:5432/rehab_app?sslmode=disable`

  3) Run the server:
     `go run .`

  The app will be available at `http://localhost:8080`.

  ### Seed accounts (if `SEED_DATA=true`)
  - Employee: `10001` / `password`
  - Manager: `20001` / `password`
  - Admin: `90000` / `password`

  ### API
  - `POST /api/v1/auth/login` -> `{ "employee_id": "...", "password": "..." }`
  - Use `Authorization: Bearer <token>` for other API calls.
  
