import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { Workout } from "@/features/workouts/types";
import { workoutKeys } from "@/lib/queryKeys";

export type CreateWorkoutRequest = {
  name: string;
  description?: string;
  is_library?: boolean;
  exercises: {
    exercise_id: number;
    rest_timer: number;
    note?: string;
    order: number;
    sets: {
      order: number;
      features: {
        feature_name: string;
        value: number;
      }[];
    }[];
  }[];
};

async function createWorkout(data: CreateWorkoutRequest): Promise<Workout> {
  return apiClient.post<Workout>("/workouts", data);
}

export function useCreateWorkout() {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: createWorkout,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: workoutKeys.all });
      toast.success(t("workouts.toasts.created"));
    },
  });

  return {
    createWorkout: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
  };
}
