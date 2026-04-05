import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutPlan } from "../types";
import { workoutPlanKeys } from "@/lib/queryKeys";

export function useWorkoutPlan(id: MaybeRefOrGetter<number | null | undefined>) {
  const { data, isLoading, error } = useQuery({
    queryKey: computed(() => workoutPlanKeys.detail(toValue(id) ?? 0)),
    queryFn: () => apiClient.get<WorkoutPlan>(`/workout-plans/${toValue(id)}`),
    enabled: computed(() => !!toValue(id)),
  });

  return {
    plan: computed(() => data.value ?? null),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
