package models

import "time"

type User struct {
  ID         string
  Name       string
  EmployeeID string
  Role       string
  Department string
  Position   string
}

type UserProfile struct {
  Age           int
  FitnessLevel  string
  Restrictions  []string
  Goals         []string
  OnboardingComplete bool
}

type Workout struct {
  ID          string
  Name        string
  Description string
  Duration    int
  Difficulty  string
  Category    string
}

type Exercise struct {
  ID          string
  Name        string
  Description string
  Category    string
  Difficulty  string
  Sets        int
  Reps        string
  Duration    int
  Rest        int
  MuscleGroups []string
  Equipment   []string
  VideoURL    string
}

type WorkoutSession struct {
  ID                 string
  WorkoutName        string
  StartedAt          time.Time
  CompletedAt        *time.Time
  DurationMinutes    int
  TotalExercises     int
  CompletedExercises int
  CaloriesBurned     int
}

type Achievement struct {
  ID          string
  Title       string
  Description string
  Icon        string
  Unlocked    bool
  UnlockedAt  *time.Time
  Progress    int
  Total       int
}

type Goal struct {
  ID          string
  Title       string
  Description string
  TargetDate  string
  Progress    int
  Category    string
}

type Notification struct {
  ID      string
  Title   string
  Message string
  Type    string
  Created string
  Read    bool
}

type Recommendation struct {
  ID       string
  Title    string
  Body     string
  Category string
}

type CommunityPost struct {
  ID        string
  Title     string
  Body      string
  Author    string
  CreatedAt string
  Likes     int
}

type SupportTicket struct {
  ID        string
  Category  string
  Subject   string
  Status    string
  CreatedAt string
}

type VideoTutorial struct {
  ID          string
  Title       string
  Description string
  Duration    int
  Category    string
  Difficulty  string
  URL         string
}

type NutritionItem struct {
  ID          string
  Title       string
  Description string
  Calories    int
  Category    string
}

type Reward struct {
  ID          string
  Title       string
  Description string
  PointsCost  int
  Category    string
}

type EmployeeStats struct {
  UserID        string
  Name          string
  Department    string
  WorkoutsCount int
  HoursTotal    float64
  Points        int
}
