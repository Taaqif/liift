<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import type { WorkoutPlanProgress } from "@/features/workout-plans/types";
import { Button } from "@/components/ui/button";
import { Target, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  progress: WorkoutPlanProgress;
}>();

const router = useRouter();

const progressPct = computed(() => {
  const p = props.progress;
  const total = p.plan.numberOfWeeks * p.plan.daysPerWeek;
  if (total === 0) return 0;
  const done = p.current_week * p.plan.daysPerWeek + p.current_day;
  return Math.min(Math.round((done / total) * 100), 100);
});
</script>

<template>
  <div
    class="relative overflow-hidden rounded-2xl border border-violet-500/30 bg-gradient-to-br from-violet-500/8 via-blue-500/4 to-transparent p-5 shadow-sm"
  >
    <div class="pointer-events-none absolute -bottom-8 -right-8 h-32 w-32 rounded-full bg-violet-400/8 blur-3xl" />

    <div class="flex items-start justify-between gap-4">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-2">
          <Target class="w-3.5 h-3.5 text-violet-500" />
          <span class="text-xs font-semibold uppercase tracking-wider text-violet-600 dark:text-violet-400">
            Active Plan
          </span>
        </div>

        <p class="text-lg font-bold leading-tight truncate">{{ progress.plan.name }}</p>

        <div class="flex items-center gap-3 mt-1 text-sm text-muted-foreground">
          <span>Week {{ progress.current_week + 1 }} of {{ progress.plan.numberOfWeeks }}</span>
          <span class="opacity-40">·</span>
          <span>Day {{ progress.current_day + 1 }} of {{ progress.plan.daysPerWeek }}</span>
        </div>

        <!-- Progress bar -->
        <div class="mt-3 flex items-center gap-2">
          <div class="flex-1 h-1.5 rounded-full bg-violet-200/40 dark:bg-violet-900/40 overflow-hidden">
            <div
              class="h-full rounded-full bg-violet-500 transition-all duration-500"
              :style="{ width: `${progressPct}%` }"
            />
          </div>
          <span class="text-xs text-muted-foreground tabular-nums w-8 text-right">{{ progressPct }}%</span>
        </div>
      </div>

      <Button
        variant="outline"
        size="sm"
        class="shrink-0 border-violet-500/30 hover:bg-violet-500/10 hover:border-violet-500/60"
        @click="router.push('/workout-plans/active')"
      >
        View
        <ChevronRight class="w-3.5 h-3.5 ml-0.5" />
      </Button>
    </div>
  </div>
</template>
