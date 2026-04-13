<script setup lang="ts">
import { computed } from "vue";
import { generateId } from "@/lib/utils";
import { useFieldArray } from "vee-validate";
import type { WorkoutExerciseForm, WorkoutSetForm, WorkoutSetFeatureForm } from "../types";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Plus, Trash2, GripVertical } from "lucide-vue-next";
import { VueDraggable } from "vue-draggable-plus";
import ExerciseInfoDialog from "@/features/exercises/components/ExerciseInfoDialog.vue";
import type { Exercise } from "@/features/exercises/types";

const props = defineProps<{
  exerciseIndex: number;
  field: { value: WorkoutExerciseForm; key: string };
  exerciseOptions: { value: number; label: string }[];
  getExerciseFeatures: (exerciseId: number | null) => string[];
  getExercise: (exerciseId: number | null) => Exercise | undefined;
  drawerScrollRef: HTMLElement | null;
}>();

const emit = defineEmits<{
  (e: "remove"): void;
}>();

const setsPath = computed(
  () => `exercises.${props.exerciseIndex}.sets`,
);

const {
  fields: setFields,
  push: pushSet,
  remove: removeSet,
  move: moveSet,
  update: updateSet,
} = useFieldArray<WorkoutSetForm>(setsPath);

const getSetFeatureValue = (set: WorkoutSetForm, featureName: string): string => {
  const val = set.features?.find((f) => f.feature_name === featureName)?.value;
  return val != null ? String(val) : "";
};

const addSet = () => {
  const featureNames = props.getExerciseFeatures(
    props.field.value?.exercise_id ?? null,
  );
  pushSet({
    _key: generateId(),
    order: setFields.value.length,
    features: featureNames.map((name) => ({ feature_name: name, value: null })),
  });
};

const onSetsReorder = (event: { oldIndex?: number; newIndex?: number }) => {
  const oldIndex = event.oldIndex ?? 0;
  const newIndex = event.newIndex ?? 0;
  if (oldIndex === newIndex) return;
  moveSet(oldIndex, newIndex);
};

const updateFeatureValue = (
  setIndex: number,
  featureName: string,
  value: number | null,
) => {
  const set = setFields.value[setIndex]?.value;
  if (!set) return;
  const features = [...(set.features || [])];
  const existingIndex = features.findIndex(
    (f: WorkoutSetFeatureForm) => f.feature_name === featureName,
  );
  if (existingIndex >= 0) {
    const f = features[existingIndex];
    if (f) features[existingIndex] = { ...f, value };
  } else {
    features.push({ feature_name: featureName, value });
  }
  updateSet(setIndex, { ...set, features });
};

const setsForDraggable = computed(
  () => props.field.value?.sets ?? [],
);
</script>

