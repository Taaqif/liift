import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { Exercise } from "@/features/exercises/types";
import { exerciseKeys } from "@/lib/queryKeys";

export function useExercise(id: MaybeRefOrGetter<number | null | undefined>) {
  const { data, isLoading, error } = useQuery({
    queryKey: computed(() => exerciseKeys.detail(toValue(id) ?? 0)),
    queryFn: () => apiClient.get<Exercise>(`/exercises/${toValue(id)}`),
    enabled: computed(() => !!toValue(id)),
  });

  return {
    exercise: computed(() => data.value ?? null),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
  };
}
