package session

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository interface para persistência de sessões
type Repository interface {
	// CRUD básico
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetByName(ctx context.Context, name string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Consultas específicas
	List(ctx context.Context, limit, offset int) ([]*Session, error)
	ListConnected(ctx context.Context) ([]*Session, error)
	ListByStatus(ctx context.Context, connected bool) ([]*Session, error)

	// Operações de status
	UpdateConnectionStatus(ctx context.Context, id uuid.UUID, connected bool) error
	UpdateLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error

	// QR Code operations
	UpdateQRCode(ctx context.Context, id uuid.UUID, qrCode string, expiresAt time.Time) error
	ClearQRCode(ctx context.Context, id uuid.UUID) error

	// Device operations
	UpdateDeviceJID(ctx context.Context, id uuid.UUID, deviceJID string) error

	// Verificações
	ExistsByName(ctx context.Context, name string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

// WhatsAppGateway interface para integração com WhatsApp
type WhatsAppGateway interface {
	// Gerenciamento de sessão
	CreateSession(ctx context.Context, sessionName string) error
	ConnectSession(ctx context.Context, sessionName string) error
	DisconnectSession(ctx context.Context, sessionName string) error
	DeleteSession(ctx context.Context, sessionName string) error
	RestoreSession(ctx context.Context, sessionName string) error // Restaura sessão individual
	RestoreAllSessions(ctx context.Context, sessionNames []string) error
	RegisterSessionUUID(sessionName, sessionUUID string) // Registra mapeamento nome -> UUID
	SessionExists(sessionName string) bool               // Verifica se sessão existe no gateway

	// Status e informações
	IsSessionConnected(ctx context.Context, sessionName string) (bool, error)
	GetSessionInfo(ctx context.Context, sessionName string) (*DeviceInfo, error)

	// QR Code para pareamento
	GenerateQRCode(ctx context.Context, sessionName string) (*QRCodeResponse, error)

	// Configuração de proxy
	SetProxy(ctx context.Context, sessionName string, proxy *ProxyConfig) error

	// Eventos e callbacks
	SetEventHandler(handler EventHandler)

	// Envio de mensagens
	SendTextMessage(ctx context.Context, sessionName, to, content string) (*MessageSendResult, error)
	SendMediaMessage(ctx context.Context, sessionName, to, mediaURL, caption, mediaType string) (*MessageSendResult, error)
	SendLocationMessage(ctx context.Context, sessionName, to string, latitude, longitude float64, address string) (*MessageSendResult, error)
	SendContactMessage(ctx context.Context, sessionName, to, contactName, contactPhone string) (*MessageSendResult, error)
}

// EventHandler interface para eventos do WhatsApp
type EventHandler interface {
	OnSessionConnected(sessionName string, deviceInfo *DeviceInfo)
	OnSessionDisconnected(sessionName string, reason string)
	OnQRCodeGenerated(sessionName string, qrCode string, expiresAt time.Time)
	OnConnectionError(sessionName string, err error)
	OnMessageReceived(sessionName string, message *WhatsAppMessage)
	OnMessageSent(sessionName string, messageID string, status string)
}

// WhatsAppMessage representa uma mensagem do WhatsApp
type WhatsAppMessage struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Chat      string                 `json:"chat"`
	Type      string                 `json:"type"`
	Content   string                 `json:"content,omitempty"`
	MediaURL  string                 `json:"media_url,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	FromMe    bool                   `json:"from_me"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageSendResult representa o resultado do envio de uma mensagem
type MessageSendResult struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	To        string    `json:"to"`
}

// QRCodeGenerator interface para geração de QR codes
type QRCodeGenerator interface {
	Generate(ctx context.Context, sessionName string) (*QRCodeResponse, error)
	GenerateImage(ctx context.Context, qrCode string) ([]byte, error)
	IsExpired(expiresAt time.Time) bool
}