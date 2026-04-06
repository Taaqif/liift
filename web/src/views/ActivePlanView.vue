<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { toast } from "vue-sonner";
import { useActivePlanProgress } from "@/features/workout-plans/composables/useActivePlanProgress";
import { useUpdatePlanPosition } from "@/features/workout-plans/composables/useUpdatePlanPosition";
import { useCompletePlanProgress } from "@/features/workout-plans/composables/useCompletePlanProgress";
import { useStopPlanProgress } from "@/features/workout-plans/composables/useStopPlanProgress";
import { useStartPlanDay } from "@/features/workout-plans/composables/useStartPlanDay";
import { useActiveWorkoutSession } from "@/features/workout-session/composables/useActiveWorkoutSession";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";
import type { Workout } from "@/features/workouts/types";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ArrowLeft, ChevronLeft, ChevronRight, Check, Play, Trophy, Dumbbell, History } from "lucide-vue-next";
import ExerciseLogDrawer from "@/features/exercises/components/ExerciseLogDrawer.vue";

const router = useRouter();
const { t } = useI18n();

const { progress, loading } = useActivePlanProgress();
const { updatePosition, isPending: isUpdating } = useUpdatePlanPosition();
const { completePlan, isPending: isCompleting } = useCompletePlanProgress();
const { stopPlan, isPending: isStopping } = useStopPlanProgress();
const { startDay, isPending: isStarting } = useStartPlanDay();
const { session: activeSession } = useActiveWorkoutSession();

const activeDaySession = computed(() => {
  const s = activeSession.value;
  if (!s || !progress.value) return null;
  if (s.plan_progress_id === progress.value.id) return s;
  return null;
});

const { workouts: allWorkouts } = useWorkouts({ limit: 1000, includeAll: true });
const workoutMap = computed(() => {
  const map = new Map<number, Workout>();
  for (const w of allWorkouts.value) {
    map.set(w.id, w);
  }
  return map;
});

const viewWeek = ref<number | null>(null);
const currentViewWeek = computed(() => {
  if (viewWeek.value !== null) return viewWeek.value;
  return progress.value?.current_week ?? 0;
});

const plan = computed(() => progress.value?.plan ?? null);
const weeks = computed(() => plan.value?.weeks ?? []);
const totalWeeks = computed(() => weeks.value.length);

const currentWeekDays = computed(() => weeks.value[currentViewWeek.value]?.days ?? []);

function isCurrentPosition(dayIndex: number): boolean {
  return (
    progress.value !== null &&
    currentViewWeek.value === progress.value.current_week &&
    dayIndex === progress.value.current_day
  );
}

function isDayPast(dayIndex: number): boolean {
  if (!progress.value) return false;
  const w = currentViewWeek.value;
  const cw = progress.value.current_week;
  const cd = progress.value.current_day;
  if (w < cw) return true;
  if (w === cw && dayIndex < cd) return true;
  return false;
}

function findNextNonRestDay(fromWeek: number, fromDay: number): { week: number; day: number } | null {
  const ws = weeks.value;
  for (let w = fromWeek; w < ws.length; w++) {
    const weekDays = ws[w]?.days ?? [];
    const startDay = w === fromWeek ? fromDay : 0;
    for (let d = startDay; d < weekDays.length; d++) {
      if (!weekDays[d]?.isRest) return { week: w, day: d };
    }
  }
  return null;
}

const nextPosition = computed(() => {
  if (!progress.value) return null;
  return findNextNonRestDay(progress.value.current_week, progress.value.current_day + 1);
});

const isOnLastDay = computed(() => nextPosition.value === null);

async function handleAdvance() {
  if (!progress.value || !nextPosition.value) return;
  await updatePosition({
    id: progress.value.id,
    week: nextPosition.value.week,
    day: nextPosition.value.day,
  });
  viewWeek.value = nextPosition.value.week;
}

async function handleJumpToDay(weekIndex: number, dayIndex: number) {
  if (!progress.value) return;
  await updatePosition({ id: progress.value.id, week: weekIndex, day: dayIndex });
  viewWeek.value = weekIndex;
}

async function handleStartDay() {
  if (!progress.value) return;
  try {
    await startDay(progress.value.id);
  } catch (err) {
    if (err instanceof Error && err.message === "active_session_exists") {
      toast.error(t("workoutSession.activeSessionExists"));
    } else {
      toast.error(t("workoutSession.toasts.saveFailed"));
    }
  }
}

const showCompleteDialog = ref(false);
async function handleComplete() {
  if (!progress.value) return;
  await completePlan(progress.value.id);
  showCompleteDialog.value = false;
  router.push({ name: "workout-plans" });
}

const showStopDialog = ref(false);
async function handleStop() {
  if (!progress.value) return;
  await stopPlan(progress.value.id);
  showStopDialog.value = false;
  router.push({ name: "workout-plans" });
}

function getWorkout(id: number): Workout | undefined {
  return workoutMap.value.get(id);
}

function setLabel(count: number): string {
  return t("workouts.sets", count);
}

