import type { MaybeRefOrGetter } from "vue";
import { toValue } from "vue";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type {
  WorkoutSession,
  UpdateWorkoutSessionPayload,
} from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

async function updateWorkoutSession(
  sessionId: number,
  payload: UpdateWorkoutSessionPayload,
): Promise<WorkoutSession> {
  return apiClient.patch<WorkoutSession>(
    `/workout-sessions/${sessionId}`,
    payload,
  );
}

export function useUpdateWorkoutSession(sessionId: MaybeRefOrGetter<number>) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (payload: UpdateWorkoutSessionPayload) =>
      updateWorkoutSession(toValue(sessionId), payload),
    onSuccess: (data) => {
      queryClient.setQueryData(workoutSessionKeys.detail(data.id), data);
      queryClient.setQueryData(workoutSessionKeys.active(), data);
    },
  });

  return {
    updateSession: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
