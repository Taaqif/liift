<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useExercises } from "@/features/exercises/composables/useExercises";
import ExerciseList from "@/features/exercises/components/ExerciseList.vue";
import CreateExerciseDrawer from "@/features/exercises/components/CreateExerciseDrawer.vue";
import EditExerciseDrawer from "@/features/exercises/components/EditExerciseDrawer.vue";
import type { Exercise } from "@/features/exercises/types";
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

const route = useRoute();
const router = useRouter();

const limit = ref(20);
const offset = ref(0);

const isCreateDrawerOpen = ref(false);
const isEditDrawerOpen = ref(false);
const selectedExercise = ref<Exercise | null>(null);
const isFormDirty = ref(false);
const showUnsavedChangesDialog = ref(false);
const pendingCreateDrawerClose = ref(false);
const pendingEditDrawerClose = ref(false);

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

const checkQueryParam = () => {
  if (route.query.action === "create") {
    isCreateDrawerOpen.value = true;
    router.replace({ query: {} });
  }
};

onMounted(() => {
  checkQueryParam();
});

watch(
  () => route.query.action,
  () => {
    checkQueryParam();
  },
);
const handleEditExercise = (exercise: Exercise) => {
  selectedExercise.value = exercise;
  isEditDrawerOpen.value = true;
  nextTick(() => {
    isFormDirty.value = false;
  });
};

const handleExerciseCreated = () => {
  resetCreateDrawer();
};

const handleExerciseUpdated = () => {
  resetEditDrawer();
};

const handleExerciseDeleted = async () => {
  resetEditDrawer();
  try {
    await refetch();
  } catch (err) {
    console.error("Failed to refetch exercises:", err);
  }
};

const handleCreateDrawerOpenChange = (open: boolean) => {
  if (open) {
    isCreateDrawerOpen.value = true;
    pendingCreateDrawerClose.value = false;
    nextTick(() => {
      isFormDirty.value = false;
    });
    return;
  }

  if (isFormDirty.value) {
    pendingCreateDrawerClose.value = true;
    showUnsavedChangesDialog.value = true;
    isCreateDrawerOpen.value = true;
    return;
  }

  resetCreateDrawer();
};

const handleEditDrawerOpenChange = (open: boolean) => {
  if (open) {
    isEditDrawerOpen.value = true;
    pendingEditDrawerClose.value = false;
    nextTick(() => {
      isFormDirty.value = false;
    });
    return;
  }

  if (isFormDirty.value) {
    pendingEditDrawerClose.value = true;
    showUnsavedChangesDialog.value = true;
    isEditDrawerOpen.value = true;
    return;
  }

  resetEditDrawer();
};

const handleFormDirty = (dirty: boolean) => {
  if (!isCreateDrawerOpen.value && !isEditDrawerOpen.value) {
    isFormDirty.value = false;
    return;
  }
  isFormDirty.value = dirty;
};

const resetCreateDrawer = () => {
  isCreateDrawerOpen.value = false;
  isFormDirty.value = false;
  pendingCreateDrawerClose.value = false;
};

const resetEditDrawer = () => {
  isEditDrawerOpen.value = false;
  selectedExercise.value = null;
  isFormDirty.value = false;
  pendingEditDrawerClose.value = false;
};

const handleKeepEditing = () => {
  showUnsavedChangesDialog.value = false;
  pendingCreateDrawerClose.value = false;
  pendingEditDrawerClose.value = false;
};

const handleDiscardChanges = () => {
  showUnsavedChangesDialog.value = false;
  if (pendingCreateDrawerClose.value) {
    resetCreateDrawer();
  } else if (pendingEditDrawerClose.value) {
    resetEditDrawer();
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
      <Drawer :open="isCreateDrawerOpen" :dismissible="true" @update:open="handleCreateDrawerOpenChange">
        <DrawerTrigger as-child>
          <Button>{{ $t("exercises.createNew") }}</Button>
        </DrawerTrigger>
        <CreateExerciseDrawer :open="isCreateDrawerOpen" @exercise-created="handleExerciseCreated"
          @form-dirty="handleFormDirty" />
      </Drawer>
      <Drawer :open="isEditDrawerOpen" :dismissible="true" @update:open="handleEditDrawerOpenChange">
        <EditExerciseDrawer :open="isEditDrawerOpen" :exercise="selectedExercise"
          @exercise-updated="handleExerciseUpdated" @exercise-deleted="handleExerciseDeleted"
          @form-dirty="handleFormDirty" />
      </Drawer>
      <Dialog v-model:open="showUnsavedChangesDialog">
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

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>{{ $t("exercises.errorLoading", { message: error.message }) }}</p>
    </div>

    <ExerciseList :exercises="exercises" :loading="loading" @edit="handleEditExercise" />

    <div v-if="!loading && total > 0" class="mt-8 flex items-center justify-between">
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
        <button @click="goToPage(currentPage - 1)" :disabled="!hasPrevPage"
          class="px-4 py-2 border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent">
          {{ $t("pagination.previous") }}
        </button>
        <button @click="goToPage(currentPage + 1)" :disabled="!hasNextPage"
          class="px-4 py-2 border rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-accent">
          {{ $t("pagination.next") }}
        </button>
      </div>
    </div>
  </div>
</template>
