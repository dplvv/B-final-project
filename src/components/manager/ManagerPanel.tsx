import { useEffect, useMemo, useState, type ReactNode } from 'react';
import type { Screen } from '../../types';
import { apiRequest } from '../../api/client';

interface ManagerPanelProps {
  onNavigate: (screen: Screen) => void;
}

type ManagerTab = 'dashboard' | 'employees' | 'awards' | 'redemptions' | 'support';

const tabs: { id: ManagerTab; label: string }[] = [
  { id: 'dashboard', label: 'Сводка' },
  { id: 'employees', label: 'Сотрудники' },
  { id: 'awards', label: 'Начисления' },
  { id: 'redemptions', label: 'Заявки' },
  { id: 'support', label: 'Поддержка' },
];

function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: ReactNode }) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-end justify-center z-50" onClick={onClose}>
      <div
        className="bg-white rounded-t-3xl w-full max-w-xl p-6 max-h-[85vh] overflow-y-auto"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl">{title}</h2>
          <button onClick={onClose} className="text-gray-400">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

export function ManagerPanel({ onNavigate }: ManagerPanelProps) {
  const [activeTab, setActiveTab] = useState<ManagerTab>('dashboard');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [dashboard, setDashboard] = useState<any>(null);
  const [employees, setEmployees] = useState<any[]>([]);
  const [selectedEmployee, setSelectedEmployee] = useState<any | null>(null);
  const [employeeDetail, setEmployeeDetail] = useState<any | null>(null);
  const [redemptions, setRedemptions] = useState<any[]>([]);
  const [supportTickets, setSupportTickets] = useState<any[]>([]);

  const [awardForm, setAwardForm] = useState({ employeeId: '', points: '', reason: '' });
  const [showAwardResult, setShowAwardResult] = useState<string | null>(null);

  const [showSupportResponse, setShowSupportResponse] = useState(false);
  const [supportTicket, setSupportTicket] = useState<any | null>(null);
  const [supportResponse, setSupportResponse] = useState({ response: '', status: 'resolved' });

  const loadDashboard = async () => {
    const data = await apiRequest<any>('/manager/dashboard');
    setDashboard(data);
  };

  const loadEmployees = async () => {
    const data = await apiRequest<any>('/manager/employees');
    setEmployees(data.employees || []);
  };

  const loadRedemptions = async () => {
    const data = await apiRequest<any>('/manager/redemptions');
    setRedemptions(data.redemptions || []);
  };

  const loadSupport = async () => {
    const data = await apiRequest<any>('/manager/support/tickets');
    setSupportTickets(data.tickets || []);
  };

  const loadActive = async (tab: ManagerTab) => {
    setLoading(true);
    setError(null);
    try {
      switch (tab) {
        case 'dashboard':
          await loadDashboard();
          break;
        case 'employees':
          await loadEmployees();
          break;
        case 'awards':
          await loadEmployees();
          break;
        case 'redemptions':
          await loadRedemptions();
          break;
        case 'support':
          await loadSupport();
          break;
        default:
          break;
      }
    } catch (err: any) {
      setError(err?.message || 'Ошибка загрузки');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadActive(activeTab);
  }, [activeTab]);

  const openEmployee = async (employee: any) => {
    setSelectedEmployee(employee);
    const data = await apiRequest<any>(`/manager/employees/${employee.id}`);
    setEmployeeDetail(data);
  };

  const submitAward = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!awardForm.employeeId || !awardForm.points) return;
    await apiRequest('/manager/award', {
      method: 'POST',
      body: JSON.stringify({
        employee_id: awardForm.employeeId,
        points: Number(awardForm.points),
        reason: awardForm.reason,
      }),
    });
    setShowAwardResult('Баллы успешно начислены');
    setAwardForm({ employeeId: '', points: '', reason: '' });
    await loadEmployees();
  };

  const approveRedemption = async (item: any) => {
    await apiRequest(`/manager/redemptions/${item.id}/approve`, { method: 'POST' });
    await loadRedemptions();
  };

  const rejectRedemption = async (item: any) => {
    await apiRequest(`/manager/redemptions/${item.id}/reject`, { method: 'POST' });
    await loadRedemptions();
  };

  const openSupportResponse = (ticket: any) => {
    setSupportTicket(ticket);
    setSupportResponse({ response: ticket.response || '', status: ticket.status || 'resolved' });
    setShowSupportResponse(true);
  };

  const saveSupportResponse = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!supportTicket) return;
    await apiRequest(`/manager/support/tickets/${supportTicket.id}/respond`, {
      method: 'POST',
      body: JSON.stringify({ response: supportResponse.response, status: supportResponse.status }),
    });
    setShowSupportResponse(false);
    await loadSupport();
  };

  const tabContent = useMemo(() => {
    switch (activeTab) {
      case 'dashboard':
        return (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-white rounded-2xl p-4 shadow-sm">
                <div className="text-xs text-gray-500">Сотрудников</div>
                <div className="text-2xl">{dashboard?.employees_count || 0}</div>
              </div>
              <div className="bg-white rounded-2xl p-4 shadow-sm">
                <div className="text-xs text-gray-500">Тренировок</div>
                <div className="text-2xl">{dashboard?.total_workouts || 0}</div>
              </div>
              <div className="bg-white rounded-2xl p-4 shadow-sm">
                <div className="text-xs text-gray-500">Часов активности</div>
                <div className="text-2xl">{Number(dashboard?.total_hours || 0).toFixed(1)}</div>
              </div>
            </div>
            <div className="bg-white rounded-2xl p-4 shadow-sm">
              <h3 className="mb-3">Топ сотрудников</h3>
              <div className="space-y-2">
                {(dashboard?.employees || []).map((employee: any) => (
                  <div key={employee.id} className="flex items-center justify-between text-sm">
                    <div>
                      <div>{employee.name}</div>
                      <div className="text-xs text-gray-500">{employee.department}</div>
                    </div>
                    <div className="text-gray-600">{employee.workouts} тренировок</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        );
      case 'employees':
        return (
          <div className="space-y-4">
            <h2 className="text-lg">Сотрудники</h2>
            <div className="grid gap-3">
              {employees.map((employee) => (
                <div key={employee.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{employee.name}</div>
                      <div className="text-xs text-gray-500">{employee.department}</div>
                    </div>
                    <button className="text-blue-600" onClick={() => openEmployee(employee)}>
                      Подробнее
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'awards':
        return (
          <div className="space-y-4">
            <h2 className="text-lg">Начисление баллов</h2>
            {showAwardResult && (
              <div className="bg-green-50 border border-green-200 text-green-700 rounded-xl p-4 text-sm">
                {showAwardResult}
              </div>
            )}
            <form onSubmit={submitAward} className="bg-white rounded-2xl p-4 shadow-sm space-y-3">
              <div>
                <label className="text-sm text-gray-600">Сотрудник</label>
                <select
                  value={awardForm.employeeId}
                  onChange={(event) => setAwardForm({ ...awardForm, employeeId: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  required
                >
                  <option value="">Выберите сотрудника</option>
                  {employees.map((employee) => (
                    <option key={employee.id} value={employee.id}>
                      {employee.name} ({employee.department})
                    </option>
                  ))}
                </select>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-sm text-gray-600">Баллы</label>
                  <input
                    type="number"
                    value={awardForm.points}
                    onChange={(event) => setAwardForm({ ...awardForm, points: event.target.value })}
                    className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                    required
                  />
                </div>
                <div>
                  <label className="text-sm text-gray-600">Причина</label>
                  <input
                    value={awardForm.reason}
                    onChange={(event) => setAwardForm({ ...awardForm, reason: event.target.value })}
                    className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  />
                </div>
              </div>
              <button type="submit" className="w-full bg-blue-600 text-white py-3 rounded-xl">
                Начислить баллы
              </button>
            </form>
          </div>
        );
      case 'redemptions':
        return (
          <div className="space-y-4">
            <h2 className="text-lg">Заявки на награды</h2>
            <div className="bg-white rounded-2xl shadow-sm overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-gray-600">
                  <tr>
                    <th className="px-4 py-3 text-left">Сотрудник</th>
                    <th className="px-4 py-3 text-left">Награда</th>
                    <th className="px-4 py-3 text-left">Статус</th>
                    <th className="px-4 py-3"></th>
                  </tr>
                </thead>
                <tbody>
                  {redemptions.map((item) => (
                    <tr key={item.id} className="border-t">
                      <td className="px-4 py-3">{item.user}</td>
                      <td className="px-4 py-3">{item.reward}</td>
                      <td className="px-4 py-3">{item.status}</td>
                      <td className="px-4 py-3 text-right">
                        {item.status === 'pending' && (
                          <div className="flex items-center gap-2 justify-end">
                            <button className="text-green-600" onClick={() => approveRedemption(item)}>
                              Одобрить
                            </button>
                            <button className="text-red-600" onClick={() => rejectRedemption(item)}>
                              Отклонить
                            </button>
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        );
      case 'support':
        return (
          <div className="space-y-4">
            <h2 className="text-lg">Обращения</h2>
            <div className="bg-white rounded-2xl shadow-sm overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-gray-600">
                  <tr>
                    <th className="px-4 py-3 text-left">Сотрудник</th>
                    <th className="px-4 py-3 text-left">Тема</th>
                    <th className="px-4 py-3 text-left">Статус</th>
                    <th className="px-4 py-3"></th>
                  </tr>
                </thead>
                <tbody>
                  {supportTickets.map((ticket) => (
                    <tr key={ticket.id} className="border-t">
                      <td className="px-4 py-3">{ticket.user}</td>
                      <td className="px-4 py-3">{ticket.subject}</td>
                      <td className="px-4 py-3">{ticket.status}</td>
                      <td className="px-4 py-3 text-right">
                        <button className="text-blue-600" onClick={() => openSupportResponse(ticket)}>
                          Ответить
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        );
      default:
        return null;
    }
  }, [activeTab, dashboard, employees, redemptions, supportTickets, awardForm, showAwardResult]);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-blue-900 text-white p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl">Панель менеджера</h1>
            <p className="text-blue-200 text-sm">Контроль команды и мотивации</p>
          </div>
          <button onClick={() => onNavigate('home')} className="text-blue-100 text-sm">
            На главную
          </button>
        </div>
      </div>

      <div className="p-6 space-y-6">
        <div className="flex gap-2 overflow-x-auto">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-4 py-2 rounded-full text-sm whitespace-nowrap ${
                activeTab === tab.id ? 'bg-blue-900 text-white' : 'bg-white border border-gray-200 text-gray-600'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        {loading ? <div className="text-sm text-gray-500">Загрузка...</div> : tabContent}
      </div>

      {selectedEmployee && employeeDetail && (
        <Modal title={employeeDetail.employee?.name || 'Сотрудник'} onClose={() => setSelectedEmployee(null)}>
          <div className="space-y-3 text-sm">
            <div>Табельный номер: {employeeDetail.employee?.employee_id}</div>
            <div>Отдел: {employeeDetail.employee?.department}</div>
            <div>Должность: {employeeDetail.employee?.position}</div>
            <div>Тренировок: {employeeDetail.workouts}</div>
            <div>Часов: {Number(employeeDetail.hours || 0).toFixed(1)}</div>
            <div>Баллы: {employeeDetail.points}</div>
          </div>
        </Modal>
      )}

      {showSupportResponse && supportTicket && (
        <Modal title={`Ответ для: ${supportTicket.subject}`} onClose={() => setShowSupportResponse(false)}>
          <form onSubmit={saveSupportResponse} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Ответ</label>
              <textarea
                value={supportResponse.response}
                onChange={(event) => setSupportResponse({ ...supportResponse, response: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={4}
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Статус</label>
              <select
                value={supportResponse.status}
                onChange={(event) => setSupportResponse({ ...supportResponse, status: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              >
                <option value="open">Открыт</option>
                <option value="in_progress">В работе</option>
                <option value="resolved">Решено</option>
              </select>
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowSupportResponse(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-blue-900 text-white">
                Отправить
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  );
}
