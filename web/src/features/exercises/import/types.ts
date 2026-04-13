// Types matching the free-exercise-db schema
export type FreeExercise = {
  id: string;
  name: string;
  force: string | null;
  level: "beginner" | "intermediate" | "expert";
  mechanic: string | null;
  equipment: string | null;
  primaryMuscles: string[];
  secondaryMuscles: string[];
  instructions: string[];
  category: string;
  images: string[];
};

// A FreeExercise mapped to our CreateExerciseRequest shape, ready to POST
export type MappedImportExercise = {
  // Source data for display
  source: FreeExercise;
  // Mapped fields
  name: string;
  description: string;
  force?: string;
  category?: string;
  instructions: string[];
  primary_muscle_groups: string[];
  secondary_muscle_groups: string[];
  equipment: string[];
  exercise_features: string[];
  // First image URL from GitHub raw CDN (used to fetch and upload)
  imageUrl: string | null;
};
