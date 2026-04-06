import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutPlanProgress } from "../types";
import { workoutPlanProgressKeys } from "@/lib/queryKeys";

export function useActivePlanProgress() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: workoutPlanProgressKeys.active(),
    queryFn: () => apiClient.get<WorkoutPlanProgress>("/workout-plan-progress/active").catch((err) => {
      if (err instanceof Error && err.message === "no_active_plan_progress") return null;
      throw err;
    }),
    retry: false,
  });

  return {
    progress: computed(() => data.value ?? null),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}
