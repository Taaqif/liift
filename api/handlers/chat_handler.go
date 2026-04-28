package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"liift/api/middleware"
	"liift/api/types"
	"liift/internal/ai"
	"liift/internal/models"
	"liift/internal/repository"

	"github.com/labstack/echo/v4"
)

func generateSlug() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type ChatHandler struct {
	chatRepo    *repository.ChatRepository
	settingsRepo *repository.AISettingsRepository
	exerciseRepo *repository.ExerciseRepository
	workoutRepo  *repository.WorkoutRepository
	sessionRepo  *repository.WorkoutSessionRepository
}

func NewChatHandler(
	chatRepo *repository.ChatRepository,
	settingsRepo *repository.AISettingsRepository,
	exerciseRepo *repository.ExerciseRepository,
	workoutRepo *repository.WorkoutRepository,
	sessionRepo *repository.WorkoutSessionRepository,
) *ChatHandler {
	return &ChatHandler{
		chatRepo:     chatRepo,
		settingsRepo: settingsRepo,
		exerciseRepo: exerciseRepo,
		workoutRepo:  workoutRepo,
		sessionRepo:  sessionRepo,
	}
}

// Response types

type ChatMessageResponse struct {
	ID        uint      `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Metadata  string    `json:"metadata,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatSessionResponse struct {
	ID        uint                  `json:"id"`
	Slug      string                `json:"slug"`
	Title     string                `json:"title"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Messages  []ChatMessageResponse `json:"messages"`
}

type ChatSessionsListResponse struct {
	Data   []ChatSessionResponse `json:"data"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func mapMessage(m models.ChatMessage) ChatMessageResponse {
	return ChatMessageResponse{
		ID:        m.ID,
		Role:      m.Role,
		Content:   m.Content,
		Metadata:  m.Metadata,
		CreatedAt: m.CreatedAt,
	}
}

func mapSession(s models.ChatSession, includeMsgs bool) ChatSessionResponse {
	r := ChatSessionResponse{
		ID:        s.ID,
		Slug:      s.Slug,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		Messages:  []ChatMessageResponse{},
	}
	if includeMsgs {
		for _, m := range s.Messages {
			r.Messages = append(r.Messages, mapMessage(m))
		}
	}
	return r
}

// Handlers

func (h *ChatHandler) ListSessions(c echo.Context) error {
	userID := middleware.GetUserID(c)
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit <= 0 {
		limit = 20
	}

	sessions, total, err := h.chatRepo.ListSessions(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_fetch_sessions"})
	}

	data := make([]ChatSessionResponse, 0, len(sessions))
	for _, s := range sessions {
		data = append(data, mapSession(s, false))
	}
	return c.JSON(http.StatusOK, ChatSessionsListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *ChatHandler) CreateSession(c echo.Context) error {
	userID := middleware.GetUserID(c)
	var req struct {
		Title string `json:"title"`
	}
	c.Bind(&req)
	if req.Title == "" {
		req.Title = "New Chat"
	}

	s := &models.ChatSession{
		UserID: userID,
		Slug:   generateSlug(),
		Title:  req.Title,
	}
	if err := h.chatRepo.CreateSession(c.Request().Context(), s); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_create_session"})
	}
	return c.JSON(http.StatusCreated, mapSession(*s, true))
}

func (h *ChatHandler) GetSession(c echo.Context) error {
	userID := middleware.GetUserID(c)
	slug := c.Param("slug")
	if slug == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_slug"})
	}

	s, err := h.chatRepo.GetSessionBySlug(c.Request().Context(), slug, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
	}
	return c.JSON(http.StatusOK, mapSession(*s, true))
}

