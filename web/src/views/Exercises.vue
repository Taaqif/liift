<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useExercises } from "@/features/exercises/composables/useExercises";
import ExerciseList from "@/features/exercises/components/ExerciseList.vue";
import ExerciseDrawer from "@/features/exercises/components/ExerciseDrawer.vue";
import ExerciseFilter from "@/features/exercises/components/ExerciseFilter.vue";
import type { Exercise } from "@/features/exercises/types";
import type { ExerciseFilter as ExerciseFilterType } from "@/features/exercises/components/ExerciseFilter.vue";
import { Button } from "@/components/ui/button";
import { Drawer, DrawerTrigger } from "@/components/ui/drawer";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";

const route = useRoute();
const router = useRouter();

const limit = ref(10);
const offset = ref(0);

const createDrawerOpen = ref(false);
const editDrawerOpen = ref(false);
const selectedExercise = ref<Exercise | null>(null);
const formDirty = ref(false);
const showUnsavedDialog = ref(false);
const pendingCreateClose = ref(false);
const pendingEditClose = ref(false);

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

const { exercises, total, loading, error, refetch } = useExercises(params);

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

onMounted(() => {
  if (route.query.action === "create") {
    createDrawerOpen.value = true;
    router.replace({ query: {} });
  }
});

watch(
  () => route.query.action,
  () => {
    if (route.query.action === "create") {
      createDrawerOpen.value = true;
      router.replace({ query: {} });
    }
  },
);
const handleEditExercise = (exercise: Exercise) => {
  selectedExercise.value = exercise;
  editDrawerOpen.value = true;
  nextTick(() => {
    formDirty.value = false;
  });
};

const handleExerciseCreated = () => {
  createDrawerOpen.value = false;
  formDirty.value = false;
  pendingCreateClose.value = false;
};

const handleExerciseUpdated = () => {
  editDrawerOpen.value = false;
  selectedExercise.value = null;
  formDirty.value = false;
  pendingEditClose.value = false;
};

const handleExerciseDeleted = async () => {
  editDrawerOpen.value = false;
  selectedExercise.value = null;
  formDirty.value = false;
  pendingEditClose.value = false;
  try {
    await refetch();
  } catch (err) {
    console.error("Failed to refetch exercises:", err);
  }
};

const handleCreateDrawerOpenChange = (open: boolean) => {
  if (open) {
    createDrawerOpen.value = true;
    pendingCreateClose.value = false;
    nextTick(() => {
      formDirty.value = false;
    });
    return;
  }

  if (formDirty.value) {
    pendingCreateClose.value = true;
    showUnsavedDialog.value = true;
    createDrawerOpen.value = true;
    return;
  }

  createDrawerOpen.value = false;
  formDirty.value = false;
  pendingCreateClose.value = false;
};

const handleEditDrawerOpenChange = (open: boolean) => {
  if (open) {
    editDrawerOpen.value = true;
    pendingEditClose.value = false;
    nextTick(() => {
      formDirty.value = false;
    });
    return;
  }

  if (formDirty.value) {
    pendingEditClose.value = true;
    showUnsavedDialog.value = true;
    editDrawerOpen.value = true;
    return;
  }

  editDrawerOpen.value = false;
  selectedExercise.value = null;
  formDirty.value = false;
  pendingEditClose.value = false;
};

const handleFormDirty = (dirty: boolean) => {
  if (!createDrawerOpen.value && !editDrawerOpen.value) {
    formDirty.value = false;
    return;
  }
  formDirty.value = dirty;
};

const handleKeepEditing = () => {
  showUnsavedDialog.value = false;
  pendingCreateClose.value = false;
  pendingEditClose.value = false;
};

const handleDiscardChanges = () => {
  showUnsavedDialog.value = false;
  if (pendingCreateClose.value) {
    createDrawerOpen.value = false;
    formDirty.value = false;
    pendingCreateClose.value = false;
  } else if (pendingEditClose.value) {
    editDrawerOpen.value = false;
    selectedExercise.value = null;
    formDirty.value = false;
    pendingEditClose.value = false;
  }
};

const handleFilter = (newFilter: ExerciseFilterType) => {
  // Update filter - this will trigger params computed to update
  filter.value = {
    search: newFilter.search,
    muscleGroup: [...newFilter.muscleGroup],
    equipment: [...newFilter.equipment],
  };
  // Reset to first page when filtering; params computed will trigger refetch
  offset.value = 0;
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
      <Drawer
        :open="createDrawerOpen"
        :dismissible="true"
        @update:open="handleCreateDrawerOpenChange"
      >
        <DrawerTrigger as-child>
          <Button>{{ $t("exercises.createNew") }}</Button>
        </DrawerTrigger>
        <ExerciseDrawer
          :open="createDrawerOpen"
          :exercise="null"
          @exercise-created="handleExerciseCreated"
          @form-dirty="handleFormDirty"
        />
      </Drawer>
      <Drawer
        :open="editDrawerOpen"
        :dismissible="true"
        @update:open="handleEditDrawerOpenChange"
      >
        <ExerciseDrawer
          :open="editDrawerOpen"
          :exercise="selectedExercise"
          @exercise-updated="handleExerciseUpdated"
          @exercise-deleted="handleExerciseDeleted"
          @form-dirty="handleFormDirty"
        />
      </Drawer>
      <Dialog v-model:open="showUnsavedDialog">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Unsaved changes</DialogTitle>
            <DialogDescription>
              You have unsaved changes. Are you sure you want to close?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" @click="handleKeepEditing">
              Keep editing
            </Button>
            <Button variant="destructive" @click="handleDiscardChanges">
              Discard changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>

    <div
      v-if="error"
      class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg"
    >
      <p>{{ $t("exercises.errorLoading", { message: error.message }) }}</p>
    </div>

    <div class="mb-6">
      <ExerciseFilter
        :model-value="filter"
        @update:model-value="handleFilter"
      />
    </div>

    <ExerciseList
      :exercises="exercises"
      :loading="loading"
      @edit="handleEditExercise"
    />

    <div
      v-if="!loading && total > 0"
      class="mt-8 flex flex-col gap-2 items-center justify-between"
    >
      <Pagination
        v-slot="{ page }"
        v-model:page="currentPage"
        :items-per-page="limit"
        :total="total"
        :default-page="currentPage"
      >
        <PaginationContent v-slot="{ items }">
          <PaginationPrevious />

          <template v-for="(item, index) in items" :key="index">
            <PaginationItem
              v-if="item.type === 'page'"
              :value="item.value"
              :is-active="item.value === page"
            >
              {{ item.value }}
            </PaginationItem>
            <PaginationEllipsis v-else :key="item.type" :index="index" />
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
</template>
