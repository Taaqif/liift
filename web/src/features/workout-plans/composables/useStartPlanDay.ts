import { useMutation, useQueryClient } from "@tanstack/vue-query";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { apiClient } from "@/lib/api";
import type { WorkoutSession } from "@/features/workout-session/types";
import { workoutSessionKeys } from "@/lib/queryKeys";

export function useStartPlanDay() {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { t } = useI18n();

  const mutation = useMutation({
    mutationFn: (progressId: number) =>
      apiClient.post<WorkoutSession>(`/workout-plan-progress/${progressId}/start-day`),
    onSuccess: (session) => {
      queryClient.setQueryData(workoutSessionKeys.detail(session.id), session);
      queryClient.setQueryData(workoutSessionKeys.active(), session);
      queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
      router.push({ name: "active-workout" });
      toast.success(t("workoutSession.toasts.started"));
    },
  });

  return {
    startDay: mutation.mutateAsync,
    isPending: mutation.isPending,
    error: mutation.error,
  };
}
