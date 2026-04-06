<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { VueDraggable } from "vue-draggable-plus";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";
import type { Workout } from "@/features/workouts/types";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { GripVertical, X, Pencil } from "lucide-vue-next";

const props = withDefaults(
  defineProps<{
    modelValue: number[];
    placeholder?: string;
    class?: string;
    scrollRef?: HTMLElement | null;
  }>(),
  {
    placeholder: "Add workout...",
  },
);

const emits = defineEmits<{
  (e: "update:modelValue", value: number[]): void;
  (e: "edit", workoutId: number): void;
}>();

const { workouts, loading } = useWorkouts({ limit: 500, includeAll: true });
const selectedIdToAdd = ref<string | undefined>(undefined);

const ids = computed(() => props.modelValue ?? []);

const workoutById = computed(() => {
  const map = new Map<number, Workout>();
  workouts.value.forEach((w) => map.set(w.id, w));
  return map;
});

const options = computed(() =>
  workouts.value
    .filter((w) => w.is_library !== false && !ids.value.includes(w.id))
    .map((w) => ({ value: w.id.toString(), label: w.name }))
    .sort((a, b) => a.label.localeCompare(b.label)),
);

const orderedItems = computed(() =>
  ids.value.map((id) => workoutById.value.get(id)).filter((w): w is Workout => !!w),
);

watch(selectedIdToAdd, (val) => {
  if (val) {
    const id = parseInt(val, 10);
    if (!Number.isNaN(id) && !ids.value.includes(id)) {
      emits("update:modelValue", [...ids.value, id]);
    }
    selectedIdToAdd.value = undefined;
  }
});

function removeAt(index: number) {
  const next = ids.value.filter((_, i) => i !== index);
  emits("update:modelValue", next);
}

function onReorder(event: { oldIndex?: number; newIndex?: number }) {
  const oldIndex = event.oldIndex ?? 0;
  const newIndex = event.newIndex ?? 0;
  if (oldIndex === newIndex) return;
  const next = [...ids.value];
  const [removed] = next.splice(oldIndex, 1);
  next.splice(newIndex, 0, removed);
  emits("update:modelValue", next);
}
</script>

<template>
  <div class="space-y-2" :class="props.class">
    <Select
      v-model="selectedIdToAdd"
      :disabled="loading || options.length === 0"
    >
      <SelectTrigger class="w-full">
        <SelectValue :placeholder="placeholder" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem
          v-for="opt in options"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </SelectItem>
      </SelectContent>
    </Select>

    <p
      v-if="!loading && workouts.length > 0 && options.length === 0 && ids.length > 0"
      class="text-xs text-muted-foreground"
    >
      {{ $t("workoutPlans.allWorkoutsAdded") }}
    </p>
    <p
      v-else-if="!loading && workouts.length === 0"
      class="text-xs text-muted-foreground"
    >
      {{ $t("workoutPlans.noWorkoutsYet") }}
    </p>

    <VueDraggable
      v-if="orderedItems.length > 0"
      :model-value="orderedItems"
      :custom-update="onReorder"
      handle=".workout-list-drag-handle"
      :force-fallback="true"
      :fallback-on-body="true"
      ghost-class="workout-drag-ghost"
      chosen-class="workout-drag-chosen"
      fallback-class="workout-drag-fallback"
      :animation="200"
      :scroll="scrollRef || true"
      :bubble-scroll="true"
      :scroll-sensitivity="80"
      :scroll-speed="16"
      class="space-y-1"
    >
      <div
        v-for="(item, index) in orderedItems"
        :key="item.id"
        class="flex items-center gap-2 rounded border bg-background px-3 py-2 text-sm"
      >
        <span
          class="workout-list-drag-handle cursor-grab touch-none text-muted-foreground hover:text-foreground active:cursor-grabbing"
          aria-hidden="true"
        >
          <GripVertical class="h-4 w-4" />
        </span>
        <span class="flex-1">{{ item.name }}</span>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-8 w-8 shrink-0"
          @click="emits('edit', item.id)"
        >
          <Pencil class="h-3.5 w-3.5" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-8 w-8 shrink-0"
          @click="removeAt(index)"
        >
          <X class="h-4 w-4" />
        </Button>
      </div>
    </VueDraggable>
  </div>
</template>

<style scoped>
:deep(.workout-drag-ghost) {
  border-radius: 0.5rem;
  border: 2px dashed hsl(var(--primary) / 0.4);
  background: hsl(var(--primary) / 0.05);
  opacity: 0.9;
}

:deep(.workout-drag-chosen) {
  opacity: 0.4;
}
</style>

<style>
.workout-drag-fallback {
  border-radius: 0.5rem;
  border: 1px solid hsl(var(--border));
  background: hsl(var(--background));
  box-shadow: 0 12px 40px -12px rgb(0 0 0 / 0.25);
  cursor: grabbing;
  z-index: 9999;
}
</style>
