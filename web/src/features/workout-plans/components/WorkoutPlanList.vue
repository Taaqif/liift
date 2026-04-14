<script setup lang="ts">
import type { WorkoutPlan } from "@/features/workout-plans/types";
import { Button } from "@/components/ui/button";
import { Target, Play, Pencil, Eye, CheckCircle2 } from "lucide-vue-next";

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
    (acc, w) => acc + w.days.filter((d) => d.workoutIds.length > 0).length,
    0,
  );
}
</script>

<template>
  <!-- Loading -->
  <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
    <div v-for="i in 4" :key="i" class="rounded-2xl border bg-card p-5 space-y-3">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-muted animate-pulse shrink-0" />
        <div class="space-y-2 flex-1">
          <div class="h-4 w-32 bg-muted animate-pulse rounded" />
          <div class="h-3 w-20 bg-muted animate-pulse rounded" />
        </div>
      </div>
      <div class="h-px bg-muted" />
      <div class="h-8 bg-muted animate-pulse rounded-lg" />
    </div>
  </div>

  <!-- Empty -->
  <div v-else-if="plans.length === 0" class="flex flex-col items-center gap-3 py-20 text-center text-muted-foreground/50">
    <CalendarDays class="w-8 h-8" />
    <p class="text-sm">{{ $t("workoutPlans.noPlans") }}</p>
  </div>

  <!-- Plan grid -->
  <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
    <div
      v-for="plan in plans"
      :key="plan.id"
      class="rounded-2xl border bg-card p-5 flex flex-col gap-4 transition-all hover:shadow-sm"
      :class="plan.id === activePlanId ? 'border-violet-500/30 bg-violet-500/[0.02]' : ''"
    >
      <!-- Top: icon + name + badge -->
      <div class="flex items-start gap-3">
        <div
          class="shrink-0 w-10 h-10 rounded-xl flex items-center justify-center"
          :class="plan.id === activePlanId ? 'bg-violet-500/15' : 'bg-muted'"
        >
          <Target
            class="w-5 h-5"
            :class="plan.id === activePlanId ? 'text-violet-500' : 'text-muted-foreground'"
          />
        </div>
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <p class="font-semibold text-sm leading-tight truncate">{{ plan.name }}</p>
            <span
              v-if="plan.id === activePlanId"
              class="inline-flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full bg-violet-500/15 text-violet-600 dark:text-violet-400 border border-violet-500/25 shrink-0 font-medium"
            >
              <CheckCircle2 class="w-3 h-3" />
              {{ $t("workoutPlans.progress.activeBadge") }}
            </span>
          </div>
          <div class="flex items-center gap-1.5 mt-1 text-xs text-muted-foreground">
            <span>{{ $t("workoutPlans.weeksDays", { weeks: plan.numberOfWeeks, days: plan.daysPerWeek }) }}</span>
            <span class="opacity-40">·</span>
            <span>{{ $t("workoutPlans.workoutSlots", { count: totalWorkoutSlots(plan) }) }}</span>
          </div>
          <p v-if="plan.description" class="text-xs text-muted-foreground mt-1 line-clamp-2">{{ plan.description }}</p>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex items-center gap-2 pt-1 border-t border-border/60 mt-auto">
        <Button variant="ghost" size="sm" class="gap-1.5 text-muted-foreground" @click="emits('view', plan)">
          <Eye class="w-3.5 h-3.5" />
          {{ $t("workoutPlans.detail.view") }}
        </Button>
        <Button variant="ghost" size="sm" class="gap-1.5 text-muted-foreground" @click="emits('edit', plan)">
          <Pencil class="w-3.5 h-3.5" />
          {{ $t("edit") }}
        </Button>
        <div class="flex-1" />
        <Button
          v-if="plan.id !== activePlanId"
          size="sm"
          :disabled="isStarting || !!activePlanId"
          class="gap-1.5"
          @click="emits('start', plan)"
        >
          <Play class="w-3 h-3 fill-current" />
          {{ $t("workoutPlans.progress.start") }}
        </Button>
      </div>
    </div>
  </div>
</template>
