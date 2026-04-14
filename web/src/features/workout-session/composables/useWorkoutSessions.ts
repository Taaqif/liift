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
  date: MaybeRefOrGetter<string | null> = null,
  from: MaybeRefOrGetter<string | null> = null,
  to: MaybeRefOrGetter<string | null> = null,
) {
  const params = computed(() => ({
    limit: toValue(limit),
    offset: toValue(offset),
    workoutId: toValue(workoutId),
    date: toValue(date),
    from: toValue(from),
    to: toValue(to),
  }));

  const { data, isLoading, error } = useQuery({
    queryKey: computed(() =>
      workoutSessionKeys.list({
        limit: params.value.limit,
        offset: params.value.offset,
        workoutId: params.value.workoutId ?? undefined,
        date: params.value.date ?? undefined,
        from: params.value.from ?? undefined,
        to: params.value.to ?? undefined,
      }),
    ),
    queryFn: () => {
      const l = toValue(limit);
      const o = toValue(offset);
      const wid = toValue(workoutId);
      const d = toValue(date);
      const f = toValue(from);
      const t = toValue(to);
      const qs = new URLSearchParams({ limit: String(l), offset: String(o) });
      if (wid) qs.set("workout_id", String(wid));
      if (d) qs.set("date", d);
      if (f) qs.set("from", f);
      if (t) qs.set("to", t);
      return apiClient.get<WorkoutSessionsResponse>(`/workout-sessions?${qs}`);
    },
    staleTime: 0,
    enabled: computed(() => toValue(date) !== null || toValue(workoutId) !== null || (toValue(limit) > 0)),
  });

  return {
    sessions: computed(() => data.value?.data ?? []),
    total: computed(() => data.value?.total ?? 0),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
