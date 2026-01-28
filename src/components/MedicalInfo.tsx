import { useEffect, useState } from 'react';
import type { Screen, MedicalInfo as MedicalInfoType } from '../types';
import { apiRequest } from '../api/client';

interface MedicalInfoProps {
  onNavigate: (screen: Screen) => void;
}

const emptyMedical: MedicalInfoType = {
  chronicDiseases: [],
  injuries: [],
  medications: [],
  allergies: [],
  doctorApproval: false,
  lastCheckup: '',
  restrictions: [],
};

export function MedicalInfo({ onNavigate }: MedicalInfoProps) {
  const [medicalData, setMedicalData] = useState<MedicalInfoType>(emptyMedical);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [contact, setContact] = useState<{ name: string; phone: string } | null>(null);
  const [showContactForm, setShowContactForm] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem('emergency_contact');
    if (stored) {
      try {
        setContact(JSON.parse(stored));
      } catch {
        setContact(null);
      }
    }
  }, []);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiRequest<any>('/medical-info');
      setMedicalData({
        chronicDiseases: data.chronic_diseases || [],
        injuries: data.injuries || [],
        medications: data.medications || [],
        allergies: data.allergies || [],
        doctorApproval: Boolean(data.doctor_approval),
        lastCheckup: data.last_checkup || '',
        restrictions: data.restrictions || [],
      });
    } catch (err: any) {
      setError(err?.message || 'Не удалось загрузить медицинские данные');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const save = async () => {
    setSaving(true);
    setError(null);
    try {
      await apiRequest('/medical-info', {
        method: 'PUT',
        body: JSON.stringify({
          chronic_diseases: medicalData.chronicDiseases,
          injuries: medicalData.injuries,
          medications: medicalData.medications,
          allergies: medicalData.allergies,
          doctor_approval: medicalData.doctorApproval,
          last_checkup: medicalData.lastCheckup,
          restrictions: medicalData.restrictions,
        }),
      });
      setEditing(false);
    } catch (err: any) {
      setError(err?.message || 'Не удалось сохранить данные');
    } finally {
      setSaving(false);
    }
  };

  const addItem = (key: keyof MedicalInfoType) => {
    if (!editing) {
      setEditing(true);
    }
    const list = medicalData[key] as string[];
    setMedicalData({ ...medicalData, [key]: [...list, ''] });
  };

  const updateItem = (key: keyof MedicalInfoType, index: number, value: string) => {
    const list = [...(medicalData[key] as string[])];
    list[index] = value;
    setMedicalData({ ...medicalData, [key]: list });
  };

  const removeItem = (key: keyof MedicalInfoType, index: number) => {
    const list = [...(medicalData[key] as string[])];
    list.splice(index, 1);
    setMedicalData({ ...medicalData, [key]: list });
  };

  const handleContactSave = (event: React.FormEvent) => {
    event.preventDefault();
    if (!contact?.name || !contact.phone) return;
    localStorage.setItem('emergency_contact', JSON.stringify(contact));
    setShowContactForm(false);
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <div className="text-sm text-slate-500">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-red-600 text-white p-6">
        <button
          onClick={() => onNavigate('profile')}
          className="flex items-center gap-2 mb-4"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          <span>Назад</span>
        </button>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl mb-2">Медицинская информация</h1>
            <p className="text-red-100">Данные о здоровье и ограничениях</p>
          </div>
          <button
            className="text-white/90 bg-white/10 px-3 py-2 rounded-lg"
            onClick={() => setEditing((prev) => !prev)}
          >
            {editing ? 'Отмена' : 'Редактировать'}
          </button>
        </div>
      </div>

      <div className="p-6 space-y-4">
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        <div className="bg-red-50 border border-red-200 rounded-2xl p-5">
          <div className="flex items-start gap-3">
            <svg className="w-6 h-6 text-red-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <h4 className="text-red-900 mb-1">Конфиденциальность</h4>
              <p className="text-sm text-red-700">
                Медицинская информация защищена и используется только для адаптации программы тренировок.
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${
                medicalData.doctorApproval ? 'bg-green-100' : 'bg-amber-100'
              }`}>
                <svg className={`w-6 h-6 ${medicalData.doctorApproval ? 'text-green-600' : 'text-amber-600'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  {medicalData.doctorApproval ? (
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
                  ) : (
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  )}
                </svg>
              </div>
              <div>
                <h4>Допуск врача</h4>
                <p className="text-sm text-gray-600">
                  {medicalData.doctorApproval ? 'Получен' : 'Не получен'}
                </p>
              </div>
            </div>
            {editing && (
              <button
                className="text-blue-600"
                onClick={() => setMedicalData({ ...medicalData, doctorApproval: !medicalData.doctorApproval })}
              >
                {medicalData.doctorApproval ? 'Отменить' : 'Подтвердить'}
              </button>
            )}
          </div>
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="bg-blue-100 w-12 h-12 rounded-xl flex items-center justify-center">
                <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
              <div>
                <h4>Последний медосмотр</h4>
                <p className="text-sm text-gray-600">
                  {medicalData.lastCheckup
                    ? new Date(medicalData.lastCheckup).toLocaleDateString('ru-RU', {
                        day: 'numeric',
                        month: 'long',
                        year: 'numeric',
                      })
                    : 'Не указано'}
                </p>
              </div>
            </div>
            {editing && (
              <input
                type="date"
                value={medicalData.lastCheckup || ''}
                onChange={(event) => setMedicalData({ ...medicalData, lastCheckup: event.target.value })}
                className="text-sm px-2 py-1 border border-gray-200 rounded-lg"
              />
            )}
          </div>
        </div>

        {(
          [
            { key: 'chronicDiseases', title: 'Хронические заболевания' },
            { key: 'injuries', title: 'Травмы и операции' },
            { key: 'medications', title: 'Принимаемые препараты' },
            { key: 'allergies', title: 'Аллергии' },
          ] as const
        ).map((section) => {
          const items = medicalData[section.key] as string[];
          return (
            <div key={section.key} className="bg-white rounded-2xl p-5 shadow-sm">
              <div className="flex items-center justify-between mb-3">
                <h3>{section.title}</h3>
                <button className="text-blue-600 text-sm" onClick={() => addItem(section.key)}>
                  Добавить
                </button>
              </div>
              {items.length > 0 ? (
                <div className="space-y-2">
                  {items.map((item, index) => (
                    <div key={index} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
                      {editing ? (
                        <input
                          value={item}
                          onChange={(event) => updateItem(section.key, index, event.target.value)}
                          className="flex-1 bg-transparent outline-none text-sm"
                        />
                      ) : (
                        <span>{item}</span>
                      )}
                      <button className="text-red-600" onClick={() => removeItem(section.key, index)}>
                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-gray-500">Не указано</p>
              )}
            </div>
          );
        })}

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <div className="flex items-center justify-between mb-3">
            <h3>Ограничения по нагрузке</h3>
            <button className="text-blue-600 text-sm" onClick={() => addItem('restrictions')}>
              Добавить
            </button>
          </div>
          {medicalData.restrictions.length > 0 ? (
            <div className="space-y-2">
              {medicalData.restrictions.map((restriction, index) => (
                <div key={index} className="p-3 bg-amber-50 border border-amber-200 rounded-xl">
                  <div className="flex items-start gap-2">
                    <svg className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                    {editing ? (
                      <input
                        value={restriction}
                        onChange={(event) => updateItem('restrictions', index, event.target.value)}
                        className="flex-1 bg-transparent outline-none text-sm text-amber-900"
                      />
                    ) : (
                      <span className="text-sm text-amber-900">{restriction}</span>
                    )}
                    <button className="text-red-600" onClick={() => removeItem('restrictions', index)}>
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500">Не указано</p>
          )}
        </div>

        <div className="bg-white rounded-2xl p-5 shadow-sm">
          <h3 className="mb-3">Экстренный контакт</h3>
          {contact ? (
            <div className="flex items-center justify-between p-3 bg-red-50 rounded-xl">
              <div>
                <div className="text-sm">{contact.name}</div>
                <div className="text-xs text-gray-600">{contact.phone}</div>
              </div>
              <button
                className="text-red-600"
                onClick={() => setShowContactForm(true)}
              >
                Изменить
              </button>
            </div>
          ) : (
            <button
              className="w-full py-3 bg-red-50 text-red-600 rounded-xl border-2 border-red-200 hover:bg-red-100 transition-colors flex items-center justify-center gap-2"
              onClick={() => setShowContactForm(true)}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
              </svg>
              Добавить экстренный контакт
            </button>
          )}
        </div>

        {editing && (
          <button
            onClick={save}
            disabled={saving}
            className="w-full bg-red-600 text-white py-4 rounded-xl disabled:opacity-60"
          >
            {saving ? 'Сохраняем...' : 'Сохранить изменения'}
          </button>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded-2xl p-5">
          <div className="flex items-start gap-3">
            <svg className="w-6 h-6 text-blue-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <h4 className="text-blue-900 mb-1">Почему это важно?</h4>
              <p className="text-sm text-blue-700">
                Полная медицинская информация позволяет системе адаптировать программу тренировок с учётом ваших особенностей и обеспечить максимальную безопасность.
              </p>
            </div>
          </div>
        </div>
      </div>

      {showContactForm && (
        <div className="fixed inset-0 bg-black/40 flex items-end justify-center z-50" onClick={() => setShowContactForm(false)}>
          <div
            className="bg-white rounded-t-3xl w-full max-w-md p-6"
            onClick={(event) => event.stopPropagation()}
          >
            <h2 className="text-xl mb-4">Экстренный контакт</h2>
            <form onSubmit={handleContactSave} className="space-y-4">
              <div>
                <label className="text-sm text-gray-600">Имя</label>
                <input
                  value={contact?.name || ''}
                  onChange={(event) => setContact({ name: event.target.value, phone: contact?.phone || '' })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  required
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Телефон</label>
                <input
                  value={contact?.phone || ''}
                  onChange={(event) => setContact({ name: contact?.name || '', phone: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  required
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  className="flex-1 py-3 rounded-xl border border-gray-200"
                  onClick={() => setShowContactForm(false)}
                >
                  Отмена
                </button>
                <button type="submit" className="flex-1 py-3 rounded-xl bg-red-600 text-white">
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
