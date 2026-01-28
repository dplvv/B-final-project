import { useEffect, useMemo, useState } from 'react';
import type { Screen, Goal } from '../types';
import { apiRequest } from '../api/client';

interface GoalsProps {
  onNavigate: (screen: Screen) => void;
}

const categoryOptions = [
  { value: 'strength', label: 'Сила' },
  { value: 'flexibility', label: 'Гибкость' },
  { value: 'endurance', label: 'Выносливость' },
  { value: 'weight', label: 'Вес' },
  { value: 'other', label: 'Другое' },
];

const categoryStyles: Record<string, { bg: string; text: string; bar: string; badge: string }> = {
  strength: {
    bg: 'bg-blue-100',
    text: 'text-blue-600',
    bar: 'bg-blue-500',
    badge: 'bg-blue-50 text-blue-700',
  },
  flexibility: {
    bg: 'bg-purple-100',
    text: 'text-purple-600',
    bar: 'bg-purple-500',
    badge: 'bg-purple-50 text-purple-700',
  },
  endurance: {
    bg: 'bg-green-100',
    text: 'text-green-600',
    bar: 'bg-green-500',
    badge: 'bg-green-50 text-green-700',
  },
  weight: {
    bg: 'bg-orange-100',
    text: 'text-orange-600',
    bar: 'bg-orange-500',
    badge: 'bg-orange-50 text-orange-700',
  },
  other: {
    bg: 'bg-gray-100',
    text: 'text-gray-600',
    bar: 'bg-gray-500',
    badge: 'bg-gray-50 text-gray-700',
  },
};

const emptyForm: Omit<Goal, 'id'> = {
  title: '',
  description: '',
  targetDate: '',
  progress: 0,
  category: 'strength',
};

