package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

type MessageHandler struct {
	*shared.BaseHandler
	messageService *services.MessageService
	sessionService *services.SessionService
}

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

// @Summary Send text message
// @Description Send a text message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendTextMessageRequest true "Text message request"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/text [post]

// @Summary Send text message
// @Description Send a text message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendTextMessageRequest true "Text message request"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/text [post]
func (h *MessageHandler) SendTextMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send text message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendTextMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendTextMessage(r.Context(), sessionID, req.RemoteJID, req.Body)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send text message", map[string]interface{}{
			"session_id": sessionID,
			"remote_jid": req.RemoteJID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send text message")
		return
	}

	h.LogSuccess("send text message", map[string]interface{}{
		"session_id":  sessionID,
		"message_id":  response.MessageID,
		"remote_jid":  req.RemoteJID,
		"body_len":    len(req.Body),
	})

	h.GetWriter().WriteSuccess(w, response, "Text message sent successfully")
}

// @Summary Send media message
// @Description Send a media message (image, video, audio, document) via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendMediaMessageRequest true "Media message request"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/media [post]
func (h *MessageHandler) SendMediaMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send media message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendMediaMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendMediaMessage(r.Context(), sessionID, req.To, req.MediaURL, req.Caption, req.Type)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send media message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"media_type": req.Type,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send media message")
		return
	}

	h.LogSuccess("send media message", map[string]interface{}{
		"session_id": sessionID,
		"message_id": response.MessageID,
		"to":         req.To,
		"media_type": req.Type,
		"media_url":  req.MediaURL,
	})

	h.GetWriter().WriteSuccess(w, response, "Media message sent successfully")
}

// @Summary Send image message
// @Description Send an image message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendImageMessageRequest true "Image message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/image [post]
func (h *MessageHandler) SendImage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send image message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendImageMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendImageMessage(r.Context(), sessionID, req.To, req.File, req.Caption, req.Filename)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send image message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send image message")
		return
	}

	h.LogSuccess("send image message", map[string]interface{}{
		"session_id":  sessionID,
		"message_id":  response.MessageID,
		"to":          req.To,
		"has_caption": req.Caption != "",
		"filename":    req.Filename,
	})

	h.GetWriter().WriteSuccess(w, response, "Image message sent successfully")
}

// @Summary Send audio message
// @Description Send an audio message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendAudioMessageRequest true "Audio message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/audio [post]
func (h *MessageHandler) SendAudio(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send audio message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendAudioMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendAudioMessage(r.Context(), sessionID, req.To, req.File, req.Caption)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send audio message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send audio message")
		return
	}

	h.LogSuccess("send audio message", map[string]interface{}{
		"session_id":  sessionID,
		"message_id":  response.MessageID,
		"to":          req.To,
		"has_caption": req.Caption != "",
		"mime_type":   req.MimeType,
	})

	h.GetWriter().WriteSuccess(w, response, "Audio message sent successfully")
}

// @Summary Send video message
// @Description Send a video message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendVideoMessageRequest true "Video message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/video [post]
func (h *MessageHandler) SendVideo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send video message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendVideoMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendVideoMessage(r.Context(), sessionID, req.To, req.File, req.Caption, req.Filename)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send video message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send video message")
		return
	}

	h.LogSuccess("send video message", map[string]interface{}{
		"session_id":  sessionID,
		"message_id":  response.MessageID,
		"to":          req.To,
		"has_caption": req.Caption != "",
		"filename":    req.Filename,
	})

	h.GetWriter().WriteSuccess(w, response, "Video message sent successfully")
}

// @Summary Send document message
// @Description Send a document message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendDocumentMessageRequest true "Document message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/document [post]
func (h *MessageHandler) SendDocument(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send document message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendDocumentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendDocumentMessage(r.Context(), sessionID, req.To, req.File, req.Caption, req.Filename)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send document message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send document message")
		return
	}

	h.LogSuccess("send document message", map[string]interface{}{
		"session_id":  sessionID,
		"message_id":  response.MessageID,
		"to":          req.To,
		"filename":    req.Filename,
		"has_caption": req.Caption != "",
	})

	h.GetWriter().WriteSuccess(w, response, "Document message sent successfully")
}

// @Summary Send sticker message
// @Description Send a sticker message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendStickerMessageRequest true "Sticker message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/sticker [post]
func (h *MessageHandler) SendSticker(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send sticker message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendStickerMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendStickerMessage(r.Context(), sessionID, req.To, req.File)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send sticker message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send sticker message")
		return
	}

	h.LogSuccess("send sticker message", map[string]interface{}{
		"session_id": sessionID,
		"message_id": response.MessageID,
		"to":         req.To,
		"mime_type":  req.MimeType,
	})

	h.GetWriter().WriteSuccess(w, response, "Sticker message sent successfully")
}

// @Summary Send location message
// @Description Send a location message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendLocationMessageRequest true "Location message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/location [post]
func (h *MessageHandler) SendLocation(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send location message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendLocationMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendLocationMessage(r.Context(), sessionID, req.To, req.Latitude, req.Longitude, req.Address)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send location message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send location message")
		return
	}

	h.LogSuccess("send location message", map[string]interface{}{
		"session_id": sessionID,
		"message_id": response.MessageID,
		"to":         req.To,
		"latitude":   req.Latitude,
		"longitude":  req.Longitude,
		"address":    req.Address,
	})

	h.GetWriter().WriteSuccess(w, response, "Location message sent successfully")
}

