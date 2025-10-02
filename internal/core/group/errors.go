package group

import "errors"

// Erros de domínio para grupos
var (
	// Erros de validação
	ErrInvalidGroupName        = errors.New("invalid group name")
	ErrInvalidGroupDescription = errors.New("invalid group description")
	ErrInvalidParticipant      = errors.New("invalid participant")
	ErrInvalidInviteLink       = errors.New("invalid invite link")
	ErrInvalidJID              = errors.New("invalid JID")
	ErrInvalidGroupSettings    = errors.New("invalid group settings")

	// Erros de negócio
	ErrGroupNotFound           = errors.New("group not found")
	ErrGroupAlreadyExists      = errors.New("group already exists")
	ErrParticipantNotFound     = errors.New("participant not found")
	ErrParticipantAlreadyExists = errors.New("participant already exists")
	ErrNotGroupAdmin           = errors.New("user is not a group admin")
	ErrNotGroupOwner           = errors.New("user is not the group owner")
	ErrNotGroupParticipant     = errors.New("user is not a group participant")

	// Erros de permissão
	ErrCannotRemoveOwner       = errors.New("cannot remove group owner")
	ErrCannotLeaveAsOwner      = errors.New("group owner cannot leave")
	ErrCannotDemoteOwner       = errors.New("cannot demote group owner")
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// Erros de participantes
	ErrNoParticipants          = errors.New("no participants provided")
	ErrTooManyParticipants     = errors.New("too many participants")
	ErrDuplicateParticipant    = errors.New("duplicate participant")
	ErrParticipantAlreadyAdmin = errors.New("participant is already an admin")
	ErrParticipantNotAdmin     = errors.New("participant is not an admin")

	// Erros de convite
	ErrInviteLinkNotFound      = errors.New("invite link not found")
	ErrInviteLinkExpired       = errors.New("invite link expired")
	ErrInviteLinkInactive      = errors.New("invite link is inactive")

	// Erros de solicitação
	ErrGroupRequestNotFound    = errors.New("group request not found")
	ErrGroupRequestAlreadyExists = errors.New("group request already exists")
	ErrGroupRequestAlreadyProcessed = errors.New("group request already processed")

	// Erros de operação
	ErrGroupLocked             = errors.New("group is locked")
	ErrOperationNotAllowed     = errors.New("operation not allowed")
	ErrInvalidAction           = errors.New("invalid action")
)

// GroupError representa um erro específico de grupo com contexto
type GroupError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

func (e *GroupError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *GroupError) Unwrap() error {
	return e.Cause
}

// NewGroupError cria um novo erro de grupo
func NewGroupError(code, message string, cause error) *GroupError {
	return &GroupError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adiciona contexto ao erro
func (e *GroupError) WithContext(key string, value interface{}) *GroupError {
	e.Context[key] = value
	return e
}

// Códigos de erro específicos
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeAlreadyExists  = "ALREADY_EXISTS"
	ErrCodePermission     = "PERMISSION_DENIED"
	ErrCodeOperation      = "OPERATION_FAILED"
	ErrCodeInviteLink     = "INVITE_LINK_ERROR"
	ErrCodeParticipant    = "PARTICIPANT_ERROR"
	ErrCodeSettings       = "SETTINGS_ERROR"
)

// Funções helper para criar erros específicos
func NewValidationError(message string, cause error) *GroupError {
	return NewGroupError(ErrCodeValidation, message, cause)
}

func NewNotFoundError(message string) *GroupError {
	return NewGroupError(ErrCodeNotFound, message, nil)
}

func NewAlreadyExistsError(message string) *GroupError {
	return NewGroupError(ErrCodeAlreadyExists, message, nil)
}

func NewPermissionError(message string) *GroupError {
	return NewGroupError(ErrCodePermission, message, nil)
}

func NewOperationError(message string, cause error) *GroupError {
	return NewGroupError(ErrCodeOperation, message, cause)
}

func NewInviteLinkError(message string, cause error) *GroupError {
	return NewGroupError(ErrCodeInviteLink, message, cause)
}

func NewParticipantError(message string, cause error) *GroupError {
	return NewGroupError(ErrCodeParticipant, message, cause)
}

func NewSettingsError(message string, cause error) *GroupError {
	return NewGroupError(ErrCodeSettings, message, cause)
}
