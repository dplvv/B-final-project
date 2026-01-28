package api

import (
  "net/http"
  "time"
)

type supportTicketRequest struct {
  Category string `json:"category"`
  Subject  string `json:"subject"`
  Message  string `json:"message"`
}

func (api *API) SupportFAQ(w http.ResponseWriter, r *http.Request) {
  faq := []map[string]any{
    {"id": "1", "question": "Как начать использовать приложение?", "answer": "Начните с прохождения первичного опросника на главной странице. Это поможет нам составить персональную программу реабилитации с учетом вашего уровня подготовки и ограничений по здоровью.", "category": "Начало работы"},
    {"id": "2", "question": "Как часто нужно заниматься?", "answer": "Рекомендуется заниматься 3-4 раза в неделю. Регулярность важнее интенсивности. Соблюдайте дни отдыха для восстановления мышц.", "category": "Тренировки"},
    {"id": "3", "question": "Что делать, если упражнение вызывает боль?", "answer": "Немедленно прекратите выполнение упражнения. Боль - это сигнал тела о проблеме. Обратитесь к медицинскому специалисту перед продолжением тренировок.", "category": "Безопасность"},
    {"id": "4", "question": "Можно ли изменить программу тренировок?", "answer": "Да, вы можете пройти опросник повторно для корректировки программы. Также можно выбирать отдельные тренировки из библиотеки упражнений.", "category": "Программы"},
    {"id": "5", "question": "Как отслеживать прогресс?", "answer": "Используйте раздел Прогресс для просмотра статистики тренировок, достижений и календаря. Все данные автоматически сохраняются.", "category": "Прогресс"},
    {"id": "6", "question": "Нужно ли специальное оборудование?", "answer": "Большинство упражнений не требуют оборудования. Для некоторых может понадобиться коврик, резинка или гантели. Это указано в описании каждого упражнения.", "category": "Оборудование"},
    {"id": "7", "question": "Как работают напоминания?", "answer": "Настройте напоминания в разделе Настройки. Приложение будет уведомлять вас о запланированных тренировках и новых достижениях.", "category": "Уведомления"},
    {"id": "8", "question": "Можно ли заниматься при хронических заболеваниях?", "answer": "Обязательно проконсультируйтесь с врачом перед началом тренировок. Укажите все ограничения по здоровью в опроснике для составления безопасной программы.", "category": "Здоровье"},
  }

  categories := []string{"Все", "Начало работы", "Тренировки", "Безопасность", "Программы", "Прогресс", "Оборудование", "Уведомления", "Здоровье"}

  writeJSON(w, http.StatusOK, map[string]any{
    "faq": faq,
    "categories": categories,
  })
}

func (api *API) SupportTickets(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  rows, err := api.DB.Query(
    `select id, category, subject, status, created_at
     from support_tickets
     where user_id = $1
     order by created_at desc`,
    userID,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }
  defer rows.Close()

  tickets := []map[string]any{}
  for rows.Next() {
    var id, category, subject, status string
    var createdAt time.Time
    _ = rows.Scan(&id, &category, &subject, &status, &createdAt)
    tickets = append(tickets, map[string]any{
      "id": id,
      "category": category,
      "subject": subject,
      "status": status,
      "created_at": createdAt.Format("2006-01-02"),
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"tickets": tickets})
}

func (api *API) SupportCreate(w http.ResponseWriter, r *http.Request) {
  userID := userIDFromContext(r.Context())

  var req supportTicketRequest
  if err := decodeJSON(r, &req); err != nil {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payload"})
    return
  }

  if req.Category == "" || req.Subject == "" || req.Message == "" {
    writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing fields"})
    return
  }

  _, err := api.DB.Exec(
    `insert into support_tickets (user_id, category, subject, message)
     values ($1, $2, $3, $4)`,
    userID,
    req.Category,
    req.Subject,
    req.Message,
  )
  if err != nil {
    writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "server error"})
    return
  }

  writeJSON(w, http.StatusCreated, map[string]any{"status": "created"})
}
