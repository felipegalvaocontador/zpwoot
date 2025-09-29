package handlers

import (
	"zpwoot/internal/app/common"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	logger        *logger.Logger
	wameowManager *wameow.Manager
}

func NewHealthHandler(logger *logger.Logger, wameowManager *wameow.Manager) *HealthHandler {
	return &HealthHandler{
		logger:        logger,
		wameowManager: wameowManager,
	}
}

// @Summary Health check
// @Description Check if the API is running and healthy
// @Tags Health
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} common.HealthResponse "API is healthy"
// @Failure 500 {object} object "Internal Server Error"
// @Router /health [get]
func (h *HealthHandler) GetHealth(c *fiber.Ctx) error {
	response := &common.HealthResponse{
		Status:  "ok",
		Service: "zpwoot",
	}
	return c.JSON(response)
}

// @Summary WhatsApp manager health check
// @Description Check if WhatsApp manager and whatsmeow tables are available
// @Tags Health
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} object "WhatsApp manager is healthy"
// @Failure 503 {object} object "Service Unavailable"
// @Router /health/wameow [get]
func (h *HealthHandler) GetWameowHealth(c *fiber.Ctx) error {
	if h.wameowManager == nil {
		return c.Status(503).JSON(fiber.Map{
			"status":  "error",
			"service": "wameow",
			"message": "WhatsApp manager not initialized",
		})
	}

	healthData := h.wameowManager.HealthCheck()
	healthData["service"] = "wameow"
	healthData["message"] = "WhatsApp manager is healthy and whatsmeow tables are available"

	return c.JSON(healthData)
}
