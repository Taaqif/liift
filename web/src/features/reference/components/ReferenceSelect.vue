<script setup lang="ts">
import { computed, ref, watch } from "vue";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useReferenceOptions } from "../composables/useReferenceOptions";
import type { ReferenceType } from "../types";

const ALL_VALUE = "__all__";

const props = withDefaults(
  defineProps<{
    referenceType: ReferenceType;
    modelValue: string;
    placeholder?: string;
    allOptionLabel?: string;
    class?: string;
  }>(),
  {
    placeholder: "Select...",
    allOptionLabel: "",
  },
);

const emits = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const { options, loading } = useReferenceOptions(props.referenceType);

const innerValue = ref(props.modelValue || (props.allOptionLabel ? ALL_VALUE : ""));

watch(
  () => props.modelValue,
  (v) => {
    innerValue.value = v || (props.allOptionLabel ? ALL_VALUE : "");
  },
);

const selectOptions = computed(() =>
  props.allOptionLabel
    ? [{ value: ALL_VALUE, label: props.allOptionLabel }, ...options.value]
    : options.value,
);

function onUpdate(value: string) {
  const emitted = value === ALL_VALUE ? "" : value;
  innerValue.value = value;
  emits("update:modelValue", emitted);
}
</script>

<template>
  <Select :model-value="innerValue" :disabled="loading" @update:model-value="onUpdate">
    <SelectTrigger :class="props.class">
      <SelectValue :placeholder="placeholder" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="opt in selectOptions" :key="opt.value" :value="opt.value">
        {{ opt.label }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>
