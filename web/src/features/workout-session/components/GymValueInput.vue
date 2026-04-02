<script setup lang="ts">
import { ref, computed } from "vue";
import { Button } from "@/components/ui/button";
import { Minus, Plus } from "lucide-vue-next";

const KG_TO_LBS = 2.20462;
const KM_TO_MILES = 0.621371;

const props = withDefaults(
  defineProps<{
    featureName: string;
    modelValue: number;
    disabled?: boolean;
    label?: string;
  }>(),
  { disabled: false },
);

const emit = defineEmits<{
  (e: "update:modelValue", value: number): void;
  (e: "blur"): void;
}>();

const weightUnit = ref<"kg" | "lbs">("kg");
const distanceUnit = ref<"km" | "mi">("km");

const displayValue = computed(() => {
  if (props.featureName === "weight") {
    return weightUnit.value === "lbs" ? kgToLbs(props.modelValue) : props.modelValue;
  }
  if (props.featureName === "distance") {
    if (distanceUnit.value === "mi") {
      return Math.round(props.modelValue * KM_TO_MILES * 10) / 10;
    }
    return props.modelValue;
  }
  return props.modelValue;
});

function kgToLbs(kg: number): number {
  return Math.round(kg * KG_TO_LBS * 10) / 10;
}
function lbsToKg(lbs: number): number {
  return Math.round((lbs / KG_TO_LBS) * 100) / 100;
}

const step = computed(() => {
  if (props.featureName === "rep") return 1;
  if (props.featureName === "weight") {
    return weightUnit.value === "kg" ? 1.25 : 2.5;
  }
  if (props.featureName === "distance") {
    return distanceUnit.value === "km" ? 0.1 : 0.1;
  }
  return 1;
});

const min = computed(() => (props.featureName === "rep" ? 0 : 0));

function clamp(v: number): number {
  const m = min.value;
  if (v < m) return m;
  if (props.featureName === "rep") return Math.round(v);
  return v;
}

function stepUp() {
  const next = clamp(props.modelValue + step.value);
  emit("update:modelValue", props.featureName === "weight" ? Math.round(next * 100) / 100 : next);
}

function stepDown() {
  const next = clamp(props.modelValue - step.value);
  emit("update:modelValue", props.featureName === "weight" ? Math.round(next * 100) / 100 : next);
}

function onNumericInput(e: Event) {
  const raw = (e.target as HTMLInputElement).value;
  const num = parseFloat(raw);
  if (Number.isNaN(num)) return;
  if (props.featureName === "weight") {
    emit(
      "update:modelValue",
      weightUnit.value === "lbs" ? lbsToKg(num) : num,
    );
  } else if (props.featureName === "distance") {
    emit(
      "update:modelValue",
      distanceUnit.value === "mi" ? num / KM_TO_MILES : num,
    );
  } else if (props.featureName === "duration") {
    emit("update:modelValue", Math.max(0, num));
  }
}

const inputStep = computed(() => {
  if (props.featureName === "rep") return "1";
  if (props.featureName === "weight") return String(step.value);
  if (props.featureName === "distance") return "0.1";
  return "1";
});
</script>

<template>
  <div class="flex flex-col gap-1">
    <label v-if="label" class="text-xs text-muted-foreground whitespace-nowrap">
      {{ label }}
    </label>
    <div class="flex items-center gap-0.5 rounded-md border border-input bg-transparent overflow-hidden">
      <template v-if="featureName === 'rep'">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-8 w-8 shrink-0 rounded-none"
          :disabled="disabled || modelValue <= 0"
          @click="stepDown"
        >
          <Minus class="size-4" />
        </Button>
        <input
          type="number"
          inputmode="numeric"
          :value="modelValue"
          min="0"
          step="1"
          class="h-8 w-14 border-0 bg-transparent text-center text-sm focus:outline-none focus:ring-0 disabled:opacity-50 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
          :disabled="disabled"
          @input="(e) => { const n = parseInt((e.target as HTMLInputElement).value, 10); if (!Number.isNaN(n) && n >= 0) emit('update:modelValue', n); }"
          @blur="emit('blur')"
        />
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-8 w-8 shrink-0 rounded-none"
          :disabled="disabled"
          @click="stepUp"
        >
          <Plus class="size-4" />
        </Button>
      </template>
      <template v-else>
        <input
          type="number"
          inputmode="decimal"
          :value="featureName === 'weight' || featureName === 'distance' ? displayValue : modelValue"
          :min="min"
          :step="inputStep"
          class="h-8 min-w-0 flex-1 border-0 bg-transparent px-2 py-1 text-sm focus:outline-none focus:ring-0 disabled:opacity-50 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
          :disabled="disabled"
          @input="onNumericInput"
          @blur="emit('blur')"
        />
        <template v-if="featureName === 'weight'">
          <button
            type="button"
            class="shrink-0 px-2 py-1 text-xs font-medium text-muted-foreground hover:text-foreground border-l border-input"
            :disabled="disabled"
            @click="weightUnit = weightUnit === 'kg' ? 'lbs' : 'kg'"
          >
            {{ weightUnit }}
          </button>
        </template>
        <template v-else-if="featureName === 'distance'">
          <button
            type="button"
            class="shrink-0 px-2 py-1 text-xs font-medium text-muted-foreground hover:text-foreground border-l border-input"
            :disabled="disabled"
            @click="distanceUnit = distanceUnit === 'km' ? 'mi' : 'km'"
          >
            {{ distanceUnit === 'km' ? 'km' : 'mi' }}
          </button>
        </template>
        <template v-else-if="featureName === 'duration'">
          <span class="shrink-0 px-2 py-1 text-xs text-muted-foreground border-l border-input">
            sec
          </span>
        </template>
      </template>
    </div>
  </div>
</template>
