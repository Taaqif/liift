import { computed, type MaybeRefOrGetter, toValue } from "vue";
import { useQuery, useMutation, useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { chatSessionKeys } from "@/lib/queryKeys";
import type { ChatSession, ChatSessionDetail } from "@/features/ai-chat/types";

interface ChatSessionsListResponse {
  data: ChatSession[];
  total: number;
  limit: number;
  offset: number;
}

export function useChatSessions() {
  const { data, isLoading, error } = useQuery({
    queryKey: chatSessionKeys.list(),
    queryFn: () => apiClient.get<ChatSessionsListResponse>("/chats").then((r) => r.data),
  });

  return {
    sessions: data,
    loading: isLoading,
    error,
  };
}

export function useChatSession(slug: MaybeRefOrGetter<string>) {
  const { data, isLoading, error } = useQuery({
    queryKey: computed(() => chatSessionKeys.detail(toValue(slug))),
    queryFn: () => apiClient.get<ChatSessionDetail>(`/chats/${toValue(slug)}`),
    enabled: computed(() => !!toValue(slug)),
  });

  return {
    session: data,
    loading: isLoading,
    error,
  };
}

export function useCreateChatSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (title?: string) =>
      apiClient.post<ChatSession>("/chats", { title: title ?? "New Chat" }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: chatSessionKeys.list() });
    },
  });
}

export function useDeleteChatSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (slug: string) => apiClient.delete(`/chats/${slug}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: chatSessionKeys.list() });
    },
  });
}
