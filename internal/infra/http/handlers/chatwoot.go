package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/chatwoot"
	"zpwoot/internal/app/common"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type ChatwootHandler struct {
	chatwootUC      chatwoot.UseCase
	sessionResolver *helpers.SessionResolver
	logger          *logger.Logger
}

func NewChatwootHandler(
	chatwootUC chatwoot.UseCase,
	sessionRepo helpers.SessionRepository,
	logger *logger.Logger,
) *ChatwootHandler {
	sessionResolver := helpers.NewSessionResolver(logger, sessionRepo)

	return &ChatwootHandler{
		chatwootUC:      chatwootUC,
		sessionResolver: sessionResolver,
		logger:          logger,
	}
}

// resolveSession resolves session from URL parameter
func (h *ChatwootHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		return nil, errors.New("session identifier is required")
	}

	return h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
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

	var req chatwoot.CreateChatwootConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	// Validate required fields
	if req.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("URL is required"))
		return
	}

	if req.Token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Token is required"))
		return
	}

	if req.AccountID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Account ID is required"))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to create Chatwoot configuration"))
		return
	}

	response := common.NewSuccessResponse(result, "Chatwoot configuration created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	h.logger.InfoWithFields("Getting Chatwoot configuration", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.chatwootUC.GetConfig(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Chatwoot configuration not found"))
			return
		}

		h.logger.Error("Failed to get Chatwoot configuration: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get Chatwoot configuration"))
		return
	}

	response := common.NewSuccessResponse(result, "Chatwoot configuration retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	var req chatwoot.UpdateChatwootConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	h.logger.InfoWithFields("Updating Chatwoot configuration", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.chatwootUC.UpdateConfig(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Chatwoot configuration not found"))
			return
		}

		h.logger.Error("Failed to update Chatwoot configuration: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to update Chatwoot configuration"))
		return
	}

	response := common.NewSuccessResponse(result, "Chatwoot configuration updated successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	h.logger.InfoWithFields("Deleting Chatwoot configuration", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	err = h.chatwootUC.DeleteConfig(r.Context())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Chatwoot configuration not found"))
			return
		}

		h.logger.Error("Failed to delete Chatwoot configuration: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to delete Chatwoot configuration"))
		return
	}

	response := common.NewSuccessResponse(nil, "Chatwoot configuration deleted successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	h.logger.InfoWithFields("Testing Chatwoot connection", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.chatwootUC.TestConnection(r.Context())
	if err != nil {
		h.logger.Error("Failed to test Chatwoot connection: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to test Chatwoot connection"))
		return
	}

	response := common.NewSuccessResponse(result, "Connection test completed")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	h.logger.InfoWithFields("Getting Chatwoot statistics", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.chatwootUC.GetStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get Chatwoot statistics: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get Chatwoot statistics"))
		return
	}

	response := common.NewSuccessResponse(result, "Statistics retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	var req struct {
		InboxName  string `json:"inboxName"`
		WebhookURL string `json:"webhookURL"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.InboxName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Inbox name is required"))
		return
	}

	if req.WebhookURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Webhook URL is required"))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to auto-create Chatwoot inbox"))
		return
	}

	response := common.NewSuccessResponse(nil, "Inbox created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Session identifier is required"))
		return
	}

	// Resolve session
	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
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

	var payload chatwoot.ChatwootWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode webhook payload: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid webhook payload"))
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to process webhook"))
		return
	}

	response := common.NewSuccessResponse(nil, "Webhook processed successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
