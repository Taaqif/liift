import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutPlanProgress } from "../types";
import { workoutPlanProgressKeys } from "@/lib/queryKeys";

export function useCompletePlanProgress() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (id: number) =>
      apiClient.post<WorkoutPlanProgress>(`/workout-plan-progress/${id}/complete`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanProgressKeys.all });
      toast.success(t("workoutPlans.progress.toasts.completed"));
    },
  });

  return {
    completePlan: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
