import { useEffect, useMemo, useState } from 'react';
import type { Screen } from '../types';
import { apiRequest } from '../api/client';

interface StatisticsProps {
  onNavigate: (screen: Screen) => void;
}

interface WeeklyStat {
  day: string;
  workouts: number;
  duration: number;
}

interface MonthlyTrend {
  month: string;
  workouts: number;
}

interface CategoryStat {
  category: string;
  percentage: number;
  color?: string;
  color_class?: string;
}

interface RecordStat {
  title: string;
  value: string;
  icon_class?: string;
  bg_class?: string;
}

interface StatisticsResponse {
  weekly: WeeklyStat[];
  monthly: MonthlyTrend[];
  categories: CategoryStat[];
  records: RecordStat[];
  total_workouts: number;
  total_hours: number;
  max_weekly: number;
  max_monthly: number;
  monthly_points: string;
}

export function Statistics({ onNavigate }: StatisticsProps) {
  const [stats, setStats] = useState<StatisticsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiRequest<StatisticsResponse>('/statistics');
      setStats(data);
    } catch (err: any) {
      setError(err?.message || 'Не удалось загрузить статистику');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const insights = useMemo(() => {
    if (!stats) {
      return [] as string[];
    }
    const monthly = stats.monthly || [];
    const weekly = stats.weekly || [];
    const lastMonth = monthly[monthly.length - 1]?.workouts || 0;
    const prevMonth = monthly.length > 1 ? monthly[monthly.length - 2].workouts : 0;
    const delta = lastMonth - prevMonth;
    const topDay = weekly.reduce((best, current) => (current.duration > best.duration ? current : best), weekly[0] || { day: '', duration: 0, workouts: 0 });
    const avgDuration = stats.total_workouts > 0 ? Math.round((stats.total_hours * 60) / stats.total_workouts) : 0;

    const messages: string[] = [];
    if (monthly.length > 1) {
      messages.push(
        delta >= 0
          ? `Ваша активность выросла на ${delta} тренировок по сравнению с прошлым месяцем`
          : `В этом месяце вы сделали на ${Math.abs(delta)} тренировок меньше, чем в прошлом`
      );
    }
    if (topDay.day) {
      messages.push(`Самый активный день недели: ${topDay.day}`);
    }
    if (avgDuration > 0) {
      messages.push(`Средняя продолжительность тренировки: ${avgDuration} минут`);
    }
    return messages;
  }, [stats]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <div className="text-sm text-slate-500">Загрузка...</div>
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="min-h-screen bg-gray-50 p-6">
        <button
          onClick={() => onNavigate('progress')}
          className="flex items-center gap-2 mb-6 text-blue-600"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <div className="bg-red-50 border border-red-200 rounded-xl p-4 text-red-700 text-sm">
          {error || 'Нет данных для отображения'}
        </div>
      </div>
    );
  }

  const maxWeekly = stats.max_weekly || Math.max(1, ...stats.weekly.map((item) => item.duration));
  const maxMonthly = stats.max_monthly || Math.max(1, ...stats.monthly.map((item) => item.workouts));
  const monthlyPoints = stats.monthly_points || stats.monthly
    .map((item, index) => {
      if (stats.monthly.length <= 1) return '';
      const x = (index / (stats.monthly.length - 1)) * 300;
      const y = 100 - ((item.workouts / maxMonthly) * 80);
      return `${Math.round(x)},${Math.round(y)}`;
    })
    .join(' ');

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white p-6">
        <button
          onClick={() => onNavigate('progress')}
          className="flex items-center gap-2 mb-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="text-2xl mb-2">Статистика</h1>
        <p className="text-blue-100">Детальный анализ вашего прогресса</p>
      </div>

      <div className="p-6 space-y-6">
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-xl p-4 text-red-700 text-sm">
            {error}
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div className="bg-gradient-to-br from-blue-500 to-blue-600 text-white rounded-2xl p-5 shadow-lg">
            <div className="text-3xl mb-1">{stats.total_workouts}</div>
            <div className="text-sm text-blue-100">Всего тренировок</div>
          </div>
          <div className="bg-gradient-to-br from-green-500 to-green-600 text-white rounded-2xl p-5 shadow-lg">
            <div className="text-3xl mb-1">{stats.total_hours.toFixed(1)}</div>
            <div className="text-sm text-green-100">Часов активности</div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Активность за неделю</h3>
          <div className="space-y-3">
            {stats.weekly.map((item, index) => (
              <div key={item.day + index}>
                <div className="flex items-center justify-between text-sm mb-1">
                  <span className="text-gray-600">{item.day}</span>
                  <span>{item.duration} мин</span>
                </div>
                <div className="h-8 bg-gray-100 rounded-lg overflow-hidden">
                  <div
                    className="h-full bg-blue-500 rounded-lg transition-all duration-500"
                    style={{
                      width: `${(item.duration / maxWeekly) * 100}%`,
                      animationDelay: `${index * 0.1}s`,
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Динамика по месяцам</h3>
          <div className="relative h-48">
            <div className="absolute inset-0 flex flex-col justify-between">
              {[0, 1, 2, 3, 4].map((i) => (
                <div key={i} className="border-t border-gray-100" />
              ))}
            </div>

            <svg className="absolute inset-0 w-full h-full" viewBox="0 0 300 100" preserveAspectRatio="none">
              <polyline
                points={monthlyPoints}
                fill="none"
                stroke="#3b82f6"
                strokeWidth="2"
                vectorEffect="non-scaling-stroke"
              />
              {stats.monthly.map((item, i) => {
                const x = (i / Math.max(1, stats.monthly.length - 1)) * 300;
                const y = 100 - ((item.workouts / maxMonthly) * 80);
                return (
                  <circle
                    key={item.month + i}
                    cx={x}
                    cy={y}
                    r="4"
                    fill="#3b82f6"
                    vectorEffect="non-scaling-stroke"
                  />
                );
              })}
            </svg>

            <div className="absolute bottom-0 left-0 right-0 flex justify-between text-xs text-gray-500">
              {stats.monthly.map((item, i) => (
                <span key={item.month + i}>{item.month}</span>
              ))}
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Распределение по категориям</h3>
          <div className="space-y-4">
            {stats.categories.map((item, index) => (
              <div key={item.category + index}>
                <div className="flex justify-between text-sm mb-2">
                  <span>{item.category}</span>
                  <span>{item.percentage}%</span>
                </div>
                <div className="h-3 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className={`h-full ${item.color_class || 'bg-blue-500'} rounded-full transition-all duration-500`}
                    style={{
                      width: `${item.percentage}%`,
                      animationDelay: `${index * 0.1}s`,
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-4">Личные рекорды</h3>
          <div className="space-y-3">
            {stats.records.map((record, index) => (
              <div key={record.title + index} className={`flex items-center justify-between p-3 rounded-xl ${record.bg_class || 'bg-gray-50'}`}>
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center ${record.bg_class || 'bg-gray-200'}`}>
                    <svg className={`w-5 h-5 ${record.icon_class || 'text-gray-600'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                    </svg>
                  </div>
                  <div>
                    <div>{record.title}</div>
                    <div className="text-sm text-gray-600">{record.value}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-2xl p-5">
          <h3 className="mb-3">Инсайты</h3>
          {insights.length === 0 ? (
            <p className="text-sm text-gray-600">Накопите больше данных для подробных инсайтов.</p>
          ) : (
            <div className="space-y-3 text-sm">
              {insights.map((text, index) => (
                <div key={index} className="flex items-start gap-2">
                  <svg className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                  </svg>
                  <p className="text-gray-700">{text}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
