<script setup lang="ts">
import { computed, watch, ref, unref } from "vue";
import { useForm, useFieldArray } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { useI18n } from "vue-i18n";
import { useCreateWorkout } from "../composables/useCreateWorkout";
import { useUpdateWorkout } from "../composables/useUpdateWorkout";
import { useDeleteWorkout } from "../composables/useDeleteWorkout";
import { useExercises } from "@/features/exercises/composables/useExercises";
import type {
  Workout,
  WorkoutExerciseForm,
  WorkoutFormValues,
  WorkoutSetForm,
  WorkoutSetFeatureForm,
} from "../types";
import { workoutFormSchema } from "../types";
import {
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Plus, Trash2, GripVertical } from "lucide-vue-next";
import { VueDraggable } from "vue-draggable-plus";

const props = defineProps<{
  open?: boolean;
  modal?: boolean;
  workout?: Workout | null;
}>();

const emits = defineEmits<{
  (e: "workout-created"): void;
  (e: "workout-updated"): void;
  (e: "workout-deleted"): void;
  (e: "form-dirty", value: boolean): void;
}>();

const { t } = useI18n();
const isEditMode = computed(() => !!props.workout);

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

// Fetch all exercises for selection
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

const isFormDirty = computed(() => meta.value.dirty);

watch(
  [isFormDirty, () => props.open],
  ([dirty, isOpen]) => {
    if (isOpen) {
      emits("form-dirty", dirty);
    }
  },
  { immediate: true },
);

const populateForm = (workout: Workout | null) => {
  if (workout) {
    const exercises = (workout.exercises || []).map((ex) => ({
      id: ex.id,
      exercise_id: ex.exercise_id,
      rest_timer: ex.rest_timer,
      note: ex.note || "",
      order: ex.order,
      sets: (ex.sets || []).map((set) => ({
        id: set.id,
        order: set.order,
        features: (set.features || []).map((f) => ({
          id: f.id,
          feature_name: f.feature_name,
          value: f.value,
        })),
      })),
    }));

    resetForm({
      values: {
        name: workout.name,
        description: workout.description || "",
        exercises,
      },
    });
  }
};

watch(
  () => props.workout,
  (workout) => {
    if (workout && props.open && isEditMode.value) {
      populateForm(workout);
    }
  },
  { immediate: true },
);

watch(
  () => props.open,
  (newValue) => {
    if (newValue && props.workout) {
      populateForm(props.workout);
    } else if (!newValue) {
      resetForm();
      showDeleteDialog.value = false;
    } else if (newValue && !props.workout) {
      resetForm();
    }
  },
);

const addExercise = () => {
  pushExercise({
    exercise_id: null,
    rest_timer: 60,
    note: "",
    order: exerciseFields.value.length,
    sets: [],
  });
};

const syncExerciseOrders = () => {
  exerciseFields.value.forEach((field, idx) => {
    const ex = field.value;
    if (ex && ex.order !== idx) {
      updateExercise(idx, { ...ex, order: idx });
    }
  });
};

const removeExerciseAtIndex = (index: number) => {
  removeExercise(index);
  syncExerciseOrders();
};

const exercisesForDraggable = computed(
  () => (unref(values) as WorkoutFormValues)?.exercises ?? [],
);

const onExercisesReorder = (event: { oldIndex?: number; newIndex?: number }) => {
  const oldIndex = event.oldIndex ?? 0;
  const newIndex = event.newIndex ?? 0;
  if (oldIndex === newIndex) return;
  moveExercise(oldIndex, newIndex);
  syncExerciseOrders();
};

const setSetsInOrder = (exerciseIndex: number, newSets: WorkoutSetForm[]) => {
  const exercise = exerciseFields.value[exerciseIndex]?.value;
  if (!exercise) return;
  const withOrder = newSets.map((set, idx) => ({
    ...set,
    _key: String(set._key ?? set.id ?? crypto.randomUUID()),
    order: idx,
  }));
  updateExercise(exerciseIndex, { ...exercise, sets: withOrder });
};

const addSet = (exerciseIndex: number) => {
  const exercise = exerciseFields.value[exerciseIndex]?.value;
  if (!exercise) return;
  const currentSets = exercise.sets || [];
  const featureNames = exercise.exercise_id
    ? getExerciseFeatures(exercise.exercise_id)
    : [];
  const newSet = {
    _key: crypto.randomUUID(),
    order: currentSets.length,
    features: featureNames.map((name) => ({
      feature_name: name,
      value: 0,
    })),
  };
  updateExercise(exerciseIndex, {
    ...exercise,
    sets: [...currentSets, newSet],
  });
};

const removeSet = (exerciseIndex: number, setIndex: number) => {
  const exercise = exerciseFields.value[exerciseIndex]?.value;
  if (!exercise) return;
  const currentSets = exercise.sets || [];
  const newSets = currentSets
    .filter((_, idx) => idx !== setIndex)
    .map((set, idx) => ({ ...set, order: idx }));
  updateExercise(exerciseIndex, { ...exercise, sets: newSets });
};

