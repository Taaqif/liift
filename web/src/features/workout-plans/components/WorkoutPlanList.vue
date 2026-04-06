<script setup lang="ts">
import type { WorkoutPlan } from "@/features/workout-plans/types";
import Card from "@/components/ui/card/Card.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { CalendarDays, Play } from "lucide-vue-next";

const props = defineProps<{
  plans: WorkoutPlan[];
  loading?: boolean;
  activePlanId?: number | null;
  isStarting?: boolean;
}>();

const emits = defineEmits<{
  (e: "edit", plan: WorkoutPlan): void;
  (e: "start", plan: WorkoutPlan): void;
  (e: "view", plan: WorkoutPlan): void;
}>();

function totalWorkoutSlots(plan: WorkoutPlan): number {
  return plan.weeks.reduce(
    (acc, w) => acc + w.days.filter((d) => !d.isRest && d.workoutIds.length > 0).length,
    0,
  );
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card v-for="i in 4" :key="i" class="gap-2">
        <CardContent>
          <div class="flex gap-4 items-center">
            <div class="shrink-0 w-16 h-16 rounded-lg bg-muted animate-pulse" />
            <div class="flex-1 space-y-2">
              <div class="h-5 w-36 bg-muted animate-pulse rounded" />
              <div class="h-4 w-24 bg-muted animate-pulse rounded" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <div v-else-if="plans.length === 0" class="text-center py-12">
      <p class="text-muted-foreground">{{ $t("workoutPlans.noPlans") }}</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card
        v-for="plan in plans"
        :key="plan.id"
        class="gap-2"
        :class="{ 'ring-2 ring-green-500/50': plan.id === activePlanId }"
      >
        <CardContent>
          <div class="flex items-start justify-between gap-4">
            <div class="flex gap-4 flex-1 items-center">
              <div class="shrink-0 w-16 h-16 rounded-lg border overflow-hidden bg-muted flex items-center justify-center">
                <CalendarDays class="w-8 h-8 text-muted-foreground" />
              </div>
              <div class="flex-1 flex flex-col gap-1">
                <div class="flex items-center gap-2 flex-wrap">
                  <CardTitle class="line-clamp-1">{{ plan.name }}</CardTitle>
                  <span
                    v-if="plan.id === activePlanId"
                    class="text-xs px-2 py-0.5 rounded-full bg-green-500/20 text-green-700 dark:text-green-400 border border-green-500/30 shrink-0"
                  >
                    {{ $t("workoutPlans.progress.activeBadge") }}
                  </span>
                </div>
                <CardDescription v-if="plan.description" class="line-clamp-2">
                  {{ plan.description }}
                </CardDescription>
                <p class="text-sm text-muted-foreground">
                  {{ $t("workoutPlans.weeksDays", { weeks: plan.numberOfWeeks, days: plan.daysPerWeek }) }}
                  · {{ $t("workoutPlans.workoutSlots", { count: totalWorkoutSlots(plan) }) }}
                </p>
              </div>
            </div>
            <div class="flex flex-col items-end gap-1 shrink-0">
              <div class="flex items-center gap-1">
                <Button variant="ghost" size="sm" @click="emits('view', plan)">
                  {{ $t("workoutPlans.detail.view") }}
                </Button>
                <Button variant="outline" size="sm" @click="emits('edit', plan)">
                  {{ $t("edit") }}
                </Button>
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
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
