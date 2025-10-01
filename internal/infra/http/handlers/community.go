package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/community"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

const (
	ErrSessionNotFound     = "session not found"
	ErrSessionNotConnected = "session is not connected"
	ErrValidationFailed    = "validation failed"
	ErrCommunityNotFound   = "community not found"
	ErrInvalidRequestData  = "Invalid request data"
)

type CommunityHandler struct {
	*BaseHandler
	communityUC community.UseCase
}

func NewCommunityHandler(appLogger *logger.Logger, communityUC community.UseCase, sessionRepo helpers.SessionRepository) *CommunityHandler {
	sessionResolver := &SessionResolver{
		logger:      appLogger,
		sessionRepo: sessionRepo,
	}
	return &CommunityHandler{
		BaseHandler: NewBaseHandler(appLogger, sessionResolver),
		communityUC: communityUC,
	}
}

// resolveSession removido - usar h.resolveSessionFromURL(r) do BaseHandler

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
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	req, err := parseFunc(r)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
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
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to %s group", actionName))); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s group successfully", actionName), map[string]interface{}{
		"session_id": sess.ID,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(response, fmt.Sprintf("Group %sed successfully", actionName))); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}

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
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	paramValue := r.URL.Query().Get(paramName)
	if paramValue == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("%s parameter is required", paramName))); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
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
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName))); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s successfully", actionName), map[string]interface{}{
		"session_id": sess.ID,
		paramName:    paramValue,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(response, fmt.Sprintf("%s completed successfully", actionName))); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}

func (h *CommunityHandler) parseLinkGroupRequest(r *http.Request) (interface{}, error) {
	var req community.LinkGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

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
			linkReq, ok := req.(*community.LinkGroupRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			result, err := h.communityUC.LinkGroup(ctx, sessionID, linkReq)
			return result, err
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
			unlinkReq, ok := req.(*community.UnlinkGroupRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			result, err := h.communityUC.UnlinkGroup(ctx, sessionID, unlinkReq)
			return result, err
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
			infoReq, ok := req.(*community.GetCommunityInfoRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			result, err := h.communityUC.GetCommunityInfo(ctx, sessionID, infoReq)
			return result, err
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
			subGroupsReq, ok := req.(*community.GetSubGroupsRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			result, err := h.communityUC.GetSubGroups(ctx, sessionID, subGroupsReq)
			return result, err
		},
	)
}
