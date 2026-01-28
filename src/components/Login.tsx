import { useState } from 'react';
import type { Screen } from '../types';
import { useAuth } from '../context/AuthContext';
import { apiRequest } from '../api/client';

interface LoginProps {
  onNavigate: (screen: Screen) => void;
}

export function Login({ onNavigate }: LoginProps) {
  const { login } = useAuth();
  const [employeeId, setEmployeeId] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [resetOpen, setResetOpen] = useState(false);
  const [resetEmployeeId, setResetEmployeeId] = useState('');
  const [resetMessage, setResetMessage] = useState('');
  const [resetStatus, setResetStatus] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(employeeId, password);
    } catch (err: any) {
      setError(err.message || 'Не удалось войти');
    } finally {
      setLoading(false);
    }
  };

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setResetStatus('');
    try {
      await apiRequest('/auth/forgot', {
        method: 'POST',
        body: JSON.stringify({ employee_id: resetEmployeeId, message: resetMessage }),
      });
      setResetStatus('Запрос отправлен. Администратор свяжется с вами.');
      setResetEmployeeId('');
      setResetMessage('');
    } catch (err: any) {
      setResetStatus(err.message || 'Не удалось отправить запрос');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-600 to-blue-800 flex items-center justify-center p-6">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <div className="bg-white w-20 h-20 rounded-3xl flex items-center justify-center mx-auto mb-4 shadow-lg">
            <svg className="w-10 h-10 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>
          <h1 className="text-white text-3xl mb-2">Спорт РОСАТОМ</h1>
          <p className="text-blue-200">Система спортивной реабилитации</p>
        </div>

        <div className="bg-white rounded-3xl p-6 shadow-xl">
          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <label className="block text-gray-700 mb-2">Табельный номер</label>
              <input
                type="text"
                value={employeeId}
                onChange={(e) => setEmployeeId(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="123456"
                required
              />
            </div>

            <div>
              <label className="block text-gray-700 mb-2">Пароль</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="••••••••"
                required
              />
            </div>

            {error && (
              <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg p-3">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 text-white py-3 rounded-xl hover:bg-blue-700 transition-colors disabled:opacity-60"
            >
              {loading ? 'Входим...' : 'Войти'}
            </button>
          </form>

          <div className="mt-6 space-y-3">
            <button
              className="w-full text-blue-600 text-sm hover:underline"
              type="button"
              onClick={() => setResetOpen(true)}
            >
              Забыли пароль?
            </button>
            <div className="text-center text-gray-500 text-sm">
              Нет аккаунта?{' '}
              <button className="text-blue-600 hover:underline" onClick={() => onNavigate('register')}>
                Зарегистрироваться
              </button>
            </div>
          </div>
        </div>

        <div className="mt-6 text-center text-blue-100 text-sm">
          <p>Приложение предназначено для сотрудников ГК РОСАТОМ</p>
          <p className="mt-2">Версия 1.0.0</p>
        </div>
      </div>

      {resetOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-6 z-50">
          <div className="bg-white rounded-2xl p-6 w-full max-w-md shadow-xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg">Сброс пароля</h3>
              <button onClick={() => setResetOpen(false)}>
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleReset} className="space-y-3">
              <div>
                <label className="block text-sm text-gray-600 mb-1">Табельный номер</label>
                <input
                  value={resetEmployeeId}
                  onChange={(e) => setResetEmployeeId(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-200 rounded-lg"
                  required
                />
              </div>
              <div>
                <label className="block text-sm text-gray-600 mb-1">Комментарий</label>
                <textarea
                  value={resetMessage}
                  onChange={(e) => setResetMessage(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-200 rounded-lg"
                  rows={3}
                  placeholder="Уточните контактный телефон или почту"
                />
              </div>
              {resetStatus && <div className="text-sm text-gray-600">{resetStatus}</div>}
              <button className="w-full bg-blue-600 text-white py-2.5 rounded-lg">Отправить</button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
