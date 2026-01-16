import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type {
  ExercisesListResponse,
  ExercisesListParams,
} from "@/features/exercises/types";
import { exerciseKeys } from "@/lib/queryKeys";

async function fetchExercises(
  params?: ExercisesListParams,
): Promise<ExercisesListResponse> {
  const queryParams = new URLSearchParams();
  if (params?.limit) {
    queryParams.append("limit", params.limit.toString());
  }
  if (params?.offset) {
    queryParams.append("offset", params.offset.toString());
  }

  const queryString = queryParams.toString();
  const url = `/exercises${queryString ? `?${queryString}` : ""}`;

  return apiClient.get<ExercisesListResponse>(url);
}

export function useExercises(params?: MaybeRefOrGetter<ExercisesListParams>) {
  const {
    data: exercisesData,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: exerciseKeys.list(toValue(params)),
    queryFn: () => fetchExercises(toValue(params)),
  });

  return {
    exercises: computed(() => exercisesData.value?.data ?? []),
    total: computed(() => exercisesData.value?.total ?? 0),
    limit: computed(() => exercisesData.value?.limit ?? 20),
    offset: computed(() => exercisesData.value?.offset ?? 0),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}

