<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { useElapsedTimer, formatElapsed } from "@/composables/useElapsedTimer";
import { useRouter } from "vue-router";
import { useQueryClient } from "@tanstack/vue-query";
import { useActiveWorkoutSession } from "@/features/workout-session/composables/useActiveWorkoutSession";
import { useUpdateWorkoutSession } from "@/features/workout-session/composables/useUpdateWorkoutSession";
import { useEndWorkoutSession } from "@/features/workout-session/composables/useEndWorkoutSession";
import { useCancelWorkoutSession } from "@/features/workout-session/composables/useCancelWorkoutSession";
import { useDeleteWorkoutSession } from "@/features/workout-session/composables/useDeleteWorkoutSession";
import { useAddExerciseToSession } from "@/features/workout-session/composables/useAddExerciseToSession";
import { workoutSessionKeys } from "@/lib/queryKeys";
import type { Exercise } from "@/features/exercises/types";
import ExercisePickerSheet from "@/features/exercises/components/ExercisePickerSheet.vue";
import type {
  WorkoutSession,
  WorkoutSessionExercise,
  WorkoutSessionSet,
  UpdateWorkoutSessionPayload,
} from "@/features/workout-session/types";
import { Button } from "@/components/ui/button";
import GymValueInput from "@/features/workout-session/components/GymValueInput.vue";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import {
  Plus,
  Minus,
  Timer,
  StopCircle,
  Check,
  Circle,
  History,
  ChevronLeft,
  ChevronRight,
  LayoutList,
  Trash2,
} from "lucide-vue-next";
import { useI18n } from "vue-i18n";
import { toast } from "vue-sonner";
import ExerciseLogDrawer from "@/features/exercises/components/ExerciseLogDrawer.vue";
import ExerciseInfoDialog from "@/features/exercises/components/ExerciseInfoDialog.vue";
import RestCountdownDonut from "@/features/workout-session/components/RestCountdownDonut.vue";

const router = useRouter();
const queryClient = useQueryClient();
const { t } = useI18n();
const { session, loading, error, refetch } = useActiveWorkoutSession();
const endingWorkout = ref(false);

const localSession = ref<WorkoutSession | null>(null);
const restRemaining = ref<number | null>(null);
const restTotal = ref<number>(0);
const restExerciseId = ref<number | null>(null);
const restSetIdx = ref<number | null>(null);
const restLabel = ref<string>("");
let restTimerId: ReturnType<typeof setInterval> | null = null;

watch(
  session,
  (s) => {
    if (s) {
      localSession.value = JSON.parse(JSON.stringify(s));
    } else {
      localSession.value = null;
    }
  },
  { immediate: true },
);

const sessionId = computed(() => localSession.value?.id ?? 0);
const isActive = computed(
  () => !!localSession.value && !localSession.value.ended_at,
);
const endSessionMutation = useEndWorkoutSession(sessionId);
const cancelSessionMutation = useCancelWorkoutSession(sessionId);
const deleteSessionMutation = useDeleteWorkoutSession();
const updateSessionMutation = useUpdateWorkoutSession(sessionId);
const addExerciseMutation = useAddExerciseToSession(sessionId);

const showAddExercisePicker = ref(false);

async function onSelectExercise(exerciseId: number) {
  if (sessionId.value === 0) return;
  try {
    const updated = await addExerciseMutation.addExercise({
      exerciseId,
      restTimer: 60,
    });
    if (updated) {
      localSession.value = JSON.parse(JSON.stringify(updated));
    }
    showAddExercisePicker.value = false;
  } catch (err) {
    console.error("Failed to add exercise:", err);
    toast.error(t("workoutSession.toasts.saveFailed"));
  }
}

async function handleAddExercises(exercises: Exercise[]) {
  for (const ex of exercises) {
    await onSelectExercise(ex.id);
  }
}

function buildPayload(): UpdateWorkoutSessionPayload {
  const s = localSession.value;
  if (!s) return { exercises: [] };
  return {
    exercises: s.exercises.map((ex) => ({
      id: ex.id,
      workout_exercise_id: ex.workout_exercise_id ?? 0,
      order: ex.order,
      note: ex.note,
      rest_timer: ex.rest_timer ?? 0,
      sets: ex.sets.map((set) => ({
        id: set.id,
        workout_set_id: set.workout_set_id ?? undefined,
        order: set.order,
        completed_at: set.completed_at,
        values: set.values.map((v) => ({
          id: v.id,
          feature_name: v.feature_name,
          value: v.value,
        })),
      })),
    })),
  };
}

async function save(): Promise<void> {
  if (sessionId.value === 0) return;
  const payload = buildPayload();
  await updateSessionMutation.updateSession(payload);
  await refetch();
}

const defaultRestSeconds = 60;
const REST_TIMER_KEY = "liift:restTimer";

// Donut timer: r=8 in a 20×20 viewBox, stroke-width=3 leaves a visible hole

function clearRestTimerStorage() {
  localStorage.removeItem(REST_TIMER_KEY);
}

function tickRestTimer() {
  const stored = localStorage.getItem(REST_TIMER_KEY);
  if (!stored) {
    restRemaining.value = null;
    return;
  }
  try {
    const { startedAt, durationSeconds } = JSON.parse(stored) as {
      startedAt: string;
      durationSeconds: number;
    };
    const elapsed = Math.floor(
      (Date.now() - new Date(startedAt).getTime()) / 1000,
    );
    const remaining = durationSeconds - elapsed;
    if (remaining <= 0) {
      clearRestTimerStorage();
      if (restTimerId) {
        clearInterval(restTimerId);
        restTimerId = null;
      }
      restRemaining.value = null;
    } else {
      restRemaining.value = remaining;
    }
  } catch {
    clearRestTimerStorage();
    restRemaining.value = null;
  }
}

