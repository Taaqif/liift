import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { exerciseKeys } from "@/lib/queryKeys";

export type ExerciseLogValue = {
  feature_name: string;
  value: number;
};

export type ExerciseLogSet = {
  order: number;
  completed_at: string;
  values: ExerciseLogValue[];
};

export type ExerciseLogEntry = {
  session_id: number;
  date: string;
  workout_name: string;
  sets: ExerciseLogSet[];
};

export type ExerciseLogsResponse = {
  data: ExerciseLogEntry[];
  total: number;
  limit: number;
  offset: number;
};

export function useExerciseLogs(
  exerciseId: MaybeRefOrGetter<number | null>,
  limit: MaybeRefOrGetter<number> = 20,
  offset: MaybeRefOrGetter<number> = 0,
  from: MaybeRefOrGetter<string | null> = null,
  to: MaybeRefOrGetter<string | null> = null,
) {
  const { data, isLoading, error } = useQuery({
    queryKey: computed(() =>
      exerciseKeys.logs(toValue(exerciseId) ?? 0, toValue(limit), toValue(offset), toValue(from) ?? undefined, toValue(to) ?? undefined),
    ),
    queryFn: () => {
      const id = toValue(exerciseId);
      const l = toValue(limit);
      const o = toValue(offset);
      const f = toValue(from);
      const t = toValue(to);
      const qs = new URLSearchParams({ limit: String(l), offset: String(o) });
      if (f) qs.set("from", f);
      if (t) qs.set("to", t);
      return apiClient.get<ExerciseLogsResponse>(`/exercises/${id}/logs?${qs}`);
    },
    enabled: computed(() => !!toValue(exerciseId)),
    staleTime: 0,
  });

  return {
    entries: computed(() => data.value?.data ?? []),
    total: computed(() => data.value?.total ?? 0),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
