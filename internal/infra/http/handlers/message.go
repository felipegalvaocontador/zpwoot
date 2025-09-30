package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/message"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/logger"
)



type MessageHandler struct {
	messageUC       message.UseCase
	wameowManager   *wameow.Manager
	sessionResolver *helpers.SessionResolver
	logger          *logger.Logger
}

func NewMessageHandler(
	messageUC message.UseCase,
	wameowManager *wameow.Manager,
	sessionRepo helpers.SessionRepository,
	logger *logger.Logger,
) *MessageHandler {
	sessionResolver := helpers.NewSessionResolver(logger, sessionRepo)

	return &MessageHandler{
		messageUC:       messageUC,
		wameowManager:   wameowManager,
		sessionResolver: sessionResolver,
		logger:          logger,
	}
}

func (h *MessageHandler) handleMessageActionWithTwoFields(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	field1Name, field1ValidationMessage string,
	field2Name, field2ValidationMessage string,
	requestBuilder func(string, string) *message.SendMessageRequest,
	logFields func(string, string) map[string]interface{},
) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var reqData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	field1Value := reqData[field1Name]
	if field1Value == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse(field1ValidationMessage))
		return
	}

	field2Value := reqData[field2Name]
	if field2Value == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse(field2ValidationMessage))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := requestBuilder(field1Value, field2Value)
	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		logFieldsMap := logFields(field1Value, field2Value)
		logFieldsMap["session_id"] = sess.ID.String()
		logFieldsMap["error"] = err.Error()
		h.logger.ErrorWithFields("Failed to "+actionName, logFieldsMap)

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to " + actionName))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, successMessage))
}

func (h *MessageHandler) handleMediaMessage(
	w http.ResponseWriter,
	r *http.Request,
	messageType string,
	parseFunc func(*http.Request) (*message.SendMessageRequest, error),
) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	req, err := parseFunc(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to send %s message", messageType), map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"has_reply":  req.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to send %s message", messageType)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, fmt.Sprintf("%s message sent successfully", titleCase(messageType))))
}

func parseMediaRequest(r *http.Request, messageType string, parseBody func(*http.Request) (string, string, string, string, string, *message.ContextInfo, error)) (*message.SendMessageRequest, error) {
	remoteJID, file, caption, mimeType, filename, contextInfo, err := parseBody(r)
	if err != nil {
		return nil, fmt.Errorf("invalid %s message format", messageType)
	}

	if remoteJID == "" {
		return nil, fmt.Errorf("'Phone' field is required")
	}

	if file == "" {
		return nil, fmt.Errorf("'file' field is required")
	}

	if contextInfo != nil {
		if contextInfo.StanzaID == "" {
			return nil, fmt.Errorf("'contextInfo.stanzaId' is required when replying")
		}
	}

	return &message.SendMessageRequest{
		RemoteJID:   remoteJID,
		Type:        messageType,
		File:        file,
		Caption:     caption,
		MimeType:    mimeType,
		Filename:    filename,
		ContextInfo: contextInfo,
	}, nil
}

// @Summary Send media message
// @Description Send a media message (image, video, audio, document) through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.MediaMessageRequest true "Media message request"
// @Success 200 {object} message.MessageResponse "Media message sent successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/messages/send/media [post]
func (h *MessageHandler) SendMedia(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var mediaReq message.MediaMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&mediaReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid media message format"))
		return
	}

	if mediaReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if mediaReq.File == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'file' field is required"))
		return
	}


	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	mediaType := h.detectMediaType(mediaReq.File, mediaReq.MimeType)

	req := &message.SendMessageRequest{
		RemoteJID:   mediaReq.RemoteJID,
		Type:        mediaType,
		File:        mediaReq.File,
		Caption:     mediaReq.Caption,
		MimeType:    mediaReq.MimeType,
		Filename:    mediaReq.Filename,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send media message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"type":       mediaType,
			"has_reply":  req.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send media message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Media message sent successfully"))
}

