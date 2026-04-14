<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useRouter } from "vue-router";
import { useExercises } from "@/features/exercises/composables/useExercises";
import ExerciseList from "@/features/exercises/components/ExerciseList.vue";
import ExerciseFilter from "@/features/exercises/components/ExerciseFilter.vue";
import type { Exercise } from "@/features/exercises/types";
import type { ExerciseFilter as ExerciseFilterType } from "@/features/exercises/components/ExerciseFilter.vue";
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

const limit = ref(10);
const offset = ref(0);

const filter = ref<ExerciseFilterType>({
  search: "",
  muscleGroup: [],
  equipment: [],
});

const params = computed(() => {
  const search = filter.value.search;
  const muscleGroups = filter.value.muscleGroup;
  const equipment = filter.value.equipment;

  return {
    limit: limit.value,
    offset: offset.value,
    search: search || undefined,
    muscleGroup: muscleGroups.length > 0 ? muscleGroups : undefined,
    equipment: equipment.length > 0 ? equipment : undefined,
  };
});

const { exercises, total, loading, error } = useExercises(params);

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

const handleEditExercise = (exercise: Exercise) => {
  router.push({ name: "exercise-edit", params: { id: exercise.id } });
};

const handleFilter = (newFilter: ExerciseFilterType) => {
  filter.value = {
    search: newFilter.search,
    muscleGroup: [...newFilter.muscleGroup],
    equipment: [...newFilter.equipment],
  };
  offset.value = 0;
};
</script>

<template>
  <div>
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold mb-2">{{ $t("exercises.title") }}</h1>
        <p class="text-muted-foreground">
          {{ $t("exercises.subtitle") }}
        </p>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" @click="router.push({ name: 'exercise-import' })">Import</Button>
        <Button @click="router.push({ name: 'exercise-create' })">
          {{ $t("exercises.createNew") }}
        </Button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>{{ $t("exercises.errorLoading", { message: error.message }) }}</p>
    </div>

    <div class="flex flex-col lg:flex-row gap-6 lg:gap-8">
      <!-- Filter: top on mobile, left sidebar on desktop -->
      <div class="lg:w-72 lg:shrink-0">
        <ExerciseFilter :model-value="filter" @update:model-value="handleFilter" />
      </div>

      <!-- List + pagination -->
      <div class="flex-1 min-w-0">
        <ExerciseList :exercises="exercises" :loading="loading" @edit="handleEditExercise" />

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
            {{ $t("exercises.titleLower") }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
