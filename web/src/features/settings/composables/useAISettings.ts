import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { aiSettingsKeys } from "@/lib/queryKeys";
import type { AISettings, UpdateAISettingsPayload, ProviderInfo } from "@/features/settings/types";

export function useAISettings() {
  const { data, isLoading, error } = useQuery({
    queryKey: aiSettingsKeys.settings(),
    queryFn: () => apiClient.get<AISettings>("/ai/settings"),
  });

  return { settings: data, loading: isLoading, error };
}

export function useAIProviders() {
  const { data, isLoading } = useQuery({
    queryKey: aiSettingsKeys.providers(),
    queryFn: () => apiClient.get<ProviderInfo[]>("/ai/providers"),
  });

  return { providers: data, loading: isLoading };
}

export function useUpdateAISettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: UpdateAISettingsPayload) =>
      apiClient.put<AISettings>("/ai/settings", payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: aiSettingsKeys.settings() });
    },
  });
}
