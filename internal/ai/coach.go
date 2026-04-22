package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"liift/internal/models"

	"github.com/tmc/langchaingo/llms"
)

// SSEEvent types sent to the frontend.
const (
	EventTextDelta = "text_delta"
	EventArtifact  = "artifact"
	EventDone      = "done"
	EventError     = "error"
)

// ArtifactType identifies what kind of artifact is being sent.
const (
	ArtifactWorkout     = "workout"
	ArtifactWorkoutPlan = "workout_plan"
)

// SSEWriter writes Server-Sent Events to a response writer.
type SSEWriter struct {
	w io.Writer
	f interface{ Flush() }
}

func NewSSEWriter(w io.Writer, f interface{ Flush() }) *SSEWriter {
	return &SSEWriter{w: w, f: f}
}

func (s *SSEWriter) Write(eventType string, data any) {
	b, _ := json.Marshal(map[string]any{
		"type": eventType,
		"data": data,
	})
	fmt.Fprintf(s.w, "data: %s\n\n", b)
	if s.f != nil {
		s.f.Flush()
	}
}

func (s *SSEWriter) WriteTextDelta(delta string) {
	s.Write(EventTextDelta, map[string]string{"delta": delta})
}

func (s *SSEWriter) WriteArtifact(artifactType string, data any) {
	s.Write(EventArtifact, map[string]any{
		"artifactType": artifactType,
		"artifact":     data,
	})
}

func (s *SSEWriter) WriteDone(messageID uint) {
	s.Write(EventDone, map[string]any{"messageId": messageID})
}

func (s *SSEWriter) WriteError(msg string) {
	s.Write(EventError, map[string]string{"error": msg})
}

// HistoryMessage is the incoming message format from the client.
type HistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RunChat executes the agentic chat loop and streams results via SSEWriter.
func RunChat(
	ctx context.Context,
	llm llms.Model,
	systemPrompt string,
	history []HistoryMessage,
	sse *SSEWriter,
) (assistantText string, artifactType string, artifactData any, err error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
	}
	for _, h := range history {
		role := llms.ChatMessageTypeHuman
		if h.Role == "assistant" {
			role = llms.ChatMessageTypeAI
		}
		messages = append(messages, llms.TextParts(role, h.Content))
	}

	tools := CoachTools()

	const maxIterations = 5
	for i := 0; i < maxIterations; i++ {
		resp, callErr := llm.GenerateContent(ctx, messages, llms.WithTools(tools))
		if callErr != nil {
			sse.WriteError("AI generation failed: " + callErr.Error())
			return "", "", nil, callErr
		}

		if len(resp.Choices) == 0 {
			break
		}

		choice := resp.Choices[0]

		if len(choice.ToolCalls) > 0 {
			assistantParts := []llms.ContentPart{}
			if choice.Content != "" {
				assistantParts = append(assistantParts, llms.TextContent{Text: choice.Content})
			}
			for _, tc := range choice.ToolCalls {
				assistantParts = append(assistantParts, tc)
			}
			messages = append(messages, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: assistantParts,
			})

			for _, tc := range choice.ToolCalls {
				result, toolArtifactType, toolArtifact, execErr := executeCoachTool(tc)
				if execErr != nil {
					result = fmt.Sprintf("Tool error: %s", execErr.Error())
				}

				if toolArtifact != nil {
					artifactType = toolArtifactType
					artifactData = toolArtifact
					sse.WriteArtifact(toolArtifactType, toolArtifact)
				}

				messages = append(messages, llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: tc.ID,
							Name:       tc.FunctionCall.Name,
							Content:    result,
						},
					},
				})
			}
			continue
		}

		assistantText = choice.Content
		if assistantText != "" {
			streamText(sse, assistantText)
		}
		break
	}

	return assistantText, artifactType, artifactData, nil
}

