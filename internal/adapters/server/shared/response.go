package shared

import (
	"encoding/json"
	"net/http"

	"zpwoot/platform/logger"
)


type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
	Success bool        `json:"success" example:"true"`
} // @name SuccessResponse


type ErrorResponse struct {
	Details interface{} `json:"details,omitempty"`
	Error   string      `json:"error" example:"Invalid request"`
	Code    string      `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Success bool        `json:"success" example:"false"`
} // @name ErrorResponse


type ValidationError struct {
	Field   string `json:"field" example:"name"`
	Message string `json:"message" example:"Name is required"`
	Value   string `json:"value,omitempty" example:""`
}


type ValidationErrorResponse struct {
	Error   string            `json:"error" example:"Validation failed"`
	Details []ValidationError `json:"details"`
	Success bool              `json:"success" example:"false"`
}


type PaginationResponse struct {
	Total   int  `json:"total" example:"100"`
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	Page    int  `json:"page" example:"1"`
	Pages   int  `json:"pages" example:"5"`
	HasNext bool `json:"hasNext" example:"true"`
	HasPrev bool `json:"hasPrev" example:"false"`
}


type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"zpwoot"`
	Version string `json:"version,omitempty" example:"1.0.0"`
	Uptime  string `json:"uptime,omitempty" example:"2h30m15s"`
} // @name HealthResponse


type ResponseWriter struct {
	logger *logger.Logger
}


func NewResponseWriter(logger *logger.Logger) *ResponseWriter {
	return &ResponseWriter{
		logger: logger,
	}
}


func (rw *ResponseWriter) WriteSuccess(w http.ResponseWriter, data interface{}, message ...string) {
	response := NewSuccessResponse(data, message...)
	rw.writeJSON(w, http.StatusOK, response)
}


func (rw *ResponseWriter) WriteCreated(w http.ResponseWriter, data interface{}, message ...string) {
	response := NewSuccessResponse(data, message...)
	rw.writeJSON(w, http.StatusCreated, response)
}


func (rw *ResponseWriter) WriteError(w http.ResponseWriter, statusCode int, message string, details ...interface{}) {
	response := NewErrorResponse(message, details...)
	rw.writeJSON(w, statusCode, response)
}


func (rw *ResponseWriter) WriteBadRequest(w http.ResponseWriter, message string, details ...interface{}) {
	rw.WriteError(w, http.StatusBadRequest, message, details...)
}


func (rw *ResponseWriter) WriteUnauthorized(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusUnauthorized, message)
}


func (rw *ResponseWriter) WriteNotFound(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusNotFound, message)
}


func (rw *ResponseWriter) WriteConflict(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusConflict, message)
}


func (rw *ResponseWriter) WriteValidationError(w http.ResponseWriter, errors []ValidationError) {
	response := NewValidationErrorResponse(errors)
	rw.writeJSON(w, http.StatusBadRequest, response)
}


func (rw *ResponseWriter) WriteInternalError(w http.ResponseWriter, message string) {
	rw.WriteError(w, http.StatusInternalServerError, message)
}


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


func NewValidationErrorResponse(errors []ValidationError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success: false,
		Error:   "Validation failed",
		Details: errors,
	}
}


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


func NewHealthResponse(service, version, uptime string) *HealthResponse {
	return &HealthResponse{
		Status:  "ok",
		Service: service,
		Version: version,
		Uptime:  uptime,
	}
}




func IsSuccessStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}


func IsClientError(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}


func IsServerError(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}


func GetStatusText(statusCode int) string {
	return http.StatusText(statusCode)
}
