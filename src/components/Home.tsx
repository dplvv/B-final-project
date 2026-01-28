import { useEffect, useState } from 'react';
import type { Screen, Workout } from '../types';
import { apiRequest } from '../api/client';
import { useAuth } from '../context/AuthContext';

interface HomeProps {
  onNavigate: (screen: Screen) => void;
}

export function Home({ onNavigate }: HomeProps) {
  const { user, profile } = useAuth();
  const [stats, setStats] = useState({ workouts: 0, hours: 0, achievements: 0 });
  const [recommended, setRecommended] = useState<Workout | null>(null);

  const needsAssessment = !profile?.fitnessLevel;

  useEffect(() => {
    const load = async () => {
      try {
        const progress = await apiRequest<any>('/progress');
        setStats({
          workouts: progress.workouts || 0,
          hours: Number(progress.hours || 0),
          achievements: progress.achievements || 0,
        });
      } catch {
        setStats({ workouts: 0, hours: 0, achievements: 0 });
      }
    };

    const loadProgram = async () => {
      try {
        const data = await apiRequest<any>('/program');
        const workouts: Workout[] = (data.workouts || []).map((w: any) => ({
          id: w.id,
          name: w.name,
          description: w.description,
          duration: w.duration,
          difficulty: w.difficulty,
          category: w.category,
          exercisesCount: w.exercises_count,
          completed: w.completed,
        }));
        const next = workouts.find((w) => !w.completed) || workouts[0] || null;
        setRecommended(next || null);
      } catch {
        setRecommended(null);
      }
    };

    load();
    loadProgram();
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-50 to-white">
      <div className="bg-blue-600 text-white p-6 rounded-b-3xl">
        <div className="flex items-center gap-2 mb-2">
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <span className="text-sm opacity-90">ГК РОСАТОМ</span>
        </div>
        <h1 className="text-2xl mb-1">Привет, {user?.name?.split(' ')[0] || 'коллега'}!</h1>
        <p className="text-blue-100 text-sm">Готовы к тренировке?</p>
      </div>

      <div className="p-6 space-y-4">
        {needsAssessment && (
          <div className="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex items-start gap-3">
            <svg className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div className="flex-1">
              <h3 className="text-amber-900 mb-1">Пройдите оценку</h3>
              <p className="text-sm text-amber-700 mb-3">
                Для составления персональной программы реабилитации необходимо пройти первичный опросник
              </p>
              <button
                onClick={() => onNavigate('questionnaire')}
                className="bg-amber-600 text-white px-4 py-2 rounded-lg text-sm flex items-center gap-2"
              >
                Начать опросник
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="bg-blue-100 w-10 h-10 rounded-full flex items-center justify-center mb-2">
              <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div className="text-2xl mb-1">{stats.workouts}</div>
            <div className="text-xs text-gray-600">Тренировок</div>
          </div>

          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="bg-green-100 w-10 h-10 rounded-full flex items-center justify-center mb-2">
              <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="text-2xl mb-1">{Number(stats.hours).toFixed(1)}</div>
            <div className="text-xs text-gray-600">Часов</div>
          </div>

          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <div className="bg-purple-100 w-10 h-10 rounded-full flex items-center justify-center mb-2">
              <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
              </svg>
            </div>
            <div className="text-2xl mb-1">{stats.achievements}</div>
            <div className="text-xs text-gray-600">Достижения</div>
          </div>
        </div>

        {!needsAssessment && recommended && (
          <div className="bg-gradient-to-br from-blue-600 to-blue-700 rounded-2xl p-6 text-white shadow-lg">
            <div className="flex items-center gap-2 mb-3">
              <div className="bg-white/20 px-3 py-1 rounded-full text-xs">Сегодня</div>
            </div>
            <h3 className="text-xl mb-2">{recommended.name}</h3>
            <p className="text-blue-100 text-sm mb-4">{recommended.description}</p>
            <div className="flex items-center gap-4 text-sm mb-4">
              <div className="flex items-center gap-1">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>{recommended.duration} мин</span>
              </div>
              <div className="flex items-center gap-1">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
                <span>{recommended.difficulty}</span>
              </div>
            </div>
            <button
              onClick={() => onNavigate('program')}
              className="bg-white text-blue-600 px-6 py-3 rounded-xl w-full flex items-center justify-center gap-2"
            >
              Перейти к программе
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        )}

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Быстрые ссылки</h3>
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <button
              onClick={() => onNavigate('exerciseLibrary')}
              className="p-3 bg-blue-50 rounded-xl hover:bg-blue-100 transition-colors"
            >
              <svg className="w-6 h-6 text-blue-600 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
              <div className="text-xs">Упражнения</div>
            </button>
            <button
              onClick={() => onNavigate('goals')}
              className="p-3 bg-purple-50 rounded-xl hover:bg-purple-100 transition-colors"
            >
              <svg className="w-6 h-6 text-purple-600 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
              </svg>
              <div className="text-xs">Цели</div>
            </button>
            <button
              onClick={() => onNavigate('videoTutorials')}
              className="p-3 bg-red-50 rounded-xl hover:bg-red-100 transition-colors"
            >
              <svg className="w-6 h-6 text-red-600 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div className="text-xs">Видео</div>
            </button>
            <button
              onClick={() => onNavigate('nutrition')}
              className="p-3 bg-green-50 rounded-xl hover:bg-green-100 transition-colors"
            >
              <svg className="w-6 h-6 text-green-600 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              <div className="text-xs">Питание</div>
            </button>
            <button
              onClick={() => onNavigate('calendar')}
              className="p-3 bg-yellow-50 rounded-xl hover:bg-yellow-100 transition-colors"
            >
              <svg className="w-6 h-6 text-yellow-600 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <div className="text-xs">Календарь</div>
            </button>
            <button
              onClick={() => onNavigate('achievements')}
              className="p-3 bg-orange-50 rounded-xl hover:bg-orange-100 transition-colors"
            >
              <svg className="w-6 h-6 text-orange-600 mx-auto mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
              </svg>
              <div className="text-xs">Награды</div>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
