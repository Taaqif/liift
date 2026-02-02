<script setup lang="ts">
import { computed, watch, ref, onUnmounted } from "vue";
import { useForm, useField } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useCreateExercise } from "../composables/useCreateExercise";
import { useUpdateExercise } from "../composables/useUpdateExercise";
import { useDeleteExercise } from "../composables/useDeleteExercise";
import { useMuscleGroupOptions } from "@/features/reference/composables/useMuscleGroupOptions";
import { useEquipmentOptions } from "@/features/reference/composables/useEquipmentOptions";
import { useExerciseFeatureOptions } from "@/features/reference/composables/useExerciseFeatureOptions";
import type { Exercise } from "../types";
import { getImageUrl, revokeImageUrl } from "@/lib/api";
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
  FieldGroup,
  FieldSet,
  FieldLegend,
  FieldDescription,
  FieldLabel,
  Field,
  FieldTitle,
} from "@/components/ui/field";
import { Checkbox } from "@/components/ui/checkbox";
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
  exercise?: Exercise | null;
}>();

const emits = defineEmits<{
  (e: "exercise-created"): void;
  (e: "exercise-updated"): void;
  (e: "exercise-deleted"): void;
  (e: "form-dirty", value: boolean): void;
}>();

const { t } = useI18n();
const isEditMode = computed(() => !!props.exercise);

const {
  createExercise,
  isPending: isCreating,
  error: createError,
} = useCreateExercise();
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

const { options: muscleGroupOptions } = useMuscleGroupOptions();
const { options: equipmentOptions } = useEquipmentOptions();
const { options: exerciseFeatureOptions } = useExerciseFeatureOptions();

const formSchema = z.object({
  name: z.string().min(1, t("exercises.validation.nameRequired")),
  description: z.string().optional(),
  primary_muscle_groups: z
    .array(z.string())
    .min(1, t("exercises.validation.primaryMuscleGroupsRequired")),
  secondary_muscle_groups: z.array(z.string()).optional(),
  image: z.union([z.instanceof(File), z.null()]).optional(),
  equipment: z
    .array(z.string())
    .min(1, t("exercises.validation.equipmentRequired")),
  exercise_features: z
    .array(z.string())
    .min(1, t("exercises.validation.exerciseFeaturesRequired")),
});

const { handleSubmit, resetForm, meta, setFieldValue } = useForm({
  validationSchema: toTypedSchema(formSchema),
  initialValues: {
    name: "",
    description: "",
    primary_muscle_groups: [] as string[],
    secondary_muscle_groups: [] as string[],
    equipment: [] as string[],
    exercise_features: [] as string[],
  },
});

const isFormDirty = computed(() => meta.value.dirty);

const imageUrl = ref<string | null>(null);

// Load existing image when exercise changes (edit mode only)
const loadExistingImage = async (imagePath: string | null | undefined) => {
  if (!imagePath) {
    imageUrl.value = null;
    return;
  }

  const fullPath = imagePath.startsWith("http")
    ? imagePath
    : `${window.location.origin}${imagePath}`;

  const blobUrl = await getImageUrl(fullPath);
  imageUrl.value = blobUrl ?? null;
};

watch(
  () => props.exercise?.image,
  async (imagePath) => {
    if (isEditMode.value) {
      await loadExistingImage(imagePath ?? null);
    }
  },
  { immediate: true },
);

const { value: imageValue } = useField<File | null | undefined>("image");

watch(imageValue, (file) => {
  if (file instanceof File) {
    const reader = new FileReader();
    reader.onload = (e) => {
      imageUrl.value = (e.target?.result as string) ?? null;
    };
    reader.readAsDataURL(file);
  } else if (!isEditMode.value) {
    // In create mode, clear image if no file selected
    imageUrl.value = null;
  }
  // In edit mode, if file is cleared (null), we keep the existing image
  // The existing image is managed by the watch on props.exercise?.image
});

const clearImage = () => {
  imageUrl.value = null;
  setFieldValue("image", null);
};

watch(
  [isFormDirty, () => props.open],
  ([dirty, isOpen]) => {
    if (isOpen) {
      emits("form-dirty", dirty);
    }
  },
  { immediate: true },
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
        exercise_features:
          exercise.exercise_features?.map((ef) => ef.name) ?? [],
        image: exercise.image ? undefined : null,
      },
    });
  }
};

watch(
  () => props.exercise,
  (exercise) => {
    if (exercise && props.open && isEditMode.value) {
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
      imageUrl.value = null;
      showDeleteDialog.value = false;
    } else if (newValue && !props.exercise) {
      // Create mode - reset form when opening
      resetForm();
      imageUrl.value = null;
    }
  },
);

