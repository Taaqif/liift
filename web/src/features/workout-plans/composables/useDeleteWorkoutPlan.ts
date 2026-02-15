import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import { workoutPlanKeys } from "@/lib/queryKeys";

export function useDeleteWorkoutPlan() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (id: number) =>
      apiClient.delete(`/workout-plans/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanKeys.all });
      toast.success(t("workoutPlans.toasts.deleted"));
    },
  });

  return {
    deletePlan: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