func (h *ChatHandler) UpdateSession(c echo.Context) error {
	userID := middleware.GetUserID(c)
	slug := c.Param("slug")
	if slug == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_slug"})
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil || req.Title == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "title_required"})
	}

	s, err := h.chatRepo.GetSessionBySlug(c.Request().Context(), slug, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
	}
	if err := h.chatRepo.UpdateSessionTitle(c.Request().Context(), s.ID, userID, req.Title); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_update_session"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ChatHandler) DeleteSession(c echo.Context) error {
	userID := middleware.GetUserID(c)
	slug := c.Param("slug")
	if slug == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_slug"})
	}

	s, err := h.chatRepo.GetSessionBySlug(c.Request().Context(), slug, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
	}
	if err := h.chatRepo.DeleteSession(c.Request().Context(), s.ID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_delete_session"})
	}
	return c.NoContent(http.StatusNoContent)
}

// SendMessage handles the streaming chat endpoint.
// POST /chats/:slug/messages
// Body: { "content": "..." }
// Response: text/event-stream SSE
func (h *ChatHandler) SendMessage(c echo.Context) error {
	userID := middleware.GetUserID(c)
	slug := c.Param("slug")
	if slug == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid_slug"})
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.Bind(&req); err != nil || req.Content == "" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "content_required"})
	}

	ctx := c.Request().Context()

	// Verify session ownership
	session, err := h.chatRepo.GetSessionBySlug(ctx, slug, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "session_not_found"})
	}
	sessionID := session.ID

	// Get AI settings
	settings, _, err := h.settingsRepo.Get(ctx)
	if err != nil || settings.APIKey == "" && settings.Provider != "ollama" {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "ai_not_configured"})
	}

	// Create LLM
	llm, err := ai.NewLLM(ctx, settings)
	if err != nil {
		return c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: err.Error()})
	}

	// Get current user from context for system prompt
	user := middleware.GetUser(c)

	// Build system prompt with context
	systemPrompt := ai.BuildSystemPrompt(ctx, user, h.exerciseRepo, h.workoutRepo, h.sessionRepo)

	// Save user message
	userMsg := &models.ChatMessage{
		SessionID: uint(sessionID),
		Role:      "user",
		Content:   req.Content,
	}
	if err := h.chatRepo.CreateMessage(ctx, userMsg); err != nil {
		return c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "failed_to_save_message"})
	}

	// Build history from existing messages (excluding the just-saved user msg since we append it)
	history := ai.MessagesToHistory(session.Messages)
	history = append(history, ai.HistoryMessage{Role: "user", Content: req.Content})

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)

	flusher, _ := c.Response().Writer.(http.Flusher)
	sse := ai.NewSSEWriter(c.Response().Writer, flusher)

	// Run chat and stream response
	assistantText, artifactType, artifactData, runErr := ai.RunChat(ctx, llm, systemPrompt, history, sse)
	if runErr != nil {
		sse.WriteError(runErr.Error())
		return nil
	}

	// Save assistant message
	metadata := ""
	if artifactData != nil {
		metaObj := map[string]any{
			"artifactType": artifactType,
			"artifact":     artifactData,
		}
		if b, err := json.Marshal(metaObj); err == nil {
			metadata = string(b)
		}
	}

	assistantMsg := &models.ChatMessage{
		SessionID: uint(sessionID),
		Role:      "assistant",
		Content:   assistantText,
		Metadata:  metadata,
	}
	if err := h.chatRepo.CreateMessage(ctx, assistantMsg); err != nil {
		c.Logger().Errorf("failed to save assistant message: %v", err)
	}

	// Auto-title the session from the first user message
	if session.Title == "New Chat" && len(session.Messages) == 0 {
		title := req.Content
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		h.chatRepo.UpdateSessionTitle(ctx, uint(sessionID), userID, title)
	}

	h.chatRepo.TouchSession(ctx, uint(sessionID))
	sse.WriteDone(assistantMsg.ID)

	return nil
}

func RegisterChatRoutes(api *echo.Group, handler *ChatHandler) {
	api.GET("/chats", handler.ListSessions)
	api.POST("/chats", handler.CreateSession)
	api.GET("/chats/:slug", handler.GetSession)
	api.PUT("/chats/:slug", handler.UpdateSession)
	api.DELETE("/chats/:slug", handler.DeleteSession)
	api.POST("/chats/:slug/messages", handler.SendMessage)
}
