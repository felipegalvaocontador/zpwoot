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

// BaseHandler fornece funcionalidades comuns para todos os handlers HTTP
type BaseHandler struct {
	logger     *logger.Logger
	writer     *ResponseWriter
	validator  *validation.Validator
}

// NewBaseHandler cria nova instância do base handler
func NewBaseHandler(logger *logger.Logger) *BaseHandler {
	return &BaseHandler{
		logger:    logger,
		writer:    NewResponseWriter(logger),
		validator: validation.New(),
	}
}

// GetLogger retorna logger do handler
func (h *BaseHandler) GetLogger() *logger.Logger {
	return h.logger
}

// GetWriter retorna response writer
func (h *BaseHandler) GetWriter() *ResponseWriter {
	return h.writer
}

// GetValidator retorna validator
func (h *BaseHandler) GetValidator() *validation.Validator {
	return h.validator
}

// ===== URL PARAMETER EXTRACTION =====

// GetSessionIDFromURL extrai session ID da URL (aceita UUID ou nome de sessão)
func (h *BaseHandler) GetSessionIDFromURL(r *http.Request) (uuid.UUID, error) {
	sessionIDStr := chi.URLParam(r, "sessionId")
	if sessionIDStr == "" {
		return uuid.Nil, fmt.Errorf("session ID is required")
	}

	// Tentar primeiro como UUID
	sessionID, err := uuid.Parse(sessionIDStr)
	if err == nil {
		return sessionID, nil
	}

	// Se não for UUID, é um nome de sessão - retornar erro especial para indicar que é nome
	return uuid.Nil, fmt.Errorf("session_name:%s", sessionIDStr)
}

// GetSessionNameFromURL extrai nome da sessão da URL
func (h *BaseHandler) GetSessionNameFromURL(r *http.Request) (string, error) {
	sessionName := chi.URLParam(r, "sessionId")
	if sessionName == "" {
		return "", fmt.Errorf("session identifier is required")
	}
	return sessionName, nil
}

// GetStringParam extrai parâmetro string da URL
func (h *BaseHandler) GetStringParam(r *http.Request, paramName string) (string, error) {
	value := chi.URLParam(r, paramName)
	if value == "" {
		return "", fmt.Errorf("%s is required", paramName)
	}
	return value, nil
}

// GetIntParam extrai parâmetro int da URL
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

// ===== QUERY PARAMETER EXTRACTION =====

// GetQueryString extrai parâmetro string da query
func (h *BaseHandler) GetQueryString(r *http.Request, paramName string, defaultValue ...string) string {
	value := r.URL.Query().Get(paramName)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// GetQueryInt extrai parâmetro int da query
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

// GetQueryBool extrai parâmetro bool da query
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

// ===== REQUEST BODY PARSING =====

// ParseJSONBody faz parse do body JSON para struct
func (h *BaseHandler) ParseJSONBody(r *http.Request, dest interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Rejeitar campos desconhecidos

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	return nil
}

// ParseAndValidateJSON faz parse e validação do body JSON
func (h *BaseHandler) ParseAndValidateJSON(r *http.Request, dest interface{}) error {
	// Parse JSON
	if err := h.ParseJSONBody(r, dest); err != nil {
		return err
	}

	// Validar struct
	if err := h.validator.ValidateStruct(dest); err != nil {
		return err
	}

	return nil
}

// ===== PAGINATION HELPERS =====

// GetPaginationParams extrai parâmetros de paginação da query
func (h *BaseHandler) GetPaginationParams(r *http.Request) (limit, offset int, err error) {
	limit, err = h.GetQueryInt(r, "limit", 20)
	if err != nil {
		return 0, 0, err
	}

	offset, err = h.GetQueryInt(r, "offset", 0)
	if err != nil {
		return 0, 0, err
	}

	// Validar limites
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

// ===== ERROR HANDLING =====

// HandleError processa erro e escreve resposta apropriada
func (h *BaseHandler) HandleError(w http.ResponseWriter, err error, operation string) {
	h.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", operation), map[string]interface{}{
		"error": err.Error(),
	})

	// Determinar status code baseado no tipo de erro
	statusCode := h.getStatusCodeFromError(err)
	message := h.getMessageFromError(err, operation)

	h.writer.WriteError(w, statusCode, message)
}

// getStatusCodeFromError determina status code baseado no erro
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
		// Se contém "validation" na mensagem, é erro de validação
		if contains(err.Error(), "validation") {
			return http.StatusBadRequest
		}
		// Se contém "not found" na mensagem, é 404
		if contains(err.Error(), "not found") {
			return http.StatusNotFound
		}
		// Se contém "already exists" na mensagem, é 409
		if contains(err.Error(), "already exists") {
			return http.StatusConflict
		}
		// Padrão: erro interno
		return http.StatusInternalServerError
	}
}

// getMessageFromError determina mensagem baseada no erro
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
		// Para outros erros, usar mensagem genérica
		return fmt.Sprintf("Failed to %s", operation)
	}
}

// ===== LOGGING HELPERS =====

// LogRequest registra informações da requisição
func (h *BaseHandler) LogRequest(r *http.Request, operation string) {
	h.logger.InfoWithFields(fmt.Sprintf("Processing %s request", operation), map[string]interface{}{
		"method":     r.Method,
		"path":       r.URL.Path,
		"query":      r.URL.RawQuery,
		"user_agent": r.Header.Get("User-Agent"),
		"ip":         getClientIP(r),
	})
}

// LogSuccess registra sucesso da operação
func (h *BaseHandler) LogSuccess(operation string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["operation"] = operation
	
	h.logger.InfoWithFields(fmt.Sprintf("%s completed successfully", operation), details)
}

// ===== UTILITY FUNCTIONS =====

// getClientIP extrai IP do cliente
func getClientIP(r *http.Request) string {
	// Verificar headers de proxy
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// contains verifica se string contém substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}

// containsSubstring helper para busca de substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
