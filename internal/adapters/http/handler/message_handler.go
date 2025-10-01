package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zpwoot/internal/adapters/http/shared"
	"zpwoot/internal/services"
	"zpwoot/internal/services/shared/dto"
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

// Todos os DTOs agora estão centralizados em internal/services/shared/dto/message_dto.go

// DTOs removidos - agora usando os centralizados de internal/services/shared/dto/message_dto.go

// Todos os DTOs foram movidos para internal/services/shared/dto/message_dto.go

// ===== CRUD OPERATIONS =====

// CreateMessage cria uma nova mensagem
// @Summary Create message
// @Description Create a new message in the system
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.CreateMessageRequest true "Message creation request"
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
	var req dto.CreateMessageRequest
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
// @Param request body dto.UpdateSyncStatusRequest true "Sync status update request"
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
	var req dto.UpdateSyncStatusRequest
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
// @Param request body dto.SendTextMessageRequest true "Text message request"
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
	var req dto.SendTextMessageRequest
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
// @Param request body dto.SendMediaMessageRequest true "Media message request"
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
	var req dto.SendMediaMessageRequest
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

// SendImage envia uma mensagem de imagem
// @Summary Send image message
// @Description Send an image message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendImageMessageRequest true "Image message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/image [post]
func (h *MessageHandler) SendImage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send image message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req dto.SendImageMessageRequest
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
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send image message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"has_caption":  req.Caption != "",
		"filename":     req.Filename,
	})

	h.GetWriter().WriteSuccess(w, response, "Image message sent successfully")
}

// SendAudio envia uma mensagem de áudio
// @Summary Send audio message
// @Description Send an audio message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendAudioMessageRequest true "Audio message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/audio [post]
func (h *MessageHandler) SendAudio(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send audio message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req dto.SendAudioMessageRequest
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
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send audio message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"has_caption":  req.Caption != "",
		"mime_type":    req.MimeType,
	})

	h.GetWriter().WriteSuccess(w, response, "Audio message sent successfully")
}

// SendVideo envia uma mensagem de vídeo
// @Summary Send video message
// @Description Send a video message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendVideoMessageRequest true "Video message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/video [post]
func (h *MessageHandler) SendVideo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send video message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendVideoMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send video message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"has_caption":  req.Caption != "",
		"filename":     req.Filename,
	})

	h.GetWriter().WriteSuccess(w, response, "Video message sent successfully")
}

// SendDocument envia uma mensagem de documento
// @Summary Send document message
// @Description Send a document message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendDocumentMessageRequest true "Document message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/document [post]
func (h *MessageHandler) SendDocument(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send document message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendDocumentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send document message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"filename":     req.Filename,
		"has_caption":  req.Caption != "",
	})

	h.GetWriter().WriteSuccess(w, response, "Document message sent successfully")
}

// SendSticker envia uma mensagem de sticker
// @Summary Send sticker message
// @Description Send a sticker message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendStickerMessageRequest true "Sticker message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/sticker [post]
func (h *MessageHandler) SendSticker(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send sticker message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendStickerMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send sticker message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"mime_type":    req.MimeType,
	})

	h.GetWriter().WriteSuccess(w, response, "Sticker message sent successfully")
}

// SendLocation envia uma mensagem de localização
// @Summary Send location message
// @Description Send a location message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendLocationMessageRequest true "Location message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/location [post]
func (h *MessageHandler) SendLocation(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send location message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendLocationMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send location message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"latitude":     req.Latitude,
		"longitude":    req.Longitude,
		"address":      req.Address,
	})

	h.GetWriter().WriteSuccess(w, response, "Location message sent successfully")
}

// SendContact envia uma mensagem de contato
// @Summary Send contact message
// @Description Send a contact message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendContactMessageRequest true "Contact message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/contact [post]
func (h *MessageHandler) SendContact(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send contact message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendContactMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send contact message", map[string]interface{}{
		"session_id":    sessionID,
		"session_name":  session.Session.Name,
		"to":            req.To,
		"contact_name":  req.ContactName,
		"contact_phone": req.ContactPhone,
	})

	h.GetWriter().WriteSuccess(w, response, "Contact message sent successfully")
}

// SendContactList envia uma lista de contatos
// @Summary Send contact list message
// @Description Send a contact list message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendContactListMessageRequest true "Contact list message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendContactListResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/contact-list [post]
func (h *MessageHandler) SendContactList(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send contact list message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendContactListMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	contactResults := make([]dto.ContactResult, len(req.Contacts))
	for i, contact := range req.Contacts {
		contactResults[i] = dto.ContactResult{
			ContactName: contact.Name,
			MessageID:   uuid.New().String(),
			Status:      "sent",
		}
	}

	response := &dto.SendContactListResponse{
		SessionID:      sessionID,
		RemoteJID:      req.To,
		ContactCount:   len(req.Contacts),
		ContactResults: contactResults,
		SentAt:         time.Now(),
	}

	h.LogSuccess("send contact list message", map[string]interface{}{
		"session_id":    sessionID,
		"session_name":  session.Session.Name,
		"to":            req.To,
		"contact_count": len(req.Contacts),
	})

	h.GetWriter().WriteSuccess(w, response, "Contact list sent successfully")
}

// SendBusinessProfile envia um perfil de negócio
// @Summary Send business profile message
// @Description Send a business profile message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendBusinessProfileMessageRequest true "Business profile message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/profile/business [post]
func (h *MessageHandler) SendBusinessProfile(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send business profile message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendBusinessProfileMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send business profile message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"business_jid": req.BusinessJID,
	})

	h.GetWriter().WriteSuccess(w, response, "Business profile sent successfully")
}

