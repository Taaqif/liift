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
import type { Workout } from "@/features/workouts/types";
import {
  workoutPlanFormSchema,
  createEmptyPlan,
  resizeWeeks,
} from "@/features/workout-plans/types";
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
import { ArrowLeft, Plus } from "lucide-vue-next";
import WorkoutListSelect from "@/features/workout-plans/components/WorkoutListSelect.vue";
import InlineWorkoutCreator from "@/features/workout-plans/components/InlineWorkoutCreator.vue";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const planId = computed(() => {
  const id = route.params.id;
  return id ? Number(id) : null;
});
const isEditMode = computed(() => !!planId.value);

const { plan, loading: planLoading } = useWorkoutPlan(planId);

const {
  createPlan,
  isPending: isCreating,
  error: createError,
} = useCreateWorkoutPlan();
const {
  updatePlan,
  isPending: isUpdating,
  error: updateError,
} = useUpdateWorkoutPlan();
const {
  deletePlan,
  isPending: isDeleting,
  error: deleteError,
} = useDeleteWorkoutPlan();

const { handleSubmit, resetForm, meta, values, setFieldValue, errors, submitCount } =
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
      workoutIds: [...(d.workoutIds ?? [])],
      description: d.description ?? "",
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

watch(
  plan,
  (p) => {
    if (p) populateForm(p);
  },
  { immediate: true },
);

watch(
  () => [values.numberOfWeeks, values.daysPerWeek],
  ([numWeeks, daysPerWeek]) => {
    const n = Number(numWeeks);
    const d = Number(daysPerWeek);
    const current = scheduleWeeks.value.length
      ? scheduleWeeks.value
      : values.weeks;
    if (
      n >= 1 &&
      d >= 1 &&
      current &&
      (current.length !== n || current[0]?.days.length !== d)
    ) {
      const resized = resizeWeeks(current, n, d);
      scheduleWeeks.value = resized;
      setFieldValue("weeks", resized);
      if (selectedWeek.value >= n) selectedWeek.value = 0;
    }
  },
);

function updateDay(
  weekIndex: number,
  dayIndex: number,
  patch: Partial<PlanDay>,
) {
  const source = scheduleWeeks.value.length
    ? scheduleWeeks.value
    : values.weeks;
  const weeks = (source ?? []).map((w, wi) => {
    if (wi !== weekIndex) return w;
    return {
      days: w.days.map((d, di) => (di === dayIndex ? { ...d, ...patch } : d)),
    };
  });
  scheduleWeeks.value = weeks;
  setFieldValue("weeks", weeks);
}

