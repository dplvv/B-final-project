import { useEffect, useMemo, useState } from 'react';
import type { WorkoutHistory, Screen } from '../types';
import { apiRequest } from '../api/client';

interface HistoryProps {
  onNavigate: (screen: Screen) => void;
  onStartWorkout?: (workoutId: string) => void;
}

export function History({ onNavigate, onStartWorkout }: HistoryProps) {
  const [history, setHistory] = useState<WorkoutHistory[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'completed' | 'skipped'>('all');
  const [selectedWorkout, setSelectedWorkout] = useState<WorkoutHistory | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadHistory = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiRequest<{ history: any[] }>('/history');
      const items = (data.history || []).map((item) => ({
        id: item.id,
        workoutId: item.workout_id,
        workoutName: item.workout_name,
        date: item.date,
        duration: Number(item.duration || 0),
        completedExercises: Number(item.completed_exercises || 0),
        totalExercises: Number(item.total_exercises || 0),
        completed: Boolean(item.completed),
        calories: Number(item.calories || 0),
        rating: item.rating ? Number(item.rating) : undefined,
      }));
      setHistory(items);
    } catch (err: any) {
      setError(err?.message || 'Не удалось загрузить историю');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadHistory();
  }, []);

  const filteredHistory = useMemo(() => {
    return history.filter((workout) => {
      if (filter === 'completed') return workout.completed;
      if (filter === 'skipped') return !workout.completed;
      return true;
    });
  }, [history, filter]);

  const groupedHistory = useMemo(() => {
    const grouped: Record<string, WorkoutHistory[]> = {};
    filteredHistory.forEach((workout) => {
      const date = new Date(workout.date);
      const monthYear = date.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
      if (!grouped[monthYear]) {
        grouped[monthYear] = [];
      }
      grouped[monthYear].push(workout);
    });
    return grouped;
  }, [filteredHistory]);

  const completedCount = history.filter((workout) => workout.completed).length;
  const totalDuration = history.filter((workout) => workout.completed).reduce((sum, workout) => sum + workout.duration, 0);
  const totalCalories = history.filter((workout) => workout.completed).reduce((sum, workout) => sum + (workout.calories || 0), 0);
  const ratings = history.filter((workout) => workout.rating).map((workout) => workout.rating || 0);
  const averageRating = ratings.length > 0 ? ratings.reduce((sum, value) => sum + value, 0) / ratings.length : 0;

  const handleRepeat = async () => {
    if (selectedWorkout?.workoutId && onStartWorkout) {
      await onStartWorkout(selectedWorkout.workoutId);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-gradient-to-br from-teal-600 to-cyan-600 text-white p-6">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('profile')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>История тренировок</h1>
        </div>
        <p className="text-teal-100 text-sm">Полный журнал активности</p>
      </div>

      <div className="p-4 space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        <div className="grid grid-cols-2 gap-3">
          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="flex items-center gap-2 text-blue-600 mb-2">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-xs">Завершено</span>
            </div>
            <div className="text-2xl">{completedCount}</div>
          </div>

          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="flex items-center gap-2 text-green-600 mb-2">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-xs">Минут</span>
            </div>
            <div className="text-2xl">{totalDuration}</div>
          </div>

          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="flex items-center gap-2 text-orange-600 mb-2">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
              </svg>
              <span className="text-xs">Калорий</span>
            </div>
            <div className="text-2xl">{totalCalories}</div>
          </div>

          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="flex items-center gap-2 text-yellow-600 mb-2">
              <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
              </svg>
              <span className="text-xs">Средний рейтинг</span>
            </div>
            <div className="text-2xl">{averageRating > 0 ? averageRating.toFixed(1) : '—'}</div>
          </div>
        </div>

        <div className="flex gap-2">
          <button
            onClick={() => setFilter('all')}
            className={`flex-1 py-2 rounded-xl text-sm transition-colors ${
              filter === 'all' ? 'bg-teal-600 text-white' : 'bg-white text-gray-700'
            }`}
          >
            Все ({history.length})
          </button>
          <button
            onClick={() => setFilter('completed')}
            className={`flex-1 py-2 rounded-xl text-sm transition-colors ${
              filter === 'completed' ? 'bg-teal-600 text-white' : 'bg-white text-gray-700'
            }`}
          >
            Завершено ({completedCount})
          </button>
          <button
            onClick={() => setFilter('skipped')}
            className={`flex-1 py-2 rounded-xl text-sm transition-colors ${
              filter === 'skipped' ? 'bg-teal-600 text-white' : 'bg-white text-gray-700'
            }`}
          >
            Пропущено ({history.length - completedCount})
          </button>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Загрузка истории...</div>
        ) : (
          Object.keys(groupedHistory).map((monthYear) => (
            <div key={monthYear}>
              <h3 className="mb-3 capitalize text-gray-700">{monthYear}</h3>
              <div className="space-y-3 mb-4">
                {groupedHistory[monthYear].map((workout) => (
                  <div
                    key={workout.id}
                    onClick={() => setSelectedWorkout(workout)}
                    className={`rounded-2xl p-4 shadow-sm cursor-pointer transition-all ${
                      workout.completed
                        ? 'bg-white border-2 border-transparent hover:border-teal-200'
                        : 'bg-gray-50 border-2 border-gray-200'
                    }`}
                  >
                    <div className="flex items-start gap-4">
                      <div className={`rounded-xl p-3 text-center flex-shrink-0 ${
                        workout.completed ? 'bg-teal-50' : 'bg-gray-100'
                      }`}>
                        <div className={`text-lg mb-1 ${
                          workout.completed ? 'text-teal-600' : 'text-gray-500'
                        }`}>
                          {new Date(workout.date).getDate()}
                        </div>
                        <div className="text-xs text-gray-600">
                          {new Date(workout.date).toLocaleDateString('ru-RU', { weekday: 'short' })}
                        </div>
                      </div>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between mb-2">
                          <h4 className={workout.completed ? '' : 'text-gray-500'}>
                            {workout.workoutName}
                          </h4>
                          {workout.completed ? (
                            <svg className="w-6 h-6 text-green-600 flex-shrink-0 ml-2" fill="currentColor" viewBox="0 0 24 24">
                              <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                          ) : (
                            <svg className="w-6 h-6 text-gray-400 flex-shrink-0 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          )}
                        </div>

                        <div className="flex flex-wrap items-center gap-4 text-sm text-gray-600 mb-2">
                          <div className="flex items-center gap-1">
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            {workout.duration} мин
                          </div>
                          <div className="flex items-center gap-1">
                            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 10l-2 1m0 0l-2-1m2 1v2.5M20 7l-2 1m2-1l-2-1m2 1v2.5M14 4l-2-1-2 1M4 7l2-1M4 7l2 1M4 7v2.5M12 21l-2-1m2 1l2-1m-2 1v-2.5M6 18l-2-1v-2.5M18 18l2-1v-2.5" />
                            </svg>
                            {workout.completedExercises}/{workout.totalExercises} упр.
                          </div>
                          {workout.calories && workout.completed && (
                            <div className="flex items-center gap-1">
                              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z" />
                              </svg>
                              {workout.calories} ккал
                            </div>
                          )}
                        </div>

                        {workout.rating && (
                          <div className="flex items-center gap-1">
                            {Array.from({ length: 5 }).map((_, i) => (
                              <svg
                                key={i}
                                className={`w-4 h-4 ${i < (workout.rating || 0) ? 'text-yellow-400' : 'text-gray-300'}`}
                                fill="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                              </svg>
                            ))}
                          </div>
                        )}

                        {!workout.completed && (
                          <div className="text-sm text-gray-500 italic mt-2">Пропущена</div>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))
        )}

        {!loading && filteredHistory.length === 0 && (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">📊</div>
            <h3 className="mb-2">История пуста</h3>
            <p className="text-gray-600 text-sm">Начните тренироваться, чтобы увидеть историю</p>
          </div>
        )}
      </div>

      {selectedWorkout && (
        <div
          className="fixed inset-0 bg-black/50 flex items-end justify-center z-50"
          onClick={() => setSelectedWorkout(null)}
        >
          <div
            className="bg-white rounded-t-3xl w-full max-w-md max-h-[80vh] overflow-y-auto"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="p-6 relative">
              <button
                onClick={() => setSelectedWorkout(null)}
                className="absolute top-4 right-4 bg-gray-100 rounded-full p-2"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>

              <h2 className="mb-4">{selectedWorkout.workoutName}</h2>

              <div className="bg-teal-50 rounded-2xl p-4 mb-4">
                <div className="flex items-center gap-2 text-teal-700 mb-2">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  <span className="text-sm">
                    {new Date(selectedWorkout.date).toLocaleDateString('ru-RU', {
                      weekday: 'long',
                      day: 'numeric',
                      month: 'long',
                      year: 'numeric',
                    })}
                  </span>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3 mb-4">
                <div className="bg-gray-50 rounded-xl p-3">
                  <div className="text-sm text-gray-600 mb-1">Длительность</div>
                  <div className="text-xl">{selectedWorkout.duration} мин</div>
                </div>

                <div className="bg-gray-50 rounded-xl p-3">
                  <div className="text-sm text-gray-600 mb-1">Упражнений</div>
                  <div className="text-xl">{selectedWorkout.completedExercises}/{selectedWorkout.totalExercises}</div>
                </div>

                {selectedWorkout.calories ? (
                  <div className="bg-gray-50 rounded-xl p-3">
                    <div className="text-sm text-gray-600 mb-1">Калорий</div>
                    <div className="text-xl">{selectedWorkout.calories}</div>
                  </div>
                ) : null}

                {selectedWorkout.rating ? (
                  <div className="bg-gray-50 rounded-xl p-3">
                    <div className="text-sm text-gray-600 mb-1">Оценка</div>
                    <div className="flex items-center gap-1">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <svg
                          key={i}
                          className={`w-5 h-5 ${i < (selectedWorkout.rating || 0) ? 'text-yellow-400' : 'text-gray-300'}`}
                          fill="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                        </svg>
                      ))}
                    </div>
                  </div>
                ) : null}
              </div>

              {selectedWorkout.completed && (
                <button
                  className="w-full bg-teal-600 text-white py-3 rounded-xl disabled:opacity-60"
                  onClick={handleRepeat}
                  disabled={!onStartWorkout}
                >
                  Повторить тренировку
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
