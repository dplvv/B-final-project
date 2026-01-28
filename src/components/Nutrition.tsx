import { useEffect, useMemo, useState } from 'react';
import type { NutritionItem, Screen } from '../types';
import { apiRequest } from '../api/client';

interface NutritionProps {
  onNavigate: (screen: Screen) => void;
}

interface WaterState {
  date: string;
  amount: number;
  goal: number;
}

interface MealEntry {
  id: string;
  title: string;
  calories: number;
  mealType: string;
}

const mealTypes = [
  { id: 'breakfast', label: 'Завтрак' },
  { id: 'lunch', label: 'Обед' },
  { id: 'dinner', label: 'Ужин' },
  { id: 'snack', label: 'Перекус' },
];

export function Nutrition({ onNavigate }: NutritionProps) {
  const [selectedTab, setSelectedTab] = useState<'tips' | 'water' | 'meals'>('tips');
  const [items, setItems] = useState<NutritionItem[]>([]);
  const [category, setCategory] = useState('all');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [water, setWater] = useState<WaterState>({ date: '', amount: 0, goal: 2500 });
  const [waterLoading, setWaterLoading] = useState(false);

  const [meals, setMeals] = useState<MealEntry[]>([]);
  const [mealLoading, setMealLoading] = useState(false);
  const [showAddMeal, setShowAddMeal] = useState(false);
  const [mealForm, setMealForm] = useState({ title: '', calories: '', mealType: 'breakfast' });

  const today = useMemo(() => new Date().toISOString().slice(0, 10), []);

  const categories = useMemo(() => {
    const list = Array.from(new Set(items.map((item) => item.category).filter(Boolean)));
    return ['all', ...list];
  }, [items]);

  const loadItems = async (cat = category) => {
    setLoading(true);
    setError(null);
    try {
      const categoryParam = cat === 'all' ? '' : cat;
      const data = await apiRequest<{ items: any[] }>(`/nutrition?category=${encodeURIComponent(categoryParam)}`);
      const mapped = (data.items || []).map((item) => ({
        id: item.id,
        title: item.title,
        description: item.description,
        calories: Number(item.calories || 0),
        category: item.category || 'Общее',
      }));
      setItems(mapped);
    } catch (err: any) {
      setError(err?.message || 'Не удалось загрузить рекомендации по питанию');
    } finally {
      setLoading(false);
    }
  };

  const loadWater = async () => {
    setWaterLoading(true);
    try {
      const data = await apiRequest<WaterState>(`/nutrition/water?date=${today}`);
      setWater({
        date: data.date || today,
        amount: Number(data.amount || 0),
        goal: Number(data.goal || 2500),
      });
    } catch {
      setWater({ date: today, amount: 0, goal: 2500 });
    } finally {
      setWaterLoading(false);
    }
  };

  const addWater = async (delta: number) => {
    if (!delta) return;
    setWaterLoading(true);
    try {
      const data = await apiRequest<WaterState>('/nutrition/water', {
        method: 'POST',
        body: JSON.stringify({ date: today, delta }),
      });
      setWater({
        date: data.date || today,
        amount: Number(data.amount || 0),
        goal: Number(data.goal || water.goal),
      });
    } catch {
      setWater((prev) => ({ ...prev, amount: prev.amount + delta }));
    } finally {
      setWaterLoading(false);
    }
  };

  const loadMeals = async () => {
    setMealLoading(true);
    try {
      const data = await apiRequest<{ meals: any[] }>(`/nutrition/diary?date=${today}`);
      const mapped = (data.meals || []).map((meal) => ({
        id: meal.id,
        title: meal.title,
        calories: Number(meal.calories || 0),
        mealType: meal.meal_type || 'breakfast',
      }));
      setMeals(mapped);
    } catch {
      setMeals([]);
    } finally {
      setMealLoading(false);
    }
  };

  const addMeal = async (title: string, calories: number, mealType: string) => {
    if (!title) return;
    setMealLoading(true);
    try {
      await apiRequest('/nutrition/diary', {
        method: 'POST',
        body: JSON.stringify({ date: today, title, calories, meal_type: mealType }),
      });
      await loadMeals();
    } finally {
      setMealLoading(false);
    }
  };

  const removeMeal = async (id: string) => {
    setMealLoading(true);
    try {
      await apiRequest(`/nutrition/diary/${id}`, { method: 'DELETE' });
      await loadMeals();
    } finally {
      setMealLoading(false);
    }
  };

  useEffect(() => {
    loadItems();
    loadWater();
    loadMeals();
  }, []);

  const waterPercentage = water.goal ? Math.min(100, Math.round((water.amount / water.goal) * 100)) : 0;

  const groupedMeals = useMemo(() => {
    const group: Record<string, MealEntry[]> = {};
    mealTypes.forEach((type) => {
      group[type.id] = [];
    });
    meals.forEach((meal) => {
      if (!group[meal.mealType]) {
        group[meal.mealType] = [];
      }
      group[meal.mealType].push(meal);
    });
    return group;
  }, [meals]);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-green-600 text-white p-6">
        <button
          onClick={() => onNavigate('home')}
          className="flex items-center gap-2 mb-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <h1 className="text-2xl mb-2">Питание</h1>
        <p className="text-green-100">Рекомендации по питанию для спортсменов</p>
      </div>

      <div className="p-6 space-y-6">
        <div className="bg-white rounded-2xl p-2 shadow-sm flex gap-2">
          <button
            onClick={() => setSelectedTab('tips')}
            className={`flex-1 py-2 px-4 rounded-xl transition-colors ${
              selectedTab === 'tips' ? 'bg-green-600 text-white' : 'text-gray-600 hover:bg-gray-50'
            }`}
          >
            Советы
          </button>
          <button
            onClick={() => setSelectedTab('water')}
            className={`flex-1 py-2 px-4 rounded-xl transition-colors ${
              selectedTab === 'water' ? 'bg-green-600 text-white' : 'text-gray-600 hover:bg-gray-50'
            }`}
          >
            Вода
          </button>
          <button
            onClick={() => setSelectedTab('meals')}
            className={`flex-1 py-2 px-4 rounded-xl transition-colors ${
              selectedTab === 'meals' ? 'bg-green-600 text-white' : 'text-gray-600 hover:bg-gray-50'
            }`}
          >
            Приёмы пищи
          </button>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        {selectedTab === 'tips' && (
          <div className="space-y-4">
            <div className="flex gap-2 overflow-x-auto pb-2 -mx-6 px-6">
              {categories.map((item) => (
                <button
                  key={item}
                  onClick={() => {
                    setCategory(item);
                    loadItems(item);
                  }}
                  className={`px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                    category === item
                      ? 'bg-green-600 text-white'
                      : 'bg-white text-gray-700 border border-gray-200'
                  }`}
                >
                  {item === 'all' ? 'Все' : item}
                </button>
              ))}
            </div>

            {loading ? (
              <div className="text-sm text-gray-500">Загрузка рекомендаций...</div>
            ) : items.length === 0 ? (
              <div className="bg-white rounded-2xl p-6 text-center text-gray-600">
                Пока нет рекомендаций по выбранной категории.
              </div>
            ) : (
              <div className="space-y-3">
                {items.map((tip) => (
                  <div key={tip.id} className="bg-white rounded-2xl p-5 shadow-sm">
                    <div className="flex items-start gap-3">
                      <div className="bg-green-100 w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0">
                        <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <h4 className="mb-1">{tip.title}</h4>
                          {tip.calories ? (
                            <span className="text-xs text-gray-500">{tip.calories} ккал</span>
                          ) : null}
                        </div>
                        <p className="text-sm text-gray-600">{tip.description}</p>
                        {tip.category && (
                          <span className="text-xs text-green-700 bg-green-50 px-2 py-1 rounded-full mt-2 inline-block">
                            {tip.category}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {selectedTab === 'water' && (
          <div className="space-y-4">
            <div className="bg-gradient-to-br from-cyan-50 to-blue-50 rounded-2xl p-6">
              <div className="flex items-center justify-between mb-4">
                <h3>Вода сегодня</h3>
                <span className="text-sm text-gray-600">{water.amount}/{water.goal} мл</span>
              </div>

              <div className="flex items-end justify-center gap-2 mb-4">
                {[...Array(8)].map((_, index) => {
                  const filled = index < Math.floor((water.amount / water.goal) * 8);
                  return (
                    <div
                      key={index}
                      className={`w-8 rounded-t-lg transition-all ${
                        filled ? 'bg-cyan-500 h-16' : 'bg-gray-200 h-12'
                      }`}
                    />
                  );
                })}
              </div>

              <div className="h-3 bg-white/50 rounded-full overflow-hidden mb-4">
                <div
                  className="h-full bg-cyan-500 rounded-full transition-all duration-500"
                  style={{ width: `${waterPercentage}%` }}
                />
              </div>

              <div className="flex gap-2">
                <button
                  className="flex-1 bg-cyan-600 text-white py-3 rounded-xl hover:bg-cyan-700 disabled:opacity-60"
                  onClick={() => addWater(250)}
                  disabled={waterLoading}
                >
                  + 250 мл
                </button>
                <button
                  className="px-4 bg-white text-cyan-600 rounded-xl border border-cyan-200"
                  onClick={() => {
                    const value = prompt('Введите объем (мл)');
                    const parsed = value ? Number(value) : 0;
                    if (parsed > 0) {
                      addWater(parsed);
                    }
                  }}
                  disabled={waterLoading}
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                </button>
              </div>
            </div>

            <div className="bg-white rounded-2xl p-5 shadow-sm">
              <h3 className="mb-3">Советы по питьевому режиму</h3>
              <ul className="space-y-2 text-sm text-gray-700">
                <li className="flex items-start gap-2">
                  <span className="text-cyan-600">•</span>
                  <span>Начинайте день со стакана воды</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-cyan-600">•</span>
                  <span>Пейте воду за 30 минут до еды</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-cyan-600">•</span>
                  <span>Во время тренировки пейте каждые 15-20 минут</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-cyan-600">•</span>
                  <span>После тренировки восполните потерю жидкости</span>
                </li>
              </ul>
            </div>
          </div>
        )}

        {selectedTab === 'meals' && (
          <div className="space-y-4">
            <div className="bg-white rounded-2xl p-5 shadow-sm">
              <div className="flex items-center justify-between mb-3">
                <h3>Дневник питания</h3>
                <button
                  className="text-green-600 text-sm"
                  onClick={() => setShowAddMeal(true)}
                >
                  Добавить
                </button>
              </div>
              {mealLoading ? (
                <div className="text-sm text-gray-500">Загрузка...</div>
              ) : meals.length === 0 ? (
                <p className="text-sm text-gray-500">Сегодня записей нет</p>
              ) : (
                <div className="space-y-3">
                  {mealTypes.map((type) => (
                    <div key={type.id}>
                      <div className="text-xs uppercase text-gray-500 mb-2">{type.label}</div>
                      <div className="space-y-2">
                        {(groupedMeals[type.id] || []).map((meal) => (
                          <div key={meal.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
                            <div>
                              <div className="text-sm">{meal.title}</div>
                              {meal.calories ? (
                                <div className="text-xs text-gray-500">{meal.calories} ккал</div>
                              ) : null}
                            </div>
                            <button
                              className="text-red-500"
                              onClick={() => removeMeal(meal.id)}
                            >
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                              </svg>
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="space-y-3">
              <h3>Рекомендации</h3>
              {items.slice(0, 6).map((item) => (
                <div key={item.id} className="bg-white rounded-2xl p-4 shadow-sm flex items-center justify-between">
                  <div>
                    <div className="text-sm mb-1">{item.title}</div>
                    <div className="text-xs text-gray-500">{item.category}</div>
                  </div>
                  <button
                    className="text-green-600"
                    onClick={() => addMeal(item.title, item.calories, 'lunch')}
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="bg-amber-50 border border-amber-200 rounded-2xl p-5">
          <div className="flex items-start gap-3">
            <svg className="w-6 h-6 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <h4 className="text-amber-900 mb-1">Важно</h4>
              <p className="text-sm text-amber-700">
                Рекомендации носят общий характер. Для составления индивидуального плана питания проконсультируйтесь с диетологом или врачом.
              </p>
            </div>
          </div>
        </div>
      </div>

      {showAddMeal && (
        <div className="fixed inset-0 bg-black/40 flex items-end justify-center z-50" onClick={() => setShowAddMeal(false)}>
          <div
            className="bg-white rounded-t-3xl w-full max-w-md p-6"
            onClick={(event) => event.stopPropagation()}
          >
            <h2 className="text-xl mb-4">Добавить приём пищи</h2>
            <form
              className="space-y-4"
              onSubmit={(event) => {
                event.preventDefault();
                addMeal(mealForm.title, Number(mealForm.calories || 0), mealForm.mealType);
                setMealForm({ title: '', calories: '', mealType: 'breakfast' });
                setShowAddMeal(false);
              }}
            >
              <div>
                <label className="text-sm text-gray-600">Название</label>
                <input
                  value={mealForm.title}
                  onChange={(event) => setMealForm({ ...mealForm, title: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  required
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Калории</label>
                <input
                  type="number"
                  value={mealForm.calories}
                  onChange={(event) => setMealForm({ ...mealForm, calories: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Тип приёма пищи</label>
                <select
                  value={mealForm.mealType}
                  onChange={(event) => setMealForm({ ...mealForm, mealType: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                >
                  {mealTypes.map((type) => (
                    <option key={type.id} value={type.id}>
                      {type.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  className="flex-1 py-3 rounded-xl border border-gray-200"
                  onClick={() => setShowAddMeal(false)}
                >
                  Отмена
                </button>
                <button type="submit" className="flex-1 py-3 rounded-xl bg-green-600 text-white">
                  Сохранить
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
