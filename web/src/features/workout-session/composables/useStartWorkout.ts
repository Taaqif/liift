import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

async function startWorkout(workoutId: number): Promise<WorkoutSession> {
  return apiClient.post<WorkoutSession>(`/workouts/${workoutId}/start`);
}

export function useStartWorkout() {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: startWorkout,
    onSuccess: (session) => {
      queryClient.setQueryData(workoutSessionKeys.detail(session.id), session);
      queryClient.setQueryData(workoutSessionKeys.active(), session);
      queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
      router.push({ name: "active-workout" });
      toast.success(t("workoutSession.toasts.started"));
    },
  });

  return {
    startWorkout: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
