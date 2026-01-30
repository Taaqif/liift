<script setup lang="ts">
import { computed, watch, ref, onUnmounted } from "vue";
import { useForm, useFieldArray } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useCreateWorkout } from "../composables/useCreateWorkout";
import { useUpdateWorkout } from "../composables/useUpdateWorkout";
import { useDeleteWorkout } from "../composables/useDeleteWorkout";
import { useExercises } from "@/features/exercises/composables/useExercises";
import type { Workout, WorkoutExercise, WorkoutSet } from "../types";
import type { Exercise } from "@/features/exercises/types";
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
import { Plus, Trash2 } from "lucide-vue-next";

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

// Create a dynamic schema based on exercise features
const createExerciseSchema = (exerciseId: number | null) => {
  if (!exerciseId) {
    return z.object({
      exercise_id: z.number().min(1, t("workouts.validation.exercisesRequired")),
      rest_timer: z.number().min(0),
      note: z.string().optional(),
      order: z.number(),
      sets: z.array(z.any()).min(1, t("workouts.validation.setsRequired")),
    });
  }

  const exercise = allExercises.value.find((e) => e.id === exerciseId);
  if (!exercise || !exercise.exercise_features) {
    return z.object({
      exercise_id: z.number(),
      rest_timer: z.number().min(0),
      note: z.string().optional(),
      order: z.number(),
      sets: z.array(z.any()).min(1),
    });
  }

  const featureNames = exercise.exercise_features.map((f) => f.name);

  const setSchema = z.object({
    order: z.number(),
    features: z
      .array(
        z.object({
          feature_name: z.string(),
          value: z.number().min(0),
        }),
      )
      .refine(
        (features) => {
          // Check that all enabled features have values
          return featureNames.every((name) =>
            features.some((f) => f.feature_name === name),
          );
        },
        {
          message: t("workouts.validation.featuresRequired"),
        },
      ),
  });

  return z.object({
    exercise_id: z.number().min(1),
    rest_timer: z.number().min(0),
    note: z.string().optional(),
    order: z.number(),
    sets: z.array(setSchema).min(1, t("workouts.validation.setsRequired")),
  });
};

const formSchema = z.object({
  name: z.string().min(1, t("workouts.validation.nameRequired")),
  description: z.string().optional(),
  exercises: z
    .array(z.any())
    .min(1, t("workouts.validation.exercisesRequired"))
    .refine(
      (arr) => arr.every((ex: { sets?: unknown[] }) => (ex?.sets?.length ?? 0) >= 1),
      { message: t("workouts.validation.setsRequired") },
    ),
});

const { handleSubmit, resetForm, meta, setFieldValue } = useForm({
  validationSchema: toTypedSchema(formSchema),
  initialValues: {
    name: "",
    description: "",
    exercises: [] as any[],
  },
});

const { fields: exerciseFields, push: pushExercise, remove: removeExercise } =
  useFieldArray("exercises");

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

const removeExerciseAtIndex = (index: number) => {
  removeExercise(index);
  // Update order for remaining exercises
  exerciseFields.value.forEach((field, idx) => {
    setFieldValue(`exercises.${idx}.order`, idx);
  });
};

const addSet = (exerciseIndex: number) => {
  const currentSets = exerciseFields.value[exerciseIndex].value?.sets || [];
  const exerciseId = exerciseFields.value[exerciseIndex].value?.exercise_id;
  const featureNames = exerciseId ? getExerciseFeatures(exerciseId) : [];
  const newSet = {
    order: currentSets.length,
    features: featureNames.map((name) => ({
      feature_name: name,
      value: 0,
    })),
  };
  setFieldValue(`exercises.${exerciseIndex}.sets`, [...currentSets, newSet]);
};

const removeSet = (exerciseIndex: number, setIndex: number) => {
  const currentSets = exerciseFields.value[exerciseIndex].value?.sets || [];
  const newSets = currentSets.filter((_: any, idx: number) => idx !== setIndex);
  setFieldValue(`exercises.${exerciseIndex}.sets`, newSets);
  // Update order
  newSets.forEach((_: any, idx: number) => {
    setFieldValue(`exercises.${exerciseIndex}.sets.${idx}.order`, idx);
  });
};

const getExerciseFeatures = (exerciseId: number | null): string[] => {
  if (!exerciseId) return [];
  const exercise = allExercises.value.find((e) => e.id === exerciseId);
  return exercise?.exercise_features?.map((f) => f.name) || [];
};

