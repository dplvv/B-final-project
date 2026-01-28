import { useEffect, useState } from 'react';
import type { Screen } from '../types';
import { apiRequest } from '../api/client';

interface WorkoutCompleteProps {
  sessionId: string;
  onNavigate: (screen: Screen) => void;
}

export function WorkoutComplete({ sessionId, onNavigate }: WorkoutCompleteProps) {
  const [summary, setSummary] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiRequest<any>(`/workout-sessions/${sessionId}/summary`);
        setSummary(data);
      } catch {
        setSummary(null);
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [sessionId]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <div className="text-sm text-slate-500">Загрузка...</div>
      </div>
    );
  }

  if (!summary) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gradient-to-b from-green-50 to-white flex flex-col items-center justify-center p-6">
      <div className="bg-white rounded-3xl p-8 shadow-lg max-w-md w-full text-center">
        <div className="bg-green-100 w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-6">
          <svg className="w-10 h-10 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>

        <h1 className="text-2xl mb-2">Тренировка завершена!</h1>
        <p className="text-gray-600 mb-6">{summary.workout_name}</p>

        <div className="grid grid-cols-2 gap-4 mb-6">
          <div className="bg-gray-50 rounded-xl p-4">
            <div className="text-sm text-gray-600 mb-1">Время</div>
            <div className="text-xl text-gray-900">{summary.duration} мин</div>
          </div>
          <div className="bg-gray-50 rounded-xl p-4">
            <div className="text-sm text-gray-600 mb-1">Упражнения</div>
            <div className="text-xl text-gray-900">{summary.completed_exercises}/{summary.total_exercises}</div>
          </div>
          <div className="bg-gray-50 rounded-xl p-4">
            <div className="text-sm text-gray-600 mb-1">Калории</div>
            <div className="text-xl text-gray-900">{summary.calories}</div>
          </div>
          <div className="bg-gray-50 rounded-xl p-4">
            <div className="text-sm text-gray-600 mb-1">Баллы</div>
            <div className="text-xl text-gray-900">+10</div>
          </div>
        </div>

        <button
          onClick={() => onNavigate('feedback')}
          className="w-full bg-blue-600 text-white py-3 rounded-xl mb-3"
        >
          Оставить отзыв
        </button>
        <button
          onClick={() => onNavigate('home')}
          className="w-full bg-gray-900 text-white py-3 rounded-xl"
        >
          Вернуться на главную
        </button>
      </div>
    </div>
  );
}
