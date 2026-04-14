<script setup lang="ts">
import { useRouter } from "vue-router";
import { useWorkoutPlans } from "@/features/workout-plans/composables/useWorkoutPlans";
import { useActivePlanProgress } from "@/features/workout-plans/composables/useActivePlanProgress";
import { useStartPlan } from "@/features/workout-plans/composables/useStartPlan";
import WorkoutPlanList from "@/features/workout-plans/components/WorkoutPlanList.vue";
import ActivePlanCard from "@/features/workout-plans/components/ActivePlanCard.vue";
import type { WorkoutPlan } from "@/features/workout-plans/types";
import { Button } from "@/components/ui/button";
import { toast } from "vue-sonner";
import { useI18n } from "vue-i18n";
import { Plus } from "lucide-vue-next";

const router = useRouter();
const { t } = useI18n();

const { plans, loading, error } = useWorkoutPlans();
const { progress } = useActivePlanProgress();
const { startPlan, isPending: isStarting } = useStartPlan();

const handleViewPlan = (plan: WorkoutPlan) => {
  router.push({ name: "workout-plan-detail", params: { id: plan.id } });
};

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
  <div class="pb-12 space-y-8">
    <!-- Header -->
    <div class="flex items-start justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">{{ $t("workoutPlans.title") }}</h1>
        <p class="text-muted-foreground mt-1">{{ $t("workoutPlans.subtitle") }}</p>
      </div>
      <Button class="gap-1.5 shrink-0" @click="router.push({ name: 'workout-plan-create' })">
        <Plus class="w-4 h-4" />
        {{ $t("workoutPlans.createNew") }}
      </Button>
    </div>

    <!-- Active plan banner -->
    <ActivePlanCard v-if="progress" :progress="progress" />

    <!-- Error -->
    <div v-if="error" class="p-4 rounded-xl bg-destructive/10 text-destructive text-sm">
      {{ $t("workoutPlans.errorLoading", { message: error?.message }) }}
    </div>

    <WorkoutPlanList
      :plans="plans"
      :loading="loading"
      :active-plan-id="progress?.plan_id ?? null"
      :is-starting="isStarting"
      @view="handleViewPlan"
      @edit="handleEditPlan"
      @start="handleStartPlan"
    />
  </div>
</template>
