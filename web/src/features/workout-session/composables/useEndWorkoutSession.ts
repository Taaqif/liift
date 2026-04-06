import type { MaybeRefOrGetter } from "vue";
import { toValue } from "vue";
import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys, workoutPlanProgressKeys, exerciseKeys } from "@/lib/queryKeys";

async function endWorkoutSession(sessionId: number): Promise<WorkoutSession> {
  return apiClient.post<WorkoutSession>(
    `/workout-sessions/${sessionId}/end`,
  );
}

export function useEndWorkoutSession(sessionId: MaybeRefOrGetter<number>) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: () => endWorkoutSession(toValue(sessionId)),
    onSuccess: (data) => {
      queryClient.setQueryData(workoutSessionKeys.detail(data.id), data);
      queryClient.removeQueries({ queryKey: workoutSessionKeys.active() });
      queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
      queryClient.invalidateQueries({ queryKey: exerciseKeys.all });
      if (data.plan_progress_id) {
        queryClient.invalidateQueries({ queryKey: workoutPlanProgressKeys.all });
        router.push({ name: "active-plan" });
      } else {
        router.push({ name: "workouts" });
      }
      toast.success(t("workoutSession.toasts.ended"));
    },
  });

  return {
    endSession: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
