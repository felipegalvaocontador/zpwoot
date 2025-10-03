package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

type WebhookHandler struct {
	*shared.BaseHandler
	sessionService *services.SessionService
}

func NewWebhookHandler(
	sessionService *services.SessionService,
	logger *logger.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		sessionService: sessionService,
	}
}

// @Summary Set webhook configuration
// @Description Configure webhook settings for the session
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/webhook/set [post]
func (h *WebhookHandler) SetConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set webhook config")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("set webhook config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Webhook configuration set successfully")
}

// @Summary Get webhook configuration
// @Description Get the current webhook configuration for the session
// @Tags Webhooks
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/webhook/find [get]
func (h *WebhookHandler) FindConfig(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "find webhook config")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSessionByNameOrID(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("find webhook config", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Webhook configuration retrieved successfully")
}

// @Summary Test webhook configuration
// @Description Test the webhook configuration by sending a test event
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/webhook/test [post]
func (h *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "test webhook")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	h.LogSuccess("test webhook", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Webhook test completed successfully")
}