func (h *MessageHandler) detectMediaType(file, mimeType string) string {
	if mimeType != "" {
		if strings.HasPrefix(mimeType, "image/") {
			return "image"
		}
		if strings.HasPrefix(mimeType, "video/") {
			return "video"
		}
		if strings.HasPrefix(mimeType, "audio/") {
			return "audio"
		}
		return "document"
	}

	file = strings.ToLower(file)
	if strings.HasSuffix(file, ".jpg") || strings.HasSuffix(file, ".jpeg") || 
	   strings.HasSuffix(file, ".png") || strings.HasSuffix(file, ".gif") || 
	   strings.HasSuffix(file, ".webp") {
		return "image"
	}
	if strings.HasSuffix(file, ".mp4") || strings.HasSuffix(file, ".avi") || 
	   strings.HasSuffix(file, ".mov") || strings.HasSuffix(file, ".webm") {
		return "video"
	}
	if strings.HasSuffix(file, ".mp3") || strings.HasSuffix(file, ".wav") || 
	   strings.HasSuffix(file, ".ogg") || strings.HasSuffix(file, ".m4a") {
		return "audio"
	}

	return "image"
}

// @Summary Send text message
// @Description Send a text message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.TextMessageRequest true "Text message request"
// @Success 200 {object} message.MessageResponse "Text message sent successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/messages/send/text [post]
func (h *MessageHandler) SendText(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var textReq message.TextMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&textReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid text message format"))
		return
	}

	if textReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if textReq.Body == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'body' field is required"))
		return
	}

	if textReq.ContextInfo != nil && textReq.ContextInfo.StanzaID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID:   textReq.RemoteJID,
		Type:        "text",
		Body:        textReq.Body,
		ContextInfo: textReq.ContextInfo,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send text message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"has_reply":  req.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send text message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Text message sent successfully"))
}

// @Summary Send image message
// @Description Send an image message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.ImageMessageRequest true "Image message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/image [post]
func (h *MessageHandler) SendImage(w http.ResponseWriter, r *http.Request) {
	h.handleMediaMessage(w, r, "image", func(r *http.Request) (*message.SendMessageRequest, error) {
		return parseMediaRequest(r, "image", func(r *http.Request) (string, string, string, string, string, *message.ContextInfo, error) {
			var imageReq message.ImageMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&imageReq); err != nil {
				return "", "", "", "", "", nil, err
			}
			return imageReq.RemoteJID, imageReq.File, imageReq.Caption, imageReq.MimeType, imageReq.Filename, imageReq.ContextInfo, nil
		})
	})
}

// @Summary Send audio message
// @Description Send an audio message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.AudioMessageRequest true "Audio message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/audio [post]
func (h *MessageHandler) SendAudio(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var audioReq message.AudioMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&audioReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid audio message format"))
		return
	}

	if audioReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if audioReq.File == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'file' field is required"))
		return
	}

	if audioReq.ContextInfo != nil && audioReq.ContextInfo.StanzaID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID:   audioReq.RemoteJID,
		Type:        "audio",
		File:        audioReq.File,
		MimeType:    audioReq.MimeType,
		Caption:     audioReq.Caption,
		ContextInfo: audioReq.ContextInfo,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send audio message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"has_reply":  req.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send audio message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Audio message sent successfully"))
}

// @Summary Send video message
// @Description Send a video message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.VideoMessageRequest true "Video message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/video [post]
func (h *MessageHandler) SendVideo(w http.ResponseWriter, r *http.Request) {
	h.handleMediaMessage(w, r, "video", func(r *http.Request) (*message.SendMessageRequest, error) {
		return parseMediaRequest(r, "video", func(r *http.Request) (string, string, string, string, string, *message.ContextInfo, error) {
			var videoReq message.VideoMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&videoReq); err != nil {
				return "", "", "", "", "", nil, err
			}
			return videoReq.RemoteJID, videoReq.File, videoReq.Caption, videoReq.MimeType, videoReq.Filename, videoReq.ContextInfo, nil
		})
	})
}

// @Summary Send document message
// @Description Send a document message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.DocumentMessageRequest true "Document message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/document [post]
func (h *MessageHandler) SendDocument(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var docReq message.DocumentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&docReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid document message format"))
		return
	}

	if docReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if docReq.File == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'file' field is required"))
		return
	}

	if docReq.ContextInfo != nil && docReq.ContextInfo.StanzaID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID:   docReq.RemoteJID,
		Type:        "document",
		File:        docReq.File,
		Caption:     docReq.Caption,
		MimeType:    docReq.MimeType,
		Filename:    docReq.Filename,
		ContextInfo: docReq.ContextInfo,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send document message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"filename":   req.Filename,
			"has_reply":  req.ContextInfo != nil,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send document message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Document message sent successfully"))
}

