<script setup lang="ts">
import { computed, watch, ref } from "vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { useI18n } from "vue-i18n";
import { useCreateWorkoutPlan } from "../composables/useCreateWorkoutPlan";
import { useUpdateWorkoutPlan } from "../composables/useUpdateWorkoutPlan";
import { useDeleteWorkoutPlan } from "../composables/useDeleteWorkoutPlan";
import type {
  WorkoutPlan,
  WorkoutPlanFormValues,
  PlanDay,
  PlanWeek,
} from "../types";
import {
  workoutPlanFormSchema,
  createEmptyPlan,
  resizeWeeks,
} from "../types";
import {
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Checkbox } from "@/components/ui/checkbox";
import WorkoutListSelect from "./WorkoutListSelect.vue";

const props = defineProps<{
  open?: boolean;
  plan?: WorkoutPlan | null;
}>();

const emits = defineEmits<{
  (e: "plan-created"): void;
  (e: "plan-updated"): void;
  (e: "plan-deleted"): void;
  (e: "form-dirty", value: boolean): void;
}>();

const { t } = useI18n();
const isEditMode = computed(() => !!props.plan);

const { createPlan, isPending: isCreating, error: createError } =
  useCreateWorkoutPlan();
const { updatePlan, isPending: isUpdating, error: updateError } =
  useUpdateWorkoutPlan();
const { deletePlan, isPending: isDeleting, error: deleteError } =
  useDeleteWorkoutPlan();

const { handleSubmit, resetForm, meta, values, setFieldValue } =
  useForm<WorkoutPlanFormValues>({
    validationSchema: toTypedSchema(workoutPlanFormSchema),
    initialValues: {
      name: "",
      description: "",
      numberOfWeeks: 3,
      daysPerWeek: 4,
      weeks: createEmptyPlan(3, 4),
    },
  });

const scheduleWeeks = ref<PlanWeek[]>(createEmptyPlan(3, 4));

const isFormDirty = computed(() => meta.value.dirty);

watch(
  [isFormDirty, () => props.open],
  ([dirty, isOpen]) => {
    if (isOpen) {
      emits("form-dirty", !!dirty);
    }
  },
  { immediate: true },
);

function populateForm(plan: WorkoutPlan | null) {
  if (plan) {
    const weeks = plan.weeks.map((w) => ({
      days: w.days.map((d) => ({
        isRest: d.isRest,
        workoutIds: [...d.workoutIds],
      })),
    }));
    resetForm({
      values: {
        name: plan.name,
        description: plan.description ?? "",
        numberOfWeeks: plan.numberOfWeeks,
        daysPerWeek: plan.daysPerWeek,
        weeks,
      },
    });
    scheduleWeeks.value = weeks;
  }
}

watch(
  () => props.plan,
  (plan) => {
    if (plan && props.open && isEditMode.value) {
      populateForm(plan);
    }
  },
  { immediate: true },
);

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen && props.plan) {
      populateForm(props.plan);
    } else if (!isOpen) {
      resetForm();
      showDeleteDialog.value = false;
    } else if (isOpen && !props.plan) {
      const initialWeeks = createEmptyPlan(3, 4);
      resetForm({
        values: {
          name: "",
          description: "",
          numberOfWeeks: 3,
          daysPerWeek: 4,
          weeks: initialWeeks,
        },
      });
      scheduleWeeks.value = initialWeeks;
    }
  },
);

watch(
  () => [values.numberOfWeeks, values.daysPerWeek],
  ([numWeeks, daysPerWeek]) => {
    const n = Number(numWeeks);
    const d = Number(daysPerWeek);
    const current = scheduleWeeks.value.length ? scheduleWeeks.value : values.weeks;
    if (
      n >= 1 &&
      d >= 1 &&
      current &&
      (current.length !== n || current[0]?.days.length !== d)
    ) {
      const resized = resizeWeeks(current, n, d);
      scheduleWeeks.value = resized;
      setFieldValue("weeks", resized);
    }
  },
);

