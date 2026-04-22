export interface ChatSession {
  id: number;
  slug: string;
  title: string;
  created_at: string;
  updated_at: string;
}

export interface ChatMessage {
  id: number;
  session_id: number;
  role: "user" | "assistant";
  content: string;
  metadata?: string;
  created_at: string;
}

export interface MessageArtifact {
  artifactType: ArtifactType;
  artifact: WorkoutArtifact | WorkoutPlanArtifact;
}

export interface ChatSessionDetail extends ChatSession {
  messages: ChatMessage[];
}

// Artifact types matching the Go structs

export interface WorkoutSetArtifact {
  reps?: number;
  weight?: number;
  duration?: number;
  distance?: number;
  rest_seconds?: number;
}

export interface WorkoutExerciseArtifact {
  exercise_id: number;
  exercise_name: string;
  sets: WorkoutSetArtifact[];
  note?: string;
}

export interface WorkoutArtifact {
  name: string;
  description?: string;
  exercises: WorkoutExerciseArtifact[];
}

export interface PlanDayArtifact {
  day_number: number;
  is_rest: boolean;
  workout_name?: string;
  workout_description?: string;
  exercises?: WorkoutExerciseArtifact[];
  note?: string;
}

export interface PlanWeekArtifact {
  week_number: number;
  days: PlanDayArtifact[];
}

export interface WorkoutPlanArtifact {
  name: string;
  description?: string;
  weeks: PlanWeekArtifact[];
}

export type ArtifactType = "workout" | "workout_plan";

export interface ActiveArtifact {
  type: ArtifactType;
  data: WorkoutArtifact | WorkoutPlanArtifact;
}

// SSE event shapes
export interface SSETextDelta {
  type: "text_delta";
  data: { delta: string };
}

export interface SSEArtifact {
  type: "artifact";
  data: { artifactType: ArtifactType; artifact: WorkoutArtifact | WorkoutPlanArtifact };
}

export interface SSEDone {
  type: "done";
  data: { messageId: number };
}

export interface SSEError {
  type: "error";
  data: { error: string };
}

export type SSEEvent = SSETextDelta | SSEArtifact | SSEDone | SSEError;
