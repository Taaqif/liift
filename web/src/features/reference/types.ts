export type Equipment = { name: string };
export type EquipmentResponse = Equipment[];

export type MuscleGroup = { name: string };
export type MuscleGroupResponse = MuscleGroup[];

export type ExerciseFeature = { name: string };
export type ExerciseFeatureResponse = ExerciseFeature[];

export type ReferenceType = "exerciseFeature" | "muscleGroup" | "equipment";

export type ReferenceOption = { value: string; label: string };