// @Summary Send sticker message
// @Description Send a sticker message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.StickerMessageRequest true "Sticker message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/sticker [post]
func (h *MessageHandler) SendSticker(w http.ResponseWriter, r *http.Request) {
	h.sendSpecificMessageType(w, r, "sticker")
}

// @Summary Send location message
// @Description Send a location message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.LocationMessageRequest true "Location message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/location [post]
func (h *MessageHandler) SendLocation(w http.ResponseWriter, r *http.Request) {
	h.sendSpecificMessageType(w, r, "location")
}

func (h *MessageHandler) sendSpecificMessageType(w http.ResponseWriter, r *http.Request, messageType string) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		h.logger.Warn("Session identifier is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		h.logger.WarnWithFields("Session not found", map[string]interface{}{
			"session_identifier": sessionIdentifier,
			"error":              err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	var req *message.SendMessageRequest
	switch messageType {
	case "sticker":
		var stickerReq message.MediaMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&stickerReq); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid sticker message format"))
			return
		}

		if stickerReq.RemoteJID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
			return
		}

		if stickerReq.File == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("'file' field is required"))
			return
		}

		req = &message.SendMessageRequest{
			RemoteJID: stickerReq.RemoteJID,
			Type:      "sticker",
			File:      stickerReq.File,
			MimeType:  stickerReq.MimeType,
		}

	case "location":
		var locationReq message.LocationMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&locationReq); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid location message format"))
			return
		}

		if locationReq.RemoteJID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
			return
		}

		req = &message.SendMessageRequest{
			RemoteJID: locationReq.RemoteJID,
			Type:      "location",
			Latitude:  locationReq.Latitude,
			Longitude: locationReq.Longitude,
			Address:   locationReq.Address,
		}

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Unsupported message type"))
		return
	}

	if req.ContextInfo != nil && req.ContextInfo.StanzaID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'contextInfo.stanzaId' is required when replying"))
		return
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.handleSendMessageError(w, err, messageType, sess.ID.String(), req.RemoteJID)
		return
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s message sent successfully", titleCase(messageType)), map[string]interface{}{
		"session_id": sess.ID.String(),
		"to":         req.RemoteJID,
		"message_id": response.ID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, fmt.Sprintf("%s message sent successfully", titleCase(messageType))))
}

func (h *MessageHandler) handleSendMessageError(w http.ResponseWriter, err error, messageType, sessionID, remoteJID string) {
	h.logger.ErrorWithFields("Failed to send "+messageType+" message", map[string]interface{}{
		"session_id": sessionID,
		"to":         remoteJID,
		"type":       messageType,
		"error":      err.Error(),
	})

	if strings.Contains(err.Error(), "not connected") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
		return
	}

	if strings.Contains(err.Error(), "not found") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session or target not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to send %s message", messageType)))
}

// @Summary Send contact message(s)
// @Description Send a single contact or multiple contacts through WhatsApp. Automatically detects if it's a single contact (ContactMessage) or multiple contacts (ContactsArrayMessage) based on the array length.
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.ContactMessageRequest true "Contact message request (single contact or contact list)"
// @Success 200 {object} common.SuccessResponse{data=message.ContactMessageResponse} "Contact sent successfully"
// @Success 200 {object} common.SuccessResponse{data=message.ContactListMessageResponse} "Contact list sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/contact [post]
func (h *MessageHandler) SendContact(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var contactListReq message.ContactListMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&contactListReq); err == nil && len(contactListReq.Contacts) > 0 {
		h.handleContactList(w, r, sessionIdentifier)
	} else {
		h.handleSingleContact(w, r, sessionIdentifier)
	}
}

func (h *MessageHandler) handleSingleContact(w http.ResponseWriter, r *http.Request, sessionIdentifier string) {
	var contactReq message.ContactMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&contactReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid single contact format"))
		return
	}

	if contactReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if contactReq.ContactName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'contactName' field is required"))
		return
	}

	if contactReq.ContactPhone == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'contactPhone' field is required"))
		return
	}


	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID:    contactReq.RemoteJID,
		Type:         "contact",
		ContactName:  contactReq.ContactName,
		ContactPhone: contactReq.ContactPhone,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"to":            req.RemoteJID,
			"contact_name":  req.ContactName,
			"contact_phone": req.ContactPhone,
			"has_reply":     req.ContextInfo != nil,
			"error":         err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send contact message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Contact message sent successfully"))
}

