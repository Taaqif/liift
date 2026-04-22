<script setup lang="ts">
import { ref, nextTick, watch } from "vue";
import { BotIcon, UserIcon, DumbbellIcon, CalendarDaysIcon, ChevronRightIcon } from "lucide-vue-next";
import type { ChatMessage, MessageArtifact } from "@/features/ai-chat/types";

const props = defineProps<{
  messages: ChatMessage[];
  streamingText?: string;
  streaming?: boolean;
}>();

const emit = defineEmits<{
  openArtifact: [artifact: MessageArtifact];
}>();

function parseArtifact(metadata?: string): MessageArtifact | null {
  if (!metadata) return null;
  try {
    return JSON.parse(metadata) as MessageArtifact;
  } catch {
    return null;
  }
}

function artifactLabel(a: MessageArtifact): string {
  return a.artifactType === "workout"
    ? (a.artifact as { name: string }).name
    : (a.artifact as { name: string }).name;
}

const scrollEl = ref<HTMLElement | null>(null);

watch(
  () => [props.messages.length, props.streamingText],
  () => {
    nextTick(() => {
      scrollEl.value?.scrollTo({ top: scrollEl.value.scrollHeight, behavior: "smooth" });
    });
  },
);
</script>

<template>
  <div ref="scrollEl" class="flex-1 overflow-y-auto p-4 space-y-6">
    <div
      v-for="msg in messages"
      :key="msg.id"
      class="flex gap-3"
      :class="msg.role === 'user' ? 'flex-row-reverse' : ''"
    >
      <!-- Avatar -->
      <div
        class="shrink-0 size-8 rounded-full flex items-center justify-center"
        :class="msg.role === 'user' ? 'bg-primary text-primary-foreground' : 'bg-muted'"
      >
        <UserIcon v-if="msg.role === 'user'" class="size-4" />
        <BotIcon v-else class="size-4" />
      </div>

      <!-- Bubble + artifact card -->
      <div class="flex flex-col gap-2 max-w-[75%]">
        <div
          class="rounded-2xl px-4 py-2.5 text-sm whitespace-pre-wrap"
          :class="
            msg.role === 'user'
              ? 'bg-primary text-primary-foreground rounded-tr-sm'
              : 'bg-muted rounded-tl-sm'
          "
        >
          {{ msg.content }}
        </div>

        <!-- Artifact card -->
        <button
          v-if="parseArtifact(msg.metadata)"
          class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3 text-left text-sm hover:bg-accent hover:text-accent-foreground transition-colors group"
          @click="emit('openArtifact', parseArtifact(msg.metadata)!)"
        >
          <DumbbellIcon v-if="parseArtifact(msg.metadata)!.artifactType === 'workout'" class="size-4 shrink-0 text-muted-foreground group-hover:text-accent-foreground" />
          <CalendarDaysIcon v-else class="size-4 shrink-0 text-muted-foreground group-hover:text-accent-foreground" />
          <div class="flex-1 min-w-0">
            <p class="font-medium truncate">{{ artifactLabel(parseArtifact(msg.metadata)!) }}</p>
            <p class="text-xs text-muted-foreground">{{ parseArtifact(msg.metadata)!.artifactType === 'workout' ? 'Workout' : 'Workout Plan' }} · Click to review</p>
          </div>
          <ChevronRightIcon class="size-4 shrink-0 text-muted-foreground" />
        </button>
      </div>
    </div>

    <!-- Streaming response -->
    <div v-if="streaming || streamingText" class="flex gap-3">
      <div class="shrink-0 size-8 rounded-full flex items-center justify-center bg-muted">
        <BotIcon class="size-4" />
      </div>
      <div class="max-w-[75%] rounded-2xl rounded-tl-sm px-4 py-2.5 text-sm bg-muted whitespace-pre-wrap">
        <span v-if="streamingText">{{ streamingText }}</span>
        <span v-else class="flex gap-1 items-center h-5">
          <span class="size-1.5 rounded-full bg-current animate-bounce [animation-delay:-0.3s]" />
          <span class="size-1.5 rounded-full bg-current animate-bounce [animation-delay:-0.15s]" />
          <span class="size-1.5 rounded-full bg-current animate-bounce" />
        </span>
      </div>
    </div>

  </div>
</template>
