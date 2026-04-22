<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAISettings, useAIProviders, useUpdateAISettings } from "@/features/settings/composables/useAISettings";

const { settings } = useAISettings();
const { providers } = useAIProviders();
const { mutateAsync: updateSettings, isPending } = useUpdateAISettings();

const provider = ref("");
const apiKey = ref("");
const model = ref("");
const ollamaBaseURL = ref("");
const customBaseURL = ref("");

watch(
  settings,
  (s) => {
    if (!s) return;
    provider.value = s.provider;
    model.value = s.model;
    ollamaBaseURL.value = s.ollamaBaseURL;
    customBaseURL.value = s.customBaseURL;
  },
  { immediate: true },
);

const selectedProvider = computed(() =>
  providers.value?.find((p) => p.id === provider.value),
);

watch(provider, (newProvider) => {
  const p = providers.value?.find((prov) => prov.id === newProvider);
  if (p) {
    model.value = p.defaultModel;
  }
});

async function save() {
  try {
    await updateSettings({
      provider: provider.value,
      apiKey: apiKey.value || undefined,
      model: model.value || undefined,
      ollamaBaseURL: ollamaBaseURL.value || undefined,
      customBaseURL: customBaseURL.value || undefined,
    });
    apiKey.value = "";
    toast.success("AI settings saved");
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "Failed to save settings");
  }
}
</script>

<template>
  <form class="space-y-6" @submit.prevent="save">
    <!-- Provider -->
    <div class="space-y-2">
      <Label>Provider</Label>
      <Select :model-value="provider" @update:model-value="provider = $event as string">
        <SelectTrigger>
          <SelectValue placeholder="Select provider" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="p in providers" :key="p.id" :value="p.id">
            {{ p.name }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <!-- API Key -->
    <div v-if="selectedProvider?.needsApiKey" class="space-y-2">
      <Label>API Key</Label>
      <Input
        v-model="apiKey"
        type="password"
        :placeholder="settings?.hasApiKey ? `Current key: ${settings.apiKeyMasked}` : 'Enter API key'"
        autocomplete="off"
      />
      <p v-if="settings?.hasApiKey" class="text-xs text-muted-foreground">
        Leave blank to keep existing key.
      </p>
    </div>

    <!-- Ollama base URL -->
    <div v-if="provider === 'ollama'" class="space-y-2">
      <Label>Ollama Base URL</Label>
      <Input v-model="ollamaBaseURL" placeholder="http://localhost:11434" />
    </div>

    <!-- Custom base URL -->
    <div v-if="provider === 'custom'" class="space-y-2">
      <Label>Base URL</Label>
      <Input v-model="customBaseURL" placeholder="https://openrouter.ai/api/v1" />
      <p class="text-xs text-muted-foreground">
        Any OpenAI-compatible endpoint, e.g. OpenRouter, LM Studio, vLLM.
      </p>
    </div>

    <!-- Model — dropdown for known providers, free text for custom -->
    <div v-if="selectedProvider" class="space-y-2">
      <Label>Model</Label>
      <Input
        v-if="provider === 'custom'"
        v-model="model"
        placeholder="e.g. openai/gpt-4o, mistralai/mistral-7b"
      />
      <Select v-else :model-value="model" @update:model-value="model = $event as string">
        <SelectTrigger>
          <SelectValue placeholder="Select model" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="m in selectedProvider.models" :key="m" :value="m">
            {{ m }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <Button type="submit" :disabled="isPending">
      {{ isPending ? "Saving..." : "Save Settings" }}
    </Button>
  </form>
</template>
