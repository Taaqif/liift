import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutPlan, WorkoutPlanFormValues } from "../types";
import { workoutPlanKeys } from "@/lib/queryKeys";

export function useCreateWorkoutPlan() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (data: WorkoutPlanFormValues) =>
      apiClient.post<WorkoutPlan>("/workout-plans", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanKeys.all });
      toast.success(t("workoutPlans.toasts.created"));
    },
  });

  return {
    createPlan: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
