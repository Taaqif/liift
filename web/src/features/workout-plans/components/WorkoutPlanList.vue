<script setup lang="ts">
import type { WorkoutPlan } from "@/features/workout-plans/types";
import Card from "@/components/ui/card/Card.vue";
import CardHeader from "@/components/ui/card/CardHeader.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Calendar, Play } from "lucide-vue-next";

const props = defineProps<{
  plans: WorkoutPlan[];
  loading?: boolean;
  activePlanId?: number | null;
  isStarting?: boolean;
}>();

const emits = defineEmits<{
  (e: "edit", plan: WorkoutPlan): void;
  (e: "start", plan: WorkoutPlan): void;
}>();

function totalWorkoutSlots(plan: WorkoutPlan): number {
  return plan.weeks.reduce(
    (acc, w) =>
      acc +
      w.days.filter((d) => !d.isRest && d.workoutIds.length > 0).length,
    0,
  );
}
</script>

<template>
  <div class="space-y-4">
    <div
      v-if="loading"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
    >
      <Card v-for="i in 6" :key="i">
        <CardHeader>
          <CardTitle>
            <div class="h-6 w-48 bg-gray-200 animate-pulse rounded"></div>
          </CardTitle>
          <CardDescription>
            <div
              class="h-4 w-full bg-gray-200 animate-pulse rounded mt-2"
            ></div>
          </CardDescription>
        </CardHeader>
      </Card>
    </div>

    <div
      v-else-if="plans.length === 0"
      class="text-center py-12"
    >
      <p class="text-muted-foreground">{{ $t("workoutPlans.noPlans") }}</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card
        v-for="plan in plans"
        :key="plan.id"
        class="hover:shadow-md transition-shadow"
        :class="{ 'ring-2 ring-green-500/50': plan.id === activePlanId }"
      >
        <CardHeader>
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <CardTitle class="line-clamp-2">{{ plan.name }}</CardTitle>
                <span
                  v-if="plan.id === activePlanId"
                  class="text-xs px-2 py-0.5 rounded-full bg-green-500/20 text-green-700 dark:text-green-400 border border-green-500/30 shrink-0"
                >
                  {{ $t("workoutPlans.progress.activeBadge") }}
                </span>
              </div>
              <CardDescription
                v-if="plan.description"
                class="line-clamp-2 mt-2"
              >
                {{ plan.description }}
              </CardDescription>
            </div>
            <Button
              variant="outline"
              size="sm"
              @click="emits('edit', plan)"
              class="shrink-0"
            >
              {{ $t("edit") }}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-4 text-sm text-muted-foreground">
              <div class="flex items-center gap-1">
                <Calendar class="w-4 h-4" />
                <span>
                  {{ $t("workoutPlans.weeksDays", {
                    weeks: plan.numberOfWeeks,
                    days: plan.daysPerWeek,
                  }) }}
                </span>
              </div>
              <div class="flex items-center gap-1">
                <span>
                  {{ $t("workoutPlans.workoutSlots", {
                    count: totalWorkoutSlots(plan),
                  }) }}
                </span>
              </div>
            </div>
            <Button
              v-if="plan.id !== activePlanId"
              size="sm"
              variant="secondary"
              :disabled="isStarting || !!activePlanId"
              @click="emits('start', plan)"
            >
              <Play class="w-3 h-3 mr-1" />
              {{ $t("workoutPlans.progress.start") }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
