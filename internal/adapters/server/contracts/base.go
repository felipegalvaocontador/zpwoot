package contracts

import (
	"time"
)

type BaseRequest struct {
	RequestID string    `json:"request_id,omitempty" example:"req-123"`
	Timestamp time.Time `json:"timestamp,omitempty" example:"2024-01-01T12:00:00Z"`
}

type BaseResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message,omitempty" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty" example:"req-123"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

type ErrorResponse struct {
	Success   bool        `json:"success" example:"false"`
	Error     string      `json:"error" example:"Invalid request"`
	Code      string      `json:"code,omitempty" example:"VALIDATION_ERROR"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty" example:"req-123"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

type ValidationError struct {
	Field   string `json:"field" example:"name"`
	Message string `json:"message" example:"Name is required"`
	Value   string `json:"value,omitempty" example:""`
	Code    string `json:"code,omitempty" example:"REQUIRED"`
}

type ValidationErrorResponse struct {
	Success   bool              `json:"success" example:"false"`
	Error     string            `json:"error" example:"Validation failed"`
	Code      string            `json:"code" example:"VALIDATION_ERROR"`
	Details   []ValidationError `json:"details"`
	RequestID string            `json:"request_id,omitempty" example:"req-123"`
	Timestamp time.Time         `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

type PaginationRequest struct {
	Limit  int `json:"limit" validate:"omitempty,min=1,max=100" example:"20"`
	Offset int `json:"offset" validate:"omitempty,min=0" example:"0"`
	Page   int `json:"page" validate:"omitempty,min=1" example:"1"`
}

type PaginationResponse struct {
	Total   int  `json:"total" example:"100"`
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	Page    int  `json:"page" example:"1"`
	Pages   int  `json:"pages" example:"5"`
	HasNext bool `json:"has_next" example:"true"`
	HasPrev bool `json:"has_prev" example:"false"`
}

type ListResponse struct {
	BaseResponse
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type SessionID string

type MessageID string

type GroupID string

type ContactID string

type PhoneNumber string

type JID string

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
	MessageTypeContact  MessageType = "contact"
	MessageTypeLocation MessageType = "location"
	MessageTypeSticker  MessageType = "sticker"
	MessageTypeButton   MessageType = "button"
	MessageTypeList     MessageType = "list"
	MessageTypePoll     MessageType = "poll"
)

type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusReconnecting ConnectionStatus = "reconnecting"
	ConnectionStatusFailed       ConnectionStatus = "failed"
)

func NewBaseResponse(success bool, message string, data interface{}) *BaseResponse {
	return &BaseResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewErrorResponse(error, code string, details interface{}) *ErrorResponse {
	return &ErrorResponse{
		Success:   false,
		Error:     error,
		Code:      code,
		Details:   details,
		Timestamp: time.Now(),
	}
}

func NewValidationErrorResponse(errors []ValidationError) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success:   false,
		Error:     "Validation failed",
		Code:      "VALIDATION_ERROR",
		Details:   errors,
		Timestamp: time.Now(),
	}
}

func NewPaginationResponse(total, limit, offset int) *PaginationResponse {
	page := 1
	if limit > 0 {
		page = (offset / limit) + 1
	}

	pages := 1
	if limit > 0 && total > 0 {
		pages = (total + limit - 1) / limit
	}

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

func IsValidMessageType(msgType string) bool {
	switch MessageType(msgType) {
	case MessageTypeText, MessageTypeImage, MessageTypeAudio, MessageTypeVideo,
		MessageTypeDocument, MessageTypeContact, MessageTypeLocation, MessageTypeSticker,
		MessageTypeButton, MessageTypeList, MessageTypePoll:
		return true
	default:
		return false
	}
}

func IsValidConnectionStatus(status string) bool {
	switch ConnectionStatus(status) {
	case ConnectionStatusConnected, ConnectionStatusDisconnected, ConnectionStatusConnecting,
		ConnectionStatusReconnecting, ConnectionStatusFailed:
		return true
	default:
		return false
	}
}

func IsValidStatus(status string) bool {
	switch Status(status) {
	case StatusActive, StatusInactive, StatusPending, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}
