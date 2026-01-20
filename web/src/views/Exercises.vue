<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useExercises } from "@/features/exercises/composables/useExercises";
import ExerciseList from "@/features/exercises/components/ExerciseList.vue";
import CreateExerciseDrawer from "@/features/exercises/components/CreateExerciseDrawer.vue";
import EditExerciseDrawer from "@/features/exercises/components/EditExerciseDrawer.vue";
import type { Exercise } from "@/features/exercises/types";
import { Button } from "@/components/ui/button";
import { Drawer, DrawerTrigger } from "@/components/ui/drawer";

const route = useRoute();
const router = useRouter();

const limit = ref(20);
const offset = ref(0);
const isDrawerOpen = ref(false);
const isEditDrawerOpen = ref(false);
const selectedExercise = ref<Exercise | null>(null);

const params = computed(() => ({
  limit: limit.value,
  offset: offset.value,
}));

const { exercises, total, loading, error, refetch } = useExercises(params);

const totalPages = computed(() => Math.ceil(total.value / limit.value));
const currentPage = computed(() => Math.floor(offset.value / limit.value) + 1);

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  offset.value = (page - 1) * limit.value;
};

const hasNextPage = computed(() => offset.value + limit.value < total.value);
const hasPrevPage = computed(() => offset.value > 0);

// Handle query parameter to open drawer
const checkQueryParam = () => {
  if (route.query.action === "create") {
    isDrawerOpen.value = true;
    // Clean up the query parameter
    router.replace({ query: {} });
  }
};

onMounted(() => {
  checkQueryParam();
});

// Watch for route changes to handle query parameter
watch(
  () => route.query.action,
  () => {
    checkQueryParam();
  },
);

const handleEditExercise = (exercise: Exercise) => {
  selectedExercise.value = exercise;
  isEditDrawerOpen.value = true;
};

const handleExerciseUpdated = () => {
  isEditDrawerOpen.value = false;
  selectedExercise.value = null;
};

const handleExerciseDeleted = async () => {
  // Close drawer immediately
  isEditDrawerOpen.value = false;
  selectedExercise.value = null;
  // Explicitly refetch to ensure data is updated
  try {
    await refetch();
  } catch (err) {
    console.error("Failed to refetch exercises:", err);
  }
};
</script>

<template>
  <div class="p-8 max-w-[1200px] mx-auto">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold mb-2">{{ $t("exercises.title") }}</h1>
        <p class="text-muted-foreground">
          {{ $t("exercises.subtitle") }}
        </p>
      </div>
      <Drawer v-model:open="isDrawerOpen" :dismissible="false">
        <DrawerTrigger as-child>
          <Button>{{ $t("exercises.createNew") }}</Button>
        </DrawerTrigger>
        <CreateExerciseDrawer
          :open="isDrawerOpen"
          @exercise-created="isDrawerOpen = false"
        />
      </Drawer>
      <Drawer v-model:open="isEditDrawerOpen" :dismissible="false">
        <EditExerciseDrawer
          :open="isEditDrawerOpen"
          :exercise="selectedExercise"
          @exercise-updated="handleExerciseUpdated"
          @exercise-deleted="handleExerciseDeleted"
        />
      </Drawer>
    </div>

    <div
      v-if="error"
      class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg"
    >
      <p>{{ $t("exercises.errorLoading", { message: error.message }) }}</p>
    </div>

    <ExerciseList
      :exercises="exercises"
      :loading="loading"
      @edit="handleEditExercise"
    />

    <div
      v-if="!loading && total > 0"
      class="mt-8 flex items-center justify-between"
    >
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
      <div class="flex gap-2">
        <button
          @click="goToPage(currentPage - 1)"
          :disabled="!hasPrevPage"
          class="px-4 py-2 border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
        >
          {{ $t("pagination.previous") }}
        </button>
        <button
          @click="goToPage(currentPage + 1)"
          :disabled="!hasNextPage"
          class="px-4 py-2 border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent"
        >
          {{ $t("pagination.next") }}
        </button>
      </div>
    </div>
  </div>
</template>