<template>
  <div class="border rounded-lg overflow-hidden">
    <!-- Exercise header -->
    <div class="flex items-center gap-2 px-3 py-2.5 bg-muted/40 border-b">
      <button
        type="button"
        class="exercise-drag-handle p-1 rounded cursor-grab active:cursor-grabbing text-muted-foreground hover:text-foreground touch-none"
        tabindex="-1"
      >
        <GripVertical class="w-4 h-4" />
      </button>
      <span class="flex-1 text-sm font-semibold truncate">
        {{ exerciseOptions.find(o => o.value === field.value?.exercise_id)?.label ?? $t('exercises.title') }}
      </span>
      <ExerciseInfoDialog
        v-if="field.value?.exercise_id"
        :exercise="getExercise(field.value.exercise_id)"
      />
      <Button
        type="button"
        variant="ghost"
        size="icon"
        class="size-7 text-muted-foreground hover:text-destructive shrink-0"
        @click="emit('remove')"
      >
        <Trash2 class="w-3.5 h-3.5" />
      </Button>
    </div>

    <div class="p-3 space-y-3">
      <!-- Rest timer + note -->
      <div class="flex gap-3">
        <FormField
          v-slot="{ componentField }"
          :name="`exercises.${exerciseIndex}.rest_timer`"
        >
          <FormItem class="w-28 shrink-0">
            <FormLabel class="text-xs">{{ $t("workouts.restTimer") }}</FormLabel>
            <FormControl>
              <Input type="number" min="0" class="h-8 text-sm" v-bind="componentField" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField
          v-slot="{ componentField }"
          :name="`exercises.${exerciseIndex}.note`"
        >
          <FormItem class="flex-1">
            <FormLabel class="text-xs">{{ $t("workouts.note") }}</FormLabel>
            <FormControl>
              <Textarea
                :placeholder="$t('workouts.notePlaceholder')"
                rows="1"
                class="text-sm resize-none"
                v-bind="componentField"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>
      </div>

      <!-- Sets -->
      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <Label class="text-xs font-medium text-muted-foreground uppercase tracking-wide">{{ $t("workouts.sets") }}</Label>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            class="h-7 px-2 text-xs"
            @click="addSet"
            :disabled="!field.value?.exercise_id"
          >
            <Plus class="w-3.5 h-3.5 mr-1" />
            {{ $t("workouts.addSet") }}
          </Button>
        </div>
        <FormField :name="`exercises.${exerciseIndex}.sets`" v-slot>
          <FormItem>
            <FormMessage />
          </FormItem>
        </FormField>
        <VueDraggable
          :model-value="setsForDraggable"
          :custom-update="(e: { oldIndex?: number; newIndex?: number }) => onSetsReorder(e)"
          handle=".set-drag-handle"
          :force-fallback="true"
          :fallback-on-body="true"
          ghost-class="workout-drag-ghost-set"
          chosen-class="workout-drag-chosen"
          fallback-class="workout-drag-fallback-set"
          :animation="150"
          :scroll="drawerScrollRef || true"
          :bubble-scroll="true"
          :scroll-sensitivity="80"
          :scroll-speed="16"
          class="space-y-2"
        >
          <div
            v-for="(setField, setIndex) in setFields"
            :key="setField.key"
            class="flex items-center gap-2 bg-muted/30 rounded-md px-2 py-2"
          >
            <button
              type="button"
              class="set-drag-handle p-0.5 text-muted-foreground/50 hover:text-muted-foreground touch-none cursor-grab active:cursor-grabbing shrink-0"
              tabindex="-1"
            >
              <GripVertical class="w-3.5 h-3.5" />
            </button>
            <span class="text-xs font-medium text-muted-foreground w-8 shrink-0">
              {{ $t("workouts.setNumber", { number: setIndex + 1 }) }}
            </span>
            <div
              v-if="field.value?.exercise_id"
              class="flex flex-1 gap-2 min-w-0"
            >
              <FormField
                v-for="featureName in getExerciseFeatures(field.value.exercise_id)"
                :key="featureName"
                :name="`exercises.${exerciseIndex}.sets.${setIndex}.features.${getExerciseFeatures(field.value.exercise_id).indexOf(featureName)}.value`"
              >
                <FormItem class="flex-1 min-w-0">
                  <FormLabel class="text-xs text-muted-foreground capitalize sr-only">
                    {{ $t(`exerciseFeature.${featureName}`) }}
                  </FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      step="0.01"
                      min="0"
                      :placeholder="$t(`exerciseFeature.${featureName}`)"
                      class="h-8 text-sm"
                      :model-value="getSetFeatureValue(setField.value, featureName)"
                      @update:model-value="
                        (value: string | number) => {
                          const num = Number(value);
                          updateFeatureValue(
                            Number(setIndex),
                            featureName,
                            value === '' || Number.isNaN(num) ? null : num,
                          );
                        }
                      "
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              </FormField>
            </div>
            <Button
              v-if="setFields.length > 1"
              type="button"
              variant="ghost"
              size="icon"
              class="size-7 text-muted-foreground hover:text-destructive shrink-0"
              @click="removeSet(setIndex)"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </Button>
          </div>
        </VueDraggable>
      </div>
    </div>
  </div>
</template>
