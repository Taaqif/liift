<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkoutPlans } from "@/features/workout-plans/composables/useWorkoutPlans";
import WorkoutPlanList from "@/features/workout-plans/components/WorkoutPlanList.vue";
import WorkoutPlanDrawer from "@/features/workout-plans/components/WorkoutPlanDrawer.vue";
import type { WorkoutPlan } from "@/features/workout-plans/types";
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

const createDrawerOpen = ref(false);
const editDrawerOpen = ref(false);
const selectedPlan = ref<WorkoutPlan | null>(null);
const formDirty = ref(false);
const showUnsavedDialog = ref(false);
const pendingCreateClose = ref(false);
const pendingEditClose = ref(false);

const { plans, loading, error, refetch } = useWorkoutPlans();

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

const handleEditPlan = (plan: WorkoutPlan) => {
  selectedPlan.value = plan;
  editDrawerOpen.value = true;
  nextTick(() => {
    formDirty.value = false;
  });
};

const handlePlanCreated = () => {
  nextTick(() => {
    createDrawerOpen.value = false;
    formDirty.value = false;
    pendingCreateClose.value = false;
  });
};

const handlePlanUpdated = () => {
  nextTick(() => {
    editDrawerOpen.value = false;
    selectedPlan.value = null;
    formDirty.value = false;
    pendingEditClose.value = false;
  });
};

const handlePlanDeleted = async () => {
  nextTick(() => {
    editDrawerOpen.value = false;
    selectedPlan.value = null;
    formDirty.value = false;
    pendingEditClose.value = false;
  });
  try {
    await refetch();
  } catch (err) {
    console.error("Failed to refetch plans:", err);
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

  nextTick(() => {
    createDrawerOpen.value = false;
    formDirty.value = false;
    pendingCreateClose.value = false;
  });
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

  nextTick(() => {
    editDrawerOpen.value = false;
    selectedPlan.value = null;
    formDirty.value = false;
    pendingEditClose.value = false;
  });
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
  nextTick(() => {
    if (pendingCreateClose.value) {
      createDrawerOpen.value = false;
      formDirty.value = false;
      pendingCreateClose.value = false;
    } else if (pendingEditClose.value) {
      editDrawerOpen.value = false;
      selectedPlan.value = null;
      formDirty.value = false;
      pendingEditClose.value = false;
    }
  });
};
</script>

<template>
  <div class="p-8 max-w-[1200px] mx-auto">
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold mb-2">{{ $t("workoutPlans.title") }}</h1>
        <p class="text-muted-foreground">
          {{ $t("workoutPlans.subtitle") }}
        </p>
      </div>
      <Drawer :open="createDrawerOpen" :dismissible="true" :handle-only="true"
        @update:open="handleCreateDrawerOpenChange">
        <DrawerTrigger as-child>
          <Button>{{ $t("workoutPlans.createNew") }}</Button>
        </DrawerTrigger>
        <WorkoutPlanDrawer :open="createDrawerOpen" :plan="null" @plan-created="handlePlanCreated"
          @form-dirty="handleFormDirty" />
      </Drawer>
      <Drawer :open="editDrawerOpen" :dismissible="true" :handle-only="true" @update:open="handleEditDrawerOpenChange">
        <WorkoutPlanDrawer :open="editDrawerOpen" :plan="selectedPlan" @plan-updated="handlePlanUpdated"
          @plan-deleted="handlePlanDeleted" @form-dirty="handleFormDirty" />
      </Drawer>
      <Dialog v-model:open="showUnsavedDialog">
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{{ $t("workoutPlans.unsavedChangesTitle") }}</DialogTitle>
            <DialogDescription>
              {{ $t("workoutPlans.unsavedChangesDescription") }}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" @click="handleKeepEditing">
              {{ $t("workoutPlans.keepEditing") }}
            </Button>
            <Button variant="destructive" @click="handleDiscardChanges">
              {{ $t("workoutPlans.discardChanges") }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>{{ $t("workoutPlans.errorLoading", { message: error?.message }) }}</p>
    </div>

    <WorkoutPlanList :plans="plans" :loading="loading" @edit="handleEditPlan" />
  </div>
</template>
