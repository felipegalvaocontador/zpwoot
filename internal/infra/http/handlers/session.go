package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/session"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	pkgErrors "zpwoot/pkg/errors"
	"zpwoot/platform/logger"
)

type SessionHandler struct {
	logger          *logger.Logger
	sessionUC       session.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewSessionHandler(appLogger *logger.Logger, sessionUC session.UseCase, sessionRepo helpers.SessionRepository) *SessionHandler {
	return &SessionHandler{
		logger:          appLogger,
		sessionUC:       sessionUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func NewSessionHandlerWithoutUseCase(appLogger *logger.Logger, sessionRepo helpers.SessionRepository) *SessionHandler {
	return &SessionHandler{
		logger:          appLogger,
		sessionUC:       nil,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func (h *SessionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(common.NewErrorResponse(message)); err != nil {
		h.logger.Error("Failed to encode error response: " + err.Error())
	}
}

func (h *SessionHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	response := common.NewSuccessResponse(data, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode success response: " + err.Error())
	}
}

func (h *SessionHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
	idOrName := chi.URLParam(r, "sessionId")

	sess, err := h.sessionResolver.ResolveSession(r.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       r.URL.Path,
		})

		if err.Error() == "session not found" || errors.Is(err, domainSession.ErrSessionNotFound) {
			return nil, fmt.Errorf("session not found")
		}

		return nil, fmt.Errorf("failed to resolve session")
	}

	return sess, nil
}

func (h *SessionHandler) handleSessionAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	actionFunc func(context.Context, string) (interface{}, error),
) {
	if h.sessionUC == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Session use case not initialized")
		return
	}

	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	result, err := actionFunc(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s", actionName))
		return
	}

	response := common.NewSuccessResponse(result, fmt.Sprintf("%s retrieved successfully", titleCase(actionName)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) handleSessionActionNoReturn(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	actionFunc func(context.Context, string) error,
	successMessage string,
) {
	if h.sessionUC == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Session use case not initialized")
		return
	}

	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	err = actionFunc(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))

		appErr := &pkgErrors.AppError{}
		if errors.As(err, &appErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(appErr.Code)
			json.NewEncoder(w).Encode(common.NewErrorResponse(appErr.Message))
			return
		}

		if err.Error() == "session not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Session not found")
			return
		}

		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s", actionName))
		return
	}

	response := common.NewSuccessResponse(nil, successMessage)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Create new session
// @Description Create a new WhatsApp session with optional proxy configuration. If qrCode is true, returns QR code immediately for connection.
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body session.CreateSessionRequest true "Session creation request with optional qrCode flag"
// @Success 201 {object} session.CreateSessionResponse "Session created successfully. If qrCode was true, includes QR code data."
// @Failure 400 {object} object "Bad Request"
// @Failure 409 {object} object "Session already exists"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/create [post]
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating new session")

	if h.sessionUC == nil {
		h.logger.Error("Session use case not initialized")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Session service not available",
		})
		return
	}

	var req session.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	if isValid, errorMsg := h.sessionResolver.ValidateSessionName(req.Name); !isValid {
		h.logger.WarnWithFields("Invalid session name provided", map[string]interface{}{
			"name":  req.Name,
			"error": errorMsg,
		})

		suggested := h.sessionResolver.SuggestValidName(req.Name)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         "Invalid session name",
			"message":       errorMsg,
			"suggestedName": suggested,
			"namingRules": []string{
				"Must be 3-50 characters long",
				"Must start with a letter",
				"Can contain letters, numbers, hyphens, and underscores",
				"Cannot use reserved names (create, list, info, etc.)",
			},
		})
		return
	}

	result, err := h.sessionUC.CreateSession(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create session: " + err.Error())

		if strings.Contains(err.Error(), "Session already exists") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Session already exists",
				"message": fmt.Sprintf("A session with the name '%s' already exists. Please choose a different name.", req.Name),
			})
			return
		}

		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	response := common.NewSuccessResponse(result, "Session created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// @Summary List sessions
// @Description Get a list of all WhatsApp sessions with optional filtering
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param isConnected query bool false "Filter by connection status"
// @Param deviceJid query string false "Filter by device JID"
// @Param limit query int false "Number of sessions to return (default: 20)"
// @Param offset query int false "Number of sessions to skip (default: 0)"
// @Success 200 {object} session.ListSessionsResponse "Sessions retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/list [get]
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Listing sessions")

	if h.sessionUC == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Session use case not initialized")
		return
	}

	var req session.ListSessionsRequest

	if isConnectedStr := r.URL.Query().Get("isConnected"); isConnectedStr != "" {
		switch isConnectedStr {
		case "true":
			isConnected := true
			req.IsConnected = &isConnected
		case "false":
			isConnected := false
			req.IsConnected = &isConnected
		}
	}

	if deviceJid := r.URL.Query().Get("deviceJid"); deviceJid != "" {
		req.DeviceJid = &deviceJid
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}
	req.Limit = limit

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	req.Offset = offset

	result, err := h.sessionUC.ListSessions(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list sessions: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	response := common.NewSuccessResponse(result, "Sessions retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get session information
// @Description Get detailed information about a specific WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} session.SessionInfoResponse "Session information retrieved successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/info [get]
func (h *SessionHandler) GetSessionInfo(w http.ResponseWriter, r *http.Request) {
	h.handleSessionAction(w, r, "get session info", func(ctx context.Context, sessionID string) (interface{}, error) {
		return h.sessionUC.GetSessionInfo(ctx, sessionID)
	})
}

// @Summary Delete session
// @Description Delete a WhatsApp session and all associated data
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} common.SuccessResponse "Session deleted successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/delete [delete]
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	h.handleSessionActionNoReturn(w, r, "delete session", h.sessionUC.DeleteSession, "Session deleted successfully")
}

