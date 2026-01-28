import { useEffect, useState } from 'react';
import { apiRequest } from '../api/client';

interface WorkoutSessionProps {
  sessionId: string;
  onComplete: () => void;
}

type SessionState = 'ready' | 'exercise' | 'rest' | 'completed';

export function WorkoutSession({ sessionId, onComplete }: WorkoutSessionProps) {
  const [state, setState] = useState<SessionState>('ready');
  const [timer, setTimer] = useState(0);
  const [isPaused, setIsPaused] = useState(false);
  const [session, setSession] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const loadSession = async () => {
    try {
      const data = await apiRequest<any>(`/workout-sessions/${sessionId}`);
      if (data.status === 'completed') {
        setState('completed');
        setSession(data);
        return data;
      }
      setSession(data);
      return data;
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadSession();
  }, [sessionId]);

  useEffect(() => {
    if (state === 'rest' && !isPaused) {
      const interval = setInterval(() => {
        setTimer((prev) => {
          if (prev <= 1) {
            handleNextSet();
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [state, isPaused]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-white">
        <div className="text-sm text-gray-500">Загрузка тренировки...</div>
      </div>
    );
  }

  if (!session) {
    return null;
  }

  if (state === 'completed') {
    return (
      <div className="min-h-screen bg-gradient-to-b from-green-50 to-white flex items-center justify-center p-6">
        <div className="text-center">
          <div className="bg-green-100 w-24 h-24 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-12 h-12 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h2 className="text-2xl mb-3">Тренировка завершена!</h2>
          <p className="text-gray-600 mb-8">Отличная работа!</p>

          <button onClick={onComplete} className="bg-blue-600 text-white px-8 py-4 rounded-xl">
            Завершить
          </button>
        </div>
      </div>
    );
  }

  const currentExercise = session.current_exercise;

  const handleStart = () => {
    setState('exercise');
  };

  const handleCompleteSet = async () => {
    await apiRequest(`/workout-sessions/${sessionId}/complete-set`, { method: 'POST' });
    const updated = await loadSession();
    if (updated?.status === 'completed') {
      setState('completed');
      return;
    }

    const updatedExercise = updated?.current_exercise || currentExercise;
    const isLastSet = (updated?.current_set || 0) >= updatedExercise.sets;
    if (isLastSet) {
      setState('ready');
    } else {
      setTimer(updatedExercise.rest);
      setState('rest');
    }
  };

  const handleNextSet = () => {
    setState('exercise');
    setTimer(0);
  };

  const handleSkipRest = () => {
    handleNextSet();
  };

  const togglePause = () => {
    setIsPaused(!isPaused);
  };

  return (
    <div className="min-h-screen bg-white">
      <div className="bg-blue-600 text-white p-6">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-xl">{session.workout?.name}</h1>
          <button onClick={onComplete} className="p-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="flex items-center gap-2 text-sm text-blue-100 mb-2">
          <span>Упражнение {session.current_index + 1} из {session.total_exercises}</span>
          <span>•</span>
          <span>Подход {session.current_set} из {currentExercise.sets}</span>
        </div>

        <div className="bg-blue-500 rounded-full h-2 overflow-hidden">
          <div
            className="bg-white h-full transition-all"
            style={{ width: `${session.progress_percent}%` }}
          />
        </div>
      </div>

      <div className="p-6">
        <div className="bg-gradient-to-br from-blue-50 to-purple-50 rounded-2xl p-6 mb-6">
          <h2 className="text-2xl mb-3">{currentExercise.name}</h2>

          <div className="flex items-center gap-4 mb-4">
            <div className="bg-white px-4 py-2 rounded-lg">
              <div className="text-xs text-gray-600">Подходы</div>
              <div className="text-lg">{currentExercise.sets}</div>
            </div>
            <div className="bg-white px-4 py-2 rounded-lg">
              <div className="text-xs text-gray-600">Повторения</div>
              <div className="text-lg">{currentExercise.reps}</div>
            </div>
            <div className="bg-white px-4 py-2 rounded-lg">
              <div className="text-xs text-gray-600">Отдых</div>
              <div className="text-lg">{currentExercise.rest}с</div>
            </div>
          </div>

          <div className="bg-white/60 rounded-xl p-4 flex items-start gap-3">
            <svg className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm text-gray-700">{currentExercise.description}</p>
          </div>
        </div>

        <div className="bg-white border-2 border-gray-200 rounded-2xl p-6 mb-6">
          <div className="text-center mb-4">
            <div className="text-sm text-gray-600 mb-2">Текущий подход</div>
            <div className="text-5xl text-blue-600">{session.current_set}</div>
            <div className="text-gray-500">из {currentExercise.sets}</div>
          </div>

          <div className="flex gap-2 justify-center">
            {Array.from({ length: currentExercise.sets }).map((_, i) => (
              <div
                key={i}
                className={`w-3 h-3 rounded-full ${
                  i < session.current_set ? 'bg-green-500' : i === session.current_set - 1 ? 'bg-blue-500' : 'bg-gray-300'
                }`}
              />
            ))}
          </div>
        </div>

        {state === 'rest' && (
          <div className="bg-amber-50 border-2 border-amber-200 rounded-2xl p-6 mb-6">
            <div className="text-center">
              <svg className="w-12 h-12 text-amber-600 mx-auto mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div className="text-sm text-amber-700 mb-2">Отдых</div>
              <div className="text-5xl text-amber-900 mb-4">{timer}с</div>

              <div className="flex gap-3 justify-center">
                <button
                  onClick={togglePause}
                  className="bg-amber-600 text-white px-6 py-3 rounded-xl flex items-center gap-2"
                >
                  {isPaused ? (
                    <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M8 5v14l11-7z" />
                    </svg>
                  ) : (
                    <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                      <path d="M6 4h4v16H6V4zm8 0h4v16h-4V4z" />
                    </svg>
                  )}
                  {isPaused ? 'Продолжить' : 'Пауза'}
                </button>

                <button
                  onClick={handleSkipRest}
                  className="bg-white text-amber-700 border-2 border-amber-300 px-6 py-3 rounded-xl flex items-center gap-2"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
                  </svg>
                  Пропустить
                </button>
              </div>
            </div>
          </div>
        )}

        {state === 'ready' && (
          <button
            onClick={handleStart}
            className="w-full bg-blue-600 text-white py-5 rounded-xl flex items-center justify-center gap-2 text-lg"
          >
            <svg className="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
              <path d="M8 5v14l11-7z" />
            </svg>
            Начать упражнение
          </button>
        )}

        {state === 'exercise' && (
          <button
            onClick={handleCompleteSet}
            className="w-full bg-green-600 text-white py-5 rounded-xl flex items-center justify-center gap-2 text-lg"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Завершить подход
          </button>
        )}

        <div className="mt-6 p-4 bg-blue-50 rounded-xl">
          <h4 className="text-sm text-blue-900 mb-2">Советы:</h4>
          <ul className="text-sm text-blue-700 space-y-1">
            <li>• Контролируйте дыхание</li>
            <li>• Соблюдайте правильную технику</li>
            <li>• Не торопитесь</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
