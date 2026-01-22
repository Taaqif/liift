// Exercise types matching the API response
export type MuscleGroup = {
  name: string;
};

export type Equipment = {
  name: string;
};

export type ExerciseFeature = {
  name: string;
};

export type Exercise = {
  id: number;
  name: string;
  description?: string;
  image?: string;
  primary_muscle_groups: MuscleGroup[];
  secondary_muscle_groups: MuscleGroup[];
  equipment: Equipment[];
  exercise_features: ExerciseFeature[];
  created_at?: string;
  updated_at?: string;
};

export type ExercisesListParams = {
  limit?: number;
  offset?: number;
  search?: string;
  muscleGroup?: string[];
  equipment?: string[];
};

export type ExercisesListResponse = {
  data: Exercise[];
  total: number;
  limit: number;
  offset: number;
};

