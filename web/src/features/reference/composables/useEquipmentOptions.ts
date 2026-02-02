import { computed, type ComputedRef } from "vue";
import { useI18n } from "vue-i18n";
import { useEquipment } from "./useEquipment";
import type { ReferenceOption } from "../types";

export function useEquipmentOptions(): {
  options: ComputedRef<ReferenceOption[]>;
  loading: ComputedRef<boolean>;
} {
  const { t } = useI18n();
  const { equipment, loading } = useEquipment();

  const options = computed<ReferenceOption[]>(() =>
    equipment.value
      .map((item) => ({
        value: item.name,
        label: t(`equipment.${item.name}`),
      }))
      .sort((a, b) => a.label.localeCompare(b.label)),
  );

  return { options, loading };
}
