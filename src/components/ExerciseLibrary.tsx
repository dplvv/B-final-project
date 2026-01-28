import { useEffect, useMemo, useState } from 'react';
import type { Exercise, Screen } from '../types';
import { apiRequest } from '../api/client';

interface ExerciseLibraryProps {
  onNavigate: (screen: Screen) => void;
  onExerciseSelect: (exerciseId: string) => void;
}

export function ExerciseLibrary({ onNavigate, onExerciseSelect }: ExerciseLibraryProps) {
  const [selectedCategory, setSelectedCategory] = useState('Все');
  const [selectedDifficulty, setSelectedDifficulty] = useState('Все');
  const [searchQuery, setSearchQuery] = useState('');
  const [exercises, setExercises] = useState<Exercise[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = async () => {
      try {
        const data = await apiRequest<any>(`/exercises?q=${encodeURIComponent(searchQuery)}&category=${encodeURIComponent(selectedCategory === 'Все' ? '' : selectedCategory)}&difficulty=${encodeURIComponent(selectedDifficulty === 'Все' ? '' : selectedDifficulty)}`);
        const items = (data.exercises || []).map((ex: any) => ({
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
        }));
        setExercises(items);
      } catch {
        setExercises([]);
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [searchQuery, selectedCategory, selectedDifficulty]);

  const categories = useMemo(() => {
    const uniq = new Set<string>();
    exercises.forEach((ex) => {
      if (ex.category) uniq.add(ex.category);
    });
    return ['Все', ...Array.from(uniq)];
  }, [exercises]);

  const difficulties = useMemo(() => {
    const uniq = new Set<string>();
    exercises.forEach((ex) => {
      if (ex.difficulty) uniq.add(ex.difficulty);
    });
    return ['Все', ...Array.from(uniq)];
  }, [exercises]);

  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty.toLowerCase()) {
      case 'начальный':
      case 'легкая':
        return 'bg-green-100 text-green-700';
      case 'средний':
      case 'средняя':
        return 'bg-yellow-100 text-yellow-700';
      case 'продвинутый':
      case 'сложная':
        return 'bg-red-100 text-red-700';
      default:
        return 'bg-gray-100 text-gray-700';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-blue-600 text-white p-6 rounded-b-3xl">
        <button onClick={() => onNavigate('home')} className="flex items-center gap-2 mb-4">
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="mb-4">Библиотека упражнений</h1>

        <div className="relative">
          <svg className="w-5 h-5 absolute left-3 top-1/2 -translate-y-1/2 text-blue-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Поиск упражнений..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-3 rounded-xl bg-blue-700/50 placeholder-blue-300 text-white outline-none"
          />
        </div>
      </div>

      <div className="p-4 space-y-4">
        <div>
          <h3 className="text-sm text-gray-600 mb-2">Категория</h3>
          <div className="flex gap-2 overflow-x-auto pb-2">
            {categories.map((category) => (
              <button
                key={category}
                onClick={() => setSelectedCategory(category)}
                className={`px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                  selectedCategory === category
                    ? 'bg-blue-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-200'
                }`}
              >
                {category}
              </button>
            ))}
          </div>
        </div>

        <div>
          <h3 className="text-sm text-gray-600 mb-2">Сложность</h3>
          <div className="flex gap-2 overflow-x-auto pb-2">
            {difficulties.map((difficulty) => (
              <button
                key={difficulty}
                onClick={() => setSelectedDifficulty(difficulty)}
                className={`px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                  selectedDifficulty === difficulty
                    ? 'bg-blue-600 text-white'
                    : 'bg-white text-gray-700 border border-gray-200'
                }`}
              >
                {difficulty}
              </button>
            ))}
          </div>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Загрузка упражнений...</div>
        ) : (
          <div className="space-y-3">
            {exercises.map((exercise) => (
              <div
                key={exercise.id}
                className="bg-white rounded-2xl p-4 shadow-sm hover:shadow-md transition-shadow cursor-pointer"
                onClick={() => onExerciseSelect(exercise.id)}
              >
                <div className="flex items-start gap-4">
                  <div className="bg-blue-100 rounded-xl w-16 h-16 flex items-center justify-center text-2xl flex-shrink-0">
                    🏋️
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center justify-between mb-1">
                      <h3>{exercise.name}</h3>
                      <span className={`text-xs px-2 py-1 rounded-full ${getDifficultyColor(exercise.difficulty || '')}`}>
                        {exercise.difficulty}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 mb-2">{exercise.description}</p>
                    <div className="flex flex-wrap gap-2 text-xs text-gray-500">
                      <span>{exercise.sets} подхода</span>
                      <span>• {exercise.reps}</span>
                      <span>• отдых {exercise.rest}с</span>
                      {exercise.category && <span>• {exercise.category}</span>}
                    </div>
                  </div>
                </div>
              </div>
            ))}
            {exercises.length === 0 && (
              <div className="text-sm text-gray-500">Ничего не найдено.</div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
