import type { Exercise } from "@/features/exercises/types";
import { z } from "zod";
import { i18n } from "@/i18n";

const t = i18n.global.t;

export const workoutSetFeatureSchema = z.object({
  id: z.number().optional(),
  feature_name: z.string(),
  value: z.number().min(0).nullable(),
});

export const workoutSetSchema = z.object({
  id: z.number().optional(),
  _key: z.string().optional(),
  order: z.number(),
  features: z.array(workoutSetFeatureSchema),
});

const setsSchema = z
  .array(workoutSetSchema)
  .min(1, t("workouts.validation.setsRequired"));

const exerciseIdSchema = z.union([
  z.number().min(1),
  z.null(),
]);

export const workoutExerciseSchema = z.object({
  id: z.number().optional(),
  exercise_id: z.preprocess((val) => (val === undefined ? null : val), exerciseIdSchema),
  rest_timer: z.number().min(0),
  note: z.string().optional(),
  order: z.number(),
  sets: z.preprocess((val) => val ?? [], setsSchema),
});

export type WorkoutSetFeatureForm = z.infer<typeof workoutSetFeatureSchema>;
export type WorkoutSetForm = z.infer<typeof workoutSetSchema>;
export type WorkoutExerciseForm = z.infer<typeof workoutExerciseSchema>;

export const workoutFormSchema = z.object({
  name: z.string().min(1, t("workouts.validation.nameRequired")),
  description: z.string().optional(),
  exercises: z
    .array(workoutExerciseSchema)
    .min(1, t("workouts.validation.exercisesRequired")),
});

export type WorkoutFormValues = z.infer<typeof workoutFormSchema>;

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
  is_library?: boolean;
  exercises: WorkoutExercise[];
  created_at?: string;
  updated_at?: string;
};

export type WorkoutsListParams = {
  limit?: number;
  offset?: number;
  search?: string;
  exerciseFeatures?: string[];
  exerciseIds?: number[];
  muscleGroup?: string[];
  equipment?: string[];
  includeAll?: boolean;
};

export type WorkoutsListResponse = {
  data: Workout[];
  total: number;
  limit: number;
  offset: number;
};
