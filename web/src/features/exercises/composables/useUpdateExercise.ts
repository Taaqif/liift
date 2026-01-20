import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { Exercise } from "@/features/exercises/types";
import { exerciseKeys } from "@/lib/queryKeys";

export type UpdateExerciseRequest = {
  name: string;
  description?: string;
  primary_muscle_groups?: string[];
  secondary_muscle_groups?: string[];
  equipment?: string[];
};

async function updateExercise(
  id: number,
  data: UpdateExerciseRequest,
): Promise<Exercise> {
  return apiClient.put<Exercise>(`/exercises/${id}`, data);
}

export function useUpdateExercise() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateExerciseRequest }) =>
      updateExercise(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: exerciseKeys.all });
      toast.success(t("exercises.toasts.updated"));
    },
  });

  return {
    updateExercise: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
  };
}
