<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { useActiveWorkoutSession } from "@/features/workout-session/composables/useActiveWorkoutSession";
import { useUpdateWorkoutSession } from "@/features/workout-session/composables/useUpdateWorkoutSession";
import { useEndWorkoutSession } from "@/features/workout-session/composables/useEndWorkoutSession";
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
import { Plus, Timer, StopCircle, Check, Circle } from "lucide-vue-next";
import { useI18n } from "vue-i18n";
import { toast } from "vue-sonner";

const router = useRouter();
const { t } = useI18n();
const { session, loading, error, refetch } = useActiveWorkoutSession();
const endingWorkout = ref(false);

const localSession = ref<WorkoutSession | null>(null);
const restRemaining = ref<number | null>(null);
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

function startRestTimer(ex: WorkoutSessionExercise) {
  const restSeconds = ex.rest_timer > 0 ? ex.rest_timer : defaultRestSeconds;
  restRemaining.value = restSeconds;
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
    startRestTimer(ex);
  } else {
    if (restTimerId) clearInterval(restTimerId);
    restTimerId = null;
    restRemaining.value = null;
  }

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

async function handleEndWorkout() {
  if (sessionId.value === 0) return;
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
</script>

<template>
  <div class="p-8 max-w-[900px] mx-auto">
    <div v-if="loading && !localSession" class="flex items-center justify-center py-16">
      <p class="text-muted-foreground">{{ $t("workoutSession.loading") }}</p>
    </div>

    <template v-else-if="localSession">
      <div class="mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold">
            {{ localSession.workout?.name ?? $t("workoutSession.activeWorkout") }}
          </h1>
          <div class="flex flex-wrap items-center gap-6 mt-2 text-muted-foreground">
            <div class="flex items-center gap-2">
              <Timer class="w-4 h-4 shrink-0" />
              <span>{{ $t("workoutSession.startedAt") }} {{ formatElapsed(elapsedSeconds) }}</span>
            </div>
            <div
              class="flex items-center gap-2 min-w-0"
              :class="restRemaining !== null && restRemaining > 0 ? 'text-green-600 dark:text-green-400 font-medium' : ''"
            >
              <Timer class="w-4 h-4 shrink-0" />
              <span>
                {{ $t("workoutSession.rest") }}:
                <template v-if="restRemaining !== null && restRemaining > 0">
                  {{ restRemaining }}s
                  <span v-if="restLabel" class="text-muted-foreground font-normal">· {{ restLabel }}</span>
                </template>
                <template v-else>—</template>
              </span>
            </div>
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
            @click="handleEndWorkout"
          >
            <StopCircle class="w-4 h-4 mr-2" />
            {{ endingWorkout || endSessionMutation.isPending.value ? $t("workoutSession.ending") : $t("workoutSession.endWorkout") }}
          </Button>
        </div>
      </div>

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

      <div class="space-y-6">
        <Card v-for="ex in localSession.exercises" :key="ex.id">
          <CardHeader>
            <CardTitle class="text-lg">{{ ex.exercise?.name ?? "" }}</CardTitle>
            <p v-if="ex.exercise?.primary_muscle_groups?.length" class="text-sm text-muted-foreground">
              {{ ex.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
            </p>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <div
                v-for="(set, setIdx) in ex.sets"
                :key="set.id || setIdx"
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
                      @update:model-value="(val: number) => updateSetValue(ex, set, vIdx, val)"
                      @blur="save()"
                    />
                  </template>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  :disabled="!isActive || isSetCheckboxDisabled(ex, setIdx)"
                  :class="[
                    'shrink-0 size-9',
                    set.completed_at
                      ? 'bg-green-600 border-green-600 text-white hover:bg-green-700 hover:text-white'
                      : '',
                  ]"
                  @click="setCompleted(ex, setIdx, !set.completed_at)"
                >
                  <Check v-if="set.completed_at" class="size-5" />
                  <Circle v-else class="size-5" />
                </Button>
              </div>
            </div>
            <Button
              v-if="isActive"
              type="button"
              variant="outline"
              size="sm"
              @click="addSet(ex)"
            >
              <Plus class="w-4 h-4 mr-2" />
              {{ $t("workoutSession.addSet") }}
            </Button>
            <div class="pt-2">
              <Label class="text-sm text-muted-foreground">{{ $t("workouts.note") }}</Label>
              <Textarea
                class="mt-1"
                :model-value="ex.note"
                :placeholder="$t('workouts.notePlaceholder')"
                rows="2"
                :disabled="!isActive"
                @update:model-value="(val: string | number) => updateNote(ex, typeof val === 'string' ? val : '')"
                @blur="save()"
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </template>
  </div>
</template>
