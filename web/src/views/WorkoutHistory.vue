<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useWorkoutSessions } from "@/features/workout-session/composables/useWorkoutSessions";
import { useWorkoutSession } from "@/features/workout-session/composables/useWorkoutSession";
import type { WorkoutSessionSummary } from "@/features/workout-session/composables/useWorkoutSessions";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import Card from "@/components/ui/card/Card.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Skeleton } from "@/components/ui/skeleton";
import { Dumbbell } from "lucide-vue-next";

const limit = 20;
const offset = ref(0);
const currentPage = ref(1);

const { sessions, total, loading } = useWorkoutSessions(limit, offset);

watch(currentPage, (page) => {
  offset.value = (page - 1) * limit;
});

const selectedSessionId = ref<number | null>(null);
const drawerOpen = ref(false);

const { session: detailSession, loading: detailLoading } = useWorkoutSession(
  computed(() => selectedSessionId.value) as unknown as number | null,
);

function openSession(summary: WorkoutSessionSummary) {
  selectedSessionId.value = summary.id;
  drawerOpen.value = true;
}

watch(drawerOpen, (open) => {
  if (!open) selectedSessionId.value = null;
});

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    weekday: "short",
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatDuration(startIso: string, endIso: string | null): string {
  if (!endIso) return "";
  const ms = new Date(endIso).getTime() - new Date(startIso).getTime();
  const mins = Math.floor(ms / 60000);
  if (mins < 60) return `${mins}m`;
  const hours = Math.floor(mins / 60);
  const rem = mins % 60;
  return rem > 0 ? `${hours}h ${rem}m` : `${hours}h`;
}

const featureUnitMap: Record<string, string> = {
  weight: "kg",
  rep: "reps",
  duration: "s",
  distance: "m",
};

function formatValue(featureName: string, value: number): string {
  const formatted = Number.isInteger(value) ? value.toString() : value.toFixed(1);
  const unit = featureUnitMap[featureName];
  return unit ? `${formatted} ${unit}` : formatted;
}

function featureLabel(name: string): string {
  const unit = featureUnitMap[name];
  const label = name.charAt(0).toUpperCase() + name.slice(1);
  return unit ? `${label} (${unit})` : label;
}

