<script setup lang="ts">
import { computed, ref } from "vue";
import { VueDraggable } from "vue-draggable-plus";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";
import type { Workout } from "@/features/workouts/types";
import { Button } from "@/components/ui/button";
import { GripVertical, X, ChevronDown, Plus } from "lucide-vue-next";
import InlineWorkoutEditor from "./InlineWorkoutEditor.vue";
import WorkoutPickerSheet from "./WorkoutPickerSheet.vue";

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
}>();

const { workouts, loading } = useWorkouts({ limit: 500, includeAll: true });
const pickerOpen = ref(false);
const expandedId = ref<number | null>(null);

const ids = computed(() => props.modelValue ?? []);

const workoutById = computed(() => {
  const map = new Map<number, Workout>();
  workouts.value.forEach((w) => map.set(w.id, w));
  return map;
});

const orderedItems = computed(() =>
  ids.value
    .map((id) => workoutById.value.get(id))
    .filter((w): w is Workout => !!w),
);

function handlePickerAdd(newIds: number[]) {
  const toAdd = newIds.filter((id) => !ids.value.includes(id));
  if (toAdd.length > 0) emits("update:modelValue", [...ids.value, ...toAdd]);
}

function removeAt(index: number) {
  const id = ids.value[index];
  if (expandedId.value === id) expandedId.value = null;
  const next = ids.value.filter((_, i) => i !== index);
  emits("update:modelValue", next);
}

function toggleExpand(id: number) {
  expandedId.value = expandedId.value === id ? null : id;
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
    <p
      v-if="!loading && workouts.length === 0"
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
        class="rounded border bg-background text-sm overflow-hidden"
      >
        <div class="flex items-center gap-2 px-3 py-2">
          <span
            class="workout-list-drag-handle cursor-grab touch-none text-muted-foreground hover:text-foreground active:cursor-grabbing"
            aria-hidden="true"
          >
            <GripVertical class="h-4 w-4" />
          </span>
          <span class="flex-1 min-w-0 truncate">{{ item.name }}</span>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            class="h-8 w-8 shrink-0"
            @click="toggleExpand(item.id)"
          >
            <ChevronDown
              class="h-4 w-4 transition-transform duration-200"
              :class="{ 'rotate-180': expandedId === item.id }"
            />
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

        <div
          v-if="expandedId !== item.id && item.exercises?.length"
          class="px-3 pb-2 text-xs text-muted-foreground truncate"
        >
          {{ item.exercises.map((e) => e.exercise?.name).filter(Boolean).join(" · ") }}
        </div>

        <InlineWorkoutEditor
          v-if="expandedId === item.id"
          :workout-id="item.id"
          :scroll-ref="scrollRef"
          @close="expandedId = null"
        />
      </div>
    </VueDraggable>
    <Button
      type="button"
      variant="outline"
      class="w-full"
      :disabled="loading || workouts.length === 0"
      @click="pickerOpen = true"
    >
      <Plus class="w-4 h-4 mr-2" />
      {{ placeholder }}
    </Button>
    <WorkoutPickerSheet
      :open="pickerOpen"
      @update:open="pickerOpen = $event"
      @add="handlePickerAdd"
    />
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
