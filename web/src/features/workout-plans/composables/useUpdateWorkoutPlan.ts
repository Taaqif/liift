import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutPlan, WorkoutPlanFormValues } from "../types";
import { workoutPlanKeys } from "@/lib/queryKeys";

export function useUpdateWorkoutPlan() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: WorkoutPlanFormValues }) =>
      apiClient.put<WorkoutPlan>(`/workout-plans/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanKeys.all });
      toast.success(t("workoutPlans.toasts.updated"));
    },
  });

  return {
    updatePlan: (id: number, data: WorkoutPlanFormValues) =>
      mutation.mutateAsync({ id, data }),
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
