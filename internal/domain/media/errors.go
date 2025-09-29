package media

import "errors"

// Domain errors for media operations
var (
	// Validation errors
	ErrInvalidSessionID = errors.New("invalid session ID")
	ErrInvalidMessageID = errors.New("invalid message ID")
	ErrInvalidMediaType = errors.New("invalid media type")
	ErrInvalidLimit     = errors.New("invalid limit: must be between 1 and 100")
	ErrInvalidOffset    = errors.New("invalid offset: must be >= 0")
	ErrInvalidOlderThan = errors.New("invalid older_than: must be >= 0")

	// Business logic errors
	ErrSessionNotFound   = errors.New("session not found")
	ErrMessageNotFound   = errors.New("message not found")
	ErrNoMediaInMessage  = errors.New("no media in message")
	ErrMediaTypeMismatch = errors.New("media type mismatch")
	ErrMediaNotCached    = errors.New("media not cached")
	ErrCacheExpired      = errors.New("cache expired")

	// Technical errors
	ErrDownloadFailed    = errors.New("download failed")
	ErrCacheWriteFailed  = errors.New("cache write failed")
	ErrCacheReadFailed   = errors.New("cache read failed")
	ErrFileNotFound      = errors.New("file not found")
	ErrInsufficientSpace = errors.New("insufficient disk space")
	ErrFileTooLarge      = errors.New("file too large")
	ErrUnsupportedFormat = errors.New("unsupported media format")

	// WhatsApp specific errors
	ErrClientNotLoggedIn = errors.New("WhatsApp client not logged in")
	ErrMediaKeyMissing   = errors.New("media key missing")
	ErrDecryptionFailed  = errors.New("media decryption failed")
	ErrUploadFailed      = errors.New("media upload failed")
)
