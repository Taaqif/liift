package handlers

import (
	"net/http"

	"liift/api/middleware"
	"liift/api/types"
	"liift/internal/ai"
	"liift/internal/models"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
)

type AISettingsHandler struct {
	repo *repository.AISettingsRepository
}

func NewAISettingsHandler(repo *repository.AISettingsRepository) *AISettingsHandler {
	return &AISettingsHandler{repo: repo}
}

type AISettingsResponse struct {
	Provider      string `json:"provider"`
	APIKeyMasked  string `json:"apiKeyMasked"`
	HasAPIKey     bool   `json:"hasApiKey"`
	Model         string `json:"model"`
	OllamaBaseURL string `json:"ollamaBaseURL"`
	CustomBaseURL string `json:"customBaseURL"`
	IsConfigured  bool   `json:"isConfigured"`
}

type UpdateAISettingsRequest struct {
	Provider      string `json:"provider"`
	APIKey        string `json:"apiKey"`
	AIModel       string `json:"model"`
	OllamaBaseURL string `json:"ollamaBaseURL"`
	CustomBaseURL string `json:"customBaseURL"`
}

func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

func mapSettingsToResponse(s *models.AISettings, isConfigured bool) AISettingsResponse {
	return AISettingsResponse{
		Provider:      s.Provider,
		APIKeyMasked:  maskAPIKey(s.APIKey),
		HasAPIKey:     s.APIKey != "",
		Model:         s.AIModel,
		OllamaBaseURL: s.OllamaBaseURL,
		CustomBaseURL: s.CustomBaseURL,
		IsConfigured:  isConfigured,
	}
}

func (h *AISettingsHandler) GetSettings(c echo.Context) error {
	userID := middleware.GetUserID(c)
	s, isConfigured, err := h.repo.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_settings"})
	}
	return c.JSON(http.StatusOK, mapSettingsToResponse(s, isConfigured))
}

func (h *AISettingsHandler) UpdateSettings(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req UpdateAISettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_request_body"})
	}

	validProviders := map[string]bool{"openai": true, "anthropic": true, "google": true, "ollama": true, "custom": true}
	if req.Provider != "" && !validProviders[req.Provider] {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_provider"})
	}

	existing, _, err := h.repo.GetByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_settings"})
	}

	s := &models.AISettings{
		UserID:        userID,
		Provider:      req.Provider,
		APIKey:        req.APIKey,
		AIModel:       req.AIModel,
		OllamaBaseURL: req.OllamaBaseURL,
		CustomBaseURL: req.CustomBaseURL,
	}
	if s.Provider == "" {
		s.Provider = existing.Provider
	}
	// If apiKey is empty in request, preserve existing key (upsert repo handles this)

	if err := h.repo.Upsert(c.Request().Context(), s); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_save_settings"})
	}

	updated, _, _ := h.repo.GetByUserID(c.Request().Context(), userID)
	return c.JSON(http.StatusOK, mapSettingsToResponse(updated, true))
}

// GetProviders returns the list of supported providers and their default models.
func (h *AISettingsHandler) GetProviders(c echo.Context) error {
	type ProviderInfo struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		DefaultModel string   `json:"defaultModel"`
		Models       []string `json:"models"`
		NeedsAPIKey  bool     `json:"needsApiKey"`
		NeedsBaseURL bool     `json:"needsBaseUrl"`
	}
	providers := []ProviderInfo{
		{
			ID:           "openai",
			Name:         "OpenAI",
			DefaultModel: ai.DefaultModels["openai"],
			Models:       []string{"gpt-4o-mini", "gpt-4o", "gpt-4-turbo", "gpt-3.5-turbo"},
			NeedsAPIKey:  true,
		},
		{
			ID:           "anthropic",
			Name:         "Anthropic",
			DefaultModel: ai.DefaultModels["anthropic"],
			Models:       []string{"claude-3-5-haiku-20241022", "claude-3-5-sonnet-20241022", "claude-opus-4-5"},
			NeedsAPIKey:  true,
		},
		{
			ID:           "google",
			Name:         "Google AI",
			DefaultModel: ai.DefaultModels["google"],
			Models:       []string{"gemini-1.5-flash", "gemini-1.5-pro", "gemini-2.0-flash"},
			NeedsAPIKey:  true,
		},
		{
			ID:           "ollama",
			Name:         "Ollama (Local)",
			DefaultModel: ai.DefaultModels["ollama"],
			Models:       []string{"llama3", "llama3.1", "mistral", "phi3", "gemma2"},
			NeedsAPIKey:  false,
			NeedsBaseURL: true,
		},
		{
			ID:           "custom",
			Name:         "Custom (OpenAI-compatible)",
			DefaultModel: "",
			Models:       []string{},
			NeedsAPIKey:  true,
			NeedsBaseURL: true,
		},
	}
	return c.JSON(http.StatusOK, providers)
}

func RegisterAISettingsRoutes(api *echo.Group, handler *AISettingsHandler) {
	api.GET("/ai/settings", handler.GetSettings)
	api.PUT("/ai/settings", handler.UpdateSettings)
	api.GET("/ai/providers", handler.GetProviders)
}
