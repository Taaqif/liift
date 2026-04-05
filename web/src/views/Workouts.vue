<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useRouter } from "vue-router";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";
import { useStartWorkout } from "@/features/workout-session/composables/useStartWorkout";
import { useI18n } from "vue-i18n";
import { toast } from "vue-sonner";
import WorkoutList from "@/features/workouts/components/WorkoutList.vue";
import WorkoutFilter from "@/features/workouts/components/WorkoutFilter.vue";
import type { Workout } from "@/features/workouts/types";
import type { WorkoutFilter as WorkoutFilterType } from "@/features/workouts/components/WorkoutFilter.vue";
import { Button } from "@/components/ui/button";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";

const router = useRouter();
const { t } = useI18n();
const { startWorkout } = useStartWorkout();
const startingWorkoutId = ref<number | null>(null);

const limit = ref(12);
const offset = ref(0);

const filter = ref<WorkoutFilterType>({
  search: "",
  exerciseFeatures: [],
  exerciseIds: [],
  muscleGroup: [],
  equipment: [],
});

const params = computed(() => ({
  limit: limit.value,
  offset: offset.value,
  search: filter.value.search || undefined,
  exerciseFeatures: filter.value.exerciseFeatures.length ? filter.value.exerciseFeatures : undefined,
  exerciseIds: filter.value.exerciseIds.length ? filter.value.exerciseIds : undefined,
  muscleGroup: filter.value.muscleGroup.length ? filter.value.muscleGroup : undefined,
  equipment: filter.value.equipment.length ? filter.value.equipment : undefined,
}));

const { workouts, total, loading, error } = useWorkouts(params);

const currentPage = ref(1);

watch(
  () => offset.value,
  () => {
    currentPage.value = Math.floor(offset.value / limit.value) + 1;
  },
  { immediate: true },
);

watch(currentPage, (page) => {
  offset.value = (page - 1) * limit.value;
});

const handleEditWorkout = (workout: Workout) => {
  router.push({ name: "workout-edit", params: { id: workout.id } });
};

const handleStartWorkout = async (workout: Workout) => {
  startingWorkoutId.value = workout.id;
  try {
    await startWorkout(workout.id);
  } catch (e) {
    const err = e as Error;
    if (err.message === "active_session_exists") {
      toast.error(t("workoutSession.activeSessionExists"));
      router.push({ name: "active-workout" });
    } else {
      toast.error(err.message);
    }
  } finally {
    startingWorkoutId.value = null;
  }
};

const handleFilter = (newFilter: WorkoutFilterType) => {
  filter.value = {
    search: newFilter.search,
    exerciseFeatures: [...newFilter.exerciseFeatures],
    exerciseIds: [...newFilter.exerciseIds],
    muscleGroup: [...newFilter.muscleGroup],
    equipment: [...newFilter.equipment],
  };
  offset.value = 0;
};
</script>

<template>
  <div class="p-8 max-w-[1200px] mx-auto">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold mb-2">{{ $t("workouts.title") }}</h1>
        <p class="text-muted-foreground">
          {{ $t("workouts.subtitle") }}
        </p>
      </div>
      <Button @click="router.push({ name: 'workout-create' })">
        {{ $t("workouts.createNew") }}
      </Button>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>{{ $t("workouts.errorLoading", { message: error.message }) }}</p>
    </div>

    <div class="mb-6">
      <WorkoutFilter :model-value="filter" @update:model-value="handleFilter" />
    </div>

    <WorkoutList
      :workouts="workouts"
      :loading="loading"
      :starting-workout-id="startingWorkoutId"
      @edit="handleEditWorkout"
      @start="handleStartWorkout"
    />

    <div v-if="!loading && total > 0" class="mt-8 flex flex-col gap-2 items-center justify-between">
      <Pagination v-slot="{ page }" v-model:page="currentPage" :items-per-page="limit" :total="total"
        :default-page="currentPage">
        <PaginationContent v-slot="{ items }">
          <PaginationPrevious />

          <template v-for="(item, index) in items" :key="index">
            <PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === page">
              {{ item.value }}
            </PaginationItem>
            <PaginationEllipsis v-else :key="item.type" />
          </template>

          <PaginationNext />
        </PaginationContent>
      </Pagination>
      <div class="text-sm text-muted-foreground">
        {{
          $t("pagination.showingFromToOfTotal", {
            from: offset + 1,
            to: Math.min(offset + limit, total),
            total: total,
          })
        }}
        {{ $t("workouts.titleLower") }}
      </div>
    </div>
  </div>
</template>
