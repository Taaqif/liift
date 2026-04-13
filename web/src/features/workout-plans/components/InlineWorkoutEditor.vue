<script setup lang="ts">
import { computed, ref, unref, watch } from "vue";
import { generateId } from "@/lib/utils";
import { useForm, useFieldArray } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { useI18n } from "vue-i18n";
import { useWorkout } from "@/features/workouts/composables/useWorkout";
import { useUpdateWorkout } from "@/features/workouts/composables/useUpdateWorkout";
import { useExercises } from "@/features/exercises/composables/useExercises";
import type { WorkoutExerciseForm, WorkoutFormValues } from "@/features/workouts/types";
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
import { Plus } from "lucide-vue-next";
import { VueDraggable } from "vue-draggable-plus";
import WorkoutExerciseItem from "@/features/workouts/components/WorkoutExerciseItem.vue";

const props = defineProps<{
  workoutId: number;
  scrollRef?: HTMLElement | null;
}>();

const emit = defineEmits<{
  "workout-updated": [];
  close: [];
}>();

const { t } = useI18n();
const { updateWorkout, isPending } = useUpdateWorkout();
const { exercises: allExercises } = useExercises({ limit: 1000 });
const { workout } = useWorkout(computed(() => props.workoutId));

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

function populateForm(w: typeof workout.value) {
  if (!w || w.id !== props.workoutId) return;
  resetForm({
    values: {
      name: w.name,
      description: w.description || "",
      exercises: (w.exercises || []).map((ex) => ({
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
      })),
    },
  });
}

watch(() => workout.value, (w) => { if (w) populateForm(w); }, { immediate: true });

const getExerciseFeatures = (exerciseId: number | null): string[] => {
  if (!exerciseId) return [];
  const exercise = allExercises.value.find((e) => e.id === exerciseId);
  return exercise?.exercise_features?.map((f) => f.name) ?? [];
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
      ? [{ _key: generateId(), order: 0, features: featureNames.map((name) => ({ feature_name: name, value: null })) }]
      : currentSets.map((set) => ({
          ...set,
          features: featureNames.map((name) => {
            const existing = set.features?.find((f) => f.feature_name === name);
            return existing ?? { feature_name: name, value: null };
          }),
        }));
  updateExercise(exerciseIndex, { ...exercise, exercise_id: exerciseId, sets: newSets });
};

const exercisesForDraggable = computed(() => (unref(values) as WorkoutFormValues)?.exercises ?? []);

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
        id: ex.id,
        exercise_id: ex.exercise_id,
        rest_timer: ex.rest_timer || 0,
        note: ex.note || "",
        order: idx,
        sets: ex.sets.map((set, setIdx) => ({
          id: set.id,
          order: setIdx,
          features: featureNames.map((name) => {
            const existing = set.features?.find((f) => f.feature_name === name);
            return { id: existing?.id, feature_name: name, value: existing?.value ?? 0 };
          }),
        })),
      };
    });

    await updateWorkout({
      id: props.workoutId,
      data: {
        name: formValues.name.trim(),
        description: formValues.description?.trim() || undefined,
        exercises,
      },
    });

    emit("workout-updated");
    emit("close");
  } catch {
    error.value = t("workouts.error");
  }
});
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

    <div class="flex gap-2 justify-end pt-2 border-t">
      <Button variant="outline" type="button" @click="emit('close')" :disabled="isPending">
        {{ $t("cancel") }}
      </Button>
      <Button type="button" @click="onSubmit" :disabled="isPending">
        {{ isPending ? $t("updating") : $t("workoutPlans.editWorkout.save") }}
      </Button>
    </div>
  </div>
</template>