func (h *MessageHandler) handleContactList(w http.ResponseWriter, r *http.Request, sessionIdentifier string) {
	contactListReq, err := h.parseContactListRequest(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	sess, sessionErr := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if sessionErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	result, err := h.sendContactListViaWameow(r.Context(), sess.ID.String(), contactListReq)
	if err != nil {
		h.handleContactSendError(w, err)
		return
	}

	h.buildContactListResponse(w, result, sess.ID.String(), contactListReq.RemoteJID, len(contactListReq.Contacts))
}

func (h *MessageHandler) parseContactListRequest(r *http.Request) (*message.ContactListMessageRequest, error) {
	var contactListReq message.ContactListMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&contactListReq); err != nil {
		return nil, fmt.Errorf("invalid contact list format")
	}

	if contactListReq.RemoteJID == "" {
		return nil, fmt.Errorf("'Phone' field is required")
	}

	if len(contactListReq.Contacts) == 0 {
		return nil, fmt.Errorf("'contacts' array cannot be empty")
	}

	if len(contactListReq.Contacts) > 5 {
		return nil, fmt.Errorf("maximum 5 contacts allowed per message")
	}

	for i, contact := range contactListReq.Contacts {
		if contact.Name == "" {
			return nil, fmt.Errorf("contact %d: 'name' field is required", i+1)
		}
		if contact.Phone == "" {
			return nil, fmt.Errorf("contact %d: 'phone' field is required", i+1)
		}
	}


	return &contactListReq, nil
}

func (h *MessageHandler) sendContactListViaWameow(ctx context.Context, sessionID string, contactListReq *message.ContactListMessageRequest) (*wameow.ContactListResult, error) {
	if h.wameowManager == nil {
		return nil, fmt.Errorf("wameow manager not available")
	}

	contacts := make([]wameow.ContactInfo, len(contactListReq.Contacts))
	for i, contact := range contactListReq.Contacts {
		contacts[i] = wameow.ContactInfo{
			Name:  contact.Name,
			Phone: contact.Phone,
		}
	}

	req := &message.SendMessageRequest{
		RemoteJID: contactListReq.RemoteJID,
		Type:      "contact_list",
	}

	response, err := h.messageUC.SendMessage(ctx, sessionID, req)
	if err != nil {
		return nil, err
	}

	result := &wameow.ContactListResult{
		Results: make([]wameow.ContactResult, len(contactListReq.Contacts)),
	}

	for i, contact := range contactListReq.Contacts {
		result.Results[i] = wameow.ContactResult{
			ContactName: contact.Name,
			MessageID:   response.ID,
			Status:      "sent",
		}
	}

	return result, nil
}

func (h *MessageHandler) handleContactSendError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "not connected") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send contact list"))
}

