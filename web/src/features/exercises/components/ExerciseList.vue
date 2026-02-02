<script setup lang="ts">
import { ref, watch, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import type { Exercise } from "@/features/exercises/types";
import Card from "@/components/ui/card/Card.vue";
import CardHeader from "@/components/ui/card/CardHeader.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Dumbbell } from "lucide-vue-next";
import { getImageUrl, revokeImageUrl } from "@/lib/api";

const props = defineProps<{
  exercises: Exercise[];
  loading?: boolean;
}>();

const emits = defineEmits<{
  (e: "edit", exercise: Exercise): void;
}>();

const { t } = useI18n();

const formatMuscleGroups = (groups: { name: string }[]) =>
  groups.map((g) => t(`muscleGroup.${g.name}`)).join(", ");
const formatEquipment = (items: { name: string }[]) =>
  items.map((e) => t(`equipment.${e.name}`)).join(", ");

const handleEdit = (exercise: Exercise) => {
  emits("edit", exercise);
};

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
  <div class="space-y-4">
    <div v-if="loading" class="space-y-4">
      <Card v-for="i in 5" :key="i">
        <CardHeader>
          <CardTitle>
            <div class="h-6 w-48 bg-gray-200 animate-pulse rounded"></div>
          </CardTitle>
          <CardDescription>
            <div
              class="h-4 w-full bg-gray-200 animate-pulse rounded mt-2"
            ></div>
          </CardDescription>
        </CardHeader>
      </Card>
    </div>

    <div v-else-if="exercises.length === 0" class="text-center py-12">
      <p class="text-muted-foreground">{{ $t("exercises.noExercises") }}</p>
    </div>

    <div v-else class="space-y-4">
      <Card v-for="exercise in exercises" :key="exercise.id" class="gap-2">
        <CardContent>
          <div class="flex items-start justify-between gap-4">
            <div class="flex gap-4 flex-1 items-center">
              <div
                class="shrink-0 w-20 h-20 rounded-lg border overflow-hidden bg-muted flex items-center justify-center"
              >
                <img
                  v-if="getImageUrlForExercise(exercise)"
                  :src="getImageUrlForExercise(exercise)"
                  :alt="exercise.name"
                  class="w-full h-full object-cover"
                />
                <Dumbbell v-else class="w-10 h-10 text-muted-foreground" />
              </div>
              <div class="flex-1 gap-2 flex flex-col">
                <CardTitle>{{ exercise.name }}</CardTitle>
                <CardDescription v-if="exercise.description">
                  {{ exercise.description }}
                </CardDescription>
                <div class="flex flex-col gap-1 text-sm">
                  <div v-if="exercise.primary_muscle_groups.length > 0">
                    <span class="font-medium text-muted-foreground">{{
                      $t("exercises.primaryLabel")
                    }}</span>
                    <span class="ml-2">
                      {{ formatMuscleGroups(exercise.primary_muscle_groups) }}
                    </span>
                  </div>
                  <div v-if="exercise.secondary_muscle_groups.length > 0">
                    <span class="font-medium text-muted-foreground">{{
                      $t("exercises.secondaryLabel")
                    }}</span>
                    <span class="ml-2">
                      {{ formatMuscleGroups(exercise.secondary_muscle_groups) }}
                    </span>
                  </div>
                  <div v-if="exercise.equipment.length > 0">
                    <span class="font-medium text-muted-foreground">{{
                      $t("exercises.equipmentLabel")
                    }}</span>
                    <span class="ml-2">
                      {{ formatEquipment(exercise.equipment) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              @click="handleEdit(exercise)"
              class="shrink-0"
            >
              {{ $t("edit") }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
