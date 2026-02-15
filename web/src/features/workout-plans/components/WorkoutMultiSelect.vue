<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";

const props = withDefaults(
  defineProps<{
    modelValue: number[];
    placeholder?: string;
    class?: string;
  }>(),
  {
    placeholder: "Select workouts...",
  },
);

const emits = defineEmits<{
  (e: "update:modelValue", value: number[]): void;
}>();

const { workouts, loading } = useWorkouts({ limit: 500 });

const options = computed(() =>
  workouts.value
    .map((w) => ({ value: w.id.toString(), label: w.name }))
    .sort((a, b) => a.label.localeCompare(b.label)),
);

const normalizedModel = computed(() =>
  (props.modelValue ?? []).map((id) => id.toString()),
);

const innerValue = ref<string[]>([...normalizedModel.value]);

watch(
  normalizedModel,
  (v) => {
    innerValue.value = [...v];
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
  <div class="space-y-1">
    <MultiSelectTags
      :model-value="innerValue"
      :options="options"
      :placeholder="placeholder"
      :disabled="loading"
      :class="props.class"
      @update:model-value="onUpdate"
    />
    <p
      v-if="!loading && options.length === 0"
      class="text-xs text-muted-foreground"
    >
      {{ $t("workoutPlans.noWorkoutsYet") }}
    </p>
  </div>
</template>
