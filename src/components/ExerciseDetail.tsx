import { useEffect, useState } from 'react';
import type { Exercise, Screen } from '../types';
import { apiRequest } from '../api/client';

interface ExerciseDetailProps {
  exerciseId: string;
  onNavigate: (screen: Screen) => void;
}

export function ExerciseDetail({ exerciseId, onNavigate }: ExerciseDetailProps) {
  const [exercise, setExercise] = useState<Exercise | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiRequest<any>(`/exercises/${exerciseId}`);
        const ex = data.exercise;
        setExercise({
          id: ex.id,
          name: ex.name,
          sets: ex.sets,
          reps: ex.reps,
          duration: ex.duration,
          rest: ex.rest,
          description: ex.description,
          videoUrl: ex.video_url,
          category: ex.category,
          difficulty: ex.difficulty,
          muscleGroups: ex.muscle_groups || [],
          equipment: ex.equipment || [],
        });
      } catch {
        setExercise(null);
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [exerciseId]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-sm text-gray-500">Загрузка...</div>
      </div>
    );
  }

  if (!exercise) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-blue-600 text-white p-6 rounded-b-3xl">
        <button onClick={() => onNavigate('exerciseLibrary')} className="flex items-center gap-2 mb-4">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="mb-2">{exercise.name}</h1>
        <p className="text-blue-100 text-sm">{exercise.category}</p>
      </div>

      <div className="p-6 space-y-4">
        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Описание</h3>
          <p className="text-sm text-gray-600">{exercise.description}</p>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Параметры</h3>
          <div className="grid grid-cols-2 gap-3">
            <div className="bg-gray-50 rounded-xl p-3">
              <div className="text-xs text-gray-500">Подходы</div>
              <div className="text-lg">{exercise.sets}</div>
            </div>
            <div className="bg-gray-50 rounded-xl p-3">
              <div className="text-xs text-gray-500">Повторения</div>
              <div className="text-lg">{exercise.reps}</div>
            </div>
            <div className="bg-gray-50 rounded-xl p-3">
              <div className="text-xs text-gray-500">Отдых</div>
              <div className="text-lg">{exercise.rest}с</div>
            </div>
            <div className="bg-gray-50 rounded-xl p-3">
              <div className="text-xs text-gray-500">Сложность</div>
              <div className="text-lg">{exercise.difficulty}</div>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Задействованные мышцы</h3>
          <div className="flex flex-wrap gap-2">
            {exercise.muscleGroups?.map((muscle) => (
              <span key={muscle} className="px-3 py-1 bg-blue-50 text-blue-700 rounded-full text-xs">
                {muscle}
              </span>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Оборудование</h3>
          <div className="flex flex-wrap gap-2">
            {exercise.equipment?.map((item) => (
              <span key={item} className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-xs">
                {item}
              </span>
            ))}
          </div>
        </div>

        {exercise.videoUrl && (
          <div className="bg-white rounded-2xl p-5 shadow-sm">
            <h3 className="mb-3">Видео</h3>
            <a
              className="text-blue-600 text-sm underline"
              href={exercise.videoUrl}
              target="_blank"
              rel="noreferrer"
            >
              Открыть видео
            </a>
          </div>
        )}
      </div>
    </div>
  );
}