export function Goals({ onNavigate }: GoalsProps) {
  const [goals, setGoals] = useState<Goal[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingGoal, setEditingGoal] = useState<Goal | null>(null);
  const [form, setForm] = useState<Omit<Goal, 'id'>>(emptyForm);
  const [saving, setSaving] = useState(false);

  const loadGoals = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiRequest<{ goals: any[] }>('/goals');
      const items = (data.goals || []).map((goal) => ({
        id: goal.id,
        title: goal.title,
        description: goal.description || '',
        targetDate: goal.target_date || '',
        progress: Number(goal.progress || 0),
        category: goal.category || 'other',
      }));
      setGoals(items);
    } catch (err: any) {
      setError(err?.message || 'Не удалось загрузить цели');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadGoals();
  }, []);

  const openCreate = () => {
    setEditingGoal(null);
    setForm(emptyForm);
    setShowForm(true);
  };

  const openEdit = (goal: Goal) => {
    setEditingGoal(goal);
    setForm({
      title: goal.title,
      description: goal.description,
      targetDate: goal.targetDate,
      progress: goal.progress,
      category: goal.category,
    });
    setShowForm(true);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setSaving(true);
    try {
      if (editingGoal) {
        await apiRequest(`/goals/${editingGoal.id}`, {
          method: 'PUT',
          body: JSON.stringify({
            title: form.title,
            description: form.description,
            target_date: form.targetDate,
            category: form.category,
            progress: form.progress,
          }),
        });
      } else {
        await apiRequest('/goals', {
          method: 'POST',
          body: JSON.stringify({
            title: form.title,
            description: form.description,
            target_date: form.targetDate,
            category: form.category,
            progress: form.progress,
          }),
        });
      }
      setShowForm(false);
      await loadGoals();
    } catch (err: any) {
      setError(err?.message || 'Не удалось сохранить цель');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (goalId: string) => {
    if (!confirm('Удалить цель?')) return;
    try {
      await apiRequest(`/goals/${goalId}`, { method: 'DELETE' });
      await loadGoals();
    } catch (err: any) {
      setError(err?.message || 'Не удалось удалить цель');
    }
  };

  const summary = useMemo(() => {
    const total = goals.length;
    const completed = goals.filter((goal) => goal.progress >= 100).length;
    const average = total > 0 ? Math.round(goals.reduce((sum, goal) => sum + goal.progress, 0) / total) : 0;
    return { total, completed, average };
  }, [goals]);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-600 text-white p-6">
        <button
          onClick={() => onNavigate('home')}
          className="flex items-center gap-2 mb-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="text-2xl mb-2">Мои цели</h1>
        <p className="text-blue-100">Отслеживайте свой прогресс</p>
      </div>

      <div className="p-6 space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        <div className="grid grid-cols-3 gap-3">
          <div className="bg-white rounded-2xl p-4 text-center shadow-sm">
            <div className="text-2xl mb-1">{summary.total}</div>
            <div className="text-xs text-gray-600">Целей</div>
          </div>
          <div className="bg-white rounded-2xl p-4 text-center shadow-sm">
            <div className="text-2xl text-green-600 mb-1">{summary.completed}</div>
            <div className="text-xs text-gray-600">Достигнуто</div>
          </div>
          <div className="bg-white rounded-2xl p-4 text-center shadow-sm">
            <div className="text-2xl text-blue-600 mb-1">{summary.average}%</div>
            <div className="text-xs text-gray-600">Средний</div>
          </div>
        </div>

        {loading ? (
          <div className="text-sm text-gray-500">Загрузка целей...</div>
        ) : goals.length === 0 ? (
          <div className="bg-white rounded-2xl p-6 text-center text-gray-600">
            Пока нет целей. Добавьте новую цель, чтобы начать отслеживание прогресса.
          </div>
        ) : (
          <div className="space-y-3">
            {goals.map((goal) => {
              const styles = categoryStyles[goal.category] || categoryStyles.other;
              const daysLeft = goal.targetDate
                ? Math.ceil((new Date(goal.targetDate).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24))
                : null;

              return (
                <div key={goal.id} className="bg-white rounded-2xl p-5 shadow-sm">
                  <div className="flex items-start gap-3 mb-3">
                    <div className={`${styles.bg} w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0`}>
                      <svg className={`w-6 h-6 ${styles.text}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                      </svg>
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <h3 className="mb-1">{goal.title}</h3>
                        <span className={`text-xs px-2 py-1 rounded-full ${styles.badge}`}>
                          {categoryOptions.find((item) => item.value === goal.category)?.label || 'Другое'}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600">{goal.description}</p>
                    </div>
                  </div>

                  <div className="mb-3">
                    <div className="flex justify-between text-sm mb-2">
                      <span className="text-gray-600">Прогресс</span>
                      <span>{goal.progress}%</span>
                    </div>
                    <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                      <div
                        className={`h-full ${styles.bar} rounded-full transition-all duration-300`}
                        style={{ width: `${goal.progress}%` }}
                      />
                    </div>
                  </div>

                  <div className="flex flex-wrap items-center justify-between gap-3 text-sm">
                    <div className="flex items-center gap-1 text-gray-600">
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                      <span>
                        {goal.targetDate
                          ? daysLeft && daysLeft > 0
                            ? `${daysLeft} дней осталось`
                            : 'Срок истёк'
                          : 'Дата не указана'}
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <button
                        onClick={() => openEdit(goal)}
                        className="text-blue-600 hover:underline"
                      >
                        Подробнее
                      </button>
                      <button
                        onClick={() => handleDelete(goal.id)}
                        className="text-red-500 hover:underline"
                      >
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <button
          onClick={openCreate}
          className="w-full bg-blue-600 text-white py-4 rounded-xl flex items-center justify-center gap-2 hover:bg-blue-700 transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Добавить цель
        </button>

        <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-2xl p-5">
          <h3 className="mb-3">Советы по достижению целей</h3>
          <ul className="space-y-2 text-sm text-gray-700">
            <li className="flex items-start gap-2">
              <span className="text-blue-600">•</span>
              <span>Ставьте конкретные и измеримые цели</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-blue-600">•</span>
              <span>Разбивайте большие цели на маленькие шаги</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-blue-600">•</span>
              <span>Отслеживайте прогресс регулярно</span>
            </li>
            <li className="flex items-start gap-2">
              <span className="text-blue-600">•</span>
              <span>Корректируйте цели при необходимости</span>
            </li>
          </ul>
        </div>
      </div>

      {showForm && (
        <div className="fixed inset-0 bg-black/40 flex items-end justify-center z-50" onClick={() => setShowForm(false)}>
          <div
            className="bg-white rounded-t-3xl w-full max-w-md p-6"
            onClick={(event) => event.stopPropagation()}
          >
            <h2 className="text-xl mb-4">
              {editingGoal ? 'Редактировать цель' : 'Новая цель'}
            </h2>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="text-sm text-gray-600">Название</label>
                <input
                  value={form.title}
                  onChange={(event) => setForm({ ...form, title: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  placeholder="Например, укрепить спину"
                  required
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Описание</label>
                <textarea
                  value={form.description}
                  onChange={(event) => setForm({ ...form, description: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  rows={3}
                  placeholder="Опишите цель подробнее"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Категория</label>
                <select
                  value={form.category}
                  onChange={(event) => setForm({ ...form, category: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                >
                  {categoryOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-sm text-gray-600">Дата завершения</label>
                <input
                  type="date"
                  value={form.targetDate}
                  onChange={(event) => setForm({ ...form, targetDate: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Прогресс</label>
                <div className="flex items-center gap-3">
                  <input
                    type="range"
                    min={0}
                    max={100}
                    value={form.progress}
                    onChange={(event) => setForm({ ...form, progress: Number(event.target.value) })}
                    className="w-full"
                  />
                  <span className="text-sm w-12 text-right">{form.progress}%</span>
                </div>
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="flex-1 py-3 rounded-xl border border-gray-200"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className="flex-1 py-3 rounded-xl bg-blue-600 text-white disabled:opacity-60"
                >
                  {saving ? 'Сохраняем...' : 'Сохранить'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
