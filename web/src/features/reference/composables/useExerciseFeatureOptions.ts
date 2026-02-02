import { computed, type ComputedRef } from "vue";
import { useI18n } from "vue-i18n";
import { useExerciseFeature } from "./useExerciseFeature";
import type { ReferenceOption } from "../types";

export function useExerciseFeatureOptions(): {
  options: ComputedRef<ReferenceOption[]>;
  loading: ComputedRef<boolean>;
} {
  const { t } = useI18n();
  const { exerciseFeatures, loading } = useExerciseFeature();

  const options = computed<ReferenceOption[]>(() =>
    exerciseFeatures.value
      .map((item) => ({
        value: item.name,
        label: t(`exerciseFeature.${item.name}`),
      }))
      .sort((a, b) => a.label.localeCompare(b.label)),
  );

  return { options, loading };
}