function updateDay(
  weekIndex: number,
  dayIndex: number,
  patch: Partial<PlanDay>,
) {
  const source = scheduleWeeks.value.length ? scheduleWeeks.value : values.weeks;
  const weeks = (source ?? []).map((w, wi) => {
    if (wi !== weekIndex) return w;
    return {
      days: w.days.map((d, di) =>
        di === dayIndex ? { ...d, ...patch } : d,
      ),
    };
  });
  scheduleWeeks.value = weeks;
  setFieldValue("weeks", weeks);
}

function setDayRest(weekIndex: number, dayIndex: number, isRest: boolean) {
  const current =
    scheduleWeeks.value[weekIndex]?.days[dayIndex] ??
    values.weeks[weekIndex]?.days[dayIndex];
  updateDay(weekIndex, dayIndex, {
    isRest,
    workoutIds: isRest ? [] : [...(current?.workoutIds ?? [])],
  });
}

function setDayWorkoutIds(
  weekIndex: number,
  dayIndex: number,
  workoutIds: number[],
) {
  updateDay(weekIndex, dayIndex, {
    isRest: workoutIds.length === 0,
    workoutIds,
  });
}

const onSubmit = handleSubmit(async (formValues) => {
  try {
    const payload = {
      name: formValues.name.trim(),
      description: formValues.description?.trim() ?? "",
      numberOfWeeks: formValues.numberOfWeeks,
      daysPerWeek: formValues.daysPerWeek,
      weeks: formValues.weeks.map((w) => ({
        days: w.days.map((d) => ({
          isRest: d.isRest,
          workoutIds: [...(d.workoutIds ?? [])],
        })),
      })),
    };
    if (isEditMode.value && props.plan) {
      await updatePlan(props.plan.id, payload);
      resetForm();
      emits("plan-updated");
    } else {
      await createPlan(payload);
      resetForm();
      emits("plan-created");
    }
  } catch (err) {
    console.error("Failed to save plan:", err);
  }
});

async function onDelete() {
  if (!props.plan) return;
  try {
    await deletePlan(props.plan.id);
    resetForm();
    showDeleteDialog.value = false;
    emits("plan-deleted");
  } catch (err) {
    console.error("Failed to delete plan:", err);
  }
}

const error = computed(
  () => createError.value || updateError.value || deleteError.value,
);
const isPending = computed(
  () => isCreating.value || isUpdating.value || isDeleting.value,
);
const showDeleteDialog = ref(false);

const title = computed(() =>
  isEditMode.value ? t("workoutPlans.editTitle") : t("workoutPlans.createNew"),
);
const description = computed(() =>
  isEditMode.value
    ? t("workoutPlans.editDescription")
    : t("workoutPlans.createDescription"),
);
const submitButtonText = computed(() => {
  if (isEditMode.value) {
    return isUpdating.value ? t("updating") : t("workoutPlans.update");
  }
  return isCreating.value ? t("creating") : t("workoutPlans.create");
});

const drawerScrollRef = ref<HTMLElement | null>(null);
</script>

