<script setup lang="ts">
import { ref, watch, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import type { Exercise } from "@/features/exercises/types";
import Card from "@/components/ui/card/Card.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Dumbbell, History, Check } from "lucide-vue-next";
import { getImageUrl, revokeImageUrl } from "@/lib/api";
import ExerciseLogDrawer from "@/features/exercises/components/ExerciseLogDrawer.vue";

const props = defineProps<{
  exercises: Exercise[];
  loading?: boolean;
  selectable?: boolean;
  selectedIds?: Set<number>;
}>();

const emits = defineEmits<{
  (e: "edit", exercise: Exercise): void;
  (e: "select", exercise: Exercise): void;
}>();

const { t } = useI18n();

const formatMuscleGroups = (groups: { name: string }[]) =>
  groups.map((g) => t(`muscleGroup.${g.name}`)).join(", ");
const formatEquipment = (items: { name: string }[]) =>
  items.map((e) => t(`equipment.${e.name}`)).join(", ");

const handleEdit = (exercise: Exercise) => {
  emits("edit", exercise);
};

const logDrawerOpen = ref(false);
const selectedExercise = ref<Exercise | null>(null);

function openLogs(exercise: Exercise) {
  selectedExercise.value = exercise;
  logDrawerOpen.value = true;
}

type ImageCacheEntry = { url: string; path: string };
const imageCache = ref<Map<number, ImageCacheEntry>>(new Map());

const setCache = (id: number, entry: ImageCacheEntry) => {
  const next = new Map(imageCache.value);
  next.set(id, entry);
  imageCache.value = next;
};

const deleteCache = (id: number) => {
  const next = new Map(imageCache.value);
  next.delete(id);
  imageCache.value = next;
};

const getFullImagePath = (image: string) =>
  image.startsWith("http") ? image : `${window.location.origin}${image}`;

const loadImage = async (exercise: Exercise) => {
  if (!exercise.image) {
    const cached = imageCache.value.get(exercise.id);
    if (cached) {
      revokeImageUrl(cached.path);
      deleteCache(exercise.id);
    }
    return;
  }

  const imagePath = getFullImagePath(exercise.image);
  const cached = imageCache.value.get(exercise.id);

  // If unchanged, keep current blob URL
  if (cached && cached.path === imagePath) return;

  // If changed, revoke old and reload
  if (cached && cached.path !== imagePath) {
    revokeImageUrl(cached.path);
    deleteCache(exercise.id);
  }

  const blobUrl = await getImageUrl(imagePath);
  if (blobUrl) {
    setCache(exercise.id, { url: blobUrl, path: imagePath });
  }
};

const loadAllImages = async () => {
  for (const exercise of props.exercises) {
    await loadImage(exercise);
  }
};

watch(() => props.exercises, loadAllImages, { immediate: true });

onUnmounted(() => {
  imageCache.value.forEach((entry) => {
    revokeImageUrl(entry.path);
  });
  imageCache.value = new Map();
});

const getImageUrlForExercise = (exercise: Exercise): string | undefined => {
  return imageCache.value.get(exercise.id)?.url;
};
</script>

<template>
  <div :class="selectable ? '' : 'space-y-4'">
    <!-- Loading skeletons -->
    <div v-if="loading" :class="selectable ? 'divide-y' : 'space-y-4'">
      <template v-if="selectable">
        <div v-for="i in 8" :key="i" class="flex items-center gap-3 px-4 py-3">
          <div class="shrink-0 w-10 h-10 rounded-md bg-muted animate-pulse" />
          <div class="flex-1 space-y-1.5">
            <div class="h-4 w-36 bg-muted animate-pulse rounded" />
            <div class="h-3 w-24 bg-muted animate-pulse rounded" />
          </div>
        </div>
      </template>
      <template v-else>
        <Card v-for="i in 5" :key="i">
          <CardContent>
            <div class="flex gap-4 items-center">
              <div class="shrink-0 w-20 h-20 rounded-lg bg-muted animate-pulse" />
              <div class="flex-1 space-y-2">
                <div class="h-5 w-48 bg-muted animate-pulse rounded" />
                <div class="h-4 w-full bg-muted animate-pulse rounded" />
              </div>
            </div>
          </CardContent>
        </Card>
      </template>
    </div>

    <div v-else-if="exercises.length === 0" class="text-center py-12">
      <p class="text-muted-foreground">{{ $t("exercises.noExercises") }}</p>
    </div>

    <!-- Selectable compact list -->
    <div v-else-if="selectable" class="divide-y">
      <div
        v-for="exercise in exercises"
        :key="exercise.id"
        class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-muted/50 transition-colors"
        @click="emits('select', exercise)"
      >
        <div class="shrink-0 w-10 h-10 rounded-md border overflow-hidden bg-muted flex items-center justify-center">
          <img
            v-if="getImageUrlForExercise(exercise)"
            :src="getImageUrlForExercise(exercise)"
            :alt="exercise.name"
            class="w-full h-full object-cover"
          />
          <Dumbbell v-else class="w-5 h-5 text-muted-foreground" />
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium truncate">{{ exercise.name }}</p>
          <p v-if="exercise.primary_muscle_groups.length > 0" class="text-xs text-muted-foreground truncate">
            {{ formatMuscleGroups(exercise.primary_muscle_groups) }}
          </p>
        </div>
        <div
          class="shrink-0 w-6 h-6 rounded-full border-2 flex items-center justify-center transition-colors"
          :class="selectedIds?.has(exercise.id)
            ? 'bg-primary border-primary text-primary-foreground'
            : 'border-muted-foreground/40'"
        >
          <Check v-if="selectedIds?.has(exercise.id)" class="w-3.5 h-3.5" />
        </div>
      </div>
    </div>

    <!-- Default card list -->
    <div v-else class="space-y-3">
      <Card v-for="exercise in exercises" :key="exercise.id" class="gap-2">
        <CardContent>
          <div class="flex gap-3 items-center">
            <div
              class="shrink-0 w-14 h-14 rounded-lg border overflow-hidden bg-muted flex items-center justify-center"
            >
              <img
                v-if="getImageUrlForExercise(exercise)"
                :src="getImageUrlForExercise(exercise)"
                :alt="exercise.name"
                class="w-full h-full object-cover"
              />
              <Dumbbell v-else class="w-7 h-7 text-muted-foreground" />
            </div>
            <div class="flex-1 min-w-0 flex flex-col gap-1">
              <CardTitle class="truncate">{{ exercise.name }}</CardTitle>
              <div class="flex flex-col gap-0.5 text-xs text-muted-foreground">
                <span v-if="exercise.primary_muscle_groups.length > 0" class="truncate">
                  {{ formatMuscleGroups(exercise.primary_muscle_groups) }}
                </span>
                <span v-if="exercise.equipment.length > 0" class="truncate">
                  {{ formatEquipment(exercise.equipment) }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-2 mt-3 pt-3 border-t justify-end">
            <Button
              variant="ghost"
              size="icon"
              class="size-8"
              @click="openLogs(exercise)"
            >
              <History class="size-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              @click="handleEdit(exercise)"
            >
              {{ $t("edit") }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>

<ExerciseLogDrawer
  v-if="!selectable"
  v-model:open="logDrawerOpen"
  :exercise-id="selectedExercise?.id ?? null"
  :exercise-name="selectedExercise?.name"
/>
</template>
