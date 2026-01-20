<script setup lang="ts">
import { computed, watch } from "vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useCreateExercise } from "../composables/useCreateExercise";
import { useMuscleGroup } from "@/features/reference/composables/useMuscleGroup";
import { useEquipment } from "@/features/reference/composables/useEquipment";
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
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { MultiSelectTags } from "@/components/ui/multi-select-tags";

const props = defineProps<{
  open?: boolean;
  modal?: boolean;
}>();

const emits = defineEmits<{
  (e: "exercise-created"): void;
}>();

const { createExercise, isPending, error } = useCreateExercise();
const { muscleGroup } = useMuscleGroup();
const { equipment } = useEquipment();

const formSchema = z.object({
  name: z.string().min(1, "Exercise name is required"),
  description: z.string().optional(),
  primary_muscle_groups: z
    .array(z.string())
    .min(1, "At least one primary muscle group is required"),
  secondary_muscle_groups: z.array(z.string()).optional(),
  equipment: z.array(z.string()).min(1, "At least one equipment is required"),
});

const { handleSubmit, resetForm } = useForm({
  validationSchema: toTypedSchema(formSchema),
  initialValues: {
    name: "",
    description: "",
    primary_muscle_groups: [] as string[],
    secondary_muscle_groups: [] as string[],
    equipment: [] as string[],
  },
});

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
    });
    resetForm();
    emits("exercise-created");
  } catch (err) {
    console.error("Failed to create exercise:", err);
  }
});

watch(
  () => props.open,
  (newValue) => {
    if (!newValue) {
      resetForm();
    }
  },
);
</script>

<template>
  <DrawerContent class="max-h-[95vh]">
    <div class="mx-auto w-full max-w-2xl overflow-y-auto">
      <DrawerHeader>
        <DrawerTitle>Create New Exercise</DrawerTitle>
        <DrawerDescription>
          Add a new exercise to your library. Fill in the details below.
        </DrawerDescription>
      </DrawerHeader>
      <div class="p-4 pb-0 space-y-6">
        <div
          v-if="error"
          class="p-4 bg-destructive/10 text-destructive rounded-lg"
        >
          <p>Error: {{ error.message }}</p>
        </div>

        <form @submit="onSubmit" class="space-y-4">
          <FormField v-slot="{ componentField }" name="name">
            <FormItem>
              <FormLabel>Exercise Name</FormLabel>
              <FormControl>
                <Input
                  placeholder="e.g., Bench Press"
                  v-bind="componentField"
                  required
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="description">
            <FormItem>
              <FormLabel>Description</FormLabel>
              <FormControl>
                <Textarea
                  placeholder="Describe the exercise..."
                  rows="3"
                  v-bind="componentField"
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="primary_muscle_groups">
            <FormItem>
              <FormLabel>Primary Muscle Groups</FormLabel>
              <FormControl>
                <MultiSelectTags
                  :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']"
                  :options="muscleGroupOptions"
                  placeholder="Select primary muscle groups..."
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="secondary_muscle_groups">
            <FormItem>
              <FormLabel>Secondary Muscle Groups</FormLabel>
              <FormControl>
                <MultiSelectTags
                  :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']"
                  :options="muscleGroupOptions"
                  placeholder="Select secondary muscle groups..."
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="equipment">
            <FormItem>
              <FormLabel>Equipment</FormLabel>
              <FormControl>
                <MultiSelectTags
                  :model-value="(componentField.modelValue ?? []) as string[]"
                  @update:model-value="componentField['onUpdate:modelValue']"
                  :options="equipmentOptions"
                  placeholder="Select equipment..."
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          </FormField>
        </form>
      </div>
      <DrawerFooter>
        <Button @click="onSubmit" :disabled="isPending">
          {{ isPending ? "Creating..." : "Create Exercise" }}
        </Button>
        <DrawerClose as-child>
          <Button variant="outline">Cancel</Button>
        </DrawerClose>
      </DrawerFooter>
    </div>
  </DrawerContent>
</template>
