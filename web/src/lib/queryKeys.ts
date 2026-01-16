/**
 * Query key factory for centralized query key management.
 * Organized by feature for scalability and maintainability.
 */

const baseKeys = {
  auth: ["auth"] as const,
  reference: ["reference"] as const,
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

export const queryKeys = {
  auth: authKeys,
  reference: referenceKeys,
} as const;
