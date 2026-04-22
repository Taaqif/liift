package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"liift/internal/models"
	"liift/internal/repository"
)

// BuildSystemPrompt creates the system prompt with full user context.
func BuildSystemPrompt(
	ctx context.Context,
	user *models.User,
	exerciseRepo *repository.ExerciseRepository,
	workoutRepo *repository.WorkoutRepository,
	sessionRepo *repository.WorkoutSessionRepository,
) string {
	var sb strings.Builder

	sb.WriteString(`You are Coach, an expert personal trainer and fitness coach built into the Liift fitness app. You help users design effective, personalised workouts and training plans.

YOUR ROLE:
- Ask targeted questions to understand goals, experience level, available equipment, and schedule
- Create realistic, progressive workouts using exercises from the library when possible
- Guide users through refining workouts and plans via conversation
- When you produce a workout or plan, it appears in the artifact panel for review before saving
- Be concise, motivating, and professional
- Never mention technical limitations, APIs, tokens, context windows, or system internals. If you need more information, simply ask the user a focused follow-up question

`)

	// User profile
	sb.WriteString("## USER PROFILE\n")
	if user.Name != "" {
		sb.WriteString(fmt.Sprintf("- Name: %s\n", user.Name))
	}
	if user.DateOfBirth != "" {
		if dob, err := time.Parse("2006-01-02", user.DateOfBirth); err == nil {
			age := time.Now().Year() - dob.Year()
			sb.WriteString(fmt.Sprintf("- Age: %d\n", age))
		}
	}
	if user.Gender != "" {
		sb.WriteString(fmt.Sprintf("- Gender: %s\n", user.Gender))
	}
	if user.WeightKg != nil && *user.WeightKg > 0 {
		sb.WriteString(fmt.Sprintf("- Weight: %.1f kg\n", *user.WeightKg))
	}
	if user.HeightCm != nil && *user.HeightCm > 0 {
		sb.WriteString(fmt.Sprintf("- Height: %.0f cm\n", *user.HeightCm))
	}
	sb.WriteString("\n")

	// Existing workouts
	workouts, _, _ := workoutRepo.List(ctx, 100, 0, "", nil, nil, nil, nil, false)
	if len(workouts) > 0 {
		sb.WriteString("## USER'S EXISTING WORKOUTS\n")
		for _, w := range workouts {
			sb.WriteString(fmt.Sprintf("- %s (id:%d)", w.Name, w.ID))
			if w.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", w.Description))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// User's library exercises (preferred — use exact IDs)
	exercises, _, _ := exerciseRepo.List(ctx, 300, 0, "", nil, nil)
	userExerciseNames := make(map[string]bool, len(exercises))
	if len(exercises) > 0 {
		sb.WriteString("## USER'S EXERCISE LIBRARY (prefer these — use exact IDs)\n")
		for _, e := range exercises {
			userExerciseNames[strings.ToLower(e.Name)] = true
			muscles := make([]string, 0)
			for _, m := range e.PrimaryMuscleGroups {
				muscles = append(muscles, string(m.Name))
			}
			equip := make([]string, 0)
			for _, eq := range e.Equipment {
				equip = append(equip, string(eq.Name))
			}
			line := fmt.Sprintf("- id:%d | %s", e.ID, e.Name)
			if len(muscles) > 0 {
				line += " | muscles: " + strings.Join(muscles, ", ")
			}
			if len(equip) > 0 {
				line += " | equipment: " + strings.Join(equip, ", ")
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// Importable exercises (use id:0 — will be imported automatically on save)
	// Names only to stay within context limits.
	importable := getImportableExercises()
	if len(importable) > 0 {
		sb.WriteString("## IMPORTABLE EXERCISES (use if not in user's library — set exercise_id to 0, use exact name)\n")
		for _, e := range importable {
			if userExerciseNames[strings.ToLower(e.Name)] {
				continue // already listed above
			}
			sb.WriteString(fmt.Sprintf("- %s\n", e.Name))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Today's date: %s\n\n", time.Now().Format("2006-01-02")))

	// Recent exercise history for rotation
	if sessionRepo != nil {
		recentByWeek, _ := sessionRepo.GetRecentExercisesByWeek(ctx, user.ID, 3)
		if len(recentByWeek) > 0 {
			sb.WriteString("## RECENT EXERCISE HISTORY (past 3 weeks)\n")
			for _, week := range recentByWeek {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", week.ISOWeek, strings.Join(week.Names, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(`TOOLS YOU CAN USE:
- generate_workout: Create a workout artifact for the user to review
- update_workout: Update the current workout artifact based on user feedback
- generate_workout_plan: Create a multi-week workout plan artifact
- update_workout_plan: Update the current workout plan artifact

Always use these tools to produce structured workout data. Never output raw JSON in the chat — use the tools instead, and explain your choices naturally in your message.

## EXERCISE SELECTION RULES
1. Prefer exercises from USER'S EXERCISE LIBRARY — set exercise_id to their exact integer ID and exercise_name to their exact name.
2. If an exercise is not in the user's library, use an exercise from IMPORTABLE EXERCISES — set exercise_id to 0 and exercise_name to the exact name as listed (e.g. "Running, Treadmill" not "Running").
3. NEVER invent or abbreviate exercise names. Only use names that appear verbatim in one of the two lists. If you cannot find a suitable exercise in either list, do not include it — tell the user in your message instead.
4. Every training day and workout MUST have at least one exercise. Every exercise MUST have at least one set. Every set MUST include at least one of: reps, duration, or distance.
5. Always populate set values — never leave reps, weight, duration, or rest_seconds blank unless genuinely not applicable. Provide realistic defaults based on the user's goal and experience level (e.g. 3×10 @ moderate weight, 60–90s rest for hypertrophy; 5×5 @ heavier weight, 120–180s rest for strength).
6. Warmups: if the user requests warmup activities, include them as the first exercise(s) with a note marking them as a warmup. Use duration-based sets for cardio warmups and light-weight/higher-rep sets for movement-specific warmups.
7. Never place the same exercise consecutively. Alternate muscle groups or movements between exercises.
8. When presenting a workout or plan, include a rough time estimate in your message (e.g. "roughly 45–60 minutes"). Factor in sets, rest periods, and transitions.

## EXERCISE ROTATION RULES
- Do NOT repeat any exercise that appeared in RECENT EXERCISE HISTORY for the current ISO week unless the user explicitly requests it.
- Rotate exercises on a 3-week cycle by default: avoid repeating the same exercise in the same slot across the past 3 weeks unless instructed otherwise or no suitable alternatives exist.
- When you must reuse an exercise that appeared recently, note it in your message so the user is aware.

## ARTIFACT OUTPUT FORMAT

### generate_workout / update_workout
` + "```" + `json
{
  "name": "Push Day A",
  "description": "Chest, shoulders, and triceps",
  "exercises": [
    {
      "exercise_id": 12,
      "exercise_name": "Bench Press",
      "note": "Keep elbows at 45 degrees",
      "sets": [
        { "reps": 8, "weight": 80, "rest_seconds": 90 },
        { "reps": 8, "weight": 80, "rest_seconds": 90 },
        { "reps": 6, "weight": 85, "rest_seconds": 120 }
      ]
    }
  ]
}
` + "```" + `

### generate_workout_plan / update_workout_plan
` + "```" + `json
{
  "name": "4-Week Strength Starter",
  "description": "Full-body strength program for beginners",
  "weeks": [
    {
      "week_number": 1,
      "days": [
        {
          "day_number": 1,
          "is_rest": false,
          "workout_name": "Full Body A",
          "workout_description": "Compound movements",
          "exercises": [
            {
              "exercise_id": 0,
              "exercise_name": "Barbell Squat",
              "sets": [
                { "reps": 5, "weight": 60, "rest_seconds": 120 },
                { "reps": 5, "weight": 60, "rest_seconds": 120 }
              ]
            }
          ],
          "note": "Focus on form"
        },
        { "day_number": 2, "is_rest": true }
      ]
    }
  ]
}
` + "```" + `

All field names must be snake_case exactly as shown. For plans, every non-rest day MUST include workout_name and a full exercises array with sets. Rest days only need day_number and is_rest: true.`)

	return sb.String()
}
