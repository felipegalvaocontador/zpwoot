package wameow

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sync"
	"time"

	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
)

type ConnectionManager struct {
	logger *logger.Logger
}

func NewConnectionManager(logger *logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		logger: logger,
	}
}

func (c *ConnectionManager) SafeConnect(client *whatsmeow.Client, sessionID string) error {
	if err := ValidateClientAndStore(client, sessionID); err != nil {
		return newConnectionError(sessionID, "connect", err)
	}

	if client.IsConnected() {
		c.logger.InfoWithFields("Client already connected", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil
	}

	c.logger.InfoWithFields("Connecting client", map[string]interface{}{
		"session_id": sessionID,
	})

	err := client.Connect()
	if err != nil {
		return newConnectionError(sessionID, "connect", err)
	}

	return nil
}

func (c *ConnectionManager) SafeDisconnect(client *whatsmeow.Client, sessionID string) {
	if client == nil {
		c.logger.WarnWithFields("Cannot disconnect nil client", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	if !client.IsConnected() {
		c.logger.InfoWithFields("Client already disconnected", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	c.logger.InfoWithFields("Disconnecting client", map[string]interface{}{
		"session_id": sessionID,
	})
	client.Disconnect()
}

func (c *ConnectionManager) ConnectWithRetry(client *whatsmeow.Client, sessionID string, config *RetryConfig) error {
	if config == nil {
		config = defaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.InfoWithFields("Retry attempt", map[string]interface{}{
				"session_id":  sessionID,
				"attempt":     attempt,
				"max_retries": config.MaxRetries,
			})
			time.Sleep(config.RetryInterval)
		}

		err := c.SafeConnect(client, sessionID)
		if err == nil {
			c.logger.InfoWithFields("Successfully connected", map[string]interface{}{
				"session_id": sessionID,
				"attempt":    attempt + 1,
			})
			return nil
		}

		lastErr = err
		c.logger.WarnWithFields("Connection attempt failed", map[string]interface{}{
			"session_id": sessionID,
			"attempt":    attempt + 1,
			"error":      err.Error(),
		})
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", config.MaxRetries+1, lastErr)
}

type QRCodeGenerator struct {
	logger     *logger.Logger
	lastQRCode string
	mu         sync.Mutex
}

func NewQRCodeGenerator(logger *logger.Logger) *QRCodeGenerator {
	return &QRCodeGenerator{
		logger: logger,
	}
}

func (q *QRCodeGenerator) GenerateQRCodeImage(qrText string) string {
	if qrText == "" {
		q.logger.Warn("Empty QR text provided")
		return ""
	}

	png, err := qrcode.Encode(qrText, qrcode.Medium, 256)
	if err != nil {
		q.logger.ErrorWithFields("Failed to generate QR code", map[string]interface{}{
			"error": err.Error(),
		})
		return ""
	}

	base64String := base64.StdEncoding.EncodeToString(png)
	return "data:image/png;base64," + base64String
}

func (q *QRCodeGenerator) DisplayQRCodeInTerminal(qrCode, sessionID string) {
	if qrCode == "" {
		q.logger.WarnWithFields("Empty QR code", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	// Protege a saída do terminal contra concorrência
	q.mu.Lock()
	defer q.mu.Unlock()

	// Evita exibir o mesmo QR code repetidamente
	if q.lastQRCode == qrCode {
		q.logger.DebugWithFields("Skipping duplicate QR code display", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	// Log QR code string for debugging (separate from terminal display)
	q.logger.InfoWithFields("QR code generated", map[string]interface{}{
		"session_id": sessionID,
		"qr_code":    qrCode, // Log the actual QR string for debugging
	})

	// Clear terminal before displaying new QR code (WhatsApp renews every 20-30 seconds)
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top

	// Display QR code using HalfBlock for better terminal compatibility
	fmt.Println("Scan this QR code with WhatsApp:")
	qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)

	// Armazena o QR code atual para evitar duplicates
	q.lastQRCode = qrCode
}

// sessionManager implements SessionUpdater interface
type sessionManager struct {
	sessionRepo ports.SessionRepository
	logger      *logger.Logger
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionRepo ports.SessionRepository, logger *logger.Logger) SessionUpdater {
	return &sessionManager{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (s *sessionManager) UpdateConnectionStatus(sessionID string, isConnected bool) {
	if s.sessionRepo == nil {
		s.logger.WarnWithFields("No session repository available", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionEntity, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		// Check if it's a "not found" error (session may have been deleted)
		if err.Error() == "session not found" {
			s.logger.InfoWithFields("Session not found during status update (may have been deleted)", map[string]interface{}{
				"session_id": sessionID,
			})
		} else {
			s.logger.ErrorWithFields("Failed to get session", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
		return
	}

	if sessionEntity == nil {
		s.logger.InfoWithFields("Session entity is nil (may have been deleted)", map[string]interface{}{
			"session_id": sessionID,
		})
		return
	}

	currentConnectionStatus := sessionEntity.IsConnected
	if currentConnectionStatus == isConnected {
		return
	}

	sessionEntity.SetConnected(isConnected)

	if isConnected {
		sessionEntity.QRCode = ""
		sessionEntity.QRCodeExpiresAt = nil
	}

	if err := s.sessionRepo.Update(ctx, sessionEntity); err != nil {
		s.logger.ErrorWithFields("Failed to update session in database", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	s.logger.InfoWithFields("Session connection status updated", map[string]interface{}{
		"session_id":   sessionID,
		"is_connected": isConnected,
	})
}

func (s *sessionManager) GetSession(sessionID string) (*session.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.sessionRepo.GetByID(ctx, sessionID)
}

func (s *sessionManager) GetSessionRepo() ports.SessionRepository {
	return s.sessionRepo
}

type RetryConfig struct {
	MaxRetries    int
	RetryInterval time.Duration
}

func defaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    5,
		RetryInterval: 30 * time.Second,
	}
}
