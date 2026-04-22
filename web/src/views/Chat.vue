<script setup lang="ts">
import { ref, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { PanelLeftIcon, PlusIcon, SettingsIcon } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent } from "@/components/ui/sheet";
import SessionList from "@/features/ai-chat/components/SessionList.vue";
import ChatWindow from "@/features/ai-chat/components/ChatWindow.vue";
import NewChatPrompt from "@/features/ai-chat/components/NewChatPrompt.vue";
import { useCreateChatSession } from "@/features/ai-chat/composables/useChatSessions";
import { useAISettings } from "@/features/settings/composables/useAISettings";

const route = useRoute();
const router = useRouter();

// Desktop: inline sidebar open/close. Mobile: sheet.
const desktopSidebarOpen = ref(true);
const mobileSidebarOpen = ref(false);

const sessionSlug = computed(() => {
  const slug = route.params.slug;
  return typeof slug === "string" && slug ? slug : "";
});

const { mutateAsync: createSession, isPending: creating } = useCreateChatSession();

const { settings, loading: settingsLoading } = useAISettings();
const isConfigured = computed(() => !!settings.value?.isConfigured);

function handleNew() {
  mobileSidebarOpen.value = false;
  router.push("/coach");
}

function handleSelect(slug: string) {
  mobileSidebarOpen.value = false;
  router.push(`/coach/${slug}`);
}

async function handleFirstMessage(content: string) {
  const session = await createSession(undefined);
  router.push({ path: `/coach/${session.slug}`, state: { initialMessage: content } });
}
</script>

<template>
  <div class="-mx-4 -my-6 md:-mx-8 md:-my-8 flex overflow-hidden" style="height: calc(var(--vh, 100dvh) - 57px)">

    <!-- Desktop sidebar (hidden on mobile) -->
    <div
      class="hidden md:flex shrink-0 border-r flex-col overflow-hidden transition-all duration-200"
      :class="desktopSidebarOpen ? 'w-64' : 'w-0'"
    >
      <SessionList
        v-if="desktopSidebarOpen"
        :active-session-slug="sessionSlug || undefined"
        :disabled="!isConfigured"
        @select="handleSelect"
      />
    </div>

    <!-- Mobile sidebar sheet -->
    <Sheet v-model:open="mobileSidebarOpen">
      <SheetContent side="left" class="p-0 w-72 flex flex-col">
        <SessionList
          :active-session-slug="sessionSlug || undefined"
          :disabled="!isConfigured"
          @select="handleSelect"
        />
      </SheetContent>
    </Sheet>

    <!-- Main area -->
    <div class="flex-1 min-w-0 flex flex-col overflow-hidden">
      <!-- Top bar -->
      <div class="flex items-center gap-1 px-2 py-2 border-b shrink-0">
        <!-- Mobile: sheet trigger -->
        <Button
          variant="ghost"
          size="icon"
          class="size-8 md:hidden"
          @click="mobileSidebarOpen = true"
        >
          <PanelLeftIcon class="size-4" />
        </Button>
        <!-- Desktop: toggle inline sidebar -->
        <Button
          variant="ghost"
          size="icon"
          class="size-8 hidden md:flex"
          @click="desktopSidebarOpen = !desktopSidebarOpen"
        >
          <PanelLeftIcon class="size-4" />
        </Button>

        <span class="text-sm font-semibold truncate flex-1 px-1">Coach</span>

        <Button variant="ghost" size="icon" class="size-8" :disabled="!isConfigured" @click="handleNew">
          <PlusIcon class="size-4" />
        </Button>
      </div>

      <!-- AI not configured -->
      <div v-if="!settingsLoading && !isConfigured" class="flex-1 flex flex-col items-center justify-center gap-4 text-center p-8">
        <SettingsIcon class="size-14 text-muted-foreground/40" />
        <div class="space-y-1">
          <p class="font-semibold">Set up AI Provider</p>
          <p class="text-sm text-muted-foreground max-w-xs">
            Configure an AI provider before chatting with Coach.
          </p>
        </div>
        <Button @click="router.push('/settings')">
          <SettingsIcon class="size-4 mr-1.5" />
          Go to Settings
        </Button>
      </div>

      <!-- No session: new chat prompt with starters -->
      <NewChatPrompt
        v-else-if="isConfigured && !sessionSlug"
        class="flex-1 min-h-0 overflow-hidden"
        :disabled="creating"
        @send="handleFirstMessage"
      />

      <!-- Active session -->
      <ChatWindow
        v-else-if="isConfigured && sessionSlug"
        :key="sessionSlug"
        :session-slug="sessionSlug"
        class="flex-1 min-h-0 overflow-hidden"
      />
    </div>
  </div>
</template>
