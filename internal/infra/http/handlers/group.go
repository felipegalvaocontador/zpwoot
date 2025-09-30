package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/group"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type GroupHandler struct {
	*BaseHandler
	groupUC group.UseCase
}

func NewGroupHandler(appLogger *logger.Logger, groupUC group.UseCase, sessionRepo helpers.SessionRepository) *GroupHandler {
	sessionResolver := &SessionResolver{
		logger:      appLogger,
		sessionRepo: sessionRepo,
	}
	return &GroupHandler{
		BaseHandler: NewBaseHandler(appLogger, sessionResolver),
		groupUC:     groupUC,
	}
}

func (h *GroupHandler) handleGroupActionWithValidation(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	fieldName string,
	fieldValidationMessage string,
	responseBuilder func(string) map[string]interface{},
) {
	sess, err := h.resolveSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	var reqData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	fieldValue := reqData[fieldName]
	if fieldValue == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, fieldValidationMessage)
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		fieldName:      fieldValue,
	})

	response := responseBuilder(fieldValue)
	h.writeSuccessResponse(w, response, successMessage)
}

func (h *GroupHandler) handleGroupActionWithTwoFields(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	field1Name, field1ValidationMessage string,
	field2Name, field2ValidationMessage string,
	responseBuilder func(string, string) map[string]interface{},
) {
	sess, err := h.resolveSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	var reqData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	field1Value := reqData[field1Name]
	if field1Value == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, field1ValidationMessage)
		return
	}

	field2Value := reqData[field2Name]
	if field2Value == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, field2ValidationMessage)
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		field1Name:     field1Value,
		field2Name:     field2Value,
	})

	response := responseBuilder(field1Value, field2Value)
	h.writeSuccessResponse(w, response, successMessage)
}

func (h *GroupHandler) GetGroupInviteLink(w http.ResponseWriter, r *http.Request) {
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

	groupJid := r.URL.Query().Get("groupJid")
	if groupJid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required"))
		return
	}

	h.logger.InfoWithFields("Getting group invite link", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"group_jid":    groupJid,
	})

	response := map[string]interface{}{
		"groupJid":   groupJid,
		"inviteLink": "https://chat.whatsapp.com/placeholder",
		"message":    "Group invite link functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group invite link retrieved successfully"))
}

func (h *GroupHandler) JoinGroupViaLink(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithValidation(
		w,
		r,
		"Joining group via link",
		"Join group request processed",
		"inviteLink",
		"Invite link is required",
		func(inviteLink string) map[string]interface{} {
			return map[string]interface{}{
				"inviteLink": inviteLink,
				"status":     "pending",
				"message":    "Join group via link functionality needs to be implemented in use case",
			}
		},
	)
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithValidation(
		w,
		r,
		"Leaving group",
		"Left group successfully",
		"groupJid",
		"Group JID is required",
		func(groupJid string) map[string]interface{} {
			return map[string]interface{}{
				"groupJid": groupJid,
				"status":   "left",
				"message":  "Leave group functionality needs to be implemented in use case",
			}
		},
	)
}

func (h *GroupHandler) UpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
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
		GroupJID string `json:"groupJid"`
		Announce *bool  `json:"announce,omitempty"`
		Locked   *bool  `json:"locked,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.GroupJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required"))
		return
	}

	h.logger.InfoWithFields("Updating group settings", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"group_jid":    req.GroupJID,
		"announce":     req.Announce,
		"locked":       req.Locked,
	})

	response := map[string]interface{}{
		"groupJid": req.GroupJID,
		"announce": req.Announce,
		"locked":   req.Locked,
		"status":   "updated",
		"message":  "Update group settings functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group settings updated successfully"))
}

func (h *GroupHandler) GetGroupRequestParticipants(w http.ResponseWriter, r *http.Request) {
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

	groupJid := r.URL.Query().Get("groupJid")
	if groupJid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required"))
		return
	}

	h.logger.InfoWithFields("Getting group request participants", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"group_jid":    groupJid,
	})

	response := map[string]interface{}{
		"groupJid":     groupJid,
		"participants": []interface{}{},
		"message":      "Get group request participants functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group request participants retrieved successfully"))
}

func (h *GroupHandler) UpdateGroupRequestParticipants(w http.ResponseWriter, r *http.Request) {
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
		GroupJID     string   `json:"groupJid"`
		Action       string   `json:"action"`
		Participants []string `json:"participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.GroupJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required"))
		return
	}

	if req.Action == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Action is required (approve or reject)"))
		return
	}

	if len(req.Participants) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("At least one participant is required"))
		return
	}

	h.logger.InfoWithFields("Updating group request participants", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	response := map[string]interface{}{
		"groupJid":     req.GroupJID,
		"action":       req.Action,
		"participants": req.Participants,
		"status":       "processed",
		"message":      "Update group request participants functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group request participants updated successfully"))
}

