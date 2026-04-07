<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useActiveWorkoutSession } from "@/features/workout-session/composables/useActiveWorkoutSession";
import { useUpdateWorkoutSession } from "@/features/workout-session/composables/useUpdateWorkoutSession";
import { useEndWorkoutSession } from "@/features/workout-session/composables/useEndWorkoutSession";
import { useCancelWorkoutSession } from "@/features/workout-session/composables/useCancelWorkoutSession";
import { useAddExerciseToSession } from "@/features/workout-session/composables/useAddExerciseToSession";
import { useExercises } from "@/features/exercises/composables/useExercises";
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
import Card from "@/components/ui/card/Card.vue";
import CardHeader from "@/components/ui/card/CardHeader.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
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
import { Plus, Minus, Timer, StopCircle, Check, Circle, History, ChevronLeft, ChevronRight, LayoutList } from "lucide-vue-next";
import { useI18n } from "vue-i18n";
import { toast } from "vue-sonner";
import ExerciseLogDrawer from "@/features/exercises/components/ExerciseLogDrawer.vue";
import ExerciseInfoDialog from "@/features/exercises/components/ExerciseInfoDialog.vue";

const router = useRouter();
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
let elapsedTimerId: ReturnType<typeof setInterval> | null = null;

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
const isActive = computed(() => !!localSession.value && !localSession.value.ended_at);
const endSessionMutation = useEndWorkoutSession(sessionId);
const cancelSessionMutation = useCancelWorkoutSession(sessionId);
const updateSessionMutation = useUpdateWorkoutSession(sessionId);
const addExerciseMutation = useAddExerciseToSession(sessionId);

const showAddExerciseDialog = ref(false);
const { exercises: libraryExercises, loading: exercisesLoading } = useExercises(
  computed(() => ({ limit: 200 })),
);

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
    showAddExerciseDialog.value = false;
  } catch (err) {
    console.error("Failed to add exercise:", err);
    toast.error(t("workoutSession.toasts.saveFailed"));
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

function startRestTimer(ex: WorkoutSessionExercise, setIdx: number) {
  const restSeconds = ex.rest_timer > 0 ? ex.rest_timer : defaultRestSeconds;
  restRemaining.value = restSeconds;
  restTotal.value = restSeconds;
  restLabel.value = ex.exercise?.name ?? "";
  if (restTimerId) clearInterval(restTimerId);
  restTimerId = setInterval(() => {
    if (restRemaining.value === null) return;
    restRemaining.value -= 1;
    if (restRemaining.value <= 0 && restTimerId) {
      clearInterval(restTimerId);
      restTimerId = null;
      restRemaining.value = null;
    }
  }, 1000);

  // Point to the target set — the one we're resting *for*
  const exs = localSession.value?.exercises ?? [];
  const exIdx = exs.findIndex((e) => e.id === ex.id);
  if (setIdx < ex.sets.length - 1) {
    // Next set in the same exercise
    restExerciseId.value = ex.id;
    restSetIdx.value = setIdx + 1;
  } else if (exIdx >= 0 && exIdx < exs.length - 1) {
    // Last set of this exercise — rest is for the first set of the next exercise
    const nextExercise = exs[exIdx + 1];
    restExerciseId.value = nextExercise?.id ?? null;
    restSetIdx.value = 0;
  } else {
    // Last set of the last exercise — nothing upcoming
    restExerciseId.value = null;
    restSetIdx.value = null;
  }
}

async function setCompleted(
  ex: WorkoutSessionExercise,
  setIdx: number,
  checked: boolean,
) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const sets = localSession.value.exercises[exIndex]?.sets;
  if (!sets || setIdx < 0 || setIdx >= sets.length) return;
  const s = sets[setIdx];
  if (!s) return;

  s.completed_at = checked ? new Date().toISOString() : null;

  if (checked) {
    startRestTimer(ex, setIdx);
  } else {
    if (restTimerId) clearInterval(restTimerId);
    restTimerId = null;
    restRemaining.value = null;
    restExerciseId.value = null;
    restSetIdx.value = null;
  }

  const viewedExIndex = activeExIndex.value;

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

  // Auto-advance only if the user is still viewing the exercise that just became fully complete
  if (checked && viewedExIndex === activeExIndex.value) {
    const ex = localSession.value?.exercises[viewedExIndex];
    if (ex && isExerciseComplete(ex) && viewedExIndex < (localSession.value?.exercises.length ?? 1) - 1) {
      setTimeout(() => { activeExIndex.value++; }, 400);
    }
  }
}

