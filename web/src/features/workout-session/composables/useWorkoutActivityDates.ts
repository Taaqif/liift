import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { workoutSessionKeys } from "@/lib/queryKeys";

type ActivityDatesResponse = {
  dates: string[];
  year: number;
  month: number;
};

export function useWorkoutActivityDates(
  year: MaybeRefOrGetter<number>,
  month: MaybeRefOrGetter<number>,
) {
  const { data, isLoading, error } = useQuery({
    queryKey: computed(() => workoutSessionKeys.activityDates(toValue(year), toValue(month))),
    queryFn: () =>
      apiClient.get<ActivityDatesResponse>(
        `/workout-sessions/activity?year=${toValue(year)}&month=${toValue(month)}`,
      ),
    staleTime: 60_000,
  });

  return {
    activityDates: computed<Set<string>>(() => new Set(data.value?.dates ?? [])),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