// streamText sends text content in word chunks for a streaming feel.
func streamText(sse *SSEWriter, text string) {
	words := strings.Fields(text)
	for i, word := range words {
		chunk := word
		if i < len(words)-1 {
			chunk += " "
		}
		sse.WriteTextDelta(chunk)
	}
}

// executeCoachTool runs a tool call and returns a result string and optional artifact.
func executeCoachTool(tc llms.ToolCall) (result string, artifactType string, artifactData any, err error) {
	if tc.FunctionCall == nil {
		return "No function call.", "", nil, nil
	}
	args := tc.FunctionCall.Arguments

	switch tc.FunctionCall.Name {
	case "generate_workout", "update_workout":
		var workout WorkoutArtifact
		if parseErr := json.Unmarshal([]byte(args), &workout); parseErr != nil {
			return "", "", nil, fmt.Errorf("invalid workout data: %w", parseErr)
		}
		if validErr := validateWorkoutArtifact(&workout); validErr != nil {
			return "", "", nil, fmt.Errorf("workout artifact invalid: %w", validErr)
		}
		return "Workout artifact generated and displayed to user for review.", ArtifactWorkout, workout, nil

	case "generate_workout_plan", "update_workout_plan":
		var plan WorkoutPlanArtifact
		if parseErr := json.Unmarshal([]byte(args), &plan); parseErr != nil {
			return "", "", nil, fmt.Errorf("invalid plan data: %w", parseErr)
		}
		if validErr := validatePlanArtifact(&plan); validErr != nil {
			return "", "", nil, fmt.Errorf("plan artifact invalid: %w", validErr)
		}
		return "Workout plan artifact generated and displayed to user for review.", ArtifactWorkoutPlan, plan, nil

	default:
		return "Unknown tool.", "", nil, nil
	}
}

// validateWorkoutArtifact ensures the parsed workout has required, non-empty fields.
func validateWorkoutArtifact(w *WorkoutArtifact) error {
	if w.Name == "" {
		return fmt.Errorf("workout name is required")
	}
	if len(w.Exercises) == 0 {
		return fmt.Errorf("workout must have at least one exercise")
	}
	for i, ex := range w.Exercises {
		if ex.ExerciseName == "" {
			return fmt.Errorf("exercise %d: exercise_name is required", i+1)
		}
		if len(ex.Sets) == 0 {
			return fmt.Errorf("exercise %d (%s): must have at least one set", i+1, ex.ExerciseName)
		}
	}
	return nil
}

// validatePlanArtifact ensures the parsed plan has required, non-empty fields.
func validatePlanArtifact(p *WorkoutPlanArtifact) error {
	if p.Name == "" {
		return fmt.Errorf("plan name is required")
	}
	if len(p.Weeks) == 0 {
		return fmt.Errorf("plan must have at least one week")
	}
	for wi, week := range p.Weeks {
		if len(week.Days) == 0 {
			return fmt.Errorf("week %d: must have at least one day", wi+1)
		}
		for di, day := range week.Days {
			if !day.IsRest {
				if day.WorkoutName == "" {
					return fmt.Errorf("week %d day %d: workout_name required for training days", wi+1, day.DayNumber)
				}
				if len(day.Exercises) == 0 {
					return fmt.Errorf("week %d day %d (%s): exercises required for training days", wi+1, day.DayNumber, day.WorkoutName)
				}
				for ei, ex := range day.Exercises {
					if ex.ExerciseName == "" {
						return fmt.Errorf("week %d day %d exercise %d: exercise_name is required", wi+1, di+1, ei+1)
					}
					if len(ex.Sets) == 0 {
						return fmt.Errorf("week %d day %d exercise %d (%s): must have at least one set", wi+1, di+1, ei+1, ex.ExerciseName)
					}
				}
			}
		}
	}
	return nil
}

// MessagesToHistory converts stored DB messages to HistoryMessage slice.
func MessagesToHistory(msgs []models.ChatMessage) []HistoryMessage {
	out := make([]HistoryMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, HistoryMessage{Role: m.Role, Content: m.Content})
	}
	return out
}
