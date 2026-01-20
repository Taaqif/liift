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
const isDrawerOpen = ref(false);
const isEditDrawerOpen = ref(false);
const selectedExercise = ref<Exercise | null>(null);
const isEditFormDirty = ref(false);
const isCreateFormDirty = ref(false);
const showUnsavedChangesDialog = ref(false);
const pendingDrawerClose = ref(false);
const isCreateDrawerPendingClose = ref(false);

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
  isEditFormDirty.value = false;
  isEditDrawerOpen.value = true;
};

const handleExerciseCreated = () => {
  isDrawerOpen.value = false;
  isCreateFormDirty.value = false;
};

const handleExerciseUpdated = () => {
  isEditDrawerOpen.value = false;
  selectedExercise.value = null;
  isEditFormDirty.value = false;
};

const handleExerciseDeleted = async () => {
  // Close drawer immediately
  isEditDrawerOpen.value = false;
  selectedExercise.value = null;
  isEditFormDirty.value = false;
  try {
    await refetch();
  } catch (err) {
    console.error("Failed to refetch exercises:", err);
  }
};

const handleFormDirty = (dirty: boolean) => {
  if (!isEditDrawerOpen.value) return;
  isEditFormDirty.value = dirty;
};

const handleEditDrawerOpenChange = (open: boolean) => {
  if (open) {
    isEditFormDirty.value = false;
    isEditDrawerOpen.value = true;
    return;
  }

  if (isEditFormDirty.value) {
    pendingDrawerClose.value = true;
    showUnsavedChangesDialog.value = true;
    isEditDrawerOpen.value = true;
    return;
  }

  isEditDrawerOpen.value = false;
  selectedExercise.value = null;
  isEditFormDirty.value = false;
};

const handleKeepEditing = () => {
  showUnsavedChangesDialog.value = false;
  pendingDrawerClose.value = false;
  isCreateDrawerPendingClose.value = false;
};

const handleCreateFormDirty = (dirty: boolean) => {
  if (!isDrawerOpen.value) return;
  isCreateFormDirty.value = dirty;
};

const handleCreateDrawerOpenChange = (open: boolean) => {
  if (open) {
    isCreateFormDirty.value = false;
    isDrawerOpen.value = true;
    return;
  }

  if (isCreateFormDirty.value) {
    isCreateDrawerPendingClose.value = true;
    showUnsavedChangesDialog.value = true;
    isDrawerOpen.value = true;
    return;
  }

  isDrawerOpen.value = false;
  isCreateFormDirty.value = false;
};

const handleDiscardChanges = () => {
  showUnsavedChangesDialog.value = false;
  if (pendingDrawerClose.value) {
    isEditDrawerOpen.value = false;
    selectedExercise.value = null;
    isEditFormDirty.value = false;
    pendingDrawerClose.value = false;
  } else if (isCreateDrawerPendingClose.value) {
    isDrawerOpen.value = false;
    isCreateFormDirty.value = false;
    isCreateDrawerPendingClose.value = false;
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
      <Drawer
        :open="isDrawerOpen"
        :dismissible="true"
        @update:open="handleCreateDrawerOpenChange"
      >
        <DrawerTrigger as-child>
          <Button>{{ $t("exercises.createNew") }}</Button>
        </DrawerTrigger>
        <CreateExerciseDrawer
          :open="isDrawerOpen"
          @exercise-created="handleExerciseCreated"
          @form-dirty="handleCreateFormDirty"
        />
      </Drawer>
      <Drawer
        :open="isEditDrawerOpen"
        :dismissible="true"
        @update:open="handleEditDrawerOpenChange"
      >
        <EditExerciseDrawer
          :open="isEditDrawerOpen"
          :exercise="selectedExercise"
          @exercise-updated="handleExerciseUpdated"
          @exercise-deleted="handleExerciseDeleted"
          @form-dirty="handleFormDirty"
        />
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
