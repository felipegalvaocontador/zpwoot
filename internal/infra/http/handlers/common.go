package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"zpwoot/internal/app/common"
	"zpwoot/internal/domain/session"
	"zpwoot/platform/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func titleCase(s string) string {
	caser := cases.Title(language.English)
	return caser.String(s)
}

type SessionRepository interface {
	GetByID(ctx context.Context, id string) (*session.Session, error)
	GetByName(ctx context.Context, name string) (*session.Session, error)
}

type SessionResolver struct {
	logger      *logger.Logger
	sessionRepo SessionRepository
}

func NewSessionResolver(logger *logger.Logger, sessionRepo SessionRepository) *SessionResolver {
	return &SessionResolver{
		logger:      logger,
		sessionRepo: sessionRepo,
	}
}

func (sr *SessionResolver) ResolveSessionIdentifier(idOrName string) (identifierType string, value string, isValid bool) {
	idOrName = strings.TrimSpace(idOrName)

	if idOrName == "" {
		sr.logger.Warn("Empty session identifier provided")
		return "", "", false
	}

	if sr.isValidUUID(idOrName) {
		sr.logger.DebugWithFields("Resolved as UUID", map[string]interface{}{
			"identifier": idOrName,
			"type":       "uuid",
		})
		return "uuid", idOrName, true
	}

	if sr.isValidSessionName(idOrName) {
		sr.logger.DebugWithFields("Resolved as session name", map[string]interface{}{
			"identifier": idOrName,
			"type":       "name",
		})
		return "name", idOrName, true
	}

	sr.logger.WarnWithFields("Invalid session identifier", map[string]interface{}{
		"identifier": idOrName,
		"reason":     "not a valid UUID or session name",
	})
	return "", "", false
}

func (sr *SessionResolver) isValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}

func (sr *SessionResolver) isValidSessionName(name string) bool {
	if len(name) < 1 || len(name) > 100 {
		return false
	}

	urlSafePattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !urlSafePattern.MatchString(name) {
		return false
	}

	if sr.looksLikeUUID(name) {
		return false
	}

	reservedNames := []string{
		"create", "list", "info", "delete", "connect", "logout",
		"qr", "pair", "proxy", "webhook", "chatwoot", "health",
		"swagger", "api", "admin", "config", "status", "test",
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return false
		}
	}

	return true
}

func (sr *SessionResolver) looksLikeUUID(str string) bool {
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidPattern.MatchString(str)
}

func (sr *SessionResolver) ValidateSessionName(name string) (bool, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		return false, "Session name cannot be empty"
	}

	if len(name) < 3 {
		return false, "Session name must be at least 3 characters long"
	}

	if len(name) > 50 {
		return false, "Session name must be at most 50 characters long"
	}

	creationPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !creationPattern.MatchString(name) {
		return false, "Session name must start with a letter and contain only letters, numbers, hyphens, and underscores"
	}

	reservedNames := []string{
		"create", "list", "info", "delete", "connect", "logout",
		"qr", "pair", "proxy", "webhook", "chatwoot", "health",
		"swagger", "api", "admin", "config", "status", "test",
		"new", "add", "remove", "update", "edit", "view", "show",
	}

	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return false, "Session name '" + name + "' is reserved and cannot be used"
		}
	}

	return true, ""
}

func (sr *SessionResolver) SuggestValidName(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	validPattern := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	suggested := validPattern.ReplaceAllString(input, "-")

	multiHyphen := regexp.MustCompile(`-+`)
	suggested = multiHyphen.ReplaceAllString(suggested, "-")

	if len(suggested) > 0 && !regexp.MustCompile(`^[a-zA-Z]`).MatchString(suggested) {
		suggested = "session-" + suggested
	}

	suggested = strings.Trim(suggested, "-")

	if len(suggested) < 3 {
		suggested = "session-" + suggested
	}

	if len(suggested) > 50 {
		suggested = suggested[:50]
		suggested = strings.TrimRight(suggested, "-")
	}

	return suggested
}

func (sr *SessionResolver) ResolveSession(ctx context.Context, idOrName string) (*session.Session, error) {
	identifierType, value, isValid := sr.ResolveSessionIdentifier(idOrName)
	if !isValid {
		return nil, session.ErrSessionNotFound
	}

	sr.logger.InfoWithFields("Resolving session", map[string]interface{}{
		"identifier":      value,
		"identifier_type": identifierType,
	})

	var sess *session.Session
	var err error

	switch identifierType {
	case "uuid":
		sess, err = sr.sessionRepo.GetByID(ctx, value)
	case "name":
		sess, err = sr.sessionRepo.GetByName(ctx, value)
	default:
		return nil, session.ErrSessionNotFound
	}

	if err != nil {
		sr.logger.ErrorWithFields("Failed to resolve session", map[string]interface{}{
			"identifier":      value,
			"identifier_type": identifierType,
			"error":           err.Error(),
		})
		return nil, err
	}

	sr.logger.InfoWithFields("Session resolved successfully", map[string]interface{}{
		"identifier":      value,
		"identifier_type": identifierType,
		"session_id":      sess.ID.String(),
		"session_name":    sess.Name,
	})

	return sess, nil
}

type BaseHandler struct {
	logger          *logger.Logger
	sessionResolver *SessionResolver
}

func NewBaseHandler(logger *logger.Logger, sessionResolver *SessionResolver) *BaseHandler {
	return &BaseHandler{
		logger:          logger,
		sessionResolver: sessionResolver,
	}
}

func (h *BaseHandler) resolveSession(r *http.Request) (*session.Session, error) {
	if sess, ok := r.Context().Value("session").(*session.Session); ok {
		return sess, nil
	}

	sessionIdentifier := r.URL.Query().Get("sessionId")
	if sessionIdentifier == "" {
		if urlParam := r.Context().Value("sessionId"); urlParam != nil {
			if sessionID, ok := urlParam.(string); ok {
				sessionIdentifier = sessionID
			}
		}
	}

	if sessionIdentifier == "" {
		return nil, session.ErrSessionNotFound
	}

	return h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
}

func (h *BaseHandler) resolveSessionFromURL(r *http.Request) (*session.Session, error) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		return nil, session.ErrSessionNotFound
	}

	sess, err := h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": sessionIdentifier,
			"error":      err.Error(),
			"path":       r.URL.Path,
		})
		return nil, err
	}

	return sess, nil
}

func (h *BaseHandler) handleActionRequest(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	parseFunc func(*http.Request, *session.Session) (interface{}, error),
	actionFunc func(context.Context, interface{}) (interface{}, error),
) {
	sess, err := h.resolveSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, session.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	req, err := parseFunc(r, sess)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(r.Context(), req)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))
		h.writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to %s", actionName))
		return
	}

	h.writeSuccessResponse(w, result, successMessage)
}

func (h *BaseHandler) handleSimpleGetRequest(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	actionFunc func(context.Context, string) (interface{}, error),
) {
	sess, err := h.resolveSessionFromURL(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, session.ErrSessionNotFound) {
			statusCode = 404
		}
		h.writeErrorResponse(w, statusCode, err.Error())
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(r.Context(), sess.ID.String())
	if err != nil {
		h.logger.Error("Failed to " + actionName + ": " + err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to "+actionName)
		return
	}

	h.writeSuccessResponse(w, result, successMessage)
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
