import { useEffect, useState } from 'react';
import type { Notification, Screen } from '../types';
import { apiRequest } from '../api/client';

interface NotificationsProps {
  onNavigate: (screen: Screen) => void;
}

export function Notifications({ onNavigate }: NotificationsProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);

  const load = async () => {
    const data = await apiRequest<any>('/notifications');
    const items = (data.notifications || []).map((n: any) => ({
      id: n.id,
      title: n.title,
      message: n.message,
      type: n.type,
      date: n.created_at,
      read: n.read,
    }));
    setNotifications(items);
  };

  useEffect(() => {
    load();
  }, []);

  const markRead = async (id: string) => {
    await apiRequest(`/notifications/${id}/read`, { method: 'POST' });
    await load();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-gradient-to-br from-indigo-600 to-blue-700 text-white p-6">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('profile')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Уведомления</h1>
        </div>
        <p className="text-blue-100 text-sm">Все сообщения и напоминания</p>
      </div>

      <div className="p-4 space-y-3">
        {notifications.map((notification) => (
          <div key={notification.id} className={`bg-white rounded-2xl p-4 shadow-sm ${!notification.read ? 'border-l-4 border-blue-600' : ''}`}>
            <div className="flex items-start justify-between gap-3">
              <div>
                <h4 className="mb-1">{notification.title}</h4>
                <p className="text-sm text-gray-600">{notification.message}</p>
                <div className="text-xs text-gray-400 mt-2">{notification.date}</div>
              </div>
              {!notification.read && (
                <button className="text-xs text-blue-600" onClick={() => markRead(notification.id)}>
                  Отметить
                </button>
              )}
            </div>
          </div>
        ))}
        {notifications.length === 0 && (
          <div className="text-sm text-gray-500">Нет уведомлений.</div>
        )}
      </div>
    </div>
  );
}
