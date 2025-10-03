package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

type GroupHandler struct {
	*shared.BaseHandler
	groupService   *services.GroupService
	sessionService *services.SessionService
}

func NewGroupHandler(
	groupService *services.GroupService,
	sessionService *services.SessionService,
	logger *logger.Logger,
) *GroupHandler {
	return &GroupHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		groupService:   groupService,
		sessionService: sessionService,
	}
}

// @Summary Create new WhatsApp group
// @Description Create a new WhatsApp group with specified participants
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.CreateGroupRequest true "Group creation request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.CreateGroupResponse}
// @Failure 400 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /sessions/{sessionId}/groups [post]
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "create group")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.CreateGroupRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	response, err := h.groupService.CreateGroup(r.Context(), sessionID, &req)
	if err != nil {
		h.HandleError(w, err, "create group")
		return
	}

	h.LogSuccess("create group", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    response.GroupJID,
		"group_name":   response.Name,
		"participants": len(response.Participants),
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary List WhatsApp groups
// @Description List all WhatsApp groups for a session
// @Tags Groups
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.ListGroupsResponse}
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /sessions/{sessionId}/groups [get]
func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list groups")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	response, err := h.groupService.ListGroups(r.Context(), sessionID)
	if err != nil {
		h.HandleError(w, err, "list groups")
		return
	}

	h.LogSuccess("list groups", map[string]interface{}{
		"session_id":  sessionID,
		"group_count": response.Count,
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary Get group information
// @Description Get detailed information about a WhatsApp group
// @Tags Groups
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param groupJid query string true "Group JID"
// @Success 200 {object} shared.SuccessResponse{data=contracts.GetGroupInfoResponse}
// @Failure 400 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /sessions/{sessionId}/groups/info [get]
func (h *GroupHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group info")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	groupJID := r.URL.Query().Get("groupJid")
	if groupJID == "" {
		h.GetWriter().WriteBadRequest(w, "Group JID is required")
		return
	}

	response, err := h.groupService.GetGroupInfo(r.Context(), sessionID, groupJID)
	if err != nil {
		h.HandleError(w, err, "get group info")
		return
	}

	h.LogSuccess("get group info", map[string]interface{}{
		"session_id":        sessionID,
		"group_jid":         groupJID,
		"group_name":        response.Name,
		"participant_count": len(response.Participants),
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary Update group participants
// @Description Add, remove, promote or demote group participants
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.UpdateParticipantsRequest true "Participants update request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.UpdateParticipantsResponse}
// @Failure 400 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /sessions/{sessionId}/groups/participants [post]
func (h *GroupHandler) UpdateGroupParticipants(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update group participants")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.UpdateParticipantsRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	response, err := h.groupService.UpdateGroupParticipants(r.Context(), sessionID, &req)
	if err != nil {
		h.HandleError(w, err, "update group participants")
		return
	}

	h.LogSuccess("update group participants", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary Set group name
// @Description Change the name of a WhatsApp group
// @Tags Groups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body contracts.SetGroupNameRequest true "Group name request"
// @Success 200 {object} shared.SuccessResponse{data=contracts.SetGroupNameResponse}
// @Failure 400 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /sessions/{sessionId}/groups/name [put]
func (h *GroupHandler) SetGroupName(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group name")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.SetGroupNameRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	response, err := h.groupService.SetGroupName(r.Context(), sessionID, &req)
	if err != nil {
		h.HandleError(w, err, "set group name")
		return
	}

	h.LogSuccess("set group name", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"new_name":   req.Name,
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

func (h *GroupHandler) SetGroupDescription(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group description")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group description not implemented yet")
}

func (h *GroupHandler) SetGroupPhoto(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group photo")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group photo not implemented yet")
}

func (h *GroupHandler) GetGroupInviteLink(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group invite link")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group invite link not implemented yet")
}

func (h *GroupHandler) JoinGroupViaLink(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "join group via link")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Join group via link not implemented yet")
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "leave group")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Leave group not implemented yet")
}

func (h *GroupHandler) UpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update group settings")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Update group settings not implemented yet")
}

func (h *GroupHandler) GetGroupRequestParticipants(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group request participants")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group request participants not implemented yet")
}

func (h *GroupHandler) UpdateGroupRequestParticipants(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update group request participants")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Update group request participants not implemented yet")
}

func (h *GroupHandler) SetGroupJoinApprovalMode(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group join approval mode")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group join approval mode not implemented yet")
}

func (h *GroupHandler) SetGroupMemberAddMode(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group member add mode")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group member add mode not implemented yet")
}

func (h *GroupHandler) GetGroupInfoFromLink(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group info from link")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group info from link not implemented yet")
}

func (h *GroupHandler) GetGroupInfoFromInvite(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group info from invite")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group info from invite not implemented yet")
}

func (h *GroupHandler) JoinGroupWithInvite(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "join group with invite")

	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Join group with invite not implemented yet")
}
