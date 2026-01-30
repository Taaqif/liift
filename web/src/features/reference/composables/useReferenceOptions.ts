import { computed, type ComputedRef } from "vue";
import { useExerciseFeature } from "./useExerciseFeature";
import { useMuscleGroup } from "./useMuscleGroup";
import { useEquipment } from "./useEquipment";
import type { ReferenceOption, ReferenceType } from "../types";

export function useReferenceOptions(type: ReferenceType): {
  options: ComputedRef<ReferenceOption[]>;
  loading: ComputedRef<boolean>;
} {
  const { exerciseFeatures, loading: loadingFeatures } = useExerciseFeature();
  const { muscleGroup, loading: loadingMuscle } = useMuscleGroup();
  const { equipment, loading: loadingEquipment } = useEquipment();

  const options = computed<ReferenceOption[]>(() => {
    const items =
      type === "exerciseFeature"
        ? exerciseFeatures.value
        : type === "muscleGroup"
          ? muscleGroup.value
          : equipment.value;
    return items
      .map((item) => ({
        value: item.name,
        label:
          type === "exerciseFeature"
            ? item.name.charAt(0).toUpperCase() + item.name.slice(1)
            : item.name,
      }))
      .sort((a, b) => a.label.localeCompare(b.label));
  });

  const loading = computed(
    () =>
      (type === "exerciseFeature" && loadingFeatures.value) ||
      (type === "muscleGroup" && loadingMuscle.value) ||
      (type === "equipment" && loadingEquipment.value),
  );

  return { options, loading };
}
