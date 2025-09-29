package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"zpwoot/internal/app/chatwoot"
	domainChatwoot "zpwoot/internal/domain/chatwoot"
	"zpwoot/pkg/errors"
	"zpwoot/platform/logger"
)

type ChatwootHandler struct {
	chatwootUC chatwoot.UseCase
	logger     *logger.Logger
}

type ChatwootService interface {
	CreateConfig(ctx context.Context, req *chatwoot.CreateChatwootConfigRequest) (*chatwoot.CreateChatwootConfigResponse, error)
	GetConfig(ctx context.Context) (*chatwoot.ChatwootConfigResponse, error)
	UpdateConfig(ctx context.Context, req *chatwoot.UpdateChatwootConfigRequest) (*chatwoot.ChatwootConfigResponse, error)
	DeleteConfig(ctx context.Context) error
	SyncContact(ctx context.Context, req *chatwoot.SyncContactRequest) (*chatwoot.SyncContactResponse, error)
	SyncConversation(ctx context.Context, req *chatwoot.SyncConversationRequest) (*chatwoot.SyncConversationResponse, error)
	ProcessWebhook(ctx context.Context, payload *chatwoot.ChatwootWebhookPayload) error
}

func NewChatwootHandler(chatwootUC chatwoot.UseCase, logger *logger.Logger) *ChatwootHandler {
	return &ChatwootHandler{
		chatwootUC: chatwootUC,
		logger:     logger,
	}
}

// @Summary Set Chatwoot configuration
// @Description Set or update Chatwoot integration configuration for a WhatsApp session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body chatwoot.CreateChatwootConfigRequest true "Chatwoot configuration request"
// @Success 200 {object} chatwoot.CreateChatwootConfigResponse "Chatwoot configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) CreateConfig(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	var req chatwoot.CreateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if auto-create is requested
	if req.AutoCreate != nil && *req.AutoCreate {
		// Generate webhook URL dynamically
		baseURL := h.getBaseURL(c)
		webhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", baseURL, sessionID)

		inboxName := "WhatsApp zpwoot"
		if req.InboxName != nil && *req.InboxName != "" {
			inboxName = *req.InboxName
		}

		// Try to auto-create inbox
		err := h.chatwootUC.AutoCreateInbox(c.Context(), sessionID, inboxName, webhookURL)
		if err != nil {
			h.logger.WarnWithFields("Failed to auto-create inbox", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			// Don't fail the request, just log the warning
		}
	}

	config, err := h.chatwootUC.CreateConfig(c.Context(), sessionID, &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration created successfully",
		"data":    config,
	})
}

func (h *ChatwootHandler) GetConfig(c *fiber.Ctx) error {
	config, err := h.chatwootUC.GetConfig(c.Context())
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

func (h *ChatwootHandler) UpdateConfig(c *fiber.Ctx) error {
	var req chatwoot.UpdateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	config, err := h.chatwootUC.UpdateConfig(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    config,
	})
}

func (h *ChatwootHandler) DeleteConfig(c *fiber.Ctx) error {
	err := h.chatwootUC.DeleteConfig(c.Context())
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration deleted successfully",
	})
}

func (h *ChatwootHandler) SyncContacts(c *fiber.Ctx) error {
	var req chatwoot.SyncContactRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	contact, err := h.chatwootUC.SyncContact(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    contact,
	})
}

func (h *ChatwootHandler) SyncConversations(c *fiber.Ctx) error {
	var req chatwoot.SyncConversationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	conversation, err := h.chatwootUC.SyncConversation(c.Context(), &req)
	if err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			return c.Status(appErr.Code).JSON(fiber.Map{
				"error":   appErr.Message,
				"details": appErr.Details,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    conversation,
	})
}

func (h *ChatwootHandler) ReceiveWebhook(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	// Log webhook reception
	h.logWebhookReception(sessionID, c)

	// Validate session ID
	if err := h.validateSessionID(sessionID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Parse webhook payload
	payload, err := h.parseWebhookPayload(c, sessionID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate event type
	if err := h.validateEventType(sessionID, payload.Event, c.Body()); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error(), "event": payload.Event})
	}

	// Process webhook
	if err := h.processWebhook(c, sessionID, payload); err != nil {
		return h.handleWebhookError(c, sessionID, payload.Event, err)
	}

	// Return success response
	return h.returnWebhookSuccess(c, sessionID, payload.Event)
}