const onSubmit = handleSubmit(async (values) => {
  try {
    if (isEditMode.value && props.exercise) {
      // Edit mode
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
          exercise_features: values.exercise_features,
          image: values.image,
        },
      });
      resetForm();
      emits("exercise-updated");
    } else {
      // Create mode
      await createExercise({
        name: values.name.trim(),
        description: values.description?.trim() || undefined,
        primary_muscle_groups: values.primary_muscle_groups,
        secondary_muscle_groups:
          values.secondary_muscle_groups &&
          values.secondary_muscle_groups.length > 0
            ? values.secondary_muscle_groups
            : undefined,
        equipment: values.equipment,
        exercise_features: values.exercise_features,
        image: values.image ?? null,
      });
      resetForm();
      imageUrl.value = null;
      emits("exercise-created");
    }
  } catch (err) {
    console.error(
      `Failed to ${isEditMode.value ? "update" : "create"} exercise:`,
      err,
    );
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

const error = computed(
  () => createError.value || updateError.value || deleteError.value,
);
const isPending = computed(
  () => isCreating.value || isUpdating.value || isDeleting.value,
);
const showDeleteDialog = ref(false);

const title = computed(() =>
  isEditMode.value ? t("exercises.editTitle") : t("exercises.createNew"),
);

const description = computed(() =>
  isEditMode.value
    ? t("exercises.editDescription")
    : t("exercises.createDescription"),
);

const submitButtonText = computed(() => {
  if (isEditMode.value) {
    return isUpdating.value ? t("updating") : t("exercises.updateExercise");
  }
  return isCreating.value ? t("creating") : t("exercises.create");
});

onUnmounted(() => {
  if (props.exercise?.image) {
    const fullPath = props.exercise.image.startsWith("http")
      ? props.exercise.image
      : `${window.location.origin}${props.exercise.image}`;
    revokeImageUrl(fullPath);
  }
});
</script>

<template>
  <DrawerContent class="max-h-[95vh]">
    <div class="mx-auto w-full max-w-2xl overflow-y-auto">
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
          <p>{{ $t("exercises.error") }}: {{ error.message }}</p>
        </div>

        <form @submit="onSubmit" class="space-y-4">
          <FormField v-slot="{ componentField }" name="name">
            <FormItem>
              <FormLabel>{{ $t("exercises.name") }}</FormLabel>
              <FormControl>
                <Input
                  :placeholder="$t('exercises.namePlaceholder')"
                  v-bind="componentField"
                  required
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="description">
            <FormItem>
              <FormLabel>{{ $t("exercises.description") }}</FormLabel>
              <FormControl>
                <Textarea
                  :placeholder="$t('exercises.descriptionPlaceholder')"
                  rows="3"
                  v-bind="componentField"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <div class="space-y-2">
            <FormField v-slot="{ handleChange }" name="image">
              <FormItem>
                <FormLabel>{{ $t("exercises.image") }}</FormLabel>
                <FormControl>
                  <Input
                    @change="
                      (e) => {
                        const target = e.target as HTMLInputElement;
                        const file = target.files?.[0];
                        handleChange(file ?? (isEditMode ? undefined : null));
                      }
                    "
                    type="file"
                    accept="image/*"
                    class="cursor-pointer"
                  />
                </FormControl>
                <FormMessage />
                <div class="flex items-center gap-3 mt-2">
                  <div v-if="imageUrl">
                    <img
                      :src="imageUrl"
                      :alt="exercise?.name || 'Preview'"
                      class="h-32 w-32 rounded-lg object-cover border"
                    />
                  </div>
                  <Button
                    v-if="imageUrl"
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="clearImage"
                    :disabled="!imageUrl"
                  >
                    {{ $t("exercises.clearImage") }}
                  </Button>
                </div>
              </FormItem>
            </FormField>
          </div>

          <FormField v-slot="{ componentField }" name="primary_muscle_groups">
            <FormItem>
              <FormLabel>{{ $t("exercises.primaryMuscleGroups") }}</FormLabel>
              <FormControl>
                <MultiSelectTags
                  :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']"
                  :options="muscleGroupOptions"
                  :placeholder="$t('exercises.primaryMuscleGroupsPlaceholder')"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="secondary_muscle_groups">
            <FormItem>
              <FormLabel>{{ $t("exercises.secondaryMuscleGroups") }}</FormLabel>
              <FormControl>
                <MultiSelectTags
                  :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']"
                  :options="muscleGroupOptions"
                  :placeholder="
                    $t('exercises.secondaryMuscleGroupsPlaceholder')
                  "
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="equipment">
            <FormItem>
              <FormLabel>{{ $t("exercises.equipment") }}</FormLabel>
              <FormControl>
                <MultiSelectTags
                  :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']"
                  :options="equipmentOptions"
                  :placeholder="$t('exercises.equipmentPlaceholder')"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="exercise_features">
            <FormItem>
              <FormControl>
                <FieldGroup>
                  <FieldSet class="gap-4">
                    <FieldLegend>{{
                      $t("exercises.exerciseFeatures")
                    }}</FieldLegend>
                    <FieldDescription class="line-clamp-1">
                      {{ $t("exercises.exerciseFeaturesDescription") }}
                    </FieldDescription>
                    <FieldGroup
                      class="flex flex-row flex-wrap gap-2 [--radius:9999rem]"
                    >
                      <FieldLabel
                        v-for="option in exerciseFeatureOptions"
                        :key="option.value"
                        class="!w-fit"
                      >
                        <Field
                          orientation="horizontal"
                          class="gap-1.5 overflow-hidden !px-3 !py-1.5 transition-all duration-100 ease-linear group-has-data-[state=checked]/field-label:!px-2"
                        >
                          <Checkbox
                            :id="option.value"
                            :model-value="
                              (componentField.modelValue ?? []).includes(
                                option.value,
                              )
                            "
                            @update:model-value="
                              (value: boolean | 'indeterminate') => {
                                const checked = value === true;
                                const current = (componentField.modelValue ??
                                  []) as string[];
                                if (checked) {
                                  componentField['onUpdate:modelValue']?.([
                                    ...current,
                                    option.value,
                                  ]);
                                } else {
                                  componentField['onUpdate:modelValue']?.(
                                    current.filter((v) => v !== option.value),
                                  );
                                }
                              }
                            "
                            class="-ml-6 -translate-x-1 rounded-full transition-all duration-100 ease-linear data-[state=checked]:ml-0 data-[state=checked]:translate-x-0"
                          />
                          <FieldTitle>{{ option.label }}</FieldTitle>
                        </Field>
                      </FieldLabel>
                    </FieldGroup>
                  </FieldSet>
                </FieldGroup>
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
      </div>
      <DrawerFooter class="flex-col gap-2 justify-between">
        <Dialog v-if="isEditMode" v-model:open="showDeleteDialog">
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
