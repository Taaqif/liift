import { ref } from "vue";
import type { ResolvedExerciseRequest } from "@/features/ai-chat/composables/useExerciseResolver";

export type AIWorkoutState = {
  name: string;
  description?: string;
  exercises: ResolvedExerciseRequest[];
};

export type AIResolvedDay = {
  day_number: number;
  is_rest: boolean;
  workout_name?: string;
  workout_description?: string;
  exercises?: ResolvedExerciseRequest[];
  note?: string;
};

export type AIPlanState = {
  name: string;
  description?: string;
  weeks: {
    week_number: number;
    days: AIResolvedDay[];
  }[];
};

// Module-level singletons — survive across component lifecycle but not page refresh
const _aiWorkout = ref<AIWorkoutState | null>(null);
const _aiPlan = ref<AIPlanState | null>(null);

export function useAIFormState() {
  function setAIWorkout(data: AIWorkoutState) {
    _aiWorkout.value = data;
  }

  /** Read and clear the pending AI workout state. */
  function takeAIWorkout(): AIWorkoutState | null {
    const v = _aiWorkout.value;
    _aiWorkout.value = null;
    return v;
  }

  function setAIPlan(data: AIPlanState) {
    _aiPlan.value = data;
  }

  /** Read and clear the pending AI plan state. */
  function takeAIPlan(): AIPlanState | null {
    const v = _aiPlan.value;
    _aiPlan.value = null;
    return v;
  }

  return { setAIWorkout, takeAIWorkout, setAIPlan, takeAIPlan };
}