// logWebhookReception logs the webhook reception details
func (h *ChatwootHandler) logWebhookReception(sessionID string, c *fiber.Ctx) {
	h.logger.InfoWithFields("Received Chatwoot webhook", map[string]interface{}{
		"session_id": sessionID,
		"ip":         c.IP(),
		"user_agent": c.Get("User-Agent"),
	})

	// Log raw body for debugging
	rawBody := c.Body()
	h.logger.DebugWithFields("Raw webhook payload", map[string]interface{}{
		"session_id": sessionID,
		"body":       string(rawBody),
		"body_size":  len(rawBody),
	})
}

// validateSessionID validates the session ID format
func (h *ChatwootHandler) validateSessionID(sessionID string) error {
	if _, err := uuid.Parse(sessionID); err != nil {
		h.logger.WarnWithFields("Invalid session ID format", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("Invalid session ID format")
	}
	return nil
}

// parseWebhookPayload parses the webhook payload from request body
func (h *ChatwootHandler) parseWebhookPayload(c *fiber.Ctx, sessionID string) (*chatwoot.ChatwootWebhookPayload, error) {
	rawBody := c.Body()
	var payload chatwoot.ChatwootWebhookPayload

	if err := c.BodyParser(&payload); err != nil {
		h.logger.WarnWithFields("Failed to parse webhook payload", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
			"raw_body":   string(rawBody),
		})
		return nil, fmt.Errorf("Invalid webhook payload")
	}

	// Log parsed payload for debugging
	h.logger.DebugWithFields("Parsed webhook payload", map[string]interface{}{
		"session_id": sessionID,
		"event":      payload.Event,
		"account_id": payload.Account.ID,
		"conv_id":    payload.Conversation.ID,
		"message_id": payload.ID,
	})

	return &payload, nil
}

// validateEventType validates the webhook event type
func (h *ChatwootHandler) validateEventType(sessionID, event string, rawBody []byte) error {
	if !domainChatwoot.IsValidChatwootEvent(event) {
		h.logger.WarnWithFields("Invalid event type", map[string]interface{}{
			"session_id": sessionID,
			"event":      event,
			"raw_body":   string(rawBody),
		})
		return fmt.Errorf("Invalid event type")
	}
	return nil
}

// processWebhook processes the webhook using the use case
func (h *ChatwootHandler) processWebhook(c *fiber.Ctx, sessionID string, payload *chatwoot.ChatwootWebhookPayload) error {
	return h.chatwootUC.ProcessWebhook(c.Context(), sessionID, payload)
}

// handleWebhookError handles webhook processing errors
func (h *ChatwootHandler) handleWebhookError(c *fiber.Ctx, sessionID, event string, err error) error {
	h.logger.ErrorWithFields("Failed to process webhook", map[string]interface{}{
		"session_id": sessionID,
		"event":      event,
		"error":      err.Error(),
	})

	if appErr := errors.GetAppError(err); appErr != nil {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"error":   appErr.Message,
			"details": appErr.Details,
		})
	}
	return c.Status(500).JSON(fiber.Map{
		"error": "Internal server error",
	})
}

// returnWebhookSuccess returns a successful webhook response
func (h *ChatwootHandler) returnWebhookSuccess(c *fiber.Ctx, sessionID, event string) error {
	h.logger.InfoWithFields("Webhook processed successfully", map[string]interface{}{
		"session_id": sessionID,
		"event":      event,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Webhook processed successfully",
		"event":   event,
	})
}

func (h *ChatwootHandler) TestConnection(c *fiber.Ctx) error {
	ctx := c.Context()

	result, err := h.chatwootUC.TestConnection(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Chatwoot connection test failed",
			"error":   err.Error(),
			"status":  "failed",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot connection test completed",
		"data":    result,
		"status":  "connected",
	})
}

