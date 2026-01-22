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
  exercise_features?: string[];
  image?: File | null;
};

async function updateExercise(
  id: number,
  data: UpdateExerciseRequest,
): Promise<Exercise> {
  // If image is provided, send as FormData
  if (data.image) {
    const formData = new FormData();
    formData.append("name", data.name);
    if (data.description) {
      formData.append("description", data.description);
    }
    data.primary_muscle_groups?.forEach((mg) => {
      formData.append("primary_muscle_groups", mg);
    });
    data.secondary_muscle_groups?.forEach((mg) => {
      formData.append("secondary_muscle_groups", mg);
    });
    data.equipment?.forEach((eq) => {
      formData.append("equipment", eq);
    });
    data.exercise_features?.forEach((ef) => {
      formData.append("exercise_features", ef);
    });
    formData.append("image", data.image);
    return apiClient.put<Exercise>(`/exercises/${id}`, formData);
  }

  const jsonData = {
    name: data.name,
    description: data.description,
    primary_muscle_groups: data.primary_muscle_groups,
    secondary_muscle_groups: data.secondary_muscle_groups,
    equipment: data.equipment,
    exercise_features: data.exercise_features,
    image_guid: data.image === null ? "" : undefined,
  };

  return apiClient.put<Exercise>(`/exercises/${id}`, jsonData);
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
