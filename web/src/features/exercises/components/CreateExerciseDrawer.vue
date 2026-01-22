<script setup lang="ts">
import { computed, watch, ref } from "vue";
import { useForm, useField } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useCreateExercise } from "../composables/useCreateExercise";
import { useMuscleGroup } from "@/features/reference/composables/useMuscleGroup";
import { useEquipment } from "@/features/reference/composables/useEquipment";
import { useExerciseFeature } from "@/features/reference/composables/useExerciseFeature";
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

const props = defineProps<{
  open?: boolean;
  modal?: boolean;
}>();

const emits = defineEmits<{
  (e: "exercise-created"): void;
  (e: "form-dirty", value: boolean): void;
}>();

import { useI18n } from "vue-i18n";

const { t } = useI18n();
const { createExercise, isPending, error } = useCreateExercise();
const { muscleGroup } = useMuscleGroup();
const { equipment } = useEquipment();
const { exerciseFeatures } = useExerciseFeature();

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
    image: null as File | null,
    equipment: [] as string[],
    exercise_features: [] as string[],
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

const exerciseFeatureOptions = computed(() =>
  exerciseFeatures.value.map((feature) => ({
    value: feature.name,
    label: feature.name,
  })),
);

const imageUrl = ref<string | null>(null);
const { value: imageValue } = useField<File | null | undefined>("image");

watch(imageValue, (file) => {
  if (file instanceof File) {
    const reader = new FileReader();
    reader.onload = (e) => {
      imageUrl.value = (e.target?.result as string) ?? null;
    };
    reader.readAsDataURL(file);
  } else {
    imageUrl.value = null;
  }
});

const clearImage = () => {
  imageUrl.value = null;
  setFieldValue("image", null);
};

const onSubmit = handleSubmit(async (values) => {
  try {
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
  } catch (err) {
    console.error("Failed to create exercise:", err);
  }
});

watch(
  () => props.open,
  (newValue) => {
    if (newValue) {
      resetForm();
      imageUrl.value = null;
    }
  },
);
</script>

<template>
  <DrawerContent class="max-h-[95vh]">
    <div class="mx-auto w-full max-w-2xl overflow-y-auto">
      <DrawerHeader>
        <DrawerTitle>{{ $t("exercises.createNew") }}</DrawerTitle>
        <DrawerDescription>
          {{ $t("exercises.createDescription") }}
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
                    type="file"
                    accept="image/*"
                    class="cursor-pointer"
                    @change="
                      (e) => {
                        const target = e.target as HTMLInputElement;
                        const file = target.files?.[0];
                        handleChange(file ?? null);
                      }
                    "
                  />
                </FormControl>
                <FormMessage />
                <div class="flex items-center gap-3 mt-2">
                  <div v-if="imageUrl">
                    <img
                      :src="imageUrl"
                      alt="Preview"
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
      <DrawerFooter>
        <Button @click="onSubmit" :disabled="isPending">
          {{ isPending ? $t("creating") : $t("exercises.create") }}
        </Button>
        <DrawerClose as-child>
          <Button variant="outline">{{ $t("cancel") }}</Button>
        </DrawerClose>
      </DrawerFooter>
    </div>
  </DrawerContent>
</template>
