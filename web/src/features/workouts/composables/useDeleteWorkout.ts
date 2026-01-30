import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import { workoutKeys } from "@/lib/queryKeys";

async function deleteWorkout(id: number): Promise<void> {
  return apiClient.delete(`/workouts/${id}`);
}

export function useDeleteWorkout() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: deleteWorkout,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutKeys.all });
      toast.success(t("workouts.toasts.deleted"));
    },
  });

  return {
    deleteWorkout: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
  };
}