const completedExercises = computed(() => {
  if (!detailSession.value) return [];
  return detailSession.value.exercises.filter((ex) =>
    ex.sets.some((s) => s.completed_at),
  );
});
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-3xl font-bold mb-2">{{ $t("workoutHistory.title") }}</h1>
      <p class="text-muted-foreground">{{ $t("workoutHistory.subtitle") }}</p>
    </div>

    <!-- Loading skeletons -->
    <div v-if="loading" class="space-y-4">
      <Card v-for="i in 5" :key="i">
        <CardContent class="py-4">
          <div class="flex items-center justify-between">
            <div class="space-y-2">
              <Skeleton class="h-5 w-40" />
              <Skeleton class="h-4 w-28" />
            </div>
            <div class="flex gap-4 text-right">
              <Skeleton class="h-4 w-16" />
              <Skeleton class="h-4 w-16" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Empty state -->
    <div
      v-else-if="sessions.length === 0"
      class="flex flex-col items-center justify-center py-24 gap-4 text-center"
    >
      <Dumbbell class="w-12 h-12 text-muted-foreground" />
      <p class="text-muted-foreground">{{ $t("workoutHistory.noHistory") }}</p>
    </div>

    <!-- Session list -->
    <div v-else class="space-y-3">
      <Card
        v-for="session in sessions"
        :key="session.id"
        class="cursor-pointer hover:bg-muted/30 transition-colors"
        @click="openSession(session)"
      >
        <CardContent class="py-4">
          <div class="flex items-center justify-between gap-4">
            <div class="min-w-0">
              <p class="font-semibold truncate">
                {{ session.workout_name || $t("workoutHistory.unnamedWorkout") }}
              </p>
              <p class="text-sm text-muted-foreground mt-0.5">
                {{ formatDate(session.started_at) }}
                <span class="mx-1">·</span>
                {{ formatTime(session.started_at) }}
              </p>
            </div>
            <div class="flex items-center gap-6 shrink-0 text-sm text-muted-foreground">
              <div class="text-right">
                <p class="font-medium text-foreground">
                  {{ session.exercise_count }}
                </p>
                <p class="text-xs">{{ $t("workoutHistory.exercises") }}</p>
              </div>
              <div class="text-right">
                <p class="font-medium text-foreground">
                  {{ session.sets_completed }}
                </p>
                <p class="text-xs">{{ $t("workoutHistory.sets") }}</p>
              </div>
              <div v-if="session.ended_at" class="text-right">
                <p class="font-medium text-foreground">
                  {{ formatDuration(session.started_at, session.ended_at) }}
                </p>
                <p class="text-xs">{{ $t("workoutHistory.duration") }}</p>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Pagination -->
    <div v-if="!loading && total > limit" class="mt-8 flex flex-col gap-2 items-center">
      <Pagination
        v-slot="{ page }"
        v-model:page="currentPage"
        :items-per-page="limit"
        :total="total"
        :default-page="currentPage"
      >
        <PaginationContent v-slot="{ items }">
          <PaginationPrevious />
          <template v-for="(item, index) in items" :key="index">
            <PaginationItem v-if="item.type === 'page'" :value="item.value" :is-active="item.value === page">
              {{ item.value }}
            </PaginationItem>
            <PaginationEllipsis v-else :key="item.type" />
          </template>
          <PaginationNext />
        </PaginationContent>
      </Pagination>
      <p class="text-sm text-muted-foreground">
        {{
          $t("pagination.showingFromToOfTotal", {
            from: offset + 1,
            to: Math.min(offset + limit, total),
            total,
          })
        }}
        {{ $t("workoutHistory.titleLower") }}
      </p>
    </div>

    <!-- Detail sidebar -->
    <Sheet :open="drawerOpen" @update:open="(v: boolean) => (drawerOpen = v)">
      <SheetContent class="sm:max-w-md flex flex-col gap-0 p-0">
        <SheetHeader>
          <SheetTitle>
            {{ detailSession?.workout?.name || $t("workoutHistory.unnamedWorkout") }}
          </SheetTitle>
          <SheetDescription v-if="detailSession">
            {{ formatDate(detailSession.started_at) }}
            <span v-if="detailSession.ended_at">
              · {{ formatDuration(detailSession.started_at, detailSession.ended_at) }}
            </span>
          </SheetDescription>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-4 min-h-0">
          <!-- Loading -->
          <div v-if="detailLoading" class="space-y-6">
            <div v-for="i in 3" :key="i" class="space-y-2">
              <Skeleton class="h-4 w-32" />
              <Skeleton class="h-20 w-full" />
            </div>
          </div>

          <!-- Exercises -->
          <div v-else-if="completedExercises.length > 0" class="relative">
            <!-- Timeline line -->
            <div class="absolute left-3 top-2 bottom-2 w-px bg-border" />
            <div class="space-y-8">
              <div
                v-for="exercise in completedExercises"
                :key="exercise.id"
                class="relative pl-10"
              >
                <!-- Timeline dot -->
                <div class="absolute left-0 top-1 flex items-center justify-center w-6 h-6 rounded-full bg-background border-2 border-primary" />

                <p class="text-sm font-semibold leading-none mb-2">
                  {{ exercise.exercise?.name ?? $t("workoutHistory.unknownExercise") }}
                </p>

                <!-- Sets table -->
                <div class="rounded-lg border overflow-hidden text-sm">
                  <table class="w-full">
                    <thead>
                      <tr class="bg-muted/50 text-xs text-muted-foreground">
                        <th class="px-3 py-2 text-left font-medium w-12">
                          {{ $t("exercises.logs.set") }}
                        </th>
                        <th
                          v-for="val in exercise.sets.find(s => s.completed_at)?.values ?? []"
                          :key="val.feature_name"
                          class="px-3 py-2 text-right font-medium"
                        >
                          {{ featureLabel(val.feature_name) }}
                        </th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-border">
                      <tr
                        v-for="(set, idx) in exercise.sets.filter(s => s.completed_at)"
                        :key="set.id"
                        class="hover:bg-muted/30"
                      >
                        <td class="px-3 py-2 text-muted-foreground font-medium">{{ idx + 1 }}</td>
                        <td
                          v-for="val in set.values"
                          :key="val.feature_name"
                          class="px-3 py-2 text-right tabular-nums"
                        >
                          {{ formatValue(val.feature_name, val.value) }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="flex flex-col items-center justify-center py-16 text-center gap-2">
            <p class="text-muted-foreground text-sm">{{ $t("workoutHistory.noSetsCompleted") }}</p>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
