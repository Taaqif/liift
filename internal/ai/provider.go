// Package ai provides the AI coaching layer using LangChain Go.
package ai

import (
	"context"
	"fmt"

	"liift/internal/models"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// DefaultModels maps each provider to a sensible default model.
var DefaultModels = map[string]string{
	"openai":    "gpt-4o-mini",
	"anthropic": "claude-3-5-haiku-20241022",
	"google":    "gemini-1.5-flash",
	"ollama":    "llama3",
}

// NewLLM creates an LLM instance from the user's AI settings.
func NewLLM(ctx context.Context, s *models.AISettings) (llms.Model, error) {
	model := s.AIModel
	if model == "" {
		model = DefaultModels[s.Provider]
	}

	switch s.Provider {
	case "openai":
		if s.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key not configured")
		}
		return openai.New(
			openai.WithToken(s.APIKey),
			openai.WithModel(model),
		)
	case "anthropic":
		if s.APIKey == "" {
			return nil, fmt.Errorf("Anthropic API key not configured")
		}
		return anthropic.New(
			anthropic.WithToken(s.APIKey),
			anthropic.WithModel(model),
		)
	case "google":
		if s.APIKey == "" {
			return nil, fmt.Errorf("Google AI API key not configured")
		}
		return googleai.New(ctx,
			googleai.WithAPIKey(s.APIKey),
			googleai.WithDefaultModel(model),
		)
	case "ollama":
		baseURL := s.OllamaBaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return ollama.New(
			ollama.WithModel(model),
			ollama.WithServerURL(baseURL),
		)
	case "custom":
		if s.APIKey == "" {
			return nil, fmt.Errorf("API key not configured")
		}
		if s.CustomBaseURL == "" {
			return nil, fmt.Errorf("custom base URL not configured")
		}
		return openai.New(
			openai.WithToken(s.APIKey),
			openai.WithModel(model),
			openai.WithBaseURL(s.CustomBaseURL),
		)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", s.Provider)
	}
}