function startRestTimer(ex: WorkoutSessionExercise, setIdx: number) {
  const restSeconds = ex.rest_timer > 0 ? ex.rest_timer : defaultRestSeconds;
  restTotal.value = restSeconds;
  restLabel.value = ex.exercise?.name ?? "";

  // Point to the target set — the one we're resting *for*
  const exs = localSession.value?.exercises ?? [];
  const exIdx = exs.findIndex((e) => e.id === ex.id);
  let targetExId: number | null = null;
  let targetSetIdx: number | null = null;
  if (setIdx < ex.sets.length - 1) {
    // Next set in the same exercise
    targetExId = ex.id;
    targetSetIdx = setIdx + 1;
  } else if (exIdx >= 0 && exIdx < exs.length - 1) {
    // Last set of this exercise — rest is for the first set of the next exercise
    const nextExercise = exs[exIdx + 1];
    targetExId = nextExercise?.id ?? null;
    targetSetIdx = 0;
  }
  restExerciseId.value = targetExId;
  restSetIdx.value = targetSetIdx;

  localStorage.setItem(
    REST_TIMER_KEY,
    JSON.stringify({
      startedAt: new Date().toISOString(),
      durationSeconds: restSeconds,
      restExerciseId: targetExId,
      restSetIdx: targetSetIdx,
    }),
  );
  if (restTimerId) clearInterval(restTimerId);
  tickRestTimer();
  restTimerId = setInterval(tickRestTimer, 1000);
}

async function setCompleted(
  ex: WorkoutSessionExercise,
  setIdx: number,
  checked: boolean,
  force = false,
) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const sets = localSession.value.exercises[exIndex]?.sets;
  if (!sets || setIdx < 0 || setIdx >= sets.length) return;
  const s = sets[setIdx];
  if (!s) return;

  // If completing the last set while earlier sets are still unchecked, ask first
  if (
    !force &&
    checked &&
    setIdx === sets.length - 1 &&
    sets.slice(0, setIdx).some((s) => !s.completed_at)
  ) {
    pendingSetCompletion.value = { ex, setIdx };
    showSkippedSetsDialog.value = true;
    return;
  }

  s.completed_at = checked ? new Date().toISOString() : null;

  if (checked) {
    startRestTimer(ex, setIdx);
  } else {
    if (restTimerId) clearInterval(restTimerId);
    restTimerId = null;
    restRemaining.value = null;
    restExerciseId.value = null;
    restSetIdx.value = null;
    clearRestTimerStorage();
  }

  const viewedExIndex = activeExIndex.value;
  const isLastExercise = exIndex === localSession.value.exercises.length - 1;
  const isLastSet = setIdx === sets.length - 1;

  if (sessionId.value === 0) return;
  const payload = buildPayload();
  try {
    const updated = await updateSessionMutation.updateSession(payload);
    if (updated) {
      localSession.value = JSON.parse(JSON.stringify(updated));
    }
  } catch (err) {
    s.completed_at = checked ? null : new Date().toISOString();
    toast.error(t("workoutSession.toasts.saveFailed"));
    console.error("Failed to save set completion:", err);
    return;
  }

  if (checked) {
    // Prompt to end if this was the last set of the last exercise
    if (isLastExercise && isLastSet) {
      openFinishDialog();
      return;
    }
    // Auto-advance only if the user is still viewing the exercise that just became fully complete
    if (viewedExIndex === activeExIndex.value) {
      const ex = localSession.value?.exercises[viewedExIndex];
      if (
        ex &&
        isExerciseComplete(ex) &&
        viewedExIndex < (localSession.value?.exercises.length ?? 1) - 1
      ) {
        setTimeout(() => {
          activeExIndex.value++;
        }, 400);
      }
    }
  }
}

async function confirmSkippedSets() {
  showSkippedSetsDialog.value = false;
  const pending = pendingSetCompletion.value;
  pendingSetCompletion.value = null;
  if (!pending) return;
  await setCompleted(pending.ex, pending.setIdx, true, true);
}

function isSetCheckboxDisabled(
  ex: WorkoutSessionExercise,
  setIdx: number,
): boolean {
  const set = ex.sets[setIdx];
  // Already ticked — always allow unchecking
  if (set?.completed_at) return false;
  // Unticked — only disable if the previous set isn't done yet
  if (setIdx === 0) return false;
  const prev = ex.sets[setIdx - 1];
  return !prev?.completed_at;
}

function updateSetValue(
  ex: WorkoutSessionExercise,
  set: WorkoutSessionSet,
  valueIndex: number,
  value: number,
) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const sets = localSession.value.exercises[exIndex]?.sets;
  const setIndex = sets?.findIndex((s) => s.id === set.id) ?? -1;
  if (setIndex === -1 || !sets) return;
  const v = sets[setIndex]?.values[valueIndex];
  if (v) v.value = value;
}

function updateNote(ex: WorkoutSessionExercise, note: string) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const e = localSession.value.exercises[exIndex];
  if (e) e.note = note;
}

function addSet(ex: WorkoutSessionExercise) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const exItem = localSession.value.exercises[exIndex];
  if (!exItem) return;
  const lastSet = ex.sets[ex.sets.length - 1];
  const newSet: WorkoutSessionSet = {
    id: 0,
    order: ex.sets.length,
    completed_at: null,
    values:
      lastSet?.values.map((v) => ({
        id: 0,
        feature_name: v.feature_name,
        value: v.value,
      })) ?? [],
  };
  exItem.sets.push(newSet);
  save();
}

