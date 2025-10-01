package middleware

import (
	"context"
	"net/http"

	"zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"

	"github.com/go-chi/chi/v5"
)

// SessionContextKey é a chave usada para armazenar a sessão no context
type SessionContextKey string

const (
	SessionKey SessionContextKey = "resolved_session"
)

// SessionMiddleware é um middleware que resolve automaticamente a sessão
type SessionMiddleware struct {
	sessionResolver *helpers.SessionResolver
	logger          *logger.Logger
}

// NewSessionMiddleware cria uma nova instância do middleware de sessão
func NewSessionMiddleware(sessionResolver *helpers.SessionResolver, logger *logger.Logger) *SessionMiddleware {
	return &SessionMiddleware{
		sessionResolver: sessionResolver,
		logger:          logger,
	}
}

// WithSession é um middleware que resolve a sessão automaticamente e a adiciona ao context
func (sm *SessionMiddleware) WithSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := sm.ResolveSessionFromRequest(r)
		if err != nil {
			sm.handleSessionError(w, r, err)
			return
		}

		// Adiciona a sessão ao context
		ctx := context.WithValue(r.Context(), SessionKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ResolveSessionFromRequest resolve a sessão a partir da requisição
// Tenta múltiplas fontes: URL param, query param, context
func (sm *SessionMiddleware) ResolveSessionFromRequest(r *http.Request) (*session.Session, error) {
	// 1. Verifica se já existe no context
	if sess, ok := r.Context().Value(SessionKey).(*session.Session); ok {
		return sess, nil
	}

	// 2. Tenta pegar da URL (chi.URLParam)
	sessionIdentifier := chi.URLParam(r, "sessionId")

	// 3. Se não encontrou na URL, tenta query param
	if sessionIdentifier == "" {
		sessionIdentifier = r.URL.Query().Get("sessionId")
	}

	// 4. Se ainda não encontrou, tenta context value
	if sessionIdentifier == "" {
		if urlParam := r.Context().Value("sessionId"); urlParam != nil {
			if sessionID, ok := urlParam.(string); ok {
				sessionIdentifier = sessionID
			}
		}
	}

	// 5. Se não encontrou identificador, retorna erro
	if sessionIdentifier == "" {
		return nil, session.ErrSessionNotFound
	}

	// 6. Resolve a sessão
	sess, err := sm.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		sm.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": sessionIdentifier,
			"error":      err.Error(),
			"path":       r.URL.Path,
			"method":     r.Method,
		})
		return nil, err
	}

	return sess, nil
}

// handleSessionError trata erros de resolução de sessão
func (sm *SessionMiddleware) handleSessionError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	if err == session.ErrSessionNotFound {
		statusCode = http.StatusNotFound
		message = "Session not found"
	}

	sm.logger.WarnWithFields("Session resolution failed", map[string]interface{}{
		"error":  err.Error(),
		"path":   r.URL.Path,
		"method": r.Method,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Resposta JSON simples
	response := `{"success": false, "error": "` + message + `"}`
	if _, writeErr := w.Write([]byte(response)); writeErr != nil {
		sm.logger.Error("Failed to write error response: " + writeErr.Error())
	}
}

// GetSessionFromContext extrai a sessão do context da requisição
func GetSessionFromContext(ctx context.Context) (*session.Session, bool) {
	sess, ok := ctx.Value(SessionKey).(*session.Session)
	return sess, ok
}

// GetSessionFromRequest é um helper para extrair sessão da requisição
func GetSessionFromRequest(r *http.Request) (*session.Session, bool) {
	return GetSessionFromContext(r.Context())
}

// MustGetSessionFromRequest extrai a sessão ou entra em pânico (para casos onde sabemos que existe)
func MustGetSessionFromRequest(r *http.Request) *session.Session {
	sess, ok := GetSessionFromRequest(r)
	if !ok {
		panic("session not found in request context - ensure SessionMiddleware is applied")
	}
	return sess
}