const getExerciseFeatures = (exerciseId: number | null): string[] => {
  if (!exerciseId) return [];
  const exercise = allExercises.value.find((e) => e.id === exerciseId);
  return exercise?.exercise_features?.map((f) => f.name) || [];
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
          _key: crypto.randomUUID(),
          order: 0,
          features: featureNames.map((name) => ({ feature_name: name, value: 0 })),
        },
      ]
    : currentSets.map((set) => {
        const features = featureNames.map((name) => {
          const existing = set.features?.find((f) => f.feature_name === name);
          return existing ?? { feature_name: name, value: 0 };
        });
        return { ...set, features };
      });
  updateExercise(exerciseIndex, {
    ...exercise,
    exercise_id: exerciseId,
    sets: newSets,
  });
};

const getSetFeatureValue = (
  set: WorkoutSetForm,
  featureName: string,
): number =>
  set.features?.find((f) => f.feature_name === featureName)?.value ?? 0;

const updateFeatureValue = (
  exerciseIndex: number,
  setIndex: number,
  featureName: string,
  value: number,
) => {
  const exercise = exerciseFields.value[exerciseIndex]?.value;
  if (!exercise) return;
  const sets = [...(exercise.sets || [])];
  const set = sets[setIndex];
  if (!set) return;
  const features = [...(set.features || [])];
  const existingIndex = features.findIndex(
    (f: WorkoutSetFeatureForm) => f.feature_name === featureName,
  );
  if (existingIndex >= 0) {
    const f = features[existingIndex];
    if (f) {
      features[existingIndex] = { ...f, value };
    }
  } else {
    features.push({ feature_name: featureName, value });
  }
  sets[setIndex] = { ...set, features };
  updateExercise(exerciseIndex, { ...exercise, sets });
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
            value: existing?.value || 0,
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

    if (isEditMode.value && props.workout) {
      await updateWorkout({
        id: props.workout.id,
        data: {
          name: values.name.trim(),
          description: values.description?.trim() || undefined,
          exercises,
        },
      });
      resetForm();
      emits("workout-updated");
    } else {
      await createWorkout({
        name: values.name.trim(),
        description: values.description?.trim() || undefined,
        exercises,
      });
      resetForm();
      emits("workout-created");
    }
  } catch (err) {
    console.error(
      `Failed to ${isEditMode.value ? "update" : "create"} workout:`,
      err,
    );
  }
});

