import { useEffect, useState } from 'react';
import type { CalendarDay, CalendarWorkout, Screen } from '../types';
import { apiRequest } from '../api/client';

interface CalendarProps {
  onNavigate: (screen: Screen) => void;
}

export function Calendar({ onNavigate }: CalendarProps) {
  const [monthParam, setMonthParam] = useState<string>('');
  const [selectedDate, setSelectedDate] = useState<string>('');
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const load = async (month = monthParam, date = selectedDate) => {
    setLoading(true);
    try {
      const url = `/calendar?month=${encodeURIComponent(month)}&date=${encodeURIComponent(date)}`;
      const res = await apiRequest<any>(url);
      setData(res);
      setMonthParam(res.month_param || month);
      setSelectedDate(date);
    } catch {
      setData(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load('');
  }, []);

  if (loading && !data) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-sm text-gray-500">Загрузка...</div>
      </div>
    );
  }

  const days: CalendarDay[] = data?.days || [];
  const startingDay = data?.starting_day || 0;
  const monthLabel = data?.month_label || '';
  const selectedWorkouts: CalendarWorkout[] = data?.selected_workouts || [];

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-blue-600 text-white p-6 rounded-b-3xl">
        <div className="flex items-center gap-4 mb-4">
          <button onClick={() => onNavigate('home')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Календарь</h1>
        </div>

        <div className="grid grid-cols-3 gap-2">
          <div className="bg-white/20 backdrop-blur rounded-xl p-3 text-center">
            <div className="text-2xl mb-1">{data?.total_workouts || 0}</div>
            <div className="text-xs text-white/80">Тренировок</div>
          </div>
          <div className="bg-white/20 backdrop-blur rounded-xl p-3 text-center">
            <div className="text-2xl mb-1">{data?.current_streak || 0}</div>
            <div className="text-xs text-white/80">Дней подряд</div>
          </div>
          <div className="bg-white/20 backdrop-blur rounded-xl p-3 text-center">
            <div className="text-2xl mb-1">{data?.total_hours || 0}</div>
            <div className="text-xs text-white/80">Часов</div>
          </div>
        </div>
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-white rounded-2xl p-4 shadow-sm">
          <div className="flex items-center justify-between mb-4">
            <button onClick={() => load(data?.prev_month || '')} className="p-2 hover:bg-gray-100 rounded-lg">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <h3 className="capitalize">{monthLabel}</h3>
            <button onClick={() => load(data?.next_month || '')} className="p-2 hover:bg-gray-100 rounded-lg">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>

          <div className="grid grid-cols-7 gap-2 mb-2">
            {['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'].map((day) => (
              <div key={day} className="text-center text-xs text-gray-600 py-2">
                {day}
              </div>
            ))}
          </div>

          <div className="grid grid-cols-7 gap-2">
            {Array.from({ length: startingDay }).map((_, i) => (
              <div key={`empty-${i}`} className="aspect-square" />
            ))}

            {days.map((day) => (
              <button
                key={day.date}
                onClick={() => {
                  setSelectedDate(day.date);
                  load(data?.month_param || '', day.date);
                }}
                className={`aspect-square rounded-xl text-sm flex items-center justify-center relative ${
                  day.isSelected
                    ? 'bg-blue-600 text-white'
                    : day.isToday
                      ? 'bg-blue-100 text-blue-700'
                      : 'bg-gray-50 text-gray-700'
                }`}
              >
                {day.day}
                {day.isWorkout && (
                  <span className={`absolute bottom-1 w-1.5 h-1.5 rounded-full ${day.isSelected ? 'bg-white' : 'bg-green-500'}`} />
                )}
              </button>
            ))}
          </div>
        </div>

        {selectedDate && (
          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <h3 className="mb-3">Тренировки {data?.selected_label}</h3>
            <div className="space-y-3">
              {selectedWorkouts.map((workout) => (
                <div key={workout.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
                  <div>
                    <div className="text-sm">{workout.name}</div>
                    <div className="text-xs text-gray-500">{workout.duration} мин • {workout.exercises} упр.</div>
                  </div>
                  {workout.completed ? (
                    <span className="text-xs text-green-600">Выполнено</span>
                  ) : (
                    <span className="text-xs text-gray-500">Не завершено</span>
                  )}
                </div>
              ))}
              {selectedWorkouts.length === 0 && (
                <div className="text-sm text-gray-500">Нет тренировок на выбранную дату.</div>
              )}
            </div>
          </div>
        )}

        <div className="bg-white rounded-2xl p-4 shadow-sm">
          <h3 className="mb-3">Последние тренировки</h3>
          <div className="space-y-3">
            {(data?.recent_workouts || []).map((workout: CalendarWorkout) => (
              <div key={workout.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
                <div>
                  <div className="text-sm">{workout.name}</div>
                  <div className="text-xs text-gray-500">{workout.date} • {workout.duration} мин</div>
                </div>
                {workout.completed ? (
                  <span className="text-xs text-green-600">Готово</span>
                ) : (
                  <span className="text-xs text-gray-500">В процессе</span>
                )}
              </div>
            ))}
            {(data?.recent_workouts || []).length === 0 && (
              <div className="text-sm text-gray-500">Пока нет тренировок.</div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
