<script setup lang="ts">
import { computed, ref } from "vue";
import { Info } from "lucide-vue-next";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  exercise?: {
    name?: string;
    description?: string;
    instructions?: string[];
    image?: string;
    force?: string;
    category?: string;
    primary_muscle_groups?: { name: string }[];
    secondary_muscle_groups?: { name: string }[];
    equipment?: { name: string }[];
    exercise_features?: { name: string }[];
  } | null;
}>();

const { t } = useI18n();
const open = ref(false);

const hasContent = computed(() => {
  const e = props.exercise;
  if (!e) return false;
  return (
    !!e.description ||
    !!e.image ||
    !!e.force ||
    !!e.category ||
    (e.instructions?.length ?? 0) > 0 ||
    (e.primary_muscle_groups?.length ?? 0) > 0 ||
    (e.secondary_muscle_groups?.length ?? 0) > 0 ||
    (e.equipment?.length ?? 0) > 0
  );
});

const imageUrl = computed(() => {
  const img = props.exercise?.image;
  if (!img) return null;
  return img.startsWith("http") ? img : `${window.location.origin}${img}`;
});
</script>

<template>
  <template v-if="hasContent">
    <Button
      type="button"
      variant="ghost"
      size="icon"
      class="size-7 shrink-0 text-muted-foreground hover:text-foreground"
      @click.stop="open = true"
    >
      <Info class="size-4" />
    </Button>

    <Dialog v-model:open="open">
      <DialogContent class="max-w-sm max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{{ exercise?.name ?? t("exercises.info") }}</DialogTitle>
        </DialogHeader>

        <div class="space-y-4 text-sm">
          <!-- Image -->
          <img
            v-if="imageUrl"
            :src="imageUrl"
            :alt="exercise?.name"
            class="w-full rounded-md object-cover max-h-48"
          />

          <!-- Category & Force -->
          <div v-if="exercise?.category || exercise?.force" class="flex flex-wrap gap-1.5">
            <span
              v-if="exercise?.category"
              class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-secondary text-secondary-foreground capitalize"
            >
              {{ t(`exercises.categoryValues.${exercise.category}`, exercise.category) }}
            </span>
            <span
              v-if="exercise?.force"
              class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border border-border text-muted-foreground capitalize"
            >
              {{ t(`exercises.forceValues.${exercise.force}`, exercise.force) }}
            </span>
          </div>

          <!-- Muscle groups -->
          <div
            v-if="exercise?.primary_muscle_groups?.length || exercise?.secondary_muscle_groups?.length"
            class="space-y-2"
          >
            <div v-if="exercise?.primary_muscle_groups?.length">
              <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1">
                {{ t("exercises.primaryMuscleGroups") }}
              </p>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="mg in exercise.primary_muscle_groups"
                  :key="mg.name"
                  class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium bg-secondary text-secondary-foreground capitalize"
                >
                  {{ mg.name }}
                </span>
              </div>
            </div>
            <div v-if="exercise?.secondary_muscle_groups?.length">
              <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1">
                {{ t("exercises.secondaryMuscleGroups") }}
              </p>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="mg in exercise.secondary_muscle_groups"
                  :key="mg.name"
                  class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border border-border text-muted-foreground capitalize"
                >
                  {{ mg.name }}
                </span>
              </div>
            </div>
          </div>

          <!-- Equipment -->
          <div v-if="exercise?.equipment?.length">
            <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-1">
              {{ t("exercises.equipment") }}
            </p>
            <div class="flex flex-wrap gap-1">
              <span
                v-for="eq in exercise.equipment"
                :key="eq.name"
                class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border border-border text-muted-foreground capitalize"
              >
                {{ eq.name }}
              </span>
            </div>
          </div>

          <!-- Description -->
          <p v-if="exercise?.description" class="text-muted-foreground leading-relaxed">
            {{ exercise.description }}
          </p>

          <!-- Instructions -->
          <div v-if="exercise?.instructions?.length" class="space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              {{ t("exercises.instructions") }}
            </p>
            <ol class="space-y-1.5">
              <li v-for="(step, i) in exercise.instructions" :key="i" class="flex gap-2">
                <span class="text-muted-foreground shrink-0 tabular-nums font-medium">{{ i + 1 }}.</span>
                <span class="leading-snug">{{ step }}</span>
              </li>
            </ol>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </template>
</template>
