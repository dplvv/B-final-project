import { useState } from 'react';
import type { Screen } from '../types';
import { useAuth } from '../context/AuthContext';

interface RegisterProps {
  onNavigate: (screen: Screen) => void;
}

export function Register({ onNavigate }: RegisterProps) {
  const { register } = useAuth();
  const [name, setName] = useState('');
  const [employeeId, setEmployeeId] = useState('');
  const [department, setDepartment] = useState('');
  const [position, setPosition] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (password !== confirm) {
      setError('Пароли не совпадают');
      return;
    }
    setLoading(true);
    try {
      await register({ name, employeeId, department, position, password });
    } catch (err: any) {
      setError(err.message || 'Не удалось зарегистрироваться');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 to-slate-800 flex items-center justify-center p-6">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <div className="bg-white w-20 h-20 rounded-3xl flex items-center justify-center mx-auto mb-4 shadow-lg">
            <svg className="w-10 h-10 text-slate-800" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h1 className="text-white text-3xl mb-2">Создание аккаунта</h1>
          <p className="text-slate-300">Заполните данные сотрудника</p>
        </div>

        <div className="bg-white rounded-3xl p-6 shadow-xl">
          <form onSubmit={handleRegister} className="space-y-4">
            <div>
              <label className="block text-gray-700 mb-2">ФИО</label>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
                placeholder="Иван Петров"
                required
              />
            </div>
            <div>
              <label className="block text-gray-700 mb-2">Табельный номер</label>
              <input
                value={employeeId}
                onChange={(e) => setEmployeeId(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
                required
              />
            </div>
            <div>
              <label className="block text-gray-700 mb-2">Отдел</label>
              <input
                value={department}
                onChange={(e) => setDepartment(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
              />
            </div>
            <div>
              <label className="block text-gray-700 mb-2">Должность</label>
              <input
                value={position}
                onChange={(e) => setPosition(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
              />
            </div>
            <div>
              <label className="block text-gray-700 mb-2">Пароль</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
                required
              />
            </div>
            <div>
              <label className="block text-gray-700 mb-2">Подтверждение пароля</label>
              <input
                type="password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl"
                required
              />
            </div>
            {error && <div className="text-sm text-red-600">{error}</div>}
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-slate-900 text-white py-3 rounded-xl hover:bg-slate-800 disabled:opacity-60"
            >
              {loading ? 'Создаем...' : 'Создать аккаунт'}
            </button>
          </form>
          <div className="mt-4 text-center text-sm text-slate-500">
            Уже есть аккаунт?{' '}
            <button className="text-slate-800 hover:underline" onClick={() => onNavigate('login')}>
              Войти
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
