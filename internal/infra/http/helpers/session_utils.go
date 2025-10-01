package helpers

import (
	"encoding/json"
	"errors"
	"net/http"

	"zpwoot/internal/app/common"
	"zpwoot/internal/domain/session"
	"zpwoot/platform/logger"

	"github.com/go-chi/chi/v5"
)

// SessionUtils fornece utilitários para trabalhar com sessões em handlers
type SessionUtils struct {
	sessionResolver *SessionResolver
	logger          *logger.Logger
}

// NewSessionUtils cria uma nova instância dos utilitários de sessão
func NewSessionUtils(sessionResolver *SessionResolver, logger *logger.Logger) *SessionUtils {
	return &SessionUtils{
		sessionResolver: sessionResolver,
		logger:          logger,
	}
}

// ResolveSessionStrategy define diferentes estratégias para resolver sessões
type ResolveSessionStrategy int

const (
	// FromURL resolve apenas da URL (chi.URLParam)
	FromURL ResolveSessionStrategy = iota
	// FromQuery resolve apenas de query parameters
	FromQuery
	// FromContext resolve apenas do context
	FromContext
	// FromAny tenta todas as estratégias (URL -> Query -> Context)
	FromAny
)

// SessionResolutionOptions configura como resolver a sessão
type SessionResolutionOptions struct {
	Strategy    ResolveSessionStrategy
	ParamName   string // nome do parâmetro (default: "sessionId")
	Required    bool   // se true, retorna erro se não encontrar
	LogFailures bool   // se true, loga falhas de resolução
}

// DefaultOptions retorna opções padrão para resolução de sessão
func DefaultOptions() SessionResolutionOptions {
	return SessionResolutionOptions{
		Strategy:    FromAny,
		ParamName:   "sessionId",
		Required:    true,
		LogFailures: true,
	}
}

// ResolveSession resolve uma sessão usando as opções especificadas
func (su *SessionUtils) ResolveSession(r *http.Request, opts SessionResolutionOptions) (*session.Session, error) {
	var sessionIdentifier string

	switch opts.Strategy {
	case FromURL:
		sessionIdentifier = chi.URLParam(r, opts.ParamName)
	case FromQuery:
		sessionIdentifier = r.URL.Query().Get(opts.ParamName)
	case FromContext:
		if value := r.Context().Value(opts.ParamName); value != nil {
			if id, ok := value.(string); ok {
				sessionIdentifier = id
			}
		}
	case FromAny:
		// Tenta URL primeiro
		sessionIdentifier = chi.URLParam(r, opts.ParamName)

		// Se não encontrou, tenta query
		if sessionIdentifier == "" {
			sessionIdentifier = r.URL.Query().Get(opts.ParamName)
		}

		// Se ainda não encontrou, tenta context
		if sessionIdentifier == "" {
			if value := r.Context().Value(opts.ParamName); value != nil {
				if id, ok := value.(string); ok {
					sessionIdentifier = id
				}
			}
		}
	}

	// Se não encontrou identificador e é obrigatório
	if sessionIdentifier == "" && opts.Required {
		return nil, session.ErrSessionNotFound
	}

	// Se não encontrou mas não é obrigatório
	if sessionIdentifier == "" {
		return nil, nil
	}

	// Resolve a sessão
	sess, err := su.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil && opts.LogFailures {
		su.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": sessionIdentifier,
			"strategy":   opts.Strategy,
			"error":      err.Error(),
			"path":       r.URL.Path,
			"method":     r.Method,
		})
	}

	return sess, err
}

// WithSessionHandler é um helper que executa um handler apenas se conseguir resolver a sessão
func (su *SessionUtils) WithSessionHandler(
	w http.ResponseWriter,
	r *http.Request,
	opts SessionResolutionOptions,
	handler func(*session.Session, http.ResponseWriter, *http.Request),
) {
	sess, err := su.ResolveSession(r, opts)
	if err != nil {
		su.WriteSessionError(w, err)
		return
	}

	if sess == nil && opts.Required {
		su.WriteSessionError(w, session.ErrSessionNotFound)
		return
	}

	handler(sess, w, r)
}

// WriteSessionError escreve um erro de sessão como resposta JSON
func (su *SessionUtils) WriteSessionError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal server error"

	if errors.Is(err, session.ErrSessionNotFound) {
		statusCode = http.StatusNotFound
		message = "Session not found"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := common.NewErrorResponse(message)
	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		su.logger.Error("Failed to encode error response: " + encErr.Error())
	}
}

// QuickResolveFromURL é um helper rápido para resolver sessão da URL
func (su *SessionUtils) QuickResolveFromURL(r *http.Request) (*session.Session, error) {
	return su.ResolveSession(r, SessionResolutionOptions{
		Strategy:    FromURL,
		ParamName:   "sessionId",
		Required:    true,
		LogFailures: true,
	})
}

// QuickResolveFromQuery é um helper rápido para resolver sessão de query params
func (su *SessionUtils) QuickResolveFromQuery(r *http.Request) (*session.Session, error) {
	return su.ResolveSession(r, SessionResolutionOptions{
		Strategy:    FromQuery,
		ParamName:   "sessionId",
		Required:    true,
		LogFailures: true,
	})
}

// QuickResolveAny é um helper rápido para resolver sessão de qualquer fonte
func (su *SessionUtils) QuickResolveAny(r *http.Request) (*session.Session, error) {
	return su.ResolveSession(r, DefaultOptions())
}

// HandleWithSession é um wrapper que simplifica handlers que precisam de sessão
func (su *SessionUtils) HandleWithSession(
	handler func(*session.Session, http.ResponseWriter, *http.Request),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		su.WithSessionHandler(w, r, DefaultOptions(), handler)
	}
}

// HandleWithOptionalSession é um wrapper para handlers onde sessão é opcional
func (su *SessionUtils) HandleWithOptionalSession(
	handler func(*session.Session, http.ResponseWriter, *http.Request),
) http.HandlerFunc {
	opts := DefaultOptions()
	opts.Required = false

	return func(w http.ResponseWriter, r *http.Request) {
		su.WithSessionHandler(w, r, opts, handler)
	}
}
