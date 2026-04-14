<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useQueryClient } from "@tanstack/vue-query";
import { exerciseKeys } from "@/lib/queryKeys";
import { useExerciseLogs } from "@/features/exercises/composables/useExerciseLogs";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { ChevronLeft, ChevronRight } from "lucide-vue-next";

const props = defineProps<{
  open: boolean;
  exerciseId: number | null;
  exerciseName?: string;
}>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const limit = 10;
const offset = ref(0);
const fromDate = ref<string>("");
const toDate = ref<string>("");

const queryClient = useQueryClient();

watch(() => props.exerciseId, () => {
  offset.value = 0;
  fromDate.value = "";
  toDate.value = "";
});

watch(() => props.open, (open) => {
  if (open && props.exerciseId) {
    queryClient.invalidateQueries({ queryKey: exerciseKeys.all });
  }
});

watch([fromDate, toDate], () => { offset.value = 0; });

const { entries, total, loading } = useExerciseLogs(
  computed(() => props.exerciseId),
  limit,
  offset,
  computed(() => fromDate.value || null),
  computed(() => toDate.value || null),
);

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
</script>

<template>
  <Sheet :open="open" @update:open="(v: boolean) => emit('update:open', v)">
    <SheetContent class="sm:max-w-md flex flex-col gap-0 p-0">
      <SheetHeader>
        <SheetTitle>{{ exerciseName ?? $t("exercises.logs.title") }}</SheetTitle>
        <SheetDescription>{{ $t("exercises.logs.description") }}</SheetDescription>
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
            <div class="h-4 w-32 bg-muted animate-pulse rounded" />
            <div class="h-20 w-full bg-muted animate-pulse rounded" />
          </div>
        </div>

        <div
          v-else-if="entries.length === 0"
          class="flex flex-col items-center justify-center py-16 text-center gap-2"
        >
          <p class="text-muted-foreground text-sm">{{ $t("exercises.logs.noLogs") }}</p>
        </div>

        <div v-else class="relative">
          <!-- Timeline -->
          <div class="absolute left-3 top-2 bottom-2 w-px bg-border" />
          <div class="space-y-8">
            <div
              v-for="entry in entries"
              :key="entry.session_id"
              class="relative pl-10"
            >
              <!-- Timeline dot -->
              <div class="absolute left-0 top-1 flex items-center justify-center w-6 h-6 rounded-full bg-background border-2 border-primary" />

              <!-- Date + workout -->
              <div class="mb-2">
                <p class="text-sm font-semibold leading-none">{{ formatDate(entry.date) }}</p>
                <p v-if="entry.workout_name" class="text-xs text-muted-foreground mt-0.5">
                  {{ entry.workout_name }}
                </p>
              </div>

              <!-- Sets table -->
              <div class="rounded-lg border overflow-hidden text-sm">
                <table class="w-full">
                  <thead>
                    <tr class="bg-muted/50 text-xs text-muted-foreground">
                      <th class="px-3 py-2 text-left font-medium w-12">{{ $t("exercises.logs.set") }}</th>
                      <th
                        v-for="val in entry.sets[0]?.values ?? []"
                        :key="val.feature_name"
                        class="px-3 py-2 text-right font-medium"
                      >
                        {{ featureLabel(val.feature_name) }}
                      </th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-border">
                    <tr
                      v-for="(set, idx) in entry.sets"
                      :key="set.order"
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
      </div>

      <!-- Pagination -->
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
    </SheetContent>
  </Sheet>
</template>