func (h *MessageHandler) buildContactListResponse(w http.ResponseWriter, result *wameow.ContactListResult, sessionID, remoteJID string, contactCount int) {
	contactResults := make([]map[string]interface{}, 0, len(result.Results))
	for _, r := range result.Results {
		contactResults = append(contactResults, map[string]interface{}{
			"contactName": r.ContactName,
			"messageId":   r.MessageID,
			"status":      r.Status,
		})
	}

	response := map[string]interface{}{
		"sessionId":      sessionID,
		"remoteJid":      remoteJID,
		"contactCount":   contactCount,
		"contactResults": contactResults,
		"sentAt":         time.Now(),
	}

	h.logger.InfoWithFields("Contact list sent successfully", map[string]interface{}{
		"session_id":    sessionID,
		"to":            remoteJID,
		"contact_count": contactCount,
		"success_count": len(contactResults),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Contact list sent successfully"))
}

// @Summary Send business profile
// @Description Send a business profile through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.BusinessProfileMessageRequest true "Business profile message request"
// @Success 200 {object} common.SuccessResponse{data=message.SendMessageResponse} "Business profile sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/profile/business [post]
func (h *MessageHandler) SendBusinessProfile(w http.ResponseWriter, r *http.Request) {


	h.handleMessageActionWithTwoFields(
		w,
		r,
		"send business profile message",
		"Business profile sent successfully",
		"Phone",
		"'Phone' field is required",
		"businessJid",
		"'businessJid' field is required",
		func(phone, businessJid string) *message.SendMessageRequest {
			return &message.SendMessageRequest{
				RemoteJID: phone,
				Type:      "business_profile",
				Body:      fmt.Sprintf("Business Profile: %s", businessJid),
			}
		},
		func(phone, businessJid string) map[string]interface{} {
			return map[string]interface{}{
				"to":           phone,
				"business_jid": businessJid,
			}
		},
	)
}

// @Summary Send button message
// @Description Send a button message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.ButtonMessageRequest true "Button message request"
// @Success 200 {object} common.SuccessResponse{data=message.MessageResponse} "Button message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/button [post]
func (h *MessageHandler) SendButtonMessage(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var buttonReq message.ButtonMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&buttonReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid button message format"))
		return
	}

	if buttonReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if buttonReq.Body == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'body' field is required"))
		return
	}

	if len(buttonReq.Buttons) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'buttons' array cannot be empty"))
		return
	}

	if len(buttonReq.Buttons) > 3 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("maximum 3 buttons allowed"))
		return
	}

	for i, button := range buttonReq.Buttons {
		if button.ID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("button %d: 'id' field is required", i+1)))
			return
		}
		if button.Text == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("button %d: 'text' field is required", i+1)))
			return
		}
	}


	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	buttonText := buttonReq.Body + "\n\nOptions:\n"
	for i, button := range buttonReq.Buttons {
		buttonText += fmt.Sprintf("%d. %s (ID: %s)\n", i+1, button.Text, button.ID)
	}

	req := &message.SendMessageRequest{
		RemoteJID: buttonReq.RemoteJID,
		Type:      "text",
		Body:      buttonText,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send button message", map[string]interface{}{
			"session_id":   sess.ID.String(),
			"to":           req.RemoteJID,
			"button_count": len(buttonReq.Buttons),
			"error":        err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send button message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Button message sent successfully"))
}

// @Summary Send list message
// @Description Send a list message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.ListMessageRequest true "List message request"
// @Success 200 {object} common.SuccessResponse{data=message.MessageResponse} "List message sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/list [post]
func (h *MessageHandler) SendListMessage(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var listReq message.ListMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&listReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid list message format"))
		return
	}

	if listReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if listReq.Body == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'body' field is required"))
		return
	}

	if listReq.ButtonText == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'buttonText' field is required"))
		return
	}

	if len(listReq.Sections) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'sections' array cannot be empty"))
		return
	}

	if len(listReq.Sections) > 10 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("maximum 10 sections allowed"))
		return
	}

	totalRows := 0
	for i, section := range listReq.Sections {
		if section.Title == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("section %d: 'title' field is required", i+1)))
			return
		}

		if len(section.Rows) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("section %d: 'rows' array cannot be empty", i+1)))
			return
		}

		totalRows += len(section.Rows)
		if totalRows > 10 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("maximum 10 rows allowed across all sections"))
			return
		}

		for j, row := range section.Rows {
			if row.ID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("section %d, row %d: 'id' field is required", i+1, j+1)))
				return
			}
			if row.Title == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("section %d, row %d: 'title' field is required", i+1, j+1)))
				return
			}
		}
	}


	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	listText := listReq.Body + "\n\n" + listReq.ButtonText + ":\n"
	for _, section := range listReq.Sections {
		listText += fmt.Sprintf("\n%s:\n", section.Title)
		for i, row := range section.Rows {
			listText += fmt.Sprintf("%d. %s", i+1, row.Title)
			if row.Description != "" {
				listText += fmt.Sprintf(" - %s", row.Description)
			}
			listText += fmt.Sprintf(" (ID: %s)\n", row.ID)
		}
	}

	req := &message.SendMessageRequest{
		RemoteJID: listReq.RemoteJID,
		Type:      "text",
		Body:      listText,
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send list message", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"to":            req.RemoteJID,
			"section_count": len(listReq.Sections),
			"total_rows":    totalRows,
			"error":         err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send list message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "List message sent successfully"))
}

// @Summary Send reaction
// @Description Send a reaction to a message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.ReactionRequest true "Reaction request"
// @Success 200 {object} common.SuccessResponse{data=message.ReactionResponse} "Reaction sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/reaction [post]
func (h *MessageHandler) SendReaction(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var reactionReq message.ReactionMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&reactionReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid reaction format"))
		return
	}

	if reactionReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if reactionReq.MessageID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'messageId' field is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID: reactionReq.RemoteJID,
		Type:      "text",
		Body:      fmt.Sprintf("Reaction %s to message %s", reactionReq.Reaction, reactionReq.MessageID),
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send reaction", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"message_id": reactionReq.MessageID,
			"reaction":   reactionReq.Reaction,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send reaction"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Reaction sent successfully"))
}

