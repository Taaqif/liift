<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from "vue";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import ChatMessages from "@/features/ai-chat/components/ChatMessages.vue";
import ChatInput from "@/features/ai-chat/components/ChatInput.vue";
import ArtifactPanel from "@/features/ai-chat/components/ArtifactPanel.vue";
import NewChatPrompt from "@/features/ai-chat/components/NewChatPrompt.vue";
import { useChatSession } from "@/features/ai-chat/composables/useChatSessions";
import { useStreamingChat } from "@/features/ai-chat/composables/useStreamingChat";
import type { ChatMessage, ActiveArtifact, MessageArtifact } from "@/features/ai-chat/types";

const props = defineProps<{ sessionSlug: string }>();

const { session, loading } = useChatSession(() => props.sessionSlug);
const { sendMessage, streaming, streamingText, error, artifact: streamArtifact } =
  useStreamingChat(props.sessionSlug);

const artifact = ref<ActiveArtifact | null>(null);
const artifactSheetOpen = ref(false);
const localMessages = ref<ChatMessage[]>([]);
const pendingUserMessage = ref<string | null>(null);

watch(session, (s) => {
  if (s?.messages) localMessages.value = [...s.messages];
}, { immediate: true });

watch(streamArtifact, (a) => {
  if (a) {
    artifact.value = a;
    artifactSheetOpen.value = true; // auto-open on mobile
  }
});

watch(streaming, (isStreaming) => {
  if (!isStreaming) {
    if (session.value?.messages) {
      localMessages.value = [...session.value.messages];
    }
    pendingUserMessage.value = null;
  }
});

const displayMessages = computed<ChatMessage[]>(() => {
  const msgs = [...localMessages.value];
  if (pendingUserMessage.value) {
    msgs.push({
      id: -1,
      session_id: session.value?.id ?? 0,
      role: "user",
      content: pendingUserMessage.value,
      created_at: new Date().toISOString(),
    });
  }
  return msgs;
});

async function handleSend(content: string) {
  pendingUserMessage.value = content;
  await sendMessage(content);
}

onMounted(async () => {
  const initial = (history.state as Record<string, unknown>)?.initialMessage as string | undefined;
  if (initial) {
    history.replaceState({ ...history.state, initialMessage: undefined }, "");
    await nextTick();
    handleSend(initial);
  }
});

function openArtifact(a: MessageArtifact) {
  artifact.value = { type: a.artifactType, data: a.artifact };
  artifactSheetOpen.value = true;
}

function closeArtifact() {
  artifact.value = null;
  artifactSheetOpen.value = false;
}
</script>

<template>
  <div class="flex flex-1 w-full h-full overflow-hidden">
    <!-- Messages + Input -->
    <div class="flex flex-col flex-1 min-w-0 min-h-0 overflow-hidden">
      <div v-if="loading" class="flex-1 flex items-center justify-center">
        <span class="text-sm text-muted-foreground">Loading...</span>
      </div>
      <template v-else>
        <!-- No messages yet: show starters -->
        <NewChatPrompt
          v-if="!displayMessages.length && !streaming"
          class="flex-1 min-h-0 overflow-hidden"
          :disabled="streaming"
          @send="handleSend"
        />
        <template v-else>
          <ChatMessages
            :messages="displayMessages"
            :streaming-text="streamingText"
            :streaming="streaming"
            class="flex-1 min-h-0"
            @open-artifact="openArtifact"
          />
          <div v-if="error" class="px-4 py-2 text-xs text-destructive border-t bg-destructive/5">
            {{ error }}
          </div>
          <ChatInput :disabled="streaming" @send="handleSend" />
        </template>
      </template>
    </div>

    <!-- Artifact sheet (right side, all screen sizes) -->
    <Sheet v-model:open="artifactSheetOpen">
      <SheetContent side="right" class="w-full sm:w-96 p-0 flex flex-col">
        <SheetHeader class="sr-only">
          <SheetTitle>{{ artifact?.type === 'workout' ? 'Workout' : 'Workout Plan' }}</SheetTitle>
        </SheetHeader>
        <ArtifactPanel v-if="artifact" :artifact="artifact" />
      </SheetContent>
    </Sheet>
  </div>
</template>
