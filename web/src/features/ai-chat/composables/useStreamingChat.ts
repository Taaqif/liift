import { ref } from "vue";
import { useQueryClient } from "@tanstack/vue-query";
import { apiClient } from "@/lib/api";
import { chatSessionKeys } from "@/lib/queryKeys";
import type { ChatMessage, ActiveArtifact, SSEEvent } from "@/features/ai-chat/types";

export function useStreamingChat(sessionSlug: string) {
  const queryClient = useQueryClient();
  const streaming = ref(false);
  const streamingText = ref("");
  const error = ref<string | null>(null);
  const artifact = ref<ActiveArtifact | null>(null);

  async function sendMessage(content: string): Promise<void> {
    if (streaming.value) return;

    streaming.value = true;
    streamingText.value = "";
    error.value = null;

    const token = apiClient.getToken();

    try {
      const response = await fetch(`/api/chats/${sessionSlug}/messages`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: token ? `Bearer ${token}` : "",
        },
        body: JSON.stringify({ content }),
      });

      if (!response.ok) {
        const body = await response.json().catch(() => ({ error: "Request failed" }));
        throw new Error(body.error ?? `HTTP ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error("No response body");

      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });

        // Process complete SSE lines
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          if (!line.startsWith("data: ")) continue;
          const jsonStr = line.slice(6).trim();
          if (!jsonStr) continue;

          try {
            const event = JSON.parse(jsonStr) as SSEEvent;
            handleEvent(event);
          } catch {
            // malformed JSON, skip
          }
        }
      }
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Unknown error";
    } finally {
      streaming.value = false;
      streamingText.value = "";
      // Refresh session messages
      queryClient.invalidateQueries({ queryKey: chatSessionKeys.detail(sessionSlug) });
      queryClient.invalidateQueries({ queryKey: chatSessionKeys.list() });
    }
  }

  function handleEvent(event: SSEEvent) {
    switch (event.type) {
      case "text_delta":
        streamingText.value += event.data.delta;
        break;
      case "artifact":
        artifact.value = {
          type: event.data.artifactType,
          data: event.data.artifact,
        };
        break;
      case "done":
        break;
      case "error":
        error.value = event.data.error;
        break;
    }
  }

  return {
    sendMessage,
    streaming,
    streamingText,
    error,
    artifact,
  };
}

// Optimistic message helper for the chat UI
export function createOptimisticUserMessage(sessionId: number, content: string): ChatMessage {
  return {
    id: -Date.now(),
    session_id: sessionId,
    role: "user",
    content,
    created_at: new Date().toISOString(),
  };
}
