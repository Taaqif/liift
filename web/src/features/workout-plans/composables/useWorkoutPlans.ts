import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { WorkoutPlan } from "../types";
import { workoutPlanKeys } from "@/lib/queryKeys";

export type WorkoutPlansListParams = {
  limit?: number;
  offset?: number;
};

export type WorkoutPlansListResponse = {
  data: WorkoutPlan[];
  total: number;
  limit: number;
  offset: number;
};

async function fetchWorkoutPlans(
  params?: WorkoutPlansListParams,
): Promise<WorkoutPlansListResponse> {
  const queryParams = new URLSearchParams();
  if (params?.limit != null) {
    queryParams.append("limit", params.limit.toString());
  }
  if (params?.offset != null) {
    queryParams.append("offset", params.offset.toString());
  }
  const queryString = queryParams.toString();
  const url = `/workout-plans${queryString ? `?${queryString}` : ""}`;
  return apiClient.get<WorkoutPlansListResponse>(url);
}

export function useWorkoutPlans(
  params?: MaybeRefOrGetter<WorkoutPlansListParams>,
) {
  const {
    data: plansData,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: computed(() => workoutPlanKeys.list(toValue(params))),
    queryFn: () => fetchWorkoutPlans(toValue(params)),
  });

  return {
    plans: computed(() => plansData.value?.data ?? []),
    total: computed(() => plansData.value?.total ?? 0),
    limit: computed(() => plansData.value?.limit ?? 20),
    offset: computed(() => plansData.value?.offset ?? 0),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}