// SendButton envia uma mensagem com botões
// @Summary Send button message
// @Description Send a button message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendButtonMessageRequest true "Button message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/button [post]
func (h *MessageHandler) SendButton(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send button message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendButtonMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send button message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"button_count": len(req.Buttons),
	})

	h.GetWriter().WriteSuccess(w, response, "Button message sent successfully")
}

// SendList envia uma mensagem com lista
// @Summary Send list message
// @Description Send a list message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendListMessageRequest true "List message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/list [post]
func (h *MessageHandler) SendList(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send list message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendListMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Contar total de rows
	totalRows := 0
	for _, section := range req.Sections {
		totalRows += len(section.Rows)
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send list message", map[string]interface{}{
		"session_id":    sessionID,
		"session_name":  session.Session.Name,
		"to":            req.To,
		"section_count": len(req.Sections),
		"total_rows":    totalRows,
	})

	h.GetWriter().WriteSuccess(w, response, "List message sent successfully")
}

// SendPoll envia uma mensagem de poll
// @Summary Send poll message
// @Description Send a poll message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendPollMessageRequest true "Poll message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/poll [post]
func (h *MessageHandler) SendPoll(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send poll message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendPollMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send poll message", map[string]interface{}{
		"session_id":         sessionID,
		"session_name":       session.Session.Name,
		"to":                 req.To,
		"poll_name":          req.Name,
		"option_count":       len(req.Options),
		"selectable_count":   req.SelectableCount,
		"allow_multiple":     req.AllowMultipleVote,
	})

	h.GetWriter().WriteSuccess(w, response, "Poll message sent successfully")
}

// SendReaction envia uma reação a uma mensagem
// @Summary Send reaction message
// @Description Send a reaction to a message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendReactionMessageRequest true "Reaction message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/reaction [post]
func (h *MessageHandler) SendReaction(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send reaction message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendReactionMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send reaction message", map[string]interface{}{
		"session_id":        sessionID,
		"session_name":      session.Session.Name,
		"to":                req.To,
		"target_message_id": req.MessageID,
		"reaction":          req.Reaction,
	})

	h.GetWriter().WriteSuccess(w, response, "Reaction sent successfully")
}

// SendPresence envia status de presença
// @Summary Send presence status
// @Description Send presence status (typing, recording, etc.) via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.SendPresenceMessageRequest true "Presence message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/send/presence [post]
func (h *MessageHandler) SendPresence(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send presence message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.SendPresenceMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar envio via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send presence message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"presence":     req.Presence,
	})

	h.GetWriter().WriteSuccess(w, response, "Presence sent successfully")
}

// EditMessage edita uma mensagem enviada
// @Summary Edit message
// @Description Edit a previously sent message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.EditMessageRequest true "Edit message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/edit [post]
func (h *MessageHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "edit message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar edição via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: req.MessageID,
		To:        req.To,
		Status:    "edited",
		Timestamp: time.Now(),
	}

	h.LogSuccess("edit message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"message_id":   req.MessageID,
		"new_body_len": len(req.NewBody),
	})

	h.GetWriter().WriteSuccess(w, response, "Message edited successfully")
}

// RevokeMessage revoga uma mensagem enviada
// @Summary Revoke message
// @Description Revoke a previously sent message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body dto.RevokeMessageRequest true "Revoke message request"
// @Success 200 {object} shared.APIResponse{data=dto.SendMessageResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/revoke [post]
func (h *MessageHandler) RevokeMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "revoke message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req dto.RevokeMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar revogação via WhatsApp Gateway
	response := &dto.SendMessageResponse{
		MessageID: req.MessageID,
		To:        req.To,
		Status:    "revoked",
		Timestamp: time.Now(),
	}

	h.LogSuccess("revoke message", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"to":           req.To,
		"message_id":   req.MessageID,
	})

	h.GetWriter().WriteSuccess(w, response, "Message revoked successfully")
}

// GetPollResults obtém resultados de um poll
// @Summary Get poll results
// @Description Get results of a poll message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID"
// @Success 200 {object} shared.APIResponse{data=dto.GetPollResultsResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/messages/poll/{messageId}/results [get]
func (h *MessageHandler) GetPollResults(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get poll results")

	sessionID := chi.URLParam(r, "sessionId")
	messageID := chi.URLParam(r, "messageId")

	if sessionID == "" || messageID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID and Message ID are required")
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar busca de resultados via WhatsApp Gateway
	// Por enquanto, simular resposta de sucesso
	voteResults := []dto.PollVoteInfo{
		{
			OptionName: "Option 1",
			Voters:     []string{"5511888888888@s.whatsapp.net", "5511777777777@s.whatsapp.net"},
			VoteCount:  2,
		},
		{
			OptionName: "Option 2",
			Voters:     []string{"5511666666666@s.whatsapp.net"},
			VoteCount:  1,
		},
	}

	response := &dto.GetPollResultsResponse{
		MessageID:   messageID,
		PollName:    "Sample Poll",
		TotalVotes:  3,
		VoteResults: voteResults,
		CreatedAt:   time.Now().Add(-24 * time.Hour), // Simular poll criado há 1 dia
	}

	h.LogSuccess("get poll results", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"message_id":   messageID,
		"total_votes":  response.TotalVotes,
	})

	h.GetWriter().WriteSuccess(w, response, "Poll results retrieved successfully")
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
// @Param request body dto.MarkAsReadRequest true "Mark as read request"
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
	var req dto.MarkAsReadRequest
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

	response := &dto.MarkAsReadResponse{
		ChatJID:     req.ChatJID,
		MarkedCount: len(req.MessageIDs),
		Status:      "success",
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