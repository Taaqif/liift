<script setup lang="ts">
import { computed, watch, ref, unref, onMounted } from "vue";
import { generateId } from "@/lib/utils";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { useForm, useFieldArray } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { useI18n } from "vue-i18n";
import { useWorkout } from "@/features/workouts/composables/useWorkout";
import { useCreateWorkout } from "@/features/workouts/composables/useCreateWorkout";
import { useUpdateWorkout } from "@/features/workouts/composables/useUpdateWorkout";
import { useDeleteWorkout } from "@/features/workouts/composables/useDeleteWorkout";
import { useExercises } from "@/features/exercises/composables/useExercises";
import type {
  Workout,
  WorkoutExerciseForm,
  WorkoutFormValues,
} from "@/features/workouts/types";
import { workoutFormSchema } from "@/features/workouts/types";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Plus, Trash2 } from "lucide-vue-next";
import { VueDraggable } from "vue-draggable-plus";
import WorkoutExerciseItem from "@/features/workouts/components/WorkoutExerciseItem.vue";
import ExercisePickerSheet from "@/features/exercises/components/ExercisePickerSheet.vue";
import type { Exercise } from "@/features/exercises/types";
import { useAIFormState } from "@/features/ai-chat/composables/useAIFormState";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const workoutId = computed(() => {
  const id = route.params.id;
  return id ? Number(id) : null;
});
const isEditMode = computed(() => !!workoutId.value);

const { workout, loading: workoutLoading } = useWorkout(workoutId);

const {
  createWorkout,
  isPending: isCreating,
  error: createError,
} = useCreateWorkout();
const {
  updateWorkout,
  isPending: isUpdating,
  error: updateError,
} = useUpdateWorkout();
const {
  deleteWorkout,
  isPending: isDeleting,
  error: deleteError,
} = useDeleteWorkout();

const { exercises: allExercises } = useExercises({ limit: 1000 });

const exerciseOptions = computed(() =>
  allExercises.value.map((ex) => ({
    value: ex.id,
    label: ex.name,
    exercise: ex,
  })),
);

const { handleSubmit, resetForm, meta, values } = useForm<WorkoutFormValues>({
  validationSchema: toTypedSchema(workoutFormSchema),
  initialValues: {
    name: "",
    description: "",
    exercises: [],
  },
});

const {
  fields: exerciseFields,
  push: pushExercise,
  remove: removeExercise,
  move: moveExercise,
  update: updateExercise,
} = useFieldArray<WorkoutExerciseForm>("exercises");

const getExerciseFeatures = (exerciseId: number | null): string[] => {
  if (!exerciseId) return [];
  const exercise = allExercises.value.find((e) => e.id === exerciseId);
  return exercise?.exercise_features?.map((f) => f.name) || [];
};

const populateForm = (w: Workout) => {
  const exercises = (w.exercises || []).map((ex) => {
    // Merge saved features with current exercise feature list so new features
    // added after the workout was created are included with null values.
    const featureNames = getExerciseFeatures(ex.exercise_id);
    return {
      id: ex.id,
      exercise_id: ex.exercise_id,
      rest_timer: ex.rest_timer,
      note: ex.note || "",
      order: ex.order,
      sets: (ex.sets || []).map((set) => {
        const savedFeatures = set.features || [];
        const features = featureNames.length > 0
          ? featureNames.map((name) => {
              const saved = savedFeatures.find((f) => f.feature_name === name);
              return {
                ...(saved?.id != null && { id: saved.id }),
                feature_name: name,
                value: saved?.value || null,
              };
            })
          : savedFeatures.map((f) => ({
              id: f.id,
              feature_name: f.feature_name,
              value: f.value || null,
            }));
        return { id: set.id, order: set.order, features };
      }),
    };
  });
  resetForm({
    values: {
      name: w.name,
      description: w.description || "",
      exercises,
    },
  });
};

const formPopulated = ref(false);
const { takeAIWorkout } = useAIFormState();

// AI pre-fill: apply immediately on mount since exercises are pre-resolved
onMounted(() => {
  const aiWorkout = takeAIWorkout();
  if (aiWorkout?.exercises?.length) {
    resetForm({
      values: {
        name: aiWorkout.name,
        description: aiWorkout.description || "",
        exercises: aiWorkout.exercises.map((ex) => ({
          exercise_id: ex.exercise_id,
          rest_timer: ex.rest_timer ?? 60,
          note: ex.note || "",
          order: ex.order,
          sets: ex.sets.map((set) => ({
            order: set.order,
            features: set.features.map((f) => ({ feature_name: f.feature_name, value: f.value ?? null })),
          })),
        })),
      },
    });
    formPopulated.value = true;
  }
});

// Wait for both workout and allExercises before populating so getExerciseFeatures works
watch([workout, allExercises], ([w]) => {
  if (w && allExercises.value.length > 0 && !formPopulated.value) {
    populateForm(w);
    formPopulated.value = true;
  }
}, { immediate: true });

const showExercisePicker = ref(false);

const addExercise = () => {
  showExercisePicker.value = true;
};