function removeSet(ex: WorkoutSessionExercise, setIdx: number) {
  if (!localSession.value) return;
  const exItem = localSession.value.exercises.find((e) => e.id === ex.id);
  if (!exItem || exItem.sets.length <= 1) return;
  exItem.sets.splice(setIdx, 1);
  // Re-number orders
  exItem.sets.forEach((s, i) => {
    s.order = i;
  });
  save();
}

const showSkippedSetsDialog = ref(false);
// showCancelDialog / showEndDialog removed — use openFinishDialog("cancel"/"complete")
const pendingSetCompletion = ref<{
  ex: WorkoutSessionExercise;
  setIdx: number;
} | null>(null);

const showRemoveExerciseDialog = ref(false);
const exerciseToRemove = ref<WorkoutSessionExercise | null>(null);

function confirmRemoveExercise(ex: WorkoutSessionExercise) {
  exerciseToRemove.value = ex;
  showRemoveExerciseDialog.value = true;
}

function removeExercise() {
  const ex = exerciseToRemove.value;
  if (!ex || !localSession.value) return;
  showRemoveExerciseDialog.value = false;
  const idx = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (idx === -1) return;
  localSession.value.exercises.splice(idx, 1);
  // Clamp activeExIndex so it stays in bounds
  if (activeExIndex.value >= localSession.value.exercises.length) {
    activeExIndex.value = Math.max(0, localSession.value.exercises.length - 1);
  }
  save();
}

function adjustRestTimer(ex: WorkoutSessionExercise, delta: number) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const exItem = localSession.value.exercises[exIndex];
  if (!exItem) return;
  const current =
    exItem.rest_timer > 0 ? exItem.rest_timer : defaultRestSeconds;
  exItem.rest_timer = Math.max(15, current + delta);
  save();
}

const { elapsedSeconds } = useElapsedTimer(
  () => localSession.value?.started_at,
);

// ── Unified finish dialog ─────────────────────────────────
const showFinishDialog = ref(false);
const cancellingWorkout = ref(false);

function openFinishDialog() {
  showFinishDialog.value = true;
}

const incompleteSets = computed(() => {
  const exs = localSession.value?.exercises ?? [];
  return exs.flatMap((ex) =>
    ex.sets
      .map((s, i) => ({ ex, setIdx: i, set: s }))
      .filter(({ set }) => !set.completed_at),
  );
});

// Blank quick-start session with zero completed sets — no activity worth keeping
const isBlankWithNoActivity = computed(() => {
  const s = localSession.value;
  if (!s || s.workout) return false;
  return !s.exercises.some((ex) => ex.sets.some((set) => set.completed_at));
});

async function handleCancelWorkout() {
  if (sessionId.value === 0) return;
  showFinishDialog.value = false;
  cancellingWorkout.value = true;
  try {
    await cancelSessionMutation.cancelSession();
  } catch (err) {
    toast.error(t("workoutSession.toasts.cancelFailed"));
    console.error("Failed to cancel workout:", err);
  } finally {
    cancellingWorkout.value = false;
  }
}

async function handleEndWorkout() {
  if (sessionId.value === 0) return;
  showFinishDialog.value = false;
  endingWorkout.value = true;
  try {
    if (isBlankWithNoActivity.value) {
      await deleteSessionMutation.deleteSession(sessionId.value);
      queryClient.setQueryData(workoutSessionKeys.active(), null);
      router.push({ name: "workouts" });
    } else {
      await endSessionMutation.endSession();
    }
  } catch (err) {
    toast.error(t("workoutSession.toasts.endFailed"));
    console.error("Failed to end workout:", err);
  } finally {
    endingWorkout.value = false;
  }
}

watch(
  [loading, session],
  ([l, s]) => {
    if (!l && s === null) {
      router.replace({ name: "workouts" });
    }
  },
  { immediate: true },
);

function onVisibilityChange() {
  if (document.visibilityState === "visible") tickRestTimer();
}

onMounted(() => {
  // Restore rest timer if one was running before the tab went away
  const stored = localStorage.getItem(REST_TIMER_KEY);
  if (stored) {
    try {
      const parsed = JSON.parse(stored) as {
        startedAt: string;
        durationSeconds: number;
        restExerciseId?: number | null;
        restSetIdx?: number | null;
      };
      restTotal.value = parsed.durationSeconds;
      restExerciseId.value = parsed.restExerciseId ?? null;
      restSetIdx.value = parsed.restSetIdx ?? null;
      tickRestTimer();
      if (restRemaining.value !== null) {
        if (restTimerId) clearInterval(restTimerId);
        restTimerId = setInterval(tickRestTimer, 1000);
      }
    } catch {
      clearRestTimerStorage();
    }
  }
  document.addEventListener("visibilitychange", onVisibilityChange);
});

onUnmounted(() => {
  if (restTimerId) clearInterval(restTimerId);
  document.removeEventListener("visibilitychange", onVisibilityChange);
});

const logDrawerOpen = ref(false);
const logExerciseId = ref<number | null>(null);
const logExerciseName = ref<string | undefined>(undefined);

function openLogs(exerciseId: number, name?: string) {
  logExerciseId.value = exerciseId;
  logExerciseName.value = name;
  logDrawerOpen.value = true;
}

function isExerciseComplete(ex: WorkoutSessionExercise): boolean {
  return ex.sets.length > 0 && ex.sets.every((s) => !!s.completed_at);
}

function completedSetsCount(ex: WorkoutSessionExercise): number {
  return ex.sets.filter((s) => !!s.completed_at).length;
}

