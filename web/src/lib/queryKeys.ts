/**
 * Query key factory for centralized query key management.
 * Organized by feature for scalability and maintainability.
 */

const baseKeys = {
  auth: ["auth"] as const,
  reference: ["reference"] as const,
  exercises: ["exercises"] as const,
} as const;

export const authKeys = {
  all: baseKeys.auth,
  user: () => [...baseKeys.auth, "user"] as const,
} as const;

export const referenceKeys = {
  all: baseKeys.reference,
  equipment: () => [...baseKeys.reference, "equipment"] as const,
  muscleGroup: () => [...baseKeys.reference, "muscleGroup"] as const,
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
} as const;

export const queryKeys = {
  auth: authKeys,
  reference: referenceKeys,
  exercises: exerciseKeys,
} as const;
