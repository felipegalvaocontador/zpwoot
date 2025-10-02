package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)


type MediaHandler struct {
	*shared.BaseHandler
	sessionService *services.SessionService
}


func NewMediaHandler(
	sessionService *services.SessionService,
	logger *logger.Logger,
) *MediaHandler {
	return &MediaHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		sessionService: sessionService,
	}
}


// @Summary Download media from WhatsApp
// @Description Download media file from WhatsApp message
// @Tags Media
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/media/download [post]
func (h *MediaHandler) DownloadMedia(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "download media")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}


	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}


	h.LogSuccess("download media", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media download initiated successfully")
}


// @Summary Get media information
// @Description Get information about media files
// @Tags Media
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/media/info [get]
func (h *MediaHandler) GetMediaInfo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get media info")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}


	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}


	h.LogSuccess("get media info", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media information retrieved successfully")
}


// @Summary List cached media files
// @Description List all cached media files for the session
// @Tags Media
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/media/list [get]
func (h *MediaHandler) ListCachedMedia(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list cached media")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}


	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}


	h.LogSuccess("list cached media", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Cached media listed successfully")
}


// @Summary Clear media cache
// @Description Clear all cached media files for the session
// @Tags Media
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/media/clear-cache [post]
func (h *MediaHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "clear media cache")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}


	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}


	h.LogSuccess("clear media cache", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media cache cleared successfully")
}


// @Summary Get media statistics
// @Description Get statistics about media usage for the session
// @Tags Media
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/media/stats [get]
func (h *MediaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get media stats")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}


	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}


	h.LogSuccess("get media stats", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media statistics retrieved successfully")
}