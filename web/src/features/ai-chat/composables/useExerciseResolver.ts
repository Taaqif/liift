import { ref, computed } from "vue";
import { apiClient } from "@/lib/api";
import { useExercises } from "@/features/exercises/composables/useExercises";
import { mapExercise } from "@/features/exercises/import/mapping";
import type { FreeExercise } from "@/features/exercises/import/types";
import type { WorkoutExerciseArtifact, WorkoutSetArtifact } from "@/features/ai-chat/types";
import type { Exercise } from "@/features/exercises/types";

export type ResolvedExerciseRequest = {
  exercise_id: number;
  rest_timer: number;
  note?: string;
  order: number;
  sets: {
    order: number;
    features: { feature_name: string; value: number }[];
  }[];
};

export function useExerciseResolver() {
  const { exercises } = useExercises(() => ({ limit: 999 }));
  const freeExerciseDb = ref<FreeExercise[]>([]);
  const freeDbLoaded = ref(false);

  const userExerciseIds = computed(() => new Set(exercises.value.map((e) => e.id)));
  const userExerciseNames = computed(() => new Set(exercises.value.map((e) => e.name.toLowerCase())));

  async function ensureFreeDb() {
    if (freeDbLoaded.value) return;
    const res = await fetch("/exercises.json");
    if (res.ok) freeExerciseDb.value = await res.json();
    freeDbLoaded.value = true;
  }

  function findBestImportMatch(name: string): FreeExercise | null {
    if (freeExerciseDb.value.length === 0) return null;
    const q = name.toLowerCase().trim();

    let m = freeExerciseDb.value.find((e) => e.name.toLowerCase() === q);
    if (m) return m;

    m = freeExerciseDb.value.find((e) => {
      const el = e.name.toLowerCase();
      return el.startsWith(q + ",") || el.startsWith(q + " (");
    });
    if (m) return m;

    const words = q.split(/[\s,]+/).filter(Boolean);
    if (words.length > 0) {
      m = freeExerciseDb.value.find((e) => {
        const el = e.name.toLowerCase();
        return words.every((w) => el.includes(w));
      });
      if (m) return m;
    }

    return null;
  }

  async function resolveExerciseId(exerciseId: number, exerciseName: string): Promise<number> {
    const nameLower = exerciseName.toLowerCase();

    if (exerciseId > 0 && userExerciseIds.value.has(exerciseId)) return exerciseId;

    const existing = exercises.value.find((e) => e.name.toLowerCase() === nameLower);
    if (existing) return existing.id;

    await ensureFreeDb();
    const match =
      freeExerciseDb.value.find((e) => e.name.toLowerCase() === nameLower) ??
      findBestImportMatch(exerciseName);

    if (!match) {
      throw new Error(
        `"${exerciseName}" couldn't be found in the exercise library. Ask the coach to suggest an alternative.`,
      );
    }

    const mapped = mapExercise(match);
    const created = await apiClient.post<Exercise>("/exercises", {
      name: mapped.name,
      description: mapped.description,
      force: mapped.force,
      category: mapped.category,
      instructions: mapped.instructions,
      primary_muscle_groups: mapped.primary_muscle_groups,
      secondary_muscle_groups: mapped.secondary_muscle_groups,
      equipment: mapped.equipment,
      exercise_features: mapped.exercise_features,
    });
    return created.id;
  }

  function buildFeatures(set: WorkoutSetArtifact): { feature_name: string; value: number }[] {
    const features: { feature_name: string; value: number }[] = [];
    if (set.reps != null) features.push({ feature_name: "rep", value: set.reps });
    if (set.weight != null) features.push({ feature_name: "weight", value: set.weight });
    if (set.duration != null) features.push({ feature_name: "duration", value: set.duration });
    if (set.distance != null) features.push({ feature_name: "distance", value: set.distance });
    return features;
  }

  async function buildExerciseRequests(exs: WorkoutExerciseArtifact[]): Promise<ResolvedExerciseRequest[]> {
    await ensureFreeDb();
    return Promise.all(
      exs.map(async (ex, i) => ({
        exercise_id: await resolveExerciseId(ex.exercise_id, ex.exercise_name),
        rest_timer: ex.sets[0]?.rest_seconds ?? 60,
        note: ex.note,
        order: i + 1,
        sets: ex.sets.map((set, si) => ({
          order: si + 1,
          features: buildFeatures(set),
        })),
      })),
    );
  }

  return {
    exercises,
    resolveExerciseId,
    buildExerciseRequests,
    ensureFreeDb,
    freeExerciseDb,
    findBestImportMatch,
    buildFeatures,
  };
}
