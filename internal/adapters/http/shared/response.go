package shared

import (
	"encoding/json"
	"net/http"

	"zpwoot/platform/logger"
)

// SuccessResponse estrutura padrão para respostas de sucesso
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
	Success bool        `json:"success" example:"true"`
} // @name SuccessResponse

// ErrorResponse estrutura padrão para respostas de erro
type ErrorResponse struct {
	Details interface{} `json:"details,omitempty"`
	Error   string      `json:"error" example:"Invalid request"`
	Code    string      `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Success bool        `json:"success" example:"false"`
} // @name ErrorResponse

// ValidationError representa um erro de validação específico
type ValidationError struct {
	Field   string `json:"field" example:"name"`
	Message string `json:"message" example:"Name is required"`
	Value   string `json:"value,omitempty" example:""`
}

// ValidationErrorResponse resposta para erros de validação
type ValidationErrorResponse struct {
	Error   string            `json:"error" example:"Validation failed"`
	Details []ValidationError `json:"details"`
	Success bool              `json:"success" example:"false"`
}

// PaginationResponse informações de paginação
type PaginationResponse struct {
	Total   int  `json:"total" example:"100"`
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	Page    int  `json:"page" example:"1"`
	Pages   int  `json:"pages" example:"5"`
	HasNext bool `json:"hasNext" example:"true"`
	HasPrev bool `json:"hasPrev" example:"false"`
}

// HealthResponse resposta para health check
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"zpwoot"`
	Version string `json:"version,omitempty" example:"1.0.0"`
	Uptime  string `json:"uptime,omitempty" example:"2h30m15s"`
} // @name HealthResponse

// ResponseWriter utilitário para escrever respostas HTTP
type ResponseWriter struct {
	logger *logger.Logger
}

// NewResponseWriter cria nova instância do response writer
func NewResponseWriter(logger *logger.Logger) *ResponseWriter {
	return &ResponseWriter{
		logger: logger,
	}
}

// WriteSuccess escreve resposta de sucesso
func (rw *ResponseWriter) WriteSuccess(w http.ResponseWriter, data interface{}, message ...string) {
	response := NewSuccessResponse(data, message...)
	rw.writeJSON(w, http.StatusOK, response)
}

// WriteCreated escreve resposta de criação (201)
func (rw *ResponseWriter) WriteCreated(w http.ResponseWriter, data interface{}, message ...string) {
	response := NewSuccessResponse(data, message...)
	rw.writeJSON(w, http.StatusCreated, response)
}

// WriteError escreve resposta de erro
func (rw *ResponseWriter) WriteError(w http.ResponseWriter, statusCode int, message string, details ...interface{}) {
	response := NewErrorResponse(message, details...)
	rw.writeJSON(w, statusCode, response)
}

// WriteBadRequest escreve resposta de bad request (400)
func (rw *ResponseWriter) WriteBadRequest(w http.ResponseWriter, message string, details ...interface{}) {
	rw.WriteError(w, http.StatusBadRequest, message, details...)
}

// WriteUnauthorized escreve resposta de não autorizado (401)
func (rw *ResponseWriter) WriteUnauthorized(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusUnauthorized, message)
}

// WriteNotFound escreve resposta de não encontrado (404)
func (rw *ResponseWriter) WriteNotFound(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusNotFound, message)
}

// WriteConflict escreve resposta de conflito (409)
func (rw *ResponseWriter) WriteConflict(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusConflict, message)
}

// WriteValidationError escreve resposta de erro de validação
func (rw *ResponseWriter) WriteValidationError(w http.ResponseWriter, errors []ValidationError) {
	response := NewValidationErrorResponse(errors)
	rw.writeJSON(w, http.StatusBadRequest, response)
}

// WriteInternalError escreve resposta de erro interno (500)
func (rw *ResponseWriter) WriteInternalError(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusInternalServerError, message)
}

// writeJSON escreve resposta JSON
func (rw *ResponseWriter) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		rw.logger.ErrorWithFields("Failed to encode JSON response", map[string]interface{}{
			"error":       err.Error(),
			"status_code": statusCode,
		})
	}
}

// ===== FACTORY FUNCTIONS =====

// NewSuccessResponse cria nova resposta de sucesso
func NewSuccessResponse(data interface{}, message ...string) *SuccessResponse {
	response := &SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(message) > 0 {
		response.Message = message[0]
	}

	return response
}

// NewErrorResponse cria nova resposta de erro
func NewErrorResponse(message string, details ...interface{}) *ErrorResponse {
	response := &ErrorResponse{
		Success: false,
		Error:   message,
	}

	if len(details) > 0 {
		response.Details = details[0]
	}

	return response
}

// NewValidationErrorResponse cria nova resposta de erro de validação
func NewValidationErrorResponse(errors []ValidationError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Details: errors,
	}
}

// NewPaginationResponse cria nova resposta de paginação
func NewPaginationResponse(total, limit, offset int) *PaginationResponse {
	page := (offset / limit) + 1
	pages := (total + limit - 1) / limit

	return &PaginationResponse{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Page:    page,
		Pages:   pages,
		HasNext: offset+limit < total,
		HasPrev: offset > 0,
	}
}

// NewHealthResponse cria nova resposta de health
func NewHealthResponse(service, version, uptime string) *HealthResponse {
	return &HealthResponse{
		Status:  "ok",
		Service: service,
		Version: version,
		Uptime:  uptime,
	}
}

// ===== HTTP STATUS HELPERS =====

// IsSuccessStatus verifica se status code é de sucesso (2xx)
func IsSuccessStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// IsClientError verifica se status code é erro do cliente (4xx)
func IsClientError(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

// IsServerError verifica se status code é erro do servidor (5xx)
func IsServerError(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}

// GetStatusText retorna texto descritivo do status code
func GetStatusText(statusCode int) string {
	return http.StatusText(statusCode)
}
