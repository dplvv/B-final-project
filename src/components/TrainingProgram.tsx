import { useEffect, useState } from 'react';
import type { Workout } from '../types';
import { apiRequest } from '../api/client';

interface TrainingProgramProps {
  onStartWorkout: (workoutId: string) => void;
}

export function TrainingProgram({ onStartWorkout }: TrainingProgramProps) {
  const [workouts, setWorkouts] = useState<Workout[]>([]);
  const [completedCount, setCompletedCount] = useState(0);
  const [totalDuration, setTotalDuration] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiRequest<any>('/program');
        const items = (data.workouts || []).map((w: any) => ({
          id: w.id,
          name: w.name,
          description: w.description,
          duration: w.duration,
          difficulty: w.difficulty,
          category: w.category,
          exercisesCount: w.exercises_count,
          completed: w.completed,
          recommendedDate: w.recommended_date,
        }));
        setWorkouts(items);
        setCompletedCount(data.completed_count || 0);
        setTotalDuration(data.total_duration || 0);
      } catch {
        setWorkouts([]);
        setCompletedCount(0);
        setTotalDuration(0);
      } finally {
        setLoading(false);
      }
    };

    load();
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white p-6">
        <h1 className="text-2xl mb-2">Программа тренировок</h1>
        <p className="text-blue-100 text-sm">Персональная программа реабилитации</p>
      </div>

      <div className="bg-white border-b p-6">
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-green-50 rounded-xl p-4">
            <div className="flex items-center gap-2 text-green-700 mb-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm">Выполнено</span>
            </div>
            <div className="text-2xl text-green-900">{completedCount}</div>
          </div>

          <div className="bg-purple-50 rounded-xl p-4">
            <div className="flex items-center gap-2 text-purple-700 mb-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm">Общее время</span>
            </div>
            <div className="text-2xl text-purple-900">{totalDuration} мин</div>
          </div>
        </div>
      </div>

      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg">Недельный план</h2>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Загрузка программы...</div>
        ) : (
          <div className="space-y-3">
            {workouts.map((workout, index) => (
              <div key={workout.id} className="bg-white rounded-2xl shadow-sm overflow-hidden">
                <div className="p-5">
                  <div className="flex items-start gap-4">
                    <div
                      className={`w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0 ${
                        workout.completed ? 'bg-green-100 text-green-600' : 'bg-blue-100 text-blue-600'
                      }`}
                    >
                      {workout.completed ? (
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      ) : (
                        <span className="text-lg">{index + 1}</span>
                      )}
                    </div>

                    <div className="flex-1">
                      <h3 className="mb-2">{workout.name}</h3>

                      <div className="flex items-center gap-4 text-sm text-gray-600 mb-3">
                        <div className="flex items-center gap-1">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                          <span>{workout.duration} мин</span>
                        </div>
                        <div className="flex items-center gap-1">
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                          </svg>
                          <span>{workout.difficulty}</span>
                        </div>
                      </div>

                      <div className="text-sm text-gray-600 mb-3">
                        {workout.exercisesCount} упражнений
                      </div>

                      <button
                        onClick={() => onStartWorkout(workout.id)}
                        className={`px-5 py-2.5 rounded-lg flex items-center gap-2 transition-colors ${
                          workout.completed
                            ? 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                            : 'bg-blue-600 text-white hover:bg-blue-700'
                        }`}
                      >
                        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                          <path d="M8 5v14l11-7z" />
                        </svg>
                        <span>{workout.completed ? 'Повторить' : 'Начать'}</span>
                      </button>
                    </div>
                  </div>
                </div>

                {!workout.completed && workout.recommendedDate && (
                  <div className="bg-blue-50 px-5 py-3 text-sm text-blue-700">
                    Рекомендуемая дата: {new Date(workout.recommendedDate).toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' })}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