function setDayWorkoutIds(
  weekIndex: number,
  dayIndex: number,
  workoutIds: number[],
) {
  updateDay(weekIndex, dayIndex, { workoutIds });
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
          workoutIds: [...(d.workoutIds ?? [])],
          description: d.description?.trim() ?? "",
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

// Inline workout creation per day
const creatorTarget = ref<{ weekIndex: number; dayIndex: number } | null>(null);

function openCreator(weekIndex: number, dayIndex: number) {
  creatorTarget.value = { weekIndex, dayIndex };
}

function onWorkoutCreated(
  weekIndex: number,
  dayIndex: number,
  workout: Workout,
) {
  const current = scheduleWeeks.value[weekIndex]?.days[dayIndex];
  const existingIds = current?.workoutIds ?? [];
  setDayWorkoutIds(weekIndex, dayIndex, [...existingIds, workout.id]);
  creatorTarget.value = null;
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
const selectedWeek = ref(0);

onBeforeRouteLeave(() => {
  if (meta.value.dirty && !isPending.value) {
    return window.confirm(
      t("unsavedChanges.confirmLeave") ||
        "You have unsaved changes. Leave anyway?",
    );
  }
});
</script>

<template>
  <div ref="pageScrollRef">
    <div class="mb-8">
      <button class="text-sm text-muted-foreground hover:text-foreground transition-colors mb-1" @click="router.push({ name: 'workout-plans' })">
        ← {{ $t("nav.workoutPlans") }}
      </button>
      <h1 class="text-3xl font-bold tracking-tight">{{ title }}</h1>
      <p class="text-muted-foreground mt-1">{{ description }}</p>
    </div>

    <div
      v-if="isEditMode && planLoading"
      class="flex items-center justify-center py-24"
    >
      <div class="text-muted-foreground">{{ $t("loading") }}</div>
    </div>

    <div v-else class="space-y-6">
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
          <span class="text-base font-medium">{{
            $t("workoutPlans.schedule")
          }}</span>
          <div class="flex flex-wrap justify-center gap-2">
            <Button
              v-for="(_, weekIndex) in scheduleWeeks"
              :key="weekIndex"
              type="button"
              :variant="selectedWeek === weekIndex ? 'default' : 'outline'"
              class="min-w-12 h-10 text-base font-semibold"
              @click="selectedWeek = weekIndex"
            >
              {{ weekIndex + 1 }}
            </Button>
          </div>
          <div v-if="scheduleWeeks[selectedWeek]" class="grid gap-3">
                <div
                  v-for="(day, dayIndex) in scheduleWeeks[selectedWeek].days"
                  :key="dayIndex"
                  class="rounded border p-4 bg-muted/30 space-y-3"
                >
                  <h5 class="font-medium text-sm">
                    {{ $t("workoutPlans.dayLabel", { number: dayIndex + 1 }) }}
                  </h5>
                  <Textarea
                    :placeholder="$t('workoutPlans.dayDescriptionPlaceholder')"
                    rows="2"
                    :value="day.description ?? ''"
                    @input="
                      (e: Event) =>
                        updateDay(selectedWeek, dayIndex, {
                          description: (e.target as HTMLTextAreaElement).value,
                        })
                    "
                  />
                  <div class="space-y-1">
                    <span class="text-sm text-muted-foreground block">
                      {{ $t("workoutPlans.workoutsPerDay") }}
                    </span>
                    <p
                      v-if="
                        submitCount > 0 &&
                        errors[
                          `weeks[${selectedWeek}].days[${dayIndex}].workoutIds`
                        ]
                      "
                      class="text-sm font-medium text-destructive"
                    >
                      {{
                        errors[
                          `weeks[${selectedWeek}].days[${dayIndex}].workoutIds`
                        ]
                      }}
                    </p>
                    <WorkoutListSelect
                      :key="`week-${selectedWeek}-day-${dayIndex}`"
                      :model-value="day.workoutIds ?? []"
                      :placeholder="$t('workoutPlans.addWorkout')"
                      :scroll-ref="pageScrollRef"
                      @update:model-value="
                        (ids) => setDayWorkoutIds(selectedWeek, dayIndex, ids)
                      "
                    />
                    <Button
                      v-if="
                        !(
                          creatorTarget?.weekIndex === selectedWeek &&
                          creatorTarget?.dayIndex === dayIndex
                        )
                      "
                      type="button"
                      variant="outline"
                      size="sm"
                      class="w-full"
                      @click="openCreator(selectedWeek, dayIndex)"
                    >
                      <Plus class="w-4 h-4 mr-2" />
                      {{ $t("workoutPlans.adHocWorkout.create") }}
                    </Button>
                  </div>
                  <InlineWorkoutCreator
                    v-if="
                      creatorTarget?.weekIndex === selectedWeek &&
                      creatorTarget?.dayIndex === dayIndex
                    "
                    :save-to-library="false"
                    :scroll-ref="pageScrollRef"
                    @workout-created="
                      (w) => onWorkoutCreated(selectedWeek, dayIndex, w)
                    "
                    @close="creatorTarget = null"
                  />
                </div>
          </div>
        </div>
      </form>

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
        <Button
          variant="outline"
          @click="router.push({ name: 'workout-plans' })"
        >
          {{ $t("cancel") }}
        </Button>
      </div>
    </div>
  </div>
</template>
