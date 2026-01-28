package handlers

import "net/http"

type faqItem struct {
  ID       string
  Question string
  Answer   string
  Category string
}

func (a *App) Support(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  category := r.URL.Query().Get("category")

  faq := []faqItem{
    {ID: "1", Question: "Как начать использовать приложение?", Answer: "Начните с прохождения первичного опросника на главной странице. Это поможет нам составить персональную программу реабилитации с учетом вашего уровня подготовки и ограничений по здоровью.", Category: "Начало работы"},
    {ID: "2", Question: "Как часто нужно заниматься?", Answer: "Рекомендуется заниматься 3-4 раза в неделю. Регулярность важнее интенсивности. Соблюдайте дни отдыха для восстановления мышц.", Category: "Тренировки"},
    {ID: "3", Question: "Что делать, если упражнение вызывает боль?", Answer: "Немедленно прекратите выполнение упражнения. Боль - это сигнал тела о проблеме. Обратитесь к медицинскому специалисту перед продолжением тренировок.", Category: "Безопасность"},
    {ID: "4", Question: "Можно ли изменить программу тренировок?", Answer: "Да, вы можете пройти опросник повторно для корректировки программы. Также можно выбирать отдельные тренировки из библиотеки упражнений.", Category: "Программы"},
    {ID: "5", Question: "Как отслеживать прогресс?", Answer: "Используйте раздел " + "Прогресс" + " для просмотра статистики тренировок, достижений и календаря. Все данные автоматически сохраняются.", Category: "Прогресс"},
    {ID: "6", Question: "Нужно ли специальное оборудование?", Answer: "Большинство упражнений не требуют оборудования. Для некоторых может понадобиться коврик, резинка или гантели. Это указано в описании каждого упражнения.", Category: "Оборудование"},
    {ID: "7", Question: "Как работают напоминания?", Answer: "Настройте напоминания в разделе " + "Настройки" + ". Приложение будет уведомлять вас о запланированных тренировках и новых достижениях.", Category: "Уведомления"},
    {ID: "8", Question: "Можно ли заниматься при хронических заболеваниях?", Answer: "Обязательно проконсультируйтесь с врачом перед началом тренировок. Укажите все ограничения по здоровью в опроснике для составления безопасной программы.", Category: "Здоровье"},
  }

  categories := []string{"Все", "Начало работы", "Тренировки", "Безопасность", "Программы", "Прогресс"}

  filtered := []faqItem{}
  for _, item := range faq {
    if category == "" || category == "Все" || item.Category == category {
      filtered = append(filtered, item)
    }
  }

  data := map[string]any{
    "FAQ":        filtered,
    "Categories": categories,
    "Category":   category,
  }

  a.renderPage(w, r, "support", "Помощь и поддержка", "", data)
}

func (a *App) SupportSubmit(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user == nil {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  category := r.FormValue("category")
  subject := r.FormValue("subject")
  message := r.FormValue("message")

  if category == "" || subject == "" || message == "" {
    a.setFlash(w, "Заполните все поля")
    http.Redirect(w, r, "/support", http.StatusSeeOther)
    return
  }

  _, _ = a.DB.Exec(
    `insert into support_tickets (user_id, category, subject, message)
     values ($1, $2, $3, $4)`,
    user.ID,
    category,
    subject,
    message,
  )

  a.setFlash(w, "Обращение отправлено")
  http.Redirect(w, r, "/support", http.StatusSeeOther)
}