// @Summary Send presence
// @Description Send presence status (typing, recording, paused) through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.PresenceRequest true "Presence request"
// @Success 200 {object} common.SuccessResponse{data=message.PresenceResponse} "Presence sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/presence [post]
func (h *MessageHandler) SendPresence(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var presenceReq message.PresenceMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&presenceReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid presence format"))
		return
	}

	if presenceReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if presenceReq.Presence == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'presence' field is required"))
		return
	}

	validStates := []string{"typing", "recording", "paused"}
	isValidState := false
	for _, state := range validStates {
		if presenceReq.Presence == state {
			isValidState = true
			break
		}
	}

	if !isValidState {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'presence' must be one of: typing, recording, paused"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID: presenceReq.RemoteJID,
		Type:      "text",
		Body:      fmt.Sprintf("Presence: %s", presenceReq.Presence),
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send presence", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"presence":   presenceReq.Presence,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send presence"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Presence sent successfully"))
}

// @Summary Edit message
// @Description Edit a previously sent message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.EditMessageRequest true "Edit message request"
// @Success 200 {object} common.SuccessResponse{data=message.EditResponse} "Message edited successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/edit [post]
func (h *MessageHandler) EditMessage(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var editReq message.EditMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&editReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid edit message format"))
		return
	}

	if editReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if editReq.MessageID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'messageId' field is required"))
		return
	}

	if editReq.NewBody == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'newBody' field is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID: editReq.RemoteJID,
		Type:      "text",
		Body:      fmt.Sprintf("Edit message %s: %s", editReq.MessageID, editReq.NewBody),
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to edit message", map[string]interface{}{
			"session_id": sess.ID.String(),
			"to":         req.RemoteJID,
			"message_id": editReq.MessageID,
			"new_body":   editReq.NewBody,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to edit message"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Message edited successfully"))
}

