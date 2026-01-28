import { useEffect, useState } from 'react';
import type { Achievement, Screen } from '../types';
import { apiRequest } from '../api/client';

interface AchievementsProps {
  onNavigate: (screen: Screen) => void;
}

export function Achievements({ onNavigate }: AchievementsProps) {
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiRequest<any>('/achievements');
        const items = (data.achievements || []).map((item: any) => ({
          id: item.id,
          title: item.title,
          description: item.description,
          icon: item.icon,
          unlocked: item.unlocked,
          unlockedDate: item.unlocked_date,
          progress: item.progress,
          total: item.total,
        }));
        setAchievements(items);
      } catch {
        setAchievements([]);
      } finally {
        setLoading(false);
      }
    };

    load();
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-gradient-to-br from-yellow-500 to-orange-500 text-white p-6 rounded-b-3xl">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('home')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Достижения</h1>
        </div>
        <p className="text-orange-100 text-sm">Ваши награды и успехи</p>
      </div>

      <div className="p-4">
        {loading ? (
          <div className="text-sm text-gray-500">Загрузка...</div>
        ) : (
          <div className="space-y-3">
            {achievements.map((achievement) => (
              <div
                key={achievement.id}
                className={`bg-white rounded-2xl p-4 shadow-sm border-2 ${
                  achievement.unlocked ? 'border-yellow-200' : 'border-gray-200 opacity-70'
                }`}
              >
                <div className="flex items-start gap-4">
                  <div className="text-4xl">{achievement.icon}</div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between mb-1">
                      <h3>{achievement.title}</h3>
                      {achievement.unlocked && (
                        <span className="text-xs text-yellow-600">Получено</span>
                      )}
                    </div>
                    <p className="text-sm text-gray-600 mb-2">{achievement.description}</p>
                    {achievement.total && (
                      <div className="text-xs text-gray-500">
                        Прогресс: {achievement.progress}/{achievement.total}
                      </div>
                    )}
                    {achievement.unlockedDate && (
                      <div className="text-xs text-gray-400 mt-1">
                        Дата: {achievement.unlockedDate}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
            {achievements.length === 0 && (
              <div className="text-sm text-gray-500">Нет достижений.</div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
