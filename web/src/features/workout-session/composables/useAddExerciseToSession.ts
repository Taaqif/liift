import type { MaybeRefOrGetter } from "vue";
import { toValue } from "vue";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

async function addExerciseToSession(
  sessionId: number,
  exerciseId: number,
  restTimer: number,
): Promise<WorkoutSession> {
  return apiClient.post<WorkoutSession>(
    `/workout-sessions/${sessionId}/exercises`,
    { exercise_id: exerciseId, rest_timer: restTimer },
  );
}

export function useAddExerciseToSession(sessionId: MaybeRefOrGetter<number>) {
  const queryClient = useQueryClient();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: ({ exerciseId, restTimer }: { exerciseId: number; restTimer: number }) =>
      addExerciseToSession(toValue(sessionId), exerciseId, restTimer),
    onSuccess: (data) => {
      queryClient.setQueryData(workoutSessionKeys.detail(data.id), data);
      queryClient.setQueryData(workoutSessionKeys.active(), data);
      toast.success(t("workoutSession.toasts.exerciseAdded"));
    },
  });

  return {
    addExercise: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
