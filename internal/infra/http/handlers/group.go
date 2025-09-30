package handlers

import (
	"errors"

	"zpwoot/internal/app/group"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type GroupHandler struct {
	logger          *logger.Logger
	groupUC         group.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewGroupHandler(appLogger *logger.Logger, groupUC group.UseCase, sessionRepo helpers.SessionRepository) *GroupHandler {
	return &GroupHandler{
		logger:          appLogger,
		groupUC:         groupUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

func (h *GroupHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if errors.Is(err, domainSession.ErrSessionNotFound) {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Internal server error")
	}

	return sess, nil
}

// handleGroupActionWithJID handles common group action logic for requests with GroupJID


// CreateGroup creates a new WhatsApp group
func (h *GroupHandler) CreateGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"name":         req.Name,
		"participants": len(req.Participants),
	})

	response, err := h.groupUC.CreateGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// GetGroupInfo retrieves information about a specific group
func (h *GroupHandler) GetGroupInfo(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Query("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required as query parameter: ?groupJid=...")
	}

	req := &group.GetGroupInfoRequest{
		GroupJID: groupJID,
	}

	h.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
	})

	response, err := h.groupUC.GetGroupInfo(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group info", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// ListGroups lists all groups the user is a member of
func (h *GroupHandler) ListGroups(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	h.logger.InfoWithFields("Listing groups", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.groupUC.ListGroups(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.ErrorWithFields("Failed to list groups", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// UpdateGroupParticipants adds, removes, promotes, or demotes group participants
func (h *GroupHandler) UpdateGroupParticipants(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.UpdateGroupParticipantsRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	groupJID := req.GroupJID

	h.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"group_jid":    groupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	response, err := h.groupUC.UpdateGroupParticipants(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// SetGroupName updates the group name
func (h *GroupHandler) SetGroupName(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.SetGroupNameRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	h.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  req.GroupJID,
	})

	response, err := h.groupUC.SetGroupName(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to Setting group name", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// SetGroupDescription updates the group description
func (h *GroupHandler) SetGroupDescription(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.SetGroupDescriptionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	h.logger.InfoWithFields("Setting group description", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  req.GroupJID,
	})

	response, err := h.groupUC.SetGroupDescription(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to Setting group description", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// SetGroupPhoto updates the group photo
func (h *GroupHandler) SetGroupPhoto(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.SetGroupPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	h.logger.InfoWithFields("Setting group photo", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  req.GroupJID,
	})

	response, err := h.groupUC.SetGroupPhoto(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to Setting group photo", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// GetGroupInviteLink retrieves or generates a group invite link
func (h *GroupHandler) GetGroupInviteLink(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJID := c.Query("groupJid")
	if groupJID == "" {
		return fiber.NewError(400, "Group JID is required as query parameter: ?groupJid=...")
	}

	reset := c.QueryBool("reset", false)

	req := &group.GetGroupInviteLinkRequest{
		GroupJID: groupJID,
		Reset:    reset,
	}

	h.logger.InfoWithFields("Getting group invite link", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
		"reset":      reset,
	})

	response, err := h.groupUC.GetGroupInviteLink(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// JoinGroup joins a group using an invite link
func (h *GroupHandler) JoinGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.JoinGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Joining group", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.groupUC.JoinGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to join group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// LeaveGroup leaves a group
func (h *GroupHandler) LeaveGroup(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.LeaveGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	h.logger.InfoWithFields("Leaving group", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  req.GroupJID,
	})

	response, err := h.groupUC.LeaveGroup(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to Leaving group", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// UpdateGroupSettings updates group settings (announce, locked)
func (h *GroupHandler) UpdateGroupSettings(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.UpdateGroupSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Invalid request body", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	groupJID := req.GroupJID

	h.logger.InfoWithFields("Updating group settings", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJID,
		"announce":   req.Announce,
		"locked":     req.Locked,
	})

	response, err := h.groupUC.UpdateGroupSettings(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to update group settings", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// GetGroupRequestParticipants gets the list of participants that have requested to join the group
func (h *GroupHandler) GetGroupRequestParticipants(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	groupJid := c.Query("groupJid")
	if groupJid == "" {
		return fiber.NewError(400, "Group JID is required as query parameter: ?groupJid=...")
	}

	h.logger.InfoWithFields("Getting group request participants", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJid,
	})

	participants, err := h.groupUC.GetGroupRequestParticipants(c.Context(), sess.ID.String(), groupJid)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group request participants", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJid,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	response := map[string]interface{}{
		"groupJid":     groupJid,
		"participants": participants,
		"total":        len(participants),
	}

	return c.JSON(response)
}

// UpdateGroupRequestParticipants approves or rejects requests to join the group
func (h *GroupHandler) UpdateGroupRequestParticipants(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req struct {
		GroupJID     string   `json:"groupJid"`     // Group JID
		Action       string   `json:"action"`       // "approve" or "reject"
		Participants []string `json:"participants"` // List of participant JIDs
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	if req.Action == "" {
		return fiber.NewError(400, "Action is required")
	}

	if len(req.Participants) == 0 {
		return fiber.NewError(400, "At least one participant is required")
	}

	groupJid := req.GroupJID

	h.logger.InfoWithFields("Updating group request participants", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"group_jid":    groupJid,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	success, failed, err := h.groupUC.UpdateGroupRequestParticipants(c.Context(), sess.ID.String(), groupJid, req.Participants, req.Action)
	if err != nil {
		h.logger.ErrorWithFields("Failed to update group request participants", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJid,
			"action":     req.Action,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	response := map[string]interface{}{
		"groupJid":     groupJid,
		"action":       req.Action,
		"participants": req.Participants,
		"success":      success,
		"failed":       failed,
	}

	return c.JSON(response)
}

// SetGroupJoinApprovalMode sets the group join approval mode
func (h *GroupHandler) SetGroupJoinApprovalMode(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req struct {
		GroupJID        string `json:"groupJid"`
		RequireApproval bool   `json:"requireApproval"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	groupJid := req.GroupJID

	h.logger.InfoWithFields("Setting group join approval mode", map[string]interface{}{
		"session_id":       sess.ID.String(),
		"group_jid":        groupJid,
		"require_approval": req.RequireApproval,
	})

	err := h.groupUC.SetGroupJoinApprovalMode(c.Context(), sess.ID.String(), groupJid, req.RequireApproval)
	if err != nil {
		h.logger.ErrorWithFields("Failed to set group join approval mode", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJid,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	response := map[string]interface{}{
		"groupJid":        groupJid,
		"requireApproval": req.RequireApproval,
		"status":          "updated",
	}

	return c.JSON(response)
}

// SetGroupMemberAddMode sets the group member add mode
func (h *GroupHandler) SetGroupMemberAddMode(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req struct {
		GroupJID string `json:"groupJid"`
		Mode     string `json:"mode"` // "admin_add" or "all_member_add"
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid request body")
	}

	if req.GroupJID == "" {
		return fiber.NewError(400, "Group JID is required in request body")
	}

	if req.Mode == "" {
		return fiber.NewError(400, "Mode is required")
	}

	groupJid := req.GroupJID

	h.logger.InfoWithFields("Setting group member add mode", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  groupJid,
		"mode":       req.Mode,
	})

	err := h.groupUC.SetGroupMemberAddMode(c.Context(), sess.ID.String(), groupJid, req.Mode)
	if err != nil {
		h.logger.ErrorWithFields("Failed to set group member add mode", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  groupJid,
			"error":      err.Error(),
		})
		return fiber.NewError(500, err.Error())
	}

	response := map[string]interface{}{
		"groupJid": groupJid,
		"mode":     req.Mode,
		"status":   "updated",
	}

	return c.JSON(response)
}

// ============================================================================
// ADVANCED GROUP HANDLERS
// ============================================================================

// GetGroupInfoFromLink gets group information from an invite link
// GET /sessions/:sessionId/groups/info-from-link?inviteLink=...
func (h *GroupHandler) GetGroupInfoFromLink(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	inviteLink := c.Query("inviteLink")
	if inviteLink == "" {
		return fiber.NewError(400, "Invite link parameter is required")
	}

	req := &group.GetGroupInfoFromLinkRequest{
		InviteLink: inviteLink,
	}

	h.logger.InfoWithFields("Getting group info from link", map[string]interface{}{
		"session_id":  sess.ID.String(),
		"invite_link": inviteLink,
	})

	response, err := h.groupUC.GetGroupInfoFromLink(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group info from link", map[string]interface{}{
			"session_id":  sess.ID.String(),
			"invite_link": inviteLink,
			"error":       err.Error(),
		})

		if errors.Is(err, domainSession.ErrSessionNotConnected) {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to get group info from link")
	}

	h.logger.InfoWithFields("Group info retrieved from link successfully", map[string]interface{}{
		"session_id": sess.ID,
		"group_jid":  response.JID,
		"group_name": response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetGroupInfoFromInvite gets group information from an invite
// POST /sessions/:sessionId/groups/info-from-invite
func (h *GroupHandler) GetGroupInfoFromInvite(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.GetGroupInfoFromInviteRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse get group info from invite request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Getting group info from invite", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  req.GroupJID,
		"code":       req.Code,
	})

	response, err := h.groupUC.GetGroupInfoFromInvite(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get group info from invite", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"code":       req.Code,
			"error":      err.Error(),
		})

		if errors.Is(err, domainSession.ErrSessionNotConnected) {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to get group info from invite")
	}

	h.logger.InfoWithFields("Group info retrieved from invite successfully", map[string]interface{}{
		"session_id": sess.ID,
		"group_jid":  response.JID,
		"group_name": response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// JoinGroupWithInvite joins a group using a specific invite
// POST /sessions/:sessionId/groups/join-with-invite
func (h *GroupHandler) JoinGroupWithInvite(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req group.JoinGroupWithInviteRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse join group with invite request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Joining group with invite", map[string]interface{}{
		"session_id": sess.ID.String(),
		"group_jid":  req.GroupJID,
		"code":       req.Code,
	})

	response, err := h.groupUC.JoinGroupWithInvite(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to join group with invite", map[string]interface{}{
			"session_id": sess.ID.String(),
			"group_jid":  req.GroupJID,
			"code":       req.Code,
			"error":      err.Error(),
		})

		if errors.Is(err, domainSession.ErrSessionNotConnected) {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to join group with invite")
	}

	h.logger.InfoWithFields("Joined group with invite successfully", map[string]interface{}{
		"session_id": sess.ID,
		"group_jid":  response.GroupJID,
		"success":    response.Success,
	})

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}
