<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { DumbbellIcon, CalendarDaysIcon, ArrowRightIcon, LoaderCircleIcon } from "lucide-vue-next";
import { Button } from "@/components/ui/button";
import { useExerciseResolver } from "@/features/ai-chat/composables/useExerciseResolver";
import { useAIFormState } from "@/features/ai-chat/composables/useAIFormState";
import ArtifactExerciseList from "@/features/ai-chat/components/ArtifactExerciseList.vue";
import type { ActiveArtifact, WorkoutArtifact, WorkoutPlanArtifact, PlanDayArtifact } from "@/features/ai-chat/types";

const IMAGE_BASE_URL = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/";

const props = defineProps<{
  artifact: ActiveArtifact;
}>();

const router = useRouter();
const { exercises, freeExerciseDb, ensureFreeDb, findBestImportMatch, buildExerciseRequests } = useExerciseResolver();
const { setAIWorkout, setAIPlan } = useAIFormState();

const resolving = ref(false);

const isWorkout = computed(() => props.artifact.type === "workout");
const workout = computed(() => isWorkout.value ? props.artifact.data as WorkoutArtifact : null);
const plan = computed(() => !isWorkout.value ? props.artifact.data as WorkoutPlanArtifact : null);

const userExerciseIds = computed(() => new Set(exercises.value.map((e) => e.id)));
const userExerciseNames = computed(() => new Set(exercises.value.map((e) => e.name.toLowerCase())));
const freeExerciseNames = computed(() => new Set(freeExerciseDb.value.map((e) => e.name.toLowerCase())));

onMounted(async () => {
  await ensureFreeDb();
});

type ExerciseStatus = "library" | "importable" | "unknown";

function exerciseStatus(exerciseId: number, exerciseName: string): ExerciseStatus {
  const nameLower = exerciseName.toLowerCase();
  if ((exerciseId > 0 && userExerciseIds.value.has(exerciseId)) || userExerciseNames.value.has(nameLower)) {
    return "library";
  }
  if (freeExerciseNames.value.has(nameLower) || findBestImportMatch(exerciseName) !== null) {
    return "importable";
  }
  return "unknown";
}

type ExerciseInfo = {
  name?: string;
  description?: string;
  instructions?: string[];
  image?: string;
  force?: string;
  category?: string;
  primary_muscle_groups?: { name: string }[];
  secondary_muscle_groups?: { name: string }[];
  equipment?: { name: string }[];
  exercise_features?: { name: string }[];
} | null;

function getExercisePreview(exerciseId: number, exerciseName: string): ExerciseInfo {
  const status = exerciseStatus(exerciseId, exerciseName);

  if (status === "library") {
    const nameLower = exerciseName.toLowerCase();
    return (
      exercises.value.find(
        (e) => (exerciseId > 0 && e.id === exerciseId) || e.name.toLowerCase() === nameLower,
      ) ?? null
    );
  }

  if (status === "importable") {
    const nameLower = exerciseName.toLowerCase();
    const match =
      freeExerciseDb.value.find((e) => e.name.toLowerCase() === nameLower) ??
      findBestImportMatch(exerciseName);
    if (!match) return null;
    return {
      name: match.name,
      instructions: match.instructions,
      image: match.images.length > 0 ? `${IMAGE_BASE_URL}${match.images[0]}` : undefined,
      force: match.force ?? undefined,
      category: match.category,
      primary_muscle_groups: match.primaryMuscles.map((m) => ({ name: m })),
      secondary_muscle_groups: match.secondaryMuscles.map((m) => ({ name: m })),
      equipment: match.equipment ? [{ name: match.equipment }] : [],
    };
  }

  return null;
}

async function handleReviewWorkout() {
  if (!workout.value) return;
  resolving.value = true;
  try {
    const exerciseRequests = await buildExerciseRequests(workout.value.exercises);
    setAIWorkout({
      name: workout.value.name,
      description: workout.value.description,
      exercises: exerciseRequests,
    });
    router.push({ name: "workout-create" });
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "Failed to resolve exercises");
    resolving.value = false;
  }
}

