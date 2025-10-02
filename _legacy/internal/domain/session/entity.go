package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	CreatedAt       time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time    `json:"updatedAt" db:"updated_at"`
	ConnectionError *string      `json:"connectionError,omitempty" db:"connection_error"`
	QRCodeExpiresAt *time.Time   `json:"qrCodeExpiresAt,omitempty" db:"qr_code_expires_at"`
	ProxyConfig     *ProxyConfig `json:"proxyConfig,omitempty"`
	ConnectedAt     *time.Time   `json:"connectedAt,omitempty" db:"connected_at"`
	LastSeen        *time.Time   `json:"lastSeen,omitempty" db:"last_seen"`
	Name            string       `json:"name" db:"name"`
	DeviceJid       string       `json:"deviceJid" db:"device_jid"`
	QRCode          string       `json:"qrCode,omitempty" db:"qr_code"`
	ID              uuid.UUID    `json:"id" db:"id"`
	IsConnected     bool         `json:"isConnected" db:"is_connected"`
}

type SessionInfo struct {
	Session    *Session    `json:"session"`
	DeviceInfo *DeviceInfo `json:"deviceInfo,omitempty"`
}

type DeviceInfo struct {
	Platform    string `json:"platform"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	AppVersion  string `json:"app_version"`
}

const (
	StatusCreated      = "created"
	StatusConnecting   = "connecting"
	StatusConnected    = "connected"
	StatusDisconnected = "disconnected"
	StatusError        = "error"
	StatusLoggedOut    = "logged_out"
)

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrInvalidSessionStatus = errors.New("invalid session status")
	ErrSessionNotConnected  = errors.New("session not connected")
)

// @name ProxyConfig
type ProxyConfig struct {
	Type     string `json:"type" db:"proxy_type" example:"http"`
	Host     string `json:"host" db:"proxy_host" example:"proxy.example.com"`
	Username string `json:"username,omitempty" db:"proxy_username" example:"user"`
	Password string `json:"password,omitempty" db:"proxy_password" example:"password"`
	Port     int    `json:"port" db:"proxy_port" example:"8080"`
}

type CreateSessionRequest struct {
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	Name        string       `json:"name" validate:"required,min=1,max=100"`
	QrCode      bool         `json:"qrCode"`
}

type UpdateSessionRequest struct {
	Name        *string      `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
}

type ListSessionsRequest struct {
	IsConnected *bool   `json:"isConnected,omitempty" query:"isConnected"`
	DeviceJid   *string `json:"deviceJid,omitempty" query:"deviceJid"`
	Limit       int     `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100"`
	Offset      int     `json:"offset,omitempty" query:"offset" validate:"omitempty,min=0"`
}

type PairPhoneRequest struct {
	PhoneNumber string `json:"phoneNumber" validate:"required,e164"`
}

type QRCodeResponse struct {
	ExpiresAt   time.Time `json:"expires_at"`
	QRCode      string    `json:"qr_code"`
	QRCodeImage string    `json:"qr_code_image,omitempty"`
	Timeout     int       `json:"timeout_seconds"`
}

func NewSession(name string) *Session {
	return &Session{
		ID:          uuid.New(),
		Name:        name,
		IsConnected: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (s *Session) SetConnected(connected bool) {
	s.IsConnected = connected
	s.UpdatedAt = time.Now()

	if connected {
		now := time.Now()
		s.ConnectedAt = &now
		s.LastSeen = &now
		s.ConnectionError = nil
	}
}

func (s *Session) SetDeviceJid(deviceJid string) {
	s.DeviceJid = deviceJid
	s.UpdatedAt = time.Now()
}

func (s *Session) SetConnectionError(errorMsg string) {
	s.IsConnected = false
	s.ConnectionError = &errorMsg
	s.UpdatedAt = time.Now()
}

func (s *Session) IsActive() bool {
	return s.IsConnected
}

func (s *Session) CanConnect() bool {
	return true
}

func (s *Session) CanLogout() bool {
	return s.IsConnected
}

func (s *Session) UpdateLastSeen() {
	now := time.Now()
	s.LastSeen = &now
	s.UpdatedAt = now
}
