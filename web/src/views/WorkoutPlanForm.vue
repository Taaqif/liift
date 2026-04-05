<script setup lang="ts">
import { computed, watch, ref } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { useI18n } from "vue-i18n";
import { useWorkoutPlan } from "@/features/workout-plans/composables/useWorkoutPlan";
import { useCreateWorkoutPlan } from "@/features/workout-plans/composables/useCreateWorkoutPlan";
import { useUpdateWorkoutPlan } from "@/features/workout-plans/composables/useUpdateWorkoutPlan";
import { useDeleteWorkoutPlan } from "@/features/workout-plans/composables/useDeleteWorkoutPlan";
import type {
  WorkoutPlan,
  WorkoutPlanFormValues,
  PlanDay,
  PlanWeek,
} from "@/features/workout-plans/types";
import {
  workoutPlanFormSchema,
  createEmptyPlan,
  resizeWeeks,
} from "@/features/workout-plans/types";
import type { Workout } from "@/features/workouts/types";
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
import { ArrowLeft, Plus } from "lucide-vue-next";
import WorkoutListSelect from "@/features/workout-plans/components/WorkoutListSelect.vue";
import AdHocWorkoutDialog from "@/features/workout-plans/components/AdHocWorkoutDialog.vue";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const planId = computed(() => {
  const id = route.params.id;
  return id ? Number(id) : null;
});
const isEditMode = computed(() => !!planId.value);

const { plan, loading: planLoading } = useWorkoutPlan(planId);

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

function populateForm(p: WorkoutPlan) {
  const weeks = p.weeks.map((w) => ({
    days: w.days.map((d) => ({
      isRest: d.isRest,
      workoutIds: [...d.workoutIds],
    })),
  }));
  resetForm({
    values: {
      name: p.name,
      description: p.description ?? "",
      numberOfWeeks: p.numberOfWeeks,
      daysPerWeek: p.daysPerWeek,
      weeks,
    },
  });
  scheduleWeeks.value = weeks;
}

watch(plan, (p) => {
  if (p) populateForm(p);
}, { immediate: true });

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

function updateDay(weekIndex: number, dayIndex: number, patch: Partial<PlanDay>) {
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

function setDayWorkoutIds(weekIndex: number, dayIndex: number, workoutIds: number[]) {
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
    if (isEditMode.value && planId.value) {
      await updatePlan(planId.value, payload);
    } else {
      await createPlan(payload);
    }
    resetForm();
    router.push({ name: "workout-plans" });
  } catch (err) {
    console.error("Failed to save plan:", err);
  }
});

async function onDelete() {
  if (!planId.value) return;
  try {
    await deletePlan(planId.value);
    showDeleteDialog.value = false;
    router.push({ name: "workout-plans" });
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

// Ad-hoc workout creation per day
const adHocDialogOpen = ref(false);
const adHocTarget = ref<{ weekIndex: number; dayIndex: number } | null>(null);

function openAdHocDialog(weekIndex: number, dayIndex: number) {
  adHocTarget.value = { weekIndex, dayIndex };
  adHocDialogOpen.value = true;
}

function onAdHocWorkoutCreated(workout: Workout) {
  if (!adHocTarget.value) return;
  const { weekIndex, dayIndex } = adHocTarget.value;
  const current = scheduleWeeks.value[weekIndex]?.days[dayIndex];
  const existingIds = current?.workoutIds ?? [];
  setDayWorkoutIds(weekIndex, dayIndex, [...existingIds, workout.id]);
}

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

const pageScrollRef = ref<HTMLElement | null>(null);

onBeforeRouteLeave(() => {
  if (meta.value.dirty && !isPending.value) {
    return window.confirm(t("unsavedChanges.confirmLeave") || "You have unsaved changes. Leave anyway?");
  }
});
</script>

<template>
  <div ref="pageScrollRef">
    <div class="mb-8 flex items-center gap-4">
      <Button variant="ghost" size="icon" @click="router.push({ name: 'workout-plans' })">
        <ArrowLeft class="h-4 w-4" />
      </Button>
      <div>
        <h1 class="text-3xl font-bold">{{ title }}</h1>
        <p class="text-muted-foreground">{{ description }}</p>
      </div>
    </div>

    <div v-if="isEditMode && planLoading" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">{{ $t("loading") }}</div>
    </div>

    <div v-else class="space-y-6">
      <div v-if="error" class="p-4 bg-destructive/10 text-destructive rounded-lg">
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
                <Input type="number" min="1" max="52" v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
          <FormField v-slot="{ componentField }" name="daysPerWeek">
            <FormItem>
              <FormLabel>{{ $t("workoutPlans.daysPerWeek") }}</FormLabel>
              <FormControl>
                <Input type="number" min="1" max="14" v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>

        <div class="space-y-4">
          <span class="text-base font-medium">{{ $t("workoutPlans.schedule") }}</span>
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
                      @update:model-value="(v: boolean | 'indeterminate') => setDayRest(weekIndex, dayIndex, v === true)"
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
                      :scroll-ref="pageScrollRef"
                      @update:model-value="(ids) => setDayWorkoutIds(weekIndex, dayIndex, ids)"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      class="w-full"
                      @click="openAdHocDialog(weekIndex, dayIndex)"
                    >
                      <Plus class="w-4 h-4 mr-2" />
                      {{ $t("workoutPlans.adHocWorkout.create") }}
                    </Button>
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

      <AdHocWorkoutDialog
        v-model:open="adHocDialogOpen"
        :day-label="adHocTarget
          ? $t('workoutPlans.adHocWorkout.dayLabel', {
              week: adHocTarget.weekIndex + 1,
              day: adHocTarget.dayIndex + 1,
            })
          : ''"
        @workout-created="onAdHocWorkoutCreated"
      />

      <div class="flex flex-col gap-2 pt-4 border-t">
        <Dialog v-if="isEditMode" v-model:open="showDeleteDialog">
          <DialogTrigger as-child>
            <Button type="button" variant="destructive" :disabled="isPending">
              {{ $t("delete") }}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{{ $t("delete") }}</DialogTitle>
              <DialogDescription>{{ $t("areYouSure") }}</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="outline" @click="showDeleteDialog = false" :disabled="isDeleting">
                {{ $t("cancel") }}
              </Button>
              <Button variant="destructive" @click="onDelete" :disabled="isDeleting">
                {{ isDeleting ? $t("deleting") : $t("delete") }}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
        <Button @click="onSubmit" :disabled="isPending" class="flex-1">
          {{ submitButtonText }}
        </Button>
        <Button variant="outline" @click="router.push({ name: 'workout-plans' })">
          {{ $t("cancel") }}
        </Button>
      </div>
    </div>
  </div>
</template>
