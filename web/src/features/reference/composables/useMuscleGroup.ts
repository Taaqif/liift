import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { MuscleGroupResponse } from "@/features/reference/types";
import { referenceKeys } from "@/lib/queryKeys";

async function fetchMuscleGroup() {
  return apiClient.get<MuscleGroupResponse>("/");
}

export function useMuscleGroup() {
  const {
    data: equipment,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: referenceKeys.muscleGroup(),
    queryFn: fetchMuscleGroup,
  });

  return {
    muscleGroup: computed(() => equipment.value ?? []),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}
