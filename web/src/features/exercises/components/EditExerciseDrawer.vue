<script setup lang="ts">
import { computed, watch, ref } from "vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useUpdateExercise } from "../composables/useUpdateExercise";
import { useDeleteExercise } from "../composables/useDeleteExercise";
import { useMuscleGroup } from "@/features/reference/composables/useMuscleGroup";
import { useEquipment } from "@/features/reference/composables/useEquipment";
import type { Exercise } from "../types";
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
import { MultiSelectTags } from "@/components/ui/multi-select-tags";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

const props = defineProps<{
  open?: boolean;
  modal?: boolean;
  exercise: Exercise | null;
}>();

const emits = defineEmits<{
  (e: "exercise-updated"): void;
  (e: "exercise-deleted"): void;
  (e: "form-dirty", value: boolean): void;
}>();

const { t } = useI18n();
const {
  updateExercise,
  isPending: isUpdating,
  error: updateError,
} = useUpdateExercise();
const {
  deleteExercise,
  isPending: isDeleting,
  error: deleteError,
} = useDeleteExercise();
const { muscleGroup } = useMuscleGroup();
const { equipment } = useEquipment();

const formSchema = z.object({
  name: z.string().min(1, t("exercises.validation.nameRequired")),
  description: z.string().optional(),
  primary_muscle_groups: z
    .array(z.string())
    .min(1, t("exercises.validation.primaryMuscleGroupsRequired")),
  secondary_muscle_groups: z.array(z.string()).optional(),
  equipment: z
    .array(z.string())
    .min(1, t("exercises.validation.equipmentRequired")),
});

const { handleSubmit, resetForm, setValues, meta } = useForm({
  validationSchema: toTypedSchema(formSchema),
  initialValues: {
    name: "",
    description: "",
    primary_muscle_groups: [] as string[],
    secondary_muscle_groups: [] as string[],
    equipment: [] as string[],
  },
});

const isFormDirty = computed(() => meta.value.dirty);

watch(
  [isFormDirty, () => props.open],
  ([dirty, isOpen]) => {
    if (isOpen) {
      emits("form-dirty", dirty);
    }
  },
  { immediate: true },
);

const muscleGroupOptions = computed(() =>
  muscleGroup.value.map((group) => ({
    value: group.name,
    label: group.name,
  })),
);

const equipmentOptions = computed(() =>
  equipment.value.map((group) => ({
    value: group.name,
    label: group.name,
  })),
);

const populateForm = (exercise: Exercise | null) => {
  if (exercise) {
    // Use resetForm with values to set the form values AND reset the dirty state
    resetForm({
      values: {
        name: exercise.name,
        description: exercise.description || "",
        primary_muscle_groups: exercise.primary_muscle_groups.map(
          (mg) => mg.name,
        ),
        secondary_muscle_groups: exercise.secondary_muscle_groups.map(
          (mg) => mg.name,
        ),
        equipment: exercise.equipment.map((eq) => eq.name),
      },
    });
  }
};

watch(
  () => props.exercise,
  (exercise) => {
    if (exercise && props.open) {
      populateForm(exercise);
    }
  },
  { immediate: true },
);

// Populate form when drawer opens (in case exercise is already set)
watch(
  () => props.open,
  (newValue) => {
    if (newValue && props.exercise) {
      // Drawer opened and exercise is available, populate form
      populateForm(props.exercise);
    } else if (!newValue) {
      // Drawer closed, reset form
      resetForm();
      showDeleteDialog.value = false;
    }
  },
);

const onSubmit = handleSubmit(async (values) => {
  if (!props.exercise) return;

  try {
    await updateExercise({
      id: props.exercise.id,
      data: {
        name: values.name.trim(),
        description: values.description?.trim() || undefined,
        primary_muscle_groups: values.primary_muscle_groups,
        secondary_muscle_groups:
          values.secondary_muscle_groups &&
            values.secondary_muscle_groups.length > 0
            ? values.secondary_muscle_groups
            : undefined,
        equipment: values.equipment,
      },
    });
    resetForm();
    emits("exercise-updated");
  } catch (err) {
    console.error("Failed to update exercise:", err);
  }
});

const onDelete = async () => {
  if (!props.exercise) return;

  try {
    await deleteExercise(props.exercise.id);
    resetForm();
    showDeleteDialog.value = false;
    emits("exercise-deleted");
  } catch (err) {
    console.error("Failed to delete exercise:", err);
  }
};

const error = computed(() => updateError.value || deleteError.value);
const isPending = computed(() => isUpdating.value || isDeleting.value);
const showDeleteDialog = ref(false);
</script>

<template>
  <DrawerContent class="max-h-[95vh]">
    <div class="mx-auto w-full max-w-2xl overflow-y-auto">
      <DrawerHeader>
        <DrawerTitle>{{ $t("exercises.editTitle") }}</DrawerTitle>
        <DrawerDescription>
          {{ $t("exercises.editDescription") }}
        </DrawerDescription>
      </DrawerHeader>
      <div class="p-4 pb-0 space-y-6">
        <div v-if="error" class="p-4 bg-destructive/10 text-destructive rounded-lg">
          <p>{{ $t("exercises.error") }}: {{ error.message }}</p>
        </div>

        <form @submit="onSubmit" class="space-y-4">
          <FormField v-slot="{ componentField }" name="name">
            <FormItem>
              <FormLabel>{{ $t("exercises.name") }}</FormLabel>
              <FormControl>
                <Input :placeholder="$t('exercises.namePlaceholder')" v-bind="componentField" required />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="description">
            <FormItem>
              <FormLabel>{{ $t("exercises.description") }}</FormLabel>
              <FormControl>
                <Textarea :placeholder="$t('exercises.descriptionPlaceholder')" rows="3" v-bind="componentField" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="primary_muscle_groups">
            <FormItem>
              <FormLabel>{{ $t("exercises.primaryMuscleGroups") }}</FormLabel>
              <FormControl>
                <MultiSelectTags :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']" :options="muscleGroupOptions"
                  :placeholder="$t('exercises.primaryMuscleGroupsPlaceholder')" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="secondary_muscle_groups">
            <FormItem>
              <FormLabel>{{ $t("exercises.secondaryMuscleGroups") }}</FormLabel>
              <FormControl>
                <MultiSelectTags :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']" :options="muscleGroupOptions"
                  :placeholder="$t('exercises.secondaryMuscleGroupsPlaceholder')
                    " />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="equipment">
            <FormItem>
              <FormLabel>{{ $t("exercises.equipment") }}</FormLabel>
              <FormControl>
                <MultiSelectTags :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']" :options="equipmentOptions"
                  :placeholder="$t('exercises.equipmentPlaceholder')" />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
      </div>
      <DrawerFooter class="flex-col gap-2 justify-between">
        <Dialog v-model:open="showDeleteDialog">
          <DialogTrigger as-child>
            <Button type="button" variant="destructive" :disabled="isPending">
              {{ $t("exercises.deleteExercise") }}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{{ $t("exercises.deleteExercise") }}</DialogTitle>
              <DialogDescription>
                {{
                  $t("exercises.deleteExerciseConfirmDescription", {
                    name: exercise?.name,
                  })
                }}
              </DialogDescription>
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
          {{ isUpdating ? $t("updating") : $t("exercises.updateExercise") }}
        </Button>
        <DrawerClose as-child>
          <Button variant="outline">{{ $t("cancel") }}</Button>
        </DrawerClose>
      </DrawerFooter>
    </div>
  </DrawerContent>
</template>
