<script setup lang="ts">
import { ref, computed } from "vue";
import { useExercises } from "@/features/exercises/composables/useExercises";
import ExerciseList from "@/features/exercises/components/ExerciseList.vue";

const limit = ref(20);
const offset = ref(0);

const params = computed(() => ({
  limit: limit.value,
  offset: offset.value,
}));

const { exercises, total, loading, error } = useExercises(params);

const totalPages = computed(() => Math.ceil(total.value / limit.value));
const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1);

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  offset.value = (page - 1) * limit.value;
};

const hasNextPage = computed(() => offset.value + limit.value < total.value);
const hasPrevPage = computed(() => offset.value > 0);
</script>

<template>
  <div class="p-8 max-w-[1200px] mx-auto">
    <div class="mb-8">
      <h1 class="text-3xl font-bold mb-2">Exercises</h1>
      <p class="text-muted-foreground">
        Browse all available exercises in the library
      </p>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>Error loading exercises: {{ error.message }}</p>
    </div>

    <ExerciseList :exercises="exercises" :loading="loading" />

    <div
      v-if="!loading && total > 0"
      class="mt-8 flex items-center justify-between"
    >
      <div class="text-sm text-muted-foreground">
        Showing {{ offset + 1 }} to
        {{ Math.min(offset + limit, total) }} of {{ total }} exercises
      </div>
      <div class="flex gap-2">
        <button
          @click="goToPage(currentPage - 1)"
          :disabled="!hasPrevPage"
          class="px-4 py-2 border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
        >
          Previous
        </button>
        <button
          @click="goToPage(currentPage + 1)"
          :disabled="!hasNextPage"
          class="px-4 py-2 border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>