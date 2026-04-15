<script setup lang="ts">
import { computed, watch, ref, onUnmounted } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { useForm, useField } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useI18n } from "vue-i18n";
import { useExercise } from "@/features/exercises/composables/useExercise";
import { useCreateExercise } from "@/features/exercises/composables/useCreateExercise";
import { useUpdateExercise } from "@/features/exercises/composables/useUpdateExercise";
import { useDeleteExercise } from "@/features/exercises/composables/useDeleteExercise";
import { useMuscleGroupOptions } from "@/features/reference/composables/useMuscleGroupOptions";
import { useEquipmentOptions } from "@/features/reference/composables/useEquipmentOptions";
import { useExerciseFeatureOptions } from "@/features/reference/composables/useExerciseFeatureOptions";
import type { Exercise } from "@/features/exercises/types";
import { getImageUrl, revokeImageUrl } from "@/lib/api";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ArrowLeft, Plus, Trash2, GripVertical } from "lucide-vue-next";
import { VueDraggable } from "vue-draggable-plus";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const exerciseId = computed(() => {
  const id = route.params.id;
  return id ? Number(id) : null;
});
const isEditMode = computed(() => !!exerciseId.value);

const { exercise, loading: exerciseLoading, fetching: exerciseFetching } = useExercise(exerciseId);

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

const forceOptions = ["pull", "push", "static"] as const;
const categoryOptions = ["strength", "cardio", "stretching"] as const;

const formSchema = z.object({
  name: z.string().min(1, t("exercises.validation.nameRequired")),
  description: z.string().optional(),
  force: z.string().optional(),
  category: z.string().optional(),
  instructions: z.array(z.string()).optional(),
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
    force: undefined as string | undefined,
    category: undefined as string | undefined,
    instructions: [] as string[],
    primary_muscle_groups: [] as string[],
    secondary_muscle_groups: [] as string[],
    equipment: [] as string[],
    exercise_features: [] as string[],
  },
});

const imageUrl = ref<string | null>(null);

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
  () => exercise.value?.image,
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
    imageUrl.value = null;
  }
});

const clearImage = () => {
  imageUrl.value = null;
  setFieldValue("image", null);
};

const populateForm = (ex: Exercise) => {
  resetForm({
    values: {
      name: ex.name,
      description: ex.description || "",
      force: ex.force ?? undefined,
      category: ex.category ?? undefined,
      instructions: ex.instructions ?? [],
      primary_muscle_groups: ex.primary_muscle_groups.map((mg) => mg.name),
      secondary_muscle_groups: ex.secondary_muscle_groups.map((mg) => mg.name),
      equipment: ex.equipment.map((eq) => eq.name),
      exercise_features: ex.exercise_features?.map((ef) => ef.name) ?? [],
      image: ex.image ? undefined : null,
    },
  });
};

watch(exercise, (ex) => {
  if (ex) populateForm(ex);
}, { immediate: true });