const handleExercisesAdded = (exercises: Exercise[]) => {
  exercises.forEach((exercise) => {
    const featureNames = exercise.exercise_features?.map((f) => f.name) || [];
    pushExercise({
      exercise_id: exercise.id,
      rest_timer: 60,
      note: "",
      order: exerciseFields.value.length,
      sets: [
        {
          _key: generateId(),
          order: 0,
          features: featureNames.map((name) => ({ feature_name: name, value: null })),
        },
      ],
    });
  });
};

const removeExerciseAtIndex = (index: number) => {
  removeExercise(index);
};

const exercisesForDraggable = computed(
  () => (unref(values) as WorkoutFormValues)?.exercises ?? [],
);

const onExercisesReorder = (event: { oldIndex?: number; newIndex?: number }) => {
  const oldIndex = event.oldIndex ?? 0;
  const newIndex = event.newIndex ?? 0;
  if (oldIndex === newIndex) return;
  moveExercise(oldIndex, newIndex);
};

const getExercise = (exerciseId: number | null) => {
  if (!exerciseId) return undefined;
  return allExercises.value.find((e) => e.id === exerciseId);
};

const onExerciseSelected = (
  exerciseIndex: number,
  exercise: WorkoutExerciseForm | undefined,
  value: unknown,
) => {
  if (!value || !exercise) return;
  const exerciseId = Number(value);
  const featureNames = getExerciseFeatures(exerciseId);
  const currentSets = exercise.sets ?? [];
  const needsInitialSet = currentSets.length === 0;
  const newSets = needsInitialSet
    ? [
        {
          _key: generateId(),
          order: 0,
          features: featureNames.map((name) => ({ feature_name: name, value: null })),
        },
      ]
    : currentSets.map((set) => {
        const features = featureNames.map((name) => {
          const existing = set.features?.find((f) => f.feature_name === name);
          return existing ?? { feature_name: name, value: null };
        });
        return { ...set, features };
      });
  updateExercise(exerciseIndex, {
    ...exercise,
    exercise_id: exerciseId,
    sets: newSets,
  });
};

const onSubmit = handleSubmit(async (values) => {
  try {
    const validExercises = values.exercises.filter(
      (ex): ex is WorkoutExerciseForm & { exercise_id: number } =>
        ex.exercise_id != null,
    );
    const exercises = validExercises.map((ex, idx) => {
      const exerciseFeatures = getExerciseFeatures(ex.exercise_id);
      const sets = ex.sets.map((set, setIdx) => {
        const features = exerciseFeatures.map((featureName) => {
          const existing = set.features?.find(
            (f) => f.feature_name === featureName,
          );
          return {
            ...(existing?.id != null && { id: existing.id }),
            feature_name: featureName,
            value: existing?.value ?? null,
          };
        });
        return {
          ...(set.id != null && { id: set.id }),
          order: setIdx,
          features,
        };
      });
      return {
        ...(ex.id != null && { id: ex.id }),
        exercise_id: ex.exercise_id,
        rest_timer: ex.rest_timer || 0,
        note: ex.note || "",
        order: idx,
        sets,
      };
    });

    if (isEditMode.value && workoutId.value) {
      await updateWorkout({
        id: workoutId.value,
        data: {
          name: values.name.trim(),
          description: values.description?.trim() || undefined,
          exercises,
        },
      });
    } else {
      await createWorkout({
        name: values.name.trim(),
        description: values.description?.trim() || undefined,
        exercises,
      });
    }
    resetForm();
    router.push({ name: "workouts" });
  } catch (err) {
    console.error(
      `Failed to ${isEditMode.value ? "update" : "create"} workout:`,
      err,
    );
  }
});

const onDelete = async () => {
  if (!workoutId.value) return;
  try {
    await deleteWorkout(workoutId.value);
    showDeleteDialog.value = false;
    router.push({ name: "workouts" });
  } catch (err) {
    console.error("Failed to delete workout:", err);
  }
};

const error = computed(
  () => createError.value || updateError.value || deleteError.value,
);
const isPending = computed(
  () => isCreating.value || isUpdating.value || isDeleting.value,
);
const showDeleteDialog = ref(false);

const title = computed(() =>
  isEditMode.value ? t("workouts.editTitle") : t("workouts.createNew"),
);
const description = computed(() =>
  isEditMode.value ? t("workouts.editDescription") : t("workouts.createDescription"),
);
const submitButtonText = computed(() => {
  if (isEditMode.value) {
    return isUpdating.value ? t("updating") : t("workouts.update");
  }
  return isCreating.value ? t("creating") : t("workouts.create");
});

const pageScrollRef = ref<HTMLElement | null>(null);

onBeforeRouteLeave(() => {
  if (meta.value.dirty && !isPending.value) {
    return window.confirm(t("unsavedChanges.confirmLeave") || "You have unsaved changes. Leave anyway?");
  }
});
</script>

