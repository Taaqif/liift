import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

async function fetchActiveSession(): Promise<WorkoutSession | null> {
  try {
    return await apiClient.get<WorkoutSession>("/workout-sessions/active");
  } catch (e) {
    const err = e as Error;
    if (err.message === "no_active_session") {
      return null;
    }
    throw e;
  }
}

export function useActiveWorkoutSession() {
  const query = useQuery({
    queryKey: workoutSessionKeys.active(),
    queryFn: fetchActiveSession,
  });

  return {
    session: computed(() => query.data.value ?? null),
    loading: computed(() => query.isLoading.value),
    error: computed(() => query.error.value),
    refetch: query.refetch,
    isFetched: computed(() => query.isFetched.value),
  };
}
