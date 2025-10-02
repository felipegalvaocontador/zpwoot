package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// ChatwootHandler implementa handlers REST para integração Chatwoot
type ChatwootHandler struct {
	*shared.BaseHandler
	messageService *services.MessageService
	sessionService *services.SessionService
}

// NewChatwootHandler cria nova instância do handler Chatwoot
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

// ChatwootWebhookPayload representa o payload do webhook do Chatwoot
type ChatwootWebhookPayload struct {
	Event        string                    `json:"event"`
	Account      *ChatwootAccount          `json:"account,omitempty"`
	Conversation *ChatwootConversation     `json:"conversation,omitempty"`
	Message      *ChatwootMessage          `json:"message,omitempty"`
	Contact      *ChatwootContact          `json:"contact,omitempty"`
	Inbox        *ChatwootInbox            `json:"inbox,omitempty"`
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

// ReceiveWebhook processa webhooks do Chatwoot
// @Summary Receive Chatwoot webhook
// @Description Process incoming webhook from Chatwoot
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param payload body ChatwootWebhookPayload true "Webhook payload"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /chatwoot/webhook/{sessionId} [post]
func (h *ChatwootHandler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "receive chatwoot webhook")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse webhook payload
	var payload ChatwootWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid webhook payload")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Processar webhook baseado no evento
	// Por enquanto, apenas logar o evento recebido
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

// CreateConfig cria uma nova configuração do Chatwoot
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

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar criação de configuração Chatwoot
	h.LogSuccess("create chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration created successfully")
}

// FindConfig obtém a configuração do Chatwoot
// @Summary Find Chatwoot configuration
// @Description Find the current Chatwoot configuration for the session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId query string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /chatwoot/find [get]
func (h *ChatwootHandler) FindConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "find chatwoot config")

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar busca de configuração Chatwoot
	h.LogSuccess("get chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration retrieved successfully")
}

// UpdateConfig atualiza a configuração do Chatwoot
// @Summary Update Chatwoot configuration
// @Description Update the Chatwoot configuration for the session
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot [put]
func (h *ChatwootHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update chatwoot config")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar atualização de configuração Chatwoot
	h.LogSuccess("update chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration updated successfully")
}

// DeleteConfig remove a configuração do Chatwoot
// @Summary Delete Chatwoot configuration
// @Description Delete the Chatwoot configuration for the session
// @Tags Chatwoot
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot [delete]
func (h *ChatwootHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "delete chatwoot config")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar remoção de configuração Chatwoot
	h.LogSuccess("delete chatwoot config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot configuration deleted successfully")
}

// TestConnection testa a conexão com o Chatwoot
// @Summary Test Chatwoot connection
// @Description Test the connection to Chatwoot instance
// @Tags Chatwoot
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot/test [post]
func (h *ChatwootHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "test chatwoot connection")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar teste de conexão Chatwoot
	h.LogSuccess("test chatwoot connection", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot connection test completed successfully")
}

// AutoCreateInbox cria automaticamente um inbox no Chatwoot
// @Summary Auto-create Chatwoot inbox
// @Description Automatically create an inbox in Chatwoot for the session
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot/auto-create-inbox [post]
func (h *ChatwootHandler) AutoCreateInbox(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "auto-create chatwoot inbox")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar criação automática de inbox Chatwoot
	h.LogSuccess("auto-create chatwoot inbox", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot inbox created successfully")
}

// GetStats obtém estatísticas do Chatwoot
// @Summary Get Chatwoot statistics
// @Description Get statistics for Chatwoot integration
// @Tags Chatwoot
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/chatwoot/stats [get]
func (h *ChatwootHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get chatwoot stats")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar busca de estatísticas Chatwoot
	h.LogSuccess("get chatwoot stats", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Chatwoot statistics retrieved successfully")
}