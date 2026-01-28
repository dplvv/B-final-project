package handlers

import (
  "net/http"
  "strconv"
)

type onboardingSlide struct {
  ID          int
  Icon        string
  Title       string
  Description string
  Color       string
}

var onboardingSlides = []onboardingSlide{
  {ID: 1, Icon: "💪", Title: "Персональная программа реабилитации", Description: "Индивидуальные тренировки, адаптированные под ваш уровень подготовки и цели", Color: "from-blue-600 to-blue-700"},
  {ID: 2, Icon: "📊", Title: "Отслеживайте прогресс", Description: "Следите за достижениями, статистикой тренировок и своим развитием", Color: "from-purple-600 to-purple-700"},
  {ID: 3, Icon: "🏆", Title: "Достижения и мотивация", Description: "Получайте награды за выполнение целей и участвуйте в челленджах", Color: "from-orange-600 to-orange-700"},
  {ID: 4, Icon: "👥", Title: "Сообщество РОСАТОМ", Description: "Соревнуйтесь с коллегами, делитесь успехами и поддерживайте друг друга", Color: "from-green-600 to-green-700"},
  {ID: 5, Icon: "🔒", Title: "Безопасность прежде всего", Description: "Все упражнения разработаны с учетом медицинских рекомендаций и ограничений", Color: "from-red-600 to-red-700"},
}

func (a *App) Onboarding(w http.ResponseWriter, r *http.Request) {
  slideIndex := 0
  if value := r.URL.Query().Get("slide"); value != "" {
    if parsed, err := strconv.Atoi(value); err == nil {
      slideIndex = parsed
    }
  }
  if slideIndex < 0 {
    slideIndex = 0
  }
  if slideIndex >= len(onboardingSlides) {
    slideIndex = len(onboardingSlides) - 1
  }

  data := map[string]any{
    "Slide":      onboardingSlides[slideIndex],
    "SlideIndex": slideIndex,
    "Slides":     onboardingSlides,
    "IsLast":     slideIndex == len(onboardingSlides)-1,
  }

  a.renderFullPage(w, r, "onboarding", "Онбординг", data)
}

func (a *App) OnboardingNext(w http.ResponseWriter, r *http.Request) {
  if err := r.ParseForm(); err != nil {
    http.Error(w, "invalid form", http.StatusBadRequest)
    return
  }

  slideIndex, _ := strconv.Atoi(r.FormValue("slide"))
  slideIndex++

  if slideIndex >= len(onboardingSlides) {
    user := a.userFromRequest(r)
    if user != nil {
      _, _ = a.DB.Exec("update user_profiles set onboarding_complete = true where user_id = $1", user.ID)
    }
    http.Redirect(w, r, "/questionnaire", http.StatusSeeOther)
    return
  }

  http.Redirect(w, r, "/onboarding?slide="+strconv.Itoa(slideIndex), http.StatusSeeOther)
}

func (a *App) OnboardingSkip(w http.ResponseWriter, r *http.Request) {
  user := a.userFromRequest(r)
  if user != nil {
    _, _ = a.DB.Exec("update user_profiles set onboarding_complete = true where user_id = $1", user.ID)
  }
  http.Redirect(w, r, "/questionnaire", http.StatusSeeOther)
}
