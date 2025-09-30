package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/community"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

// Error message constants
const (
	ErrSessionNotFound     = "session not found"
	ErrSessionNotConnected = "session is not connected"
	ErrValidationFailed    = "validation failed"
	ErrCommunityNotFound   = "community not found"
	ErrInvalidRequestData  = "Invalid request data"
)

// CommunityHandler handles community-related HTTP requests
type CommunityHandler struct {
	logger          *logger.Logger
	communityUC     community.UseCase
	sessionResolver *helpers.SessionResolver
}

// NewCommunityHandler creates a new community handler
func NewCommunityHandler(appLogger *logger.Logger, communityUC community.UseCase, sessionRepo helpers.SessionRepository) *CommunityHandler {
	return &CommunityHandler{
		logger:          appLogger,
		communityUC:     communityUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// resolveSession resolves session from URL parameter
func (h *CommunityHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
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

// handleGroupLinkAction handles common group link/unlink action logic
func (h *CommunityHandler) handleGroupLinkAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	parseFunc func(*http.Request) (interface{}, error),
	actionFunc func(context.Context, string, interface{}) (interface{}, error),
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

	req, err := parseFunc(r)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s group", actionName), map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := actionFunc(r.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s group", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})

		statusCode := 500
		if err.Error() == ErrSessionNotConnected {
			statusCode = 400
		} else if err.Error() == ErrCommunityNotFound {
			statusCode = 404
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to %s group", actionName)))
		return
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s group successfully", actionName), map[string]interface{}{
		"session_id": sess.ID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, fmt.Sprintf("Group %sed successfully", actionName)))
}

// handleCommunityQueryAction handles common community query action logic
func (h *CommunityHandler) handleCommunityQueryAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	paramName string,
	createRequestFunc func(string) interface{},
	actionFunc func(context.Context, string, interface{}) (interface{}, error),
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

	paramValue := r.URL.Query().Get(paramName)
	if paramValue == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("%s parameter is required", paramName)))
		return
	}

	req := createRequestFunc(paramValue)

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id": sess.ID.String(),
		paramName:    paramValue,
	})

	response, err := actionFunc(r.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			paramName:    paramValue,
			"error":      err.Error(),
		})

		statusCode := 500
		if err.Error() == ErrSessionNotConnected {
			statusCode = 400
		} else if err.Error() == ErrCommunityNotFound {
			statusCode = 404
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName)))
		return
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s successfully", actionName), map[string]interface{}{
		"session_id": sess.ID,
		paramName:    paramValue,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, fmt.Sprintf("%s completed successfully", actionName)))
}

// parseLinkGroupRequest parses link group request from HTTP request
func (h *CommunityHandler) parseLinkGroupRequest(r *http.Request) (interface{}, error) {
	var req community.LinkGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

// parseUnlinkGroupRequest parses unlink group request from HTTP request
func (h *CommunityHandler) parseUnlinkGroupRequest(r *http.Request) (interface{}, error) {
	var req community.UnlinkGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

// @Summary Link group to community
// @Description Link a group to a community
// @Tags Communities
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body community.LinkGroupRequest true "Link group request"
// @Success 200 {object} common.SuccessResponse{data=community.LinkGroupResponse} "Group linked successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/communities/link-group [post]
func (h *CommunityHandler) LinkGroup(w http.ResponseWriter, r *http.Request) {
	h.handleGroupLinkAction(
		w,
		r,
		"link",
		h.parseLinkGroupRequest,
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.communityUC.LinkGroup(ctx, sessionID, req.(*community.LinkGroupRequest))
		},
	)
}

// @Summary Unlink group from community
// @Description Unlink a group from a community
// @Tags Communities
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body community.UnlinkGroupRequest true "Unlink group request"
// @Success 200 {object} common.SuccessResponse{data=community.UnlinkGroupResponse} "Group unlinked successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/communities/unlink-group [post]
func (h *CommunityHandler) UnlinkGroup(w http.ResponseWriter, r *http.Request) {
	h.handleGroupLinkAction(
		w,
		r,
		"unlink",
		h.parseUnlinkGroupRequest,
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.communityUC.UnlinkGroup(ctx, sessionID, req.(*community.UnlinkGroupRequest))
		},
	)
}

// @Summary Get community information
// @Description Get detailed information about a community
// @Tags Communities
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param communityJid query string true "Community JID" example("120363025246125486@g.us")
// @Success 200 {object} common.SuccessResponse{data=community.GetCommunityInfoResponse} "Community information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/communities/info [get]
func (h *CommunityHandler) GetCommunityInfo(w http.ResponseWriter, r *http.Request) {
	h.handleCommunityQueryAction(
		w,
		r,
		"get community info",
		"communityJid",
		func(jid string) interface{} {
			return &community.GetCommunityInfoRequest{CommunityJID: jid}
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.communityUC.GetCommunityInfo(ctx, sessionID, req.(*community.GetCommunityInfoRequest))
		},
	)
}

// @Summary Get community sub-groups
// @Description Get all sub-groups (linked groups) of a community
// @Tags Communities
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param communityJid query string true "Community JID" example("120363025246125486@g.us")
// @Success 200 {object} common.SuccessResponse{data=community.GetSubGroupsResponse} "Sub-groups retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/communities/subgroups [get]
func (h *CommunityHandler) GetSubGroups(w http.ResponseWriter, r *http.Request) {
	h.handleCommunityQueryAction(
		w,
		r,
		"get community sub-groups",
		"communityJid",
		func(jid string) interface{} {
			return &community.GetSubGroupsRequest{CommunityJID: jid}
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.communityUC.GetSubGroups(ctx, sessionID, req.(*community.GetSubGroupsRequest))
		},
	)
}
