import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import { workoutSessionKeys } from "@/lib/queryKeys";

export function useDeleteWorkoutSession() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (sessionId: number) =>
      apiClient.delete(`/workout-sessions/${sessionId}`),
    onSuccess: (_data, sessionId) => {
      queryClient.removeQueries({ queryKey: workoutSessionKeys.detail(sessionId) });
      queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
      toast.success(t("workoutHistory.deleted"));
    },
  });

  return {
    deleteSession: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
