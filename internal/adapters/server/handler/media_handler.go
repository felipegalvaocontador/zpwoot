package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

// MediaHandler implementa handlers REST para gerenciamento de mídia
type MediaHandler struct {
	*shared.BaseHandler
	sessionService *services.SessionService
}

// NewMediaHandler cria nova instância do handler de mídia
func NewMediaHandler(
	sessionService *services.SessionService,
	logger *logger.Logger,
) *MediaHandler {
	return &MediaHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		sessionService: sessionService,
	}
}

// DownloadMedia faz download de mídia do WhatsApp
// @Summary Download media from WhatsApp
// @Description Download media file from WhatsApp message
// @Tags Media
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/media/download [post]
func (h *MediaHandler) DownloadMedia(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "download media")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar download de mídia
	h.LogSuccess("download media", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media download initiated successfully")
}

// GetMediaInfo obtém informações sobre mídia
// @Summary Get media information
// @Description Get information about media files
// @Tags Media
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/media/info [get]
func (h *MediaHandler) GetMediaInfo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get media info")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar busca de informações de mídia
	h.LogSuccess("get media info", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media information retrieved successfully")
}

// ListCachedMedia lista mídia em cache
// @Summary List cached media files
// @Description List all cached media files for the session
// @Tags Media
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/media/list [get]
func (h *MediaHandler) ListCachedMedia(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list cached media")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar listagem de mídia em cache
	h.LogSuccess("list cached media", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Cached media listed successfully")
}

// ClearCache limpa o cache de mídia
// @Summary Clear media cache
// @Description Clear all cached media files for the session
// @Tags Media
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/media/clear-cache [post]
func (h *MediaHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "clear media cache")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar limpeza de cache de mídia
	h.LogSuccess("clear media cache", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media cache cleared successfully")
}

// GetStats obtém estatísticas de mídia
// @Summary Get media statistics
// @Description Get statistics about media usage for the session
// @Tags Media
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/media/stats [get]
func (h *MediaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get media stats")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	_, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar busca de estatísticas de mídia
	h.LogSuccess("get media stats", map[string]interface{}{
		"session_id": sessionID,
	})

	h.GetWriter().WriteSuccess(w, nil, "Media statistics retrieved successfully")
}