func (h *GroupHandler) SetGroupJoinApprovalMode(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithTwoFields(
		w,
		r,
		"Setting group join approval mode",
		"Group join approval mode updated successfully",
		"groupJid",
		"Group JID is required",
		"approvalMode",
		"Approval mode is required (on or off)",
		func(groupJid, approvalMode string) map[string]interface{} {
			return map[string]interface{}{
				"groupJid":     groupJid,
				"approvalMode": approvalMode,
				"status":       "updated",
				"message":      "Set group join approval mode functionality needs to be implemented in use case",
			}
		},
	)
}

func (h *GroupHandler) SetGroupMemberAddMode(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithTwoFields(
		w,
		r,
		"Setting group member add mode",
		"Group member add mode updated successfully",
		"groupJid",
		"Group JID is required",
		"addMode",
		"Add mode is required (all_members or only_admins)",
		func(groupJid, addMode string) map[string]interface{} {
			return map[string]interface{}{
				"groupJid": groupJid,
				"addMode":  addMode,
				"status":   "updated",
				"message":  "Set group member add mode functionality needs to be implemented in use case",
			}
		},
	)
}

func (h *GroupHandler) GetGroupInfoFromLink(w http.ResponseWriter, r *http.Request) {
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

	inviteLink := r.URL.Query().Get("inviteLink")
	if inviteLink == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invite link is required"))
		return
	}

	h.logger.InfoWithFields("Getting group info from link", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"invite_link":  inviteLink,
	})

	response := map[string]interface{}{
		"inviteLink": inviteLink,
		"groupInfo": map[string]interface{}{
			"name":         "Sample Group",
			"description":  "Group info from link functionality needs to be implemented",
			"participants": 0,
		},
		"message": "Get group info from link functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group info retrieved from link successfully"))
}

