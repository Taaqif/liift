import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutPlanProgress } from "../types";
import { workoutPlanProgressKeys } from "@/lib/queryKeys";

export function useUpdatePlanPosition() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ id, week, day }: { id: number; week: number; day: number }) =>
      apiClient.patch<WorkoutPlanProgress>(`/workout-plan-progress/${id}`, {
        current_week: week,
        current_day: day,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanProgressKeys.all });
    },
  });

  return {
    updatePosition: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
