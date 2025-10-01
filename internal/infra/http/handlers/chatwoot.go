package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/chatwoot"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/platform/logger"
)

type ChatwootHandler struct {
	*BaseHandler
	chatwootUC chatwoot.UseCase
}

func NewChatwootHandler(
	chatwootUC chatwoot.UseCase,
	sessionRepo helpers.SessionRepository,
	logger *logger.Logger,
) *ChatwootHandler {
	return &ChatwootHandler{
		BaseHandler: NewBaseHandler(logger, sessionRepo),
		chatwootUC:  chatwootUC,
	}
}

func (h *ChatwootHandler) handleChatwootAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	actionFunc func(context.Context) (interface{}, error),
) {
	sess, err := h.GetSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(r.Context())
	if err != nil {
		h.logger.Error("Failed to " + actionName + ": " + err.Error())
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to "+actionName)
		return
	}

	h.writeSuccessResponse(w, result, successMessage)
}

// @Summary Create Chatwoot configuration
// @Description Create a new Chatwoot configuration for a session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body chatwoot.CreateChatwootConfigRequest true "Chatwoot configuration request"
// @Success 200 {object} common.SuccessResponse{data=chatwoot.CreateChatwootConfigResponse} "Chatwoot configuration created successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/set [post]
func (h *ChatwootHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	sess, err := h.GetSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	var req chatwoot.CreateChatwootConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.URL == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "URL is required")
		return
	}

	if req.Token == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Token is required")
		return
	}

	if req.AccountID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Account ID is required")
		return
	}

	h.logger.InfoWithFields("Creating Chatwoot configuration", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"url":          req.URL,
		"account_id":   req.AccountID,
	})

	result, err := h.chatwootUC.CreateConfig(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to create Chatwoot configuration: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create Chatwoot configuration")
		return
	}

	h.writeSuccessResponse(w, result, "Chatwoot configuration created successfully")
}

// @Summary Get Chatwoot configuration
// @Description Get the current Chatwoot configuration for a session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=chatwoot.ChatwootConfigResponse} "Chatwoot configuration retrieved successfully"
// @Failure 404 {object} object "Session or configuration not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot [get]
func (h *ChatwootHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	h.handleChatwootAction(w, r, "Getting Chatwoot configuration", "Chatwoot configuration retrieved successfully",
		func(ctx context.Context) (interface{}, error) {
			result, err := h.chatwootUC.GetConfig(ctx)
			if err != nil && strings.Contains(err.Error(), "not found") {
				return nil, errors.New("chatwoot configuration not found")
			}
			return result, err
		})
}

// @Summary Update Chatwoot configuration
// @Description Update an existing Chatwoot configuration for a session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body chatwoot.UpdateChatwootConfigRequest true "Chatwoot configuration update request"
// @Success 200 {object} common.SuccessResponse{data=chatwoot.ChatwootConfigResponse} "Chatwoot configuration updated successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or configuration not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot [put]
func (h *ChatwootHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	sess, err := h.GetSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	var req chatwoot.UpdateChatwootConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	h.logger.InfoWithFields("Updating Chatwoot configuration", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.chatwootUC.UpdateConfig(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "Chatwoot configuration not found")
			return
		}

		h.logger.Error("Failed to update Chatwoot configuration: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update Chatwoot configuration")
		return
	}

	h.writeSuccessResponse(w, result, "Chatwoot configuration updated successfully")
}

// @Summary Delete Chatwoot configuration
// @Description Delete the Chatwoot configuration for a session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse "Chatwoot configuration deleted successfully"
// @Failure 404 {object} object "Session or configuration not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot [delete]
func (h *ChatwootHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	sess, err := h.GetSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	h.logger.InfoWithFields("Deleting Chatwoot configuration", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	err = h.chatwootUC.DeleteConfig(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "Chatwoot configuration not found")
			return
		}

		h.logger.Error("Failed to delete Chatwoot configuration: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete Chatwoot configuration")
		return
	}

	h.writeSuccessResponse(w, nil, "Chatwoot configuration deleted successfully")
}

// @Summary Test Chatwoot connection
// @Description Test the connection to Chatwoot instance
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=chatwoot.TestChatwootConnectionResponse} "Connection test completed"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/test [post]
func (h *ChatwootHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	h.handleChatwootAction(w, r, "Testing Chatwoot connection", "Connection test completed",
		func(ctx context.Context) (interface{}, error) {
			return h.chatwootUC.TestConnection(ctx)
		})
}

// @Summary Get Chatwoot statistics
// @Description Get statistics for Chatwoot integration
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=chatwoot.ChatwootStatsResponse} "Statistics retrieved successfully"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/stats [get]
func (h *ChatwootHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	h.handleChatwootAction(w, r, "Getting Chatwoot statistics", "Statistics retrieved successfully",
		func(ctx context.Context) (interface{}, error) {
			return h.chatwootUC.GetStats(ctx)
		})
}

// @Summary Auto-create Chatwoot inbox
// @Description Automatically create an inbox in Chatwoot for the session
// @Tags Chatwoot
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body object{inboxName=string,webhookURL=string} true "Auto-create inbox request"
// @Success 200 {object} common.SuccessResponse "Inbox created successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/chatwoot/auto-create-inbox [post]
func (h *ChatwootHandler) AutoCreateInbox(w http.ResponseWriter, r *http.Request) {
	sess, err := h.GetSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	var req struct {
		InboxName  string `json:"inboxName"`
		WebhookURL string `json:"webhookURL"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.InboxName == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Inbox name is required")
		return
	}

	if req.WebhookURL == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Webhook URL is required")
		return
	}

	h.logger.InfoWithFields("Auto-creating Chatwoot inbox", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"inbox_name":   req.InboxName,
		"webhook_url":  req.WebhookURL,
	})

	err = h.chatwootUC.AutoCreateInbox(r.Context(), sess.ID.String(), req.InboxName, req.WebhookURL)
	if err != nil {
		h.logger.Error("Failed to auto-create Chatwoot inbox: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to auto-create Chatwoot inbox")
		return
	}

	h.writeSuccessResponse(w, nil, "Inbox created successfully")
}

// @Summary Receive Chatwoot webhook
// @Description Receive and process webhook events from Chatwoot
// @Tags Chatwoot
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param payload body chatwoot.ChatwootWebhookPayload true "Webhook payload from Chatwoot"
// @Success 200 {object} common.SuccessResponse "Webhook processed successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /chatwoot/webhook/{sessionId} [post]
func (h *ChatwootHandler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Session identifier is required")
		return
	}

	// Resolve sessão por ID ou nome
	sess, err := h.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	var payload chatwoot.ChatwootWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode webhook payload: " + err.Error())
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid webhook payload")
		return
	}

	h.logger.InfoWithFields("Received Chatwoot webhook", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"event":        payload.Event,
	})

	err = h.chatwootUC.ProcessWebhook(r.Context(), sess.ID.String(), &payload)
	if err != nil {
		h.logger.Error("Failed to process Chatwoot webhook: " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to process webhook")
		return
	}

	h.writeSuccessResponse(w, nil, "Webhook processed successfully")
}
