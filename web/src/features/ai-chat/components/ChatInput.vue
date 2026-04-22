<script setup lang="ts">
import { ref } from "vue";
import { SendHorizonalIcon } from "lucide-vue-next";
import { Button } from "@/components/ui/button";

const props = defineProps<{
  disabled?: boolean;
}>();

const emit = defineEmits<{
  send: [content: string];
}>();

const text = ref("");

function handleSend() {
  const content = text.value.trim();
  if (!content || props.disabled) return;
  emit("send", content);
  text.value = "";
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    handleSend();
  }
}
</script>

<template>
  <div class="border-t p-4">
    <div class="flex gap-2 items-end">
      <textarea
        v-model="text"
        :disabled="disabled"
        rows="1"
        placeholder="Message Coach..."
        class="flex-1 resize-none rounded-xl border bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50 min-h-[40px] max-h-32"
        style="field-sizing: content"
        @keydown="handleKeydown"
      />
      <Button
        size="icon"
        :disabled="disabled || !text.trim()"
        class="rounded-xl shrink-0"
        @click="handleSend"
      >
        <SendHorizonalIcon class="size-4" />
      </Button>
    </div>
    <p class="text-xs text-muted-foreground mt-2">
      Shift+Enter for new line. Coach may make mistakes — verify workouts before saving.
    </p>
  </div>
</template>
