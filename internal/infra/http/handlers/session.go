package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/session"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	pkgErrors "zpwoot/pkg/errors"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
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
		sessionUC:       nil, // Will be nil until properly wired
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func (h *SessionHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if err.Error() == "session not found" || errors.Is(err, domainSession.ErrSessionNotFound) {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Failed to resolve session")
	}

	return sess, nil
}

// handleSessionAction handles common session action logic
func (h *SessionHandler) handleSessionAction(
	c *fiber.Ctx,
	actionName string,
	actionFunc func(context.Context, string) (interface{}, error),
) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	result, err := actionFunc(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))
		return c.Status(500).JSON(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName)))
	}

	response := common.NewSuccessResponse(result, fmt.Sprintf("%s retrieved successfully", titleCase(actionName)))
	return c.JSON(response)
}

// handleSessionActionNoReturn handles session actions that don't return data
func (h *SessionHandler) handleSessionActionNoReturn(
	c *fiber.Ctx,
	actionName string,
	actionFunc func(context.Context, string) error,
	successMessage string,
) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	err := actionFunc(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))

		// Handle specific error types
		appErr := &pkgErrors.AppError{}
		if errors.As(err, &appErr) {
			return c.Status(appErr.Code).JSON(common.NewErrorResponse(appErr.Message))
		}

		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName)))
	}

	response := common.NewSuccessResponse(nil, successMessage)
	return c.JSON(response)
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
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	h.logger.Info("Creating new session")

	if h.sessionUC == nil {
		h.logger.Error("Session use case not initialized")
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Session service not available",
		})
	}

	var req session.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	if isValid, errorMsg := h.sessionResolver.ValidateSessionName(req.Name); !isValid {
		h.logger.WarnWithFields("Invalid session name provided", map[string]interface{}{
			"name":  req.Name,
			"error": errorMsg,
		})

		suggested := h.sessionResolver.SuggestValidName(req.Name)
		return c.Status(400).JSON(fiber.Map{
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
	}

	result, err := h.sessionUC.CreateSession(c.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create session: " + err.Error())

		if strings.Contains(err.Error(), "Session already exists") {
			return c.Status(409).JSON(fiber.Map{
				"success": false,
				"error":   "Session already exists",
				"message": fmt.Sprintf("A session with the name '%s' already exists. Please choose a different name.", req.Name),
			})
		}

		return c.Status(500).JSON(common.NewErrorResponse("Failed to create session"))
	}

	response := common.NewSuccessResponse(result, "Session created successfully")
	return c.Status(201).JSON(response)
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
func (h *SessionHandler) ListSessions(c *fiber.Ctx) error {
	h.logger.Info("Listing sessions")

	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	var req session.ListSessionsRequest

	if isConnectedStr := c.Query("isConnected"); isConnectedStr != "" {
		switch isConnectedStr {
		case "true":
			isConnected := true
			req.IsConnected = &isConnected
		case "false":
			isConnected := false
			req.IsConnected = &isConnected
		}
	}

	if deviceJid := c.Query("deviceJid"); deviceJid != "" {
		req.DeviceJid = &deviceJid
	}

	if limit := c.QueryInt("limit", 20); limit > 0 && limit <= 100 {
		req.Limit = limit
	} else {
		req.Limit = 20
	}

	if offset := c.QueryInt("offset", 0); offset >= 0 {
		req.Offset = offset
	}

	result, err := h.sessionUC.ListSessions(c.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to list sessions: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to list sessions"))
	}

	response := common.NewSuccessResponse(result, "Sessions retrieved successfully")
	return c.JSON(response)
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
func (h *SessionHandler) GetSessionInfo(c *fiber.Ctx) error {
	return h.handleSessionAction(c, "get session info", func(ctx context.Context, sessionID string) (interface{}, error) {
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
func (h *SessionHandler) DeleteSession(c *fiber.Ctx) error {
	return h.handleSessionActionNoReturn(c, "delete session", h.sessionUC.DeleteSession, "Session deleted successfully")
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
func (h *SessionHandler) ConnectSession(c *fiber.Ctx) error {
	return h.handleSessionAction(c, "connect session", func(ctx context.Context, sessionID string) (interface{}, error) {
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
func (h *SessionHandler) LogoutSession(c *fiber.Ctx) error {
	return h.handleSessionActionNoReturn(c, "logout session", h.sessionUC.LogoutSession, "Session logged out successfully")
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
func (h *SessionHandler) GetQRCode(c *fiber.Ctx) error {
	return h.handleSessionAction(c, "get QR code", func(ctx context.Context, sessionID string) (interface{}, error) {
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
func (h *SessionHandler) PairPhone(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	var req session.PairPhoneRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse pair phone request: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	ctx := c.Context()
	err := h.sessionUC.PairPhone(ctx, sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to pair phone: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to pair phone"))
	}

	response := common.NewSuccessResponse(nil, "Phone pairing initiated successfully")
	return c.JSON(response)
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
func (h *SessionHandler) SetProxy(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Setting proxy", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	var req session.SetProxyRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	err := h.sessionUC.SetProxy(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to set proxy: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to set proxy"))
	}

	response := common.NewSuccessResponse(nil, "Proxy configuration updated successfully")
	return c.JSON(response)
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
func (h *SessionHandler) GetProxy(c *fiber.Ctx) error {
	if h.sessionUC == nil {
		return c.Status(500).JSON(common.NewErrorResponse("Session use case not initialized"))
	}

	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Getting proxy config", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.sessionUC.GetProxy(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to get proxy: " + err.Error())
		if err.Error() == "session not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Session not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get proxy"))
	}

	response := common.NewSuccessResponse(result, "Proxy configuration retrieved successfully")
	return c.JSON(response)
}
