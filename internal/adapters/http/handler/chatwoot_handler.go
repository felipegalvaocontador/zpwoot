package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/http/shared"
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
// @Success 200 {object} shared.APIResponse
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
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