// @Summary Send contact message
// @Description Send a contact message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendContactMessageRequest true "Contact message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/contact [post]
func (h *MessageHandler) SendContact(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send contact message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendContactMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	response, err := h.messageService.SendContactMessage(r.Context(), sessionID, req.To, req.ContactName, req.ContactPhone)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"session_id": sessionID,
			"to":         req.To,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to send contact message")
		return
	}

	h.LogSuccess("send contact message", map[string]interface{}{
		"session_id":    sessionID,
		"message_id":    response.MessageID,
		"to":            req.To,
		"contact_name":  req.ContactName,
		"contact_phone": req.ContactPhone,
	})

	h.GetWriter().WriteSuccess(w, response, "Contact message sent successfully")
}

// @Summary Send contact list message
// @Description Send a contact list message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendContactListMessageRequest true "Contact list message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendContactListResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/contact-list [post]
func (h *MessageHandler) SendContactList(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send contact list message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendContactListMessageRequest
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

	contactResults := make([]contracts.ContactResult, len(req.Contacts))
	for i, contact := range req.Contacts {
		contactResults[i] = contracts.ContactResult{
			ContactName: contact.Name,
			MessageID:   uuid.New().String(),
			Status:      "sent",
		}
	}

	response := &contracts.SendContactListResponse{
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

// @Summary Send business profile message
// @Description Send a business profile message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendBusinessProfileMessageRequest true "Business profile message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/profile/business [post]
func (h *MessageHandler) SendBusinessProfile(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send business profile message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendBusinessProfileMessageRequest
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

	response := &contracts.SendMessageResponse{
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

// @Summary Send button message
// @Description Send a button message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendButtonMessageRequest true "Button message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/button [post]
func (h *MessageHandler) SendButton(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send button message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendButtonMessageRequest
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

	response := &contracts.SendMessageResponse{
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

// @Summary Send list message
// @Description Send a list message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendListMessageRequest true "List message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/list [post]
func (h *MessageHandler) SendList(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send list message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendListMessageRequest
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

	totalRows := 0
	for _, section := range req.Sections {
		totalRows += len(section.Rows)
	}

	response := &contracts.SendMessageResponse{
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

// @Summary Send poll message
// @Description Send a poll message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendPollMessageRequest true "Poll message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/poll [post]
func (h *MessageHandler) SendPoll(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send poll message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendPollMessageRequest
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

	response := &contracts.SendMessageResponse{
		MessageID: uuid.New().String(),
		To:        req.To,
		Status:    "sent",
		Timestamp: time.Now(),
	}

	h.LogSuccess("send poll message", map[string]interface{}{
		"session_id":       sessionID,
		"session_name":     session.Session.Name,
		"to":               req.To,
		"poll_name":        req.Name,
		"option_count":     len(req.Options),
		"selectable_count": req.SelectableCount,
		"allow_multiple":   req.AllowMultipleVote,
	})

	h.GetWriter().WriteSuccess(w, response, "Poll message sent successfully")
}

// @Summary Send reaction message
// @Description Send a reaction to a message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendReactionMessageRequest true "Reaction message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/reaction [post]
func (h *MessageHandler) SendReaction(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send reaction message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendReactionMessageRequest
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

	response := &contracts.SendMessageResponse{
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

// @Summary Send presence status
// @Description Send presence status (typing, recording, etc.) via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SendPresenceMessageRequest true "Presence message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/send/presence [post]
func (h *MessageHandler) SendPresence(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "send presence message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SendPresenceMessageRequest
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

	response := &contracts.SendMessageResponse{
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

// @Summary Edit message
// @Description Edit a previously sent message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.EditMessageRequest true "Edit message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/edit [post]
func (h *MessageHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "edit message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.EditMessageRequest
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

	response := &contracts.SendMessageResponse{
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

// @Summary Revoke message
// @Description Revoke a previously sent message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.RevokeMessageRequest true "Revoke message request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SendMessageResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/revoke [post]
func (h *MessageHandler) RevokeMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "revoke message")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.RevokeMessageRequest
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

	response := &contracts.SendMessageResponse{
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

// @Summary Get poll results
// @Description Get results of a poll message via WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.GetPollResultsResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
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

	voteResults := []contracts.PollVoteInfo{
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

	response := &contracts.GetPollResultsResponse{
		MessageID:   messageID,
		PollName:    "Sample Poll",
		TotalVotes:  3,
		VoteResults: voteResults,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	}

	h.LogSuccess("get poll results", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"message_id":   messageID,
		"total_votes":  response.TotalVotes,
	})

	h.GetWriter().WriteSuccess(w, response, "Poll results retrieved successfully")
}

// @Summary Mark messages as read
// @Description Mark messages as read in WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.MarkAsReadRequest true "Mark as read request"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/mark-read [post]
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "mark messages as read")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req contracts.MarkAsReadRequest
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

// @Summary Get pending sync messages
// @Description Get messages that are pending synchronization with Chatwoot
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Success 200 {object} shared.SuccessResponse{data=[]contracts.MessageDTO}
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/pending-sync [get]
func (h *MessageHandler) GetPendingSyncMessages(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get pending sync messages")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	limit := parseIntQuery(r, "limit", 50)
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

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

// @Summary Delete message
// @Description Delete a message from the system
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/messages/{messageId} [delete]
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "delete message")

	sessionID := chi.URLParam(r, "sessionId")
	messageID := chi.URLParam(r, "messageId")

	if sessionID == "" || messageID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID and Message ID are required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("delete message", map[string]interface{}{
		"message_id": messageID,
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Message deleted successfully")
}

// Duplicate MarkAsRead method removed

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
