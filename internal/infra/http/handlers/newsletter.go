package handlers

import (
	"context"
	"errors"
	"fmt"

	"zpwoot/internal/app/newsletter"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

// NewsletterHandler handles newsletter-related HTTP requests
type NewsletterHandler struct {
	logger          *logger.Logger
	newsletterUC    newsletter.UseCase
	sessionResolver *helpers.SessionResolver
}

// NewNewsletterHandler creates a new newsletter handler
func NewNewsletterHandler(appLogger *logger.Logger, newsletterUC newsletter.UseCase, sessionRepo helpers.SessionRepository) *NewsletterHandler {
	return &NewsletterHandler{
		logger:          appLogger,
		newsletterUC:    newsletterUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// resolveSession resolves session from URL parameter
func (h *NewsletterHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if err.Error() == "session not found" || errors.Is(err, domainSession.ErrSessionNotFound) {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Internal server error")
	}

	return sess, nil
}

// handleNewsletterAction handles common newsletter action logic
func (h *NewsletterHandler) handleNewsletterAction(
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
		h.logger.WarnWithFields(fmt.Sprintf("Failed to parse %s request", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	// Extract NewsletterJID for logging if available
	var newsletterJID string
	if reqWithJID, ok := req.(interface{ GetNewsletterJID() string }); ok {
		newsletterJID = reqWithJID.GetNewsletterJID()
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": newsletterJID,
	})

	response, err := actionFunc(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": newsletterJID,
			"error":          err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "validation failed" {
			return fiber.NewError(400, "Invalid request data")
		}

		return fiber.NewError(500, err.Error())
	}

	return c.JSON(response)
}

// handleNewsletterActionWithFiberMap handles newsletter actions that return fiber.Map responses
func (h *NewsletterHandler) handleNewsletterActionWithFiberMap(
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
		h.logger.WarnWithFields(fmt.Sprintf("Failed to parse %s request", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := actionFunc(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to %s", actionName),
		})
	}

	h.logger.InfoWithFields(fmt.Sprintf("%s successfully", actionName), map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// CreateNewsletter creates a new WhatsApp newsletter/channel
// POST /sessions/:sessionId/newsletters/create
func (h *NewsletterHandler) CreateNewsletter(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.CreateNewsletterRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse create newsletter request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
		"session_id": sess.ID.String(),
		"name":       req.Name,
	})

	response, err := h.newsletterUC.CreateNewsletter(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to create newsletter", map[string]interface{}{
			"session_id": sess.ID.String(),
			"name":       req.Name,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to create newsletter")
	}

	h.logger.InfoWithFields("Newsletter created successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"newsletter_id": response.ID,
		"name":          response.Name,
	})

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetNewsletterInfo gets information about a newsletter by JID
// GET /sessions/:sessionId/newsletters/info?jid=...
func (h *NewsletterHandler) GetNewsletterInfo(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	newsletterJid := c.Query("newsletterJid")
	if newsletterJid == "" {
		return fiber.NewError(400, "Newsletter JID parameter is required")
	}

	req := &newsletter.GetNewsletterInfoRequest{
		NewsletterJID: newsletterJid,
	}

	h.logger.InfoWithFields("Getting newsletter info", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": newsletterJid,
	})

	response, err := h.newsletterUC.GetNewsletterInfo(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get newsletter info", map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": newsletterJid,
			"error":          err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "newsletter not found" {
			return fiber.NewError(404, "Newsletter not found")
		}

		return fiber.NewError(500, "Failed to get newsletter info")
	}

	h.logger.InfoWithFields("Newsletter info retrieved successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"newsletter_id": response.ID,
		"name":          response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetNewsletterInfoWithInvite gets newsletter information using an invite key
// POST /sessions/:sessionId/newsletters/info-from-invite
func (h *NewsletterHandler) GetNewsletterInfoWithInvite(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.GetNewsletterInfoWithInviteRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse get newsletter info with invite request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return fiber.NewError(400, "Invalid request body")
	}

	h.logger.InfoWithFields("Getting newsletter info with invite", map[string]interface{}{
		"session_id": sess.ID.String(),
		"invite_key": req.InviteKey,
	})

	response, err := h.newsletterUC.GetNewsletterInfoWithInvite(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get newsletter info with invite", map[string]interface{}{
			"session_id": sess.ID.String(),
			"invite_key": req.InviteKey,
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		if err.Error() == "newsletter not found" {
			return fiber.NewError(404, "Newsletter not found")
		}

		return fiber.NewError(500, "Failed to get newsletter info with invite")
	}

	h.logger.InfoWithFields("Newsletter info retrieved with invite successfully", map[string]interface{}{
		"session_id":    sess.ID,
		"newsletter_id": response.ID,
		"name":          response.Name,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// FollowNewsletter follows a newsletter
// POST /sessions/:sessionId/newsletters/follow
func (h *NewsletterHandler) FollowNewsletter(c *fiber.Ctx) error {
	return h.handleNewsletterAction(
		c,
		"Following newsletter",
		func(c *fiber.Ctx) (interface{}, error) {
			var req newsletter.FollowNewsletterRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleNewsletterAction wrapper

			return h.newsletterUC.FollowNewsletter(ctx, sessionID, req.(*newsletter.FollowNewsletterRequest)) //nolint:errcheck // Error is handled by handleNewsletterAction wrapper
		},
	)
}

// UnfollowNewsletter unfollows a newsletter
// POST /sessions/:sessionId/newsletters/unfollow
func (h *NewsletterHandler) UnfollowNewsletter(c *fiber.Ctx) error {
	return h.handleNewsletterAction(
		c,
		"Unfollowing newsletter",
		func(c *fiber.Ctx) (interface{}, error) {
			var req newsletter.UnfollowNewsletterRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleNewsletterAction wrapper

			return h.newsletterUC.UnfollowNewsletter(ctx, sessionID, req.(*newsletter.UnfollowNewsletterRequest)) //nolint:errcheck // Error is handled by handleNewsletterAction wrapper
		},
	)
}

// GetSubscribedNewsletters gets all newsletters the user is subscribed to
// GET /sessions/:sessionId/newsletters
func (h *NewsletterHandler) GetSubscribedNewsletters(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	h.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
		"session_id": sess.ID.String(),
	})

	response, err := h.newsletterUC.GetSubscribedNewsletters(c.Context(), sess.ID.String())
	if err != nil {
		h.logger.ErrorWithFields("Failed to get subscribed newsletters", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})

		if err.Error() == "session is not connected" {
			return fiber.NewError(400, "Session is not connected")
		}

		return fiber.NewError(500, "Failed to get subscribed newsletters")
	}

	h.logger.InfoWithFields("Subscribed newsletters retrieved successfully", map[string]interface{}{
		"session_id": sess.ID.String(),
		"count":      response.Total,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetNewsletterMessages gets messages from a newsletter
func (h *NewsletterHandler) GetNewsletterMessages(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	// Parse query parameters
	newsletterJid := c.Query("newsletterJid")
	if newsletterJid == "" {
		h.logger.ErrorWithFields("Missing Newsletter JID parameter", map[string]interface{}{
			"session_id": sess.ID.String(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Missing required parameter: newsletterJid",
		})
	}

	// Parse optional parameters
	count := c.QueryInt("count", 20) // Default to 20 messages
	before := c.Query("before", "")  // Optional pagination

	req := &newsletter.GetNewsletterMessagesRequest{
		NewsletterJID: newsletterJid,
		Count:         count,
		Before:        before,
	}

	h.logger.InfoWithFields("Getting newsletter messages", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"count":          req.Count,
		"before":         req.Before,
	})

	response, err := h.newsletterUC.GetNewsletterMessages(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get newsletter messages", map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": req.NewsletterJID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get newsletter messages",
		})
	}

	h.logger.InfoWithFields("Newsletter messages retrieved successfully", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"count":          len(response.Messages),
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetNewsletterMessageUpdates gets message updates from a newsletter
func (h *NewsletterHandler) GetNewsletterMessageUpdates(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	// Parse query parameters
	newsletterJid := c.Query("newsletterJid")
	if newsletterJid == "" {
		h.logger.ErrorWithFields("Missing Newsletter JID parameter", map[string]interface{}{
			"session_id": sess.ID.String(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Missing required parameter: newsletterJid",
		})
	}

	// Parse optional parameters
	count := c.QueryInt("count", 20) // Default to 20 updates
	since := c.Query("since", "")    // Optional timestamp filter
	after := c.Query("after", "")    // Optional pagination

	req := &newsletter.GetNewsletterMessageUpdatesRequest{
		NewsletterJID: newsletterJid,
		Count:         count,
		Since:         since,
		After:         after,
	}

	h.logger.InfoWithFields("Getting newsletter message updates", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"count":          req.Count,
		"since":          req.Since,
		"after":          req.After,
	})

	response, err := h.newsletterUC.GetNewsletterMessageUpdates(c.Context(), sess.ID.String(), req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to get newsletter message updates", map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": req.NewsletterJID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get newsletter message updates",
		})
	}

	h.logger.InfoWithFields("Newsletter message updates retrieved successfully", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"count":          len(response.Updates),
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// NewsletterMarkViewed marks newsletter messages as viewed
func (h *NewsletterHandler) NewsletterMarkViewed(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.NewsletterMarkViewedRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse newsletter mark viewed request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.InfoWithFields("Marking newsletter messages as viewed", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"count":          len(req.ServerIDs),
	})

	response, err := h.newsletterUC.NewsletterMarkViewed(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to mark newsletter messages as viewed", map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": req.NewsletterJID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to mark newsletter messages as viewed",
		})
	}

	h.logger.InfoWithFields("Newsletter messages marked as viewed successfully", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"count":          len(req.ServerIDs),
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// NewsletterSendReaction sends a reaction to a newsletter message
func (h *NewsletterHandler) NewsletterSendReaction(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.NewsletterSendReactionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse newsletter send reaction request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.InfoWithFields("Sending newsletter reaction", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"server_id":      req.ServerID,
		"reaction":       req.Reaction,
	})

	response, err := h.newsletterUC.NewsletterSendReaction(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to send newsletter reaction", map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": req.NewsletterJID,
			"server_id":      req.ServerID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to send newsletter reaction",
		})
	}

	h.logger.InfoWithFields("Newsletter reaction sent successfully", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"server_id":      req.ServerID,
		"reaction":       req.Reaction,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// NewsletterSubscribeLiveUpdates subscribes to live updates from a newsletter
func (h *NewsletterHandler) NewsletterSubscribeLiveUpdates(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return fiberErr
	}

	var req newsletter.NewsletterSubscribeLiveUpdatesRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.WarnWithFields("Failed to parse newsletter subscribe live updates request", map[string]interface{}{
			"session_id": sess.ID.String(),
			"error":      err.Error(),
		})
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.InfoWithFields("Subscribing to newsletter live updates", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
	})

	response, err := h.newsletterUC.NewsletterSubscribeLiveUpdates(c.Context(), sess.ID.String(), &req)
	if err != nil {
		h.logger.ErrorWithFields("Failed to subscribe to newsletter live updates", map[string]interface{}{
			"session_id":     sess.ID.String(),
			"newsletter_jid": req.NewsletterJID,
			"error":          err.Error(),
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to subscribe to newsletter live updates",
		})
	}

	h.logger.InfoWithFields("Subscribed to newsletter live updates successfully", map[string]interface{}{
		"session_id":     sess.ID.String(),
		"newsletter_jid": req.NewsletterJID,
		"duration":       response.Duration,
	})

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// NewsletterToggleMute toggles mute status of a newsletter
func (h *NewsletterHandler) NewsletterToggleMute(c *fiber.Ctx) error {
	return h.handleNewsletterActionWithFiberMap(
		c,
		"toggle newsletter mute status",
		func(c *fiber.Ctx) (interface{}, error) {
			var req newsletter.NewsletterToggleMuteRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleNewsletterAction wrapper

			return h.newsletterUC.NewsletterToggleMute(ctx, sessionID, req.(*newsletter.NewsletterToggleMuteRequest)) //nolint:errcheck // Error is handled by handleNewsletterAction wrapper
		},
	)
}

// AcceptTOSNotice accepts a terms of service notice
func (h *NewsletterHandler) AcceptTOSNotice(c *fiber.Ctx) error {
	return h.handleNewsletterActionWithFiberMap(
		c,
		"accept TOS notice",
		func(c *fiber.Ctx) (interface{}, error) {
			var req newsletter.AcceptTOSNoticeRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleNewsletterAction wrapper

			return h.newsletterUC.AcceptTOSNotice(ctx, sessionID, req.(*newsletter.AcceptTOSNoticeRequest)) //nolint:errcheck // Error is handled by handleNewsletterAction wrapper
		},
	)
}

// UploadNewsletter uploads media for newsletters
func (h *NewsletterHandler) UploadNewsletter(c *fiber.Ctx) error {
	return h.handleNewsletterActionWithFiberMap(
		c,
		"upload newsletter media",
		func(c *fiber.Ctx) (interface{}, error) {
			var req newsletter.UploadNewsletterRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleNewsletterAction wrapper

			return h.newsletterUC.UploadNewsletter(ctx, sessionID, req.(*newsletter.UploadNewsletterRequest)) //nolint:errcheck // Error is handled by handleNewsletterAction wrapper
		},
	)
}

// UploadNewsletterReader uploads media for newsletters from a reader
func (h *NewsletterHandler) UploadNewsletterReader(c *fiber.Ctx) error {
	return h.handleNewsletterActionWithFiberMap(
		c,
		"upload newsletter media with reader",
		func(c *fiber.Ctx) (interface{}, error) {
			var req newsletter.UploadNewsletterRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			return &req, nil
		},
		func(ctx context.Context, sessionID string, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleNewsletterAction wrapper

			return h.newsletterUC.UploadNewsletterReader(ctx, sessionID, req.(*newsletter.UploadNewsletterRequest)) //nolint:errcheck // Error is handled by handleNewsletterAction wrapper
		},
	)
}
