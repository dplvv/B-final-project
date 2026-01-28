import { useEffect, useMemo, useState } from 'react';
import type { Screen } from '../types';
import { apiRequest } from '../api/client';

interface FeedbackProps {
  onNavigate: (screen: Screen) => void;
  sessionId?: string | null;
}

export function Feedback({ onNavigate, sessionId }: FeedbackProps) {
  const [rating, setRating] = useState<number>(0);
  const [difficulty, setDifficulty] = useState<string>('');
  const [comment, setComment] = useState<string>('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [goalResult, setGoalResult] = useState<string>('');
  const [recommendation, setRecommendation] = useState<string>('');
  const [submitted, setSubmitted] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<any[]>([]);

  const difficultyOptions = [
    { id: 'too-easy', label: 'Слишком легко', emoji: '😊' },
    { id: 'just-right', label: 'В самый раз', emoji: '💪' },
    { id: 'challenging', label: 'Сложно', emoji: '😅' },
    { id: 'too-hard', label: 'Очень сложно', emoji: '😰' },
  ];

  const feedbackTags = [
    'Понятные инструкции',
    'Хорошее время',
    'Эффективные упражнения',
    'Нужно больше отдыха',
    'Слишком интенсивно',
    'Хорошая музыка',
    'Мало упражнений',
    'Отличная программа',
  ];

  useEffect(() => {
    const loadHistory = async () => {
      try {
        const data = await apiRequest<{ feedbacks: any[] }>('/feedback');
        setHistory(data.feedbacks || []);
      } catch {
        setHistory([]);
      }
    };
    loadHistory();
  }, []);

  const compiledComment = useMemo(() => {
    const parts: string[] = [];
    if (difficulty) {
      const label = difficultyOptions.find((option) => option.id === difficulty)?.label;
      if (label) parts.push(`Сложность: ${label}`);
    }
    if (selectedTags.length > 0) {
      parts.push(`Теги: ${selectedTags.join(', ')}`);
    }
    if (goalResult) {
      parts.push(`Цель: ${goalResult}`);
    }
    if (recommendation) {
      parts.push(`Рекомендация: ${recommendation}`);
    }
    if (comment) {
      parts.push(`Комментарий: ${comment}`);
    }
    return parts.join(' | ');
  }, [difficulty, selectedTags, goalResult, recommendation, comment]);

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag)
        ? prev.filter((item) => item !== tag)
        : [...prev, tag]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await apiRequest('/feedback', {
        method: 'POST',
        body: JSON.stringify({
          session_id: sessionId || '',
          rating,
          comment: compiledComment || comment,
        }),
      });
      setSubmitted(true);
      setTimeout(() => {
        onNavigate('home');
      }, 2000);
    } catch (err: any) {
      setError(err?.message || 'Не удалось отправить отзыв');
    } finally {
      setLoading(false);
    }
  };

  if (submitted) {
    return (
      <div className="min-h-screen bg-gradient-to-b from-green-50 to-white flex items-center justify-center p-6">
        <div className="text-center">
          <div className="bg-green-500 w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-2xl mb-2">Спасибо за отзыв!</h2>
          <p className="text-gray-600">Ваше мнение помогает нам становиться лучше</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-purple-600 text-white p-6">
        <button
          onClick={() => onNavigate('home')}
          className="flex items-center gap-2 mb-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="text-2xl mb-2">Обратная связь</h1>
        <p className="text-purple-100">Расскажите о своём опыте</p>
      </div>

      <form onSubmit={handleSubmit} className="p-6 space-y-6">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4 text-center">Как вам тренировка?</h3>
          <div className="flex justify-center gap-4">
            {[1, 2, 3, 4, 5].map((star) => (
              <button
                key={star}
                type="button"
                onClick={() => setRating(star)}
                className="transition-transform hover:scale-110"
              >
                <svg
                  className={`w-10 h-10 ${
                    star <= rating ? 'text-yellow-400 fill-yellow-400' : 'text-gray-300'
                  }`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
                  />
                </svg>
              </button>
            ))}
          </div>
          {rating > 0 && (
            <p className="text-center text-sm text-gray-600 mt-3">
              {rating === 5 && 'Превосходно! 🌟'}
              {rating === 4 && 'Отлично! 👍'}
              {rating === 3 && 'Хорошо ✓'}
              {rating === 2 && 'Нормально'}
              {rating === 1 && 'Нужно улучшить'}
            </p>
          )}
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Уровень сложности</h3>
          <div className="grid grid-cols-2 gap-3">
            {difficultyOptions.map((option) => (
              <button
                key={option.id}
                type="button"
                onClick={() => setDifficulty(option.id)}
                className={`p-4 rounded-xl border-2 transition-all ${
                  difficulty === option.id
                    ? 'border-purple-600 bg-purple-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <div className="text-3xl mb-1">{option.emoji}</div>
                <div className="text-sm">{option.label}</div>
              </button>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Отметьте подходящее</h3>
          <div className="flex flex-wrap gap-2">
            {feedbackTags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => toggleTag(tag)}
                className={`px-4 py-2 rounded-full text-sm transition-colors ${
                  selectedTags.includes(tag)
                    ? 'bg-purple-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {tag}
              </button>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Комментарий (необязательно)</h3>
          <textarea
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-purple-500 resize-none"
            rows={4}
            placeholder="Поделитесь подробнее своими впечатлениями..."
          />
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Дополнительные вопросы</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-gray-700 mb-2">
                Достигли ли вы своей цели на тренировке?
              </label>
              <div className="flex gap-2">
                {['Да', 'Частично', 'Нет'].map((value) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setGoalResult(value)}
                    className={`flex-1 py-2 px-4 rounded-xl border transition-colors ${
                      goalResult === value
                        ? 'border-purple-600 bg-purple-50 text-purple-700'
                        : 'border-gray-200 text-gray-700'
                    }`}
                  >
                    {value}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm text-gray-700 mb-2">
                Будете ли вы рекомендовать эту программу коллегам?
              </label>
              <div className="flex gap-2">
                {['Да, обязательно', 'Возможно', 'Нет'].map((value) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setRecommendation(value)}
                    className={`flex-1 py-2 px-4 rounded-xl border transition-colors ${
                      recommendation === value
                        ? 'border-purple-600 bg-purple-50 text-purple-700'
                        : 'border-gray-200 text-gray-700'
                    }`}
                  >
                    {value}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        <button
          type="submit"
          disabled={rating === 0 || loading}
          className={`w-full py-4 rounded-xl text-white transition-colors ${
            rating === 0 || loading ? 'bg-gray-300 cursor-not-allowed' : 'bg-purple-600 hover:bg-purple-700'
          }`}
        >
          {loading ? 'Отправляем...' : 'Отправить отзыв'}
        </button>

        <div className="bg-gray-50 rounded-2xl p-4">
          <div className="flex items-start gap-2 text-sm text-gray-600">
            <svg className="w-5 h-5 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
            <p>
              Ваши отзывы используются только для улучшения программы тренировок.
            </p>
          </div>
        </div>
      </form>

      {history.length > 0 && (
        <div className="px-6 pb-8">
          <h3 className="mb-3">Ваши последние отзывы</h3>
          <div className="space-y-3">
            {history.map((item) => (
              <div key={item.id} className="bg-white rounded-2xl p-4 shadow-sm">
                <div className="flex items-center justify-between mb-2">
                  <div className="text-sm text-gray-600">{item.workout || 'Тренировка'}</div>
                  <div className="text-xs text-gray-500">{item.created}</div>
                </div>
                <div className="flex items-center gap-1 mb-2">
                  {Array.from({ length: 5 }).map((_, index) => (
                    <svg
                      key={index}
                      className={`w-4 h-4 ${index < (item.rating || 0) ? 'text-yellow-400' : 'text-gray-300'}`}
                      fill="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                    </svg>
                  ))}
                </div>
                {item.comment && <div className="text-sm text-gray-600">{item.comment}</div>}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
