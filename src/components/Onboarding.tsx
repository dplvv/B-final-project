import { useState } from 'react';
import type { Screen } from '../types';
import { apiRequest } from '../api/client';
import { useAuth } from '../context/AuthContext';

interface OnboardingProps {
  onComplete: () => void;
  onNavigate: (screen: Screen) => void;
}

const slides = [
  {
    id: 1,
    icon: '💪',
    title: 'Персональная программа реабилитации',
    description: 'Индивидуальные тренировки, адаптированные под ваш уровень подготовки и цели',
    color: 'from-blue-600 to-blue-700'
  },
  {
    id: 2,
    icon: '📊',
    title: 'Отслеживайте прогресс',
    description: 'Следите за достижениями, статистикой тренировок и своим развитием',
    color: 'from-purple-600 to-purple-700'
  },
  {
    id: 3,
    icon: '🏆',
    title: 'Достижения и мотивация',
    description: 'Получайте награды за выполнение целей и участвуйте в челленджах',
    color: 'from-orange-600 to-orange-700'
  },
  {
    id: 4,
    icon: '👥',
    title: 'Сообщество РОСАТОМ',
    description: 'Соревнуйтесь с коллегами, делитесь успехами и поддерживайте друг друга',
    color: 'from-green-600 to-green-700'
  },
  {
    id: 5,
    icon: '🔒',
    title: 'Безопасность прежде всего',
    description: 'Все упражнения разработаны с учетом медицинских рекомендаций и ограничений',
    color: 'from-red-600 to-red-700'
  }
];

export function Onboarding({ onComplete }: OnboardingProps) {
  const { refresh } = useAuth();
  const [currentSlide, setCurrentSlide] = useState(0);
  const [saving, setSaving] = useState(false);

  const finalize = async () => {
    setSaving(true);
    try {
      await apiRequest('/onboarding/complete', { method: 'POST' });
      await refresh();
      onComplete();
    } finally {
      setSaving(false);
    }
  };

  const handleNext = () => {
    if (currentSlide < slides.length - 1) {
      setCurrentSlide(currentSlide + 1);
    } else {
      finalize();
    }
  };

  const handleSkip = () => {
    finalize();
  };

  const slide = slides[currentSlide];

  return (
    <div className="min-h-screen bg-white flex flex-col">
      <div className="p-4 flex justify-end">
        <button
          onClick={handleSkip}
          className="text-gray-600 text-sm px-4 py-2 hover:text-gray-900"
        >
          Пропустить
        </button>
      </div>

      <div className="flex-1 flex flex-col items-center justify-center p-6">
        <div className={`w-32 h-32 rounded-full bg-gradient-to-br ${slide.color} flex items-center justify-center text-6xl mb-8 shadow-lg`}>
          {slide.icon}
        </div>

        <h1 className="text-center mb-4 px-4">
          {slide.title}
        </h1>

        <p className="text-center text-gray-600 px-8 max-w-md">
          {slide.description}
        </p>
      </div>

      <div className="p-6 space-y-4">
        <div className="flex justify-center gap-2">
          {slides.map((_, index) => (
            <button
              key={index}
              onClick={() => setCurrentSlide(index)}
              className={`transition-all ${
                index === currentSlide
                  ? 'w-8 h-2 bg-blue-600 rounded-full'
                  : 'w-2 h-2 bg-gray-300 rounded-full'
              }`}
            />
          ))}
        </div>

        <button
          onClick={handleNext}
          disabled={saving}
          className="w-full bg-blue-600 text-white py-4 rounded-xl flex items-center justify-center gap-2 disabled:opacity-60"
        >
          {saving
            ? 'Сохраняем...'
            : currentSlide < slides.length - 1 ? (
              <>
                Далее
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </>
            ) : (
              <>
                Начать
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </>
            )}
        </button>

        <div className="flex items-center justify-center gap-2 text-sm text-gray-500">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <span>ГК РОСАТОМ</span>
        </div>
      </div>
    </div>
  );
}
