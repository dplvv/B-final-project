import { useState } from 'react';
import { apiRequest } from '../api/client';
import { useAuth } from '../context/AuthContext';

interface QuestionnaireProps {
  onComplete: () => void;
}

interface Question {
  id: string;
  question: string;
  type: 'single' | 'multiple';
  options: { value: string; label: string }[];
}

const questions: Question[] = [
  {
    id: 'activity',
    question: 'Как часто вы занимаетесь физической активностью?',
    type: 'single',
    options: [
      { value: 'never', label: 'Не занимаюсь' },
      { value: '1-2', label: '1-2 раза в неделю' },
      { value: '3-4', label: '3-4 раза в неделю' },
      { value: '5+', label: '5+ раз в неделю' }
    ]
  },
  {
    id: 'experience',
    question: 'Ваш опыт спортивных тренировок?',
    type: 'single',
    options: [
      { value: 'beginner', label: 'Новичок (меньше 6 месяцев)' },
      { value: 'intermediate', label: 'Средний (6 мес - 2 года)' },
      { value: 'advanced', label: 'Продвинутый (более 2 лет)' }
    ]
  },
  {
    id: 'restrictions',
    question: 'Есть ли у вас ограничения по здоровью?',
    type: 'multiple',
    options: [
      { value: 'back', label: 'Проблемы со спиной' },
      { value: 'joints', label: 'Проблемы с суставами' },
      { value: 'cardio', label: 'Сердечно-сосудистые' },
      { value: 'none', label: 'Нет ограничений' }
    ]
  },
  {
    id: 'goals',
    question: 'Какие цели вы хотите достичь?',
    type: 'multiple',
    options: [
      { value: 'rehab', label: 'Реабилитация' },
      { value: 'strength', label: 'Увеличение силы' },
      { value: 'flexibility', label: 'Улучшение гибкости' },
      { value: 'endurance', label: 'Выносливость' },
      { value: 'posture', label: 'Коррекция осанки' }
    ]
  },
  {
    id: 'pain',
    question: 'Испытываете ли вы боль или дискомфорт?',
    type: 'multiple',
    options: [
      { value: 'neck', label: 'Шея' },
      { value: 'shoulders', label: 'Плечи' },
      { value: 'back', label: 'Спина' },
      { value: 'knees', label: 'Колени' },
      { value: 'none', label: 'Нет дискомфорта' }
    ]
  }
];

export function Questionnaire({ onComplete }: QuestionnaireProps) {
  const { updateProfile } = useAuth();
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string | string[]>>({});
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const question = questions[currentQuestion];
  const progress = ((currentQuestion + 1) / questions.length) * 100;

  const handleAnswer = (value: string) => {
    if (question.type === 'single') {
      setAnswers({ ...answers, [question.id]: value });
    } else {
      const current = (answers[question.id] as string[]) || [];
      if (value === 'none') {
        setAnswers({ ...answers, [question.id]: ['none'] });
      } else {
        const filtered = current.filter(v => v !== 'none');
        if (current.includes(value)) {
          setAnswers({ ...answers, [question.id]: filtered.filter(v => v !== value) });
        } else {
          setAnswers({ ...answers, [question.id]: [...filtered, value] });
        }
      }
    }
  };

  const isAnswered = () => {
    const answer = answers[question.id];
    if (question.type === 'single') {
      return !!answer;
    } else {
      return Array.isArray(answer) && answer.length > 0;
    }
  };

  const handleNext = async () => {
    if (currentQuestion < questions.length - 1) {
      setCurrentQuestion(currentQuestion + 1);
      return;
    }

    const fitnessLevel = (answers.experience as 'beginner' | 'intermediate' | 'advanced') || 'beginner';
    const restrictions = (answers.restrictions as string[]) || [];
    const goals = (answers.goals as string[]) || [];

    setSaving(true);
    setError('');
    try {
      await apiRequest('/questionnaire', {
        method: 'POST',
        body: JSON.stringify({
          fitness_level: fitnessLevel,
          restrictions,
          goals,
          answers,
        }),
      });
      await updateProfile({ fitnessLevel, restrictions, goals });
      onComplete();
    } catch (err: any) {
      setError(err.message || 'Не удалось сохранить данные');
    } finally {
      setSaving(false);
    }
  };

  const handleBack = () => {
    if (currentQuestion > 0) {
      setCurrentQuestion(currentQuestion - 1);
    }
  };

  const isSelected = (value: string) => {
    const answer = answers[question.id];
    if (question.type === 'single') {
      return answer === value;
    } else {
      return Array.isArray(answer) && answer.includes(value);
    }
  };

  return (
    <div className="min-h-screen bg-white">
      <div className="bg-blue-600 text-white p-6">
        <div className="flex items-center gap-4 mb-4">
          {currentQuestion > 0 && (
            <button onClick={handleBack} className="p-2 -ml-2">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
            </button>
          )}
          <div className="flex-1">
            <h1 className="text-xl">Оценка состояния</h1>
            <p className="text-sm text-blue-100">
              Вопрос {currentQuestion + 1} из {questions.length}
            </p>
          </div>
        </div>

        <div className="bg-blue-500 rounded-full h-2 overflow-hidden">
          <div
            className="bg-white h-full transition-all duration-300"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      <div className="p-6">
        <h2 className="text-xl mb-6">{question.question}</h2>

        <div className="space-y-3">
          {question.options.map((option) => (
            <button
              key={option.value}
              onClick={() => handleAnswer(option.value)}
              className={`w-full p-4 rounded-xl border-2 transition-all text-left flex items-center gap-3 ${
                isSelected(option.value)
                  ? 'border-blue-600 bg-blue-50'
                  : 'border-gray-200 bg-white'
              }`}
            >
              <div className={`w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                isSelected(option.value)
                  ? 'border-blue-600 bg-blue-600'
                  : 'border-gray-300'
              }`}>
                {isSelected(option.value) && (
                  <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                )}
              </div>
              <span className={isSelected(option.value) ? 'text-blue-900' : 'text-gray-700'}>
                {option.label}
              </span>
            </button>
          ))}
        </div>

        {error && <div className="text-sm text-red-600 mt-4">{error}</div>}
      </div>

      <div className="fixed bottom-0 left-0 right-0 p-6 bg-white border-t max-w-3xl mx-auto">
        <button
          onClick={handleNext}
          disabled={!isAnswered() || saving}
          className="w-full bg-blue-600 text-white py-4 rounded-xl flex items-center justify-center gap-2 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
        >
          {saving
            ? 'Сохраняем...'
            : currentQuestion < questions.length - 1
              ? (
                <>
                  Далее
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </>
              )
              : (
                <>
                  Завершить
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </>
              )}
        </button>
      </div>
    </div>
  );
}
