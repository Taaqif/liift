<script setup lang="ts">
import { ref, watch } from "vue";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import { useMuscleGroupOptions } from "@/features/reference/composables/useMuscleGroupOptions";
import { useEquipmentOptions } from "@/features/reference/composables/useEquipmentOptions";

export type ExerciseFilter = {
  search: string;
  muscleGroup: string[];
  equipment: string[];
};

const props = defineProps<{
  modelValue: ExerciseFilter;
}>();

const emits = defineEmits<{
  (e: "update:modelValue", filter: ExerciseFilter): void;
}>();


const searchInput = ref(props.modelValue.search);
const selectedMuscleGroups = ref<string[]>([...props.modelValue.muscleGroup]);
const selectedEquipment = ref<string[]>([...props.modelValue.equipment]);

watch(
  () => props.modelValue,
  (newValue) => {
    searchInput.value = newValue.search;
    selectedMuscleGroups.value = [...newValue.muscleGroup];
    selectedEquipment.value = [...newValue.equipment];
  },
  { deep: true },
);
const { options: muscleGroupOptions } = useMuscleGroupOptions();
const { options: equipmentOptions } = useEquipmentOptions();


const handleSearch = () => {
  emits("update:modelValue", {
    search: searchInput.value.trim(),
    muscleGroup: [...selectedMuscleGroups.value],
    equipment: [...selectedEquipment.value],
  });
};

const handleClear = () => {
  searchInput.value = "";
  selectedMuscleGroups.value = [];
  selectedEquipment.value = [];
  emits("update:modelValue", {
    search: "",
    muscleGroup: [],
    equipment: [],
  });
};
</script>

<template>
  <div class="flex flex-col gap-4 p-4 border rounded-lg bg-card">
    <div class="flex flex-col gap-4">
      <div>
        <Input v-model="searchInput" :placeholder="$t('exercises.filter.searchPlaceholder')"
          @keyup.enter="handleSearch" />
      </div>

      <div>
        <MultiSelectTags v-model="selectedMuscleGroups" :options="muscleGroupOptions"
          :placeholder="$t('exercises.filter.muscleGroupPlaceholder')" :list-mode="true" />
      </div>
      <div>
        <MultiSelectTags v-model="selectedEquipment" :options="equipmentOptions"
          :placeholder="$t('exercises.filter.equipmentPlaceholder')" :list-mode="true" />
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
