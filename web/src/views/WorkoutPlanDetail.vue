<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkoutPlan } from "@/features/workout-plans/composables/useWorkoutPlan";
import { useWorkouts } from "@/features/workouts/composables/useWorkouts";
import type { Workout, WorkoutExercise } from "@/features/workouts/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowLeft, Calendar, Pencil } from "lucide-vue-next";
import ExerciseInfoDialog from "@/features/exercises/components/ExerciseInfoDialog.vue";
import { useI18n } from "vue-i18n";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const planId = computed(() => Number(route.params.id));
const { plan, loading } = useWorkoutPlan(planId);

const { workouts: allWorkouts } = useWorkouts({ limit: 1000, includeAll: true });
const workoutMap = computed(() => {
  const map = new Map<number, Workout>();
  for (const w of allWorkouts.value) map.set(w.id, w);
  return map;
});

const featureUnitMap: Record<string, string> = {
  weight: "kg",
  rep: "reps",
  duration: "s",
  distance: "m",
};

function featureLabel(name: string): string {
  const unit = featureUnitMap[name];
  const label = name.charAt(0).toUpperCase() + name.slice(1);
  return unit ? `${label} (${unit})` : label;
}

function formatValue(name: string, value: number): string {
  const formatted = Number.isInteger(value) ? value.toString() : value.toFixed(1);
  const unit = featureUnitMap[name];
  return unit ? `${formatted} ${unit}` : formatted;
}

function exerciseFeatureNames(ex: WorkoutExercise): string[] {
  return ex.sets[0]?.features.map((f) => f.feature_name) ?? [];
}

const selectedWeek = ref(0);
</script>

<template>
  <div>
    <div class="mb-8 flex items-center gap-4">
      <Button variant="ghost" size="icon" @click="router.push({ name: 'workout-plans' })">
        <ArrowLeft class="h-4 w-4" />
      </Button>
      <div class="flex-1 min-w-0">
        <div v-if="loading" class="h-8 w-48 bg-muted animate-pulse rounded" />
        <h1 v-else class="text-3xl font-bold">{{ plan?.name }}</h1>
        <p v-if="plan?.description" class="text-muted-foreground mt-1">{{ plan.description }}</p>
      </div>
      <Button
        v-if="plan"
        variant="outline"
        size="sm"
        @click="router.push({ name: 'workout-plan-edit', params: { id: planId } })"
      >
        <Pencil class="h-4 w-4 mr-2" />
        {{ $t("edit") }}
      </Button>
    </div>

    <div v-if="loading" class="space-y-6">
      <div v-for="i in 3" :key="i" class="space-y-3">
        <div class="h-6 w-24 bg-muted animate-pulse rounded" />
        <div v-for="j in 3" :key="j" class="h-24 w-full bg-muted animate-pulse rounded-lg" />
      </div>
    </div>

    <template v-else-if="plan">
      <div class="flex items-center gap-2 mb-8 text-sm text-muted-foreground">
        <Calendar class="h-4 w-4" />
        <span>{{ $t("workoutPlans.weeksDays", { weeks: plan.numberOfWeeks, days: plan.daysPerWeek }) }}</span>
      </div>

      <div class="flex flex-wrap justify-center gap-2 mb-6">
        <Button
          v-for="(_, weekIdx) in plan.weeks"
          :key="weekIdx"
          type="button"
          :variant="selectedWeek === weekIdx ? 'default' : 'outline'"
          class="min-w-12 h-10 text-base font-semibold"
          @click="selectedWeek = weekIdx"
        >
          {{ weekIdx + 1 }}
        </Button>
      </div>

      <div class="space-y-3">
        <template v-if="plan.weeks[selectedWeek]">
            <Card v-for="(day, dayIdx) in plan.weeks[selectedWeek].days" :key="dayIdx">
              <CardHeader class="pb-2">
                <CardTitle class="text-sm font-semibold flex items-center gap-2">
                  {{ $t("workoutPlans.dayLabel", { number: dayIdx + 1 }) }}
                </CardTitle>
                <p v-if="day.description" class="text-sm text-muted-foreground mt-0.5">
                  {{ day.description }}
                </p>
              </CardHeader>

              <CardContent
                  v-if="day.workoutIds.length === 0"
                  class="pt-0 pb-4"
                >
                  <p class="text-sm text-muted-foreground">{{ $t("workoutPlans.detail.noWorkouts") }}</p>
                </CardContent>

                <CardContent v-else class="pt-0 pb-4 space-y-6">
                  <div
                    v-for="workoutId in day.workoutIds"
                    :key="workoutId"
                  >
                    <template v-if="workoutMap.get(workoutId) as Workout | undefined">
                      <p class="text-sm font-medium mb-3">
                        {{ workoutMap.get(workoutId)!.name }}
                      </p>

                      <div
                        v-if="workoutMap.get(workoutId)!.exercises.length === 0"
                        class="text-sm text-muted-foreground"
                      >
                        {{ $t("workoutPlans.detail.noExercises") }}
                      </div>

                      <div v-else class="space-y-4">
                        <div
                          v-for="ex in workoutMap.get(workoutId)!.exercises"
                          :key="ex.id ?? ex.order"
                        >
                          <div class="flex items-center gap-1 mb-1.5">
                            <span class="text-sm font-medium">{{ ex.exercise?.name ?? `#${ex.exercise_id}` }}</span>
                            <span
                              v-if="ex.exercise?.primary_muscle_groups?.length"
                              class="text-xs text-muted-foreground"
                            >
                              · {{ ex.exercise.primary_muscle_groups.map((m) => m.name).join(", ") }}
                            </span>
                            <ExerciseInfoDialog
                              :exercise="ex.exercise"
                            />
                          </div>

                          <div v-if="ex.sets.length > 0" class="rounded-md border overflow-hidden text-xs">
                            <table class="w-full">
                              <thead>
                                <tr class="bg-muted/50 text-muted-foreground">
                                  <th class="px-3 py-1.5 text-left font-medium w-10">
                                    {{ $t("exercises.logs.set") }}
                                  </th>
                                  <th
                                    v-for="fname in exerciseFeatureNames(ex)"
                                    :key="fname"
                                    class="px-3 py-1.5 text-right font-medium"
                                  >
                                    {{ featureLabel(fname) }}
                                  </th>
                                </tr>
                              </thead>
                              <tbody class="divide-y divide-border">
                                <tr
                                  v-for="(set, sIdx) in ex.sets"
                                  :key="set.id ?? sIdx"
                                  class="hover:bg-muted/30"
                                >
                                  <td class="px-3 py-1.5 text-muted-foreground font-medium">{{ sIdx + 1 }}</td>
                                  <td
                                    v-for="feat in set.features"
                                    :key="feat.feature_name"
                                    class="px-3 py-1.5 text-right tabular-nums"
                                  >
                                    {{ formatValue(feat.feature_name, feat.value) }}
                                  </td>
                                </tr>
                              </tbody>
                            </table>
                          </div>

                          <p v-if="ex.note" class="text-xs text-muted-foreground mt-1.5 italic">
                            {{ ex.note }}
                          </p>
                        </div>
                      </div>
                    </template>

                    <p v-else class="text-sm text-muted-foreground italic">
                      {{ $t("workoutPlans.detail.workoutNotFound") }}
                    </p>
                  </div>
                </CardContent>
            </Card>
        </template>
      </div>
    </template>
  </div>
</template>
