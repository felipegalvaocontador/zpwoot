package common

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

type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"zpwoot"`
	Version string `json:"version,omitempty" example:"1.0.0"`
	Uptime  string `json:"uptime,omitempty" example:"2h30m15s"`
} // @name HealthResponse

type PaginationResponse struct {
	Total   int  `json:"total" example:"100"`
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	Page    int  `json:"page" example:"1"`
	Pages   int  `json:"pages" example:"5"`
	HasNext bool `json:"hasNext" example:"true"`
	HasPrev bool `json:"hasPrev" example:"false"`
}

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

type APIKeyResponse struct {
	Key       string   `json:"key" example:"zpwoot_api_key_123"`
	Name      string   `json:"name" example:"My API Key"`
	ExpiresAt string   `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
	Scopes    []string `json:"scopes" example:"sessions:read,sessions:write"`
}

type StatusResponse struct {
	Status string `json:"status" example:"active"`
}

type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
} // @name MessageResponse

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
