import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { workoutSessionKeys } from "@/lib/queryKeys";

export type WorkoutSessionSummary = {
  id: number;
  workout_id: number;
  workout_name: string;
  started_at: string;
  ended_at: string | null;
  exercise_count: number;
  sets_completed: number;
};

export type WorkoutSessionsResponse = {
  data: WorkoutSessionSummary[];
  total: number;
  limit: number;
  offset: number;
};

export function useWorkoutSessions(
  limit: MaybeRefOrGetter<number> = 20,
  offset: MaybeRefOrGetter<number> = 0,
  workoutId: MaybeRefOrGetter<number | null> = null,
) {
  const params = computed(() => ({
    limit: toValue(limit),
    offset: toValue(offset),
    workoutId: toValue(workoutId),
  }));

  const { data, isLoading, error } = useQuery({
    queryKey: computed(() =>
      workoutSessionKeys.list({ limit: params.value.limit, offset: params.value.offset, workoutId: params.value.workoutId ?? undefined }),
    ),
    queryFn: () => {
      const l = toValue(limit);
      const o = toValue(offset);
      const wid = toValue(workoutId);
      const url = wid
        ? `/workout-sessions?limit=${l}&offset=${o}&workout_id=${wid}`
        : `/workout-sessions?limit=${l}&offset=${o}`;
      return apiClient.get<WorkoutSessionsResponse>(url);
    },
    staleTime: 0,
  });

  return {
    sessions: computed(() => data.value?.data ?? []),
    total: computed(() => data.value?.total ?? 0),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
