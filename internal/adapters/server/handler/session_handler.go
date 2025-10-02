package handler

import (
	"net/http"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/platform/logger"
)

// SessionHandler gerencia endpoints relacionados a sessões
type SessionHandler struct {
	*shared.BaseHandler
	sessionService *services.SessionService
}

// NewSessionHandler cria nova instância do session handler
func NewSessionHandler(sessionService *services.SessionService, logger *logger.Logger) *SessionHandler {
	return &SessionHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		sessionService: sessionService,
	}
}

// CreateSession cria uma nova sessão
// @Summary Create new session
// @Description Create a new WhatsApp session with optional proxy configuration. If qrCode is true, returns QR code immediately for connection.
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body contracts.CreateSessionRequest true "Session creation request with optional qrCode flag"
// @Success 201 {object} shared.SuccessResponse{data=contracts.CreateSessionResponse} "Session created successfully. If qrCode was true, includes QR code data."
// @Failure 400 {object} shared.ErrorResponse "Bad Request"
// @Failure 409 {object} shared.ErrorResponse "Session already exists"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/create [post]
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "create session")

	// Parse e validar request
	var req contracts.CreateSessionRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	// Executar no service
	response, err := h.sessionService.CreateSession(r.Context(), &req)
	if err != nil {
		h.HandleError(w, err, "create session")
		return
	}

	// Log sucesso
	h.LogSuccess("create session", map[string]interface{}{
		"session_id":   response.ID,
		"session_name": response.Name,
		"has_qr_code":  response.QRCode != "",
	})

	// Resposta de sucesso
	h.GetWriter().WriteCreated(w, response, "Session created successfully")
}

// ListSessions lista sessões com filtros opcionais
// @Summary List sessions
// @Description Get a list of all WhatsApp sessions with optional filtering
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param isConnected query bool false "Filter by connection status"
// @Param deviceJid query string false "Filter by device JID"
// @Param limit query int false "Number of sessions to return (default: 20)"
// @Param offset query int false "Number of sessions to skip (default: 0)"
// @Success 200 {object} shared.SuccessResponse{data=contracts.ListSessionsResponse} "Sessions retrieved successfully"
// @Failure 400 {object} shared.ErrorResponse "Bad Request"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/list [get]
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list sessions")

	// Extrair parâmetros de query
	limit, offset, err := h.GetPaginationParams(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid pagination parameters", err.Error())
		return
	}

	// Parâmetros de filtro opcionais
	isConnected, err := h.GetQueryBool(r, "isConnected")
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid isConnected parameter", err.Error())
		return
	}

	deviceJID := h.GetQueryString(r, "deviceJid")

	// Montar request
	req := &contracts.ListSessionsRequest{
		Limit:  limit,
		Offset: offset,
	}

	// Adicionar filtros se fornecidos
	if r.URL.Query().Has("isConnected") {
		req.IsConnected = &isConnected
	}
	if deviceJID != "" {
		req.DeviceJID = &deviceJID
	}

	// Executar no service
	response, err := h.sessionService.ListSessions(r.Context(), req)
	if err != nil {
		h.HandleError(w, err, "list sessions")
		return
	}

	// Log sucesso
	h.LogSuccess("list sessions", map[string]interface{}{
		"total_sessions": response.Total,
		"limit":          response.Limit,
		"offset":         response.Offset,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, "Sessions retrieved successfully")
}

// GetSessionInfo obtém informações detalhadas de uma sessão
// @Summary Get session information
// @Description Get detailed information about a specific WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SessionInfoResponse} "Session information retrieved successfully"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/info [get]
func (h *SessionHandler) GetSessionInfo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get session info")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	response, err := h.sessionService.GetSession(r.Context(), sessionID.String())
	if err != nil {
		h.HandleError(w, err, "get session info")
		return
	}

	// Log sucesso
	h.LogSuccess("get session info", map[string]interface{}{
		"session_id": sessionID.String(),
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, "Session information retrieved successfully")
}

