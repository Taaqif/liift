import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

async function fetchWorkoutSession(id: number): Promise<WorkoutSession> {
  return apiClient.get<WorkoutSession>(`/workout-sessions/${id}`);
}

export function useWorkoutSession(id: number | undefined | null) {
  const query = useQuery({
    queryKey: computed(() => workoutSessionKeys.detail(id ?? 0)),
    queryFn: () => fetchWorkoutSession(id!),
    enabled: computed(() => typeof id === "number" && id > 0),
  });

  return {
    session: computed(() => query.data.value ?? null),
    loading: computed(() => query.isLoading.value),
    error: computed(() => query.error.value),
    refetch: query.refetch,
  };
}
