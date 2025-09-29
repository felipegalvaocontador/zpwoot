package handlers

import (
	"fmt"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/webhook"
	domainWebhook "zpwoot/internal/domain/webhook"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type WebhookHandler struct {
	webhookUC webhook.UseCase
	logger    *logger.Logger
}

func NewWebhookHandler(webhookUC webhook.UseCase, appLogger *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		webhookUC: webhookUC,
		logger:    appLogger,
	}
}

// @Summary Set webhook configuration
// @Description Create or update webhook configuration for a WhatsApp session. Set enabled=true to activate, enabled=false to disable without deleting. If enabled is not provided, defaults to true.
// @Tags Webhooks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID" format(uuid)
// @Param request body webhook.SetConfigRequest true "Webhook configuration request"
// @Success 201 {object} webhook.SetConfigResponse "Webhook configuration created/updated successfully"
// @Failure 400 {object} object "Bad Request - Invalid session ID, URL, or event types"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/set [post]
func (h *WebhookHandler) SetConfig(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	h.logger.InfoWithFields("Creating webhook config", map[string]interface{}{
		"session_id": sessionID,
	})

	// Validate session ID format
	if _, err := uuid.Parse(sessionID); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid session ID format"))
	}

	var req webhook.SetConfigRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse webhook request: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	// Validate events
	if len(req.Events) == 0 {
		return c.Status(400).JSON(common.NewErrorResponse("At least one event type is required"))
	}

	if invalidEvents := domainWebhook.ValidateEvents(req.Events); len(invalidEvents) > 0 {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid event types: " + fmt.Sprintf("%v", invalidEvents)))
	}

	req.SessionID = &sessionID

	ctx := c.Context()
	result, err := h.webhookUC.SetConfig(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create webhook: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to create webhook"))
	}

	response := common.NewSuccessResponse(result, "Webhook configuration created successfully")
	return c.Status(201).JSON(response)
}

// @Summary Get webhook configuration
// @Description Get current webhook configuration for a WhatsApp session
// @Tags Webhooks
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID" format(uuid)
// @Success 200 {object} webhook.WebhookResponse "Webhook configuration retrieved successfully"
// @Failure 400 {object} object "Bad Request - Invalid session ID format"
// @Failure 404 {object} object "Webhook not found for this session"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/find [get]
func (h *WebhookHandler) FindConfig(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	h.logger.InfoWithFields("Getting webhook config", map[string]interface{}{
		"session_id": sessionID,
	})

	// Validate session ID format
	if _, err := uuid.Parse(sessionID); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid session ID format"))
	}

	ctx := c.Context()
	webhook, err := h.webhookUC.FindConfig(ctx, sessionID)
	if err != nil {
		h.logger.Error("Failed to get webhook config: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get webhook configuration"))
	}

	response := common.NewSuccessResponse(webhook, "Webhook configuration retrieved successfully")
	return c.JSON(response)
}

// @Summary Test webhook
// @Description Test webhook endpoint for a WhatsApp session with a sample event
// @Tags Webhooks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID" format(uuid)
// @Param request body webhook.TestWebhookRequest true "Test webhook request"
// @Success 200 {object} webhook.TestWebhookResponse "Webhook tested successfully"
// @Failure 400 {object} object "Bad Request - Invalid session ID or event type"
// @Failure 404 {object} object "Webhook not found for this session"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/test [post]
func (h *WebhookHandler) TestWebhook(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	h.logger.InfoWithFields("Testing webhook", map[string]interface{}{
		"session_id": sessionID,
	})

	// Validate session ID format
	if _, err := uuid.Parse(sessionID); err != nil {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid session ID format"))
	}

	var req webhook.TestWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse test webhook request: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	// Validate event type
	if req.EventType == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Event type is required"))
	}

	if !domainWebhook.IsValidEventType(req.EventType) {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid event type: " + req.EventType))
	}

	ctx := c.Context()

	// First find the webhook for this session
	existingWebhook, err := h.webhookUC.FindConfig(ctx, sessionID)
	if err != nil {
		h.logger.Error("Failed to find webhook for session: " + err.Error())
		return c.Status(404).JSON(common.NewErrorResponse("Webhook not found for this session"))
	}

	// Test the webhook
	result, err := h.webhookUC.TestWebhook(ctx, existingWebhook.ID, &req)
	if err != nil {
		h.logger.Error("Failed to test webhook: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to test webhook"))
	}

	response := common.NewSuccessResponse(result, "Webhook tested successfully")
	return c.JSON(response)
}

// @Summary Get supported webhook events
// @Description Get list of all supported webhook event types that can be subscribed to
// @Tags Webhooks
// @Produce json
// @Success 200 {object} webhook.WebhookEventsResponse "Supported events retrieved successfully"
// @Failure 500 {object} object "Internal Server Error"
// @Router /webhook/events [get]
func (h *WebhookHandler) GetSupportedEvents(c *fiber.Ctx) error {
	h.logger.Info("Getting supported webhook events")

	ctx := c.Context()
	result, err := h.webhookUC.GetSupportedWebhookEvents(ctx)
	if err != nil {
		h.logger.Error("Failed to get supported events: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get supported events"))
	}

	response := common.NewSuccessResponse(result, "Supported events retrieved successfully")
	return c.JSON(response)
}