// DeleteSession remove uma sessão
// @Summary Delete session
// @Description Delete a WhatsApp session and all associated data
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse "Session deleted successfully"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/delete [delete]
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "delete session")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	if err := h.sessionService.DeleteSession(r.Context(), sessionID.String()); err != nil {
		h.HandleError(w, err, "delete session")
		return
	}

	// Log sucesso
	h.LogSuccess("delete session", map[string]interface{}{
		"session_id": sessionID.String(),
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, nil, "Session deleted successfully")
}

// ConnectSession conecta uma sessão
// @Summary Connect session
// @Description Connect a WhatsApp session to start receiving messages. Automatically returns QR code (both string and base64 image) if device needs to be paired. If session is already connected, returns confirmation message.
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.ConnectSessionResponse} "Session connection initiated successfully with QR code if needed, or confirmation if already connected"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/connect [post]
func (h *SessionHandler) ConnectSession(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "connect session")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	response, err := h.sessionService.ConnectSession(r.Context(), sessionID.String())
	if err != nil {
		h.HandleError(w, err, "connect session")
		return
	}

	// Log sucesso
	h.LogSuccess("connect session", map[string]interface{}{
		"session_id": sessionID.String(),
		"success":    response.Success,
		"has_qr":     response.QRCode != "",
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// DisconnectSession desconecta uma sessão
// @Summary Disconnect session
// @Description Disconnect from WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse "Session disconnected successfully"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/disconnect [post]
func (h *SessionHandler) DisconnectSession(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "disconnect session")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	if err := h.sessionService.DisconnectSession(r.Context(), sessionID.String()); err != nil {
		h.HandleError(w, err, "disconnect session")
		return
	}

	// Log sucesso
	h.LogSuccess("disconnect session", map[string]interface{}{
		"session_id": sessionID.String(),
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, nil, "Session disconnected successfully")
}

// GetQRCode obtém QR code para pareamento
// @Summary Get QR code
// @Description Get QR code for WhatsApp session pairing. Returns both raw QR code string and base64 image.
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.QRCodeResponse} "QR code generated successfully with base64 image"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/qr [get]
func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get QR code")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	response, err := h.sessionService.GetQRCode(r.Context(), sessionID.String())
	if err != nil {
		h.HandleError(w, err, "get QR code")
		return
	}

	// Log sucesso
	h.LogSuccess("get QR code", map[string]interface{}{
		"session_id": sessionID.String(),
		"expires_at": response.ExpiresAt,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, "QR code retrieved successfully")
}

// GenerateQRCode gera novo QR code
// @Summary Generate QR code
// @Description Generate a new QR code for WhatsApp session pairing
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.QRCodeResponse} "QR code generated successfully"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/qr/generate [post]
func (h *SessionHandler) GenerateQRCode(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "generate QR code")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	response, err := h.sessionService.GenerateQRCode(r.Context(), sessionID.String())
	if err != nil {
		h.HandleError(w, err, "generate QR code")
		return
	}

	// Log sucesso
	h.LogSuccess("generate QR code", map[string]interface{}{
		"session_id": sessionID.String(),
		"expires_at": response.ExpiresAt,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, "QR code generated successfully")
}

