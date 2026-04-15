import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { profileKeys } from "@/lib/queryKeys";
import type { Profile } from "@/features/profile/types";

export function useProfile() {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: profileKeys.me(),
    queryFn: () => apiClient.get<Profile>("/users/me"),
    staleTime: 1000 * 60 * 5,
    retry: false,
  });

  return {
    profile: computed(() => data.value ?? null),
    loading: isLoading,
    error,
    refetch,
  };
}
