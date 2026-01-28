import { useEffect, useState } from 'react';
import type { Screen } from '../types';
import { apiRequest } from '../api/client';

interface ProgressProps {
  onNavigate?: (screen: Screen) => void;
}

export function Progress({ onNavigate }: ProgressProps = { onNavigate: () => {} }) {
  const [summary, setSummary] = useState({ workouts: 0, hours: 0, streak: 0, achievements: 0 });
  const [weekly, setWeekly] = useState<any[]>([]);
  const [monthly, setMonthly] = useState<any[]>([]);
  const [achievements, setAchievements] = useState<any[]>([]);

  useEffect(() => {
    const load = async () => {
      const progress = await apiRequest<any>('/progress');
      setSummary({
        workouts: progress.workouts || 0,
        hours: progress.hours || 0,
        streak: progress.streak || 0,
        achievements: progress.achievements || 0,
      });

      const stats = await apiRequest<any>('/statistics');
      setWeekly(stats.weekly || []);
      setMonthly(stats.monthly || []);

      const ach = await apiRequest<any>('/achievements');
      setAchievements(ach.achievements || []);
    };

    load();
  }, []);

  const weeklyGoal = 4;
  const completedThisWeek = weekly.reduce((acc, day) => acc + (day.workouts || 0), 0);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-gradient-to-br from-purple-600 to-blue-600 text-white p-6">
        <div className="flex items-center justify-between mb-2">
          <div>
            <h1 className="text-2xl">Ваш прогресс</h1>
            <p className="text-purple-100 text-sm">Следите за достижениями</p>
          </div>
          {onNavigate && (
            <button onClick={() => onNavigate('statistics')} className="p-2 hover:bg-purple-500 rounded-lg">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </button>
          )}
        </div>
      </div>

      <div className="p-6 space-y-6">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="bg-white rounded-2xl p-5 shadow-sm">
            <div className="bg-blue-100 w-12 h-12 rounded-xl flex items-center justify-center mb-3">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div className="text-3xl mb-1">{summary.workouts}</div>
            <div className="text-sm text-gray-600">Всего тренировок</div>
          </div>

          <div className="bg-white rounded-2xl p-5 shadow-sm">
            <div className="bg-green-100 w-12 h-12 rounded-xl flex items-center justify-center mb-3">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="text-3xl mb-1">{Number(summary.hours).toFixed(1)}</div>
            <div className="text-sm text-gray-600">Часов тренировок</div>
          </div>

          <div className="bg-white rounded-2xl p-5 shadow-sm">
            <div className="bg-orange-100 w-12 h-12 rounded-xl flex items-center justify-center mb-3">
              <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
              </svg>
            </div>
            <div className="text-3xl mb-1">{summary.streak}</div>
            <div className="text-sm text-gray-600">Дней подряд</div>
          </div>

          <div className="bg-white rounded-2xl p-5 shadow-sm">
            <div className="bg-purple-100 w-12 h-12 rounded-xl flex items-center justify-center mb-3">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
              </svg>
            </div>
            <div className="text-3xl mb-1">{summary.achievements}</div>
            <div className="text-sm text-gray-600">Достижений</div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h3>Недельная цель</h3>
            </div>
            <span className="text-sm text-gray-600">
              {completedThisWeek}/{weeklyGoal}
            </span>
          </div>

          <div className="bg-gray-200 rounded-full h-3 overflow-hidden mb-4">
            <div
              className="bg-gradient-to-r from-blue-600 to-purple-600 h-full transition-all"
              style={{ width: `${Math.min((completedThisWeek / weeklyGoal) * 100, 100)}%` }}
            />
          </div>

          <div className="grid grid-cols-7 gap-2">
            {weekly.map((day, index) => (
              <div key={index} className="text-center">
                <div className={`w-full aspect-square rounded-xl mb-1 flex items-center justify-center text-xs ${
                  day.workouts > 0
                    ? 'bg-green-500 text-white'
                    : 'bg-gray-100 text-gray-400'
                }`}>
                  {day.day}
                </div>
                {day.workouts > 0 && (
                  <div className="text-xs text-gray-600">{day.duration}м</div>
                )}
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3>По месяцам</h3>
          </div>

          <div className="space-y-4">
            {monthly.map((stat, index) => (
              <div key={index}>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-600">{stat.month}</span>
                  <div className="text-sm">
                    <span className="text-blue-600">{stat.workouts} тренировок</span>
                  </div>
                </div>
                <div className="bg-gray-200 rounded-full h-2 overflow-hidden">
                  <div
                    className="bg-gradient-to-r from-blue-500 to-blue-600 h-full"
                    style={{ width: `${Math.min((stat.workouts / 20) * 100, 100)}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
            </svg>
            <h3>Достижения</h3>
          </div>

          <div className="space-y-3">
            {achievements.map((achievement: any) => (
              <div
                key={achievement.id}
                className={`p-4 rounded-xl border-2 ${
                  achievement.unlocked ? 'border-purple-200 bg-purple-50' : 'border-gray-200 bg-gray-50'
                }`}
              >
                <div className="flex items-start gap-3">
                  <div className={`text-3xl ${!achievement.unlocked && 'grayscale opacity-50'}`}>
                    {achievement.icon}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <h4 className={achievement.unlocked ? 'text-purple-900' : 'text-gray-600'}>
                        {achievement.title}
                      </h4>
                      {achievement.unlocked && (
                        <svg className="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      )}
                    </div>
                    <p className="text-sm text-gray-600 mb-1">{achievement.description}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
