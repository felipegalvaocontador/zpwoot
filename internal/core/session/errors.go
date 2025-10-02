package session

import "errors"


var (

	ErrInvalidSessionName   = errors.New("session name is required")
	ErrSessionNameTooLong   = errors.New("session name is too long (max 100 characters)")
	ErrInvalidDeviceJID     = errors.New("invalid device JID format")
	ErrInvalidProxyConfig   = errors.New("invalid proxy configuration")


	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session with this name already exists")
	ErrSessionNotConnected  = errors.New("session is not connected")
	ErrSessionAlreadyConnected = errors.New("session is already connected")


	ErrConnectionFailed     = errors.New("failed to connect to WhatsApp")
	ErrQRCodeExpired        = errors.New("QR code has expired")
	ErrQRCodeNotAvailable   = errors.New("QR code is not available")
	ErrPairingFailed        = errors.New("device pairing failed")
	ErrLogoutFailed         = errors.New("failed to logout session")


	ErrSessionBusy          = errors.New("session is busy with another operation")
	ErrInvalidOperation     = errors.New("invalid operation for current session state")
	ErrOperationTimeout     = errors.New("operation timed out")
)