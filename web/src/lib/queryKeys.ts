/**
 * Query key factory for centralized query key management.
 * Organized by feature for scalability and maintainability.
 */

const baseKeys = {
  auth: ["auth"] as const,
} as const;

export const authKeys = {
  all: baseKeys.auth,
  user: () => [...baseKeys.auth, "user"] as const,
} as const;

export const queryKeys = {
  auth: authKeys,
} as const;
