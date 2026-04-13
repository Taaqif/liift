<script setup lang="ts">
import { computed, ref, unref } from "vue";
import { generateId } from "@/lib/utils";
import { useForm, useFieldArray } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { useI18n } from "vue-i18n";
import { useCreateWorkout } from "@/features/workouts/composables/useCreateWorkout";
import { useExercises } from "@/features/exercises/composables/useExercises";
import type { Workout, WorkoutExerciseForm, WorkoutFormValues } from "@/features/workouts/types";
import { workoutFormSchema } from "@/features/workouts/types";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
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
import { Plus } from "lucide-vue-next";
import { VueDraggable } from "vue-draggable-plus";
import WorkoutExerciseItem from "@/features/workouts/components/WorkoutExerciseItem.vue";

const props = defineProps<{
  open: boolean;
  dayLabel: string;
  saveToLibrary?: boolean;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
  "workout-created": [workout: Workout];
}>();

const { t } = useI18n();
const { createWorkout, isPending } = useCreateWorkout();
const { exercises: allExercises } = useExercises({ limit: 1000 });

const exerciseOptions = computed(() =>
  allExercises.value.map((ex) => ({
    value: ex.id,
    label: ex.name,
    exercise: ex,
  })),
);

const { handleSubmit, resetForm, values } = useForm<WorkoutFormValues>({
  validationSchema: toTypedSchema(workoutFormSchema),
  initialValues: { name: "", description: "", exercises: [] },
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
  return exercise?.exercise_features?.map((f) => f.name) ?? [];
};

const getExercise = (exerciseId: number | null) => {
  if (!exerciseId) return undefined;
  return allExercises.value.find((e) => e.id === exerciseId);
};

const addExercise = () => {
  pushExercise({
    exercise_id: null,
    rest_timer: 60,
    note: "",
    order: exerciseFields.value.length,
    sets: [],
  });
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
  const newSets =
    currentSets.length === 0
      ? [
          {
            _key: generateId(),
            order: 0,
            features: featureNames.map((name) => ({ feature_name: name, value: null })),
          },
        ]
      : currentSets.map((set) => ({
          ...set,
          features: featureNames.map((name) => {
            const existing = set.features?.find((f) => f.feature_name === name);
            return existing ?? { feature_name: name, value: null };
          }),
        }));
  updateExercise(exerciseIndex, { ...exercise, exercise_id: exerciseId, sets: newSets });
};

const exercisesForDraggable = computed(
  () => (unref(values) as WorkoutFormValues)?.exercises ?? [],
);

const onExercisesReorder = (event: { oldIndex?: number; newIndex?: number }) => {
  const oldIndex = event.oldIndex ?? 0;
  const newIndex = event.newIndex ?? 0;
  if (oldIndex !== newIndex) moveExercise(oldIndex, newIndex);
};

const error = ref<string | null>(null);

const onSubmit = handleSubmit(async (formValues) => {
  error.value = null;
  try {
    const validExercises = formValues.exercises.filter(
      (ex): ex is WorkoutExerciseForm & { exercise_id: number } => ex.exercise_id != null,
    );
    const exercises = validExercises.map((ex, idx) => {
      const featureNames = getExerciseFeatures(ex.exercise_id);
      return {
        exercise_id: ex.exercise_id,
        rest_timer: ex.rest_timer || 0,
        note: ex.note || "",
        order: idx,
        sets: ex.sets.map((set, setIdx) => ({
          order: setIdx,
          features: featureNames.map((name) => {
            const existing = set.features?.find((f) => f.feature_name === name);
            return { feature_name: name, value: existing?.value ?? 0 };
          }),
        })),
      };
    });

    const isLibrary = props.saveToLibrary !== false;
    const workout = await createWorkout({
      name: formValues.name.trim(),
      description: formValues.description?.trim() || undefined,
      is_library: isLibrary,
      exercises,
    });

    emit("workout-created", workout);
    resetForm();
    emit("update:open", false);
  } catch (err) {
    error.value = t("workouts.error");
  }
});

const onCancel = () => {
  resetForm();
  error.value = null;
  emit("update:open", false);
};

// Scroll ref for drag-to-scroll inside the dialog
const scrollRef = ref<HTMLElement | null>(null);
</script>

<template>
  <Dialog :open="open" @update:open="(v) => !v && onCancel()">
    <DialogContent class="max-w-2xl max-h-[85vh] flex flex-col gap-0 p-0">
      <DialogHeader class="px-6 pt-6 pb-4 border-b shrink-0">
        <DialogTitle>{{ $t("workoutPlans.adHocWorkout.title") }}</DialogTitle>
        <DialogDescription>
          {{ saveToLibrary !== false
            ? $t("workoutPlans.adHocWorkout.description", { day: dayLabel })
            : $t("workoutPlans.adHocWorkout.descriptionPlanOnly", { day: dayLabel }) }}
        </DialogDescription>
      </DialogHeader>

      <div ref="scrollRef" class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
        <div v-if="error" class="p-3 bg-destructive/10 text-destructive rounded-lg text-sm">
          {{ error }}
        </div>

        <FormField v-slot="{ componentField }" name="name">
          <FormItem>
            <FormLabel>{{ $t("workouts.name") }}</FormLabel>
            <FormControl>
              <Input
                :placeholder="$t('workouts.namePlaceholder')"
                v-bind="componentField"
                required
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="description">
          <FormItem>
            <FormLabel>{{ $t("workouts.description") }}</FormLabel>
            <FormControl>
              <Textarea
                :placeholder="$t('workouts.descriptionPlaceholder')"
                rows="2"
                v-bind="componentField"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label class="text-sm font-medium">{{ $t("workouts.exercises") }}</Label>
            <Button type="button" variant="outline" size="sm" @click="addExercise">
              <Plus class="w-4 h-4 mr-1" />
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
            :animation="200"
            :scroll="scrollRef || true"
            :bubble-scroll="true"
            :scroll-sensitivity="80"
            :scroll-speed="16"
            class="space-y-3"
          >
            <WorkoutExerciseItem
              v-for="(field, exerciseIndex) in exerciseFields"
              :key="field.key"
              :exercise-index="exerciseIndex"
              :field="(field as { value: WorkoutExerciseForm; key: string })"
              :exercise-options="exerciseOptions"
              :get-exercise-features="getExerciseFeatures"
              :get-exercise="getExercise"
              :drawer-scroll-ref="scrollRef"
              @exercise-selected="(v) => onExerciseSelected(exerciseIndex, field.value, v)"
              @remove="removeExercise(exerciseIndex)"
            />
          </VueDraggable>

          <FormField name="exercises">
            <FormItem>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>
      </div>

      <div class="px-6 py-4 border-t shrink-0 flex gap-2 justify-end">
        <Button variant="outline" type="button" @click="onCancel" :disabled="isPending">
          {{ $t("cancel") }}
        </Button>
        <Button type="button" @click="onSubmit" :disabled="isPending">
          {{ isPending ? $t("creating") : $t("workoutPlans.adHocWorkout.save") }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>
