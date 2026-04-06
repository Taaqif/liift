import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutPlanProgress } from "../types";
import { workoutPlanProgressKeys } from "@/lib/queryKeys";

export function useStartPlan() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (planId: number) =>
      apiClient.post<WorkoutPlanProgress>("/workout-plan-progress", { plan_id: planId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanProgressKeys.all });
      toast.success(t("workoutPlans.progress.toasts.started"));
    },
  });

  return {
    startPlan: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
