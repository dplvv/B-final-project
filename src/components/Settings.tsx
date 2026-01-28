import { useEffect, useState } from 'react';
import type { Screen, Settings as SettingsType } from '../types';
import { useAuth } from '../context/AuthContext';

interface SettingsProps {
  onNavigate: (screen: Screen) => void;
}

export function Settings({ onNavigate }: SettingsProps) {
  const { settings, updateSettings } = useAuth();
  const [draft, setDraft] = useState<SettingsType | null>(null);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (settings) {
      setDraft(settings);
    }
  }, [settings]);

  if (!draft) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-sm text-gray-500">Загрузка настроек...</div>
      </div>
    );
  }

  const toggleNotification = (key: keyof SettingsType['notifications']) => {
    setDraft({
      ...draft,
      notifications: {
        ...draft.notifications,
        [key]: !draft.notifications[key],
      },
    });
  };

  const togglePrivacy = (key: keyof SettingsType['privacy']) => {
    setDraft({
      ...draft,
      privacy: {
        ...draft.privacy,
        [key]: !draft.privacy[key],
      },
    });
  };

  const save = async () => {
    setSaving(true);
    setMessage('');
    try {
      await updateSettings(draft);
      setMessage('Настройки сохранены');
    } catch (err: any) {
      setMessage(err.message || 'Не удалось сохранить');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-blue-600 text-white p-6">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('profile')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Настройки</h1>
        </div>
        <p className="text-blue-100 text-sm">Управление приложением</p>
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <div className="bg-blue-100 p-2 rounded-lg">
              <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            </div>
            <h3>Уведомления</h3>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm mb-1">Уведомления включены</div>
                <p className="text-xs text-gray-600">Получать все уведомления</p>
              </div>
              <button
                onClick={() => toggleNotification('enabled')}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  draft.notifications.enabled ? 'bg-blue-600' : 'bg-gray-300'
                }`}
              >
                <div className={`absolute w-5 h-5 bg-white rounded-full top-0.5 transition-transform ${
                  draft.notifications.enabled ? 'translate-x-6' : 'translate-x-0.5'
                }`} />
              </button>
            </div>

            <div className="h-px bg-gray-200" />

            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm mb-1">Напоминания о тренировках</div>
                <p className="text-xs text-gray-600">Напоминать о запланированных тренировках</p>
              </div>
              <button
                onClick={() => toggleNotification('workoutReminders')}
                disabled={!draft.notifications.enabled}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  draft.notifications.workoutReminders && draft.notifications.enabled ? 'bg-blue-600' : 'bg-gray-300'
                }`}
              >
                <div className={`absolute w-5 h-5 bg-white rounded-full top-0.5 transition-transform ${
                  draft.notifications.workoutReminders ? 'translate-x-6' : 'translate-x-0.5'
                }`} />
              </button>
            </div>

            <div className="h-px bg-gray-200" />

            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm mb-1">Уведомления о достижениях</div>
                <p className="text-xs text-gray-600">Получать уведомления о новых достижениях</p>
              </div>
              <button
                onClick={() => toggleNotification('achievementAlerts')}
                disabled={!draft.notifications.enabled}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  draft.notifications.achievementAlerts && draft.notifications.enabled ? 'bg-blue-600' : 'bg-gray-300'
                }`}
              >
                <div className={`absolute w-5 h-5 bg-white rounded-full top-0.5 transition-transform ${
                  draft.notifications.achievementAlerts ? 'translate-x-6' : 'translate-x-0.5'
                }`} />
              </button>
            </div>

            <div className="h-px bg-gray-200" />

            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm mb-1">Еженедельный отчёт</div>
                <p className="text-xs text-gray-600">Получать сводку за неделю</p>
              </div>
              <button
                onClick={() => toggleNotification('weeklyReports')}
                disabled={!draft.notifications.enabled}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  draft.notifications.weeklyReports && draft.notifications.enabled ? 'bg-blue-600' : 'bg-gray-300'
                }`}
              >
                <div className={`absolute w-5 h-5 bg-white rounded-full top-0.5 transition-transform ${
                  draft.notifications.weeklyReports ? 'translate-x-6' : 'translate-x-0.5'
                }`} />
              </button>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <div className="bg-green-100 p-2 rounded-lg">
              <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <h3>Конфиденциальность</h3>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm mb-1">Делиться прогрессом</div>
                <p className="text-xs text-gray-600">Показывать прогресс в сообществе</p>
              </div>
              <button
                onClick={() => togglePrivacy('shareProgress')}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  draft.privacy.shareProgress ? 'bg-blue-600' : 'bg-gray-300'
                }`}
              >
                <div className={`absolute w-5 h-5 bg-white rounded-full top-0.5 transition-transform ${
                  draft.privacy.shareProgress ? 'translate-x-6' : 'translate-x-0.5'
                }`} />
              </button>
            </div>

            <div className="h-px bg-gray-200" />

            <div className="flex items-center justify-between py-2">
              <div>
                <div className="text-sm mb-1">Показывать в рейтинге</div>
                <p className="text-xs text-gray-600">Участвовать в рейтинге сотрудников</p>
              </div>
              <button
                onClick={() => togglePrivacy('showInLeaderboard')}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  draft.privacy.showInLeaderboard ? 'bg-blue-600' : 'bg-gray-300'
                }`}
              >
                <div className={`absolute w-5 h-5 bg-white rounded-full top-0.5 transition-transform ${
                  draft.privacy.showInLeaderboard ? 'translate-x-6' : 'translate-x-0.5'
                }`} />
              </button>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center gap-3 mb-4">
            <div className="bg-orange-100 p-2 rounded-lg">
              <svg className="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3>О приложении</h3>
          </div>

          <div className="space-y-3">
            <button
              onClick={() => onNavigate('support')}
              className="w-full flex items-center justify-between py-3 hover:bg-gray-50 rounded-lg px-2 -mx-2"
            >
              <div className="text-sm">Помощь и поддержка</div>
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>

            <div className="h-px bg-gray-200" />

            <button
              className="w-full flex items-center justify-between py-3 hover:bg-gray-50 rounded-lg px-2 -mx-2"
              onClick={() => window.alert('Политика конфиденциальности доступна у администратора.')} 
            >
              <div className="text-sm">Политика конфиденциальности</div>
              <svg className="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
            </button>

            <div className="h-px bg-gray-200" />

            <div className="py-3 px-2 -mx-2">
              <div className="text-sm text-gray-600 mb-1">Версия приложения</div>
              <p className="text-xs text-gray-500">1.0.0 (build 2026.01)</p>
            </div>
          </div>
        </div>

        {message && <div className="text-sm text-gray-600">{message}</div>}

        <button
          className="w-full bg-gray-900 text-white py-3 rounded-xl"
          onClick={save}
          disabled={saving}
        >
          {saving ? 'Сохраняем...' : 'Сохранить'}
        </button>
      </div>
    </div>
  );
}
