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
import type { Exercise } from "@/features/exercises/types";
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
import ExercisePickerSheet from "@/features/exercises/components/ExercisePickerSheet.vue";

const props = defineProps<{
  saveToLibrary?: boolean;
  scrollRef?: HTMLElement | null;
}>();

const emit = defineEmits<{
  "workout-created": [workout: Workout];
  close: [];
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
} = useFieldArray<WorkoutExerciseForm>("exercises");

const getExerciseFeatures = (exerciseId: number | null): string[] => {
  if (!exerciseId) return [];
  const exercise = allExercises.value.find((e) => e.id === exerciseId);
  return exercise?.exercise_features?.map((f) => f.name) ?? [];
};

const getExercise = (exerciseId: number | null): Exercise | undefined => {
  if (!exerciseId) return undefined;
  return allExercises.value.find((e) => e.id === exerciseId);
};

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
      sets: [{ _key: generateId(), order: 0, features: featureNames.map((name) => ({ feature_name: name, value: null })) }],
    });
  });
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
            return { feature_name: name, value: existing?.value ?? null };
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
    emit("close");
  } catch {
    error.value = t("workouts.error");
  }
});

function onCancel() {
  resetForm();
  error.value = null;
  emit("close");
}
</script>

<template>
  <div class="border-t px-3 py-4 space-y-4 bg-muted/20">
    <div v-if="error" class="p-3 bg-destructive/10 text-destructive rounded-lg text-sm">
      {{ error }}
    </div>

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
          <Textarea :placeholder="$t('workouts.descriptionPlaceholder')" rows="2" v-bind="componentField" />
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
          @remove="removeExercise(exerciseIndex)"
        />
      </VueDraggable>

      <FormField name="exercises">
        <FormItem>
          <FormMessage />
        </FormItem>
      </FormField>
    </div>

    <div class="flex gap-2 justify-end pt-2 border-t">
      <Button variant="outline" type="button" @click="onCancel" :disabled="isPending">
        {{ $t("cancel") }}
      </Button>
      <Button type="button" @click="onSubmit" :disabled="isPending">
        {{ isPending ? $t("creating") : $t("workoutPlans.adHocWorkout.save") }}
      </Button>
    </div>
  </div>

  <ExercisePickerSheet v-model:open="showExercisePicker" @add="handleExercisesAdded" />
</template>
