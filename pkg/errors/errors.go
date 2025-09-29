package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func NewWithDetails(code int, message, details string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

var (
	ErrBadRequest          = New(http.StatusBadRequest, "Bad request")
	ErrUnauthorized        = New(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden           = New(http.StatusForbidden, "Forbidden")
	ErrNotFound            = New(http.StatusNotFound, "Not found")
	ErrConflict            = New(http.StatusConflict, "Conflict")
	ErrInternalServerError = New(http.StatusInternalServerError, "Internal server error")
	ErrServiceUnavailable  = New(http.StatusServiceUnavailable, "Service unavailable")

	ErrWameowNotConnected = New(http.StatusServiceUnavailable, "Wameow not connected")
	ErrWameowSendFailed   = New(http.StatusInternalServerError, "Failed to send Wameow message")
	ErrInvalidPhoneNumber = New(http.StatusBadRequest, "Invalid phone number")

	ErrChatwootNotConfigured = New(http.StatusServiceUnavailable, "Chatwoot not configured")
	ErrChatwootAPIError      = New(http.StatusInternalServerError, "Chatwoot API error")

	ErrSessionNotFound      = New(http.StatusNotFound, "Session not found")
	ErrSessionAlreadyExists = New(http.StatusConflict, "Session already exists")
	ErrInvalidSessionData   = New(http.StatusBadRequest, "Invalid session data")

	ErrUserNotFound      = New(http.StatusNotFound, "User not found")
	ErrUserAlreadyExists = New(http.StatusConflict, "User already exists")
	ErrInvalidUserData   = New(http.StatusBadRequest, "Invalid user data")

	ErrOrderNotFound      = New(http.StatusNotFound, "Order not found")
	ErrOrderAlreadyExists = New(http.StatusConflict, "Order already exists")
	ErrInvalidOrderData   = New(http.StatusBadRequest, "Invalid order data")
)

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewWithDetails(http.StatusInternalServerError, "Internal server error", err.Error())
}