// Manually navigable active exercise index
const activeExIndex = ref(0);

// When session first loads, jump to the first incomplete exercise
watch(
  localSession,
  (s) => {
    if (!s) return;
    const firstIncomplete = s.exercises.findIndex(
      (ex) => !isExerciseComplete(ex),
    );
    activeExIndex.value = firstIncomplete >= 0 ? firstIncomplete : 0;
  },
  { once: true, immediate: true },
);

const exercises = computed(() => localSession.value?.exercises ?? []);

const currentEx = computed(() => exercises.value[activeExIndex.value] ?? null);

const nextEx = computed(() => {
  const next = activeExIndex.value + 1;
  return next < exercises.value.length ? (exercises.value[next] ?? null) : null;
});

const prevEx = computed(() => {
  const prev = activeExIndex.value - 1;
  return prev >= 0 ? (exercises.value[prev] ?? null) : null;
});

const canGoPrev = computed(() => activeExIndex.value > 0);
const canGoNext = computed(
  () => activeExIndex.value < exercises.value.length - 1,
);

function goToPrev() {
  if (canGoPrev.value) activeExIndex.value--;
}

function goToNext() {
  if (canGoNext.value) activeExIndex.value++;
}

function jumpToExercise(idx: number) {
  activeExIndex.value = idx;
  showWorkoutSheet.value = false;
}

const allExercisesComplete = computed(() => {
  const exs = localSession.value?.exercises ?? [];
  return exs.length > 0 && exs.every((ex) => isExerciseComplete(ex));
});

watch(allExercisesComplete, (complete) => {
  if (complete && isActive.value) {
    openFinishDialog();
  }
});

const showWorkoutSheet = ref(false);

function chunk2<T>(arr: T[]): T[][] {
  const result: T[][] = [];
  for (let i = 0; i < arr.length; i += 2) result.push(arr.slice(i, i + 2));
  return result;
}
</script>

