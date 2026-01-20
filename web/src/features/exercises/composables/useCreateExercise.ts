import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { Exercise } from "@/features/exercises/types";
import { exerciseKeys } from "@/lib/queryKeys";

export type CreateExerciseRequest = {
  name: string;
  description?: string;
  primary_muscle_groups?: string[];
  secondary_muscle_groups?: string[];
  equipment?: string[];
};

async function createExercise(
  data: CreateExerciseRequest,
): Promise<Exercise> {
  return apiClient.post<Exercise>("/exercises", data);
}

export function useCreateExercise() {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createExercise,
    onSuccess: () => {
      // Invalidate exercises list to refetch
      queryClient.invalidateQueries({ queryKey: exerciseKeys.all });
    },
  });

  return {
    createExercise: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
    isSuccess: mutation.isSuccess,
  };
}
