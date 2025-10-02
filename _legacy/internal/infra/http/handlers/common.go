package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"zpwoot/internal/app/common"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"github.com/go-chi/chi/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func titleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}



type BaseHandler struct {
	logger      *logger.Logger
	sessionRepo ports.SessionRepository
}

func NewBaseHandler(logger *logger.Logger, sessionRepo ports.SessionRepository) *BaseHandler {
	return &BaseHandler{
		logger:      logger,
		sessionRepo: sessionRepo,
	}
}

// ResolveSession - única função para resolver sessão por ID ou nome
func (h *BaseHandler) ResolveSession(ctx context.Context, idOrName string) (*session.Session, error) {
	// Primeiro tenta por ID, depois por nome
	sess, err := h.sessionRepo.GetByID(ctx, idOrName)
	if err == session.ErrSessionNotFound {
		// Se não encontrou por ID, tenta por nome
		return h.sessionRepo.GetByName(ctx, idOrName)
	}
	return sess, err
}

// GetSessionFromURL extrai sessionId da URL e resolve a sessão
func (h *BaseHandler) GetSessionFromURL(r *http.Request) (*session.Session, error) {
	sessionId := chi.URLParam(r, "sessionId")
	if sessionId == "" {
		return nil, session.ErrSessionNotFound
	}
	return h.ResolveSession(r.Context(), sessionId)
}



func (h *BaseHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(common.NewErrorResponse(message)); err != nil {
		h.logger.Error("Failed to encode error response: " + err.Error())
	}
}

func (h *BaseHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	response := common.NewSuccessResponse(data, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode success response: " + err.Error())
	}
}