func (h *GroupHandler) GetGroupInfoFromInvite(w http.ResponseWriter, r *http.Request) {
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
		GroupJID string `json:"groupJid"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.GroupJID == "" || req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID and code are required"))
		return
	}

	h.logger.InfoWithFields("Getting group info from invite", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"group_jid":    req.GroupJID,
		"code":         req.Code,
	})

	response := map[string]interface{}{
		"groupJid": req.GroupJID,
		"code":     req.Code,
		"groupInfo": map[string]interface{}{
			"name":         "Sample Group",
			"description":  "Group info from invite functionality needs to be implemented",
			"participants": 0,
		},
		"message": "Get group info from invite functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group info retrieved from invite successfully"))
}

func (h *GroupHandler) JoinGroupWithInvite(w http.ResponseWriter, r *http.Request) {
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
		GroupJID string `json:"groupJid"`
		Code     string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if req.GroupJID == "" || req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID and code are required"))
		return
	}

	h.logger.InfoWithFields("Joining group with invite", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"group_jid":    req.GroupJID,
		"code":         req.Code,
	})

	response := map[string]interface{}{
		"groupJid": req.GroupJID,
		"code":     req.Code,
		"status":   "joined",
		"message":  "Join group with invite functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Joined group with invite successfully"))
}

func (h *GroupHandler) handleGroupActionWithJID(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	parseFunc func(*http.Request) (interface{}, error),
	actionFunc func(context.Context, string, interface{}) (interface{}, error),
	extractJID func(interface{}) string,
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
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	groupJID := extractJID(req)
	if groupJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required in request body"))
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
	})

	response, err := actionFunc(r.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Create a new WhatsApp group
// @Description Create a new WhatsApp group with specified participants
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body group.CreateGroupRequest true "Group creation request"
// @Success 200 {object} common.SuccessResponse{data=group.CreateGroupResponse} "Group created successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups [post]
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
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

	var req group.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"name":         req.Name,
		"participants": len(req.Participants),
	})

	response, err := h.groupUC.CreateGroup(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group created successfully"))
}

// @Summary Get group information
// @Description Get detailed information about a specific group
// @Tags Groups
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param groupJid query string true "Group JID" example("120363025246125486@g.us")
// @Success 200 {object} common.SuccessResponse{data=group.GetGroupInfoResponse} "Group information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups/info [get]
func (h *GroupHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
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

	groupJID := r.URL.Query().Get("groupJid")
	if groupJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required as query parameter: ?groupJid=..."))
		return
	}

	req := &group.GetGroupInfoRequest{
		GroupJID: groupJID,
	}

	h.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
	})

	response, err := h.groupUC.GetGroupInfo(r.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group info", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Group information retrieved successfully"))
}

// @Summary List groups
// @Description List all groups the user is a member of
// @Tags Groups
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=group.ListGroupsResponse} "Groups listed successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups [get]
func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
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

	h.logger.InfoWithFields("Listing groups", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.groupUC.ListGroups(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.ErrorWithFields("Failed to list groups", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Groups listed successfully"))
}

// @Summary Update group participants
// @Description Add, remove, promote, or demote group participants
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body group.UpdateGroupParticipantsRequest true "Update participants request"
// @Success 200 {object} common.SuccessResponse{data=group.UpdateGroupParticipantsResponse} "Participants updated successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups/participants [post]
func (h *GroupHandler) UpdateGroupParticipants(w http.ResponseWriter, r *http.Request) {
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

	var req group.UpdateGroupParticipantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	if req.GroupJID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Group JID is required"))
		return
	}

	h.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	response, err := h.groupUC.UpdateGroupParticipants(r.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Participants updated successfully"))
}

// @Summary Set group name
// @Description Update the group name
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body group.SetGroupNameRequest true "Set group name request"
// @Success 200 {object} common.SuccessResponse{data=group.SetGroupNameResponse} "Group name updated successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups/name [put]
func (h *GroupHandler) SetGroupName(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithJID(
		w,
		r,
		"Setting group name",
		func(r *http.Request) (interface{}, error) {
			var req group.SetGroupNameRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.groupUC.SetGroupName(ctx, sessionID, req.(*group.SetGroupNameRequest))
		},
		func(req interface{}) string {
			return req.(*group.SetGroupNameRequest).GroupJID
		},
	)
}

// @Summary Set group description
// @Description Update the group description
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body group.SetGroupDescriptionRequest true "Set group description request"
// @Success 200 {object} common.SuccessResponse{data=group.SetGroupDescriptionResponse} "Group description updated successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups/description [put]
func (h *GroupHandler) SetGroupDescription(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithJID(
		w,
		r,
		"Setting group description",
		func(r *http.Request) (interface{}, error) {
			var req group.SetGroupDescriptionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.groupUC.SetGroupDescription(ctx, sessionID, req.(*group.SetGroupDescriptionRequest))
		},
		func(req interface{}) string {
			return req.(*group.SetGroupDescriptionRequest).GroupJID
		},
	)
}

// @Summary Set group photo
// @Description Update the group photo
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body group.SetGroupPhotoRequest true "Set group photo request"
// @Success 200 {object} common.SuccessResponse{data=group.SetGroupPhotoResponse} "Group photo updated successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/groups/photo [put]
func (h *GroupHandler) SetGroupPhoto(w http.ResponseWriter, r *http.Request) {
	h.handleGroupActionWithJID(
		w,
		r,
		"Setting group photo",
		func(r *http.Request) (interface{}, error) {
			var req group.SetGroupPhotoRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			return h.groupUC.SetGroupPhoto(ctx, sessionID, req.(*group.SetGroupPhotoRequest))
		},
		func(req interface{}) string {
			return req.(*group.SetGroupPhotoRequest).GroupJID
		},
	)
}
