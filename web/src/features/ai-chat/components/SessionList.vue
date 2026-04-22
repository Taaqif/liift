<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { PlusIcon, Trash2Icon, MessageSquareIcon } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import {
  useChatSessions,
  useDeleteChatSession,
} from "@/features/ai-chat/composables/useChatSessions";

const props = defineProps<{
  activeSessionSlug?: string;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  select: [slug: string];
}>();

const router = useRouter();
const { sessions, loading } = useChatSessions();
const { mutateAsync: deleteSession } = useDeleteChatSession();

const deletingSlug = ref<string | null>(null);

function handleNew() {
  router.push("/coach");
}

async function handleDelete(e: Event, slug: string) {
  e.stopPropagation();
  deletingSlug.value = slug;
  try {
    await deleteSession(slug);
    if (props.activeSessionSlug === slug) {
      router.push("/coach");
    }
  } finally {
    deletingSlug.value = null;
  }
}
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto p-2 space-y-0.5">
      <div v-if="loading" class="p-3 text-sm text-muted-foreground">Loading...</div>
      <div v-else-if="!sessions?.length" class="p-3 text-sm text-muted-foreground">
        No chats yet. Start one!
      </div>
      <button
        v-for="session in sessions"
        :key="session.slug"
        class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-left text-sm hover:bg-accent transition-colors group"
        :class="{ 'bg-accent': session.slug === activeSessionSlug }"
        @click="emit('select', session.slug)"
      >
        <MessageSquareIcon class="size-3.5 shrink-0 text-muted-foreground" />
        <span class="flex-1 truncate">{{ session.title }}</span>
        <button
          class="shrink-0 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity"
          :disabled="deletingSlug === session.slug"
          @click="handleDelete($event, session.slug)"
        >
          <Trash2Icon class="size-3.5" />
        </button>
      </button>
    </div>

    <div class="p-3 border-t">
      <Button class="w-full" size="sm" :disabled="props.disabled" @click="handleNew">
        <PlusIcon class="size-4 mr-1.5" />
        New Chat
      </Button>
    </div>
  </div>
</template>
