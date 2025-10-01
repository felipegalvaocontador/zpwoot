package session

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Session representa uma sessão WhatsApp
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

// ProxyConfig configuração de proxy para sessão
type ProxyConfig struct {
	Type     string `json:"type"`     // http, socks5
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// DeviceInfo informações do dispositivo conectado
type DeviceInfo struct {
	Platform    string `json:"platform"`
	DeviceModel string `json:"device_model"`
	OSVersion   string `json:"os_version"`
	AppVersion  string `json:"app_version"`
}

// QRCodeResponse resposta com QR Code para pareamento
type QRCodeResponse struct {
	QRCode      string    `json:"qr_code"`
	QRCodeImage string    `json:"qr_code_image,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	Timeout     int       `json:"timeout_seconds"`
}

// SessionStatus constantes de status da sessão
type SessionStatus string

const (
	StatusCreated      SessionStatus = "created"
	StatusConnecting   SessionStatus = "connecting"
	StatusConnected    SessionStatus = "connected"
	StatusDisconnected SessionStatus = "disconnected"
	StatusError        SessionStatus = "error"
	StatusLoggedOut    SessionStatus = "logged_out"
)

// NewSession cria uma nova sessão
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

// UpdateConnectionStatus atualiza status de conexão
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

// SetConnectionError define erro de conexão
func (s *Session) SetConnectionError(err string) {
	s.ConnectionError = &err
	s.IsConnected = false
	s.UpdatedAt = time.Now()
}

// SetQRCode define QR code para pareamento
func (s *Session) SetQRCode(qrCode string, expiresAt time.Time) {
	s.QRCode = &qrCode
	s.QRCodeExpiresAt = &expiresAt
	s.UpdatedAt = time.Now()
}

// ClearQRCode limpa QR code
func (s *Session) ClearQRCode() {
	s.QRCode = nil
	s.QRCodeExpiresAt = nil
	s.UpdatedAt = time.Now()
}

// UpdateLastSeen atualiza último acesso
func (s *Session) UpdateLastSeen() {
	now := time.Now()
	s.LastSeen = &now
	s.UpdatedAt = now
}

// IsQRCodeExpired verifica se QR code expirou
func (s *Session) IsQRCodeExpired() bool {
	if s.QRCodeExpiresAt == nil {
		return true
	}
	return time.Now().After(*s.QRCodeExpiresAt)
}

// GetStatus retorna status atual da sessão
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

// Validate valida dados da sessão
func (s *Session) Validate() error {
	if s.Name == "" {
		return ErrInvalidSessionName
	}

	if len(s.Name) > 100 {
		return ErrSessionNameTooLong
	}

	return nil
}

// ToJSON converte ProxyConfig para JSON
func (p *ProxyConfig) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON carrega ProxyConfig de JSON
func (p *ProxyConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}