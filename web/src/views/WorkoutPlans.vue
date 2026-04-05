<script setup lang="ts">
import { useRouter } from "vue-router";
import { useWorkoutPlans } from "@/features/workout-plans/composables/useWorkoutPlans";
import WorkoutPlanList from "@/features/workout-plans/components/WorkoutPlanList.vue";
import type { WorkoutPlan } from "@/features/workout-plans/types";
import { Button } from "@/components/ui/button";

const router = useRouter();

const { plans, loading, error } = useWorkoutPlans();

const handleEditPlan = (plan: WorkoutPlan) => {
  router.push({ name: "workout-plan-edit", params: { id: plan.id } });
};
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

    <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
      <p>{{ $t("workoutPlans.errorLoading", { message: error?.message }) }}</p>
    </div>

    <WorkoutPlanList :plans="plans" :loading="loading" @edit="handleEditPlan" />
  </div>
</template>