<template>
  <div ref="pageScrollRef">
    <!-- Header -->
    <div class="mb-8">
      <button class="text-sm text-muted-foreground hover:text-foreground transition-colors mb-1" @click="router.push({ name: 'workouts' })">
        ← {{ $t("nav.workouts") }}
      </button>
      <h1 class="text-3xl font-bold tracking-tight">{{ title }}</h1>
      <p class="text-muted-foreground mt-1">{{ description }}</p>
    </div>

    <div v-if="isEditMode && workoutLoading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">{{ $t("loading") }}</div>
    </div>

    <div v-else class="space-y-6">
      <div v-if="error" class="p-4 bg-destructive/10 text-destructive rounded-lg">
        <p>{{ $t("workouts.error") }}: {{ error.message }}</p>
      </div>

      <form @submit="onSubmit" class="space-y-6">
        <FormField v-slot="{ componentField }" name="name">
          <FormItem>
            <FormLabel>{{ $t("workouts.name") }}</FormLabel>
            <FormControl>
              <Input :placeholder="$t('workouts.namePlaceholder')" v-bind="componentField" required />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="description">
          <FormItem>
            <FormLabel>{{ $t("workouts.description") }}</FormLabel>
            <FormControl>
              <Textarea :placeholder="$t('workouts.descriptionPlaceholder')" rows="3" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <Label class="text-base font-medium">{{ $t("workouts.exercises") }}</Label>
            <Button type="button" variant="outline" size="sm" @click="addExercise">
              <Plus class="w-4 h-4 mr-2" />
              {{ $t("workouts.addExercise") }}
            </Button>
          </div>

          <VueDraggable
            :model-value="exercisesForDraggable"
            :custom-update="onExercisesReorder"
            handle=".exercise-drag-handle"
            :force-fallback="true"
            :fallback-on-body="true"
            ghost-class="workout-drag-ghost"
            chosen-class="workout-drag-chosen"
            fallback-class="workout-drag-fallback"
            :animation="200"
            :scroll="pageScrollRef || true"
            :bubble-scroll="true"
            :scroll-sensitivity="80"
            :scroll-speed="16"
            class="space-y-4"
          >
            <WorkoutExerciseItem
              v-for="(field, exerciseIndex) in exerciseFields"
              :key="field.key"
              :exercise-index="exerciseIndex"
              :field="(field as { value: WorkoutExerciseForm; key: string })"
              :exercise-options="exerciseOptions"
              :get-exercise-features="getExerciseFeatures"
              :get-exercise="getExercise"
              :drawer-scroll-ref="pageScrollRef"
              @remove="removeExerciseAtIndex(exerciseIndex)"
            />
          </VueDraggable>

          <FormField name="exercises">
            <FormItem>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </form>

      <!-- Action buttons -->
      <div class="flex flex-col gap-2 pt-4 border-t">
        <Dialog v-if="isEditMode" v-model:open="showDeleteDialog">
          <DialogTrigger as-child>
            <Button type="button" variant="destructive" :disabled="isPending">
              {{ $t("delete") }}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{{ $t("delete") }}</DialogTitle>
              <DialogDescription>{{ $t("areYouSure") }}</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" @click="showDeleteDialog = false" :disabled="isDeleting">
                {{ $t("cancel") }}
              </Button>
              <Button variant="destructive" @click="onDelete" :disabled="isDeleting">
                {{ isDeleting ? $t("deleting") : $t("delete") }}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        <Button @click="onSubmit" :disabled="isPending" class="flex-1">
          {{ submitButtonText }}
        </Button>
        <Button variant="outline" @click="router.push({ name: 'workouts' })">
          {{ $t("cancel") }}
        </Button>
      </div>
    </div>

    <ExercisePickerSheet
      v-model:open="showExercisePicker"
      @add="handleExercisesAdded"
    />
  </div>
</template>

<style scoped>
:deep(.workout-drag-ghost) {
  border-radius: 0.5rem;
  border: 2px dashed hsl(var(--primary) / 0.4);
  background: hsl(var(--primary) / 0.05);
  min-height: 80px;
  opacity: 0.9;
  transition: all 0.15s ease;
}

:deep(.workout-drag-ghost-set) {
  border-radius: 0.5rem;
  border: 2px dashed hsl(var(--primary) / 0.4);
  background: hsl(var(--primary) / 0.05);
  min-height: 60px;
  opacity: 0.9;
  transition: all 0.15s ease;
}

:deep(.workout-drag-chosen) {
  opacity: 0.4;
}
</style>

<style>
.workout-drag-fallback {
  border-radius: 0.5rem;
  border: 1px solid hsl(var(--border));
  background: hsl(var(--background));
  box-shadow: 0 12px 40px -12px rgb(0 0 0 / 0.25), 0 0 0 1px rgb(0 0 0 / 0.05);
  transform: scale(1.02) rotate(1deg);
  opacity: 1;
  cursor: grabbing;
  z-index: 9999;
}

.workout-drag-fallback-set {
  border-radius: 0.5rem;
  border: 1px solid hsl(var(--border));
  background: hsl(var(--background));
  box-shadow: 0 12px 40px -12px rgb(0 0 0 / 0.25), 0 0 0 1px rgb(0 0 0 / 0.05);
  transform: scale(1.01) rotate(0.5deg);
  opacity: 1;
  cursor: grabbing;
  z-index: 9999;
}
</style>
