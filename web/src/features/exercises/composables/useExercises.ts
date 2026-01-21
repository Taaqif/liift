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

  if (params?.search) {
    queryParams.append("q", params.search);
  }

  if (params?.muscleGroup && params.muscleGroup.length > 0) {
    for (const mg of params.muscleGroup) {
      queryParams.append("muscle_group", mg);
    }
  }

  if (params?.equipment && params.equipment.length > 0) {
    for (const eq of params.equipment) {
      queryParams.append("equipment", eq);
    }
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
    queryKey: computed(() => exerciseKeys.list(toValue(params))),
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
