import { useEffect, useMemo, useState } from 'react';
import { Home } from './components/Home';
import { Questionnaire } from './components/Questionnaire';
import { TrainingProgram } from './components/TrainingProgram';
import { WorkoutSession } from './components/WorkoutSession';
import { Progress } from './components/Progress';
import { Profile } from './components/Profile';
import { Navigation } from './components/Navigation';
import { Login } from './components/Login';
import { Onboarding } from './components/Onboarding';
import { ExerciseLibrary } from './components/ExerciseLibrary';
import { ExerciseDetail } from './components/ExerciseDetail';
import { Achievements } from './components/Achievements';
import { Calendar } from './components/Calendar';
import { Recommendations } from './components/Recommendations';
import { Settings } from './components/Settings';
import { Notifications } from './components/Notifications';
import { Support } from './components/Support';
import { Community } from './components/Community';
import { History } from './components/History';
import { WorkoutComplete } from './components/WorkoutComplete';
import { Goals } from './components/Goals';
import { Statistics } from './components/Statistics';
import { VideoTutorials } from './components/VideoTutorials';
import { Nutrition } from './components/Nutrition';
import { MedicalInfo } from './components/MedicalInfo';
import { Feedback } from './components/Feedback';
import { Register } from './components/Register';
import { AdminPanel } from './components/admin/AdminPanel';
import { ManagerPanel } from './components/manager/ManagerPanel';
import { useAuth } from './context/AuthContext';
import { apiRequest } from './api/client';
import type { Screen } from './types';

const allScreens: Screen[] = [
  'home',
  'questionnaire',
  'program',
  'workout',
  'progress',
  'profile',
  'onboarding',
  'login',
  'register',
  'exerciseLibrary',
  'exerciseDetail',
  'achievements',
  'calendar',
  'recommendations',
  'settings',
  'notifications',
  'support',
  'community',
  'history',
  'workoutComplete',
  'goals',
  'statistics',
  'videoTutorials',
  'nutrition',
  'medicalInfo',
  'feedback',
  'admin',
  'manager',
];

const screenSet = new Set(allScreens);

function screenFromHash(): Screen {
  const raw = window.location.hash.replace('#', '');
  if (screenSet.has(raw as Screen)) {
    return raw as Screen;
  }
  return 'login';
}

function App() {
  const { user, profile, loading } = useAuth();
  const [currentScreen, setCurrentScreen] = useState<Screen>(() => screenFromHash());
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);
  const [selectedExerciseId, setSelectedExerciseId] = useState<string | null>(null);

  const navigate = (screen: Screen) => {
    setCurrentScreen(screen);
    window.location.hash = screen;
  };

  useEffect(() => {
    const handler = () => {
      const next = screenFromHash();
      setCurrentScreen(next);
    };
    window.addEventListener('hashchange', handler);
    return () => window.removeEventListener('hashchange', handler);
  }, []);

  useEffect(() => {
    if (loading) return;

    if (!user) {
      if (currentScreen !== 'login' && currentScreen !== 'register') {
        navigate('login');
      }
      return;
    }

    if (user && (currentScreen === 'login' || currentScreen === 'register')) {
      if (profile?.onboardingComplete) {
        navigate('home');
      } else {
        navigate('onboarding');
      }
      return;
    }

    if (currentScreen === 'admin' && user.role !== 'admin') {
      navigate('home');
      return;
    }

    if (currentScreen === 'manager') {
      if (user.role === 'admin') {
        navigate('admin');
        return;
      }
      if (user.role !== 'manager') {
        navigate('home');
      }
    }
  }, [user, profile, loading, currentScreen]);

  const showNav = useMemo(() => {
    if (!user) return false;
    const hideFor = new Set<Screen>([
      'login',
      'register',
      'onboarding',
      'workout',
    ]);
    return !hideFor.has(currentScreen);
  }, [user, currentScreen]);

  const handleStartWorkout = async (workoutId: string) => {
    const data = await apiRequest<{ session_id: string }>('/workout-sessions', {
      method: 'POST',
      body: JSON.stringify({ workout_id: workoutId }),
    });
    setActiveSessionId(data.session_id);
    navigate('workout');
  };

  const renderScreen = () => {
    if (!user) {
      if (currentScreen === 'register') {
        return <Register onNavigate={navigate} />;
      }
      return <Login onNavigate={navigate} />;
    }

    switch (currentScreen) {
      case 'onboarding':
        return <Onboarding onComplete={() => navigate('home')} onNavigate={navigate} />;
      case 'home':
        return <Home onNavigate={navigate} />;
      case 'questionnaire':
        return <Questionnaire onComplete={() => navigate('program')} />;
      case 'program':
        return <TrainingProgram onStartWorkout={handleStartWorkout} />;
      case 'workout':
        return activeSessionId ? (
          <WorkoutSession sessionId={activeSessionId} onComplete={() => navigate('workoutComplete')} />
        ) : null;
      case 'workoutComplete':
        return activeSessionId ? (
          <WorkoutComplete sessionId={activeSessionId} onNavigate={navigate} />
        ) : null;
      case 'progress':
        return <Progress onNavigate={navigate} />;
      case 'profile':
        return <Profile onNavigate={navigate} />;
      case 'exerciseLibrary':
        return <ExerciseLibrary onNavigate={navigate} onExerciseSelect={(id) => {
          setSelectedExerciseId(id);
          navigate('exerciseDetail');
        }} />;
      case 'exerciseDetail':
        return selectedExerciseId ? (
          <ExerciseDetail exerciseId={selectedExerciseId} onNavigate={navigate} />
        ) : null;
      case 'achievements':
        return <Achievements onNavigate={navigate} />;
      case 'calendar':
        return <Calendar onNavigate={navigate} />;
      case 'recommendations':
        return <Recommendations onNavigate={navigate} />;
      case 'settings':
        return <Settings onNavigate={navigate} />;
      case 'notifications':
        return <Notifications onNavigate={navigate} />;
      case 'support':
        return <Support onNavigate={navigate} />;
      case 'community':
        return <Community onNavigate={navigate} />;
      case 'history':
        return <History onNavigate={navigate} onStartWorkout={handleStartWorkout} />;
      case 'goals':
        return <Goals onNavigate={navigate} />;
      case 'statistics':
        return <Statistics onNavigate={navigate} />;
      case 'videoTutorials':
        return <VideoTutorials onNavigate={navigate} />;
      case 'nutrition':
        return <Nutrition onNavigate={navigate} />;
      case 'medicalInfo':
        return <MedicalInfo onNavigate={navigate} />;
      case 'feedback':
        return <Feedback onNavigate={navigate} sessionId={activeSessionId} />;
      case 'admin':
        return <AdminPanel onNavigate={navigate} />;
      case 'manager':
        return <ManagerPanel onNavigate={navigate} />;
      default:
        return <Home onNavigate={navigate} />;
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <div className="text-sm text-slate-500">Загрузка...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-100">
      <div className="min-h-screen flex flex-col lg:flex-row">
        {showNav && (
          <Navigation currentScreen={currentScreen} onNavigate={navigate} userRole={user?.role} />
        )}
        <main className="flex-1">
          {renderScreen()}
        </main>
      </div>
    </div>
  );
}

export default App;
