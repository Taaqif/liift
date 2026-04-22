package ai

import "github.com/tmc/langchaingo/llms"

// WorkoutSetInput represents a single set within an exercise.
type WorkoutSetInput struct {
	Reps        *float64 `json:"reps,omitempty"`
	Weight      *float64 `json:"weight,omitempty"`
	Duration    *float64 `json:"duration,omitempty"`
	Distance    *float64 `json:"distance,omitempty"`
	RestSeconds *int     `json:"rest_seconds,omitempty"`
}

// WorkoutExerciseInput represents an exercise within a workout artifact.
type WorkoutExerciseInput struct {
	ExerciseID   int               `json:"exercise_id"`
	ExerciseName string            `json:"exercise_name"`
	Note         string            `json:"note,omitempty"`
	Sets         []WorkoutSetInput `json:"sets"`
}

// WorkoutArtifact is the structured data for a generated workout.
type WorkoutArtifact struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Exercises   []WorkoutExerciseInput `json:"exercises"`
}

// PlanDayInput represents a single training day in a plan.
type PlanDayInput struct {
	DayNumber          int                    `json:"day_number"`
	IsRest             bool                   `json:"is_rest"`
	WorkoutName        string                 `json:"workout_name,omitempty"`
	WorkoutDescription string                 `json:"workout_description,omitempty"`
	Exercises          []WorkoutExerciseInput `json:"exercises,omitempty"`
	Note               string                 `json:"note,omitempty"`
}

// PlanWeekInput represents a week of training.
type PlanWeekInput struct {
	WeekNumber int            `json:"week_number"`
	Days       []PlanDayInput `json:"days"`
}

// WorkoutPlanArtifact is the structured data for a generated plan.
type WorkoutPlanArtifact struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Weeks       []PlanWeekInput `json:"weeks"`
}

var setSchema = map[string]any{
	"type": "object",
	"description": "A single set. Must include at least one of: reps, duration, or distance. " +
		"Include weight (kg) when applicable. Always include rest_seconds.",
	"properties": map[string]any{
		"reps":         map[string]any{"type": "number", "description": "Number of repetitions"},
		"weight":       map[string]any{"type": "number", "description": "Weight in kg"},
		"duration":     map[string]any{"type": "number", "description": "Duration in seconds"},
		"distance":     map[string]any{"type": "number", "description": "Distance in metres"},
		"rest_seconds": map[string]any{"type": "integer", "description": "Rest after this set in seconds (e.g. 60, 90, 120)"},
	},
}

var exerciseItemSchema = map[string]any{
	"type":     "object",
	"required": []string{"exercise_id", "exercise_name", "sets"},
	"properties": map[string]any{
		"exercise_id": map[string]any{
			"type": "integer",
			"description": "Exact integer ID from USER'S EXERCISE LIBRARY. " +
				"Use 0 if the exercise comes from IMPORTABLE EXERCISES.",
		},
		"exercise_name": map[string]any{
			"type":        "string",
			"description": "Exact name as listed — do not paraphrase or abbreviate.",
		},
		"note": map[string]any{"type": "string", "description": "Optional coaching cue"},
		"sets": map[string]any{
			"type":     "array",
			"minItems": 1,
			"items":    setSchema,
		},
	},
}

var workoutSchema = map[string]any{
	"type":     "object",
	"required": []string{"name", "exercises"},
	"properties": map[string]any{
		"name":        map[string]any{"type": "string"},
		"description": map[string]any{"type": "string"},
		"exercises": map[string]any{
			"type":     "array",
			"minItems": 1,
			"items":    exerciseItemSchema,
		},
	},
}

var planDaySchema = map[string]any{
	"type":        "object",
	"required":    []string{"day_number", "is_rest"},
	"description": "Rest days: is_rest=true only. Training days: is_rest=false with workout_name and exercises.",
	"properties": map[string]any{
		"day_number":          map[string]any{"type": "integer"},
		"is_rest":             map[string]any{"type": "boolean"},
		"workout_name":        map[string]any{"type": "string", "description": "Required when is_rest=false"},
		"workout_description": map[string]any{"type": "string"},
		"exercises": map[string]any{
			"type":     "array",
			"minItems": 1,
			"items":    exerciseItemSchema,
		},
		"note": map[string]any{"type": "string"},
	},
}

var planSchema = map[string]any{
	"type":     "object",
	"required": []string{"name", "weeks"},
	"properties": map[string]any{
		"name":        map[string]any{"type": "string"},
		"description": map[string]any{"type": "string"},
		"weeks": map[string]any{
			"type":     "array",
			"minItems": 1,
			"items": map[string]any{
				"type":     "object",
				"required": []string{"week_number", "days"},
				"properties": map[string]any{
					"week_number": map[string]any{"type": "integer"},
					"days": map[string]any{
						"type":     "array",
						"minItems": 1,
						"items":    planDaySchema,
					},
				},
			},
		},
	},
}

// CoachTools returns the list of tools available to the AI coach.
func CoachTools() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "generate_workout",
				Description: "Generate a workout artifact for the user to review. Call this when creating a new workout.",
				Parameters:  workoutSchema,
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "update_workout",
				Description: "Update the current workout artifact based on user feedback.",
				Parameters:  workoutSchema,
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "generate_workout_plan",
				Description: "Generate a multi-week workout plan artifact. Each training day lists its exercises directly.",
				Parameters:  planSchema,
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "update_workout_plan",
				Description: "Update the current workout plan artifact based on user feedback.",
				Parameters:  planSchema,
			},
		},
	}
}
