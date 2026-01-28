import { useEffect, useState } from 'react';
import type { Screen, SupportTicket } from '../types';
import { apiRequest } from '../api/client';

interface SupportProps {
  onNavigate: (screen: Screen) => void;
}

export function Support({ onNavigate }: SupportProps) {
  const [faq, setFaq] = useState<any[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [category, setCategory] = useState('');
  const [tickets, setTickets] = useState<SupportTicket[]>([]);
  const [form, setForm] = useState({ category: '', subject: '', message: '' });
  const [status, setStatus] = useState('');

  const load = async () => {
    const data = await apiRequest<any>('/support/faq');
    setFaq(data.faq || []);
    setCategories(data.categories || []);

    const ticketData = await apiRequest<any>('/support/tickets');
    const list = (ticketData.tickets || []).map((t: any) => ({
      id: t.id,
      category: t.category,
      subject: t.subject,
      status: t.status,
      createdAt: t.created_at,
    }));
    setTickets(list);
  };

  useEffect(() => {
    load();
  }, []);

  const filteredFaq = faq.filter((item) => {
    if (!category || category === 'Все') return true;
    return item.category === category;
  });

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStatus('');
    await apiRequest('/support/tickets', {
      method: 'POST',
      body: JSON.stringify(form),
    });
    setForm({ category: '', subject: '', message: '' });
    setStatus('Обращение отправлено');
    await load();
  };

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="bg-gradient-to-br from-emerald-600 to-green-600 text-white p-6 rounded-b-3xl">
        <div className="flex items-center gap-4 mb-2">
          <button onClick={() => onNavigate('settings')} className="p-2 -ml-2">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <h1>Помощь и поддержка</h1>
        </div>
        <p className="text-emerald-100 text-sm">Ответы на вопросы и обращения</p>
      </div>

      <div className="p-4 space-y-4">
        <div className="bg-white rounded-2xl p-4 shadow-sm">
          <h3 className="mb-3">Категории</h3>
          <div className="flex gap-2 overflow-x-auto pb-2">
            {categories.map((item) => (
              <button
                key={item}
                onClick={() => setCategory(item)}
                className={`px-4 py-2 rounded-full whitespace-nowrap transition-colors ${
                  category === item ? 'bg-emerald-600 text-white' : 'bg-gray-100 text-gray-700'
                }`}
              >
                {item}
              </button>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-4 shadow-sm">
          <h3 className="mb-3">Частые вопросы</h3>
          <div className="space-y-3">
            {filteredFaq.map((item) => (
              <div key={item.id} className="border border-gray-100 rounded-xl p-4">
                <div className="text-sm font-medium mb-2">{item.question}</div>
                <div className="text-sm text-gray-600">{item.answer}</div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-4 shadow-sm">
          <h3 className="mb-3">Мои обращения</h3>
          <div className="space-y-2">
            {tickets.map((ticket) => (
              <div key={ticket.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
                <div>
                  <div className="text-sm">{ticket.subject}</div>
                  <div className="text-xs text-gray-500">{ticket.category} • {ticket.createdAt}</div>
                </div>
                <div className="text-xs text-emerald-600">{ticket.status}</div>
              </div>
            ))}
            {tickets.length === 0 && (
              <div className="text-sm text-gray-500">Обращений пока нет.</div>
            )}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-4 shadow-sm" id="contact">
          <h3 className="mb-3">Связаться с поддержкой</h3>
          <form className="space-y-3" onSubmit={submit}>
            <div>
              <label className="text-xs text-gray-500" htmlFor="category">Категория</label>
              <input
                id="category"
                name="category"
                className="mt-1 w-full px-3 py-2 rounded-lg border border-gray-200"
                value={form.category}
                onChange={(e) => setForm({ ...form, category: e.target.value })}
                placeholder="Тренировки, Приложение"
                required
              />
            </div>
            <div>
              <label className="text-xs text-gray-500" htmlFor="subject">Тема</label>
              <input
                id="subject"
                name="subject"
                className="mt-1 w-full px-3 py-2 rounded-lg border border-gray-200"
                value={form.subject}
                onChange={(e) => setForm({ ...form, subject: e.target.value })}
                placeholder="Короткое описание вопроса"
                required
              />
            </div>
            <div>
              <label className="text-xs text-gray-500" htmlFor="message">Сообщение</label>
              <textarea
                id="message"
                name="message"
                rows={4}
                className="mt-1 w-full px-3 py-2 rounded-lg border border-gray-200"
                value={form.message}
                onChange={(e) => setForm({ ...form, message: e.target.value })}
                placeholder="Опишите проблему"
                required
              />
            </div>
            {status && <div className="text-sm text-emerald-600">{status}</div>}
            <button className="w-full bg-green-600 text-white py-3 rounded-xl">Отправить</button>
          </form>
        </div>
      </div>
    </div>
  );
}
