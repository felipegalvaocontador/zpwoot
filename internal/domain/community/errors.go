package community

import "errors"

var (
	ErrCommunityNotFound     = errors.New("community not found")
	ErrCommunityExists       = errors.New("community already exists")
	ErrInvalidCommunityJID   = errors.New("invalid community JID")
	ErrInvalidGroupJID       = errors.New("invalid group JID")
	ErrCommunityNotConnected = errors.New("community is not connected")

	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrNotCommunityOwner       = errors.New("user is not community owner")
	ErrNotCommunityAdmin       = errors.New("user is not community admin")
	ErrNotCommunityMember      = errors.New("user is not community member")

	ErrGroupNotFound        = errors.New("group not found")
	ErrGroupAlreadyLinked   = errors.New("group is already linked to this community")
	ErrGroupLinkedElsewhere = errors.New("group is already linked to another community")
	ErrCannotLinkToSelf     = errors.New("cannot link community to itself")
	ErrGroupNotLinked       = errors.New("group is not linked to this community")
	ErrMaxGroupsReached     = errors.New("maximum number of linked groups reached")

	ErrEmptyCommunityJID    = errors.New("community JID cannot be empty")
	ErrEmptyGroupJID        = errors.New("group JID cannot be empty")
	ErrEmptyCommunityName   = errors.New("community name cannot be empty")
	ErrCommunityNameTooLong = errors.New("community name is too long")
	ErrDescriptionTooLong   = errors.New("community description is too long")
	ErrInvalidJIDFormat     = errors.New("invalid JID format")

	ErrLinkOperationFailed   = errors.New("group link operation failed")
	ErrUnlinkOperationFailed = errors.New("group unlink operation failed")
	ErrCommunityInfoFailed   = errors.New("failed to get community information")
	ErrSubGroupsFailed       = errors.New("failed to get community sub-groups")

	ErrCommunityAPIUnavailable = errors.New("community API is unavailable")
	ErrCommunityTimeout        = errors.New("community operation timed out")
	ErrCommunityRateLimited    = errors.New("community operation rate limited")

	ErrCommunityFull       = errors.New("community has reached maximum capacity")
	ErrCommunityPrivate    = errors.New("community is private")
	ErrCommunityArchived   = errors.New("community is archived")
	ErrCommunityRestricted = errors.New("community access is restricted")
)

type CommunityError struct {
	Cause   error                  `json:"-"`
	Context map[string]interface{} `json:"context,omitempty"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
}

func (e *CommunityError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

func (e *CommunityError) Unwrap() error {
	return e.Cause
}

func NewCommunityError(code, message string, cause error) *CommunityError {
	return &CommunityError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

func (e *CommunityError) WithDetails(details string) *CommunityError {
	e.Details = details
	return e
}

func (e *CommunityError) WithContext(key string, value interface{}) *CommunityError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

const (
	ErrCodeCommunityNotFound   = "COMMUNITY_NOT_FOUND"
	ErrCodeCommunityExists     = "COMMUNITY_EXISTS"
	ErrCodeInvalidJID          = "INVALID_JID"
	ErrCodeInsufficientPerms   = "INSUFFICIENT_PERMISSIONS"
	ErrCodeGroupAlreadyLinked  = "GROUP_ALREADY_LINKED"
	ErrCodeGroupNotLinked      = "GROUP_NOT_LINKED"
	ErrCodeValidationFailed    = "VALIDATION_FAILED"
	ErrCodeOperationFailed     = "OPERATION_FAILED"
	ErrCodeAPIUnavailable      = "API_UNAVAILABLE"
	ErrCodeTimeout             = "TIMEOUT"
	ErrCodeRateLimited         = "RATE_LIMITED"
	ErrCodeCommunityFull       = "COMMUNITY_FULL"
	ErrCodeCommunityPrivate    = "COMMUNITY_PRIVATE"
	ErrCodeCommunityArchived   = "COMMUNITY_ARCHIVED"
	ErrCodeCommunityRestricted = "COMMUNITY_RESTRICTED"
)

func NewNotFoundError(communityJID string) *CommunityError {
	return NewCommunityError(
		ErrCodeCommunityNotFound,
		"Community not found",
		ErrCommunityNotFound,
	).WithContext("communityJID", communityJID)
}

func NewInvalidJIDError(jid string, reason string) *CommunityError {
	return NewCommunityError(
		ErrCodeInvalidJID,
		"Invalid JID format",
		ErrInvalidJIDFormat,
	).WithContext("jid", jid).WithDetails(reason)
}

func NewInsufficientPermissionsError(userJID, operation string) *CommunityError {
	return NewCommunityError(
		ErrCodeInsufficientPerms,
		"Insufficient permissions",
		ErrInsufficientPermissions,
	).WithContext("userJID", userJID).WithContext("operation", operation)
}

func NewGroupAlreadyLinkedError(groupJID, communityJID string) *CommunityError {
	return NewCommunityError(
		ErrCodeGroupAlreadyLinked,
		"Group is already linked",
		ErrGroupAlreadyLinked,
	).WithContext("groupJID", groupJID).WithContext("communityJID", communityJID)
}

func NewGroupNotLinkedError(groupJID, communityJID string) *CommunityError {
	return NewCommunityError(
		ErrCodeGroupNotLinked,
		"Group is not linked to this community",
		ErrGroupNotLinked,
	).WithContext("groupJID", groupJID).WithContext("communityJID", communityJID)
}

func NewValidationError(field, reason string) *CommunityError {
	return NewCommunityError(
		ErrCodeValidationFailed,
		"Validation failed",
		errors.New("validation failed"),
	).WithContext("field", field).WithDetails(reason)
}

func NewOperationFailedError(operation string, cause error) *CommunityError {
	return NewCommunityError(
		ErrCodeOperationFailed,
		"Operation failed",
		cause,
	).WithContext("operation", operation)
}

func NewAPIUnavailableError(cause error) *CommunityError {
	return NewCommunityError(
		ErrCodeAPIUnavailable,
		"Community API is unavailable",
		cause,
	)
}

func NewTimeoutError(operation string) *CommunityError {
	return NewCommunityError(
		ErrCodeTimeout,
		"Operation timed out",
		ErrCommunityTimeout,
	).WithContext("operation", operation)
}

func NewRateLimitedError() *CommunityError {
	return NewCommunityError(
		ErrCodeRateLimited,
		"Operation rate limited",
		ErrCommunityRateLimited,
	)
}

func IsCommunityError(err error) bool {
	communityError := &CommunityError{}
	ok := errors.As(err, &communityError)
	return ok
}

func GetCommunityError(err error) (*CommunityError, bool) {
	communityErr := &CommunityError{}
	ok := errors.As(err, &communityErr)
	return communityErr, ok
}
