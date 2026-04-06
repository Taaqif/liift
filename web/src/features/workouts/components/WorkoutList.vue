<script setup lang="ts">
import type { Workout } from "@/features/workouts/types";
import Card from "@/components/ui/card/Card.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Dumbbell, Play } from "lucide-vue-next";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  workouts: Workout[];
  loading?: boolean;
  startingWorkoutId?: number | null;
}>();

const emits = defineEmits<{
  (e: "edit", workout: Workout): void;
  (e: "start", workout: Workout): void;
}>();

const { t } = useI18n();

const getExerciseCount = (workout: Workout): number =>
  workout.exercises?.length ?? 0;

const getTotalSets = (workout: Workout): number =>
  workout.exercises?.reduce((total, ex) => total + (ex.sets?.length ?? 0), 0) ?? 0;
</script>

<template>
  <div class="space-y-4">
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card v-for="i in 4" :key="i" class="gap-2">
        <CardContent>
          <div class="flex gap-4 items-center">
            <div class="shrink-0 w-16 h-16 rounded-lg bg-muted animate-pulse" />
            <div class="flex-1 space-y-2">
              <div class="h-5 w-36 bg-muted animate-pulse rounded" />
              <div class="h-4 w-24 bg-muted animate-pulse rounded" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <div v-else-if="workouts.length === 0" class="text-center py-12">
      <p class="text-muted-foreground">{{ $t("workouts.noWorkouts") }}</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card v-for="workout in workouts" :key="workout.id" class="gap-2">
        <CardContent>
          <div class="flex items-start justify-between gap-4">
            <div class="flex gap-4 flex-1 items-center">
              <div class="shrink-0 w-16 h-16 rounded-lg border overflow-hidden bg-muted flex items-center justify-center">
                <Dumbbell class="w-8 h-8 text-muted-foreground" />
              </div>
              <div class="flex-1 flex flex-col gap-1">
                <CardTitle class="line-clamp-1">{{ workout.name }}</CardTitle>
                <CardDescription v-if="workout.description" class="line-clamp-2">
                  {{ workout.description }}
                </CardDescription>
                <p class="text-sm text-muted-foreground">
                  {{ getExerciseCount(workout) }} {{ $t("workouts.exercises") }}
                  · {{ getTotalSets(workout) }} {{ $t("workouts.sets") }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <Button
                variant="default"
                size="sm"
                :disabled="startingWorkoutId != null"
                @click="emits('start', workout)"
              >
                <Play class="w-4 h-4 mr-1" />
                {{ $t("workoutSession.startWorkout") }}
              </Button>
              <Button variant="outline" size="sm" @click="emits('edit', workout)">
                {{ $t("edit") }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