// @Summary Mark message as read
// @Description Mark a message as read through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.MarkReadRequest true "Mark read request"
// @Success 200 {object} common.SuccessResponse{data=message.MarkReadResponse} "Message marked as read successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/mark-read [post]
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	var markReadReq struct {
		RemoteJID  string   `json:"remoteJid"`
		MessageIDs []string `json:"messageIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&markReadReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid mark read format"))
		return
	}

	if markReadReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return
	}

	if len(markReadReq.MessageIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'messageIds' array cannot be empty"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	req := &message.SendMessageRequest{
		RemoteJID: markReadReq.RemoteJID,
		Type:      "text",
		Body:      fmt.Sprintf("Mark as read: %v", markReadReq.MessageIDs),
	}

	ctx := r.Context()
	response, err := h.messageUC.SendMessage(ctx, sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to mark message as read", map[string]interface{}{
			"session_id":    sess.ID.String(),
			"to":            req.RemoteJID,
			"message_count": len(markReadReq.MessageIDs),
			"error":         err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to mark message as read"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Message marked as read successfully"))
}

// @Summary Send poll
// @Description Send a poll message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.CreatePollRequest true "Poll request"
// @Success 200 {object} common.SuccessResponse{data=message.CreatePollResponse} "Poll sent successfully"
// @Failure 400 {object} object "Invalid request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal server error"
// @Router /sessions/{sessionId}/messages/send/poll [post]
func (h *MessageHandler) SendPoll(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	pollReq, err := h.parsePollRequest(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	if err := h.validatePollRequest(w, pollReq); err != nil {
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	h.sendPollAndRespond(w, r, sess.ID.String(), pollReq)
}

func (h *MessageHandler) parsePollRequest(r *http.Request) (*message.CreatePollRequest, error) {
	var pollReq message.CreatePollRequest
	if err := json.NewDecoder(r.Body).Decode(&pollReq); err != nil {
		return nil, fmt.Errorf("invalid poll format")
	}

	return &pollReq, nil
}

func (h *MessageHandler) validatePollRequest(w http.ResponseWriter, pollReq *message.CreatePollRequest) error {
	if pollReq.RemoteJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'Phone' field is required"))
		return fmt.Errorf("phone required")
	}

	if pollReq.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("'name' field is required"))
		return fmt.Errorf("name required")
	}

	if len(pollReq.Options) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("poll must have at least 2 options"))
		return fmt.Errorf("options required")
	}

	if len(pollReq.Options) > 12 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("poll can have maximum 12 options"))
		return fmt.Errorf("too many options")
	}

	return nil
}

func (h *MessageHandler) sendPollAndRespond(w http.ResponseWriter, r *http.Request, sessionID string, pollReq *message.CreatePollRequest) {
	h.logger.InfoWithFields("Sending poll", map[string]interface{}{
		"session_id":       sessionID,
		"to":               pollReq.RemoteJID,
		"name":             pollReq.Name,
		"options_count":    len(pollReq.Options),
		"selectable_count": pollReq.SelectableOptionCount,
	})

	pollText := fmt.Sprintf("Poll: %s\n\nOptions:\n", pollReq.Name)
	for i, option := range pollReq.Options {
		pollText += fmt.Sprintf("%d. %s\n", i+1, option)
	}
	pollText += fmt.Sprintf("\nYou can select up to %d option(s)", pollReq.SelectableOptionCount)

	req := &message.SendMessageRequest{
		RemoteJID: pollReq.RemoteJID,
		Type:      "text",
		Body:      pollText,
	}

	ctx := r.Context()
	_, err := h.messageUC.SendMessage(ctx, sessionID, req)
	if err != nil {
		h.handlePollSendError(w, sessionID, pollReq, err)
		return
	}

	h.returnPollSuccess(w, sessionID, pollReq)
}

func (h *MessageHandler) handlePollSendError(w http.ResponseWriter, sessionID string, pollReq *message.CreatePollRequest, err error) {
	h.logger.ErrorWithFields("Failed to send poll", map[string]interface{}{
		"session_id": sessionID,
		"to":         pollReq.RemoteJID,
		"name":       pollReq.Name,
		"error":      err.Error(),
	})

	if strings.Contains(err.Error(), "not connected") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to send poll"))
}

func (h *MessageHandler) returnPollSuccess(w http.ResponseWriter, sessionID string, pollReq *message.CreatePollRequest) {
	h.logger.InfoWithFields("Poll sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         pollReq.RemoteJID,
		"name":       pollReq.Name,
	})

	response := map[string]interface{}{
		"sessionId":       sessionID,
		"remoteJid":       pollReq.RemoteJID,
		"name":            pollReq.Name,
		"options":         pollReq.Options,
		"selectableCount": pollReq.SelectableOptionCount,
		"sentAt":          time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Poll sent successfully"))
}

// @Summary Revoke message
// @Description Revoke a previously sent message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body message.RevokeMessageRequest true "Revoke message request"
// @Success 200 {object} common.SuccessResponse{data=message.RevokeMessageResponse} "Message revoked successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/messages/revoke [post]
func (h *MessageHandler) RevokeMessage(w http.ResponseWriter, r *http.Request) {
	h.handleMessageActionWithTwoFields(
		w,
		r,
		"revoke message",
		"Message revoked successfully",
		"remoteJid",
		"'Phone' field is required",
		"messageId",
		"'messageId' field is required",
		func(remoteJid, messageId string) *message.SendMessageRequest {
			return &message.SendMessageRequest{
				RemoteJID: remoteJid,
				Type:      "text",
				Body:      fmt.Sprintf("Revoke message: %s", messageId),
			}
		},
		func(remoteJid, messageId string) map[string]interface{} {
			return map[string]interface{}{
				"to":         remoteJid,
				"message_id": messageId,
			}
		},
	)
}

// @Summary Get poll results
// @Description Get results of a poll message through WhatsApp
// @Tags Messages
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param messageId path string true "Message ID of the poll"
// @Success 200 {object} common.SuccessResponse{data=message.GetPollResultsResponse} "Poll results retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or poll not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/messages/poll/{messageId}/results [get]
func (h *MessageHandler) GetPollResults(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	messageID := chi.URLParam(r, "messageId")
	if messageID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Message ID is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session not found"))
		return
	}

	ctx := r.Context()
	pollReq := &message.GetPollResultsRequest{
		RemoteJID:     sess.ID.String(), // Using session ID as remote JID for now
		PollMessageID: messageID,
	}
	response, err := h.messageUC.GetPollResults(ctx, pollReq)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get poll results", map[string]interface{}{
			"session_id": sess.ID.String(),
			"message_id": messageID,
			"error":      err.Error(),
		})

		if strings.Contains(err.Error(), "not connected") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Session is not connected"))
			return
		}

		if strings.Contains(err.Error(), "not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Poll not found"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get poll results"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Poll results retrieved successfully"))
}

func (h *MessageHandler) SendContactList(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		statusCode := 500
		if strings.Contains(err.Error(), "not found") {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req struct {
		Phone    string `json:"Phone"`
		Contacts []struct {
			Name  string `json:"name"`
			Phone string `json:"phone"`
		} `json:"contacts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.Phone == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Phone is required"))
		return
	}

	if len(req.Contacts) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("At least one contact is required"))
		return
	}

	h.logger.InfoWithFields("Sending contact list", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"session_name":  sess.Name,
		"to":            req.Phone,
		"contact_count": len(req.Contacts),
	})

	response := map[string]interface{}{
		"sessionId":    sess.ID.String(),
		"to":           req.Phone,
		"messageId":    "placeholder-message-id",
		"contactCount": len(req.Contacts),
		"status":       "sent",
		"message":      "SendContactList functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Contact list sent successfully"))
}

