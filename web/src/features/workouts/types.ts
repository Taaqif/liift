import type { Exercise } from "@/features/exercises/types";

export type WorkoutSetFeature = {
  id?: number;
  feature_name: string;
  value: number;
};

export type WorkoutSet = {
  id?: number;
  order: number;
  features: WorkoutSetFeature[];
};

export type WorkoutExercise = {
  id?: number;
  workout_id?: number;
  exercise_id: number;
  rest_timer: number;
  note?: string;
  order: number;
  exercise?: Exercise;
  sets: WorkoutSet[];
};

export type Workout = {
  id: number;
  name: string;
  description?: string;
  exercises: WorkoutExercise[];
  created_at?: string;
  updated_at?: string;
};

export type WorkoutsListParams = {
  limit?: number;
  offset?: number;
  search?: string;
  exerciseFeature?: string;
  exerciseIds?: number[];
  muscleGroup?: string[];
  equipment?: string[];
};

export type WorkoutsListResponse = {
  data: Workout[];
  total: number;
  limit: number;
  offset: number;
};
