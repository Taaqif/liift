<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import { useExercises } from "../composables/useExercises";

const props = withDefaults(
  defineProps<{
    modelValue: number[];
    placeholder?: string;
    class?: string;
    listMode?: boolean;
    single?: boolean;
  }>(),
  {
    placeholder: "Select exercises...",
    listMode: false,
    single: false,
  },
);

const emits = defineEmits<{
  (e: "update:modelValue", value: number[]): void;
}>();

const { exercises, loading } = useExercises({ limit: 1000 });

const options = computed(() =>
  exercises.value
    .map((e) => ({ value: e.id.toString(), label: e.name }))
    .sort((a, b) => a.label.localeCompare(b.label)),
);

const innerValue = ref<string[]>(
  props.modelValue.map((id) => id.toString()),
);

watch(
  () => props.modelValue,
  (v) => {
    innerValue.value = (v ?? []).map((id) => id.toString());
  },
  { deep: true },
);

function onUpdate(value: string[]) {
  innerValue.value = value;
  emits(
    "update:modelValue",
    value.map((s) => parseInt(s, 10)).filter((n) => !Number.isNaN(n)),
  );
}
</script>

<template>
  <MultiSelectTags :model-value="innerValue" :options="options" :placeholder="placeholder" :disabled="loading"
    :class="props.class" :list-mode="props.listMode" @update:model-value="onUpdate" />
</template>