// SetProxy configura proxy para uma sessão
// @Summary Set proxy
// @Description Configure proxy settings for a WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SetProxyRequest true "Proxy configuration"
// @Success 200 {object} shared.SuccessResponse "Proxy configured successfully"
// @Failure 400 {object} shared.ErrorResponse "Bad Request"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/proxy/set [post]
func (h *SessionHandler) SetProxy(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set proxy")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Parse e validar request
	var req contracts.SetProxyRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	// Executar no service
	if err := h.sessionService.SetProxy(r.Context(), sessionID.String(), &req); err != nil {
		h.HandleError(w, err, "set proxy")
		return
	}

	// Log sucesso
	h.LogSuccess("set proxy", map[string]interface{}{
		"session_id":  sessionID.String(),
		"proxy_type":  req.ProxyConfig.Type,
		"proxy_host":  req.ProxyConfig.Host,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, nil, "Proxy configured successfully")
}

// GetProxy obtém configuração de proxy
// @Summary Get proxy
// @Description Get proxy configuration for a WhatsApp session
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.ProxyResponse} "Proxy configuration retrieved successfully"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/proxy [get]
func (h *SessionHandler) GetProxy(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get proxy")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Executar no service
	response, err := h.sessionService.GetProxy(r.Context(), sessionID.String())
	if err != nil {
		h.HandleError(w, err, "get proxy")
		return
	}

	// Log sucesso
	h.LogSuccess("get proxy", map[string]interface{}{
		"session_id": sessionID.String(),
		"has_proxy":  response.ProxyConfig != nil,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, "Proxy configuration retrieved successfully")
}

// GetSessionStats obtém estatísticas das sessões
// @Summary Get session statistics
// @Description Get statistics about all sessions
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} shared.SuccessResponse{data=contracts.SessionStatsResponse} "Session statistics retrieved successfully"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/stats [get]
func (h *SessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get session stats")

	// Executar no service
	response, err := h.sessionService.GetSessionStats(r.Context())
	if err != nil {
		h.HandleError(w, err, "get session stats")
		return
	}

	// Log sucesso
	h.LogSuccess("get session stats", map[string]interface{}{
		"total":     response.Total,
		"connected": response.Connected,
		"offline":   response.Offline,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, "Session statistics retrieved successfully")
}

// LogoutSession desconecta e faz logout da sessão
// @Summary Logout session
// @Description Logout from WhatsApp session and disconnect
// @Tags Sessions
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse "Session logout successful"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/logout [post]
func (h *SessionHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "logout session")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// TODO: Implementar logout no service
	// Por enquanto, usar disconnect
	err = h.sessionService.DisconnectSession(r.Context(), sessionID.String())
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to logout session", map[string]interface{}{
			"session_id": sessionID.String(),
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to logout session")
		return
	}

	h.LogSuccess("logout session", map[string]interface{}{
		"session_id": sessionID.String(),
	})

	response := map[string]interface{}{
		"session_id": sessionID.String(),
		"status":     "logged_out",
		"message":    "Session logged out successfully",
	}

	h.GetWriter().WriteSuccess(w, response, "Session logged out successfully")
}

// PairPhone pareia sessão com número de telefone
// @Summary Pair phone number
// @Description Pair WhatsApp session with phone number
// @Tags Sessions
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.PairPhoneRequest true "Phone pairing request"
// @Success 200 {object} shared.SuccessResponse "Phone pairing initiated successfully"
// @Failure 400 {object} shared.ErrorResponse "Bad Request"
// @Failure 404 {object} shared.ErrorResponse "Session not found"
// @Failure 500 {object} shared.ErrorResponse "Internal Server Error"
// @Router /sessions/{sessionId}/pair [post]
func (h *SessionHandler) PairPhone(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "pair phone")

	// Extrair session ID da URL
	sessionID, err := h.GetSessionIDFromURL(r)
	if err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid session ID", err.Error())
		return
	}

	// Parse e validar request
	var req contracts.PairPhoneRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	// TODO: Implementar pair phone no service
	// Por enquanto, retornar placeholder
	h.LogSuccess("pair phone", map[string]interface{}{
		"session_id":   sessionID.String(),
		"phone_number": req.PhoneNumber,
	})

	response := map[string]interface{}{
		"session_id":   sessionID.String(),
		"phone_number": req.PhoneNumber,
		"status":       "pairing_initiated",
		"message":      "Phone pairing initiated successfully - Implementation pending",
	}

	h.GetWriter().WriteSuccess(w, response, "Phone pairing initiated successfully")
}