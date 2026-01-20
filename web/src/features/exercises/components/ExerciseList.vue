<script setup lang="ts">
import type { Exercise } from "@/features/exercises/types";
import Card from "@/components/ui/card/Card.vue";
import CardHeader from "@/components/ui/card/CardHeader.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";

defineProps<{
  exercises: Exercise[];
  loading?: boolean;
}>();

const emits = defineEmits<{
  (e: "edit", exercise: Exercise): void;
}>();

const handleEdit = (exercise: Exercise) => {
  emits("edit", exercise);
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
      <Card v-for="exercise in exercises" :key="exercise.id">
        <CardHeader>
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <CardTitle>{{ exercise.name }}</CardTitle>
              <CardDescription v-if="exercise.description">
                {{ exercise.description }}
              </CardDescription>
            </div>
            <Button
              variant="outline"
              size="sm"
              @click="handleEdit(exercise)"
              class="ml-4"
            >
              {{ $t("edit") }}
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="flex flex-wrap gap-4 text-sm">
            <div v-if="exercise.primary_muscle_groups.length > 0">
              <span class="font-medium text-muted-foreground">{{
                $t("exercises.primaryLabel")
              }}</span>
              <span class="ml-2">
                {{
                  exercise.primary_muscle_groups.map((mg) => mg.name).join(", ")
                }}
              </span>
            </div>
            <div
              v-if="exercise.secondary_muscle_groups.length > 0"
              class="text-muted-foreground"
            >
              <span class="font-medium">{{
                $t("exercises.secondaryLabel")
              }}</span>
              <span class="ml-2">
                {{
                  exercise.secondary_muscle_groups
                    .map((mg) => mg.name)
                    .join(", ")
                }}
              </span>
            </div>
            <div v-if="exercise.equipment.length > 0">
              <span class="font-medium text-muted-foreground">{{
                $t("exercises.equipmentLabel")
              }}</span>
              <span class="ml-2">
                {{ exercise.equipment.map((eq) => eq.name).join(", ") }}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