const onSubmit = handleSubmit(async (values) => {
  try {
    if (isEditMode.value && exerciseId.value) {
      await updateExercise({
        id: exerciseId.value,
        data: {
          name: values.name.trim(),
          description: values.description?.trim() || undefined,
          force: values.force || undefined,
          category: values.category || undefined,
          instructions: values.instructions?.filter((i) => i.trim()) ?? [],
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
    } else {
      await createExercise({
        name: values.name.trim(),
        description: values.description?.trim() || undefined,
        force: values.force || undefined,
        category: values.category || undefined,
        instructions: values.instructions?.filter((i) => i.trim()) ?? [],
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
    }
    resetForm();
    router.push({ name: "exercises" });
  } catch (err) {
    console.error(
      `Failed to ${isEditMode.value ? "update" : "create"} exercise:`,
      err,
    );
  }
});

const onDelete = async () => {
  if (!exerciseId.value) return;
  try {
    await deleteExercise(exerciseId.value);
    showDeleteDialog.value = false;
    router.push({ name: "exercises" });
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

onBeforeRouteLeave(() => {
  if (meta.value.dirty && !isPending.value) {
    return window.confirm(t("unsavedChanges.confirmLeave") || "You have unsaved changes. Leave anyway?");
  }
});

onUnmounted(() => {
  if (exercise.value?.image) {
    const fullPath = exercise.value.image.startsWith("http")
      ? exercise.value.image
      : `${window.location.origin}${exercise.value.image}`;
    revokeImageUrl(fullPath);
  }
});
</script>

<template>
  <div>
    <div class="mb-8">
      <button class="text-sm text-muted-foreground hover:text-foreground transition-colors mb-1" @click="router.push({ name: 'exercises' })">
        ← {{ $t("nav.exercises") }}
      </button>
      <h1 class="text-3xl font-bold tracking-tight">{{ title }}</h1>
      <p class="text-muted-foreground mt-1">{{ description }}</p>
    </div>

    <div v-if="isEditMode && (exerciseLoading || exerciseFetching)" class="flex items-center justify-center py-24">
      <div class="text-muted-foreground">{{ $t("loading") }}</div>
    </div>

    <div v-else class="space-y-6">
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

        <div class="grid grid-cols-2 gap-4">
          <FormField v-slot="{ field, handleChange }" name="force">
            <FormItem>
              <FormLabel>{{ $t("exercises.force") }}</FormLabel>
              <Select
                :model-value="field.value as string | undefined"
                @update:model-value="handleChange"
              >
                <FormControl>
                  <SelectTrigger class="w-full">
                    <SelectValue :placeholder="$t('exercises.forcePlaceholder')" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem v-for="opt in forceOptions" :key="opt" :value="opt">
                    {{ $t(`exercises.forceValues.${opt}`) }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ field, handleChange }" name="category">
            <FormItem>
              <FormLabel>{{ $t("exercises.category") }}</FormLabel>
              <Select
                :model-value="field.value as string | undefined"
                @update:model-value="handleChange"
              >
                <FormControl>
                  <SelectTrigger class="w-full">
                    <SelectValue :placeholder="$t('exercises.categoryPlaceholder')" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem v-for="opt in categoryOptions" :key="opt" :value="opt">
                    {{ $t(`exercises.categoryValues.${opt}`) }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          </FormField>
        </div>

        <FormField v-slot="{ field }" name="instructions">
          <FormItem>
            <FormLabel>{{ $t("exercises.instructions") }}</FormLabel>
            <div class="space-y-2">
              <VueDraggable
                :model-value="(field.value as string[] ?? [])"
                @update:model-value="field.onChange"
                handle=".drag-handle"
                animation="150"
                class="space-y-2"
              >
                <div
                  v-for="(_, i) in (field.value as string[] ?? [])"
                  :key="i"
                  class="flex gap-2 items-center"
                >
                  <GripVertical class="drag-handle h-4 w-4 shrink-0 cursor-grab text-muted-foreground" />
                  <Textarea
                    :model-value="(field.value as string[])[i]"
                    @input="(e: Event) => {
                      const list = [...((field.value as string[]) ?? [])];
                      list[i] = (e.target as HTMLTextAreaElement).value;
                      field.onChange(list);
                    }"
                    :placeholder="$t('exercises.instructionPlaceholder', { number: i + 1 })"
                    rows="2"
                    class="resize-none"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="() => {
                      const list = [...((field.value as string[]) ?? [])];
                      list.splice(i, 1);
                      field.onChange(list);
                    }"
                  >
                    <Trash2 class="h-4 w-4" />
                  </Button>
                </div>
              </VueDraggable>
              <Button
                type="button"
                variant="outline"
                size="sm"
                @click="() => {
                  const list = [...((field.value as string[]) ?? [])];
                  list.push('');
                  field.onChange(list);
                }"
              >
                <Plus class="h-4 w-4 mr-2" />
                {{ $t("exercises.addInstruction") }}
              </Button>
            </div>
            <FormMessage />
          </FormItem>
        </FormField>

        <div class="space-y-2">
          <FormField v-slot="{ handleChange }" name="image">
            <FormItem>
              <FormLabel>{{ $t("exercises.image") }}</FormLabel>
              <FormControl>
                <Input
                  @change="(e) => {
                    const target = e.target as HTMLInputElement;
                    const file = target.files?.[0];
                    handleChange(file ?? (isEditMode ? undefined : null));
                  }"
                  type="file"
                  accept="image/*"
                  class="cursor-pointer"
                />
              </FormControl>
              <FormMessage />
              <div class="flex items-center gap-3 mt-2">
                <div v-if="imageUrl">
                  <img :src="imageUrl" :alt="exercise?.name || 'Preview'"
                    class="h-32 w-32 rounded-lg object-cover border" />
                </div>
                <Button v-if="imageUrl" type="button" variant="outline" size="sm" @click="clearImage" :disabled="!imageUrl">
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
                :placeholder="$t('exercises.secondaryMuscleGroupsPlaceholder')"
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
                  <FieldLegend>{{ $t("exercises.exerciseFeatures") }}</FieldLegend>
                  <FieldDescription class="line-clamp-1">
                    {{ $t("exercises.exerciseFeaturesDescription") }}
                  </FieldDescription>
                  <FieldGroup class="flex flex-row flex-wrap gap-2 [--radius:9999rem]">
                    <FieldLabel v-for="option in exerciseFeatureOptions" :key="option.value" class="!w-fit">
                      <Field
                        orientation="horizontal"
                        class="gap-1.5 overflow-hidden !px-3 !py-1.5 transition-all duration-100 ease-linear group-has-data-[state=checked]/field-label:!px-2"
                      >
                        <Checkbox
                          :id="option.value"
                          :model-value="(componentField.modelValue ?? []).includes(option.value)"
                          @update:model-value="(value: boolean | 'indeterminate') => {
                            const checked = value === true;
                            const current = (componentField.modelValue ?? []) as string[];
                            if (checked) {
                              componentField['onUpdate:modelValue']?.([...current, option.value]);
                            } else {
                              componentField['onUpdate:modelValue']?.(current.filter((v) => v !== option.value));
                            }
                          }"
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

      <div class="flex flex-col gap-2 pt-4 border-t">
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
                {{ $t("exercises.deleteExerciseConfirmDescription", { name: exercise?.name }) }}
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
          {{ submitButtonText }}
        </Button>
        <Button variant="outline" @click="router.push({ name: 'exercises' })">
          {{ $t("cancel") }}
        </Button>
      </div>
    </div>
  </div>
</template>
