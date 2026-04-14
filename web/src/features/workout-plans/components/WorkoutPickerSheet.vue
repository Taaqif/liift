<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";
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
import { Dumbbell, Check, ChevronLeft, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  add: [workoutIds: number[]];
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

const { workouts, total, loading } = useWorkouts(params);

const selectedIds = ref<Set<number>>(new Set());
const selectedCount = computed(() => selectedIds.value.size);

const totalPages = computed(() => Math.ceil((total.value ?? 0) / LIMIT));
const currentPage = computed(() => Math.floor(offset.value / LIMIT) + 1);

const filteredWorkouts = computed(() => workouts.value);

function handleSearchInput(value: string) {
  search.value = value;
  offset.value = 0;
}

function toggleSelect(id: number) {
  const next = new Set(selectedIds.value);
  if (next.has(id)) {
    next.delete(id);
  } else {
    next.add(id);
  }
  selectedIds.value = next;
}

function handleAdd() {
  emit("add", [...selectedIds.value]);
  close();
}

function close() {
  emit("update:open", false);
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      selectedIds.value = new Set();
      offset.value = 0;
      search.value = "";
    }
  },
);
</script>

<template>
  <Sheet :open="open" @update:open="(v) => emit('update:open', v)">
    <SheetContent side="right" class="w-full sm:max-w-md flex flex-col gap-0 p-0" @open-auto-focus.prevent>
      <SheetHeader class="px-4 py-4 border-b shrink-0">
        <SheetTitle>{{ $t("workoutPlans.selectWorkouts") }}</SheetTitle>
        <SheetDescription v-if="selectedCount > 0">
          {{ t("workouts.exercisesSelected", { count: selectedCount }) }}
        </SheetDescription>
        <SheetDescription v-else>
          {{ $t("workouts.subtitle") }}
        </SheetDescription>
      </SheetHeader>

      <div class="px-4 py-3 border-b shrink-0">
        <Input
          :model-value="search"
          :placeholder="$t('workouts.filter.searchPlaceholder')"
          @update:model-value="handleSearchInput"
        />
      </div>

      <div class="flex-1 overflow-y-auto min-h-0">
        <div v-if="loading" class="divide-y">
          <div v-for="i in 6" :key="i" class="flex items-center gap-3 px-4 py-3">
            <div class="shrink-0 w-10 h-10 rounded-md bg-muted animate-pulse" />
            <div class="flex-1 space-y-1.5">
              <div class="h-4 w-40 bg-muted animate-pulse rounded" />
              <div class="h-3 w-24 bg-muted animate-pulse rounded" />
            </div>
          </div>
        </div>
        <div v-else-if="filteredWorkouts.length === 0" class="text-center py-12">
          <p class="text-muted-foreground text-sm">{{ $t("workouts.noWorkouts") }}</p>
        </div>
        <div v-else class="divide-y">
          <div
            v-for="workout in filteredWorkouts"
            :key="workout.id"
            class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-muted/50 transition-colors"
            @click="toggleSelect(workout.id)"
          >
            <div class="shrink-0 w-10 h-10 rounded-md border overflow-hidden bg-muted flex items-center justify-center">
              <Dumbbell class="w-5 h-5 text-muted-foreground" />
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">{{ workout.name }}</p>
              <p class="text-xs text-muted-foreground">
                {{ workout.exercises?.length ?? 0 }} {{ $t("workouts.exercises") }}
              </p>
            </div>
            <div
              class="shrink-0 w-6 h-6 rounded-full border-2 flex items-center justify-center transition-colors"
              :class="selectedIds.has(workout.id)
                ? 'bg-primary border-primary text-primary-foreground'
                : 'border-muted-foreground/40'"
            >
              <Check v-if="selectedIds.has(workout.id)" class="w-3.5 h-3.5" />
            </div>
          </div>
        </div>
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
              : $t("workoutPlans.addWorkout") }}
          </Button>
        </div>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