func (h *MessageHandler) SendContactListBusiness(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		statusCode := 500
		if strings.Contains(err.Error(), "not found") {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req struct {
		Phone    string `json:"Phone"`
		Contacts []struct {
			Name         string `json:"name"`
			Phone        string `json:"phone"`
			BusinessName string `json:"businessName"`
		} `json:"contacts"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.Phone == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Phone is required"))
		return
	}

	if len(req.Contacts) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("At least one contact is required"))
		return
	}

	h.logger.InfoWithFields("Sending business contact list", map[string]interface{}{
		"session_id":    sess.ID.String(),
		"session_name":  sess.Name,
		"to":            req.Phone,
		"contact_count": len(req.Contacts),
	})

	response := map[string]interface{}{
		"sessionId":    sess.ID.String(),
		"to":           req.Phone,
		"messageId":    "placeholder-message-id",
		"contactCount": len(req.Contacts),
		"status":       "sent",
		"message":      "SendContactListBusiness functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Business contact list sent successfully"))
}

func (h *MessageHandler) SendSingleContact(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		statusCode := 500
		if strings.Contains(err.Error(), "not found") {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req struct {
		Phone       string `json:"Phone"`
		ContactName string `json:"contactName"`
		ContactPhone string `json:"contactPhone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.Phone == "" || req.ContactName == "" || req.ContactPhone == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Phone, contactName, and contactPhone are required"))
		return
	}

	h.logger.InfoWithFields("Sending single contact", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"session_name":   sess.Name,
		"to":             req.Phone,
		"contact_name":   req.ContactName,
		"contact_phone":  req.ContactPhone,
	})

	response := map[string]interface{}{
		"sessionId":    sess.ID.String(),
		"to":           req.Phone,
		"messageId":    "placeholder-message-id",
		"contactName":  req.ContactName,
		"contactPhone": req.ContactPhone,
		"status":       "sent",
		"message":      "SendSingleContact functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Single contact sent successfully"))
}

func (h *MessageHandler) SendSingleContactBusiness(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		statusCode := 500
		if strings.Contains(err.Error(), "not found") {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req struct {
		Phone        string `json:"Phone"`
		ContactName  string `json:"contactName"`
		ContactPhone string `json:"contactPhone"`
		BusinessName string `json:"businessName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.Phone == "" || req.ContactName == "" || req.ContactPhone == "" || req.BusinessName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Phone, contactName, contactPhone, and businessName are required"))
		return
	}

	h.logger.InfoWithFields("Sending single business contact", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"session_name":   sess.Name,
		"to":             req.Phone,
		"contact_name":   req.ContactName,
		"contact_phone":  req.ContactPhone,
		"business_name":  req.BusinessName,
	})

	response := map[string]interface{}{
		"sessionId":     sess.ID.String(),
		"to":            req.Phone,
		"messageId":     "placeholder-message-id",
		"contactName":   req.ContactName,
		"contactPhone":  req.ContactPhone,
		"businessName":  req.BusinessName,
		"status":        "sent",
		"message":       "SendSingleContactBusiness functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Single business contact sent successfully"))
}