<template>
  <div class="pb-20 md:pb-0">
    <div
      v-if="loading && !localSession"
      class="flex items-center justify-center py-16"
    >
      <p class="text-muted-foreground">{{ $t("workoutSession.loading") }}</p>
    </div>

    <template v-else-if="localSession">
      <!-- Mobile sticky header -->
      <div
        class="md:hidden sticky top-0 z-10 -mx-4 -mt-6 px-4 py-2.5 bg-background/95 backdrop-blur-sm border-b flex items-center gap-2 mb-4 relative"
      >
        <div class="flex-1 min-w-0">
          <p class="text-sm font-semibold truncate">
            {{
              localSession.workout?.name ?? $t("workoutSession.activeWorkout")
            }}
          </p>
          <div class="flex items-center gap-1 text-xs text-muted-foreground">
            <Timer class="w-3 h-3 shrink-0" />
            <span class="tabular-nums">{{
              formatElapsed(elapsedSeconds)
            }}</span>
          </div>
        </div>

        <!-- Rest countdown donut -->
        <Transition name="rest-fade">
          <div
            v-if="restRemaining !== null"
            class="flex items-center gap-1.5 shrink-0 px-1"
          >
            <RestCountdownDonut :remaining="restRemaining" :total="restTotal" />
            <span
              class="tabular-nums text-sm font-semibold text-green-600 dark:text-green-400"
              >{{ restRemaining }}s</span
            >
          </div>
        </Transition>

        <Button
          v-if="isActive"
          type="button"
          variant="ghost"
          size="icon"
          class="size-9 shrink-0"
          :disabled="addExerciseMutation.isPending.value"
          @click="showAddExercisePicker = true"
        >
          <Plus class="w-4 h-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="size-9 shrink-0"
          @click="showWorkoutSheet = true"
        >
          <LayoutList class="w-4 h-4" />
        </Button>

        <Button
          v-if="isActive"
          type="button"
          variant="ghost"
          size="sm"
          class="shrink-0 h-9 px-3 text-green-600 hover:text-green-700 font-semibold text-xs"
          :disabled="endingWorkout || cancellingWorkout"
          @click="showFinishDialog = true"
        >
          {{ $t("workoutSession.endWorkout") }}
        </Button>
      </div>
      <!-- Overall progress bar -->
      <div class="md:hidden absolute bottom-0 left-0 right-0 h-0.5 bg-muted">
        <div
          class="h-full bg-green-500 transition-all duration-500"
          :style="{
            width: `${(() => {
              const exs = localSession?.exercises ?? [];
              const total = exs.reduce((s, e) => s + e.sets.length, 0);
              const done = exs.reduce((s, e) => s + completedSetsCount(e), 0);
              return total ? (done / total) * 100 : 0;
            })()}%`,
          }"
        />
      </div>

      <!-- Desktop header -->
      <div
        class="hidden md:flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6"
      >
        <div>
          <h1 class="text-2xl font-bold">
            {{
              localSession.workout?.name ?? $t("workoutSession.activeWorkout")
            }}
          </h1>
          <div class="flex items-center gap-3 mt-2 text-muted-foreground">
            <div class="flex items-center gap-1">
              <Timer class="w-4 h-4 shrink-0" />
              <span
                >{{ $t("workoutSession.startedAt") }}
                {{ formatElapsed(elapsedSeconds) }}</span
              >
            </div>
            <!-- Rest countdown circle -->
            <Transition name="rest-fade">
              <div
                v-if="restRemaining !== null"
                class="flex items-center gap-1.5"
              >
                <RestCountdownDonut :remaining="restRemaining" :total="restTotal" />
                <span
                  class="tabular-nums text-sm font-semibold text-green-600 dark:text-green-400"
                  >{{ restRemaining }}s</span
                >
              </div>
            </Transition>
          </div>
        </div>
        <div v-if="isActive" class="flex flex-wrap gap-2">
          <Button
            variant="outline"
            :disabled="addExerciseMutation.isPending.value"
            @click="showAddExercisePicker = true"
          >
            <Plus class="w-4 h-4 mr-2" />
            {{ $t("workoutSession.addExercise") }}
          </Button>
          <Button
            variant="destructive"
            :disabled="endingWorkout || endSessionMutation.isPending.value"
            @click="openFinishDialog()"
          >
            <StopCircle class="w-4 h-4 mr-2" />
            {{
              endingWorkout || endSessionMutation.isPending.value
                ? $t("workoutSession.ending")
                : $t("workoutSession.endWorkout")
            }}
          </Button>
          <Button
            variant="ghost"
            class="text-muted-foreground hover:text-destructive"
            :disabled="
              cancellingWorkout || cancelSessionMutation.isPending.value
            "
            @click="openFinishDialog()"
          >
            {{ $t("workoutSession.cancelWorkout") }}
          </Button>
        </div>
      </div>

      <!-- Unified finish dialog (complete / cancel) -->
      <Dialog v-model:open="showFinishDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{
              $t("workoutSession.endDialog.title")
            }}</DialogTitle>
            <DialogDescription>{{
              $t("workoutSession.endDialog.description")
            }}</DialogDescription>
          </DialogHeader>

          <!-- Incomplete sets list -->
          <div
            v-if="incompleteSets.length > 0"
            class="rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/40 px-3 py-2.5 space-y-1.5"
          >
            <p
              class="text-xs font-semibold text-amber-800 dark:text-amber-300 uppercase tracking-wide"
            >
              {{ incompleteSets.length }} incomplete
              {{ incompleteSets.length === 1 ? "set" : "sets" }}
            </p>
            <ul class="space-y-0.5">
              <li
                v-for="{ ex, setIdx } in incompleteSets.slice(0, 5)"
                :key="`${ex.id}-${setIdx}`"
                class="text-xs text-amber-700 dark:text-amber-400 flex items-center gap-1.5"
              >
                <span class="w-1 h-1 rounded-full bg-amber-400 shrink-0" />
                {{ ex.exercise?.name ?? $t("workoutHistory.unknownExercise") }}
                <span class="text-amber-500">· set {{ setIdx + 1 }}</span>
              </li>
              <li
                v-if="incompleteSets.length > 5"
                class="text-xs text-amber-500"
              >
                +{{ incompleteSets.length - 5 }} more
              </li>
            </ul>
          </div>

          <div class="flex flex-col gap-2 pt-1">
            <Button
              class="bg-green-600 hover:bg-green-700 text-white"
              :disabled="endingWorkout || cancellingWorkout"
              @click="handleEndWorkout"
            >
              {{
                endingWorkout
                  ? $t("workoutSession.ending")
                  : $t("workoutSession.completeWorkout")
              }}
            </Button>
            <Button
              variant="outline"
              :disabled="endingWorkout || cancellingWorkout"
              @click="showFinishDialog = false"
            >
              {{ $t("workoutSession.keepGoing") }}
            </Button>
            <Button
              variant="ghost"
              class="text-destructive hover:text-destructive hover:bg-destructive/10"
              :disabled="endingWorkout || cancellingWorkout"
              @click="handleCancelWorkout"
            >
              {{
                cancellingWorkout
                  ? $t("workoutSession.cancelling")
                  : $t("workoutSession.cancelWorkout")
              }}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog v-model:open="showSkippedSetsDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{
              $t("workoutSession.skippedSetsDialog.title")
            }}</DialogTitle>
            <DialogDescription>{{
              $t("workoutSession.skippedSetsDialog.description")
            }}</DialogDescription>
          </DialogHeader>
          <div class="flex flex-col gap-2 pt-2">
            <Button @click="confirmSkippedSets">
              {{ $t("workoutSession.skippedSetsDialog.confirm") }}
            </Button>
            <Button variant="outline" @click="showSkippedSetsDialog = false">
              {{ $t("cancel") }}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog v-model:open="showRemoveExerciseDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{
              $t("workoutSession.removeExerciseDialog.title")
            }}</DialogTitle>
            <DialogDescription>
              {{
                $t("workoutSession.removeExerciseDialog.description", {
                  name: exerciseToRemove?.exercise?.name ?? "",
                })
              }}
            </DialogDescription>
          </DialogHeader>
          <div class="flex flex-col gap-2 pt-2">
            <Button variant="destructive" @click="removeExercise">
              {{ $t("workoutSession.removeExerciseDialog.confirm") }}
            </Button>
            <Button variant="outline" @click="showRemoveExerciseDialog = false">
              {{ $t("cancel") }}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <ExercisePickerSheet
        v-model:open="showAddExercisePicker"
        @add="handleAddExercises"
      />

      <div
        v-if="error"
        class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg"
      >
        <p>{{ (error as Error).message }}</p>
      </div>

      <!-- Current exercise interactive panel -->
      <div
        v-if="currentEx"
        class="rounded-2xl border bg-card shadow-sm mb-4 overflow-hidden"
      >
        <!-- Top completion strip -->
        <div class="h-1 w-full bg-muted">
          <div
            class="h-full bg-green-500 transition-all duration-500"
            :style="{
              width: `${currentEx.sets.length ? (completedSetsCount(currentEx) / currentEx.sets.length) * 100 : 0}%`,
            }"
          />
        </div>

        <div class="px-5 pt-4 pb-2">
          <!-- Exercise progress dots -->
          <div class="flex justify-center gap-1.5 mb-3">
            <button
              v-for="(ex, idx) in exercises"
              :key="ex.id"
              type="button"
              class="rounded-full transition-all duration-300 cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :class="
                idx === activeExIndex
                  ? 'w-5 h-2 bg-primary'
                  : isExerciseComplete(ex)
                    ? 'w-2 h-2 bg-green-500'
                    : 'w-2 h-2 bg-muted-foreground/25'
              "
              @click="activeExIndex = idx"
            />
          </div>
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <span
                class="inline-flex items-center gap-1 text-xs font-semibold px-2 py-0.5 rounded-full mb-2"
                :class="
                  isExerciseComplete(currentEx)
                    ? 'bg-green-500/15 text-green-600 dark:text-green-400'
                    : 'bg-primary/10 text-primary'
                "
              >
                {{
                  isExerciseComplete(currentEx)
                    ? $t("workoutSession.done")
                    : `${completedSetsCount(currentEx)} / ${currentEx.sets.length} ${$t("workouts.sets")}`
                }}
              </span>
              <h2 class="text-xl font-bold leading-tight">
                {{ currentEx.exercise?.name }}
              </h2>
              <p
                v-if="currentEx.exercise?.primary_muscle_groups?.length"
                class="text-sm text-muted-foreground mt-0.5"
              >
                {{
                  currentEx.exercise.primary_muscle_groups
                    .map((m) => m.name)
                    .join(", ")
                }}
              </p>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <ExerciseInfoDialog :exercise="currentEx.exercise" />
              <Button
                v-if="currentEx.exercise?.id"
                variant="ghost"
                size="icon"
                class="size-8 text-muted-foreground"
                @click="
                  openLogs(currentEx.exercise.id, currentEx.exercise.name)
                "
              >
                <History class="size-4" />
              </Button>
              <Button
                v-if="isActive && currentEx.workout_exercise_id === 0"
                variant="ghost"
                size="icon"
                class="size-8 text-destructive hover:text-destructive"
                @click="confirmRemoveExercise(currentEx)"
              >
                <Trash2 class="size-4" />
              </Button>
            </div>
          </div>
        </div>

        <div class="space-y-1 px-2 pb-3">
          <!-- Column headers — chunked 2 per row, align with middle input column -->
          <template
            v-for="(chunk, chunkIdx) in chunk2(currentEx.sets[0]?.values ?? [])"
            :key="chunkIdx"
          >
            <div
              class="flex items-center gap-2 px-2"
              :class="chunkIdx === 0 ? 'pb-0.5' : 'pt-1 pb-0.5'"
            >
              <span class="w-8 shrink-0" />
              <div class="flex flex-1 gap-2">
                <span
                  v-for="fv in chunk"
                  :key="fv.feature_name"
                  class="flex-1 text-center text-xs font-medium text-muted-foreground uppercase tracking-wide"
                >
                  {{ $t(`exerciseFeature.${fv.feature_name}`) }}
                </span>
              </div>
              <span class="w-14 shrink-0" />
            </div>
          </template>

          <template
            v-for="(set, setIdx) in currentEx.sets"
            :key="set.id || setIdx"
          >
            <!-- Rest bar -->
            <div
              v-if="
                restExerciseId === currentEx.id &&
                restSetIdx === setIdx &&
                restRemaining !== null &&
                restRemaining > 0
              "
              class="flex items-center gap-2 px-2 py-1"
            >
              <div class="flex-1 h-1 bg-muted rounded-full overflow-hidden">
                <div
                  class="h-full bg-green-500 rounded-full transition-all duration-200"
                  :style="{ width: `${(restRemaining / restTotal) * 100}%` }"
                />
              </div>
              <span
                class="text-xs tabular-nums text-green-600 dark:text-green-400 font-medium shrink-0 w-8 text-right"
                >{{ restRemaining }}s</span
              >
            </div>

            <!-- Set row — inputs stacked in middle, check button spans full height -->
            <div
              class="flex items-stretch gap-2 px-2 py-1 rounded-xl transition-colors"
              :class="
                set.completed_at ? 'bg-green-600/10 dark:bg-green-500/10' : ''
              "
            >
              <!-- Set number / remove button -->
              <div class="w-8 shrink-0 flex items-center justify-center">
                <button
                  v-if="
                    isActive && !set.completed_at && currentEx!.sets.length > 1
                  "
                  type="button"
                  class="size-7 flex items-center justify-center rounded-md text-muted-foreground/40 hover:text-destructive hover:bg-destructive/10 active:text-destructive active:bg-destructive/10 transition-colors"
                  @click="removeSet(currentEx!, setIdx)"
                >
                  <Trash2 class="size-3.5" />
                </button>
                <span
                  v-else
                  class="text-sm font-semibold text-muted-foreground"
                  >{{ setIdx + 1 }}</span
                >
              </div>

              <!-- All input chunk rows stacked in middle column -->
              <div class="flex-1 flex flex-col gap-1.5 justify-center py-1">
                <div
                  v-for="(chunk, chunkIdx) in chunk2(set.values)"
                  :key="chunkIdx"
                  class="flex gap-2"
                >
                  <GymValueInput
                    v-for="(fv, chunkVIdx) in chunk"
                    :key="fv.feature_name"
                    :feature-name="fv.feature_name"
                    :model-value="fv.value"
                    :disabled="!isActive"
                    class="flex-1"
                    @update:model-value="
                      (val: number) =>
                        updateSetValue(
                          currentEx!,
                          set,
                          chunkIdx * 2 + chunkVIdx,
                          val,
                        )
                    "
                    @blur="save()"
                    @change="save()"
                  />
                  <div v-if="chunk.length === 1" class="flex-1" />
                </div>
              </div>

              <!-- Complete button — spans full row height for large tap target -->
              <button
                type="button"
                :disabled="
                  !isActive || isSetCheckboxDisabled(currentEx, setIdx)
                "
                class="self-stretch min-h-[3.5rem] w-14 shrink-0 rounded-xl border-2 flex items-center justify-center transition-colors disabled:opacity-30"
                :class="
                  set.completed_at
                    ? 'bg-green-600 border-green-600 text-white'
                    : 'border-border text-muted-foreground hover:border-green-500 hover:text-green-600'
                "
                @click="setCompleted(currentEx!, setIdx, !set.completed_at)"
              >
                <Check v-if="set.completed_at" class="size-5" />
                <Circle v-else class="size-5" />
              </button>
            </div>
          </template>

          <!-- Add set + rest timer -->
          <div class="flex items-center gap-2 pt-2 px-2 border-t mt-1">
            <Button
              v-if="isActive"
              type="button"
              variant="ghost"
              size="sm"
              class="text-muted-foreground h-9 px-3"
              @click="addSet(currentEx!)"
            >
              <Plus class="w-4 h-4 mr-1.5" />
              {{ $t("workoutSession.addSet") }}
            </Button>
            <div class="flex items-center gap-1 ml-auto">
              <Timer class="size-3.5 text-muted-foreground" />
              <Button
                v-if="isActive"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8"
                @click="adjustRestTimer(currentEx!, -15)"
              >
                <Minus class="size-3.5" />
              </Button>
              <span class="text-sm tabular-nums w-10 text-center"
                >{{
                  currentEx.rest_timer > 0
                    ? currentEx.rest_timer
                    : defaultRestSeconds
                }}s</span
              >
              <Button
                v-if="isActive"
                type="button"
                variant="ghost"
                size="icon"
                class="size-8"
                @click="adjustRestTimer(currentEx!, 15)"
              >
                <Plus class="size-3.5" />
              </Button>
            </div>
          </div>

          <!-- Note -->
          <div class="pt-1 px-2">
            <Textarea
              :model-value="currentEx.note"
              :placeholder="$t('workouts.notePlaceholder')"
              rows="2"
              class="text-sm resize-none"
              :disabled="!isActive"
              @update:model-value="
                (val: string | number) =>
                  updateNote(currentEx!, typeof val === 'string' ? val : '')
              "
              @blur="save()"
            />
          </div>
        </div>
      </div>

      <!-- Prev / Next nav cards -->
      <div
        v-if="canGoPrev || canGoNext"
        class="grid gap-3 mb-6"
        :class="canGoPrev && canGoNext ? 'grid-cols-2' : 'grid-cols-1'"
      >
        <!-- Previous -->
        <button
          v-if="canGoPrev && prevEx"
          class="group text-left rounded-xl border bg-card p-4 hover:bg-muted/50 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          @click="goToPrev"
        >
          <div class="flex items-center gap-1 mb-2 text-muted-foreground">
            <ChevronLeft class="size-3.5" />
            <span class="text-xs font-medium uppercase tracking-wide">{{
              $t("workoutSession.previous")
            }}</span>
            <Check
              v-if="isExerciseComplete(prevEx)"
              class="size-3 ml-auto text-green-500"
            />
          </div>
          <p class="font-semibold text-sm leading-snug line-clamp-2">
            {{ prevEx.exercise?.name }}
          </p>
          <p
            v-if="prevEx.exercise?.primary_muscle_groups?.length"
            class="text-xs text-muted-foreground mt-1 truncate"
          >
            {{
              prevEx.exercise.primary_muscle_groups
                .map((m) => m.name)
                .join(", ")
            }}
          </p>
          <p class="text-xs text-muted-foreground mt-1.5">
            {{
              isExerciseComplete(prevEx)
                ? $t("workoutSession.setsCompleted", {
                    count: prevEx.sets.length,
                  })
                : `${completedSetsCount(prevEx)} / ${prevEx.sets.length} ${$t("workouts.sets")}`
            }}
          </p>
        </button>

        <!-- Next -->
        <button
          v-if="canGoNext && nextEx"
          class="group text-left rounded-xl border bg-card p-4 hover:bg-muted/50 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          @click="goToNext"
        >
          <div class="flex items-center gap-1 mb-2 text-muted-foreground">
            <span class="text-xs font-medium uppercase tracking-wide">{{
              $t("workoutSession.upNext")
            }}</span>
            <ChevronRight class="size-3.5 ml-auto" />
          </div>
          <p class="font-semibold text-sm leading-snug line-clamp-1">
            {{ nextEx.exercise?.name }}
          </p>
          <p
            v-if="nextEx.exercise?.primary_muscle_groups?.length"
            class="text-xs text-muted-foreground mt-0.5 truncate"
          >
            {{
              nextEx.exercise.primary_muscle_groups
                .map((m) => m.name)
                .join(", ")
            }}
          </p>
          <!-- Compact set + rest summary -->
          <div class="flex flex-wrap items-center gap-1.5 mt-2.5">
            <span
              class="text-xs font-medium bg-muted px-2 py-0.5 rounded-full tabular-nums"
            >
              {{ nextEx.sets.length }} {{ $t("workouts.sets") }}
            </span>
            <template
              v-for="fv in nextEx.sets[0]?.values ?? []"
              :key="fv.feature_name"
            >
              <span
                v-if="fv.value"
                class="text-xs text-muted-foreground tabular-nums"
              >
                {{ fv.value }} {{ $t(`exerciseFeature.${fv.feature_name}`) }}
              </span>
            </template>
            <span
              class="text-xs text-muted-foreground flex items-center gap-0.5 ml-auto"
            >
              <Timer class="size-3 shrink-0" />
              {{
                nextEx.rest_timer > 0 ? nextEx.rest_timer : defaultRestSeconds
              }}s
            </span>
          </div>
        </button>
      </div>

      <!-- Full workout sheet trigger (desktop only — mobile has it in sticky header) -->
      <Button
        variant="ghost"
        size="sm"
        class="hidden md:flex w-full text-muted-foreground mb-6"
        @click="showWorkoutSheet = true"
      >
        <LayoutList class="size-4 mr-2" />
        {{ $t("workoutSession.viewFullWorkout") }}
      </Button>

      <!-- Workout overview sheet -->
      <Sheet v-model:open="showWorkoutSheet">
        <SheetContent class="flex flex-col gap-0 p-0 sm:max-w-sm">
          <SheetHeader class="px-6 py-5 border-b">
            <SheetTitle>{{ $t("workoutSession.previewTitle") }}</SheetTitle>
            <SheetDescription>{{
              localSession.workout?.name ?? $t("workoutSession.activeWorkout")
            }}</SheetDescription>
          </SheetHeader>
          <div class="flex-1 overflow-y-auto">
            <button
              v-for="(ex, idx) in localSession.exercises"
              :key="ex.id"
              class="w-full flex items-start gap-4 px-6 py-4 border-b last:border-0 hover:bg-muted/40 transition-colors text-left"
              :class="{ 'bg-primary/5': idx === activeExIndex }"
              @click="jumpToExercise(idx)"
            >
              <div
                class="size-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0 mt-0.5"
                :class="
                  isExerciseComplete(ex)
                    ? 'bg-green-600/15 text-green-600 dark:text-green-400'
                    : idx === activeExIndex
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground'
                "
              >
                <Check v-if="isExerciseComplete(ex)" class="size-3.5" />
                <span v-else>{{ idx + 1 }}</span>
              </div>
              <div class="min-w-0 flex-1">
                <p
                  class="font-semibold text-sm"
                  :class="{
                    'text-primary':
                      idx === activeExIndex && !isExerciseComplete(ex),
                  }"
                >
                  {{ ex.exercise?.name }}
                </p>
                <p
                  v-if="ex.exercise?.primary_muscle_groups?.length"
                  class="text-xs text-muted-foreground mt-0.5"
                >
                  {{
                    ex.exercise.primary_muscle_groups
                      .map((m) => m.name)
                      .join(", ")
                  }}
                </p>
                <div class="mt-1.5 flex items-center gap-3">
                  <span class="text-xs text-muted-foreground">
                    {{
                      isExerciseComplete(ex)
                        ? $t("workoutSession.setsCompleted", {
                            count: ex.sets.length,
                          })
                        : `${completedSetsCount(ex)} / ${ex.sets.length} ${$t("workouts.sets")}`
                    }}
                  </span>
                  <span
                    v-if="!isExerciseComplete(ex) && completedSetsCount(ex) > 0"
                    class="text-xs text-amber-500 font-medium"
                  >
                    {{ $t("workoutSession.inProgress") }}
                  </span>
                </div>
              </div>
              <ChevronRight
                class="size-4 text-muted-foreground shrink-0 mt-1"
              />
            </button>
          </div>
        </SheetContent>
      </Sheet>

      <!-- Desktop finish buttons -->
      <div
        v-if="isActive"
        class="hidden md:flex flex-col gap-2 mt-4 pt-4 border-t"
      >
        <Button
          class="w-full bg-green-600 hover:bg-green-700 text-white"
          :disabled="endingWorkout || endSessionMutation.isPending.value"
          @click="openFinishDialog()"
        >
          {{
            endingWorkout || endSessionMutation.isPending.value
              ? $t("workoutSession.ending")
              : $t("workoutSession.completeWorkout")
          }}
        </Button>
        <Button
          variant="ghost"
          class="w-full text-muted-foreground hover:text-destructive"
          :disabled="cancellingWorkout || cancelSessionMutation.isPending.value"
          @click="openFinishDialog()"
        >
          {{
            cancellingWorkout
              ? $t("workoutSession.cancelling")
              : $t("workoutSession.cancelWorkout")
          }}
        </Button>
      </div>

      <!-- Mobile sticky finish bar -->
      <Teleport to="body">
        <div
          v-if="isActive"
          class="md:hidden fixed bottom-0 left-0 right-0 z-20 bg-background/95 backdrop-blur-sm border-t px-4 py-3"
        >
          <Button
            class="w-full h-12 text-base font-semibold bg-green-600 hover:bg-green-700 text-white"
            :disabled="endingWorkout || endSessionMutation.isPending.value"
            @click="openFinishDialog()"
          >
            {{
              endingWorkout || endSessionMutation.isPending.value
                ? $t("workoutSession.ending")
                : $t("workoutSession.completeWorkout")
            }}
          </Button>
        </div>
      </Teleport>
    </template>
  </div>

  <ExerciseLogDrawer
    v-model:open="logDrawerOpen"
    :exercise-id="logExerciseId"
    :exercise-name="logExerciseName"
  />
</template>

<style scoped>
.rest-fade-enter-active,
.rest-fade-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s ease;
}
.rest-fade-enter-from,
.rest-fade-leave-to {
  opacity: 0;
  transform: scale(0.8);
}
</style>
