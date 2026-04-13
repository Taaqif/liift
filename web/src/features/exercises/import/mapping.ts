import type { FreeExercise, MappedImportExercise } from "./types";

const IMAGE_BASE_URL =
  "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/";

// free-exercise-db muscle name → our enum
const MUSCLE_MAP: Record<string, string> = {
  abdominals: "abdominals",
  abductors: "abductors",
  adductors: "adductors",
  biceps: "biceps",
  calves: "calves",
  chest: "chest",
  forearms: "forearms",
  glutes: "glutes",
  hamstrings: "hamstrings",
  lats: "lats",
  "lower back": "lower_back",
  "middle back": "upper_back",
  neck: "neck",
  quadriceps: "quadriceps",
  shoulders: "shoulders",
  traps: "traps",
  triceps: "triceps",
  "upper back": "upper_back",
};

// free-exercise-db equipment → our enum
const EQUIPMENT_MAP: Record<string, string> = {
  barbell: "barbell",
  cable: "machine",
  dumbbell: "dumbbell",
  "ez curl bar": "barbell",
  "exercise ball": "swiss_ball",
  "foam roll": "gym_mat",
  kettlebell: "kettlebell",
  machine: "machine",
  "medicine ball": "gym_mat",
  other: "bodyweight",
  bands: "resistance_band",
  "body only": "bodyweight",
  none: "bodyweight",
};

// free-exercise-db category → our enum
const CATEGORY_MAP: Record<string, string> = {
  strength: "strength",
  powerlifting: "strength",
  "olympic weightlifting": "strength",
  strongman: "strength",
  plyometrics: "strength",
  cardio: "cardio",
  stretching: "stretching",
};

function mapMuscle(muscle: string): string {
  return MUSCLE_MAP[muscle.toLowerCase()] ?? "other";
}

function mapEquipment(equipment: string | null): string {
  if (!equipment) return "bodyweight";
  return EQUIPMENT_MAP[equipment.toLowerCase()] ?? "bodyweight";
}

function mapCategory(category: string): string | undefined {
  return CATEGORY_MAP[category.toLowerCase()];
}

function mapForce(force: string | null): string | undefined {
  if (!force) return undefined;
  const valid = ["push", "pull", "static"];
  return valid.includes(force.toLowerCase()) ? force.toLowerCase() : undefined;
}

function inferFeatures(category: string): string[] {
  const cat = category.toLowerCase();
  if (cat === "cardio") return ["duration", "distance"];
  if (cat === "stretching") return ["duration"];
  // strength, powerlifting, olympic, strongman, plyometrics
  return ["weight", "rep"];
}

export function mapExercise(source: FreeExercise): MappedImportExercise {
  const primaryMuscles = [
    ...new Set(source.primaryMuscles.map(mapMuscle)),
  ].filter(Boolean);
  const secondaryMuscles = [
    ...new Set(source.secondaryMuscles.map(mapMuscle)),
  ].filter((m) => !primaryMuscles.includes(m));

  const equipment = mapEquipment(source.equipment);
  const category = mapCategory(source.category);
  const force = mapForce(source.force);
  const features = inferFeatures(source.category);

  const imageUrl =
    source.images.length > 0 ? `${IMAGE_BASE_URL}${source.images[0]}` : null;

  return {
    source,
    name: source.name,
    description: "",
    force,
    category,
    instructions: source.instructions,
    primary_muscle_groups:
      primaryMuscles.length > 0 ? primaryMuscles : ["other"],
    secondary_muscle_groups: secondaryMuscles,
    equipment: [equipment],
    exercise_features: features,
    imageUrl,
  };
}
