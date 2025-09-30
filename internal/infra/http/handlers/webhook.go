package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/webhook"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type WebhookHandler struct {
	*BaseHandler
	webhookUC webhook.UseCase
}

func NewWebhookHandler(appLogger *logger.Logger, webhookUC webhook.UseCase, sessionRepo helpers.SessionRepository) *WebhookHandler {
	sessionResolver := &SessionResolver{
		logger:      appLogger,
		sessionRepo: sessionRepo,
	}
	return &WebhookHandler{
		BaseHandler: NewBaseHandler(appLogger, sessionResolver),
		webhookUC:   webhookUC,
	}
}

// resolveSession resolves session from URL parameter
func (h *WebhookHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
	idOrName := chi.URLParam(r, "sessionId")

	sess, err := h.sessionResolver.ResolveSession(r.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       r.URL.Path,
		})

		return nil, err
	}

	return sess, nil
}

// handleWebhookAction handles common webhook action logic
func (h *WebhookHandler) handleWebhookAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	parseFunc func(*http.Request, *domainSession.Session) (interface{}, error),
	actionFunc func(context.Context, interface{}) (interface{}, error),
) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	req, err := parseFunc(r, sess)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(r.Context(), req)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName)))
		return
	}

	response := common.NewSuccessResponse(result, successMessage)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Set webhook configuration
// @Description Set or update webhook configuration for a WhatsApp session
// @Tags Webhooks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body webhook.SetConfigRequest true "Webhook configuration request"
// @Success 200 {object} common.SuccessResponse{data=webhook.SetConfigResponse} "Webhook configuration set successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/set [post]
func (h *WebhookHandler) SetConfig(w http.ResponseWriter, r *http.Request) {
	h.handleWebhookAction(
		w,
		r,
		"Setting webhook configuration",
		"Webhook configuration set successfully",
		func(r *http.Request, sess *domainSession.Session) (interface{}, error) {
			var req webhook.SetConfigRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			sessionID := sess.ID.String()
			req.SessionID = &sessionID
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return h.webhookUC.SetConfig(ctx, req.(*webhook.SetConfigRequest))
		},
	)
}

// @Summary Find webhook configuration
// @Description Get webhook configuration for a WhatsApp session
// @Tags Webhooks
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=webhook.WebhookResponse} "Webhook configuration retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/find [get]
func (h *WebhookHandler) FindConfig(w http.ResponseWriter, r *http.Request) {
	h.handleSimpleGetRequest(
		w,
		r,
		"Finding webhook configuration",
		"Webhook configuration retrieved successfully",
		func(ctx context.Context, sessionID string) (interface{}, error) {
			return h.webhookUC.FindConfig(ctx, sessionID)
		},
	)
}

// @Summary Test webhook
// @Description Test webhook endpoint with a sample event
// @Tags Webhooks
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body webhook.TestWebhookRequest true "Test webhook request"
// @Success 200 {object} common.SuccessResponse{data=webhook.TestWebhookResponse} "Webhook tested successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/webhook/test [post]
func (h *WebhookHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req webhook.TestWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields("Testing webhook", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"event_type":   req.EventType,
	})

	// For testing, we need to find the webhook first to get its ID
	webhookConfig, err := h.webhookUC.FindConfig(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to find webhook configuration for testing: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Webhook configuration not found"))
		return
	}

	result, err := h.webhookUC.TestWebhook(r.Context(), webhookConfig.ID, &req)
	if err != nil {
		h.logger.Error("Failed to test webhook: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to test webhook"))
		return
	}

	response := common.NewSuccessResponse(result, "Webhook tested successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get supported webhook events
// @Description Get list of all supported webhook event types
// @Tags Webhooks
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} common.SuccessResponse{data=webhook.WebhookEventsResponse} "Supported events retrieved successfully"
// @Failure 500 {object} object "Internal Server Error"
// @Router /webhook/events [get]
func (h *WebhookHandler) GetSupportedEvents(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting supported webhook events")

	result, err := h.webhookUC.GetSupportedWebhookEvents(r.Context())
	if err != nil {
		h.logger.Error("Failed to get supported webhook events: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get supported webhook events"))
		return
	}

	response := common.NewSuccessResponse(result, "Supported events retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