func (h *ChatwootHandler) GetStats(c *fiber.Ctx) error {
	ctx := c.Context()

	stats, err := h.chatwootUC.GetStats(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get Chatwoot statistics",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// @Summary Set Chatwoot configuration
// @Description Set or update Chatwoot integration configuration for a WhatsApp session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body chatwoot.CreateChatwootConfigRequest true "Chatwoot configuration request"
// @Success 200 {object} chatwoot.CreateChatwootConfigResponse "Chatwoot configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) SetConfig(c *fiber.Ctx) error {
	var req chatwoot.CreateChatwootConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	ctx := c.Context()

	_, err := h.chatwootUC.GetConfig(ctx)

	if err != nil {
		sessionID := c.Params("sessionId")
		result, createErr := h.chatwootUC.CreateConfig(ctx, sessionID, &req)
		if createErr != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create Chatwoot configuration",
				"error":   createErr.Error(),
			})
		}

		// Auto-create inbox if requested
		if req.AutoCreate != nil && *req.AutoCreate {
			h.logger.InfoWithFields("Auto-creating Chatwoot inbox", map[string]interface{}{
				"session_id":  sessionID,
				"auto_create": true,
			})

			// Generate inbox name (use provided name or default to session name)
			inboxName := "WhatsApp zpwoot"
			if req.InboxName != nil && *req.InboxName != "" {
				inboxName = *req.InboxName
			}

			// Generate webhook URL automatically using server configuration
			serverHost := os.Getenv("SERVER_HOST")
			if serverHost == "" {
				serverHost = "http://localhost:8080" // fallback
			}
			webhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", serverHost, sessionID)

			// Call auto-creation logic (this would need to be implemented in the use case)
			autoCreateErr := h.chatwootUC.AutoCreateInbox(ctx, sessionID, inboxName, webhookURL)
			if autoCreateErr != nil {
				h.logger.WarnWithFields("Failed to auto-create inbox", map[string]interface{}{
					"session_id": sessionID,
					"error":      autoCreateErr.Error(),
				})
				// Don't fail the entire request, just log the warning
			}
		}

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"message": "Chatwoot configuration created successfully",
			"data":    result,
		})
	}

	updateReq := chatwoot.UpdateChatwootConfigRequest{
		URL:       &req.URL,
		Token:     &req.Token,
		AccountID: &req.AccountID,
		InboxID:   req.InboxID,
	}

	result, updateErr := h.chatwootUC.UpdateConfig(ctx, &updateReq)
	if updateErr != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update Chatwoot configuration",
			"error":   updateErr.Error(),
		})
	}

	// Auto-create inbox if requested (also for updates)
	if req.AutoCreate != nil && *req.AutoCreate {
		sessionID := c.Params("sessionId")
		h.logger.InfoWithFields("Auto-creating Chatwoot inbox", map[string]interface{}{
			"session_id":  sessionID,
			"auto_create": true,
		})

		// Generate inbox name (use provided name or default to session name)
		inboxName := "WhatsApp zpwoot"
		if req.InboxName != nil && *req.InboxName != "" {
			inboxName = *req.InboxName
		}

		// Generate webhook URL automatically using server configuration
		serverHost := os.Getenv("SERVER_HOST")
		if serverHost == "" {
			serverHost = "http://localhost:8080" // fallback
		}
		webhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", serverHost, sessionID)

		// Call auto-creation logic
		autoCreateErr := h.chatwootUC.AutoCreateInbox(ctx, sessionID, inboxName, webhookURL)
		if autoCreateErr != nil {
			h.logger.WarnWithFields("Failed to auto-create inbox", map[string]interface{}{
				"session_id": sessionID,
				"error":      autoCreateErr.Error(),
			})
			// Don't fail the entire request, just log the warning
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration updated successfully",
		"data":    result,
	})
}

// @Summary Get Chatwoot configuration
// @Description Get current Chatwoot integration configuration for a WhatsApp session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} chatwoot.ChatwootConfigResponse "Chatwoot configuration retrieved successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/find [get]
func (h *ChatwootHandler) FindConfig(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	ctx := c.Context()
	config, err := h.chatwootUC.GetConfig(ctx)

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success":    false,
			"message":    "Chatwoot configuration not found for this session",
			"session_id": sessionID,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Chatwoot configuration found",
		"data":    config,
	})
}

// getBaseURL gets the base URL from server configuration
func (h *ChatwootHandler) getBaseURL(c *fiber.Ctx) string {
	// Use SERVER_HOST from environment configuration
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "http://localhost:8080" // fallback
	}

	return serverHost
}
