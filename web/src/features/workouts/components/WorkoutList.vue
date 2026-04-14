<script setup lang="ts">
import { ref } from "vue";
import type { Workout } from "@/features/workouts/types";
import Card from "@/components/ui/card/Card.vue";
import CardTitle from "@/components/ui/card/CardTitle.vue";
import CardDescription from "@/components/ui/card/CardDescription.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Dumbbell, History, Play } from "lucide-vue-next";
import { useI18n } from "vue-i18n";
import WorkoutHistoryDrawer from "@/features/workout-session/components/WorkoutHistoryDrawer.vue";

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

const historyDrawerOpen = ref(false);
const selectedWorkout = ref<Workout | null>(null);

function openHistory(workout: Workout) {
  selectedWorkout.value = workout;
  historyDrawerOpen.value = true;
}

const getExerciseCount = (workout: Workout): number =>
  workout.exercises?.length ?? 0;

const getTotalSets = (workout: Workout): number =>
  workout.exercises?.reduce((total, ex) => total + (ex.sets?.length ?? 0), 0) ??
  0;
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

    <div v-else class="flex flex-col gap-3">
      <Card v-for="workout in workouts" :key="workout.id" class="gap-2">
        <CardContent>
          <div class="flex gap-3 items-center">
            <div
              class="shrink-0 w-12 h-12 rounded-lg border overflow-hidden bg-muted flex items-center justify-center"
            >
              <Dumbbell class="w-6 h-6 text-muted-foreground" />
            </div>
            <div class="flex-1 min-w-0 flex flex-col gap-0.5">
              <CardTitle class="truncate">{{ workout.name }}</CardTitle>
              <p class="text-xs text-muted-foreground">
                {{ getExerciseCount(workout) }} {{ $t("workouts.exercises") }} ·
                {{ getTotalSets(workout) }} {{ $t("workouts.sets") }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                class="size-8 shrink-0"
                @click="openHistory(workout)"
              >
                <History class="size-4" />
              </Button>
              <div class="flex-1" />
              <Button
                variant="outline"
                size="sm"
                @click="emits('edit', workout)"
              >
                {{ $t("edit") }}
              </Button>
              <Button
                variant="default"
                size="sm"
                :disabled="startingWorkoutId != null"
                @click="emits('start', workout)"
              >
                <Play class="w-3.5 h-3.5 mr-1" />
                {{ $t("workoutSession.startWorkout") }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>

  <WorkoutHistoryDrawer
    v-model:open="historyDrawerOpen"
    :workout-id="selectedWorkout?.id ?? null"
    :workout-name="selectedWorkout?.name"
  />
</template>
