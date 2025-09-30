package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/newsletter"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type NewsletterHandler struct {
	logger          *logger.Logger
	newsletterUC    newsletter.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewNewsletterHandler(appLogger *logger.Logger, newsletterUC newsletter.UseCase, sessionRepo helpers.SessionRepository) *NewsletterHandler {
	return &NewsletterHandler{
		logger:          appLogger,
		newsletterUC:    newsletterUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// resolveSession resolves session from URL parameter
func (h *NewsletterHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
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

// handleNewsletterAction handles common newsletter action logic
func (h *NewsletterHandler) handleNewsletterAction(
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

// @Summary Create a new newsletter
// @Description Create a new WhatsApp newsletter/channel
// @Tags Newsletters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body newsletter.CreateNewsletterRequest true "Newsletter creation request"
// @Success 200 {object} common.SuccessResponse{data=newsletter.CreateNewsletterResponse} "Newsletter created successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/newsletters/create [post]
func (h *NewsletterHandler) CreateNewsletter(w http.ResponseWriter, r *http.Request) {
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

	var req newsletter.CreateNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.newsletterUC.CreateNewsletter(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to create newsletter: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to create newsletter"))
		return
	}

	response := common.NewSuccessResponse(result, "Newsletter created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get newsletter information
// @Description Get detailed information about a newsletter
// @Tags Newsletters
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param jid query string true "Newsletter JID" example("120363025246125486@newsletter")
// @Success 200 {object} common.SuccessResponse{data=newsletter.GetNewsletterInfoResponse} "Newsletter information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/newsletters/info [get]
func (h *NewsletterHandler) GetNewsletterInfo(w http.ResponseWriter, r *http.Request) {
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

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Newsletter JID is required"))
		return
	}

	h.logger.InfoWithFields("Getting newsletter info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
	})

	req := &newsletter.GetNewsletterInfoRequest{
		NewsletterJID: jid,
	}

	result, err := h.newsletterUC.GetNewsletterInfo(r.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.Error("Failed to get newsletter info: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get newsletter info"))
		return
	}

	response := common.NewSuccessResponse(result, "Newsletter information retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get newsletter info from invite
// @Description Get newsletter information using an invite key
// @Tags Newsletters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body newsletter.GetNewsletterInfoWithInviteRequest true "Newsletter invite request"
// @Success 200 {object} common.SuccessResponse{data=newsletter.GetNewsletterInfoResponse} "Newsletter information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/newsletters/info-from-invite [post]
func (h *NewsletterHandler) GetNewsletterInfoWithInvite(w http.ResponseWriter, r *http.Request) {
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

	var req newsletter.GetNewsletterInfoWithInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields("Getting newsletter info with invite", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.newsletterUC.GetNewsletterInfoWithInvite(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to get newsletter info with invite: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get newsletter info with invite"))
		return
	}

	response := common.NewSuccessResponse(result, "Newsletter information retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Follow newsletter
// @Description Follow (subscribe to) a newsletter
// @Tags Newsletters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body newsletter.FollowNewsletterRequest true "Follow newsletter request"
// @Success 200 {object} common.SuccessResponse{data=newsletter.NewsletterActionResponse} "Newsletter followed successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/newsletters/follow [post]
func (h *NewsletterHandler) FollowNewsletter(w http.ResponseWriter, r *http.Request) {
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

	var req newsletter.FollowNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields("Following newsletter", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.newsletterUC.FollowNewsletter(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to follow newsletter: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to follow newsletter"))
		return
	}

	response := common.NewSuccessResponse(result, "Newsletter followed successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Unfollow newsletter
// @Description Unfollow (unsubscribe from) a newsletter
// @Tags Newsletters
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body newsletter.UnfollowNewsletterRequest true "Unfollow newsletter request"
// @Success 200 {object} common.SuccessResponse{data=newsletter.NewsletterActionResponse} "Newsletter unfollowed successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/newsletters/unfollow [post]
func (h *NewsletterHandler) UnfollowNewsletter(w http.ResponseWriter, r *http.Request) {
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

	var req newsletter.UnfollowNewsletterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields("Unfollowing newsletter", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.newsletterUC.UnfollowNewsletter(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.Error("Failed to unfollow newsletter: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to unfollow newsletter"))
		return
	}

	response := common.NewSuccessResponse(result, "Newsletter unfollowed successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get subscribed newsletters
// @Description Get all newsletters the user is subscribed to
// @Tags Newsletters
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=newsletter.GetSubscribedNewslettersResponse} "Subscribed newsletters retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/newsletters [get]
func (h *NewsletterHandler) GetSubscribedNewsletters(w http.ResponseWriter, r *http.Request) {
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

	h.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := h.newsletterUC.GetSubscribedNewsletters(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to get subscribed newsletters: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get subscribed newsletters"))
		return
	}

	response := common.NewSuccessResponse(result, "Subscribed newsletters retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
