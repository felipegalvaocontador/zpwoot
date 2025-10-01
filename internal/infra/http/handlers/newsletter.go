package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/newsletter"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type NewsletterHandler struct {
	*BaseHandler
	newsletterUC newsletter.UseCase
}

func NewNewsletterHandler(appLogger *logger.Logger, newsletterUC newsletter.UseCase, sessionRepo helpers.SessionRepository) *NewsletterHandler {
	sessionResolver := &SessionResolver{
		logger:      appLogger,
		sessionRepo: sessionRepo,
	}
	return &NewsletterHandler{
		BaseHandler:  NewBaseHandler(appLogger, sessionResolver),
		newsletterUC: newsletterUC,
	}
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
	h.handleNewsletterAction(
		w,
		r,
		"Creating newsletter",
		"Newsletter created successfully",
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.newsletterUC.CreateNewsletter(ctx, sessionID, req.(*newsletter.CreateNewsletterRequest))
		},
		func() interface{} {
			return &newsletter.CreateNewsletterRequest{}
		},
	)
}

func (h *NewsletterHandler) handleNewsletterAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	actionFunc func(context.Context, string, interface{}) (interface{}, error),
	requestFactory func() interface{},
) {
	sess, err := h.resolveSessionFromURL(r)
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

	req := requestFactory()
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
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

	result, err := actionFunc(r.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.Error("Failed to " + actionName + ": " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to " + actionName)
		return
	}

	response := common.NewSuccessResponse(result, successMessage)
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
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get newsletter info")
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
	h.handleNewsletterAction(
		w,
		r,
		"Getting newsletter info with invite",
		"Newsletter information retrieved successfully",
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.newsletterUC.GetNewsletterInfoWithInvite(ctx, sessionID, req.(*newsletter.GetNewsletterInfoWithInviteRequest))
		},
		func() interface{} {
			return &newsletter.GetNewsletterInfoWithInviteRequest{}
		},
	)
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
	h.handleNewsletterAction(
		w,
		r,
		"Following newsletter",
		"Newsletter followed successfully",
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.newsletterUC.FollowNewsletter(ctx, sessionID, req.(*newsletter.FollowNewsletterRequest))
		},
		func() interface{} {
			return &newsletter.FollowNewsletterRequest{}
		},
	)
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
	h.handleNewsletterAction(
		w,
		r,
		"Unfollowing newsletter",
		"Newsletter unfollowed successfully",
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.newsletterUC.UnfollowNewsletter(ctx, sessionID, req.(*newsletter.UnfollowNewsletterRequest))
		},
		func() interface{} {
			return &newsletter.UnfollowNewsletterRequest{}
		},
	)
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
	h.handleSimpleGetRequest(
		w,
		r,
		"Getting subscribed newsletters",
		"Subscribed newsletters retrieved successfully",
		func(ctx context.Context, sessionID string) (interface{}, error) {
			return h.newsletterUC.GetSubscribedNewsletters(ctx, sessionID)
		},
	)
}
