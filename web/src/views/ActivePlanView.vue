<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useQueryClient } from "@tanstack/vue-query";
import { useI18n } from "vue-i18n";
import { toast } from "vue-sonner";
import { apiClient } from "@/lib/api";
import { workoutSessionKeys } from "@/lib/queryKeys";
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Check, Play, Trophy, Dumbbell, History, ChevronDown, CalendarDays, ChevronRight } from "lucide-vue-next";
import ExerciseLogDrawer from "@/features/exercises/components/ExerciseLogDrawer.vue";

const router = useRouter();
const { t } = useI18n();
const queryClient = useQueryClient();

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

const showPlanConflictDialog = ref(false);
const isStoppingAndStartingDay = ref(false);

async function handleStartDay() {
  if (!progress.value) return;
  try {
    await startDay(progress.value.id);
  } catch (err) {
    if (err instanceof Error && err.message === "active_session_exists") {
      showPlanConflictDialog.value = true;
    } else {
      toast.error(t("workoutSession.toasts.saveFailed"));
    }
  }
}

async function handleStopAndStartDay() {
  if (!activeSession.value || !progress.value) return;
  isStoppingAndStartingDay.value = true;
  try {
    await apiClient.post(`/workout-sessions/${activeSession.value.id}/cancel`);
    queryClient.removeQueries({ queryKey: workoutSessionKeys.active() });
    queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
    await startDay(progress.value.id);
    showPlanConflictDialog.value = false;
  } catch {
    toast.error(t("workoutSession.toasts.saveFailed"));
  } finally {
    isStoppingAndStartingDay.value = false;
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
  <div class="pb-10">

    <!-- Header -->
    <div class="mb-8">
      <button
        class="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors mb-3"
        @click="router.push({ name: 'workout-plans' })"
      >
        ← {{ $t("nav.workoutPlans") }}
      </button>

      <div v-if="loading" class="h-9 w-48 rounded-lg bg-muted animate-pulse" />
      <template v-else-if="plan">
        <h1 class="text-3xl font-bold tracking-tight">{{ plan.name }}</h1>
        <div class="flex items-center gap-2 mt-2">
          <CalendarDays class="size-3.5 text-muted-foreground shrink-0" />
          <span class="text-sm text-muted-foreground">
            {{ $t("workoutPlans.progress.title") }}
            · {{ $t("workoutPlans.progress.weekOf", { current: (progress?.current_week ?? 0) + 1, total: totalWeeks }) }}
          </span>
        </div>
      </template>
    </div>

    <div v-if="loading" class="space-y-3">
      <div v-for="i in 3" :key="i" class="h-20 rounded-xl bg-muted animate-pulse" />
    </div>

    <div v-else-if="!progress" class="flex flex-col items-center justify-center py-24 gap-4 text-center">
      <p class="text-muted-foreground">{{ $t("workoutPlans.progress.noActivePlan") }}</p>
      <Button @click="router.push({ name: 'workout-plans' })">
        {{ $t("workoutPlans.progress.browsePlans") }}
      </Button>
    </div>

    <template v-else>
      <!-- Active session pill -->
      <div
        v-if="activeDaySession"
        class="mb-6 flex items-center justify-between gap-3 rounded-xl bg-primary px-4 py-3 text-primary-foreground shadow-sm"
      >
        <div class="flex items-center gap-2 text-sm font-medium">
          <Dumbbell class="w-4 h-4 shrink-0 animate-pulse" />
          {{ $t("workoutPlans.progress.activeDayInProgress") }}
        </div>
        <button
          class="flex items-center gap-1 text-sm font-semibold opacity-90 hover:opacity-100 transition-opacity"
          @click="router.push({ name: 'active-workout' })"
        >
          {{ $t("workoutPlans.progress.continueWorkout") }}
          <ChevronRight class="size-4" />
        </button>
      </div>

      <!-- Week tabs -->
      <div class="flex gap-1.5 overflow-x-auto pb-1 mb-6 no-scrollbar">
        <button
          v-for="(_, weekIdx) in weeks"
          :key="weekIdx"
          type="button"
          class="shrink-0 inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-sm font-medium transition-colors"
          :class="currentViewWeek === weekIdx
            ? 'bg-foreground text-background'
            : 'text-muted-foreground hover:text-foreground hover:bg-muted'"
          @click="viewWeek = weekIdx; previewDayIndex = null"
        >
          Week {{ weekIdx + 1 }}
          <span
            v-if="weekIdx === progress.current_week && currentViewWeek !== weekIdx"
            class="size-1.5 rounded-full bg-primary inline-block"
          />
        </button>
      </div>

      <!-- Off-track nudge -->
      <div
        v-if="currentViewWeek !== progress.current_week"
        class="mb-5 flex items-center justify-between gap-2 rounded-xl border bg-muted/40 px-4 py-2.5 text-sm text-muted-foreground"
      >
        <span>
          {{ $t("workoutPlans.progress.currentPosition", {
            week: progress.current_week + 1,
            day: progress.current_day + 1,
          }) }}
        </span>
        <button
          class="font-medium text-foreground hover:underline shrink-0"
          @click="viewWeek = progress.current_week"
        >
          {{ $t("workoutPlans.progress.goToCurrent") }}
        </button>
      </div>

      <!-- Days timeline -->
      <div class="relative mb-8">
        <!-- Vertical track -->
        <div class="absolute left-[17px] top-5 bottom-5 w-px bg-border" />

        <div class="space-y-2">
          <div
            v-for="(day, dayIndex) in currentWeekDays"
            :key="dayIndex"
            class="relative pl-11"
          >
            <!-- Node -->
            <div
              class="absolute left-0 top-4 flex size-[34px] items-center justify-center rounded-full border-2 bg-background transition-colors"
              :class="[
                isDayPast(dayIndex) ? 'border-green-500 text-green-500' : '',
                isCurrentPosition(dayIndex) && !isDayPast(dayIndex) ? 'border-primary text-primary' : '',
                !isDayPast(dayIndex) && !isCurrentPosition(dayIndex) ? 'border-border text-muted-foreground' : '',
              ]"
            >
              <Check v-if="isDayPast(dayIndex)" class="size-4" />
              <Dumbbell v-else-if="isCurrentPosition(dayIndex) && activeDaySession" class="size-3.5 animate-pulse" />
              <span v-else-if="isCurrentPosition(dayIndex)" class="size-2 rounded-full bg-primary block" />
              <span v-else class="text-xs font-semibold">{{ dayIndex + 1 }}</span>
            </div>

            <!-- Card -->
            <div
              class="rounded-xl border transition-all"
              :class="[
                isCurrentPosition(dayIndex) ? 'border-primary/40 bg-primary/[0.03] shadow-sm' : 'bg-card',
                isDayPast(dayIndex) ? 'opacity-55' : '',
              ]"
            >
              <!-- Header row -->
              <button
                class="flex w-full items-center gap-3 px-4 pt-4 pb-3 text-left"
                :disabled="day.workoutIds.length === 0"
                @click="day.workoutIds.length > 0 && togglePreview(dayIndex)"
              >
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-semibold leading-none mb-1">
                    {{ $t("workoutPlans.dayLabel", { number: dayIndex + 1 }) }}
                    <span
                      v-if="isCurrentPosition(dayIndex) && !isDayPast(dayIndex)"
                      class="ml-2 inline-flex items-center rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary uppercase tracking-wide"
                    >Today</span>
                  </p>
                  <p v-if="day.description" class="text-xs text-muted-foreground">{{ day.description }}</p>
                  <p v-else-if="day.workoutIds.length > 0" class="text-xs text-muted-foreground truncate">
                    {{ day.workoutIds.map(id => getWorkout(id)?.name).filter(Boolean).join(" · ") }}
                  </p>
                  <p v-else class="text-xs text-muted-foreground">{{ $t("workoutPlans.progress.noWorkoutsAssigned") }}</p>
                </div>
                <div class="flex items-center gap-1 shrink-0">
                  <button
                    v-if="!isCurrentPosition(dayIndex) && !isDayPast(dayIndex)"
                    class="text-xs text-muted-foreground hover:text-foreground transition-colors px-2 py-1 rounded-md hover:bg-muted"
                    :disabled="isUpdating || !!activeDaySession"
                    @click.stop="handleJumpToDay(currentViewWeek, dayIndex)"
                  >
                    {{ $t("workoutPlans.progress.jumpHere") }}
                  </button>
                  <ChevronDown
                    v-if="day.workoutIds.length > 0"
                    class="size-4 text-muted-foreground transition-transform duration-200"
                    :class="{ 'rotate-180': previewDayIndex === dayIndex }"
                  />
                </div>
              </button>

              <!-- Expanded detail -->
              <div v-if="previewDayIndex === dayIndex && day.workoutIds.length > 0" class="px-4 pb-4 border-t pt-3 space-y-5">
                <div
                  v-for="workoutId in day.workoutIds"
                  :key="`preview-${workoutId}`"
                  class="space-y-3"
                >
                  <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                    {{ getWorkout(workoutId)?.name ?? `#${workoutId}` }}
                  </p>
                  <div v-if="!getWorkout(workoutId)?.exercises?.length" class="text-sm text-muted-foreground">
                    {{ $t("workoutPlans.detail.noExercises") }}
                  </div>
                  <div v-else class="space-y-4">
                    <div v-for="ex in getWorkout(workoutId)!.exercises" :key="ex.id ?? ex.order">
                      <div class="flex items-baseline gap-2 mb-1.5">
                        <span class="text-sm font-medium">{{ ex.exercise?.name ?? `#${ex.exercise_id}` }}</span>
                        <span v-if="ex.exercise?.primary_muscle_groups?.length" class="text-xs text-muted-foreground">
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
                      <div v-if="ex.sets.length > 0" class="rounded-lg border overflow-hidden text-xs">
                        <table class="w-full">
                          <thead>
                            <tr class="bg-muted/50 text-muted-foreground">
                              <th class="px-3 py-1.5 text-left font-medium w-10">{{ $t("exercises.logs.set") }}</th>
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
                            <tr v-for="(set, sIdx) in ex.sets" :key="set.id ?? sIdx" class="hover:bg-muted/30">
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
                      <p v-if="ex.note" class="text-xs text-muted-foreground mt-1.5 italic">{{ ex.note }}</p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- CTA for current day -->
              <div v-if="isCurrentPosition(dayIndex) && day.workoutIds.length > 0" class="px-4 pb-4" :class="{ 'pt-0': previewDayIndex !== dayIndex, 'pt-3': previewDayIndex === dayIndex }">
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
                  <Play class="w-4 h-4 mr-2 fill-current" />
                  {{ isStarting ? $t("workoutPlans.progress.startingDay") : $t("workoutPlans.progress.startDay") }}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom actions -->
      <div class="space-y-2.5 pt-4 border-t">
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
          variant="ghost"
          class="w-full text-muted-foreground hover:text-destructive"
          :disabled="isStopping"
          @click="showStopDialog = true"
        >
          {{ $t("workoutPlans.progress.stopPlan") }}
        </Button>
      </div>

      <!-- Active session conflict dialog -->
      <Dialog v-model:open="showPlanConflictDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutSession.conflictDialog.title") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutSession.conflictDialog.description") }}</DialogDescription>
          </DialogHeader>
          <DialogFooter class="flex-col gap-2 sm:flex-col">
            <Button variant="destructive" :disabled="isStoppingAndStartingDay" @click="handleStopAndStartDay">
              {{ isStoppingAndStartingDay ? $t("workoutSession.conflictDialog.stopping") : $t("workoutSession.conflictDialog.stopAndStartDay") }}
            </Button>
            <Button variant="outline" :disabled="isStoppingAndStartingDay" @click="showPlanConflictDialog = false; router.push({ name: 'active-workout' })">
              {{ $t("workoutSession.conflictDialog.goToCurrent") }}
            </Button>
            <Button variant="ghost" :disabled="isStoppingAndStartingDay" @click="showPlanConflictDialog = false">
              {{ $t("cancel") }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <!-- Complete confirmation dialog -->
      <Dialog v-model:open="showCompleteDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutPlans.progress.completeDialog.title") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutPlans.progress.completeDialog.description") }}</DialogDescription>
          </DialogHeader>
          <DialogFooter class="flex-col gap-2 sm:flex-col">
            <Button class="bg-green-600 hover:bg-green-700 text-white" :disabled="isCompleting" @click="handleComplete">
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
            <Button variant="destructive" :disabled="isStopping" @click="handleStop">
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
