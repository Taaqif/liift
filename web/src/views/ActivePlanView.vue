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
import { ArrowLeft, Check, Play, Trophy, Dumbbell, History, ChevronDown } from "lucide-vue-next";
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

function findNextDay(fromWeek: number, fromDay: number): { week: number; day: number } | null {
  const ws = weeks.value;
  for (let w = fromWeek; w < ws.length; w++) {
    const weekDays = ws[w]?.days ?? [];
    const startDay = w === fromWeek ? fromDay : 0;
    for (let d = startDay; d < weekDays.length; d++) {
      return { week: w, day: d };
    }
  }
  return null;
}

const nextPosition = computed(() => {
  if (!progress.value) return null;
  return findNextDay(progress.value.current_week, progress.value.current_day + 1);
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

const previewDayIndex = ref<number | null>(null);

function togglePreview(dayIndex: number) {
  previewDayIndex.value = previewDayIndex.value === dayIndex ? null : dayIndex;
}

const featureUnitMap: Record<string, string> = {
  weight: "kg",
  rep: "reps",
  duration: "s",
  distance: "m",
};

function featureLabel(name: string): string {
  const unit = featureUnitMap[name];
  const label = name.charAt(0).toUpperCase() + name.slice(1);
  return unit ? `${label} (${unit})` : label;
}

function formatValue(name: string, value: number): string {
  const formatted = Number.isInteger(value) ? value.toString() : value.toFixed(1);
  const unit = featureUnitMap[name];
  return unit ? `${formatted} ${unit}` : formatted;
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
      <div class="flex flex-wrap justify-center gap-2 mb-6">
        <Button
          v-for="(_, weekIdx) in weeks"
          :key="weekIdx"
          type="button"
          :variant="currentViewWeek === weekIdx ? 'default' : 'outline'"
          class="min-w-12 h-10 text-base font-semibold"
          @click="viewWeek = weekIdx; previewDayIndex = null"
        >
          {{ weekIdx + 1 }}
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
              </CardTitle>
              <div class="flex items-center gap-1">
                <Button
                  v-if="!isCurrentPosition(dayIndex)"
                  variant="ghost"
                  size="sm"
                  class="text-xs h-7"
                  :disabled="isUpdating || !!activeDaySession"
                  @click="handleJumpToDay(currentViewWeek, dayIndex)"
                >
                  {{ $t("workoutPlans.progress.jumpHere") }}
                </Button>
                <Button
                  v-if="day.workoutIds.length > 0"
                  variant="ghost"
                  size="icon"
                  class="size-7"
                  @click="togglePreview(dayIndex)"
                >
                  <ChevronDown
                    class="size-4 transition-transform duration-200"
                    :class="{ 'rotate-180': previewDayIndex === dayIndex }"
                  />
                </Button>
              </div>
            </div>
            <p v-if="day.description" class="text-sm text-muted-foreground mt-0.5">
              {{ day.description }}
            </p>
          </CardHeader>

          <CardContent class="pb-3">
            <div class="space-y-4">
              <div
                v-if="day.workoutIds.length === 0"
                class="text-sm text-muted-foreground"
              >
                {{ $t("workoutPlans.progress.noWorkoutsAssigned") }}
              </div>
              <template v-else>
                <!-- Collapsed: exercise name summary -->
                <div v-if="previewDayIndex !== dayIndex" class="space-y-1">
                  <div
                    v-for="workoutId in day.workoutIds"
                    :key="`summary-${workoutId}`"
                  >
                    <p class="text-sm font-medium">{{ getWorkout(workoutId)?.name ?? `#${workoutId}` }}</p>
                    <p
                      v-if="getWorkout(workoutId)?.exercises?.length"
                      class="text-xs text-muted-foreground truncate"
                    >
                      {{ getWorkout(workoutId)!.exercises.map((e) => e.exercise?.name).filter(Boolean).join(" · ") }}
                    </p>
                  </div>
                </div>

                <!-- Expanded: full workout detail -->
                <div v-else class="space-y-6">
                  <div
                    v-for="workoutId in day.workoutIds"
                    :key="`preview-${workoutId}`"
                    class="space-y-3"
                  >
                    <p class="text-sm font-semibold">{{ getWorkout(workoutId)?.name ?? `#${workoutId}` }}</p>
                    <div
                      v-if="!getWorkout(workoutId)?.exercises?.length"
                      class="text-sm text-muted-foreground"
                    >
                      {{ $t("workoutPlans.detail.noExercises") }}
                    </div>
                    <div v-else class="space-y-4">
                      <div
                        v-for="ex in getWorkout(workoutId)!.exercises"
                        :key="ex.id ?? ex.order"
                      >
                        <div class="flex items-baseline gap-2 mb-1.5">
                          <span class="text-sm font-medium">{{ ex.exercise?.name ?? `#${ex.exercise_id}` }}</span>
                          <span
                            v-if="ex.exercise?.primary_muscle_groups?.length"
                            class="text-xs text-muted-foreground"
                          >
                            {{ ex.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
                          </span>
                          <Button
                            variant="ghost"
                            size="icon"
                            class="size-6 ml-auto text-muted-foreground/60 hover:text-foreground"
                            @click="openLogs(ex.exercise_id, ex.exercise?.name)"
                          >
                            <History class="size-3.5" />
                          </Button>
                        </div>
                        <div v-if="ex.sets.length > 0" class="rounded-md border overflow-hidden text-xs">
                          <table class="w-full">
                            <thead>
                              <tr class="bg-muted/50 text-muted-foreground">
                                <th class="px-3 py-1.5 text-left font-medium w-10">
                                  {{ $t("exercises.logs.set") }}
                                </th>
                                <th
                                  v-for="feat in ex.sets[0]?.features ?? []"
                                  :key="feat.feature_name"
                                  class="px-3 py-1.5 text-right font-medium"
                                >
                                  {{ featureLabel(feat.feature_name) }}
                                </th>
                              </tr>
                            </thead>
                            <tbody class="divide-y divide-border">
                              <tr
                                v-for="(set, sIdx) in ex.sets"
                                :key="set.id ?? sIdx"
                                class="hover:bg-muted/30"
                              >
                                <td class="px-3 py-1.5 text-muted-foreground font-medium">{{ sIdx + 1 }}</td>
                                <td
                                  v-for="feat in set.features"
                                  :key="feat.feature_name"
                                  class="px-3 py-1.5 text-right tabular-nums"
                                >
                                  {{ formatValue(feat.feature_name, feat.value) }}
                                </td>
                              </tr>
                            </tbody>
                          </table>
                        </div>
                        <p v-if="ex.note" class="text-xs text-muted-foreground mt-1.5 italic">
                          {{ ex.note }}
                        </p>
                      </div>
                    </div>
                  </div>
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
