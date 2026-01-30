<script setup lang="ts">
import { ref, watch } from "vue";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import { useReferenceOptions } from "../composables/useReferenceOptions";
import type { ReferenceType } from "../types";

const props = withDefaults(
  defineProps<{
    referenceType: ReferenceType;
    modelValue: string[];
    placeholder?: string;
    class?: string;
  }>(),
  {
    placeholder: "Select...",
  },
);

const emits = defineEmits<{
  (e: "update:modelValue", value: string[]): void;
}>();

const { options, loading } = useReferenceOptions(props.referenceType);

const innerValue = ref<string[]>([...props.modelValue]);

watch(
  () => props.modelValue,
  (v) => {
    innerValue.value = [...(v ?? [])];
  },
  { deep: true },
);

function onUpdate(value: string[]) {
  innerValue.value = value;
  emits("update:modelValue", value);
}
</script>

<template>
  <MultiSelectTags :model-value="innerValue" :options="options" :placeholder="placeholder" :disabled="loading"
    :class="props.class" @update:model-value="onUpdate" />
</template>
