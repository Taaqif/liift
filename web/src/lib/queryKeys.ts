/**
 * Query key factory for centralized query key management.
 * Organized by feature for scalability and maintainability.
 */

const baseKeys = {
  auth: ["auth"] as const,
  reference: ["reference"] as const,
  exercises: ["exercises"] as const,
  workouts: ["workouts"] as const,
  workoutPlans: ["workout-plans"] as const,
  workoutPlanProgress: ["workout-plan-progress"] as const,
  profile: ["profile"] as const,
} as const;

export const authKeys = {
  all: baseKeys.auth,
  user: () => [...baseKeys.auth, "user"] as const,
} as const;

export const referenceKeys = {
  all: baseKeys.reference,
  equipment: () => [...baseKeys.reference, "equipment"] as const,
  muscleGroup: () => [...baseKeys.reference, "muscleGroup"] as const,
  exerciseFeatures: () => [...baseKeys.reference, "exerciseFeatures"] as const,
} as const;

export const exerciseKeys = {
  all: baseKeys.exercises,
  list: (params?: {
    limit?: number;
    offset?: number;
    search?: string;
    muscleGroup?: string[];
    equipment?: string[];
  }) =>
    [
      ...baseKeys.exercises,
      "list",
      params?.limit ?? 20,
      params?.offset ?? 0,
      params?.search ?? "",
      params?.muscleGroup?.sort().join(",") ?? "",
      params?.equipment?.sort().join(",") ?? "",
    ] as const,
  detail: (id: number) => [...baseKeys.exercises, "detail", id] as const,
  logs: (id: number, limit?: number, offset?: number, from?: string, to?: string) =>
    [...baseKeys.exercises, "logs", id, limit ?? 20, offset ?? 0, from ?? null, to ?? null] as const,
} as const;

export const workoutKeys = {
  all: baseKeys.workouts,
  list: (params?: {
    limit?: number;
    offset?: number;
    search?: string;
    exerciseFeatures?: string[];
    exerciseIds?: number[];
    muscleGroup?: string[];
    equipment?: string[];
  }) =>
    [
      ...baseKeys.workouts,
      "list",
      params?.limit ?? 20,
      params?.offset ?? 0,
      params?.search ?? "",
      params?.exerciseFeatures?.slice().sort().join(",") ?? "",
      params?.exerciseIds?.slice().sort((a, b) => a - b).join(",") ?? "",
      params?.muscleGroup?.sort().join(",") ?? "",
      params?.equipment?.sort().join(",") ?? "",
    ] as const,
  detail: (id: number) => [...baseKeys.workouts, "detail", id] as const,
} as const;

const workoutSessionBase = ["workout-sessions"] as const;

export const workoutSessionKeys = {
  all: workoutSessionBase,
  list: (params?: { limit?: number; offset?: number; workoutId?: number; date?: string; from?: string; to?: string }) =>
    [
      ...workoutSessionBase,
      "list",
      params?.limit ?? 20,
      params?.offset ?? 0,
      params?.workoutId ?? null,
      params?.date ?? null,
      params?.from ?? null,
      params?.to ?? null,
    ] as const,
  activityDates: (year: number, month: number) =>
    [...workoutSessionBase, "activity", year, month] as const,
  weeklyStats: (from: string, to: string) =>
    [...workoutSessionBase, "weekly-stats", from, to] as const,
  active: () => [...workoutSessionBase, "active"] as const,
  detail: (id: number) => [...workoutSessionBase, id] as const,
} as const;

export const workoutPlanKeys = {
  all: baseKeys.workoutPlans,
  list: (params?: { limit?: number; offset?: number }) =>
    [
      ...baseKeys.workoutPlans,
      "list",
      params?.limit ?? 20,
      params?.offset ?? 0,
    ] as const,
  detail: (id: number) => [...baseKeys.workoutPlans, "detail", id] as const,
} as const;

export const workoutPlanProgressKeys = {
  all: baseKeys.workoutPlanProgress,
  active: () => [...baseKeys.workoutPlanProgress, "active"] as const,
} as const;

export const profileKeys = {
  all: baseKeys.profile,
  me: () => [...baseKeys.profile, "me"] as const,
} as const;

export const queryKeys = {
  auth: authKeys,
  reference: referenceKeys,
  exercises: exerciseKeys,
  workouts: workoutKeys,
  workoutSessions: workoutSessionKeys,
  workoutPlans: workoutPlanKeys,
  workoutPlanProgress: workoutPlanProgressKeys,
  profile: profileKeys,
} as const;
