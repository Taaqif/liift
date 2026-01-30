import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type {
  WorkoutsListResponse,
  WorkoutsListParams,
} from "@/features/workouts/types";
import { workoutKeys } from "@/lib/queryKeys";

async function fetchWorkouts(
  params?: WorkoutsListParams,
): Promise<WorkoutsListResponse> {
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
  if (params?.exerciseFeatures?.length) {
    for (const f of params.exerciseFeatures) {
      queryParams.append("exercise_feature", f);
    }
  }
  if (params?.exerciseIds?.length) {
    for (const id of params.exerciseIds) {
      queryParams.append("exercise_id", id.toString());
    }
  }
  if (params?.muscleGroup?.length) {
    for (const mg of params.muscleGroup) {
      queryParams.append("muscle_group", mg);
    }
  }
  if (params?.equipment?.length) {
    for (const eq of params.equipment) {
      queryParams.append("equipment", eq);
    }
  }

  const queryString = queryParams.toString();
  const url = `/workouts${queryString ? `?${queryString}` : ""}`;

  return apiClient.get<WorkoutsListResponse>(url);
}

export function useWorkouts(params?: MaybeRefOrGetter<WorkoutsListParams>) {
  const {
    data: workoutsData,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: computed(() => workoutKeys.list(toValue(params))),
    queryFn: () => fetchWorkouts(toValue(params)),
  });

  return {
    workouts: computed(() => workoutsData.value?.data ?? []),
    total: computed(() => workoutsData.value?.total ?? 0),
    limit: computed(() => workoutsData.value?.limit ?? 20),
    offset: computed(() => workoutsData.value?.offset ?? 0),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}
