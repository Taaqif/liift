import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import type { EquipmentResponse } from "@/features/reference/types";
import { referenceKeys } from "@/lib/queryKeys";

async function fetchEquipment() {
  return apiClient.get<EquipmentResponse>("/equipment");
}

export function useEquipment() {
  const {
    data: equipment,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: referenceKeys.equipment(),
    queryFn: fetchEquipment,
  });

  return {
    equipment: computed(() => equipment.value ?? []),
    loading: computed(() => isLoading.value),
    error: computed(() => error.value),
    refetch,
  };
}