function isSetCheckboxDisabled(ex: WorkoutSessionExercise, setIdx: number): boolean {
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

function adjustRestTimer(ex: WorkoutSessionExercise, delta: number) {
  if (!localSession.value) return;
  const exIndex = localSession.value.exercises.findIndex((e) => e.id === ex.id);
  if (exIndex === -1) return;
  const exItem = localSession.value.exercises[exIndex];
  if (!exItem) return;
  const current = exItem.rest_timer > 0 ? exItem.rest_timer : defaultRestSeconds;
  exItem.rest_timer = Math.max(15, current + delta);
  save();
}

const elapsedSeconds = ref(0);
watch(
  () => localSession.value?.started_at,
  (startedAt) => {
    if (elapsedTimerId) clearInterval(elapsedTimerId);
    elapsedTimerId = null;
    if (!startedAt) return;
    const start = new Date(startedAt).getTime();
    const tick = () => {
      elapsedSeconds.value = Math.floor((Date.now() - start) / 1000);
    };
    tick();
    elapsedTimerId = setInterval(tick, 1000);
  },
  { immediate: true },
);

function formatElapsed(sec: number): string {
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  if (h > 0) return `${h}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

const showEndDialog = ref(false);
const showCancelDialog = ref(false);
const cancellingWorkout = ref(false);

async function handleCancelWorkout() {
  if (sessionId.value === 0) return;
  showCancelDialog.value = false;
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
  showEndDialog.value = false;
  endingWorkout.value = true;
  try {
    await endSessionMutation.endSession();
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

onUnmounted(() => {
  if (restTimerId) clearInterval(restTimerId);
  if (elapsedTimerId) clearInterval(elapsedTimerId);
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
    const firstIncomplete = s.exercises.findIndex((ex) => !isExerciseComplete(ex));
    activeExIndex.value = firstIncomplete >= 0 ? firstIncomplete : 0;
  },
  { once: true },
);

const exercises = computed(() => localSession.value?.exercises ?? []);

const currentEx = computed(() =>
  exercises.value[activeExIndex.value] ?? null,
);

const nextEx = computed(() => {
  const next = activeExIndex.value + 1;
  return next < exercises.value.length ? (exercises.value[next] ?? null) : null;
});

const prevEx = computed(() => {
  const prev = activeExIndex.value - 1;
  return prev >= 0 ? (exercises.value[prev] ?? null) : null;
});

const canGoPrev = computed(() => activeExIndex.value > 0);
const canGoNext = computed(() => activeExIndex.value < exercises.value.length - 1);

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
    showEndDialog.value = true;
  }
});

const showWorkoutSheet = ref(false);

</script>

<template>
  <div>
    <div v-if="loading && !localSession" class="flex items-center justify-center py-16">
      <p class="text-muted-foreground">{{ $t("workoutSession.loading") }}</p>
    </div>

    <template v-else-if="localSession">
      <div class="mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold">
            {{ localSession.workout?.name ?? $t("workoutSession.activeWorkout") }}
          </h1>
          <div class="flex items-center gap-2 mt-2 text-muted-foreground">
            <Timer class="w-4 h-4 shrink-0" />
            <span>{{ $t("workoutSession.startedAt") }} {{ formatElapsed(elapsedSeconds) }}</span>
          </div>
        </div>
        <div v-if="isActive" class="flex flex-wrap gap-2">
          <Button
            variant="outline"
            :disabled="addExerciseMutation.isPending.value"
            @click="showAddExerciseDialog = true"
          >
            <Plus class="w-4 h-4 mr-2" />
            {{ $t("workoutSession.addExercise") }}
          </Button>
          <Button
            variant="destructive"
            :disabled="endingWorkout || endSessionMutation.isPending.value"
            @click="showEndDialog = true"
          >
            <StopCircle class="w-4 h-4 mr-2" />
            {{ endingWorkout || endSessionMutation.isPending.value ? $t("workoutSession.ending") : $t("workoutSession.endWorkout") }}
          </Button>
          <Button
            variant="ghost"
            class="text-muted-foreground hover:text-destructive"
            :disabled="cancellingWorkout || cancelSessionMutation.isPending.value"
            @click="showCancelDialog = true"
          >
            {{ $t("workoutSession.cancelWorkout") }}
          </Button>
        </div>
      </div>

      <Dialog v-model:open="showCancelDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutSession.cancelDialog.title") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutSession.cancelDialog.description") }}</DialogDescription>
          </DialogHeader>
          <div class="flex flex-col gap-2 pt-2">
            <Button
              variant="destructive"
              :disabled="cancellingWorkout"
              @click="handleCancelWorkout"
            >
              {{ cancellingWorkout ? $t("workoutSession.cancelling") : $t("workoutSession.cancelWorkout") }}
            </Button>
            <Button variant="outline" :disabled="cancellingWorkout" @click="showCancelDialog = false">
              {{ $t("workoutSession.keepGoing") }}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog v-model:open="showEndDialog">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutSession.endDialog.title") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutSession.endDialog.description") }}</DialogDescription>
          </DialogHeader>
          <div class="flex flex-col gap-2 pt-2">
            <Button
              class="bg-green-600 hover:bg-green-700 text-white"
              :disabled="endingWorkout"
              @click="handleEndWorkout"
            >
              {{ endingWorkout ? $t("workoutSession.ending") : $t("workoutSession.completeWorkout") }}
            </Button>
            <Button variant="outline" :disabled="endingWorkout" @click="showEndDialog = false">
              {{ $t("cancel") }}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog v-model:open="showAddExerciseDialog">
        <DialogContent class="max-w-md max-h-[80vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutSession.addExercise") }}</DialogTitle>
            <DialogDescription>
              {{ $t("workoutSession.addExerciseDescription") }}
            </DialogDescription>
          </DialogHeader>
          <div class="flex-1 overflow-y-auto -mx-6 px-6">
            <p v-if="exercisesLoading" class="text-muted-foreground text-sm py-4">{{ $t("workoutSession.loading") }}</p>
            <ul v-else class="space-y-1">
              <li
                v-for="ex in libraryExercises"
                :key="ex.id"
                class="flex items-center justify-between gap-2 py-2 px-3 rounded-md hover:bg-muted cursor-pointer"
                @click="onSelectExercise(ex.id)"
              >
                <span class="font-medium">{{ ex.name }}</span>
                <span v-if="ex.primary_muscle_groups?.length" class="text-sm text-muted-foreground truncate">
                  {{ ex.primary_muscle_groups.map((m) => m.name).join(", ") }}
                </span>
              </li>
              <li v-if="!exercisesLoading && libraryExercises.length === 0" class="text-muted-foreground text-sm py-4">
                {{ $t("exercises.noExercises") }}
              </li>
            </ul>
          </div>
        </DialogContent>
      </Dialog>

      <div v-if="error" class="mb-4 p-4 bg-destructive/10 text-destructive rounded-lg">
        <p>{{ (error as Error).message }}</p>
      </div>

      <!-- Current exercise interactive panel -->
      <Card v-if="currentEx" class="mb-3 ring-1 ring-primary/25">
        <CardHeader class="pb-2">
          <!-- Segmented exercise progress bar -->
          <div class="flex gap-1 mb-3">
            <div
              v-for="(ex, idx) in exercises"
              :key="ex.id"
              class="h-1.5 flex-1 rounded-full transition-colors duration-300"
              :class="idx === activeExIndex
                ? 'bg-primary'
                : isExerciseComplete(ex)
                  ? 'bg-green-500'
                  : 'bg-muted'"
            />
          </div>
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <span class="text-xs font-semibold uppercase tracking-wide text-primary block mb-1.5">
                {{ isExerciseComplete(currentEx) ? $t("workoutSession.done") : $t("workoutSession.now") }}
              </span>
              <CardTitle class="text-lg">{{ currentEx.exercise?.name }}</CardTitle>
              <p v-if="currentEx.exercise?.primary_muscle_groups?.length" class="text-sm text-muted-foreground">
                {{ currentEx.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
              </p>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <ExerciseInfoDialog :exercise="currentEx.exercise" />
              <Button
                v-if="currentEx.exercise?.id"
                variant="ghost"
                size="icon"
                class="size-8 text-muted-foreground"
                @click="openLogs(currentEx.exercise.id, currentEx.exercise.name)"
              >
                <History class="size-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="space-y-2">
            <template v-for="(set, setIdx) in currentEx.sets" :key="set.id || setIdx">
              <!-- Rest bar above this set -->
              <div
                v-if="restExerciseId === currentEx.id && restSetIdx === setIdx && restRemaining !== null && restRemaining > 0"
                class="flex items-center gap-2 px-3"
              >
                <div class="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
                  <div
                    class="h-full bg-green-500 rounded-full"
                    :style="{ width: `${(restRemaining / restTotal) * 100}%`, transition: 'width 0.2s ease-out' }"
                  />
                </div>
                <span class="text-xs tabular-nums text-green-600 dark:text-green-400 font-medium shrink-0">{{ restRemaining }}s</span>
              </div>
              <div
                class="flex flex-wrap items-center gap-3 py-2 px-3 rounded-md border-b border-border/50 last:border-0 transition-colors"
                :class="{ 'bg-green-600/15 dark:bg-green-500/15': set.completed_at }"
              >
                <span class="text-sm text-muted-foreground w-6">{{ setIdx + 1 }}.</span>
                <div class="flex flex-wrap gap-3 flex-1 items-end">
                  <template v-for="(fv, vIdx) in set.values" :key="fv.feature_name">
                    <GymValueInput
                      :feature-name="fv.feature_name"
                      :model-value="fv.value"
                      :label="$t(`exerciseFeature.${fv.feature_name}`)"
                      :disabled="!isActive"
                      @update:model-value="(val: number) => updateSetValue(currentEx!, set, vIdx, val)"
                      @blur="save()"
                      @change="save()"
                    />
                  </template>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  :disabled="!isActive || isSetCheckboxDisabled(currentEx, setIdx)"
                  :class="[
                    'shrink-0 size-9',
                    set.completed_at
                      ? 'bg-green-600 border-green-600 text-white hover:bg-green-700 hover:text-white'
                      : '',
                  ]"
                  @click="setCompleted(currentEx!, setIdx, !set.completed_at)"
                >
                  <Check v-if="set.completed_at" class="size-5" />
                  <Circle v-else class="size-5" />
                </Button>
              </div>
            </template>
          </div>

          <div class="flex items-center gap-2">
            <Button v-if="isActive" type="button" variant="outline" size="sm" @click="addSet(currentEx!)">
              <Plus class="w-4 h-4 mr-2" />
              {{ $t("workoutSession.addSet") }}
            </Button>
            <div class="flex items-center gap-1 ml-auto">
              <Timer class="size-4 text-muted-foreground" />
              <Button v-if="isActive" type="button" variant="ghost" size="icon" class="size-7" @click="adjustRestTimer(currentEx!, -15)">
                <Minus class="size-3" />
              </Button>
              <span class="text-sm tabular-nums w-10 text-center">{{ currentEx.rest_timer > 0 ? currentEx.rest_timer : defaultRestSeconds }}s</span>
              <Button v-if="isActive" type="button" variant="ghost" size="icon" class="size-7" @click="adjustRestTimer(currentEx!, 15)">
                <Plus class="size-3" />
              </Button>
            </div>
          </div>

          <div class="pt-2">
            <Label class="text-sm text-muted-foreground">{{ $t("workouts.note") }}</Label>
            <Textarea
              class="mt-1"
              :model-value="currentEx.note"
              :placeholder="$t('workouts.notePlaceholder')"
              rows="2"
              :disabled="!isActive"
              @update:model-value="(val: string | number) => updateNote(currentEx!, typeof val === 'string' ? val : '')"
              @blur="save()"
            />
          </div>
        </CardContent>
      </Card>

      <!-- Prev / Next nav cards -->
      <div v-if="canGoPrev || canGoNext" class="grid gap-3 mb-6" :class="canGoPrev && canGoNext ? 'grid-cols-2' : 'grid-cols-1'">
        <!-- Previous -->
        <button
          v-if="canGoPrev && prevEx"
          class="group text-left rounded-xl border bg-card p-4 hover:bg-muted/50 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          @click="goToPrev"
        >
          <div class="flex items-center gap-1 mb-2 text-muted-foreground">
            <ChevronLeft class="size-3.5" />
            <span class="text-xs font-medium uppercase tracking-wide">{{ $t("workoutSession.previous") }}</span>
            <Check v-if="isExerciseComplete(prevEx)" class="size-3 ml-auto text-green-500" />
          </div>
          <p class="font-semibold text-sm leading-snug line-clamp-2">{{ prevEx.exercise?.name }}</p>
          <p v-if="prevEx.exercise?.primary_muscle_groups?.length" class="text-xs text-muted-foreground mt-1 truncate">
            {{ prevEx.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
          </p>
          <p class="text-xs text-muted-foreground mt-1.5">
            {{ isExerciseComplete(prevEx)
              ? $t("workoutSession.setsCompleted", { count: prevEx.sets.length })
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
            <span class="text-xs font-medium uppercase tracking-wide">{{ $t("workoutSession.upNext") }}</span>
            <ChevronRight class="size-3.5 ml-auto" />
          </div>
          <p class="font-semibold text-sm leading-snug line-clamp-2">{{ nextEx.exercise?.name }}</p>
          <p v-if="nextEx.exercise?.primary_muscle_groups?.length" class="text-xs text-muted-foreground mt-1 truncate">
            {{ nextEx.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
          </p>
          <!-- Set preview table -->
          <table v-if="nextEx.sets.length" class="w-full mt-3 text-xs">
            <thead>
              <tr class="text-muted-foreground">
                <th class="text-left font-medium pb-1 w-8">{{ $t("exercises.logs.set") }}</th>
                <th
                  v-for="fv in nextEx.sets[0]?.values ?? []"
                  :key="fv.feature_name"
                  class="text-right font-medium pb-1"
                >
                  {{ $t(`exerciseFeature.${fv.feature_name}`) }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <template v-for="(set, setIdx) in nextEx.sets.slice(0, 3)" :key="set.id || setIdx">
                <!-- Rest bar spanning full row -->
                <tr v-if="restExerciseId === nextEx.id && restSetIdx === setIdx && restRemaining !== null && restRemaining > 0">
                  <td :colspan="1 + (set.values.length)" class="py-1">
                    <div class="flex items-center gap-2">
                      <div class="flex-1 h-1.5 bg-muted rounded-full overflow-hidden">
                        <div
                          class="h-full bg-green-500 rounded-full"
                          :style="{ width: `${(restRemaining / restTotal) * 100}%`, transition: 'width 0.2s ease-out' }"
                        />
                      </div>
                      <span class="tabular-nums text-green-600 dark:text-green-400 font-medium shrink-0">{{ restRemaining }}s</span>
                    </div>
                  </td>
                </tr>
                <tr>
                  <td class="py-1.5 text-muted-foreground font-medium">{{ setIdx + 1 }}</td>
                  <td
                    v-for="fv in set.values"
                    :key="fv.feature_name"
                    class="py-1.5 text-right tabular-nums font-medium text-foreground"
                  >
                    {{ fv.value }}
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
          <p v-if="nextEx.sets.length > 3" class="text-xs text-muted-foreground mt-1">
            +{{ nextEx.sets.length - 3 }} more sets
          </p>
          <div class="flex items-center gap-1 mt-2 text-xs text-muted-foreground">
            <Timer class="size-3" />
            <span>{{ nextEx.rest_timer > 0 ? nextEx.rest_timer : defaultRestSeconds }}s rest</span>
          </div>
          <div class="flex items-center justify-end gap-1 mt-2" @click.stop>
            <ExerciseInfoDialog :exercise="nextEx.exercise" />
            <Button
              v-if="nextEx.exercise?.id"
              variant="ghost"
              size="icon"
              class="size-7 text-muted-foreground"
              @click="openLogs(nextEx.exercise.id, nextEx.exercise.name)"
            >
              <History class="size-3.5" />
            </Button>
          </div>
        </button>
      </div>

      <!-- Full workout sheet trigger -->
      <Button variant="ghost" size="sm" class="w-full text-muted-foreground mb-6" @click="showWorkoutSheet = true">
        <LayoutList class="size-4 mr-2" />
        {{ $t("workoutSession.viewFullWorkout") }}
      </Button>

      <!-- Workout overview sheet -->
      <Sheet v-model:open="showWorkoutSheet">
        <SheetContent class="flex flex-col gap-0 p-0 sm:max-w-sm">
          <SheetHeader class="px-6 py-5 border-b">
            <SheetTitle>{{ $t("workoutSession.previewTitle") }}</SheetTitle>
            <SheetDescription>{{ localSession.workout?.name ?? $t("workoutSession.activeWorkout") }}</SheetDescription>
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
                :class="isExerciseComplete(ex)
                  ? 'bg-green-600/15 text-green-600 dark:text-green-400'
                  : idx === activeExIndex
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground'"
              >
                <Check v-if="isExerciseComplete(ex)" class="size-3.5" />
                <span v-else>{{ idx + 1 }}</span>
              </div>
              <div class="min-w-0 flex-1">
                <p class="font-semibold text-sm" :class="{ 'text-primary': idx === activeExIndex && !isExerciseComplete(ex) }">
                  {{ ex.exercise?.name }}
                </p>
                <p v-if="ex.exercise?.primary_muscle_groups?.length" class="text-xs text-muted-foreground mt-0.5">
                  {{ ex.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
                </p>
                <div class="mt-1.5 flex items-center gap-3">
                  <span class="text-xs text-muted-foreground">
                    {{ isExerciseComplete(ex)
                      ? $t("workoutSession.setsCompleted", { count: ex.sets.length })
                      : `${completedSetsCount(ex)} / ${ex.sets.length} ${$t("workouts.sets")}`
                    }}
                  </span>
                  <span v-if="!isExerciseComplete(ex) && completedSetsCount(ex) > 0" class="text-xs text-amber-500 font-medium">
                    {{ $t("workoutSession.inProgress") }}
                  </span>
                </div>
              </div>
              <ChevronRight class="size-4 text-muted-foreground shrink-0 mt-1" />
            </button>
          </div>
        </SheetContent>
      </Sheet>

      <div v-if="isActive" class="mt-2 flex flex-col gap-2">
        <Button
          class="w-full bg-green-600 hover:bg-green-700 text-white"
          :disabled="endingWorkout || endSessionMutation.isPending.value"
          @click="showEndDialog = true"
        >
          <StopCircle class="w-4 h-4 mr-2" />
          {{ endingWorkout || endSessionMutation.isPending.value ? $t("workoutSession.ending") : $t("workoutSession.completeWorkout") }}
        </Button>
        <Button
          variant="ghost"
          class="w-full text-muted-foreground hover:text-destructive"
          :disabled="cancellingWorkout || cancelSessionMutation.isPending.value"
          @click="showCancelDialog = true"
        >
          {{ $t("workoutSession.cancelWorkout") }}
        </Button>
      </div>
    </template>
  </div>

  <ExerciseLogDrawer
    v-model:open="logDrawerOpen"
    :exercise-id="logExerciseId"
    :exercise-name="logExerciseName"
  />
</template>
