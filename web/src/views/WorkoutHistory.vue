<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { today, getLocalTimeZone, CalendarDate } from "@internationalized/date";
// CalendarDate has #private in newer @internationalized/date versions — keep refs as `any`
import { Calendar } from "@/components/ui/calendar";
import { useWorkoutActivityDates } from "@/features/workout-session/composables/useWorkoutActivityDates";
import { useWorkoutSessions } from "@/features/workout-session/composables/useWorkoutSessions";
import { useWorkoutSession } from "@/features/workout-session/composables/useWorkoutSession";
import { useDeleteWorkoutSession } from "@/features/workout-session/composables/useDeleteWorkoutSession";
import type { WorkoutSessionSummary } from "@/features/workout-session/composables/useWorkoutSessions";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
} from "@/components/ui/sheet";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import Card from "@/components/ui/card/Card.vue";
import CardContent from "@/components/ui/card/CardContent.vue";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Dumbbell, ChevronRight, CalendarDays, Clock, Layers, Flame, Trash2 } from "lucide-vue-next";

// ── Date state ────────────────────────────────────────────
const t = today(getLocalTimeZone());
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const selectedDate = ref<any>(new CalendarDate(t.year, t.month, t.day));
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const calendarPlaceholder = ref<any>(new CalendarDate(t.year, t.month, t.day));

const visibleYear = computed<number>(() => calendarPlaceholder.value?.year ?? t.year);
const visibleMonth = computed<number>(() => calendarPlaceholder.value?.month ?? t.month);

// ── Queries ───────────────────────────────────────────────
const { activityDates, loading: activityLoading } = useWorkoutActivityDates(visibleYear, visibleMonth);

const selectedDateKey = computed<string>(() => {
  const d = selectedDate.value;
  if (!d) return "";
  return `${d.year}-${String(d.month).padStart(2, "0")}-${String(d.day).padStart(2, "0")}`;
});

const { sessions: daySessions, loading: dayLoading } = useWorkoutSessions(100, 0, null, selectedDateKey);

// ── Detail sheet ──────────────────────────────────────────
const selectedSessionId = ref<number | null>(null);
const drawerOpen = ref(false);

// Pass reactive ref directly — useWorkoutSession now accepts MaybeRefOrGetter
const { session: detailSession, loading: detailLoading } = useWorkoutSession(selectedSessionId);
const { deleteSession, isPending: isDeleting } = useDeleteWorkoutSession();

function openSession(s: WorkoutSessionSummary) {
  selectedSessionId.value = s.id;
  drawerOpen.value = true;
}

const deleteDialogOpen = ref(false);

async function handleDelete() {
  if (!selectedSessionId.value) return;
  await deleteSession(selectedSessionId.value);
  deleteDialogOpen.value = false;
  drawerOpen.value = false;
}

watch(drawerOpen, (open) => {
  if (!open) selectedSessionId.value = null;
});

// ── Formatters ────────────────────────────────────────────
const selectedDateLabel = computed(() => {
  const d = selectedDate.value;
  if (!d) return "";
  const date = new Date(d.year, d.month - 1, d.day);
  const isToday = date.toDateString() === new Date().toDateString();
  if (isToday) return "Today";
  return date.toLocaleDateString(undefined, { weekday: "long", month: "long", day: "numeric" });
});

const dayStats = computed(() => ({
  exercises: daySessions.value.reduce((n, s) => n + s.exercise_count, 0),
  sets: daySessions.value.reduce((n, s) => n + s.sets_completed, 0),
  durationMs: daySessions.value.reduce((sum, s) => {
    if (!s.ended_at) return sum;
    return sum + (new Date(s.ended_at).getTime() - new Date(s.started_at).getTime());
  }, 0),
}));

function formatTime(iso: string): string {
  return new Date(iso).toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" });
}

