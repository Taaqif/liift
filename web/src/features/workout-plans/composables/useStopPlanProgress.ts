import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import { workoutPlanProgressKeys } from "@/lib/queryKeys";

export function useStopPlanProgress() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (id: number) =>
      apiClient.delete(`/workout-plan-progress/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutPlanProgressKeys.all });
      toast.success(t("workoutPlans.progress.toasts.stopped"));
    },
  });

  return {
    stopPlan: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
