import { useEffect, useState } from 'react';
import type { Screen } from '../types';
import { apiRequest } from '../api/client';
import { useAuth } from '../context/AuthContext';

interface CommunityProps {
  onNavigate: (screen: Screen) => void;
}

export function Community({ onNavigate }: CommunityProps) {
  const { user } = useAuth();
  const [selectedTab, setSelectedTab] = useState<'leaderboard' | 'departments' | 'challenges'>('leaderboard');
  const [leaderboard, setLeaderboard] = useState<any[]>([]);
  const [departments, setDepartments] = useState<any[]>([]);
  const [challenges, setChallenges] = useState<any[]>([]);

  useEffect(() => {
    const load = async () => {
      const lb = await apiRequest<any>('/community/leaderboard');
      setLeaderboard(lb.leaderboard || []);
      const dep = await apiRequest<any>('/community/departments');
      setDepartments(dep.departments || []);
      const ch = await apiRequest<any>('/community/challenges');
      setChallenges(ch.challenges || []);
    };
    load();
  }, []);

  const currentUser = leaderboard.find((entry) => entry.name === user?.name) || leaderboard[0];
  const currentRank = leaderboard.findIndex((entry) => entry.name === user?.name) + 1;

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-gradient-to-br from-orange-600 to-red-600 text-white p-6">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('home')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Сообщество</h1>
        </div>
        <p className="text-orange-100 text-sm">Соревнуйтесь с коллегами</p>
      </div>

      {currentUser && (
        <div className="px-4 -mt-4 mb-4">
          <div className="bg-white rounded-2xl p-4 shadow-lg">
            <div className="flex items-center gap-4 mb-3">
              <div className="text-4xl">👤</div>
              <div className="flex-1">
                <h3 className="mb-1">{currentUser.name}</h3>
                <p className="text-sm text-gray-600">{currentUser.department}</p>
              </div>
              <div className="text-center">
                <div className="bg-orange-100 text-orange-600 px-3 py-1 rounded-full mb-1">
                  #{currentRank || 1}
                </div>
                <div className="text-xs text-gray-600">место</div>
              </div>
            </div>
            <div className="grid grid-cols-4 gap-3 pt-3 border-t border-gray-100">
              <div className="text-center">
                <div className="text-lg mb-1">{currentUser.workouts}</div>
                <div className="text-xs text-gray-600">тренировок</div>
              </div>
              <div className="text-center">
                <div className="text-lg mb-1">{Number(currentUser.hours).toFixed(1)}ч</div>
                <div className="text-xs text-gray-600">времени</div>
              </div>
              <div className="text-center">
                <div className="text-lg mb-1">{currentUser.streak || 0}</div>
                <div className="text-xs text-gray-600">дней подряд</div>
              </div>
              <div className="text-center">
                <div className="text-lg mb-1">{currentUser.points}</div>
                <div className="text-xs text-gray-600">баллов</div>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="px-4 mb-4">
        <div className="flex gap-2 bg-white rounded-2xl p-1 shadow-sm">
          <button
            onClick={() => setSelectedTab('leaderboard')}
            className={`flex-1 py-2 rounded-xl text-sm transition-colors ${
              selectedTab === 'leaderboard' ? 'bg-orange-600 text-white' : 'text-gray-700'
            }`}
          >
            Рейтинг
          </button>
          <button
            onClick={() => setSelectedTab('departments')}
            className={`flex-1 py-2 rounded-xl text-sm transition-colors ${
              selectedTab === 'departments' ? 'bg-orange-600 text-white' : 'text-gray-700'
            }`}
          >
            Отделы
          </button>
          <button
            onClick={() => setSelectedTab('challenges')}
            className={`flex-1 py-2 rounded-xl text-sm transition-colors ${
              selectedTab === 'challenges' ? 'bg-orange-600 text-white' : 'text-gray-700'
            }`}
          >
            Вызовы
          </button>
        </div>
      </div>

      <div className="px-4 space-y-4">
        {selectedTab === 'leaderboard' && (
          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <h3 className="mb-3">Топ сотрудников</h3>
            <div className="space-y-3">
              {leaderboard.map((entry: any, index: number) => (
                <div key={entry.id} className="flex items-center gap-3 p-3 rounded-xl bg-gray-50">
                  <div className="text-xs text-gray-500">#{index + 1}</div>
                  <div className="text-2xl">👤</div>
                  <div className="flex-1">
                    <div className="text-sm">{entry.name}</div>
                    <div className="text-xs text-gray-500">{entry.department}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm">{entry.points} баллов</div>
                    <div className="text-xs text-gray-500">{entry.workouts} тренировок</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {selectedTab === 'departments' && (
          <div className="bg-white rounded-2xl p-4 shadow-sm">
            <h3 className="mb-3">Статистика по отделам</h3>
            <div className="space-y-3">
              {departments.map((department: any) => (
                <div key={department.name} className="flex items-center gap-3 p-3 bg-gray-50 rounded-xl">
                  <div className={`w-3 h-3 rounded-full ${department.color_class}`}></div>
                  <div className="flex-1">
                    <div className="text-sm">{department.name}</div>
                    <div className="text-xs text-gray-500">{department.members} сотрудников</div>
                  </div>
                  <div className="text-sm text-gray-600">{department.avg_workouts} трен./мес</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {selectedTab === 'challenges' && (
          <div className="space-y-3">
            {challenges.map((challenge: any) => (
              <div key={challenge.id} className="bg-white rounded-2xl p-4 shadow-sm">
                <div className="flex items-start gap-3">
                  <div className="text-3xl">{challenge.icon}</div>
                  <div className="flex-1">
                    <h3 className="mb-1">{challenge.title}</h3>
                    <p className="text-sm text-gray-600 mb-2">{challenge.description}</p>
                    <div className="text-xs text-gray-500 mb-2">Участников: {challenge.participants} • Осталось дней: {challenge.days_left}</div>
                    <div className="bg-gray-200 rounded-full h-2 overflow-hidden">
                      <div className="bg-orange-500 h-full" style={{ width: `${(challenge.progress / challenge.total) * 100}%` }}></div>
                    </div>
                    <div className="text-xs text-gray-500 mt-2">Награда: {challenge.reward}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
