package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
	gateway    WhatsAppGateway
	qrGen      QRCodeGenerator
}

func NewService(repo Repository, gateway WhatsAppGateway, qrGen QRCodeGenerator) *Service {
	return &Service{
		repository: repo,
		gateway:    gateway,
		qrGen:      qrGen,
	}
}

type CreateSessionRequest struct {
	Name        string       `json:"name" validate:"required,min=1,max=100"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	AutoConnect bool         `json:"autoConnect,omitempty"`
}

func (s *Service) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	exists, err := s.repository.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists {
		return nil, ErrSessionAlreadyExists
	}

	session := NewSession(req.Name)
	session.ProxyConfig = req.ProxyConfig

	if err := session.Validate(); err != nil {
		return nil, err
	}

	if err := s.repository.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := s.gateway.CreateSession(ctx, session.Name); err != nil {

		_ = s.repository.Delete(ctx, session.ID)
		return nil, fmt.Errorf("failed to initialize WhatsApp session: %w", err)
	}

	s.gateway.RegisterSessionUUID(session.Name, session.ID.String())

	if req.AutoConnect {
		if err := s.initiateConnection(ctx, session); err != nil {

		}
	}

	return session, nil
}

func (s *Service) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if err := s.syncSessionStatus(ctx, session); err != nil {

	}

	return session, nil
}

func (s *Service) GetSessionByName(ctx context.Context, name string) (*Session, error) {
	if name == "" {
		return nil, ErrInvalidSessionName
	}

	session, err := s.repository.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get session by name: %w", err)
	}

	if err := s.syncSessionStatus(ctx, session); err != nil {

	}

	return session, nil
}

func (s *Service) ListSessions(ctx context.Context, limit, offset int) ([]*Session, error) {

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	sessions, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

func (s *Service) ListConnectedSessions(ctx context.Context) ([]*Session, error) {
	sessions, err := s.repository.ListConnected(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list connected sessions: %w", err)
	}

	return sessions, nil
}

func (s *Service) GetAllSessionNames(ctx context.Context) ([]string, error) {
	sessions, err := s.repository.List(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	names := make([]string, len(sessions))
	for i, session := range sessions {
		names[i] = session.Name
	}

	return names, nil
}

func (s *Service) ConnectSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.IsConnected {
		return ErrSessionAlreadyConnected
	}

	connected, err := s.gateway.IsSessionConnected(ctx, session.Name)
	if err != nil {
		return fmt.Errorf("failed to check session status: %w", err)
	}

	if connected {

		session.UpdateConnectionStatus(true)
		if err := s.repository.Update(ctx, session); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}
		return ErrSessionAlreadyConnected
	}

	return s.initiateConnection(ctx, session)
}

func (s *Service) DisconnectSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if !session.IsConnected {
		return ErrSessionNotConnected
	}

	if err := s.gateway.DisconnectSession(ctx, session.Name); err != nil {
		return fmt.Errorf("failed to disconnect session: %w", err)
	}

	session.UpdateConnectionStatus(false)
	session.ClearQRCode()

	if err := s.repository.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	return nil
}

func (s *Service) DeleteSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.IsConnected {
		if err := s.DisconnectSession(ctx, id); err != nil {

		}
	}

	if err := s.gateway.DeleteSession(ctx, session.Name); err != nil {

	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *Service) GenerateQRCode(ctx context.Context, id uuid.UUID) (*QRCodeResponse, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.IsConnected {
		return nil, ErrSessionAlreadyConnected
	}

	if session.QRCode != nil && !session.IsQRCodeExpired() {
		return &QRCodeResponse{
			QRCode:    *session.QRCode,
			ExpiresAt: *session.QRCodeExpiresAt,
			Timeout:   120,
		}, nil
	}

	qrResponse, err := s.gateway.GenerateQRCode(ctx, session.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	session.SetQRCode(qrResponse.QRCode, qrResponse.ExpiresAt)
	if err := s.repository.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session with QR code: %w", err)
	}

	return qrResponse, nil
}

func (s *Service) GetQRCode(ctx context.Context, id uuid.UUID) (*QRCodeResponse, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.QRCode == nil {
		return nil, ErrQRCodeNotAvailable
	}

	if session.IsQRCodeExpired() {
		return nil, ErrQRCodeExpired
	}

	return &QRCodeResponse{
		QRCode:    *session.QRCode,
		ExpiresAt: *session.QRCodeExpiresAt,
		Timeout:   120,
	}, nil
}

func (s *Service) SetProxy(ctx context.Context, id uuid.UUID, proxy *ProxyConfig) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if err := s.validateProxyConfig(proxy); err != nil {
		return err
	}

	if err := s.gateway.SetProxy(ctx, session.Name, proxy); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	session.ProxyConfig = proxy
	session.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (s *Service) GetProxy(ctx context.Context, id uuid.UUID) (*ProxyConfig, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session.ProxyConfig, nil
}

func (s *Service) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := s.repository.UpdateLastSeen(ctx, id, now); err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	return nil
}

func (s *Service) GetSessionStats(ctx context.Context) (*SessionStats, error) {
	total, err := s.repository.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	connected, err := s.repository.ListConnected(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list connected sessions: %w", err)
	}

	return &SessionStats{
		Total:     int(total),
		Connected: len(connected),
		Offline:   int(total) - len(connected),
	}, nil
}

type SessionStats struct {
	Total     int `json:"total"`
	Connected int `json:"connected"`
	Offline   int `json:"offline"`
}

func (s *Service) validateCreateRequest(req *CreateSessionRequest) error {
	if req == nil {
		return fmt.Errorf("create request cannot be nil")
	}

	if req.Name == "" {
		return ErrInvalidSessionName
	}

	if len(req.Name) > 100 {
		return ErrSessionNameTooLong
	}

	if !isValidSessionName(req.Name) {
		return fmt.Errorf("session name contains invalid characters (only alphanumeric, dash and underscore allowed)")
	}

	if req.ProxyConfig != nil {
		if err := s.validateProxyConfig(req.ProxyConfig); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) validateProxyConfig(proxy *ProxyConfig) error {
	if proxy == nil {
		return nil
	}

	if proxy.Type == "" {
		return ErrInvalidProxyConfig
	}

	if proxy.Type != "http" && proxy.Type != "socks5" {
		return fmt.Errorf("invalid proxy type: %s (must be 'http' or 'socks5')", proxy.Type)
	}

	if proxy.Host == "" {
		return fmt.Errorf("proxy host is required")
	}

	if proxy.Port <= 0 || proxy.Port > 65535 {
		return fmt.Errorf("invalid proxy port: %d (must be between 1 and 65535)", proxy.Port)
	}

	return nil
}

func (s *Service) initiateConnection(ctx context.Context, session *Session) error {

	sessionExists := s.gateway.SessionExists(session.Name)
	if !sessionExists {

		s.gateway.RegisterSessionUUID(session.Name, session.ID.String())

		if err := s.gateway.RestoreSession(ctx, session.Name); err != nil {
			session.SetConnectionError(err.Error())
			_ = s.repository.Update(ctx, session)
			return fmt.Errorf("failed to restore session: %w", err)
		}
	}

	if session.ProxyConfig != nil {
		if err := s.gateway.SetProxy(ctx, session.Name, session.ProxyConfig); err != nil {
			return fmt.Errorf("failed to set proxy: %w", err)
		}
	}

	if err := s.gateway.ConnectSession(ctx, session.Name); err != nil {

		session.SetConnectionError(err.Error())
		_ = s.repository.Update(ctx, session)
		return fmt.Errorf("failed to connect session: %w", err)
	}

	return nil
}

func (s *Service) syncSessionStatus(ctx context.Context, session *Session) error {
	connected, err := s.gateway.IsSessionConnected(ctx, session.Name)
	if err != nil {
		return fmt.Errorf("failed to check session status: %w", err)
	}

	if session.IsConnected != connected {
		session.UpdateConnectionStatus(connected)
		if err := s.repository.Update(ctx, session); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}
	}

	return nil
}

func isValidSessionName(name string) bool {
	if name == "" {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

type SessionEventHandler struct {
	service *Service
}

func NewSessionEventHandler(service *Service) *SessionEventHandler {
	return &SessionEventHandler{
		service: service,
	}
}

func (h *SessionEventHandler) OnSessionConnected(sessionName string, deviceInfo *DeviceInfo) {
	ctx := context.Background()

	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.UpdateConnectionStatus(true)
	session.ConnectionError = nil
	session.ClearQRCode()

	_ = h.service.repository.Update(ctx, session)
}

func (h *SessionEventHandler) OnSessionDisconnected(sessionName string, reason string) {
	ctx := context.Background()

	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.UpdateConnectionStatus(false)
	if reason != "" {
		session.SetConnectionError(reason)
	}

	_ = h.service.repository.Update(ctx, session)
}

func (h *SessionEventHandler) OnQRCodeGenerated(sessionName string, qrCode string, expiresAt time.Time) {
	ctx := context.Background()

	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.SetQRCode(qrCode, expiresAt)
	_ = h.service.repository.Update(ctx, session)
}

func (h *SessionEventHandler) OnConnectionError(sessionName string, err error) {
	ctx := context.Background()

	session, err2 := h.service.repository.GetByName(ctx, sessionName)
	if err2 != nil {
		return
	}

	session.SetConnectionError(err.Error())
	_ = h.service.repository.Update(ctx, session)
}

func (h *SessionEventHandler) OnMessageReceived(sessionName string, message *WhatsAppMessage) {

	ctx := context.Background()
	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.UpdateLastSeen()
	_ = h.service.repository.Update(ctx, session)
}

func (s *Service) UpdateDeviceJID(ctx context.Context, id uuid.UUID, deviceJID string) error {
	return s.repository.UpdateDeviceJID(ctx, id, deviceJID)
}

func (s *Service) UpdateQRCode(ctx context.Context, id uuid.UUID, qrCode string, expiresAt time.Time) error {
	return s.repository.UpdateQRCode(ctx, id, qrCode, expiresAt)
}

func (s *Service) ClearQRCode(ctx context.Context, id uuid.UUID) error {
	return s.repository.ClearQRCode(ctx, id)
}

func (h *SessionEventHandler) OnMessageSent(sessionName string, messageID string, status string) {

	ctx := context.Background()
	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.UpdateLastSeen()
	_ = h.service.repository.Update(ctx, session)
}
