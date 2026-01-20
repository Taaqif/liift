import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import { exerciseKeys } from "@/lib/queryKeys";

async function deleteExercise(id: number): Promise<void> {
  return apiClient.delete<void>(`/exercises/${id}`);
}

export function useDeleteExercise() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: deleteExercise,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: exerciseKeys.all });
      toast.success(t("exercises.toasts.deleted"));
    },
  });

  return {
    deleteExercise: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
  };
}