// @Summary Connect session
// @Description Connect a WhatsApp session to start receiving messages. Automatically returns QR code (both string and base64 image) if device needs to be paired. If session is already connected, returns confirmation message.
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} session.ConnectSessionResponse "Session connection initiated successfully with QR code if needed, or confirmation if already connected"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/connect [post]
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	h.handleSessionAction(w, r, "connect session", func(ctx context.Context, sessionID string) (interface{}, error) {
		return h.sessionUC.ConnectSession(ctx, sessionID)
	})
}

// @Summary Logout session
// @Description Logout from WhatsApp session and disconnect
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} common.SuccessResponse "Session logout successful"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/logout [post]
func (h *SessionHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	h.handleSessionActionNoReturn(w, r, "logout session", h.sessionUC.LogoutSession, "Session logged out successfully")
}

// @Summary Get QR code
// @Description Get QR code for WhatsApp session pairing. Returns both raw QR code string and base64 image.
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} session.QRCodeResponse "QR code generated successfully with base64 image"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/qr [get]
func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	h.handleSessionAction(w, r, "get QR code", func(ctx context.Context, sessionID string) (interface{}, error) {
		return h.sessionUC.GetQRCode(ctx, sessionID)
	})
}

// @Summary Pair phone number
// @Description Pair WhatsApp session with phone number
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body session.PairPhoneRequest true "Phone pairing request"
// @Success 200 {object} common.SuccessResponse "Phone pairing initiated successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/pair [post]
func (h *SessionHandler) PairPhone(w http.ResponseWriter, r *http.Request) {
	if h.sessionUC == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Session use case not initialized")
		return
	}

	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req session.PairPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse pair phone request: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	ctx := r.Context()
	err = h.sessionUC.PairPhone(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to pair phone: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to pair phone")
		return
	}

	response := common.NewSuccessResponse(nil, "Phone pairing initiated successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Set proxy configuration
// @Description Set or update proxy configuration for a WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body session.SetProxyRequest true "Proxy configuration request"
// @Success 200 {object} session.ProxyResponse "Proxy configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/proxy/set [post]
func (h *SessionHandler) SetProxy(w http.ResponseWriter, r *http.Request) {
	if h.sessionUC == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Session use case not initialized")
		return
	}

	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Setting proxy", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	var req session.SetProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	err = h.sessionUC.SetProxy(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to set proxy: " + err.Error())
		if err.Error() == "session not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Session not found")
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to set proxy")
		return
	}

	response := common.NewSuccessResponse(nil, "Proxy configuration updated successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get proxy configuration
// @Description Get current proxy configuration for a WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} session.ProxyResponse "Proxy configuration retrieved successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/proxy/find [get]
func (h *SessionHandler) GetProxy(w http.ResponseWriter, r *http.Request) {
	if h.sessionUC == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Session use case not initialized")
		return
	}

	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Getting proxy config", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.sessionUC.GetProxy(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to get proxy: " + err.Error())
		if err.Error() == "session not found" {
			h.writeErrorResponse(w, http.StatusNotFound, "Session not found")
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get proxy")
		return
	}

	response := common.NewSuccessResponse(result, "Proxy configuration retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Getting session stats", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	response := map[string]interface{}{
		"sessionId":        sess.ID.String(),
		"sessionName":      sess.Name,
		"messagesReceived": 0,
		"messagesSent":     0,
		"uptime":           "0h 0m 0s",
		"status":           "active",
		"message":          "Session stats functionality needs to be implemented in use case",
	}

	h.writeSuccessResponse(w, response, "Session stats retrieved successfully")
}

func (h *SessionHandler) GetUserJID(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Getting user JID", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	response := map[string]interface{}{
		"sessionId": sess.ID.String(),
		"userJid":   "placeholder@s.whatsapp.net",
		"message":   "Get user JID functionality needs to be implemented in use case",
	}

	h.writeSuccessResponse(w, response, "User JID retrieved successfully")
}

func (h *SessionHandler) GetDeviceInfo(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Getting device info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	response := map[string]interface{}{
		"sessionId":    sess.ID.String(),
		"deviceId":     "placeholder-device-id",
		"platform":     "web",
		"appVersion":   "2.2412.54",
		"osVersion":    "0.1",
		"manufacturer": "zpwoot",
		"message":      "Get device info functionality needs to be implemented in use case",
	}

	h.writeSuccessResponse(w, response, "Device info retrieved successfully")
}
