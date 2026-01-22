import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { ExerciseFeatureResponse } from "@/features/reference/types";
import { referenceKeys } from "@/lib/queryKeys";

async function fetchExerciseFeatures() {
  return apiClient.get<ExerciseFeatureResponse>("/exercise-features");
}

export function useExerciseFeature() {
  const {
    data: exerciseFeatures,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: referenceKeys.exerciseFeatures(),
    queryFn: fetchExerciseFeatures,
  });

  return {
    exerciseFeatures: computed(() => exerciseFeatures.value ?? []),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}