<template>
  <DrawerContent class="!max-h-[95vh]">
    <div
      ref="drawerScrollRef"
      class="mx-auto w-full max-w-4xl overflow-y-auto"
    >
      <DrawerHeader>
        <DrawerTitle>{{ title }}</DrawerTitle>
        <DrawerDescription>
          {{ description }}
        </DrawerDescription>
      </DrawerHeader>
      <div class="p-4 pb-0 space-y-6">
        <div
          v-if="error"
          class="p-4 bg-destructive/10 text-destructive rounded-lg"
        >
          <p>{{ $t("workoutPlans.error") }}: {{ error.message }}</p>
        </div>

        <form @submit="onSubmit" class="space-y-6">
          <FormField v-slot="{ componentField }" name="name">
            <FormItem>
              <FormLabel>{{ $t("workoutPlans.name") }}</FormLabel>
              <FormControl>
                <Input
                  :placeholder="$t('workoutPlans.namePlaceholder')"
                  v-bind="componentField"
                  required
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="description">
            <FormItem>
              <FormLabel>{{ $t("workoutPlans.description") }}</FormLabel>
              <FormControl>
                <Textarea
                  :placeholder="$t('workoutPlans.descriptionPlaceholder')"
                  rows="2"
                  v-bind="componentField"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <div class="grid grid-cols-2 gap-4">
            <FormField v-slot="{ componentField }" name="numberOfWeeks">
              <FormItem>
                <FormLabel>{{ $t("workoutPlans.numberOfWeeks") }}</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="1"
                    max="52"
                    v-bind="componentField"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
            <FormField v-slot="{ componentField }" name="daysPerWeek">
              <FormItem>
                <FormLabel>{{ $t("workoutPlans.daysPerWeek") }}</FormLabel>
                <FormControl>
                  <Input
                    type="number"
                    min="1"
                    max="14"
                    v-bind="componentField"
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>

          <div class="space-y-4">
            <span class="text-base font-medium">
              {{ $t("workoutPlans.schedule") }}
            </span>
            <div class="space-y-6">
              <div
                v-for="(week, weekIndex) in scheduleWeeks"
                :key="weekIndex"
                class="rounded-lg border p-4 space-y-3"
              >
                <h4 class="font-medium text-sm text-muted-foreground">
                  {{ $t("workoutPlans.weekLabel", { number: weekIndex + 1 }) }}
                </h4>
                <div class="grid gap-3">
                  <div
                    v-for="(day, dayIndex) in week.days"
                    :key="dayIndex"
                    class="rounded border p-4 bg-muted/30 space-y-3"
                  >
                    <h5 class="font-medium text-sm">
                      {{ $t("workoutPlans.dayLabel", { number: dayIndex + 1 }) }}
                    </h5>
                    <div class="flex items-center gap-2">
                      <Checkbox
                        :id="`rest-${weekIndex}-${dayIndex}`"
                        :model-value="day.isRest"
                        @update:model-value="
                          (v: boolean | 'indeterminate') =>
                            setDayRest(weekIndex, dayIndex, v === true)
                        "
                      />
                      <label
                        :for="`rest-${weekIndex}-${dayIndex}`"
                        class="text-sm font-medium cursor-pointer select-none"
                      >
                        {{ $t("workoutPlans.restDay") }}
                      </label>
                    </div>
                    <template v-if="!day.isRest">
                      <span class="text-sm text-muted-foreground block">
                        {{ $t("workoutPlans.workoutsPerDay") }}
                      </span>
                      <WorkoutListSelect
                        :key="`week-${weekIndex}-day-${dayIndex}`"
                        :model-value="day.workoutIds ?? []"
                        :placeholder="$t('workoutPlans.addWorkout')"
                        :scroll-ref="drawerScrollRef"
                        @update:model-value="
                          (ids) =>
                            setDayWorkoutIds(weekIndex, dayIndex, ids)
                        "
                      />
                    </template>
                  </div>
                </div>
              </div>
            </div>
            <FormField name="weeks">
              <FormItem>
                <FormMessage />
              </FormItem>
            </FormField>
          </div>
        </form>
      </div>
      <DrawerFooter class="flex-col gap-2 justify-between">
        <Dialog v-if="isEditMode" v-model:open="showDeleteDialog">
          <DialogTrigger as-child>
            <Button type="button" variant="destructive" :disabled="isPending">
              {{ $t("delete") }}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{{ $t("delete") }}</DialogTitle>
              <DialogDescription>
                {{ $t("areYouSure") }}
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="outline"
                @click="showDeleteDialog = false"
                :disabled="isDeleting"
              >
                {{ $t("cancel") }}
              </Button>
              <Button
                variant="destructive"
                @click="onDelete"
                :disabled="isDeleting"
              >
                {{ isDeleting ? $t("deleting") : $t("delete") }}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        <Button @click="onSubmit" :disabled="isPending" class="flex-1">
          {{ submitButtonText }}
        </Button>
        <DrawerClose as-child>
          <Button variant="outline">{{ $t("cancel") }}</Button>
        </DrawerClose>
      </DrawerFooter>
    </div>
  </DrawerContent>
</template>
