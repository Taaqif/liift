import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { Workout } from "@/features/workouts/types";
import { workoutKeys } from "@/lib/queryKeys";

export function useWorkout(id: MaybeRefOrGetter<number | null | undefined>) {
  const { data, isLoading, error } = useQuery({
    queryKey: computed(() => workoutKeys.detail(toValue(id) ?? 0)),
    queryFn: () => apiClient.get<Workout>(`/workouts/${toValue(id)}`),
    enabled: computed(() => !!toValue(id)),
  });

  return {
    workout: computed(() => data.value ?? null),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
