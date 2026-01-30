import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { Workout } from "@/features/workouts/types";
import { workoutKeys } from "@/lib/queryKeys";

export type UpdateWorkoutRequest = {
  name: string;
  description?: string;
  exercises: {
    id?: number;
    exercise_id: number;
    rest_timer: number;
    note?: string;
    order: number;
    sets: {
      id?: number;
      order: number;
      features: {
        id?: number;
        feature_name: string;
        value: number;
      }[];
    }[];
  }[];
};

async function updateWorkout(
  id: number,
  data: UpdateWorkoutRequest,
): Promise<Workout> {
  return apiClient.put<Workout>(`/workouts/${id}`, data);
}

export function useUpdateWorkout() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateWorkoutRequest }) =>
      updateWorkout(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutKeys.all });
      toast.success(t("workouts.toasts.updated"));
    },
  });

  return {
    updateWorkout: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
  };
}
