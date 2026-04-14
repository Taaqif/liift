<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useQueryClient } from "@tanstack/vue-query";
import { workoutSessionKeys } from "@/lib/queryKeys";
import { useWorkoutSessions } from "@/features/workout-session/composables/useWorkoutSessions";
import { useWorkoutSession } from "@/features/workout-session/composables/useWorkoutSession";
import { useDeleteWorkoutSession } from "@/features/workout-session/composables/useDeleteWorkoutSession";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@/components/ui/sheet";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ChevronLeft, ChevronRight, Trash2 } from "lucide-vue-next";

const props = defineProps<{
  open: boolean;
  workoutId: number | null;
  workoutName?: string;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const limit = 10;
const offset = ref(0);
const selectedSessionId = ref<number | null>(null);
const fromDate = ref<string>("");
const toDate = ref<string>("");

const queryClient = useQueryClient();

watch(() => props.workoutId, () => {
  offset.value = 0;
  selectedSessionId.value = null;
  fromDate.value = "";
  toDate.value = "";
});

watch(() => props.open, (open) => {
  if (open && props.workoutId) {
    queryClient.invalidateQueries({ queryKey: workoutSessionKeys.all });
  }
  if (!open) selectedSessionId.value = null;
});

watch([fromDate, toDate], () => { offset.value = 0; });

const { sessions, total, loading } = useWorkoutSessions(
  limit,
  offset,
  computed(() => props.workoutId),
  null,
  computed(() => fromDate.value || null),
  computed(() => toDate.value || null),
);

const { session: detailSession, loading: detailLoading } = useWorkoutSession(
  computed(() => selectedSessionId.value) as unknown as number | null,
);

const { deleteSession, isPending: isDeleting } = useDeleteWorkoutSession();
const deleteDialogOpen = ref(false);

async function handleDelete() {
  if (!selectedSessionId.value) return;
  await deleteSession(selectedSessionId.value);
  deleteDialogOpen.value = false;
  selectedSessionId.value = null;
}

const currentPage = computed(() => Math.floor(offset.value / limit) + 1);
const totalPages = computed(() => Math.ceil(total.value / limit));

function prevPage() {
  if (offset.value >= limit) offset.value -= limit;
}

function nextPage() {
  if (offset.value + limit < total.value) offset.value += limit;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    weekday: "short",
    year: "numeric",
    month: "short",
    day: "numeric",
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

function featureLabel(name: string): string {
  const unit = featureUnitMap[name];
  const label = name.charAt(0).toUpperCase() + name.slice(1);
  return unit ? `${label} (${unit})` : label;
}

function formatValue(featureName: string, value: number): string {
  const formatted = Number.isInteger(value) ? value.toString() : value.toFixed(1);
  const unit = featureUnitMap[featureName];
  return unit ? `${formatted} ${unit}` : formatted;
}

const completedExercises = computed(() => {
  if (!detailSession.value) return [];
  return detailSession.value.exercises.filter((ex) =>
    ex.sets.some((s) => s.completed_at),
  );
});
</script>

<template>
  <Sheet :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <!-- Delete confirmation dialog — inside Sheet so it layers above the sheet overlay -->
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
      <!-- Detail view -->
      <template v-if="selectedSessionId">
        <SheetHeader>
          <div class="flex items-center gap-2">
            <Button variant="ghost" size="icon" class="size-7 shrink-0" @click="selectedSessionId = null">
              <ChevronLeft class="size-4" />
            </Button>
            <div class="flex-1 min-w-0">
              <SheetTitle>{{ workoutName ?? $t("workoutHistory.unnamedWorkout") }}</SheetTitle>
              <SheetDescription v-if="detailSession">
                {{ formatDate(detailSession.started_at) }}
                <span v-if="detailSession.ended_at">
                  · {{ formatDuration(detailSession.started_at, detailSession.ended_at) }}
                </span>
              </SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <div class="flex-1 overflow-y-auto px-6 py-4 min-h-0">
          <div v-if="detailLoading" class="space-y-6">
            <div v-for="i in 3" :key="i" class="space-y-2">
              <Skeleton class="h-4 w-32" />
              <Skeleton class="h-20 w-full" />
            </div>
          </div>

          <div v-else-if="completedExercises.length > 0" class="relative">
            <div class="absolute left-3 top-2 bottom-2 w-px bg-border" />
            <div class="space-y-8">
              <div
                v-for="exercise in completedExercises"
                :key="exercise.id"
                class="relative pl-10"
              >
                <div class="absolute left-0 top-1 flex items-center justify-center w-6 h-6 rounded-full bg-background border-2 border-primary" />
                <p class="text-sm font-semibold leading-none mb-2">
                  {{ exercise.exercise?.name ?? $t("workoutHistory.unknownExercise") }}
                </p>
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

        <SheetFooter class="px-6 pb-6 pt-2">
          <Button variant="ghost" class="w-full text-destructive hover:text-destructive hover:bg-destructive/10" @click="deleteDialogOpen = true">
            <Trash2 class="size-4 mr-2" />
            {{ $t("workoutHistory.deleteRecord") }}
          </Button>
        </SheetFooter>
      </template>

      <!-- List view -->
      <template v-else>
        <SheetHeader>
          <SheetTitle>{{ workoutName ?? $t("workoutHistory.unnamedWorkout") }}</SheetTitle>
          <SheetDescription>{{ $t("workoutHistory.drawerDescription") }}</SheetDescription>
        </SheetHeader>

        <div class="px-6 pb-3 flex items-center gap-2">
          <div class="flex-1 flex flex-col gap-0.5">
            <label class="text-xs text-muted-foreground">{{ $t("filters.from") }}</label>
            <input
              v-model="fromDate"
              type="date"
              :max="toDate || undefined"
              class="h-8 w-full rounded-md border border-input bg-background px-2 text-sm focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
          <div class="flex-1 flex flex-col gap-0.5">
            <label class="text-xs text-muted-foreground">{{ $t("filters.to") }}</label>
            <input
              v-model="toDate"
              type="date"
              :min="fromDate || undefined"
              class="h-8 w-full rounded-md border border-input bg-background px-2 text-sm focus:outline-none focus:ring-1 focus:ring-ring"
            />
          </div>
          <Button
            v-if="fromDate || toDate"
            variant="ghost"
            size="sm"
            class="mt-4 px-2 text-muted-foreground hover:text-foreground"
            @click="fromDate = ''; toDate = ''"
          >
            {{ $t("filters.clear") }}
          </Button>
        </div>

        <div class="flex-1 overflow-y-auto px-6 py-4 min-h-0">
          <div v-if="loading" class="space-y-4">
            <div v-for="i in 3" :key="i" class="space-y-2">
              <Skeleton class="h-4 w-32" />
              <Skeleton class="h-16 w-full" />
            </div>
          </div>

          <div
            v-else-if="sessions.length === 0"
            class="flex flex-col items-center justify-center py-16 text-center gap-2"
          >
            <p class="text-muted-foreground text-sm">{{ $t("workoutHistory.noHistory") }}</p>
          </div>

          <div v-else class="relative">
            <div class="absolute left-3 top-2 bottom-2 w-px bg-border" />
            <div class="space-y-6">
              <button
                v-for="session in sessions"
                :key="session.id"
                class="relative pl-10 w-full text-left group"
                @click="selectedSessionId = session.id"
              >
                <div class="absolute left-0 top-1 flex items-center justify-center w-6 h-6 rounded-full bg-background border-2 border-primary group-hover:border-primary/70 transition-colors" />

                <div class="flex items-start justify-between gap-2">
                  <div>
                    <p class="text-sm font-semibold leading-none">
                      {{ formatDate(session.started_at) }}
                    </p>
                    <p class="text-xs text-muted-foreground mt-1">
                      {{ session.exercise_count }} {{ $t("workoutHistory.exercises") }}
                      · {{ session.sets_completed }} {{ $t("workoutHistory.sets") }}
                      <span v-if="session.ended_at">
                        · {{ formatDuration(session.started_at, session.ended_at) }}
                      </span>
                    </p>
                  </div>
                  <ChevronRight class="size-4 text-muted-foreground mt-0.5 shrink-0 group-hover:text-foreground transition-colors" />
                </div>
              </button>
            </div>
          </div>
        </div>

        <SheetFooter v-if="totalPages > 1" class="flex items-center justify-between gap-2">
          <Button variant="outline" size="icon" :disabled="currentPage <= 1" @click="prevPage">
            <ChevronLeft class="h-4 w-4" />
          </Button>
          <span class="text-sm text-muted-foreground">
            {{ $t("exercises.logs.pageOf", { current: currentPage, total: totalPages }) }}
          </span>
          <Button variant="outline" size="icon" :disabled="currentPage >= totalPages" @click="nextPage">
            <ChevronRight class="h-4 w-4" />
          </Button>
        </SheetFooter>
      </template>
    </SheetContent>
  </Sheet>
</template>
