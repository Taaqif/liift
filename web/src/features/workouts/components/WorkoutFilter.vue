<script setup lang="ts">
import { ref, watch } from "vue";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import ReferenceMultiSelect from "@/features/reference/components/ReferenceMultiSelect.vue";
import ExerciseMultiSelect from "@/features/exercises/components/ExerciseMultiSelect.vue";

export type WorkoutFilter = {
  search: string;
  exerciseFeatures: string[];
  exerciseIds: number[];
  muscleGroup: string[];
  equipment: string[];
};

const props = defineProps<{
  modelValue: WorkoutFilter;
}>();

const emits = defineEmits<{
  (e: "update:modelValue", filter: WorkoutFilter): void;
}>();

const searchInput = ref(props.modelValue.search);
const selectedExerciseFeatures = ref<string[]>([...props.modelValue.exerciseFeatures]);
const selectedExerciseIds = ref<number[]>([...props.modelValue.exerciseIds]);
const selectedMuscleGroups = ref<string[]>([...props.modelValue.muscleGroup]);
const selectedEquipment = ref<string[]>([...props.modelValue.equipment]);

watch(
  () => props.modelValue,
  (newValue) => {
    searchInput.value = newValue.search;
    selectedExerciseFeatures.value = [...newValue.exerciseFeatures];
    selectedExerciseIds.value = [...newValue.exerciseIds];
    selectedMuscleGroups.value = [...newValue.muscleGroup];
    selectedEquipment.value = [...newValue.equipment];
  },
  { deep: true },
);

const handleSearch = () => {
  emits("update:modelValue", {
    search: searchInput.value.trim(),
    exerciseFeatures: [...selectedExerciseFeatures.value],
    exerciseIds: [...selectedExerciseIds.value],
    muscleGroup: [...selectedMuscleGroups.value],
    equipment: [...selectedEquipment.value],
  });
};

const handleClear = () => {
  searchInput.value = "";
  selectedExerciseFeatures.value = [];
  selectedExerciseIds.value = [];
  selectedMuscleGroups.value = [];
  selectedEquipment.value = [];
  emits("update:modelValue", {
    search: "",
    exerciseFeatures: [],
    exerciseIds: [],
    muscleGroup: [],
    equipment: [],
  });
};
</script>

<template>
  <div class="flex flex-col gap-4 p-4 border rounded-lg bg-card">
    <div class="flex flex-col sm:flex-row gap-4">
      <div class="flex-1">
        <Input v-model="searchInput" :placeholder="$t('workouts.filter.searchPlaceholder')"
          @keyup.enter="handleSearch" />
      </div>
      <div class="flex-1">
        <ReferenceMultiSelect reference-type="exerciseFeature" v-model="selectedExerciseFeatures"
          :placeholder="$t('workouts.filter.exerciseTypePlaceholder')" />
      </div>
      <div class="flex-1">
        <ReferenceMultiSelect reference-type="muscleGroup" v-model="selectedMuscleGroups"
          :placeholder="$t('workouts.filter.muscleGroupPlaceholder')" />
      </div>
      <div class="flex-1">
        <ReferenceMultiSelect reference-type="equipment" v-model="selectedEquipment"
          :placeholder="$t('workouts.filter.equipmentPlaceholder')" />
      </div>
      <div class="flex-1">
        <ExerciseMultiSelect v-model="selectedExerciseIds" :placeholder="$t('workouts.filter.exercisePlaceholder')" />
      </div>
    </div>
    <div class="flex gap-2">
      <Button @click="handleSearch" class="flex-1 sm:flex-none">
        {{ $t("search") }}
      </Button>
      <Button variant="outline" @click="handleClear" class="flex-1 sm:flex-none">
        {{ $t("clear") }}
      </Button>
    </div>
  </div>
</template>