async function handleReviewPlan() {
  if (!plan.value) return;
  resolving.value = true;
  try {
    // Pre-resolve all exercises per training day so the plan form
    // only needs to create workouts (no exercise resolution at that point)
    const resolvedWeeks = await Promise.all(
      plan.value.weeks.map(async (week) => ({
        week_number: week.week_number,
        days: await Promise.all(
          (week.days as PlanDayArtifact[]).map(async (day) => {
            if (day.is_rest || !day.exercises?.length) {
              return {
                day_number: day.day_number,
                is_rest: day.is_rest,
                workout_name: day.workout_name,
                workout_description: day.workout_description,
                note: day.note,
              };
            }
            const exercises = await buildExerciseRequests(day.exercises);
            return {
              day_number: day.day_number,
              is_rest: day.is_rest,
              workout_name: day.workout_name,
              workout_description: day.workout_description,
              exercises,
              note: day.note,
            };
          }),
        ),
      })),
    );

    setAIPlan({
      name: plan.value.name,
      description: plan.value.description,
      weeks: resolvedWeeks,
    });
    router.push({ name: "workout-plan-create" });
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "Failed to resolve exercises");
    resolving.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col h-full bg-background">
    <!-- Header -->
    <div class="flex items-center gap-2 px-4 py-3 border-b shrink-0">
      <DumbbellIcon v-if="isWorkout" class="size-4 text-muted-foreground" />
      <CalendarDaysIcon v-else class="size-4 text-muted-foreground" />
      <span class="font-semibold text-sm">
        {{ isWorkout ? "Workout" : "Workout Plan" }}
      </span>
    </div>

    <!-- Workout artifact -->
    <div v-if="workout" class="flex-1 overflow-y-auto p-4 space-y-4">
      <div>
        <h2 class="font-semibold text-base">{{ workout.name }}</h2>
        <p v-if="workout.description" class="text-sm text-muted-foreground mt-0.5">{{ workout.description }}</p>
      </div>
      <ArtifactExerciseList
        :exercises="workout.exercises"
        :status-fn="exerciseStatus"
        :preview-fn="getExercisePreview"
      />
    </div>

    <!-- Plan artifact -->
    <div v-else-if="plan" class="flex-1 overflow-y-auto p-4 space-y-6">
      <div>
        <h2 class="font-semibold text-base">{{ plan.name }}</h2>
        <p v-if="plan.description" class="text-sm text-muted-foreground mt-0.5">{{ plan.description }}</p>
        <p class="text-xs text-muted-foreground mt-1">{{ plan.weeks.length }} weeks</p>
      </div>

      <div v-for="week in plan.weeks" :key="week.week_number" class="space-y-3">
        <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
          Week {{ week.week_number }}
        </p>

        <div
          v-for="day in week.days"
          :key="day.day_number"
          class="border rounded-md overflow-hidden"
        >
          <!-- Day header -->
          <div class="flex items-center gap-2 px-3 py-2 bg-muted/40 border-b">
            <span class="text-xs font-medium text-muted-foreground w-10 shrink-0">
              Day {{ day.day_number }}
            </span>
            <span v-if="day.is_rest" class="text-xs text-muted-foreground">Rest</span>
            <span v-else class="text-sm font-medium">{{ day.workout_name }}</span>
          </div>

          <!-- Rest day body -->
          <div v-if="day.is_rest" class="px-3 py-2 text-xs text-muted-foreground">
            Recovery day
          </div>

          <!-- Training day body -->
          <div v-else class="px-3 py-3 space-y-2">
            <p v-if="day.workout_description" class="text-xs text-muted-foreground">
              {{ day.workout_description }}
            </p>
            <ArtifactExerciseList
              v-if="day.exercises?.length"
              :exercises="day.exercises"
              :status-fn="exerciseStatus"
              :preview-fn="getExercisePreview"
              compact
            />
            <p v-if="day.note" class="text-xs text-muted-foreground italic pt-1">{{ day.note }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Footer actions -->
    <div class="border-t p-4 shrink-0">
      <Button
        v-if="isWorkout"
        class="w-full"
        :disabled="resolving"
        @click="handleReviewWorkout"
      >
        <LoaderCircleIcon v-if="resolving" class="size-4 mr-1.5 animate-spin" />
        <ArrowRightIcon v-else class="size-4 mr-1.5" />
        {{ resolving ? "Resolving..." : "Review Workout" }}
      </Button>
      <Button
        v-else
        class="w-full"
        :disabled="resolving"
        @click="handleReviewPlan"
      >
        <LoaderCircleIcon v-if="resolving" class="size-4 mr-1.5 animate-spin" />
        <ArrowRightIcon v-else class="size-4 mr-1.5" />
        {{ resolving ? "Resolving..." : "Review Plan" }}
      </Button>
    </div>
  </div>
</template>