const logDrawerOpen = ref(false);
const logExerciseId = ref<number | null>(null);
const logExerciseName = ref<string | undefined>(undefined);

function openLogs(exerciseId: number, name?: string) {
  logExerciseId.value = exerciseId;
  logExerciseName.value = name;
  logDrawerOpen.value = true;
}

</script>

<template>
  <div>
    <div class="mb-8 flex items-center gap-4">
      <Button variant="ghost" size="icon" @click="router.push({ name: 'workout-plans' })">
        <ArrowLeft class="h-4 w-4" />
      </Button>
      <div>
        <h1 class="text-3xl font-bold">{{ $t("workoutPlans.progress.title") }}</h1>
        <p v-if="plan" class="text-muted-foreground">{{ plan.name }}</p>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">{{ $t("loading") }}</div>
    </div>

    <div v-else-if="!progress" class="flex flex-col items-center justify-center py-24 gap-4">
      <p class="text-muted-foreground">{{ $t("workoutPlans.progress.noActivePlan") }}</p>
      <Button @click="router.push({ name: 'workout-plans' })">
        {{ $t("workoutPlans.progress.browsePlans") }}
      </Button>
    </div>

    <template v-else>
      <!-- Active session banner -->
      <div
        v-if="activeDaySession"
        class="mb-6 flex items-center justify-between gap-3 rounded-lg border border-primary/30 bg-primary/5 px-4 py-3"
      >
        <div class="flex items-center gap-2 text-sm font-medium">
          <Dumbbell class="w-4 h-4 text-primary shrink-0" />
          {{ $t("workoutPlans.progress.activeDayInProgress") }}
        </div>
        <Button size="sm" @click="router.push({ name: 'active-workout' })">
          {{ $t("workoutPlans.progress.continueWorkout") }}
        </Button>
      </div>

      <!-- Week navigation -->
      <div class="flex items-center justify-between mb-6">
        <Button
          variant="outline"
          size="icon"
          :disabled="currentViewWeek === 0"
          @click="viewWeek = currentViewWeek - 1"
        >
          <ChevronLeft class="h-4 w-4" />
        </Button>

        <div class="text-center">
          <h2 class="text-xl font-semibold">
            {{ $t("workoutPlans.weekLabel", { number: currentViewWeek + 1 }) }}
          </h2>
          <p class="text-sm text-muted-foreground">
            {{ $t("workoutPlans.progress.weekOf", { current: currentViewWeek + 1, total: totalWeeks }) }}
          </p>
        </div>

        <Button
          variant="outline"
          size="icon"
          :disabled="currentViewWeek >= totalWeeks - 1"
          @click="viewWeek = currentViewWeek + 1"
        >
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>

      <!-- Current position indicator -->
      <div
        v-if="currentViewWeek !== progress.current_week"
        class="mb-4 p-3 bg-muted rounded-lg text-sm text-muted-foreground text-center"
      >
        {{ $t("workoutPlans.progress.currentPosition", {
          week: progress.current_week + 1,
          day: progress.current_day + 1,
        }) }}
        <button
          class="ml-2 underline text-foreground"
          @click="viewWeek = progress.current_week"
        >
          {{ $t("workoutPlans.progress.goToCurrent") }}
        </button>
      </div>

      <!-- Days grid -->
      <div class="space-y-3 mb-8">
        <Card
          v-for="(day, dayIndex) in currentWeekDays"
          :key="dayIndex"
          :class="[
            'transition-all',
            isCurrentPosition(dayIndex) && activeDaySession && 'ring-2 ring-primary bg-primary/5',
            isCurrentPosition(dayIndex) && !activeDaySession && 'ring-2 ring-primary',
            isDayPast(dayIndex) && 'opacity-60',
          ]"
        >
          <CardHeader class="pb-2">
            <div class="flex items-center justify-between gap-3">
              <CardTitle class="text-base flex items-center gap-2">
                <Check
                  v-if="isDayPast(dayIndex)"
                  class="size-4 text-green-600 dark:text-green-400 shrink-0"
                />
                <Dumbbell
                  v-else-if="isCurrentPosition(dayIndex) && activeDaySession"
                  class="size-4 text-primary shrink-0 animate-pulse"
                />
                <span
                  v-else-if="isCurrentPosition(dayIndex)"
                  class="size-2 rounded-full bg-primary shrink-0 inline-block"
                />
                {{ $t("workoutPlans.dayLabel", { number: dayIndex + 1 }) }}
                <span v-if="day.isRest" class="text-xs font-normal text-muted-foreground">
                  — {{ $t("workoutPlans.restDay") }}
                </span>
              </CardTitle>
              <Button
                v-if="!day.isRest && !isCurrentPosition(dayIndex)"
                variant="ghost"
                size="sm"
                class="text-xs h-7"
                :disabled="isUpdating || !!activeDaySession"
                @click="handleJumpToDay(currentViewWeek, dayIndex)"
              >
                {{ $t("workoutPlans.progress.jumpHere") }}
              </Button>
            </div>
          </CardHeader>

          <CardContent v-if="!day.isRest">
            <div class="space-y-4">
              <div
                v-if="day.workoutIds.length === 0"
                class="text-sm text-muted-foreground"
              >
                {{ $t("workoutPlans.progress.noWorkoutsAssigned") }}
              </div>
              <template v-else>
                <!-- Workouts with their exercises -->
                <div
                  v-for="workoutId in day.workoutIds"
                  :key="`w-${workoutId}`"
                  class="space-y-1"
                >
                  <p class="text-sm font-medium">
                    {{ getWorkout(workoutId)?.name ?? `#${workoutId}` }}
                  </p>
                  <ul class="space-y-1 pl-3">
                    <li
                      v-for="we in (getWorkout(workoutId)?.exercises ?? [])"
                      :key="we.id"
                      class="flex items-center justify-between gap-2 text-sm text-muted-foreground"
                    >
                      <div class="flex items-center gap-2 min-w-0">
                        <span class="size-1.5 rounded-full bg-muted-foreground/40 shrink-0" />
                        <span class="truncate">{{ we.exercise?.name ?? `#${we.exercise_id}` }}</span>
                      </div>
                      <div class="flex items-center gap-1 shrink-0">
                        <Button
                          variant="ghost"
                          size="icon"
                          class="size-6 text-muted-foreground/60 hover:text-foreground"
                          @click="openLogs(we.exercise_id, we.exercise?.name)"
                        >
                          <History class="size-3.5" />
                        </Button>
                        <span class="text-xs text-muted-foreground/70">
                          {{ we.sets.length }} {{ setLabel(we.sets.length) }}
                        </span>
                      </div>
                    </li>
                  </ul>
                </div>

                <template v-if="isCurrentPosition(dayIndex)">
                  <Button
                    v-if="activeDaySession"
                    class="w-full"
                    @click="router.push({ name: 'active-workout' })"
                  >
                    <Dumbbell class="w-4 h-4 mr-2" />
                    {{ $t("workoutPlans.progress.continueWorkout") }}
                  </Button>
                  <Button
                    v-else
                    class="w-full"
                    :disabled="isStarting"
                    @click="handleStartDay"
                  >
                    <Play class="w-4 h-4 mr-2" />
                    {{ isStarting ? $t("workoutPlans.progress.startingDay") : $t("workoutPlans.progress.startDay") }}
                  </Button>
                </template>
              </template>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Action buttons -->
      <div class="space-y-3 border-t pt-6">
        <Button
          v-if="isOnLastDay && currentViewWeek === progress.current_week"
          class="w-full bg-green-600 hover:bg-green-700 text-white"
          :disabled="isCompleting"
          @click="showCompleteDialog = true"
        >
          <Trophy class="w-4 h-4 mr-2" />
          {{ $t("workoutPlans.progress.completePlan") }}
        </Button>

        <Button
          v-else-if="currentViewWeek === progress.current_week && !isOnLastDay"
          class="w-full"
          :disabled="isUpdating || !!activeDaySession"
          @click="handleAdvance"
        >
          {{ $t("workoutPlans.progress.advanceToNext") }}
        </Button>

        <Button
          variant="outline"
          class="w-full text-destructive hover:text-destructive"
          :disabled="isStopping"
          @click="showStopDialog = true"
        >
          {{ $t("workoutPlans.progress.stopPlan") }}
        </Button>
      </div>

      <!-- Complete confirmation dialog -->
      <Dialog v-model:open="showCompleteDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutPlans.progress.completeDialog.title") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutPlans.progress.completeDialog.description") }}</DialogDescription>
          </DialogHeader>
          <DialogFooter class="flex-col gap-2 sm:flex-col">
            <Button
              class="bg-green-600 hover:bg-green-700 text-white"
              :disabled="isCompleting"
              @click="handleComplete"
            >
              <Trophy class="w-4 h-4 mr-2" />
              {{ isCompleting ? $t("workoutPlans.progress.completing") : $t("workoutPlans.progress.completePlan") }}
            </Button>
            <Button variant="outline" :disabled="isCompleting" @click="showCompleteDialog = false">
              {{ $t("cancel") }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Stop confirmation dialog -->
      <Dialog v-model:open="showStopDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutPlans.progress.stopDialog.title") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutPlans.progress.stopDialog.description") }}</DialogDescription>
          </DialogHeader>
          <DialogFooter class="flex-col gap-2 sm:flex-col">
            <Button
              variant="destructive"
              :disabled="isStopping"
              @click="handleStop"
            >
              {{ isStopping ? $t("workoutPlans.progress.stopping") : $t("workoutPlans.progress.stopPlan") }}
            </Button>
            <Button variant="outline" :disabled="isStopping" @click="showStopDialog = false">
              {{ $t("cancel") }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </template>
  </div>

  <ExerciseLogDrawer
    v-model:open="logDrawerOpen"
    :exercise-id="logExerciseId"
    :exercise-name="logExerciseName"
  />
</template>
