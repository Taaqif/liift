<script setup lang="ts">
import type { Workout } from "@/features/workouts/types";
import Card from "@/components/ui/card/Card.vue";
import CardHeader from "@/components/ui/card/CardHeader.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Calendar, Play } from "lucide-vue-next";

const props = defineProps<{
  workouts: Workout[];
  loading?: boolean;
  startingWorkoutId?: number | null;
}>();

const emits = defineEmits<{
  (e: "edit", workout: Workout): void;
  (e: "start", workout: Workout): void;
}>();

const handleEdit = (workout: Workout) => {
  emits("edit", workout);
};

const handleStart = (workout: Workout) => {
  emits("start", workout);
};

const getExerciseCount = (workout: Workout): number => {
  return workout.exercises?.length ?? 0;
};

const getTotalSets = (workout: Workout): number => {
  return workout.exercises?.reduce((total, ex) => total + (ex.sets?.length ?? 0), 0) ?? 0;
};
</script>

<template>
  <div class="space-y-4">
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card v-for="i in 6" :key="i">
        <CardHeader>
          <CardTitle>
            <div class="h-6 w-48 bg-gray-200 animate-pulse rounded"></div>
          </CardTitle>
          <CardDescription>
            <div class="h-4 w-full bg-gray-200 animate-pulse rounded mt-2"></div>
          </CardDescription>
        </CardHeader>
      </Card>
    </div>

    <div v-else-if="workouts.length === 0" class="text-center py-12">
      <p class="text-muted-foreground">{{ $t("workouts.noWorkouts") }}</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card v-for="workout in workouts" :key="workout.id" class="hover:shadow-md transition-shadow">
        <CardHeader>
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1">
              <CardTitle class="line-clamp-2">{{ workout.name }}</CardTitle>
              <CardDescription v-if="workout.description" class="line-clamp-2 mt-2">
                {{ workout.description }}
              </CardDescription>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <Button
                variant="default"
                size="sm"
                :disabled="startingWorkoutId != null"
                @click="handleStart(workout)"
              >
                <Play class="w-4 h-4 mr-1" />
                {{ $t("workoutSession.startWorkout") }}
              </Button>
              <Button
                variant="outline"
                size="sm"
                @click="handleEdit(workout)"
              >
                {{ $t("edit") }}
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div class="flex items-center gap-4 text-sm text-muted-foreground">
            <div class="flex items-center gap-1">
              <Calendar class="w-4 h-4" />
              <span>{{ getExerciseCount(workout) }} {{ $t("workouts.exercises") }}</span>
            </div>
            <div class="flex items-center gap-1">
              <span>{{ getTotalSets(workout) }} {{ $t("workouts.sets") }}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
