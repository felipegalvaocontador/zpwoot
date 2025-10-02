package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/platform/logger"
)

// GroupHandler gerencia endpoints relacionados a grupos WhatsApp
type GroupHandler struct {
	*shared.BaseHandler
	groupService   *services.GroupService
	sessionService *services.SessionService
}

// NewGroupHandler cria nova instância do group handler
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

// CreateGroup cria um novo grupo WhatsApp
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

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse e validar request
	var req contracts.CreateGroupRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	// Executar no service
	response, err := h.groupService.CreateGroup(r.Context(), sessionID, &req)
	if err != nil {
		h.HandleError(w, err, "create group")
		return
	}

	// Log sucesso
	h.LogSuccess("create group", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    response.GroupJID,
		"group_name":   response.Name,
		"participants": len(response.Participants),
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// ListGroups lista todos os grupos de uma sessão
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

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Executar no service
	response, err := h.groupService.ListGroups(r.Context(), sessionID)
	if err != nil {
		h.HandleError(w, err, "list groups")
		return
	}

	// Log sucesso
	h.LogSuccess("list groups", map[string]interface{}{
		"session_id":   sessionID,
		"group_count":  response.Count,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// GetGroupInfo obtém informações detalhadas de um grupo
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

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	groupJID := r.URL.Query().Get("groupJid")
	if groupJID == "" {
		h.GetWriter().WriteBadRequest(w, "Group JID is required")
		return
	}

	// Executar no service
	response, err := h.groupService.GetGroupInfo(r.Context(), sessionID, groupJID)
	if err != nil {
		h.HandleError(w, err, "get group info")
		return
	}

	// Log sucesso
	h.LogSuccess("get group info", map[string]interface{}{
		"session_id":       sessionID,
		"group_jid":        groupJID,
		"group_name":       response.Name,
		"participant_count": len(response.Participants),
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// UpdateGroupParticipants gerencia participantes do grupo
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

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse e validar request
	var req contracts.UpdateParticipantsRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	// Executar no service
	response, err := h.groupService.UpdateGroupParticipants(r.Context(), sessionID, &req)
	if err != nil {
		h.HandleError(w, err, "update group participants")
		return
	}

	// Log sucesso
	h.LogSuccess("update group participants", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// SetGroupName altera o nome do grupo
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

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse e validar request
	var req contracts.SetGroupNameRequest
	if err := h.ParseAndValidateJSON(r, &req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request format", err.Error())
		return
	}

	// Executar no service
	response, err := h.groupService.SetGroupName(r.Context(), sessionID, &req)
	if err != nil {
		h.HandleError(w, err, "set group name")
		return
	}

	// Log sucesso
	h.LogSuccess("set group name", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"new_name":   req.Name,
	})

	// Resposta de sucesso
	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// SetGroupDescription altera a descrição do grupo
func (h *GroupHandler) SetGroupDescription(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group description")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group description not implemented yet")
}

// SetGroupPhoto altera a foto do grupo
func (h *GroupHandler) SetGroupPhoto(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group photo")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group photo not implemented yet")
}

// GetGroupInviteLink obtém link de convite do grupo
func (h *GroupHandler) GetGroupInviteLink(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group invite link")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group invite link not implemented yet")
}

// JoinGroupViaLink entra em grupo via link
func (h *GroupHandler) JoinGroupViaLink(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "join group via link")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Join group via link not implemented yet")
}

// LeaveGroup sai do grupo
func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "leave group")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Leave group not implemented yet")
}

// UpdateGroupSettings atualiza configurações do grupo
func (h *GroupHandler) UpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update group settings")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Update group settings not implemented yet")
}

// GetGroupRequestParticipants lista solicitações pendentes
func (h *GroupHandler) GetGroupRequestParticipants(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group request participants")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group request participants not implemented yet")
}

// UpdateGroupRequestParticipants gerencia solicitações de entrada
func (h *GroupHandler) UpdateGroupRequestParticipants(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "update group request participants")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Update group request participants not implemented yet")
}

// SetGroupJoinApprovalMode define modo de aprovação
func (h *GroupHandler) SetGroupJoinApprovalMode(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group join approval mode")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group join approval mode not implemented yet")
}

// SetGroupMemberAddMode define modo de adição de membros
func (h *GroupHandler) SetGroupMemberAddMode(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "set group member add mode")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Set group member add mode not implemented yet")
}

// GetGroupInfoFromLink obtém info do grupo via link
func (h *GroupHandler) GetGroupInfoFromLink(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group info from link")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group info from link not implemented yet")
}

// GetGroupInfoFromInvite obtém info do grupo via convite
func (h *GroupHandler) GetGroupInfoFromInvite(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get group info from invite")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Get group info from invite not implemented yet")
}

// JoinGroupWithInvite entra no grupo com convite
func (h *GroupHandler) JoinGroupWithInvite(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "join group with invite")
	// TODO: Implementar
	h.GetWriter().WriteError(w, http.StatusNotImplemented, "Join group with invite not implemented yet")
}