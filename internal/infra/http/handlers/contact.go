package handlers

import (
	"context"
	"fmt"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/contact"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type ContactHandler struct {
	logger          *logger.Logger
	contactUC       contact.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewContactHandler(appLogger *logger.Logger, contactUC contact.UseCase, sessionRepo helpers.SessionRepository) *ContactHandler {
	return &ContactHandler{
		logger:          appLogger,
		contactUC:       contactUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// handleContactAction handles common contact action logic
func (h *ContactHandler) handleContactAction(
	c *fiber.Ctx,
	actionName string,
	successMessage string,
	parseFunc func(*fiber.Ctx, *session.Session) (interface{}, error),
	actionFunc func(context.Context, interface{}) (interface{}, error),
) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	req, err := parseFunc(c, sess)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		return c.Status(400).JSON(common.NewErrorResponse("Invalid request body"))
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(c.Context(), req)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))
		return c.Status(500).JSON(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName)))
	}

	response := common.NewSuccessResponse(result, successMessage)
	return c.JSON(response)
}

// @Summary Check if phone numbers are on WhatsApp
// @Description Check if one or more phone numbers are registered on WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body contact.CheckWhatsAppRequest true "Phone numbers to check"
// @Success 200 {object} common.SuccessResponse{data=contact.CheckWhatsAppResponse} "Phone numbers checked successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/check [post]
func (h *ContactHandler) CheckWhatsApp(c *fiber.Ctx) error {
	return h.handleContactAction(
		c,
		"Checking WhatsApp numbers",
		"Phone numbers checked successfully",
		func(c *fiber.Ctx, sess *session.Session) (interface{}, error) {
			var req contact.CheckWhatsAppRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			req.SessionID = sess.ID.String()
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return h.contactUC.CheckWhatsApp(ctx, req.(*contact.CheckWhatsAppRequest))
		},
	)
}

// @Summary Get profile picture
// @Description Get profile picture URL and metadata for a WhatsApp user
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param jid query string true "WhatsApp JID" example("5511999999999@s.whatsapp.net")
// @Param preview query bool false "Get preview (low resolution) image" default(false)
// @Success 200 {object} common.SuccessResponse{data=contact.ProfilePictureResponse} "Profile picture retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or user not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/avatar [get]
func (h *ContactHandler) GetProfilePicture(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	jid := c.Query("jid")
	if jid == "" {
		return c.Status(400).JSON(common.NewErrorResponse("JID is required"))
	}

	preview := c.QueryBool("preview", false)

	h.logger.InfoWithFields("Getting profile picture", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
		"preview":      preview,
	})

	req := &contact.GetProfilePictureRequest{
		SessionID: sess.ID.String(),
		JID:       jid,
		Preview:   preview,
	}

	result, err := h.contactUC.GetProfilePicture(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get profile picture: " + err.Error())
		if err.Error() == "user not found" {
			return c.Status(404).JSON(common.NewErrorResponse("User not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get profile picture"))
	}

	response := common.NewSuccessResponse(result, "Profile picture retrieved successfully")
	return c.JSON(response)
}

// @Summary Get user information
// @Description Get detailed information about WhatsApp users
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body contact.GetUserInfoRequest true "User JIDs to get info for"
// @Success 200 {object} common.SuccessResponse{data=contact.GetUserInfoResponse} "User information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/info [post]
func (h *ContactHandler) GetUserInfo(c *fiber.Ctx) error {
	return h.handleContactAction(
		c,
		"Getting user info",
		"User information retrieved successfully",
		func(c *fiber.Ctx, sess *session.Session) (interface{}, error) {
			var req contact.GetUserInfoRequest
			if err := c.BodyParser(&req); err != nil {
				return nil, err
			}
			req.SessionID = sess.ID.String()
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return h.contactUC.GetUserInfo(ctx, req.(*contact.GetUserInfoRequest))
		},
	)
}

// @Summary List contacts
// @Description List all contacts from the WhatsApp account
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Param search query string false "Search contacts by name or phone"
// @Success 200 {object} common.SuccessResponse{data=contact.ListContactsResponse} "Contacts retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts [get]
func (h *ContactHandler) ListContacts(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)
	search := c.Query("search", "")

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	h.logger.InfoWithFields("Listing contacts", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"limit":        limit,
		"offset":       offset,
		"search":       search,
	})

	req := &contact.ListContactsRequest{
		SessionID: sess.ID.String(),
		Limit:     limit,
		Offset:    offset,
		Search:    search,
	}

	result, err := h.contactUC.ListContacts(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list contacts: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to list contacts"))
	}

	response := common.NewSuccessResponse(result, "Contacts retrieved successfully")
	return c.JSON(response)
}

// @Summary Sync contacts
// @Description Synchronize contacts from the device with WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=contact.SyncContactsResponse} "Contacts synchronized successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/sync [post]
func (h *ContactHandler) SyncContacts(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	h.logger.InfoWithFields("Syncing contacts", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	req := &contact.SyncContactsRequest{
		SessionID: sess.ID.String(),
	}

	result, err := h.contactUC.SyncContacts(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to sync contacts: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to sync contacts"))
	}

	response := common.NewSuccessResponse(result, "Contacts synchronized successfully")
	return c.JSON(response)
}

// @Summary Get business profile
// @Description Get business profile information for a WhatsApp Business account
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param jid query string true "WhatsApp JID" example("5511999999999@s.whatsapp.net")
// @Success 200 {object} common.SuccessResponse{data=contact.BusinessProfileResponse} "Business profile retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or business not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/business [get]
func (h *ContactHandler) GetBusinessProfile(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	jid := c.Query("jid")
	if jid == "" {
		return c.Status(400).JSON(common.NewErrorResponse("JID is required"))
	}

	h.logger.InfoWithFields("Getting business profile", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
	})

	req := &contact.GetBusinessProfileRequest{
		SessionID: sess.ID.String(),
		JID:       jid,
	}

	result, err := h.contactUC.GetBusinessProfile(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get business profile: " + err.Error())
		if err.Error() == "business not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Business profile not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get business profile"))
	}

	response := common.NewSuccessResponse(result, "Business profile retrieved successfully")
	return c.JSON(response)
}

func (h *ContactHandler) resolveSession(c *fiber.Ctx) (*session.Session, *fiber.Error) {
	idOrName := c.Params("sessionId")

	sess, err := h.sessionResolver.ResolveSession(c.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       c.Path(),
		})

		if err.Error() == "session not found" {
			return nil, fiber.NewError(404, "Session not found")
		}

		return nil, fiber.NewError(500, "Failed to resolve session")
	}

	return sess, nil
}
