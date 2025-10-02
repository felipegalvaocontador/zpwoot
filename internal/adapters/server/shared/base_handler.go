package shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"zpwoot/internal/core/session"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)


type BaseHandler struct {
	logger     *logger.Logger
	writer     *ResponseWriter
	validator  *validation.Validator
}


func NewBaseHandler(logger *logger.Logger) *BaseHandler {
	return &BaseHandler{
		logger:    logger,
		writer:    NewResponseWriter(logger),
		validator: validation.New(),
	}
}


func (h *BaseHandler) GetLogger() *logger.Logger {
	return h.logger
}


func (h *BaseHandler) GetWriter() *ResponseWriter {
	return h.writer
}


func (h *BaseHandler) GetValidator() *validation.Validator {
	return h.validator
}




func (h *BaseHandler) GetSessionIDFromURL(r *http.Request) (uuid.UUID, error) {
	sessionIDStr := chi.URLParam(r, "sessionId")
	if sessionIDStr == "" {
		return uuid.Nil, fmt.Errorf("session ID is required")
	}


	sessionID, err := uuid.Parse(sessionIDStr)
	if err == nil {
		return sessionID, nil
	}


	return uuid.Nil, fmt.Errorf("session_name:%s", sessionIDStr)
}


func (h *BaseHandler) GetSessionNameFromURL(r *http.Request) (string, error) {
	sessionName := chi.URLParam(r, "sessionId")
	if sessionName == "" {
		return "", fmt.Errorf("session identifier is required")
	}
	return sessionName, nil
}


func (h *BaseHandler) GetStringParam(r *http.Request, paramName string) (string, error) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", fmt.Errorf("%s is required", paramName)
	}
	return value, nil
}


func (h *BaseHandler) GetIntParam(r *http.Request, paramName string) (int, error) {
	valueStr := chi.URLParam(r, paramName)
	if valueStr == "" {
		return 0, fmt.Errorf("%s is required", paramName)
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s format: %w", paramName, err)
	}

	return value, nil
}




func (h *BaseHandler) GetQueryString(r *http.Request, paramName string, defaultValue ...string) string {
	value := r.URL.Query().Get(paramName)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}


func (h *BaseHandler) GetQueryInt(r *http.Request, paramName string, defaultValue ...int) (int, error) {
	valueStr := r.URL.Query().Get(paramName)
	if valueStr == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s format: %w", paramName, err)
	}

	return value, nil
}


func (h *BaseHandler) GetQueryBool(r *http.Request, paramName string, defaultValue ...bool) (bool, error) {
	valueStr := r.URL.Query().Get(paramName)
	if valueStr == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return false, nil
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return false, fmt.Errorf("invalid %s format: %w", paramName, err)
	}

	return value, nil
}




func (h *BaseHandler) ParseJSONBody(r *http.Request, dest interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	return nil
}


func (h *BaseHandler) ParseAndValidateJSON(r *http.Request, dest interface{}) error {

	if err := h.ParseJSONBody(r, dest); err != nil {
		return err
	}


	if err := h.validator.ValidateStruct(dest); err != nil {
		return err
	}

	return nil
}




func (h *BaseHandler) GetPaginationParams(r *http.Request) (limit, offset int, err error) {
	limit, err = h.GetQueryInt(r, "limit", 20)
	if err != nil {
		return 0, 0, err
	}

	offset, err = h.GetQueryInt(r, "offset", 0)
	if err != nil {
		return 0, 0, err
	}


	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset, nil
}




func (h *BaseHandler) HandleError(w http.ResponseWriter, err error, operation string) {
	h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", operation), map[string]interface{}{
		"error": err.Error(),
	})


	statusCode := h.getStatusCodeFromError(err)
	message := h.getMessageFromError(err, operation)

	h.writer.WriteError(w, statusCode, message)
}


func (h *BaseHandler) getStatusCodeFromError(err error) int {
	switch {
	case err == session.ErrSessionNotFound:
		return http.StatusNotFound
	case err == session.ErrSessionAlreadyExists:
		return http.StatusConflict
	case err == session.ErrSessionAlreadyConnected:
		return http.StatusConflict
	case err == session.ErrInvalidSessionName:
		return http.StatusBadRequest
	case err == session.ErrInvalidProxyConfig:
		return http.StatusBadRequest
	default:

		if contains(err.Error(), "validation") {
			return http.StatusBadRequest
		}

		if contains(err.Error(), "not found") {
			return http.StatusNotFound
		}

		if contains(err.Error(), "already exists") {
			return http.StatusConflict
		}

		return http.StatusInternalServerError
	}
}


func (h *BaseHandler) getMessageFromError(err error, operation string) string {
	switch {
	case err == session.ErrSessionNotFound:
		return "Session not found"
	case err == session.ErrSessionAlreadyExists:
		return "Session already exists"
	case err == session.ErrSessionAlreadyConnected:
		return "Session is already connected"
	case err == session.ErrInvalidSessionName:
		return "Invalid session name"
	case err == session.ErrInvalidProxyConfig:
		return "Invalid proxy configuration"
	default:

		return fmt.Sprintf("Failed to %s", operation)
	}
}




func (h *BaseHandler) LogRequest(r *http.Request, operation string) {
	h.logger.InfoWithFields(fmt.Sprintf("Processing %s request", operation), map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"query":      r.URL.RawQuery,
		"user_agent": r.Header.Get("User-Agent"),
		"ip":         getClientIP(r),
	})
}


func (h *BaseHandler) LogSuccess(operation string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["operation"] = operation
	
	h.logger.InfoWithFields(fmt.Sprintf("%s completed successfully", operation), details)
}




func getClientIP(r *http.Request) string {

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}


func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}


func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
