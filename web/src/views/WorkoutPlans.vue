<script setup lang="ts">
import { useRouter } from "vue-router";
import { useWorkoutPlans } from "@/features/workout-plans/composables/useWorkoutPlans";
import { useActivePlanProgress } from "@/features/workout-plans/composables/useActivePlanProgress";
import { useStartPlan } from "@/features/workout-plans/composables/useStartPlan";
import WorkoutPlanList from "@/features/workout-plans/components/WorkoutPlanList.vue";
import type { WorkoutPlan } from "@/features/workout-plans/types";
import { Button } from "@/components/ui/button";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";

const router = useRouter();
const { t } = useI18n();

const { plans, loading, error } = useWorkoutPlans();
const { progress } = useActivePlanProgress();
const { startPlan, isPending: isStarting } = useStartPlan();

const handleEditPlan = (plan: WorkoutPlan) => {
  router.push({ name: "workout-plan-edit", params: { id: plan.id } });
};

async function handleStartPlan(plan: WorkoutPlan) {
  try {
    await startPlan(plan.id);
    router.push({ name: "active-plan" });
  } catch (err) {
    if (err instanceof Error && err.message === "active_plan_progress_exists") {
      toast.error(t("workoutPlans.progress.alreadyActive"));
    } else {
      toast.error(t("workoutPlans.error"));
    }
  }
}
</script>

<template>
  <div>
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold mb-2">{{ $t("workoutPlans.title") }}</h1>
        <p class="text-muted-foreground">
          {{ $t("workoutPlans.subtitle") }}
        </p>
      </div>
      <Button @click="router.push({ name: 'workout-plan-create' })">
        {{ $t("workoutPlans.createNew") }}
      </Button>
    </div>

    <!-- Active plan banner -->
    <div
      v-if="progress"
      class="mb-6 p-4 rounded-lg border border-green-500/30 bg-green-500/10 flex items-center justify-between gap-4"
    >
      <div>
        <p class="font-medium text-green-700 dark:text-green-400">
          {{ $t("workoutPlans.progress.activeBanner.title") }}
        </p>
        <p class="text-sm text-muted-foreground">
          {{ progress.plan.name }} —
          {{ $t("workoutPlans.progress.activeBanner.position", {
            week: progress.current_week + 1,
            day: progress.current_day + 1,
          }) }}
        </p>
      </div>
      <Button @click="router.push({ name: 'active-plan' })">
        {{ $t("workoutPlans.progress.activeBanner.view") }}
      </Button>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>{{ $t("workoutPlans.errorLoading", { message: error?.message }) }}</p>
    </div>

    <WorkoutPlanList
      :plans="plans"
      :loading="loading"
      :active-plan-id="progress?.plan_id ?? null"
      :is-starting="isStarting"
      @edit="handleEditPlan"
      @start="handleStartPlan"
    />
  </div>
</template>