function formatDuration(startIso: string, endIso: string | null): string {
  if (!endIso) return "";
  const ms = new Date(endIso).getTime() - new Date(startIso).getTime();
  const mins = Math.floor(ms / 60000);
  if (mins < 60) return `${mins}m`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

function formatMs(ms: number): string {
  const mins = Math.floor(ms / 60000);
  if (mins < 60) return `${mins}m`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m > 0 ? `${h}h ${m}m` : `${h}h`;
}

const featureUnitMap: Record<string, string> = {
  weight: "kg",
  rep: "reps",
  duration: "s",
  distance: "m",
};

function formatValue(name: string, value: number): string {
  const formatted = Number.isInteger(value) ? value.toString() : value.toFixed(1);
  const unit = featureUnitMap[name];
  return unit ? `${formatted} ${unit}` : formatted;
}

function featureLabel(name: string): string {
  const unit = featureUnitMap[name];
  const label = name.charAt(0).toUpperCase() + name.slice(1);
  return unit ? `${label} (${unit})` : label;
}

const completedExercises = computed(() => {
  if (!detailSession.value) return [];
  return detailSession.value.exercises.filter((ex) => ex.sets.some((s) => s.completed_at));
});

</script>

<template>
  <div class="pb-10">
    <!-- Header -->
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold mb-2">{{ $t("workoutHistory.title") }}</h1>
        <p class="text-muted-foreground">{{ $t("workoutHistory.subtitle") }}</p>
      </div>
    </div>

    <div class="flex flex-col lg:flex-row gap-6 lg:gap-8">

    <!-- Left: calendar sidebar -->
    <div class="lg:w-72 lg:shrink-0 space-y-3">
      <Card class="overflow-hidden">
        <CardContent class="p-0">
          <Calendar
            :model-value="selectedDate"
            @update:model-value="(v: any) => { if (v) selectedDate = v }"
            v-model:placeholder="calendarPlaceholder"
            :activity-dates="activityDates"
            :fixed-weeks="true"
          />
        </CardContent>
      </Card>

      <!-- Subtle month-load indicator -->
      <div v-if="activityLoading" class="h-0.5 rounded-full bg-muted overflow-hidden">
        <div class="h-full w-2/5 bg-primary/50 rounded-full animate-pulse mx-auto" />
      </div>
    </div>

    <!-- Right: sessions list -->
    <div class="flex-1 min-w-0 space-y-4">
      <!-- Date label + stats row -->
      <div class="flex items-baseline justify-between gap-2 min-h-[1.5rem]">
        <span class="font-semibold text-base">{{ selectedDateLabel }}</span>
        <div v-if="daySessions.length > 0" class="flex gap-3 text-xs text-muted-foreground">
          <span class="flex items-center gap-1"><Layers class="w-3.5 h-3.5" />{{ dayStats.exercises }}</span>
          <span class="flex items-center gap-1"><Flame class="w-3.5 h-3.5" />{{ dayStats.sets }} sets</span>
          <span v-if="dayStats.durationMs > 0" class="flex items-center gap-1"><Clock class="w-3.5 h-3.5" />{{ formatMs(dayStats.durationMs) }}</span>
        </div>
      </div>

      <!-- Skeletons -->
      <div v-if="dayLoading" class="space-y-3">
        <Skeleton v-for="i in 2" :key="i" class="h-24 w-full rounded-2xl" />
      </div>

      <!-- Empty state -->
      <div
        v-else-if="daySessions.length === 0"
        class="flex flex-col items-center gap-2 py-12 text-muted-foreground/60"
      >
        <CalendarDays class="w-8 h-8" />
        <p class="text-sm">No workouts on this day</p>
      </div>

      <!-- Session cards -->
      <div v-else class="space-y-3">
        <button
          v-for="session in daySessions"
          :key="session.id"
          class="w-full text-left group"
          @click="openSession(session)"
        >
          <div
            class="rounded-2xl border bg-card overflow-hidden shadow-sm hover:shadow-md active:scale-[0.99] transition-all duration-150"
          >
            <div class="px-4 py-4">
              <!-- Top row: name + chevron -->
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <p class="font-semibold text-base leading-tight truncate">
                    {{ session.workout_name || $t("workoutHistory.unnamedWorkout") }}
                  </p>
                  <p class="text-xs text-muted-foreground mt-0.5 flex items-center gap-1.5">
                    <Clock class="w-3 h-3 shrink-0" />
                    {{ formatTime(session.started_at) }}
                    <template v-if="session.ended_at">
                      <span class="opacity-40">·</span>
                      {{ formatDuration(session.started_at, session.ended_at) }}
                    </template>
                  </p>
                </div>
                <ChevronRight
                  class="w-4 h-4 text-muted-foreground/40 group-hover:text-muted-foreground shrink-0 mt-0.5 transition-colors"
                />
              </div>

              <!-- Stats pills -->
              <div class="flex gap-2 mt-3 flex-wrap">
                <span
                  class="inline-flex items-center gap-1 text-xs font-medium bg-muted px-2.5 py-1 rounded-full"
                >
                  <Layers class="w-3 h-3 text-muted-foreground" />
                  {{ session.exercise_count }}
                  {{ $t("workoutHistory.exercises") }}
                </span>
                <span
                  class="inline-flex items-center gap-1 text-xs font-medium bg-muted px-2.5 py-1 rounded-full"
                >
                  <Flame class="w-3 h-3 text-muted-foreground" />
                  {{ session.sets_completed }} {{ $t("workoutHistory.sets") }}
                </span>
                <span
                  v-if="session.ended_at"
                  class="inline-flex items-center gap-1 text-xs font-medium bg-muted px-2.5 py-1 rounded-full"
                >
                  <Clock class="w-3 h-3 text-muted-foreground" />
                  {{ formatDuration(session.started_at, session.ended_at) }}
                </span>
              </div>
            </div>
          </div>
        </button>
      </div>
    </div>

    <!-- Detail sheet -->
    <Sheet :open="drawerOpen" @update:open="(v: boolean) => (drawerOpen = v)">
      <!-- Delete confirmation dialog — inside Sheet so it teleports above the sheet overlay -->
      <Dialog v-model:open="deleteDialogOpen">
        <DialogContent class="max-w-sm">
          <DialogHeader>
            <DialogTitle>{{ $t("workoutHistory.deleteConfirmTitle") }}</DialogTitle>
            <DialogDescription>{{ $t("workoutHistory.deleteConfirmDescription") }}</DialogDescription>
          </DialogHeader>
          <DialogFooter class="flex-col gap-2 sm:flex-row">
            <Button variant="outline" :disabled="isDeleting" @click="deleteDialogOpen = false">
              {{ $t("cancel") }}
            </Button>
            <Button variant="destructive" :disabled="isDeleting" @click="handleDelete">
              {{ isDeleting ? $t("deleting") : $t("delete") }}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <SheetContent class="sm:max-w-md flex flex-col gap-0 p-0">
        <SheetHeader class="px-6 pt-6 pb-4 border-b shrink-0">
          <SheetTitle class="leading-tight">
            {{ detailSession?.workout?.name || $t("workoutHistory.unnamedWorkout") }}
          </SheetTitle>
          <SheetDescription v-if="detailSession">
            {{
              new Date(detailSession.started_at).toLocaleDateString(undefined, {
                weekday: "short",
                month: "short",
                day: "numeric",
              })
            }}
            <span v-if="detailSession.ended_at">
              · {{ formatDuration(detailSession.started_at, detailSession.ended_at) }}
            </span>
          </SheetDescription>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-5 min-h-0">
          <!-- Loading -->
          <div v-if="detailLoading" class="space-y-6">
            <div v-for="i in 3" :key="i" class="space-y-2">
              <Skeleton class="h-4 w-32" />
              <Skeleton class="h-20 w-full rounded-lg" />
            </div>
          </div>

          <!-- Empty -->
          <div
            v-else-if="completedExercises.length === 0"
            class="flex flex-col items-center justify-center py-16 text-center gap-2"
          >
            <Dumbbell class="w-8 h-8 text-muted-foreground/40" />
            <p class="text-muted-foreground text-sm">{{ $t("workoutHistory.noSetsCompleted") }}</p>
          </div>

          <!-- Exercise timeline -->
          <div v-else class="relative">
            <div class="absolute left-3 top-2 bottom-2 w-px bg-border" />
            <div class="space-y-8">
              <div
                v-for="exercise in completedExercises"
                :key="exercise.id"
                class="relative pl-10"
              >
                <div
                  class="absolute left-0 top-1 w-6 h-6 rounded-full bg-background border-2 border-primary flex items-center justify-center"
                />
                <p class="text-sm font-semibold mb-2">
                  {{ exercise.exercise?.name ?? $t("workoutHistory.unknownExercise") }}
                </p>
                <div class="rounded-lg border overflow-hidden text-sm">
                  <table class="w-full">
                    <thead>
                      <tr class="bg-muted/50 text-xs text-muted-foreground">
                        <th class="px-3 py-2 text-left font-medium w-10">
                          {{ $t("exercises.logs.set") }}
                        </th>
                        <th
                          v-for="val in exercise.sets.find((s) => s.completed_at)?.values ?? []"
                          :key="val.feature_name"
                          class="px-3 py-2 text-right font-medium"
                        >
                          {{ featureLabel(val.feature_name) }}
                        </th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-border">
                      <tr
                        v-for="(set, i) in exercise.sets.filter((s) => s.completed_at)"
                        :key="set.id"
                        class="hover:bg-muted/30"
                      >
                        <td class="px-3 py-2 text-muted-foreground font-medium">{{ i + 1 }}</td>
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
        </div>

        <div class="px-6 py-4 border-t shrink-0">
          <Button
            variant="ghost"
            class="w-full text-destructive hover:text-destructive hover:bg-destructive/10"
            :disabled="isDeleting"
            @click="deleteDialogOpen = true"
          >
            <Trash2 class="size-4 mr-2" />
            {{ $t("workoutHistory.deleteRecord") }}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
    </div><!-- end flex lg:flex-row -->
  </div><!-- end pb-10 -->
</template>
