package handlers

import (
	"fmt"
	"strconv"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/media"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	logger          *logger.Logger
	mediaUC         media.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewMediaHandler(appLogger *logger.Logger, mediaUC media.UseCase, sessionRepo helpers.SessionRepository) *MediaHandler {
	return &MediaHandler{
		logger:          appLogger,
		mediaUC:         mediaUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// @Summary Download media from message
// @Description Download media (image, video, audio, document) from a WhatsApp message
// @Tags Media
// @Security ApiKeyAuth
// @Produce application/octet-stream
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param messageId path string true "Message ID" example("3EB0C431C26A1916E07E")
// @Success 200 {file} binary "Media file content"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or message not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/download/{messageId} [get]
func (h *MediaHandler) DownloadMedia(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	messageID := c.Params("messageId")
	if messageID == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Message ID is required"))
	}

	h.logger.InfoWithFields("Downloading media", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"message_id":   messageID,
	})

	req := &media.DownloadMediaRequest{
		SessionID: sess.ID.String(),
		MessageID: messageID,
	}

	result, err := h.mediaUC.DownloadMedia(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to download media: " + err.Error())
		if err.Error() == "message not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Message not found"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to download media"))
	}

	// Set appropriate headers
	c.Set("Content-Type", result.MimeType)
	c.Set("Content-Length", strconv.FormatInt(result.FileSize, 10))
	if result.Filename != "" {
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	}

	return c.Send(result.Data)
}

// @Summary Download media by type
// @Description Download media from a message filtered by media type
// @Tags Media
// @Security ApiKeyAuth
// @Produce application/octet-stream
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param messageId path string true "Message ID" example("3EB0C431C26A1916E07E")
// @Param mediaType path string true "Media Type" Enums(image, video, audio, document, sticker)
// @Success 200 {file} binary "Media file content"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or message not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/download/{messageId}/{mediaType} [get]
func (h *MediaHandler) DownloadMediaByType(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	messageID := c.Params("messageId")
	mediaType := c.Params("mediaType")

	if messageID == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Message ID is required"))
	}

	if mediaType == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Media type is required"))
	}

	// Validate media type
	validTypes := []string{"image", "video", "audio", "document", "sticker"}
	isValid := false
	for _, validType := range validTypes {
		if mediaType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return c.Status(400).JSON(common.NewErrorResponse("Invalid media type. Must be one of: image, video, audio, document, sticker"))
	}

	h.logger.InfoWithFields("Downloading media by type", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"message_id":   messageID,
		"media_type":   mediaType,
	})

	req := &media.DownloadMediaRequest{
		SessionID: sess.ID.String(),
		MessageID: messageID,
		MediaType: mediaType,
	}

	result, err := h.mediaUC.DownloadMedia(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to download media: " + err.Error())
		if err.Error() == "message not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Message not found"))
		}
		if err.Error() == "media type mismatch" {
			return c.Status(400).JSON(common.NewErrorResponse("Message does not contain the requested media type"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to download media"))
	}

	// Set appropriate headers
	c.Set("Content-Type", result.MimeType)
	c.Set("Content-Length", strconv.FormatInt(result.FileSize, 10))
	if result.Filename != "" {
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	}

	return c.Send(result.Data)
}

// @Summary Get media info
// @Description Get information about media in a message without downloading it
// @Tags Media
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param messageId path string true "Message ID" example("3EB0C431C26A1916E07E")
// @Success 200 {object} common.SuccessResponse{data=media.MediaInfoResponse} "Media information"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or message not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/info/{messageId} [get]
func (h *MediaHandler) GetMediaInfo(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	messageID := c.Params("messageId")
	if messageID == "" {
		return c.Status(400).JSON(common.NewErrorResponse("Message ID is required"))
	}

	h.logger.InfoWithFields("Getting media info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"message_id":   messageID,
	})

	req := &media.GetMediaInfoRequest{
		SessionID: sess.ID.String(),
		MessageID: messageID,
	}

	result, err := h.mediaUC.GetMediaInfo(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get media info: " + err.Error())
		if err.Error() == "message not found" {
			return c.Status(404).JSON(common.NewErrorResponse("Message not found"))
		}
		if err.Error() == "no media in message" {
			return c.Status(404).JSON(common.NewErrorResponse("Message does not contain media"))
		}
		return c.Status(500).JSON(common.NewErrorResponse("Failed to get media info"))
	}

	response := common.NewSuccessResponse(result, "Media information retrieved successfully")
	return c.JSON(response)
}

// @Summary List cached media
// @Description List all cached media files for a session
// @Tags Media
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} common.SuccessResponse{data=media.ListCachedMediaResponse} "Cached media list"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/cache [get]
func (h *MediaHandler) ListCachedMedia(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	limit := c.QueryInt("limit", 50)
	offset := c.QueryInt("offset", 0)

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	h.logger.InfoWithFields("Listing cached media", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"limit":        limit,
		"offset":       offset,
	})

	req := &media.ListCachedMediaRequest{
		SessionID: sess.ID.String(),
		Limit:     limit,
		Offset:    offset,
	}

	result, err := h.mediaUC.ListCachedMedia(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list cached media: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to list cached media"))
	}

	response := common.NewSuccessResponse(result, "Cached media list retrieved successfully")
	return c.JSON(response)
}

// @Summary Clear media cache
// @Description Clear cached media files for a session
// @Tags Media
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param olderThan query int false "Clear files older than X hours" default(24)
// @Success 200 {object} common.SuccessResponse{data=media.ClearCacheResponse} "Cache cleared successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/cache/clear [delete]
func (h *MediaHandler) ClearMediaCache(c *fiber.Ctx) error {
	sess, fiberErr := h.resolveSession(c)
	if fiberErr != nil {
		return c.Status(fiberErr.Code).JSON(common.NewErrorResponse(fiberErr.Message))
	}

	olderThan := c.QueryInt("olderThan", 24)
	if olderThan < 0 {
		olderThan = 24
	}

	h.logger.InfoWithFields("Clearing media cache", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"older_than":   olderThan,
	})

	req := &media.ClearCacheRequest{
		SessionID: sess.ID.String(),
		OlderThan: olderThan,
	}

	result, err := h.mediaUC.ClearCache(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to clear media cache: " + err.Error())
		return c.Status(500).JSON(common.NewErrorResponse("Failed to clear media cache"))
	}

	response := common.NewSuccessResponse(result, "Media cache cleared successfully")
	return c.JSON(response)
}

func (h *MediaHandler) resolveSession(c *fiber.Ctx) (*domainSession.Session, *fiber.Error) {
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
