import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { workoutSessionKeys } from "@/lib/queryKeys";

export type WeeklyFeatureStat = {
  feature_name: string;
  total: number;
};

export function useWeeklyStats(
  from: MaybeRefOrGetter<string>,
  to: MaybeRefOrGetter<string>,
) {
  const { data, isLoading } = useQuery({
    queryKey: computed(() =>
      workoutSessionKeys.weeklyStats(toValue(from), toValue(to)),
    ),
    queryFn: () => {
      const qs = new URLSearchParams({ from: toValue(from), to: toValue(to) });
      return apiClient.get<WeeklyFeatureStat[]>(`/workout-sessions/weekly-stats?${qs}`);
    },
    staleTime: 60_000,
  });

  const statMap = computed(() => {
    const map = new Map<string, number>();
    for (const s of data.value ?? []) {
      map.set(s.feature_name, s.total);
    }
    return map;
  });

  return {
    statMap,
    loading: computed(() => isLoading.value),
  };
}
