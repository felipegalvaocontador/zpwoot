package handlers

import (
	"context"
	"errors"
	"fmt"

	"zpwoot/internal/app/community"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
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
func (h *CommunityHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if err.Error() == ErrSessionNotFound || errors.Is(err, domainSession.ErrSessionNotFound) {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Internal server error")
	}

	return sess, nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// parseLinkGroupRequest parses link group request
func (h *CommunityHandler) parseLinkGroupRequest(c *fiber.Ctx) (interface{}, error) {
	var req community.LinkGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse link group request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fiber.NewError(400, "Invalid request body")
	}
	return &req, nil
}

// parseUnlinkGroupRequest parses unlink group request
func (h *CommunityHandler) parseUnlinkGroupRequest(c *fiber.Ctx) (interface{}, error) {
	var req community.UnlinkGroupRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse unlink group request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fiber.NewError(400, "Invalid request body")
	}
	return &req, nil
}

// handleCommunityQueryAction handles common community query action logic
func (h *CommunityHandler) handleCommunityQueryAction(
	c *fiber.Ctx,
	actionName string,
	paramName string,
	createRequestFunc func(string) interface{},
	actionFunc func(context.Context, string, interface{}) (interface{}, error),
) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	paramValue := c.Query(paramName)
	if paramValue == "" {
		return fiber.NewError(400, fmt.Sprintf("%s parameter is required", paramName))
	}

	req := createRequestFunc(paramValue)

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id": sess.ID.String(),
		paramName:    paramValue,
	})

	response, err := actionFunc(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			paramName:    paramValue,
			"error":      err.Error(),
		})

		if err.Error() == ErrSessionNotConnected {
			return fiber.NewError(400, "Session is not connected")
		}
		if err.Error() == ErrCommunityNotFound {
			return fiber.NewError(404, "Community not found")
		}
		return fiber.NewError(500, fmt.Sprintf("Failed to %s", actionName))
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s successfully", actionName), map[string]interface{}{
		"session_id": sess.ID,
		paramName:    paramValue,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// handleGroupLinkAction handles common group link/unlink logic
func (h *CommunityHandler) handleGroupLinkAction(
	c *fiber.Ctx,
	actionName string,
	parseFunc func(*fiber.Ctx) (interface{}, error),
	actionFunc func(context.Context, string, interface{}) (interface{}, error),
) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	req, err := parseFunc(c)
	if err != nil {
		return err
	}

	// Extract JIDs for logging (works for both Link and Unlink requests)
	var communityJID, groupJID string
	if linkReq, ok := req.(*community.LinkGroupRequest); ok {
		communityJID = linkReq.CommunityJID
		groupJID = linkReq.GroupJID
	} else if unlinkReq, ok := req.(*community.UnlinkGroupRequest); ok {
		communityJID = unlinkReq.CommunityJID
		groupJID = unlinkReq.GroupJID
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s group", actionName), map[string]interface{}{
		"session_id":    sess.ID.String(),
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	response, err := actionFunc(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s group", actionName), map[string]interface{}{
			"session_id":    sess.ID.String(),
			"community_jid": communityJID,
			"group_jid":     groupJID,
			"error":         err.Error(),
		})

		if err.Error() == ErrSessionNotConnected {
			return fiber.NewError(400, "Session is not connected")
		}
		if err.Error() == ErrValidationFailed {
			return fiber.NewError(400, ErrInvalidRequestData)
		}
		return fiber.NewError(500, fmt.Sprintf("Failed to %s group", actionName))
	}

	h.logger.InfoWithFields(fmt.Sprintf("Group %s successfully", actionName), map[string]interface{}{
		"session_id":    sess.ID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	return c.Status(200).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// LinkGroup links a group to a community
// POST /sessions/:sessionId/communities/link-group
func (h *CommunityHandler) LinkGroup(c *fiber.Ctx) error {
	return h.handleGroupLinkAction(
		c,
		"link",
		h.parseLinkGroupRequest,
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleGroupLinkAction wrapper
			return h.communityUC.LinkGroup(ctx, sessionID, req.(*community.LinkGroupRequest))
		},
	)
}

// UnlinkGroup unlinks a group from a community
// POST /sessions/:sessionId/communities/unlink-group
func (h *CommunityHandler) UnlinkGroup(c *fiber.Ctx) error {
	return h.handleGroupLinkAction(
		c,
		"unlink",
		h.parseUnlinkGroupRequest,
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleGroupLinkAction wrapper
			return h.communityUC.UnlinkGroup(ctx, sessionID, req.(*community.UnlinkGroupRequest))
		},
	)
}

// GetCommunityInfo gets information about a community
// GET /sessions/:sessionId/communities/info?communityJid=...
func (h *CommunityHandler) GetCommunityInfo(c *fiber.Ctx) error {
	return h.handleCommunityQueryAction(
		c,
		"get community info",
		"communityJid",
		func(jid string) interface{} {
			return &community.GetCommunityInfoRequest{CommunityJID: jid}
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleCommunityAction wrapper
			return h.communityUC.GetCommunityInfo(ctx, sessionID, req.(*community.GetCommunityInfoRequest))
		},
	)
}

// GetSubGroups gets all sub-groups of a community
// GET /sessions/:sessionId/communities/subgroups?communityJid=...
func (h *CommunityHandler) GetSubGroups(c *fiber.Ctx) error {
	return h.handleCommunityQueryAction(
		c,
		"get community sub-groups",
		"communityJid",
		func(jid string) interface{} {
			return &community.GetSubGroupsRequest{CommunityJID: jid}
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleCommunityQueryAction wrapper
			return h.communityUC.GetSubGroups(ctx, sessionID, req.(*community.GetSubGroupsRequest))
		},
	)
}
