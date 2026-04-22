<script setup lang="ts">
import { DownloadIcon, AlertCircleIcon } from "lucide-vue-next";
import ExerciseInfoDialog from "@/features/exercises/components/ExerciseInfoDialog.vue";
import type { WorkoutExerciseArtifact } from "@/features/ai-chat/types";

type ExerciseStatus = "library" | "importable" | "unknown";

type ExerciseInfo = {
  name?: string;
  description?: string;
  instructions?: string[];
  image?: string;
  force?: string;
  category?: string;
  primary_muscle_groups?: { name: string }[];
  secondary_muscle_groups?: { name: string }[];
  equipment?: { name: string }[];
  exercise_features?: { name: string }[];
} | null;

defineProps<{
  exercises: WorkoutExerciseArtifact[];
  /** Resolve status for a given exercise */
  statusFn: (id: number, name: string) => ExerciseStatus;
  /** Resolve preview info for a given exercise */
  previewFn: (id: number, name: string) => ExerciseInfo;
  /** Compact mode — used inside plan day lists */
  compact?: boolean;
}>();
</script>

<template>
  <div :class="compact ? 'space-y-2' : 'space-y-4'">
    <div
      v-for="(ex, i) in exercises"
      :key="i"
      :class="compact ? 'space-y-1' : 'space-y-1.5'"
    >
      <!-- Exercise header row -->
      <div class="flex items-center gap-1.5 min-w-0">
        <span
          v-if="!compact"
          class="text-xs text-muted-foreground w-5 text-right shrink-0"
        >{{ i + 1 }}.</span>
        <span :class="compact ? 'text-xs font-medium' : 'font-medium text-sm'">{{ ex.exercise_name }}</span>
        <ExerciseInfoDialog :exercise="previewFn(ex.exercise_id, ex.exercise_name)" />
        <span
          v-if="statusFn(ex.exercise_id, ex.exercise_name) === 'importable'"
          class="flex items-center gap-0.5 text-xs text-muted-foreground shrink-0"
          title="Will be imported to your exercise library on save"
        >
          <DownloadIcon :class="compact ? 'size-2.5' : 'size-3'" />
          <span>import</span>
        </span>
        <span
          v-else-if="statusFn(ex.exercise_id, ex.exercise_name) === 'unknown'"
          class="flex items-center gap-0.5 text-xs text-destructive shrink-0"
          title="Not found in exercise library — ask the coach for an alternative"
        >
          <AlertCircleIcon :class="compact ? 'size-2.5' : 'size-3'" />
          <span>not found</span>
        </span>
      </div>

      <!-- Sets -->
      <div :class="compact ? 'ml-2 space-y-0.5' : 'ml-7 space-y-1'">
        <div
          v-for="(set, si) in ex.sets"
          :key="si"
          class="text-xs text-muted-foreground flex flex-wrap gap-2"
        >
          <span class="font-medium text-foreground">Set {{ si + 1 }}</span>
          <span v-if="set.reps != null">{{ set.reps }} reps</span>
          <span v-if="set.weight != null">@ {{ set.weight }} kg</span>
          <span v-if="set.duration != null">{{ set.duration }}s</span>
          <span v-if="set.distance != null">{{ set.distance }}m</span>
          <span v-if="set.rest_seconds != null" class="text-muted-foreground/60">{{ set.rest_seconds }}s rest</span>
        </div>
      </div>

      <!-- Exercise note -->
      <p
        v-if="ex.note"
        :class="compact ? 'ml-2 text-xs text-muted-foreground italic' : 'ml-7 text-xs text-muted-foreground italic'"
      >{{ ex.note }}</p>
    </div>
  </div>
</template>
