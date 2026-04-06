import type { MaybeRefOrGetter } from "vue";
import { toValue } from "vue";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys, workoutPlanProgressKeys } from "@/lib/queryKeys";

export function useCancelWorkoutSession(sessionId: MaybeRefOrGetter<number>) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: () =>
      apiClient.post<WorkoutSession>(`/workout-sessions/${toValue(sessionId)}/cancel`),
    onSuccess: (data) => {
      queryClient.setQueryData(workoutSessionKeys.detail(data.id), data);
      queryClient.removeQueries({ queryKey: workoutSessionKeys.active() });
      queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
      if (data.plan_progress_id) {
        queryClient.invalidateQueries({ queryKey: workoutPlanProgressKeys.all });
        router.push({ name: "active-plan" });
      } else {
        router.push({ name: "workouts" });
      }
    },
  });

  return {
    cancelSession: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
