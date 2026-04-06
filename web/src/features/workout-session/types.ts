export type WorkoutSessionSetValue = {
  id: number;
  feature_name: string;
  value: number;
};

export type WorkoutSessionSet = {
  id: number;
  workout_set_id?: number | null;
  order: number;
  completed_at: string | null;
  values: WorkoutSessionSetValue[];
};

export type WorkoutSessionExerciseRef = {
  id: number;
  name: string;
  description?: string;
  image?: string;
  primary_muscle_groups: { name: string }[];
  secondary_muscle_groups: { name: string }[];
  equipment: { name: string }[];
  exercise_features: { name: string }[];
};

export type WorkoutSessionExercise = {
  id: number;
  workout_exercise_id: number;
  order: number;
  note: string;
  rest_timer: number;
  exercise?: WorkoutSessionExerciseRef | null;
  sets: WorkoutSessionSet[];
};

export type WorkoutSessionWorkoutRef = {
  id: number;
  name: string;
  description?: string;
};

export type WorkoutSession = {
  id: number;
  user_id: number;
  workout_id: number;
  plan_progress_id?: number | null;
  started_at: string;
  ended_at: string | null;
  workout?: WorkoutSessionWorkoutRef | null;
  exercises: WorkoutSessionExercise[];
};

export type UpdateWorkoutSessionSetValue = {
  id?: number;
  feature_name: string;
  value: number;
};

export type UpdateWorkoutSessionSet = {
  id?: number;
  workout_set_id?: number | null;
  order: number;
  completed_at: string | null;
  values: UpdateWorkoutSessionSetValue[];
};

export type UpdateWorkoutSessionExercise = {
  id?: number;
  workout_exercise_id: number;
  order: number;
  note: string;
  rest_timer: number;
  sets: UpdateWorkoutSessionSet[];
};

export type UpdateWorkoutSessionPayload = {
  exercises: UpdateWorkoutSessionExercise[];
};