const updateSetFeatures = (
  exerciseIndex: number,
  setIndex: number,
  exerciseId: number,
) => {
  const features = getExerciseFeatures(exerciseId);
  const currentSet =
    exerciseFields.value[exerciseIndex].value?.sets[setIndex] || {};
  const currentFeatures = currentSet.features || [];

  // Remove features that are no longer enabled
  const validFeatures = currentFeatures.filter((f: any) =>
    features.includes(f.feature_name),
  );

  // Add missing features
  const existingFeatureNames = validFeatures.map((f: any) => f.feature_name);
  const missingFeatures = features
    .filter((name) => !existingFeatureNames.includes(name))
    .map((name) => ({
      feature_name: name,
      value: 0,
    }));

  setFieldValue(`exercises.${exerciseIndex}.sets.${setIndex}.features`, [
    ...validFeatures,
    ...missingFeatures,
  ]);
};

const onSubmit = handleSubmit(async (values) => {
  try {
    // Validate and prepare exercises
    const exercises = values.exercises.map((ex: any, idx: number) => {
      const exerciseFeatures = getExerciseFeatures(ex.exercise_id);
      const sets = ex.sets.map((set: any, setIdx: number) => {
        const features = exerciseFeatures.map((featureName) => {
          const existing = set.features?.find(
            (f: any) => f.feature_name === featureName,
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
</script>

<template>
  <DrawerContent class="max-h-[95vh]">
    <div class="mx-auto w-full max-w-4xl overflow-y-auto">
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

            <div v-for="(field, exerciseIndex) in exerciseFields" :key="field.key"
              class="border rounded-lg p-4 space-y-4">
              <div class="flex items-start justify-between gap-4">
                <div class="flex-1 space-y-4">
                  <FormField v-slot="{ componentField }" :name="`exercises.${exerciseIndex}.exercise_id`">
                    <FormItem>
                      <FormLabel>{{ $t("exercises.title") }}</FormLabel>
                      <FormControl>
                        <Select v-model="componentField.modelValue" @update:model-value="
                          (value) => {
                            componentField['onUpdate:modelValue']?.(value);
                            if (value) {
                              // Update all sets for this exercise
                              const sets = field.value?.sets || [];
                              sets.forEach((_: any, setIdx: number) => {
                                updateSetFeatures(exerciseIndex, setIdx, value);
                              });
                            }
                          }
                        ">
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

                    <div v-for="(set, setIndex) in field.value?.sets || []" :key="setIndex"
                      class="border rounded p-3 space-y-2">
                      <div class="flex items-center justify-between">
                        <Label class="text-sm font-medium">
                          {{ $t("workouts.setNumber", { number: setIndex + 1 }) }}
                        </Label>
                        <Button type="button" variant="ghost" size="sm" @click="removeSet(exerciseIndex, setIndex)">
                          <Trash2 class="w-4 h-4" />
                        </Button>
                      </div>

                      <div v-if="field.value?.exercise_id" class="grid grid-cols-2 gap-2">
                        <FormField v-for="featureName in getExerciseFeatures(
                          field.value.exercise_id,
                        )" :key="featureName"
                          :name="`exercises.${exerciseIndex}.sets.${setIndex}.features.${getExerciseFeatures(field.value.exercise_id).indexOf(featureName)}.value`">
                          <FormItem>
                            <FormLabel class="text-sm capitalize">
                              {{ featureName }}
                            </FormLabel>
                            <FormControl>
                              <Input type="number" step="0.01" min="0" :model-value="set.features?.find(
                                (f: any) => f.feature_name === featureName,
                              )?.value || 0
                                " @update:model-value="
                                  (value) => {
                                    const currentFeatures =
                                      set.features || [];
                                    const existingIndex =
                                      currentFeatures.findIndex(
                                        (f: any) =>
                                          f.feature_name === featureName,
                                      );
                                    if (existingIndex >= 0) {
                                      setFieldValue(
                                        `exercises.${exerciseIndex}.sets.${setIndex}.features.${existingIndex}.value`,
                                        parseFloat(value) || 0,
                                      );
                                    } else {
                                      const newFeatures = [
                                        ...currentFeatures,
                                        {
                                          feature_name: featureName,
                                          value: parseFloat(value) || 0,
                                        },
                                      ];
                                      setFieldValue(
                                        `exercises.${exerciseIndex}.sets.${setIndex}.features`,
                                        newFeatures,
                                      );
                                    }
                                  }
                                " />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        </FormField>
                      </div>
                    </div>
                  </div>
                </div>
                <Button type="button" variant="ghost" size="sm" @click="removeExerciseAtIndex(exerciseIndex)"
                  class="shrink-0">
                  <Trash2 class="w-4 h-4" />
                </Button>
              </div>
            </div>

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
