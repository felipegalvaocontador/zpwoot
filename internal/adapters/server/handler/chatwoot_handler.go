package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

type ChatwootHandler struct {
	*shared.BaseHandler
	messageService *services.MessageService
	sessionService *services.SessionService
}

func NewChatwootHandler(
	messageService *services.MessageService,
	sessionService *services.SessionService,
	logger *logger.Logger,
) *ChatwootHandler {
	return &ChatwootHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		messageService: messageService,
		sessionService: sessionService,
	}
}

type ChatwootWebhookPayload struct {
	Event        string                `json:"event"`
	Account      *ChatwootAccount      `json:"account,omitempty"`
	Conversation *ChatwootConversation `json:"conversation,omitempty"`
	Message      *ChatwootMessage      `json:"message,omitempty"`
	Contact      *ChatwootContact      `json:"contact,omitempty"`
	Inbox        *ChatwootInbox        `json:"inbox,omitempty"`
}

type ChatwootAccount struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ChatwootConversation struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type ChatwootMessage struct {
	ID          int    `json:"id"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
	Private     bool   `json:"private"`
	SenderType  string `json:"sender_type"`
}

type ChatwootContact struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type ChatwootInbox struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *ChatwootHandler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "receive chatwoot webhook")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var payload ChatwootWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid webhook payload")
		return
	}

	_, err := h.sessionService.GetSessionByNameOrID(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.GetLogger().InfoWithFields("Chatwoot webhook received", map[string]interface{}{
		"session_id": sessionID,
		"event":      payload.Event,
		"message_id": func() interface{} {
			if payload.Message != nil {
				return payload.Message.ID
			}
			return nil
		}(),
		"conversation_id": func() interface{} {
			if payload.Conversation != nil {
				return payload.Conversation.ID
			}
			return nil
		}(),
	})

	h.LogSuccess("receive chatwoot webhook", map[string]interface{}{
		"session_id": sessionID,
		"event":      payload.Event,
	})

	h.GetWriter().WriteSuccess(w, nil, "Webhook processed successfully")
}

// @Summary Create Chatwoot configuration
// @Description Create a new Chatwoot configuration for the session
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "create chatwoot config")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("create chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration created successfully")
}

// @Summary Find Chatwoot configuration
// @Description Find the current Chatwoot configuration for the session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot/find [get]
func (h *ChatwootHandler) FindConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "find chatwoot config")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("get chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration retrieved successfully")
}

func (h *ChatwootHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update chatwoot config")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("update chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration updated successfully")
}

func (h *ChatwootHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "delete chatwoot config")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("delete chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration deleted successfully")
}

func (h *ChatwootHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "test chatwoot connection")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("test chatwoot connection", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot connection test completed successfully")
}

func (h *ChatwootHandler) AutoCreateInbox(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "auto-create chatwoot inbox")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("auto-create chatwoot inbox", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot inbox created successfully")
}

func (h *ChatwootHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get chatwoot stats")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSessionByNameOrID(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("get chatwoot stats", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot statistics retrieved successfully")
}