const onDelete = async () => {
  if (!props.workout) return;

  try {
    await deleteWorkout(props.workout.id);
    resetForm();
    showDeleteDialog.value = false;
    emits("workout-deleted");
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
  isEditMode.value
    ? t("workouts.editDescription")
    : t("workouts.createDescription"),
);

const submitButtonText = computed(() => {
  if (isEditMode.value) {
    return isUpdating.value ? t("updating") : t("workouts.update");
  }
  return isCreating.value ? t("creating") : t("workouts.create");
});

const drawerScrollRef = ref<HTMLElement | null>(null);
</script>

<template>
  <DrawerContent class="!max-h-[95vh]">
    <div ref="drawerScrollRef" class="mx-auto w-full max-w-4xl overflow-y-auto ">
      <DrawerHeader>
        <DrawerTitle>{{ title }}</DrawerTitle>
        <DrawerDescription>
          {{ description }}
        </DrawerDescription>
      </DrawerHeader>
      <div class="p-4 pb-0 space-y-6">
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

            <VueDraggable :model-value="exercisesForDraggable" :custom-update="onExercisesReorder"
              handle=".exercise-drag-handle" :force-fallback="true" :fallback-on-body="true"
              ghost-class="workout-drag-ghost" chosen-class="workout-drag-chosen" fallback-class="workout-drag-fallback"
              :animation="200" :scroll="drawerScrollRef || true" :bubble-scroll="true" :scroll-sensitivity="80"
              :scroll-speed="16" class="space-y-4">
              <div v-for="(field, exerciseIndex) in exerciseFields" :key="field.key"
                class="border rounded-lg p-4 space-y-4">
                <div class="flex items-start justify-between gap-4">
                  <button type="button"
                    class="exercise-drag-handle mt-1 p-1.5 rounded-md cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground hover:bg-muted/80 touch-none transition-colors"
                    tabindex="-1">
                    <GripVertical class="w-4 h-4" />
                  </button>
                  <div class="flex-1 space-y-4 min-w-0">
                    <FormField v-slot="{ componentField }" :name="`exercises.${exerciseIndex}.exercise_id`">
                      <FormItem>
                        <FormLabel>{{ $t("exercises.title") }}</FormLabel>
                        <FormControl>
                          <Select
                            v-model="componentField.modelValue"
                            :disabled="!!field.value?.exercise_id"
                            @update:model-value="(v) => onExerciseSelected(exerciseIndex, field.value, v)"
                          >
                            <SelectTrigger>
                              <SelectValue :placeholder="$t('exercises.title')" />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem v-for="option in exerciseOptions" :key="option.value" :value="option.value">
                                {{ option.label }}
                              </SelectItem>
                            </SelectContent>
                          </Select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    </FormField>

                    <template v-if="field.value?.exercise_id">
                    <div class="grid grid-cols-2 gap-4">
                      <FormField v-slot="{ componentField }" :name="`exercises.${exerciseIndex}.rest_timer`">
                        <FormItem>
                          <FormLabel>{{ $t("workouts.restTimer") }}</FormLabel>
                          <FormControl>
                            <Input type="number" min="0" v-bind="componentField" />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      </FormField>

                      <FormField v-slot="{ componentField }" :name="`exercises.${exerciseIndex}.note`">
                        <FormItem>
                          <FormLabel>{{ $t("workouts.note") }}</FormLabel>
                          <FormControl>
                            <Textarea :placeholder="$t('workouts.notePlaceholder')" rows="2" v-bind="componentField" />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      </FormField>
                    </div>

                    <div class="space-y-2">
                      <div class="flex items-center justify-between">
                        <Label class="text-base font-medium">{{ $t("workouts.sets") }}</Label>
                        <Button type="button" variant="outline" size="sm" @click="addSet(exerciseIndex)"
                          :disabled="!field.value?.exercise_id">
                          <Plus class="w-4 h-4 mr-2" />
                          {{ $t("workouts.addSet") }}
                        </Button>
                      </div>
                      <FormField :name="`exercises.${exerciseIndex}.sets`" v-slot>
                        <FormItem>
                          <FormMessage />
                        </FormItem>
                      </FormField>
                      <VueDraggable :model-value="field.value?.sets || []"
                        @update:model-value="(v: unknown) => setSetsInOrder(exerciseIndex, v as WorkoutExerciseForm['sets'])"
                        handle=".set-drag-handle" :force-fallback="true" :fallback-on-body="true"
                        ghost-class="workout-drag-ghost-set" chosen-class="workout-drag-chosen"
                        fallback-class="workout-drag-fallback-set" :animation="150" :scroll="drawerScrollRef || true"
                        :bubble-scroll="true" :scroll-sensitivity="80" :scroll-speed="16" class="space-y-2">
                        <div v-for="(set, setIndex) in field.value?.sets || []" :key="set.id ?? set._key ?? setIndex"
                          class="border rounded p-3 space-y-2">
                          <div class="flex items-center justify-between">
                            <button type="button"
                              class="set-drag-handle p-1 rounded cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground hover:bg-muted/80 touch-none transition-colors"
                              tabindex="-1">
                              <GripVertical class="w-4 h-4" />
                            </button>
                            <Label class="text-sm font-medium flex-1">
                              {{ $t("workouts.setNumber", { number: setIndex + 1 }) }}
                            </Label>
                            <Button
                              v-if="(field.value?.sets?.length ?? 0) > 1"
                              type="button"
                              variant="ghost"
                              size="sm"
                              @click="removeSet(exerciseIndex, Number(setIndex))"
                            >
                              <Trash2 class="w-4 h-4" />
                            </Button>
                          </div>

                          <div v-if="field.value?.exercise_id" class="grid grid-cols-2 gap-2">
                            <FormField v-for="featureName in getExerciseFeatures(
                              field.value.exercise_id,
                            )" :key="featureName"
                              :name="`exercises[${exerciseIndex}].sets.${setIndex}.features.${getExerciseFeatures(field.value.exercise_id).indexOf(featureName)}.value`">
                              <FormItem>
                                <FormLabel class="text-sm capitalize">
                                  {{ featureName }}
                                </FormLabel>
                                <FormControl>
                                  <Input type="number" step="0.01" min="0"
                                    :model-value="getSetFeatureValue(set, featureName)" @update:model-value="
                                      (value: string | number) =>
                                        updateFeatureValue(
                                          exerciseIndex,
                                          setIndex,
                                          featureName,
                                          Number(value) || 0,
                                        )
                                    " />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            </FormField>
                          </div>
                        </div>
                      </VueDraggable>
                    </div>
                    </template>
                  </div>
                  <Button type="button" variant="ghost" size="sm" @click="removeExerciseAtIndex(exerciseIndex)"
                    class="shrink-0">
                    <Trash2 class="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </VueDraggable>

            <FormField v-slot="{ componentField }" name="exercises">
              <FormItem>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>
        </form>
      </div>
      <DrawerFooter class="flex-col gap-2 justify-between">
        <Dialog v-if="isEditMode" v-model:open="showDeleteDialog">
          <DialogTrigger as-child>
            <Button type="button" variant="destructive" :disabled="isPending">
              {{ $t("delete") }}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{{ $t("delete") }}</DialogTitle>
              <DialogDescription>
                {{ $t("areYouSure") }}
              </DialogDescription>
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
        <DrawerClose as-child>
          <Button variant="outline">{{ $t("cancel") }}</Button>
        </DrawerClose>
      </DrawerFooter>
    </div>
  </DrawerContent>
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
/* Global: fallback clone is appended to body */
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
