export type Role = 'employee' | 'manager' | 'admin';

export type Screen =
  | 'home'
  | 'questionnaire'
  | 'program'
  | 'workout'
  | 'progress'
  | 'profile'
  | 'onboarding'
  | 'login'
  | 'register'
  | 'exerciseLibrary'
  | 'exerciseDetail'
  | 'achievements'
  | 'calendar'
  | 'recommendations'
  | 'settings'
  | 'notifications'
  | 'support'
  | 'community'
  | 'history'
  | 'workoutComplete'
  | 'goals'
  | 'statistics'
  | 'videoTutorials'
  | 'nutrition'
  | 'medicalInfo'
  | 'feedback'
  | 'admin'
  | 'manager';

export interface User {
  id: string;
  name: string;
  employeeId: string;
  role: Role;
  department: string;
  position?: string;
  points?: number;
}

export interface UserProfile {
  age: number;
  fitnessLevel: 'beginner' | 'intermediate' | 'advanced' | '' | null;
  restrictions: string[];
  goals: string[];
  onboardingComplete: boolean;
}

export interface Settings {
  notifications: {
    enabled: boolean;
    workoutReminders: boolean;
    achievementAlerts: boolean;
    weeklyReports: boolean;
    remindersEnabled: boolean;
  };
  preferences: {
    theme: 'light' | 'dark' | string;
    language: string;
    units: 'metric' | 'imperial' | string;
  };
  privacy: {
    shareProgress: boolean;
    showInLeaderboard: boolean;
  };
}

export interface Workout {
  id: string;
  name: string;
  description?: string;
  duration: number;
  difficulty: string;
  category?: string;
  exercises?: Exercise[];
  exercisesCount?: number;
  completed?: boolean;
  recommendedDate?: string;
}

export interface Exercise {
  id: string;
  name: string;
  sets: number;
  reps: string;
  duration?: number;
  rest: number;
  description: string;
  videoUrl?: string;
  category?: string;
  difficulty?: string;
  muscleGroups?: string[];
  equipment?: string[];
}

export interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: string;
  unlocked: boolean;
  unlockedDate?: string;
  progress?: number;
  total?: number;
}

export interface WorkoutHistory {
  id: string;
  workoutId: string;
  workoutName: string;
  date: string;
  duration: number;
  completedExercises: number;
  totalExercises: number;
  completed: boolean;
  calories?: number;
  rating?: number;
}

export interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'warning' | 'success' | 'reminder' | string;
  date: string;
  read: boolean;
}

export interface Goal {
  id: string;
  title: string;
  description: string;
  targetDate: string;
  progress: number;
  category: 'strength' | 'flexibility' | 'endurance' | 'weight' | 'other' | string;
}

export interface MedicalInfo {
  chronicDiseases: string[];
  injuries: string[];
  medications: string[];
  allergies: string[];
  doctorApproval: boolean;
  lastCheckup?: string;
  restrictions: string[];
}

export interface CalendarDay {
  day: number;
  date: string;
  isWorkout: boolean;
  isToday: boolean;
  isSelected: boolean;
}

export interface CalendarWorkout {
  id: string;
  name: string;
  date: string;
  duration: number;
  exercises: number;
  calories: number;
  completed: boolean;
}

export interface SupportTicket {
  id: string;
  category: string;
  subject: string;
  status: string;
  createdAt: string;
  user?: string;
}

export interface Recommendation {
  id: string;
  title: string;
  category: string;
  readTime: number;
  icon: string;
  excerpt: string;
  body: string;
  bookmarked?: boolean;
}

export interface VideoTutorial {
  id: string;
  title: string;
  description: string;
  duration: number;
  category: string;
  difficulty: string;
  url: string;
}

export interface NutritionItem {
  id: string;
  title: string;
  description: string;
  calories: number;
  category: string;
}
