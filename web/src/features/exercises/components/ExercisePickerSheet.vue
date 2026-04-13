<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import type { Exercise } from "@/features/exercises/types";
import { useExercises } from "@/features/exercises/composables/useExercises";
import ExerciseList from "./ExerciseList.vue";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  add: [exercises: Exercise[]];
}>();

const { t } = useI18n();

const LIMIT = 15;
const offset = ref(0);
const search = ref("");

const params = computed(() => ({
  limit: LIMIT,
  offset: offset.value,
  search: search.value.trim() || undefined,
}));

const { exercises, total, loading } = useExercises(params);

const selectedIds = ref<Set<number>>(new Set());

const selectedCount = computed(() => selectedIds.value.size);

const totalPages = computed(() => Math.ceil((total.value ?? 0) / LIMIT));
const currentPage = computed(() => Math.floor(offset.value / LIMIT) + 1);

function handleSearchInput(value: string) {
  search.value = value;
  offset.value = 0;
}

function handleSelect(exercise: Exercise) {
  const next = new Set(selectedIds.value);
  if (next.has(exercise.id)) {
    next.delete(exercise.id);
  } else {
    next.add(exercise.id);
  }
  selectedIds.value = next;
}

function handleAdd() {
  emit("add", [...allSelected.value.values()]);
  close();
}

// Cache selected exercises across pages so user can paginate and select from multiple pages
const selectedExerciseCache = ref<Map<number, Exercise>>(new Map());

watch(exercises, (newExercises) => {
  // When exercises load, sync cache for any that are selected
  newExercises.forEach((ex) => {
    if (selectedIds.value.has(ex.id)) {
      selectedExerciseCache.value.set(ex.id, ex);
    }
  });
});

watch(selectedIds, (ids) => {
  // Remove deselected from cache
  selectedExerciseCache.value.forEach((_, id) => {
    if (!ids.has(id)) selectedExerciseCache.value.delete(id);
  });
  // Add newly selected from current page
  exercises.value.forEach((ex) => {
    if (ids.has(ex.id)) selectedExerciseCache.value.set(ex.id, ex);
  });
}, { deep: true });

const allSelected = computed(() => selectedExerciseCache.value);

function close() {
  emit("update:open", false);
}

// Reset when sheet closes
watch(() => props.open, (open) => {
  if (!open) {
    selectedIds.value = new Set();
    selectedExerciseCache.value = new Map();
    offset.value = 0;
    search.value = "";
  }
});
</script>

<template>
  <Sheet :open="open" @update:open="(v) => emit('update:open', v)">
    <SheetContent side="right" class="w-full sm:max-w-md flex flex-col gap-0 p-0" @open-auto-focus.prevent>
      <SheetHeader class="px-4 py-4 border-b shrink-0">
        <SheetTitle>{{ $t("workouts.addExercise") }}</SheetTitle>
        <SheetDescription v-if="selectedCount > 0">
          {{ t("workouts.exercisesSelected", { count: selectedCount }) }}
        </SheetDescription>
        <SheetDescription v-else>
          {{ $t("exercises.subtitle") }}
        </SheetDescription>
      </SheetHeader>

      <div class="px-4 py-3 border-b shrink-0">
        <Input
          :model-value="search"
          :placeholder="$t('exercises.filter.searchPlaceholder')"
          @update:model-value="handleSearchInput"
        />
      </div>

      <div class="flex-1 overflow-y-auto min-h-0">
        <ExerciseList
          :exercises="exercises"
          :loading="loading"
          :selectable="true"
          :selected-ids="selectedIds"
@select="handleSelect"
        />
      </div>

      <SheetFooter class="flex-col gap-2 px-4 py-4 border-t shrink-0">
        <div v-if="total > LIMIT" class="flex items-center justify-between text-sm text-muted-foreground">
          <Button
            variant="ghost"
            size="icon"
            :disabled="offset === 0"
            @click="offset = Math.max(0, offset - LIMIT)"
          >
            <ChevronLeft class="w-4 h-4" />
          </Button>
          <span>{{ currentPage }} / {{ totalPages }}</span>
          <Button
            variant="ghost"
            size="icon"
            :disabled="offset + LIMIT >= (total ?? 0)"
            @click="offset += LIMIT"
          >
            <ChevronRight class="w-4 h-4" />
          </Button>
        </div>
        <div class="flex gap-2">
          <Button variant="outline" class="flex-1" @click="close">
            {{ $t("cancel") }}
          </Button>
          <Button class="flex-1" :disabled="selectedCount === 0" @click="handleAdd">
            {{ selectedCount > 0
              ? t("workouts.addExercisesCount", { count: selectedCount })
              : $t("workouts.addExercise") }}
          </Button>
        </div>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
