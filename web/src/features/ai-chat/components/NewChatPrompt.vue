<script setup lang="ts">
import { BotIcon, DumbbellIcon, CalendarDaysIcon, BookOpenIcon, ChartBarIcon } from "lucide-vue-next";
import ChatInput from "@/features/ai-chat/components/ChatInput.vue";

const props = defineProps<{ disabled?: boolean }>();

const emit = defineEmits<{
  send: [content: string];
}>();

const starters = [
  {
    icon: DumbbellIcon,
    label: "Create a workout",
    message: "Help me create a new custom workout tailored to my goals and available equipment.",
  },
  {
    icon: CalendarDaysIcon,
    label: "Build a workout plan",
    message: "Help me build a structured workout plan to follow over the coming weeks.",
  },
  {
    icon: BookOpenIcon,
    label: "Learn about an exercise",
    message: "Tell me about a specific exercise — proper form, muscles worked, and tips.",
  },
  {
    icon: ChartBarIcon,
    label: "Review my training",
    message: "Can you review my recent training and give me feedback or suggestions?",
  },
];
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Centered content -->
    <div class="flex-1 flex flex-col items-center justify-center gap-6 p-6 overflow-y-auto">
      <div class="flex flex-col items-center gap-3 text-center">
        <div class="size-14 rounded-2xl bg-muted flex items-center justify-center">
          <BotIcon class="size-7 text-muted-foreground" />
        </div>
        <div class="space-y-1">
          <p class="font-semibold text-lg">How can I help?</p>
          <p class="text-sm text-muted-foreground max-w-xs">
            Ask me anything about training, or pick a topic to get started.
          </p>
        </div>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 w-full max-w-md">
        <button
          v-for="s in starters"
          :key="s.label"
          :disabled="props.disabled"
          class="flex items-center gap-3 rounded-xl border bg-card px-4 py-3 text-left text-sm hover:bg-accent hover:text-accent-foreground transition-colors disabled:opacity-50 disabled:pointer-events-none"
          @click="emit('send', s.message)"
        >
          <component :is="s.icon" class="size-4 shrink-0 text-muted-foreground" />
          <span class="font-medium">{{ s.label }}</span>
        </button>
      </div>
    </div>

    <!-- Input pinned to bottom -->
    <ChatInput :disabled="props.disabled" @send="emit('send', $event)" />
  </div>
</template>
