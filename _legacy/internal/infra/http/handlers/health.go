package handlers

import (
	"encoding/json"
	"net/http"

	"zpwoot/internal/app/common"
	"zpwoot/internal/infra/wameow"
	"zpwoot/platform/logger"
)

type HealthHandler struct {
	logger        *logger.Logger
	wameowManager *wameow.Manager
}

func NewHealthHandler(log *logger.Logger, wameowManager *wameow.Manager) *HealthHandler {
	return &HealthHandler{
		logger:        log,
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
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	response := &common.HealthResponse{
		Status:  "ok",
		Service: "zpwoot",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health response: " + err.Error())
	}
}

// @Summary WhatsApp manager health check
// @Description Check if WhatsApp manager and whatsmeow tables are available
// @Tags Health
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} object "WhatsApp manager is healthy"
// @Failure 503 {object} object "Service Unavailable"
// @Router /health/wameow [get]
func (h *HealthHandler) GetWameowHealth(w http.ResponseWriter, r *http.Request) {
	if h.wameowManager == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"service": "wameow",
			"message": "WhatsApp manager not initialized",
		}); err != nil {
			h.logger.Error("Failed to encode wameow error response: " + err.Error())
		}
		return
	}

	healthData := h.wameowManager.HealthCheck()
	healthData["service"] = "wameow"
	healthData["message"] = "WhatsApp manager is healthy and whatsmeow tables are available"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(healthData); err != nil {
		h.logger.Error("Failed to encode wameow health response: " + err.Error())
	}
}
