import { computed, toValue, type MaybeRefOrGetter } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

async function fetchWorkoutSession(id: number): Promise<WorkoutSession> {
  return apiClient.get<WorkoutSession>(`/workout-sessions/${id}`);
}

export function useWorkoutSession(id: MaybeRefOrGetter<number | null | undefined>) {
  const query = useQuery({
    queryKey: computed(() => workoutSessionKeys.detail(toValue(id) ?? 0)),
    queryFn: () => fetchWorkoutSession(toValue(id)!),
    enabled: computed(() => {
      const v = toValue(id);
      return typeof v === "number" && v > 0;
    }),
  });

  return {
    session: computed(() => query.data.value ?? null),
    loading: computed(() => query.isLoading.value),
    error: computed(() => query.error.value),
    refetch: query.refetch,
  };
}
