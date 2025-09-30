package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/media"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type MediaHandler struct {
	*BaseHandler
	mediaUC media.UseCase
}

func NewMediaHandler(appLogger *logger.Logger, mediaUC media.UseCase, sessionRepo helpers.SessionRepository) *MediaHandler {
	sessionResolver := &SessionResolver{
		logger:      appLogger,
		sessionRepo: sessionRepo,
	}
	return &MediaHandler{
		BaseHandler: NewBaseHandler(appLogger, sessionResolver),
		mediaUC:     mediaUC,
	}
}

// @Summary Download media from message
// @Description Download media content from a WhatsApp message
// @Tags Media
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body media.DownloadMediaRequest true "Download media request"
// @Success 200 {object} common.SuccessResponse{data=media.DownloadMediaResponse} "Media downloaded successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/download [post]
func (h *MediaHandler) DownloadMedia(w http.ResponseWriter, r *http.Request) {
	h.handleActionRequest(
		w,
		r,
		"Downloading media",
		"Media downloaded successfully",
		func(r *http.Request, sess *domainSession.Session) (interface{}, error) {
			var req media.DownloadMediaRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			req.SessionID = sess.ID.String()
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return h.mediaUC.DownloadMedia(ctx, req.(*media.DownloadMediaRequest))
		},
	)
}

// @Summary Get media information
// @Description Get information about media files in cache
// @Tags Media
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param messageId query string true "Message ID containing media"
// @Success 200 {object} common.SuccessResponse{data=media.MediaInfo} "Media information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/info [get]
func (h *MediaHandler) GetMediaInfo(w http.ResponseWriter, r *http.Request) {
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

	messageID := r.URL.Query().Get("messageId")
	if messageID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Message ID is required"))
		return
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

	result, err := h.mediaUC.GetMediaInfo(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get media info: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get media info"))
		return
	}

	response := common.NewSuccessResponse(result, "Media information retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary List cached media files
// @Description List all cached media files for a session
// @Tags Media
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Param mediaType query string false "Filter by media type (image, video, audio, document)"
// @Success 200 {object} common.SuccessResponse{data=media.ListCachedMediaResponse} "Cached media files listed successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/list [get]
func (h *MediaHandler) ListCachedMedia(w http.ResponseWriter, r *http.Request) {
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

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	mediaType := r.URL.Query().Get("mediaType")

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
		"media_type":   mediaType,
	})

	req := &media.ListCachedMediaRequest{
		SessionID: sess.ID.String(),
		Limit:     limit,
		Offset:    offset,
		MediaType: mediaType,
	}

	result, err := h.mediaUC.ListCachedMedia(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list cached media: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to list cached media"))
		return
	}

	response := common.NewSuccessResponse(result, "Cached media files listed successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Clear media cache
// @Description Clear cached media files for a session
// @Tags Media
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body media.ClearCacheRequest true "Clear cache request"
// @Success 200 {object} common.SuccessResponse{data=media.ClearCacheResponse} "Media cache cleared successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/clear-cache [post]
func (h *MediaHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.handleActionRequest(
		w,
		r,
		"Clearing media cache",
		"Media cache cleared successfully",
		func(r *http.Request, sess *domainSession.Session) (interface{}, error) {
			var req media.ClearCacheRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			req.SessionID = sess.ID.String()
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return h.mediaUC.ClearCache(ctx, req.(*media.ClearCacheRequest))
		},
	)
}

// @Summary Get media statistics
// @Description Get statistics about media usage for a session
// @Tags Media
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=media.GetMediaStatsResponse} "Media statistics retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/media/stats [get]
func (h *MediaHandler) GetMediaStats(w http.ResponseWriter, r *http.Request) {
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

	h.logger.InfoWithFields("Getting media statistics", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	req := &media.GetMediaStatsRequest{
		SessionID: sess.ID.String(),
	}

	result, err := h.mediaUC.GetMediaStats(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get media statistics: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get media statistics"))
		return
	}

	response := common.NewSuccessResponse(result, "Media statistics retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
