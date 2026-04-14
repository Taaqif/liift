<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { useI18n } from "vue-i18n";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import { useMuscleGroupOptions } from "../composables/useMuscleGroupOptions";

const { t } = useI18n();

const props = withDefaults(
  defineProps<{
    modelValue: string[];
    placeholder?: string;
    class?: string;
    listMode?: boolean;
    single?: boolean;
  }>(),
  {
    placeholder: undefined,
    listMode: false,
    single: false,
  },
);

const placeholder = computed(() => props.placeholder ?? t("selectPlaceholder"));

const emits = defineEmits<{
  (e: "update:modelValue", value: string[]): void;
}>();

const { options, loading } = useMuscleGroupOptions();
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
  <MultiSelectTags
    :model-value="innerValue"
    :options="options"
    :placeholder="placeholder"
    :disabled="loading"
    :class="props.class"
    :list-mode="props.listMode"
    :single="props.single"
    @update:model-value="onUpdate"
  />
</template>
