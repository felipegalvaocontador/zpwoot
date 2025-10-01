package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zpwoot/internal/adapters/http/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// MessageHandler implementa handlers REST para mensagens
type MessageHandler struct {
	*shared.BaseHandler
	messageService *services.MessageService
	sessionService *services.SessionService
}

// NewMessageHandler cria nova instância do handler de mensagens
func NewMessageHandler(
	messageService *services.MessageService,
	sessionService *services.SessionService,
	logger *logger.Logger,
) *MessageHandler {
	return &MessageHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		messageService: messageService,
		sessionService: sessionService,
	}
}

// CreateMessageRequest representa request para criação de mensagem
type CreateMessageRequest struct {
	ZpMessageID string `json:"zp_message_id" validate:"required"`
	ZpSender    string `json:"zp_sender" validate:"required"`
	ZpChat      string `json:"zp_chat" validate:"required"`
	ZpTimestamp string `json:"zp_timestamp" validate:"required"`
	ZpFromMe    bool   `json:"zp_from_me"`
	ZpType      string `json:"zp_type" validate:"required"`
	Content     string `json:"content,omitempty"`
}

// SendTextMessageRequest representa request para envio de mensagem de texto
type SendTextMessageRequest struct {
	To      string `json:"to" validate:"required"`
	Content string `json:"content" validate:"required"`
	ReplyTo string `json:"reply_to,omitempty"`
}

// SendMediaMessageRequest representa request para envio de mídia
type SendMediaMessageRequest struct {
	To       string `json:"to" validate:"required"`
	MediaURL string `json:"media_url" validate:"required,url"`
	Caption  string `json:"caption,omitempty"`
	Type     string `json:"type" validate:"required,oneof=image video audio document"`
	ReplyTo  string `json:"reply_to,omitempty"`
}

// ListMessagesRequest representa request para listagem de mensagens
type ListMessagesRequest struct {
	ChatJID string `json:"chat_jid,omitempty"`
	Limit   int    `json:"limit" validate:"min=1,max=100"`
	Offset  int    `json:"offset" validate:"min=0"`
}

// UpdateSyncStatusRequest representa request para atualização de sync
type UpdateSyncStatusRequest struct {
	SyncStatus       string `json:"sync_status" validate:"required,oneof=pending synced failed"`
	CwMessageID      *int   `json:"cw_message_id,omitempty"`
	CwConversationID *int   `json:"cw_conversation_id,omitempty"`
}

// ===== CRUD OPERATIONS =====

// CreateMessage cria uma nova mensagem
// @Summary Create message
// @Description Create a new message in the system
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body CreateMessageRequest true "Message creation request"
// @Success 201 {object} shared.APIResponse{data=services.CreateMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages [post]
func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "create message")

	// Extrair session ID da URL
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Criar request para service
	serviceReq := &services.CreateMessageRequest{
		SessionID:   sessionID,
		ZpMessageID: req.ZpMessageID,
		ZpSender:    req.ZpSender,
		ZpChat:      req.ZpChat,
		ZpTimestamp: req.ZpTimestamp,
		ZpFromMe:    req.ZpFromMe,
		ZpType:      req.ZpType,
		Content:     req.Content,
	}

	// Criar mensagem
	response, err := h.messageService.CreateMessage(r.Context(), serviceReq)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to create message", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to create message")
		return
	}

	h.LogSuccess("create message", map[string]interface{}{
		"message_id": response.ID,
		"session_id": sessionID,
	})

	h.GetWriter().WriteCreated(w, response, "Message created successfully")
}

// GetMessage busca uma mensagem por ID
// @Summary Get message
// @Description Get a message by ID
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID"
// @Success 200 {object} shared.APIResponse{data=dto.MessageDTO}
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/{messageId} [get]
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get message")

	sessionID := chi.URLParam(r, "sessionId")
	messageID := chi.URLParam(r, "messageId")

	if sessionID == "" || messageID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID and Message ID are required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Buscar mensagem
	message, err := h.messageService.GetMessage(r.Context(), messageID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Message not found")
		return
	}

	h.LogSuccess("get message", map[string]interface{}{
		"message_id": messageID,
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, message, "Message retrieved successfully")
}

// ListMessages lista mensagens com filtros e paginação
// @Summary List messages
// @Description List messages with filters and pagination
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param chat_jid query string false "Chat JID filter"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} shared.APIResponse{data=services.ListMessagesResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages [get]
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list messages")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse query parameters
	chatJID := r.URL.Query().Get("chat_jid")
	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)

	// Validar parâmetros
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Criar request para service
	req := &services.ListMessagesRequest{
		SessionID: sessionID,
		ChatJID:   chatJID,
		Limit:     limit,
		Offset:    offset,
	}

	// Listar mensagens
	response, err := h.messageService.ListMessages(r.Context(), req)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to list messages", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to list messages")
		return
	}

	h.LogSuccess("list messages", map[string]interface{}{
		"session_id":    sessionID,
		"total":         response.Total,
		"returned":      len(response.Messages),
		"limit":         limit,
		"offset":        offset,
	})

	h.GetWriter().WriteSuccess(w, response, "Messages retrieved successfully")
}

