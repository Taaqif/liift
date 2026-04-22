<script setup lang="ts">
import { ref } from "vue";
import { BrainCircuitIcon } from "lucide-vue-next";
import AISettingsForm from "@/features/settings/components/AISettingsForm.vue";

type SettingsSection = "ai";

const activeSection = ref<SettingsSection>("ai");

const sections: { id: SettingsSection; label: string; icon: typeof BrainCircuitIcon; description: string }[] = [
  {
    id: "ai",
    label: "AI Provider",
    icon: BrainCircuitIcon,
    description: "Configure your AI provider and model for the Coach feature.",
  },
];
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-semibold">Settings</h1>
      <p class="text-sm text-muted-foreground mt-1">Manage your app preferences and integrations.</p>
    </div>

    <div class="flex flex-col md:flex-row gap-8">
      <!-- Side nav -->
      <nav class="md:w-48 shrink-0">
        <!-- Mobile: horizontal pills -->
        <div class="flex gap-1 flex-wrap md:hidden mb-2">
          <button
            v-for="s in sections"
            :key="s.id"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-colors"
            :class="
              activeSection === s.id
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:text-foreground'
            "
            @click="activeSection = s.id"
          >
            <component :is="s.icon" class="size-3.5" />
            {{ s.label }}
          </button>
        </div>

        <!-- Desktop: vertical list -->
        <div class="hidden md:flex flex-col gap-0.5">
          <button
            v-for="s in sections"
            :key="s.id"
            class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm font-medium text-left transition-colors w-full"
            :class="
              activeSection === s.id
                ? 'bg-accent text-accent-foreground'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
            "
            @click="activeSection = s.id"
          >
            <component :is="s.icon" class="size-4 shrink-0" />
            {{ s.label }}
          </button>
        </div>
      </nav>

      <!-- Content -->
      <div class="flex-1 min-w-0 max-w-lg">
        <template v-for="s in sections" :key="s.id">
          <div v-if="activeSection === s.id" class="space-y-6">
            <div>
              <h2 class="text-base font-semibold">{{ s.label }}</h2>
              <p class="text-sm text-muted-foreground mt-0.5">{{ s.description }}</p>
            </div>
            <AISettingsForm v-if="s.id === 'ai'" />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>
