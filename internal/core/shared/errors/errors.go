package errors

import (
	"errors"
	"fmt"
)

// Erros comuns do domínio
var (
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternal      = errors.New("internal error")
)

// Erros específicos de sessão
var (
	ErrSessionNotFound           = errors.New("session not found")
	ErrSessionNameAlreadyExists  = errors.New("session name already exists")
	ErrSessionAlreadyConnected   = errors.New("session already connected")
	ErrSessionNotConnected       = errors.New("session not connected")
	ErrSessionInvalidState       = errors.New("session in invalid state")
)

// DomainError representa um erro de domínio
type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

// NewDomainError cria um novo erro de domínio
func NewDomainError(code, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError cria um novo erro de validação
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
