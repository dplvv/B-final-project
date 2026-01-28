import { useEffect, useMemo, useState, type ReactNode } from 'react';
import type { Screen } from '../../types';
import { apiRequest } from '../../api/client';

interface AdminPanelProps {
  onNavigate: (screen: Screen) => void;
}

type AdminTab =
  | 'users'
  | 'exercises'
  | 'workouts'
  | 'programs'
  | 'recommendations'
  | 'videos'
  | 'nutrition'
  | 'rewards'
  | 'support'
  | 'redemptions';

const tabs: { id: AdminTab; label: string }[] = [
  { id: 'users', label: 'Пользователи' },
  { id: 'exercises', label: 'Упражнения' },
  { id: 'workouts', label: 'Тренировки' },
  { id: 'programs', label: 'Программы' },
  { id: 'recommendations', label: 'Рекомендации' },
  { id: 'videos', label: 'Видео' },
  { id: 'nutrition', label: 'Питание' },
  { id: 'rewards', label: 'Награды' },
  { id: 'support', label: 'Поддержка' },
  { id: 'redemptions', label: 'Заявки' },
];

const parseList = (value: string) =>
  value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);

function Modal({ title, onClose, children }: { title: string; onClose: () => void; children: ReactNode }) {
  return (
    <div className="fixed inset-0 bg-black/40 flex items-end justify-center z-50" onClick={onClose}>
      <div
        className="bg-white rounded-t-3xl w-full max-w-2xl p-6 max-h-[85vh] overflow-y-auto"
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

export function AdminPanel({ onNavigate }: AdminPanelProps) {
  const [activeTab, setActiveTab] = useState<AdminTab>('users');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [users, setUsers] = useState<any[]>([]);
  const [exercises, setExercises] = useState<any[]>([]);
  const [workouts, setWorkouts] = useState<any[]>([]);
  const [programs, setPrograms] = useState<any[]>([]);
  const [recommendations, setRecommendations] = useState<any[]>([]);
  const [videos, setVideos] = useState<any[]>([]);
  const [nutritionItems, setNutritionItems] = useState<any[]>([]);
  const [rewards, setRewards] = useState<any[]>([]);
  const [supportTickets, setSupportTickets] = useState<any[]>([]);
  const [redemptions, setRedemptions] = useState<any[]>([]);

  const [showUserForm, setShowUserForm] = useState(false);
  const [editingUser, setEditingUser] = useState<any | null>(null);
  const [userForm, setUserForm] = useState({
    name: '',
    employeeId: '',
    department: '',
    position: '',
    role: 'employee',
    password: '',
  });

  const [showExerciseForm, setShowExerciseForm] = useState(false);
  const [editingExercise, setEditingExercise] = useState<any | null>(null);
  const [exerciseForm, setExerciseForm] = useState({
    name: '',
    description: '',
    category: '',
    difficulty: '',
    sets: '',
    reps: '',
    duration: '',
    rest: '',
    muscleGroups: '',
    equipment: '',
    videoUrl: '',
  });

  const [showWorkoutForm, setShowWorkoutForm] = useState(false);
  const [editingWorkout, setEditingWorkout] = useState<any | null>(null);
  const [workoutForm, setWorkoutForm] = useState({
    name: '',
    description: '',
    duration: '',
    difficulty: '',
    category: '',
  });

  const [showWorkoutExercises, setShowWorkoutExercises] = useState(false);
  const [workoutExercises, setWorkoutExercises] = useState<any[]>([]);
  const [workoutForComposition, setWorkoutForComposition] = useState<any | null>(null);
  const [allExercises, setAllExercises] = useState<any[]>([]);

  const [showProgramForm, setShowProgramForm] = useState(false);
  const [editingProgram, setEditingProgram] = useState<any | null>(null);
  const [programForm, setProgramForm] = useState({
    name: '',
    description: '',
    active: true,
  });

  const [showProgramWorkouts, setShowProgramWorkouts] = useState(false);
  const [programWorkouts, setProgramWorkouts] = useState<any[]>([]);
  const [programForComposition, setProgramForComposition] = useState<any | null>(null);
  const [allWorkouts, setAllWorkouts] = useState<any[]>([]);

  const [showRecommendationForm, setShowRecommendationForm] = useState(false);
  const [editingRecommendation, setEditingRecommendation] = useState<any | null>(null);
  const [recommendationForm, setRecommendationForm] = useState({
    title: '',
    body: '',
    category: '',
    icon: '',
    excerpt: '',
    readTime: '5',
  });

  const [showVideoForm, setShowVideoForm] = useState(false);
  const [editingVideo, setEditingVideo] = useState<any | null>(null);
  const [videoForm, setVideoForm] = useState({
    title: '',
    description: '',
    duration: '',
    category: '',
    difficulty: '',
    url: '',
  });

  const [showNutritionForm, setShowNutritionForm] = useState(false);
  const [editingNutrition, setEditingNutrition] = useState<any | null>(null);
  const [nutritionForm, setNutritionForm] = useState({
    title: '',
    description: '',
    calories: '',
    category: '',
  });

  const [showRewardForm, setShowRewardForm] = useState(false);
  const [editingReward, setEditingReward] = useState<any | null>(null);
  const [rewardForm, setRewardForm] = useState({
    title: '',
    description: '',
    pointsCost: '',
    category: '',
    active: true,
  });

  const [showSupportResponse, setShowSupportResponse] = useState(false);
  const [supportTicket, setSupportTicket] = useState<any | null>(null);
  const [supportResponse, setSupportResponse] = useState({ response: '', status: 'resolved' });

  const [actionMessage, setActionMessage] = useState<string | null>(null);

  const loadUsers = async () => {
    const data = await apiRequest<any>('/admin/users');
    setUsers(data.users || []);
  };

  const loadExercises = async () => {
    const data = await apiRequest<any>('/admin/content/exercises');
    setExercises(data.exercises || []);
  };

  const loadWorkouts = async () => {
    const data = await apiRequest<any>('/admin/content/workouts');
    setWorkouts(data.workouts || []);
  };

  const loadPrograms = async () => {
    const data = await apiRequest<any>('/admin/content/programs');
    setPrograms(data.programs || []);
  };

  const loadRecommendations = async () => {
    const data = await apiRequest<any>('/admin/content/recommendations');
    setRecommendations(data.recommendations || []);
  };

  const loadVideos = async () => {
    const data = await apiRequest<any>('/admin/content/videos');
    setVideos(data.videos || []);
  };

  const loadNutrition = async () => {
    const data = await apiRequest<any>('/admin/content/nutrition');
    setNutritionItems(data.items || []);
  };

  const loadRewards = async () => {
    const data = await apiRequest<any>('/admin/content/rewards');
    setRewards(data.rewards || []);
  };

  const loadSupport = async () => {
    const data = await apiRequest<any>('/admin/support/tickets');
    setSupportTickets(data.tickets || []);
  };

  const loadRedemptions = async () => {
    const data = await apiRequest<any>('/admin/redemptions');
    setRedemptions(data.redemptions || []);
  };

  const loadActive = async (tab: AdminTab) => {
    setLoading(true);
    setError(null);
    try {
      switch (tab) {
        case 'users':
          await loadUsers();
          break;
        case 'exercises':
          await loadExercises();
          break;
        case 'workouts':
          await loadWorkouts();
          break;
        case 'programs':
          await loadPrograms();
          break;
        case 'recommendations':
          await loadRecommendations();
          break;
        case 'videos':
          await loadVideos();
          break;
        case 'nutrition':
          await loadNutrition();
          break;
        case 'rewards':
          await loadRewards();
          break;
        case 'support':
          await loadSupport();
          break;
        case 'redemptions':
          await loadRedemptions();
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

  const openUserForm = (user?: any) => {
    if (user) {
      setEditingUser(user);
      setUserForm({
        name: user.name || '',
        employeeId: user.employee_id || '',
        department: user.department || '',
        position: user.position || '',
        role: user.role || 'employee',
        password: '',
      });
    } else {
      setEditingUser(null);
      setUserForm({ name: '', employeeId: '', department: '', position: '', role: 'employee', password: '' });
    }
    setShowUserForm(true);
  };

  const saveUser = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!userForm.name || !userForm.employeeId) return;
    if (!editingUser && !userForm.password) return;

    if (editingUser) {
      await apiRequest(`/admin/users/${editingUser.id}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: userForm.name,
          employee_id: userForm.employeeId,
          department: userForm.department,
          position: userForm.position,
          role: userForm.role,
          password: userForm.password,
        }),
      });
    } else {
      await apiRequest('/admin/users', {
        method: 'POST',
        body: JSON.stringify({
          name: userForm.name,
          employee_id: userForm.employeeId,
          department: userForm.department,
          position: userForm.position,
          role: userForm.role,
          password: userForm.password,
        }),
      });
    }
    setShowUserForm(false);
    await loadUsers();
  };

  const resetUserPassword = async (user: any) => {
    const newPassword = prompt('Введите новый пароль (или оставьте пустым для генерации):') || '';
    const data = await apiRequest<any>(`/admin/users/${user.id}/reset-password`, {
      method: 'POST',
      body: JSON.stringify({ password: newPassword }),
    });
    setActionMessage(`Новый пароль для ${user.name}: ${data.password}`);
  };

  const openExerciseForm = (exercise?: any) => {
    if (exercise) {
      setEditingExercise(exercise);
      setExerciseForm({
        name: exercise.name || '',
        description: exercise.description || '',
        category: exercise.category || '',
        difficulty: exercise.difficulty || '',
        sets: exercise.sets ? String(exercise.sets) : '',
        reps: exercise.reps || '',
        duration: exercise.duration ? String(exercise.duration) : '',
        rest: exercise.rest ? String(exercise.rest) : '',
        muscleGroups: (exercise.muscle_groups || []).join(', '),
        equipment: (exercise.equipment || []).join(', '),
        videoUrl: exercise.video_url || '',
      });
    } else {
      setEditingExercise(null);
      setExerciseForm({
        name: '',
        description: '',
        category: '',
        difficulty: '',
        sets: '',
        reps: '',
        duration: '',
        rest: '',
        muscleGroups: '',
        equipment: '',
        videoUrl: '',
      });
    }
    setShowExerciseForm(true);
  };

  const saveExercise = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!exerciseForm.name || !exerciseForm.description) return;

    const payload = {
      name: exerciseForm.name,
      description: exerciseForm.description,
      category: exerciseForm.category,
      difficulty: exerciseForm.difficulty,
      sets: Number(exerciseForm.sets || 0),
      reps: exerciseForm.reps,
      duration_seconds: Number(exerciseForm.duration || 0),
      rest_seconds: Number(exerciseForm.rest || 0),
      muscle_groups: parseList(exerciseForm.muscleGroups),
      equipment: parseList(exerciseForm.equipment),
      video_url: exerciseForm.videoUrl,
    };

    if (editingExercise) {
      await apiRequest(`/admin/content/exercises/${editingExercise.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/exercises', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowExerciseForm(false);
    await loadExercises();
  };

  const deleteExercise = async (exercise: any) => {
    if (!confirm('Удалить упражнение?')) return;
    await apiRequest(`/admin/content/exercises/${exercise.id}`, { method: 'DELETE' });
    await loadExercises();
  };

  const openWorkoutForm = (workout?: any) => {
    if (workout) {
      setEditingWorkout(workout);
      setWorkoutForm({
        name: workout.name || '',
        description: workout.description || '',
        duration: workout.duration_minutes ? String(workout.duration_minutes) : '',
        difficulty: workout.difficulty || '',
        category: workout.category || '',
      });
    } else {
      setEditingWorkout(null);
      setWorkoutForm({ name: '', description: '', duration: '', difficulty: '', category: '' });
    }
    setShowWorkoutForm(true);
  };

  const saveWorkout = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!workoutForm.name) return;

    const payload = {
      name: workoutForm.name,
      description: workoutForm.description,
      duration_minutes: Number(workoutForm.duration || 0),
      difficulty: workoutForm.difficulty,
      category: workoutForm.category,
    };

    if (editingWorkout) {
      await apiRequest(`/admin/content/workouts/${editingWorkout.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/workouts', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowWorkoutForm(false);
    await loadWorkouts();
  };

  const deleteWorkout = async (workout: any) => {
    if (!confirm('Удалить тренировку?')) return;
    await apiRequest(`/admin/content/workouts/${workout.id}`, { method: 'DELETE' });
    await loadWorkouts();
  };

  const openWorkoutComposition = async (workout: any) => {
    setWorkoutForComposition(workout);
    const exerciseData = await apiRequest<any>('/admin/content/exercises');
    setAllExercises(exerciseData.exercises || []);
    const detail = await apiRequest<any>(`/workouts/${workout.id}`);
    const existing = (detail.workout?.exercises || []).map((ex: any, index: number) => ({
      exercise_id: ex.id,
      sort_order: index + 1,
      sets: ex.sets || 0,
      reps: ex.reps || '',
      duration_seconds: ex.duration || 0,
      rest_seconds: ex.rest || 0,
    }));
    setWorkoutExercises(existing.length ? existing : [{
      exercise_id: '',
      sort_order: 1,
      sets: 0,
      reps: '',
      duration_seconds: 0,
      rest_seconds: 0,
    }]);
    setShowWorkoutExercises(true);
  };

  const saveWorkoutComposition = async () => {
    if (!workoutForComposition) return;
    await apiRequest(`/admin/content/workouts/${workoutForComposition.id}/exercises`, {
      method: 'POST',
      body: JSON.stringify(
        workoutExercises
          .filter((item) => item.exercise_id)
          .map((item, index) => ({
            exercise_id: item.exercise_id,
            sort_order: Number(item.sort_order || index + 1),
            sets: Number(item.sets || 0),
            reps: item.reps || '',
            duration_seconds: Number(item.duration_seconds || 0),
            rest_seconds: Number(item.rest_seconds || 0),
          }))
      ),
    });
    setShowWorkoutExercises(false);
    await loadWorkouts();
  };

  const openProgramForm = (program?: any) => {
    if (program) {
      setEditingProgram(program);
      setProgramForm({
        name: program.name || '',
        description: program.description || '',
        active: Boolean(program.active),
      });
    } else {
      setEditingProgram(null);
      setProgramForm({ name: '', description: '', active: true });
    }
    setShowProgramForm(true);
  };

  const saveProgram = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!programForm.name) return;

    const payload = {
      name: programForm.name,
      description: programForm.description,
      active: programForm.active,
    };

    if (editingProgram) {
      await apiRequest(`/admin/content/programs/${editingProgram.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/programs', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowProgramForm(false);
    await loadPrograms();
  };

  const deleteProgram = async (program: any) => {
    if (!confirm('Удалить программу?')) return;
    await apiRequest(`/admin/content/programs/${program.id}`, { method: 'DELETE' });
    await loadPrograms();
  };

  const openProgramComposition = async (program: any) => {
    setProgramForComposition(program);
    const workoutData = await apiRequest<any>('/admin/content/workouts');
    setAllWorkouts(workoutData.workouts || []);
    const current = await apiRequest<any>(`/admin/content/programs/${program.id}/workouts`);
    const existing = (current.workouts || []).map((item: any) => ({
      workout_id: item.workout_id,
      sort_order: item.sort_order,
    }));
    setProgramWorkouts(existing.length ? existing : [{ workout_id: '', sort_order: 1 }]);
    setShowProgramWorkouts(true);
  };

  const saveProgramComposition = async () => {
    if (!programForComposition) return;
    await apiRequest(`/admin/content/programs/${programForComposition.id}/workouts`, {
      method: 'POST',
      body: JSON.stringify(
        programWorkouts
          .filter((item) => item.workout_id)
          .map((item, index) => ({
            workout_id: item.workout_id,
            sort_order: Number(item.sort_order || index + 1),
          }))
      ),
    });
    setShowProgramWorkouts(false);
    await loadPrograms();
  };

  const openRecommendationForm = (item?: any) => {
    if (item) {
      setEditingRecommendation(item);
      setRecommendationForm({
        title: item.title || '',
        body: item.body || '',
        category: item.category || '',
        icon: item.icon || '',
        excerpt: item.excerpt || '',
        readTime: String(item.read_time || 5),
      });
    } else {
      setEditingRecommendation(null);
      setRecommendationForm({ title: '', body: '', category: '', icon: '', excerpt: '', readTime: '5' });
    }
    setShowRecommendationForm(true);
  };

  const saveRecommendation = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!recommendationForm.title || !recommendationForm.body) return;

    const payload = {
      title: recommendationForm.title,
      body: recommendationForm.body,
      category: recommendationForm.category,
      icon: recommendationForm.icon,
      excerpt: recommendationForm.excerpt,
      read_time: Number(recommendationForm.readTime || 5),
    };

    if (editingRecommendation) {
      await apiRequest(`/admin/content/recommendations/${editingRecommendation.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/recommendations', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowRecommendationForm(false);
    await loadRecommendations();
  };

  const deleteRecommendation = async (item: any) => {
    if (!confirm('Удалить рекомендацию?')) return;
    await apiRequest(`/admin/content/recommendations/${item.id}`, { method: 'DELETE' });
    await loadRecommendations();
  };

  const openVideoForm = (item?: any) => {
    if (item) {
      setEditingVideo(item);
      setVideoForm({
        title: item.title || '',
        description: item.description || '',
        duration: String(item.duration || 0),
        category: item.category || '',
        difficulty: item.difficulty || '',
        url: item.url || '',
      });
    } else {
      setEditingVideo(null);
      setVideoForm({ title: '', description: '', duration: '', category: '', difficulty: '', url: '' });
    }
    setShowVideoForm(true);
  };

  const saveVideo = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!videoForm.title || !videoForm.description) return;

    const payload = {
      title: videoForm.title,
      description: videoForm.description,
      duration_minutes: Number(videoForm.duration || 0),
      category: videoForm.category,
      difficulty: videoForm.difficulty,
      url: videoForm.url,
    };

    if (editingVideo) {
      await apiRequest(`/admin/content/videos/${editingVideo.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/videos', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowVideoForm(false);
    await loadVideos();
  };

  const deleteVideo = async (item: any) => {
    if (!confirm('Удалить видео?')) return;
    await apiRequest(`/admin/content/videos/${item.id}`, { method: 'DELETE' });
    await loadVideos();
  };

  const openNutritionForm = (item?: any) => {
    if (item) {
      setEditingNutrition(item);
      setNutritionForm({
        title: item.title || '',
        description: item.description || '',
        calories: String(item.calories || 0),
        category: item.category || '',
      });
    } else {
      setEditingNutrition(null);
      setNutritionForm({ title: '', description: '', calories: '', category: '' });
    }
    setShowNutritionForm(true);
  };

  const saveNutrition = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!nutritionForm.title || !nutritionForm.description) return;

    const payload = {
      title: nutritionForm.title,
      description: nutritionForm.description,
      calories: Number(nutritionForm.calories || 0),
      category: nutritionForm.category,
    };

    if (editingNutrition) {
      await apiRequest(`/admin/content/nutrition/${editingNutrition.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/nutrition', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowNutritionForm(false);
    await loadNutrition();
  };

  const deleteNutrition = async (item: any) => {
    if (!confirm('Удалить запись?')) return;
    await apiRequest(`/admin/content/nutrition/${item.id}`, { method: 'DELETE' });
    await loadNutrition();
  };

  const openRewardForm = (item?: any) => {
    if (item) {
      setEditingReward(item);
      setRewardForm({
        title: item.title || '',
        description: item.description || '',
        pointsCost: String(item.points_cost || 0),
        category: item.category || '',
        active: Boolean(item.active),
      });
    } else {
      setEditingReward(null);
      setRewardForm({ title: '', description: '', pointsCost: '', category: '', active: true });
    }
    setShowRewardForm(true);
  };

  const saveReward = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!rewardForm.title || !rewardForm.description) return;

    const payload = {
      title: rewardForm.title,
      description: rewardForm.description,
      points_cost: Number(rewardForm.pointsCost || 0),
      category: rewardForm.category,
      active: rewardForm.active,
    };

    if (editingReward) {
      await apiRequest(`/admin/content/rewards/${editingReward.id}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
      });
    } else {
      await apiRequest('/admin/content/rewards', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
    }
    setShowRewardForm(false);
    await loadRewards();
  };

  const deleteReward = async (item: any) => {
    if (!confirm('Удалить награду?')) return;
    await apiRequest(`/admin/content/rewards/${item.id}`, { method: 'DELETE' });
    await loadRewards();
  };

  const openSupportResponse = (ticket: any) => {
    setSupportTicket(ticket);
    setSupportResponse({ response: ticket.response || '', status: ticket.status || 'resolved' });
    setShowSupportResponse(true);
  };

  const saveSupportResponse = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!supportTicket) return;
    await apiRequest(`/admin/support/tickets/${supportTicket.id}/respond`, {
      method: 'POST',
      body: JSON.stringify({ response: supportResponse.response, status: supportResponse.status }),
    });
    setShowSupportResponse(false);
    await loadSupport();
  };

  const approveRedemption = async (item: any) => {
    await apiRequest(`/admin/redemptions/${item.id}/approve`, { method: 'POST' });
    await loadRedemptions();
  };

  const rejectRedemption = async (item: any) => {
    await apiRequest(`/admin/redemptions/${item.id}/reject`, { method: 'POST' });
    await loadRedemptions();
  };

  const tabContent = useMemo(() => {
    switch (activeTab) {
      case 'users':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Пользователи</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openUserForm()}>
                Добавить
              </button>
            </div>
            <div className="bg-white rounded-2xl shadow-sm overflow-hidden">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-gray-600">
                  <tr>
                    <th className="px-4 py-3 text-left">Сотрудник</th>
                    <th className="px-4 py-3 text-left">Отдел</th>
                    <th className="px-4 py-3 text-left">Роль</th>
                    <th className="px-4 py-3 text-left">Баллы</th>
                    <th className="px-4 py-3"></th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id} className="border-t">
                      <td className="px-4 py-3">
                        <div>{user.name}</div>
                        <div className="text-xs text-gray-500">{user.employee_id}</div>
                      </td>
                      <td className="px-4 py-3">{user.department || '—'}</td>
                      <td className="px-4 py-3">{user.role}</td>
                      <td className="px-4 py-3">{user.points}</td>
                      <td className="px-4 py-3 text-right">
                        <div className="flex items-center gap-2 justify-end">
                          <button className="text-blue-600" onClick={() => openUserForm(user)}>
                            Редактировать
                          </button>
                          <button className="text-gray-600" onClick={() => resetUserPassword(user)}>
                            Пароль
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        );
      case 'exercises':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Упражнения</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openExerciseForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {exercises.map((exercise) => (
                <div key={exercise.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{exercise.name}</div>
                      <div className="text-xs text-gray-500">{exercise.category} • {exercise.difficulty}</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openExerciseForm(exercise)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteExercise(exercise)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'workouts':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Тренировки</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openWorkoutForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {workouts.map((workout) => (
                <div key={workout.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{workout.name}</div>
                      <div className="text-xs text-gray-500">
                        {workout.duration_minutes} мин • {workout.difficulty} • {workout.category}
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openWorkoutComposition(workout)}>
                        Состав
                      </button>
                      <button className="text-blue-600" onClick={() => openWorkoutForm(workout)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteWorkout(workout)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'programs':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Программы</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openProgramForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {programs.map((program) => (
                <div key={program.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{program.name}</div>
                      <div className="text-xs text-gray-500">
                        {program.active ? 'Активна' : 'Не активна'} • {program.workouts} тренировок
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openProgramComposition(program)}>
                        Состав
                      </button>
                      <button className="text-blue-600" onClick={() => openProgramForm(program)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteProgram(program)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'recommendations':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Рекомендации</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openRecommendationForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {recommendations.map((item) => (
                <div key={item.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{item.title}</div>
                      <div className="text-xs text-gray-500">{item.category} • {item.read_time} мин</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openRecommendationForm(item)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteRecommendation(item)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'videos':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Видео</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openVideoForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {videos.map((item) => (
                <div key={item.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{item.title}</div>
                      <div className="text-xs text-gray-500">{item.category} • {item.difficulty}</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openVideoForm(item)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteVideo(item)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'nutrition':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Питание</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openNutritionForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {nutritionItems.map((item) => (
                <div key={item.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{item.title}</div>
                      <div className="text-xs text-gray-500">{item.category} • {item.calories} ккал</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openNutritionForm(item)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteNutrition(item)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        );
      case 'rewards':
        return (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-lg">Награды</h2>
              <button className="bg-blue-600 text-white px-4 py-2 rounded-lg" onClick={() => openRewardForm()}>
                Добавить
              </button>
            </div>
            <div className="grid gap-3">
              {rewards.map((item) => (
                <div key={item.id} className="bg-white rounded-2xl p-4 shadow-sm">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium">{item.title}</div>
                      <div className="text-xs text-gray-500">{item.points_cost} баллов • {item.category}</div>
                    </div>
                    <div className="flex items-center gap-3">
                      <button className="text-blue-600" onClick={() => openRewardForm(item)}>
                        Редактировать
                      </button>
                      <button className="text-red-600" onClick={() => deleteReward(item)}>
                        Удалить
                      </button>
                    </div>
                  </div>
                </div>
              ))}
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
                      <td className="px-4 py-3">
                        <div>{ticket.user}</div>
                        <div className="text-xs text-gray-500">{ticket.category}</div>
                      </td>
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
                      <td className="px-4 py-3">
                        {item.reward} • {item.points} баллов
                      </td>
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
      default:
        return null;
    }
  }, [activeTab, users, exercises, workouts, programs, recommendations, videos, nutritionItems, rewards, supportTickets, redemptions]);

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-slate-900 text-white p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl">Панель администратора</h1>
            <p className="text-slate-300 text-sm">Управление пользователями и контентом</p>
          </div>
          <button onClick={() => onNavigate('home')} className="text-slate-200 text-sm">
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
                activeTab === tab.id
                  ? 'bg-slate-900 text-white'
                  : 'bg-white border border-gray-200 text-gray-600'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {actionMessage && (
          <div className="bg-green-50 border border-green-200 text-green-700 rounded-xl p-4 text-sm">
            {actionMessage}
          </div>
        )}

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 rounded-xl p-4 text-sm">
            {error}
          </div>
        )}

        {loading ? <div className="text-sm text-gray-500">Загрузка...</div> : tabContent}
      </div>

      {showUserForm && (
        <Modal title={editingUser ? 'Редактировать пользователя' : 'Новый пользователь'} onClose={() => setShowUserForm(false)}>
          <form onSubmit={saveUser} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">ФИО</label>
              <input
                value={userForm.name}
                onChange={(event) => setUserForm({ ...userForm, name: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Табельный номер</label>
              <input
                value={userForm.employeeId}
                onChange={(event) => setUserForm({ ...userForm, employeeId: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Отдел</label>
                <input
                  value={userForm.department}
                  onChange={(event) => setUserForm({ ...userForm, department: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Должность</label>
                <input
                  value={userForm.position}
                  onChange={(event) => setUserForm({ ...userForm, position: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Роль</label>
                <select
                  value={userForm.role}
                  onChange={(event) => setUserForm({ ...userForm, role: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                >
                  <option value="employee">Сотрудник</option>
                  <option value="manager">Менеджер</option>
                  <option value="admin">Администратор</option>
                </select>
              </div>
              <div>
                <label className="text-sm text-gray-600">Пароль</label>
                <input
                  type="password"
                  value={userForm.password}
                  onChange={(event) => setUserForm({ ...userForm, password: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                  required={!editingUser}
                />
              </div>
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowUserForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showExerciseForm && (
        <Modal title={editingExercise ? 'Редактировать упражнение' : 'Новое упражнение'} onClose={() => setShowExerciseForm(false)}>
          <form onSubmit={saveExercise} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={exerciseForm.name}
                onChange={(event) => setExerciseForm({ ...exerciseForm, name: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Описание</label>
              <textarea
                value={exerciseForm.description}
                onChange={(event) => setExerciseForm({ ...exerciseForm, description: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={3}
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Категория</label>
                <input
                  value={exerciseForm.category}
                  onChange={(event) => setExerciseForm({ ...exerciseForm, category: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Сложность</label>
                <input
                  value={exerciseForm.difficulty}
                  onChange={(event) => setExerciseForm({ ...exerciseForm, difficulty: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Подходы</label>
                <input
                  type="number"
                  value={exerciseForm.sets}
                  onChange={(event) => setExerciseForm({ ...exerciseForm, sets: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Повторы</label>
                <input
                  value={exerciseForm.reps}
                  onChange={(event) => setExerciseForm({ ...exerciseForm, reps: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Длительность (сек)</label>
                <input
                  type="number"
                  value={exerciseForm.duration}
                  onChange={(event) => setExerciseForm({ ...exerciseForm, duration: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Отдых (сек)</label>
                <input
                  type="number"
                  value={exerciseForm.rest}
                  onChange={(event) => setExerciseForm({ ...exerciseForm, rest: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div>
              <label className="text-sm text-gray-600">Мышечные группы (через запятую)</label>
              <input
                value={exerciseForm.muscleGroups}
                onChange={(event) => setExerciseForm({ ...exerciseForm, muscleGroups: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Оборудование (через запятую)</label>
              <input
                value={exerciseForm.equipment}
                onChange={(event) => setExerciseForm({ ...exerciseForm, equipment: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Видео URL</label>
              <input
                value={exerciseForm.videoUrl}
                onChange={(event) => setExerciseForm({ ...exerciseForm, videoUrl: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowExerciseForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showWorkoutForm && (
        <Modal title={editingWorkout ? 'Редактировать тренировку' : 'Новая тренировка'} onClose={() => setShowWorkoutForm(false)}>
          <form onSubmit={saveWorkout} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={workoutForm.name}
                onChange={(event) => setWorkoutForm({ ...workoutForm, name: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Описание</label>
              <textarea
                value={workoutForm.description}
                onChange={(event) => setWorkoutForm({ ...workoutForm, description: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={3}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Длительность (мин)</label>
                <input
                  type="number"
                  value={workoutForm.duration}
                  onChange={(event) => setWorkoutForm({ ...workoutForm, duration: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Сложность</label>
                <input
                  value={workoutForm.difficulty}
                  onChange={(event) => setWorkoutForm({ ...workoutForm, difficulty: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div>
              <label className="text-sm text-gray-600">Категория</label>
              <input
                value={workoutForm.category}
                onChange={(event) => setWorkoutForm({ ...workoutForm, category: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowWorkoutForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showWorkoutExercises && workoutForComposition && (
        <Modal title={`Состав: ${workoutForComposition.name}`} onClose={() => setShowWorkoutExercises(false)}>
          <div className="space-y-4">
            {workoutExercises.map((item, index) => (
              <div key={index} className="grid grid-cols-8 gap-2 items-end">
                <div className="col-span-2">
                  <label className="text-xs text-gray-500">Упражнение</label>
                  <select
                    value={item.exercise_id}
                    onChange={(event) => {
                      const updated = [...workoutExercises];
                      updated[index].exercise_id = event.target.value;
                      setWorkoutExercises(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  >
                    <option value="">Выберите</option>
                    {allExercises.map((exercise) => (
                      <option key={exercise.id} value={exercise.id}>
                        {exercise.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-xs text-gray-500">Порядок</label>
                  <input
                    type="number"
                    value={item.sort_order}
                    onChange={(event) => {
                      const updated = [...workoutExercises];
                      updated[index].sort_order = event.target.value;
                      setWorkoutExercises(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  />
                </div>
                <div>
                  <label className="text-xs text-gray-500">Подходы</label>
                  <input
                    type="number"
                    value={item.sets}
                    onChange={(event) => {
                      const updated = [...workoutExercises];
                      updated[index].sets = event.target.value;
                      setWorkoutExercises(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  />
                </div>
                <div>
                  <label className="text-xs text-gray-500">Повторы</label>
                  <input
                    value={item.reps}
                    onChange={(event) => {
                      const updated = [...workoutExercises];
                      updated[index].reps = event.target.value;
                      setWorkoutExercises(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  />
                </div>
                <div>
                  <label className="text-xs text-gray-500">Длительность</label>
                  <input
                    type="number"
                    value={item.duration_seconds}
                    onChange={(event) => {
                      const updated = [...workoutExercises];
                      updated[index].duration_seconds = event.target.value;
                      setWorkoutExercises(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  />
                </div>
                <div>
                  <label className="text-xs text-gray-500">Отдых</label>
                  <input
                    type="number"
                    value={item.rest_seconds}
                    onChange={(event) => {
                      const updated = [...workoutExercises];
                      updated[index].rest_seconds = event.target.value;
                      setWorkoutExercises(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  />
                </div>
                <div className="flex gap-1">
                  <button
                    className="px-2 py-2 rounded-lg border border-gray-200"
                    onClick={() => {
                      const updated = workoutExercises.filter((_, idx) => idx !== index);
                      setWorkoutExercises(updated.length ? updated : [{
                        exercise_id: '',
                        sort_order: 1,
                        sets: 0,
                        reps: '',
                        duration_seconds: 0,
                        rest_seconds: 0,
                      }]);
                    }}
                  >
                    ✕
                  </button>
                </div>
              </div>
            ))}
            <button
              className="w-full py-2 border border-dashed border-gray-300 rounded-xl text-sm"
              onClick={() => setWorkoutExercises([...workoutExercises, {
                exercise_id: '',
                sort_order: workoutExercises.length + 1,
                sets: 0,
                reps: '',
                duration_seconds: 0,
                rest_seconds: 0,
              }])}
            >
              Добавить упражнение
            </button>
            <div className="flex gap-3">
              <button className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowWorkoutExercises(false)}>
                Отмена
              </button>
              <button className="flex-1 py-3 rounded-xl bg-slate-900 text-white" onClick={saveWorkoutComposition}>
                Сохранить
              </button>
            </div>
          </div>
        </Modal>
      )}

      {showProgramForm && (
        <Modal title={editingProgram ? 'Редактировать программу' : 'Новая программа'} onClose={() => setShowProgramForm(false)}>
          <form onSubmit={saveProgram} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={programForm.name}
                onChange={(event) => setProgramForm({ ...programForm, name: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Описание</label>
              <textarea
                value={programForm.description}
                onChange={(event) => setProgramForm({ ...programForm, description: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={3}
              />
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={programForm.active}
                onChange={(event) => setProgramForm({ ...programForm, active: event.target.checked })}
              />
              Активная программа
            </label>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowProgramForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showProgramWorkouts && programForComposition && (
        <Modal title={`Состав: ${programForComposition.name}`} onClose={() => setShowProgramWorkouts(false)}>
          <div className="space-y-4">
            {programWorkouts.map((item, index) => (
              <div key={index} className="grid grid-cols-4 gap-2 items-end">
                <div className="col-span-3">
                  <label className="text-xs text-gray-500">Тренировка</label>
                  <select
                    value={item.workout_id}
                    onChange={(event) => {
                      const updated = [...programWorkouts];
                      updated[index].workout_id = event.target.value;
                      setProgramWorkouts(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  >
                    <option value="">Выберите</option>
                    {allWorkouts.map((workout) => (
                      <option key={workout.id} value={workout.id}>
                        {workout.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="flex gap-1">
                  <input
                    type="number"
                    value={item.sort_order}
                    onChange={(event) => {
                      const updated = [...programWorkouts];
                      updated[index].sort_order = event.target.value;
                      setProgramWorkouts(updated);
                    }}
                    className="w-full mt-1 px-2 py-2 rounded-lg border border-gray-200"
                  />
                  <button
                    className="px-2 py-2 rounded-lg border border-gray-200"
                    onClick={() => {
                      const updated = programWorkouts.filter((_, idx) => idx !== index);
                      setProgramWorkouts(updated.length ? updated : [{ workout_id: '', sort_order: 1 }]);
                    }}
                  >
                    ✕
                  </button>
                </div>
              </div>
            ))}
            <button
              className="w-full py-2 border border-dashed border-gray-300 rounded-xl text-sm"
              onClick={() => setProgramWorkouts([...programWorkouts, { workout_id: '', sort_order: programWorkouts.length + 1 }])}
            >
              Добавить тренировку
            </button>
            <div className="flex gap-3">
              <button className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowProgramWorkouts(false)}>
                Отмена
              </button>
              <button className="flex-1 py-3 rounded-xl bg-slate-900 text-white" onClick={saveProgramComposition}>
                Сохранить
              </button>
            </div>
          </div>
        </Modal>
      )}

      {showRecommendationForm && (
        <Modal title={editingRecommendation ? 'Редактировать рекомендацию' : 'Новая рекомендация'} onClose={() => setShowRecommendationForm(false)}>
          <form onSubmit={saveRecommendation} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={recommendationForm.title}
                onChange={(event) => setRecommendationForm({ ...recommendationForm, title: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Категория</label>
              <input
                value={recommendationForm.category}
                onChange={(event) => setRecommendationForm({ ...recommendationForm, category: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Иконка</label>
              <input
                value={recommendationForm.icon}
                onChange={(event) => setRecommendationForm({ ...recommendationForm, icon: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                placeholder="Например: 💡"
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Короткое описание</label>
              <input
                value={recommendationForm.excerpt}
                onChange={(event) => setRecommendationForm({ ...recommendationForm, excerpt: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Текст</label>
              <textarea
                value={recommendationForm.body}
                onChange={(event) => setRecommendationForm({ ...recommendationForm, body: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={4}
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Время чтения (мин)</label>
              <input
                type="number"
                value={recommendationForm.readTime}
                onChange={(event) => setRecommendationForm({ ...recommendationForm, readTime: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowRecommendationForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showVideoForm && (
        <Modal title={editingVideo ? 'Редактировать видео' : 'Новое видео'} onClose={() => setShowVideoForm(false)}>
          <form onSubmit={saveVideo} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={videoForm.title}
                onChange={(event) => setVideoForm({ ...videoForm, title: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Описание</label>
              <textarea
                value={videoForm.description}
                onChange={(event) => setVideoForm({ ...videoForm, description: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={3}
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Длительность (мин)</label>
                <input
                  type="number"
                  value={videoForm.duration}
                  onChange={(event) => setVideoForm({ ...videoForm, duration: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Сложность</label>
                <input
                  value={videoForm.difficulty}
                  onChange={(event) => setVideoForm({ ...videoForm, difficulty: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div>
              <label className="text-sm text-gray-600">Категория</label>
              <input
                value={videoForm.category}
                onChange={(event) => setVideoForm({ ...videoForm, category: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Ссылка</label>
              <input
                value={videoForm.url}
                onChange={(event) => setVideoForm({ ...videoForm, url: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
              />
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowVideoForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showNutritionForm && (
        <Modal title={editingNutrition ? 'Редактировать запись' : 'Новая запись'} onClose={() => setShowNutritionForm(false)}>
          <form onSubmit={saveNutrition} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={nutritionForm.title}
                onChange={(event) => setNutritionForm({ ...nutritionForm, title: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Описание</label>
              <textarea
                value={nutritionForm.description}
                onChange={(event) => setNutritionForm({ ...nutritionForm, description: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={3}
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Калории</label>
                <input
                  type="number"
                  value={nutritionForm.calories}
                  onChange={(event) => setNutritionForm({ ...nutritionForm, calories: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Категория</label>
                <input
                  value={nutritionForm.category}
                  onChange={(event) => setNutritionForm({ ...nutritionForm, category: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowNutritionForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
        </Modal>
      )}

      {showRewardForm && (
        <Modal title={editingReward ? 'Редактировать награду' : 'Новая награда'} onClose={() => setShowRewardForm(false)}>
          <form onSubmit={saveReward} className="space-y-4">
            <div>
              <label className="text-sm text-gray-600">Название</label>
              <input
                value={rewardForm.title}
                onChange={(event) => setRewardForm({ ...rewardForm, title: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                required
              />
            </div>
            <div>
              <label className="text-sm text-gray-600">Описание</label>
              <textarea
                value={rewardForm.description}
                onChange={(event) => setRewardForm({ ...rewardForm, description: event.target.value })}
                className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                rows={3}
                required
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-sm text-gray-600">Стоимость (баллы)</label>
                <input
                  type="number"
                  value={rewardForm.pointsCost}
                  onChange={(event) => setRewardForm({ ...rewardForm, pointsCost: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
              <div>
                <label className="text-sm text-gray-600">Категория</label>
                <input
                  value={rewardForm.category}
                  onChange={(event) => setRewardForm({ ...rewardForm, category: event.target.value })}
                  className="w-full mt-1 px-3 py-2 rounded-xl border border-gray-200"
                />
              </div>
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={rewardForm.active}
                onChange={(event) => setRewardForm({ ...rewardForm, active: event.target.checked })}
              />
              Активная награда
            </label>
            <div className="flex gap-3">
              <button type="button" className="flex-1 py-3 rounded-xl border border-gray-200" onClick={() => setShowRewardForm(false)}>
                Отмена
              </button>
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Сохранить
              </button>
            </div>
          </form>
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
              <button type="submit" className="flex-1 py-3 rounded-xl bg-slate-900 text-white">
                Отправить
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  );
}