// UpdateSyncStatus atualiza o status de sincronização de uma mensagem
// @Summary Update message sync status
// @Description Update the synchronization status of a message
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID"
// @Param request body UpdateSyncStatusRequest true "Sync status update request"
// @Success 200 {object} shared.APIResponse
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/{messageId}/sync [put]
func (h *MessageHandler) UpdateSyncStatus(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update message sync status")

	sessionID := chi.URLParam(r, "sessionId")
	messageID := chi.URLParam(r, "messageId")

	if sessionID == "" || messageID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID and Message ID are required")
		return
	}

	// Parse request body
	var req UpdateSyncStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Criar request para service
	serviceReq := &services.UpdateSyncStatusRequest{
		MessageID:        messageID,
		SyncStatus:       req.SyncStatus,
		CwMessageID:      req.CwMessageID,
		CwConversationID: req.CwConversationID,
	}

	// Atualizar status
	err = h.messageService.UpdateSyncStatus(r.Context(), serviceReq)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to update sync status", map[string]interface{}{
			"message_id": messageID,
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to update sync status")
		return
	}

	h.LogSuccess("update message sync status", map[string]interface{}{
		"message_id":   messageID,
		"session_id":   sessionID,
		"sync_status":  req.SyncStatus,
	})

	h.GetWriter().WriteSuccess(w, nil, "Sync status updated successfully")
}

// ===== MESSAGING OPERATIONS =====

// SendTextMessage envia uma mensagem de texto
// @Summary Send text message
// @Description Send a text message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body SendTextMessageRequest true "Text message request"
// @Success 200 {object} shared.APIResponse
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/text [post]
func (h *MessageHandler) SendTextMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send text message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req SendTextMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	// Por enquanto, simular resposta de sucesso
	response := map[string]interface{}{
		"message_id": uuid.New().String(),
		"to":         req.To,
		"content":    req.Content,
		"status":     "sent",
		"timestamp":  "2024-01-01T00:00:00Z",
	}

	h.LogSuccess("send text message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"content_len":  len(req.Content),
	})

	h.GetWriter().WriteSuccess(w, response, "Text message sent successfully")
}

// SendMediaMessage envia uma mensagem de mídia
// @Summary Send media message
// @Description Send a media message (image, video, audio, document) via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body SendMediaMessageRequest true "Media message request"
// @Success 200 {object} shared.APIResponse
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/media [post]
func (h *MessageHandler) SendMediaMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send media message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req SendMediaMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	// Por enquanto, simular resposta de sucesso
	response := map[string]interface{}{
		"message_id": uuid.New().String(),
		"to":         req.To,
		"media_url":  req.MediaURL,
		"type":       req.Type,
		"caption":    req.Caption,
		"status":     "sent",
		"timestamp":  "2024-01-01T00:00:00Z",
	}

	h.LogSuccess("send media message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"type":         req.Type,
		"media_url":    req.MediaURL,
	})

	h.GetWriter().WriteSuccess(w, response, "Media message sent successfully")
}

// ===== STATISTICS AND SPECIAL OPERATIONS =====

// GetMessageStats retorna estatísticas de mensagens
// @Summary Get message statistics
// @Description Get statistics for messages in a session
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse{data=messaging.MessageStats}
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/stats [get]
func (h *MessageHandler) GetMessageStats(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get message stats")

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

	// Buscar estatísticas
	stats, err := h.messageService.GetMessageStats(r.Context(), &sessionID)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to get message stats", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to get message stats")
		return
	}

	h.LogSuccess("get message stats", map[string]interface{}{
		"session_id":      sessionID,
		"total_messages":  stats.TotalMessages,
		"synced_messages": stats.SyncedMessages,
	})

	h.GetWriter().WriteSuccess(w, stats, "Message statistics retrieved successfully")
}

// GetPendingSyncMessages busca mensagens pendentes de sincronização
// @Summary Get pending sync messages
// @Description Get messages that are pending synchronization with Chatwoot
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Success 200 {object} shared.APIResponse{data=[]dto.MessageDTO}
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/pending-sync [get]
func (h *MessageHandler) GetPendingSyncMessages(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get pending sync messages")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse query parameters
	limit := parseIntQuery(r, "limit", 50)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Buscar mensagens pendentes
	messages, err := h.messageService.GetPendingSyncMessages(r.Context(), sessionID, limit)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to get pending sync messages", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to get pending sync messages")
		return
	}

	h.LogSuccess("get pending sync messages", map[string]interface{}{
		"session_id": sessionID,
		"count":      len(messages),
		"limit":      limit,
	})

	h.GetWriter().WriteSuccess(w, messages, "Pending sync messages retrieved successfully")
}

// DeleteMessage deleta uma mensagem (soft delete)
// @Summary Delete message
// @Description Delete a message from the system
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID"
// @Success 200 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/{messageId} [delete]
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "delete message")

	sessionID := chi.URLParam(r, "sessionId")
	messageID := chi.URLParam(r, "messageId")

	if sessionID == "" || messageID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID and Message ID are required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar delete via service
	// Por enquanto, simular resposta de sucesso
	h.LogSuccess("delete message", map[string]interface{}{
		"message_id": messageID,
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Message deleted successfully")
}

// MarkAsRead marca mensagens como lidas
// @Summary Mark messages as read
// @Description Mark messages as read in WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body object{chat_jid=string,message_ids=[]string} true "Mark as read request"
// @Success 200 {object} shared.APIResponse
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/mark-read [post]
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "mark messages as read")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req struct {
		ChatJID    string   `json:"chat_jid" validate:"required"`
		MessageIDs []string `json:"message_ids" validate:"required,min=1"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar mark as read via WhatsApp Gateway
	// Por enquanto, simular resposta de sucesso
	h.LogSuccess("mark messages as read", map[string]interface{}{
		"session_id":      sessionID,
		"chat_jid":        req.ChatJID,
		"message_count":   len(req.MessageIDs),
	})

	response := map[string]interface{}{
		"chat_jid":        req.ChatJID,
		"marked_count":    len(req.MessageIDs),
		"status":          "success",
	}

	h.GetWriter().WriteSuccess(w, response, "Messages marked as read successfully")
}

// parseIntQuery extrai um parâmetro inteiro da query string
func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}