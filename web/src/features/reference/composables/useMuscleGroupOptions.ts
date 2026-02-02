import { computed, type ComputedRef } from "vue";
import { useI18n } from "vue-i18n";
import { useMuscleGroup } from "./useMuscleGroup";
import type { ReferenceOption } from "../types";

export function useMuscleGroupOptions(): {
  options: ComputedRef<ReferenceOption[]>;
  loading: ComputedRef<boolean>;
} {
  const { t } = useI18n();
  const { muscleGroup, loading } = useMuscleGroup();

  const options = computed<ReferenceOption[]>(() =>
    muscleGroup.value
      .map((item) => ({
        value: item.name,
        label: t(`muscleGroup.${item.name}`),
      }))
      .sort((a, b) => a.label.localeCompare(b.label)),
  );

  return { options, loading };
}
