import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { apiRequest, setAuthToken } from '../api/client';
import type { Settings, User, UserProfile } from '../types';

interface AuthState {
  user: User | null;
  profile: UserProfile | null;
  settings: Settings | null;
  loading: boolean;
  login: (employeeId: string, password: string) => Promise<void>;
  register: (payload: { name: string; employeeId: string; department?: string; position?: string; password: string }) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  updateProfile: (payload: Partial<UserProfile> & { name?: string; department?: string; position?: string }) => Promise<void>;
  updateSettings: (payload: Settings) => Promise<void>;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

function mapUser(data: any): User {
  return {
    id: data.id,
    name: data.name,
    employeeId: data.employee_id || data.employeeId,
    role: data.role,
    department: data.department || '',
    position: data.position || '',
    points: data.points,
  };
}

function mapProfile(data: any): UserProfile {
  return {
    age: data.age || 0,
    fitnessLevel: data.fitness_level ?? data.fitnessLevel ?? '',
    restrictions: data.restrictions || [],
    goals: data.goals || [],
    onboardingComplete: Boolean(data.onboarding_complete ?? data.onboardingComplete),
  };
}

function mapSettings(data: any): Settings {
  return {
    notifications: {
      enabled: data.notifications?.enabled ?? true,
      workoutReminders: data.notifications?.workout_reminders ?? true,
      achievementAlerts: data.notifications?.achievement_alerts ?? true,
      weeklyReports: data.notifications?.weekly_reports ?? false,
      remindersEnabled: data.notifications?.reminders_enabled ?? true,
    },
    preferences: {
      theme: data.preferences?.theme ?? 'light',
      language: data.preferences?.language ?? 'ru',
      units: data.preferences?.units ?? 'metric',
    },
    privacy: {
      shareProgress: data.privacy?.share_progress ?? true,
      showInLeaderboard: data.privacy?.show_in_leaderboard ?? true,
    },
  };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    try {
      const data = await apiRequest<any>('/auth/me');
      setUser(mapUser(data.user));
      setProfile(mapProfile(data.profile));
      setSettings(mapSettings(data.settings));
    } catch (error) {
      setAuthToken(null);
      setUser(null);
      setProfile(null);
      setSettings(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (localStorage.getItem('auth_token')) {
      refresh();
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (employeeId: string, password: string) => {
    const data = await apiRequest<any>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ employee_id: employeeId, password }),
    });
    setAuthToken(data.token);
    setUser(mapUser(data.user));
    setProfile({
      age: 0,
      fitnessLevel: data.user.fitness_level || '',
      restrictions: [],
      goals: [],
      onboardingComplete: Boolean(data.user.onboarding_complete),
    });
    await refresh();
  };

  const register = async (payload: { name: string; employeeId: string; department?: string; position?: string; password: string }) => {
    const data = await apiRequest<any>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        name: payload.name,
        employee_id: payload.employeeId,
        department: payload.department || '',
        position: payload.position || '',
        password: payload.password,
      }),
    });
    setAuthToken(data.token);
    setUser(mapUser(data.user));
    setProfile({
      age: 0,
      fitnessLevel: '',
      restrictions: [],
      goals: [],
      onboardingComplete: false,
    });
    await refresh();
  };

  const logout = async () => {
    try {
      await apiRequest('/auth/logout', { method: 'POST' });
    } finally {
      setAuthToken(null);
      setUser(null);
      setProfile(null);
      setSettings(null);
    }
  };

  const updateProfile = async (payload: Partial<UserProfile> & { name?: string; department?: string; position?: string }) => {
    const data = await apiRequest<any>('/profile', {
      method: 'PUT',
      body: JSON.stringify({
        name: payload.name,
        department: payload.department,
        position: payload.position,
        age: payload.age,
        fitness_level: payload.fitnessLevel,
        restrictions: payload.restrictions,
        goals: payload.goals,
      }),
    });
    setUser(mapUser(data.user));
    setProfile(mapProfile(data.profile));
  };

  const updateSettings = async (payload: Settings) => {
    const data = await apiRequest<any>('/settings', {
      method: 'PUT',
      body: JSON.stringify({
        notifications: {
          enabled: payload.notifications.enabled,
          workout_reminders: payload.notifications.workoutReminders,
          achievement_alerts: payload.notifications.achievementAlerts,
          weekly_reports: payload.notifications.weeklyReports,
          reminders_enabled: payload.notifications.remindersEnabled,
        },
        preferences: {
          language: payload.preferences.language,
          theme: payload.preferences.theme,
          units: payload.preferences.units,
        },
        privacy: {
          share_progress: payload.privacy.shareProgress,
          show_in_leaderboard: payload.privacy.showInLeaderboard,
        },
      }),
    });
    setSettings(mapSettings(data));
  };

  const value = useMemo(
    () => ({
      user,
      profile,
      settings,
      loading,
      login,
      register,
      logout,
      refresh,
      updateProfile,
      updateSettings,
    }),
    [user, profile, settings, loading]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return ctx;
}
