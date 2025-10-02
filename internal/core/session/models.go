package session

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)


type Session struct {
	ID              uuid.UUID    `json:"id"`
	Name            string       `json:"name"`
	DeviceJID       *string      `json:"deviceJid,omitempty"`
	IsConnected     bool         `json:"isConnected"`
	ConnectionError *string      `json:"connectionError,omitempty"`
	QRCode          *string      `json:"qrCode,omitempty"`
	QRCodeExpiresAt *time.Time   `json:"qrCodeExpiresAt,omitempty"`
	ProxyConfig     *ProxyConfig `json:"proxyConfig,omitempty"`
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
	ConnectedAt     *time.Time   `json:"connectedAt,omitempty"`
	LastSeen        *time.Time   `json:"lastSeen,omitempty"`
}


type ProxyConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}


type DeviceInfo struct {
	Platform    string `json:"platform"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	AppVersion  string `json:"app_version"`
}


type QRCodeResponse struct {
	QRCode      string    `json:"qr_code"`
	QRCodeImage string    `json:"qr_code_image,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	Timeout     int       `json:"timeout_seconds"`
}


type SessionStatus string

const (
	StatusCreated      SessionStatus = "created"
	StatusConnecting   SessionStatus = "connecting"
	StatusConnected    SessionStatus = "connected"
	StatusDisconnected SessionStatus = "disconnected"
	StatusError        SessionStatus = "error"
	StatusLoggedOut    SessionStatus = "logged_out"
)


func NewSession(name string) *Session {
	now := time.Now()
	return &Session{
		ID:          uuid.New(),
		Name:        name,
		IsConnected: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}


func (s *Session) UpdateConnectionStatus(connected bool) {
	s.IsConnected = connected
	s.UpdatedAt = time.Now()

	if connected {
		now := time.Now()
		s.ConnectedAt = &now
		s.LastSeen = &now
		s.ConnectionError = nil
	}
}


func (s *Session) SetConnectionError(err string) {
	s.ConnectionError = &err
	s.IsConnected = false
	s.UpdatedAt = time.Now()
}


func (s *Session) SetQRCode(qrCode string, expiresAt time.Time) {
	s.QRCode = &qrCode
	s.QRCodeExpiresAt = &expiresAt
	s.UpdatedAt = time.Now()
}


func (s *Session) ClearQRCode() {
	s.QRCode = nil
	s.QRCodeExpiresAt = nil
	s.UpdatedAt = time.Now()
}


func (s *Session) UpdateLastSeen() {
	now := time.Now()
	s.LastSeen = &now
	s.UpdatedAt = now
}


func (s *Session) IsQRCodeExpired() bool {
	if s.QRCodeExpiresAt == nil {
		return true
	}
	return time.Now().After(*s.QRCodeExpiresAt)
}


func (s *Session) GetStatus() SessionStatus {
	if s.IsConnected {
		return StatusConnected
	}

	if s.ConnectionError != nil {
		return StatusError
	}

	if s.QRCode != nil && !s.IsQRCodeExpired() {
		return StatusConnecting
	}

	if s.ConnectedAt != nil {
		return StatusDisconnected
	}

	return StatusCreated
}


func (s *Session) Validate() error {
	if s.Name == "" {
		return ErrInvalidSessionName
	}

	if len(s.Name) > 100 {
		return ErrSessionNameTooLong
	}

	return nil
}


func (p *ProxyConfig) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}


func (p *ProxyConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}