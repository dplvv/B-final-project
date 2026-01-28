import { useEffect, useState } from 'react';
import type { Screen } from '../types';
import { useAuth } from '../context/AuthContext';
import { apiRequest } from '../api/client';

interface ProfileProps {
  onNavigate?: (screen: Screen) => void;
}

export function Profile({ onNavigate }: ProfileProps) {
  const { user, profile, logout } = useAuth();
  const [stats, setStats] = useState({ workouts: 0, achievements: 0, streak: 0 });

  useEffect(() => {
    const load = async () => {
      const progress = await apiRequest<any>('/progress');
      setStats({
        workouts: progress.workouts || 0,
        achievements: progress.achievements || 0,
        streak: progress.streak || 0,
      });
    };
    load();
  }, []);

  const menuSections = [
    {
      title: 'Личная информация',
      items: [
        { label: 'Профиль', value: user?.name, screen: null },
        { label: 'Медицинская информация', value: 'Просмотреть', screen: 'medicalInfo' as Screen },
        { label: 'Уровень подготовки', value: profile?.fitnessLevel ? (profile.fitnessLevel === 'beginner' ? 'Начальный' : profile.fitnessLevel === 'intermediate' ? 'Средний' : 'Продвинутый') : 'Не определен', screen: null },
        { label: 'Цели', value: profile?.goals?.length ? `${profile.goals.length} активных` : 'Не установлены', screen: 'goals' as Screen }
      ]
    },
    {
      title: 'Настройки',
      items: [
        { label: 'Уведомления', value: 'Просмотреть', screen: 'notifications' as Screen },
        { label: 'Настройки приложения', value: null, screen: 'settings' as Screen },
        { label: 'Помощь и поддержка', value: null, screen: 'support' as Screen }
      ]
    }
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-gradient-to-br from-blue-600 to-purple-600 text-white p-6">
        <div className="text-center">
          <div className="bg-white/20 w-24 h-24 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <h2 className="text-2xl mb-1">{user?.name}</h2>
          <p className="text-blue-100 text-sm">{user?.department}</p>
          <div className="mt-4 inline-block bg-white/20 px-4 py-1.5 rounded-full text-sm">
            {profile?.age ? `${profile.age} лет` : 'Возраст не указан'}
          </div>
        </div>
      </div>

      <div className="p-6 space-y-6">
        <div className="grid grid-cols-3 gap-3">
          <div className="bg-white rounded-xl p-4 text-center shadow-sm">
            <div className="bg-blue-100 w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-2">
              <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div className="text-xl mb-1">{stats.workouts}</div>
            <div className="text-xs text-gray-600">Тренировок</div>
          </div>

          <div className="bg-white rounded-xl p-4 text-center shadow-sm">
            <div className="bg-green-100 w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-2">
              <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
              </svg>
            </div>
            <div className="text-xl mb-1">{stats.achievements}</div>
            <div className="text-xs text-gray-600">Достижения</div>
          </div>

          <div className="bg-white rounded-xl p-4 text-center shadow-sm">
            <div className="bg-purple-100 w-10 h-10 rounded-full flex items-center justify-center mx-auto mb-2">
              <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="text-xl mb-1">{stats.streak}</div>
            <div className="text-xs text-gray-600">Дней подряд</div>
          </div>
        </div>

        {profile?.restrictions?.length ? (
          <div className="bg-amber-50 border border-amber-200 rounded-2xl p-5">
            <h3 className="text-amber-900 mb-3">Учитываемые ограничения</h3>
            <div className="flex flex-wrap gap-2">
              {profile.restrictions.map((restriction, index) => (
                <div key={index} className="bg-white px-3 py-1.5 rounded-lg text-sm text-amber-800 border border-amber-200">
                  {restriction === 'back' && '🔹 Проблемы со спиной'}
                  {restriction === 'joints' && '🔹 Проблемы с суставами'}
                  {restriction === 'cardio' && '🔹 Сердечно-сосудистые'}
                  {restriction === 'none' && '✓ Нет ограничений'}
                  {!['back', 'joints', 'cardio', 'none'].includes(restriction) && restriction}
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {profile?.goals?.length ? (
          <div className="bg-blue-50 border border-blue-200 rounded-2xl p-5">
            <h3 className="text-blue-900 mb-3">Ваши цели</h3>
            <div className="space-y-2">
              {profile.goals.map((goal, index) => (
                <div key={index} className="bg-white px-4 py-3 rounded-xl text-sm text-blue-800 flex items-center gap-2">
                  <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
                  {goal === 'rehab' && 'Реабилитация'}
                  {goal === 'strength' && 'Увеличение силы'}
                  {goal === 'flexibility' && 'Улучшение гибкости'}
                  {goal === 'endurance' && 'Выносливость'}
                  {goal === 'posture' && 'Коррекция осанки'}
                  {!['rehab', 'strength', 'flexibility', 'endurance', 'posture'].includes(goal) && goal}
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {menuSections.map((section, sectionIndex) => (
          <div key={sectionIndex} className="bg-white rounded-2xl shadow-sm overflow-hidden">
            <div className="px-5 py-3 border-b bg-gray-50">
              <h3 className="text-sm text-gray-600">{section.title}</h3>
            </div>
            <div className="divide-y">
              {section.items.map((item, itemIndex) => (
                <button
                  key={itemIndex}
                  className="w-full px-5 py-4 flex items-center gap-4 hover:bg-gray-50 transition-colors"
                  onClick={() => item.screen && onNavigate && onNavigate(item.screen)}
                >
                  <div className="bg-gray-100 p-2 rounded-lg">
                    <svg className="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div className="flex-1 text-left">
                    <div className="text-sm mb-0.5">{item.label}</div>
                    {item.value && <div className="text-xs text-gray-500">{item.value}</div>}
                  </div>
                  <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              ))}
            </div>
          </div>
        ))}

        <button
          className="w-full bg-white border-2 border-red-200 text-red-600 py-4 rounded-xl flex items-center justify-center gap-2 hover:bg-red-50 transition-colors"
          onClick={() => logout()}
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
          Выйти из аккаунта
        </button>
      </div>
    </div>
  );
}
