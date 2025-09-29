package session

import (
	"context"
	"time"

	"zpwoot/pkg/errors"
	"zpwoot/pkg/uuid"
)

type Service struct {
	repo        Repository
	Wameow      WameowManager
	generator   *uuid.Generator
	qrGenerator QRGenerator
}

type QRGenerator interface {
	GenerateQRCodeImage(qrText string) string
}

type Repository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetByDeviceJid(ctx context.Context, deviceJid string) (*Session, error)
	List(ctx context.Context, req *ListSessionsRequest) ([]*Session, int, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
}

type WameowManager interface {
	CreateSession(sessionID string, config *ProxyConfig) error
	ConnectSession(sessionID string) error
	DisconnectSession(sessionID string) error
	LogoutSession(sessionID string) error
	GetQRCode(sessionID string) (*QRCodeResponse, error)
	PairPhone(sessionID, phoneNumber string) error
	IsConnected(sessionID string) bool
	GetDeviceInfo(sessionID string) (*DeviceInfo, error)
	SetProxy(sessionID string, config *ProxyConfig) error
	GetProxy(sessionID string) (*ProxyConfig, error)
}

func NewService(repo Repository, Wameow WameowManager, qrGenerator QRGenerator) *Service {
	return &Service{
		repo:        repo,
		Wameow:      Wameow,
		generator:   uuid.New(),
		qrGenerator: qrGenerator,
	}
}

func (s *Service) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {

	session := NewSession(req.Name)
	session.ProxyConfig = req.ProxyConfig

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	if err := s.Wameow.CreateSession(session.ID.String(), req.ProxyConfig); err != nil {
		return nil, errors.Wrap(err, "failed to initialize Wameow session")
	}

	// If QR code was requested, initiate connection to generate QR code
	if req.QrCode {
		// Start connection process which will generate QR code
		if err := s.Wameow.ConnectSession(session.ID.String()); err != nil {
			// Don't fail session creation if QR code generation fails
			// Just log the error
		}
	}

	return session, nil
}

func (s *Service) GetSession(ctx context.Context, id string) (*SessionInfo, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return nil, errors.ErrNotFound
	}

	info := &SessionInfo{
		Session: session,
	}

	if session.IsConnected {
		deviceInfo, _ := s.Wameow.GetDeviceInfo(id)
		info.DeviceInfo = deviceInfo
	}

	return info, nil
}

func (s *Service) ListSessions(ctx context.Context, req *ListSessionsRequest) ([]*Session, int, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}

	sessions, total, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list sessions")
	}

	return sessions, total, nil
}

func (s *Service) DeleteSession(ctx context.Context, id string) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	if session.IsActive() {
		if err := s.Wameow.DisconnectSession(id); err != nil {
			_ = err // Explicitly ignore error
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete session")
	}

	return nil
}

func (s *Service) ConnectSession(ctx context.Context, id string) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	session.SetConnected(false)   // Ensure it starts as disconnected during QR process
	session.ConnectionError = nil // Clear any previous errors
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session status to connecting")
	}

	if err := s.Wameow.ConnectSession(id); err != nil {
		session.SetConnectionError(err.Error())
		if updateErr := s.repo.Update(ctx, session); updateErr != nil {
			_ = updateErr // Explicitly ignore update error
		}
		return errors.Wrap(err, "failed to connect to Wameow")
	}

	return nil
}

func (s *Service) LogoutSession(ctx context.Context, id string) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	if !session.CanLogout() {
		return errors.NewWithDetails(400, "Cannot logout session", "Session is not connected")
	}

	if err := s.Wameow.LogoutSession(id); err != nil {
		return errors.Wrap(err, "failed to logout from Wameow")
	}

	session.SetConnected(false)
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session status")
	}

	return nil
}

func (s *Service) GetQRCode(ctx context.Context, id string) (*QRCodeResponse, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return nil, errors.ErrNotFound
	}

	// Check if QR code exists in database
	if session.QRCode == "" {
		return nil, errors.NewWithDetails(404, "QR code not found", "no QR code available for this session")
	}

	// Check if QR code is expired
	if session.QRCodeExpiresAt != nil && time.Now().After(*session.QRCodeExpiresAt) {
		return nil, errors.NewWithDetails(410, "QR code expired", "QR code has expired")
	}

	// Generate QR code image from the stored code
	qrCodeImage := s.qrGenerator.GenerateQRCodeImage(session.QRCode)

	// Calculate expiry time (default 2 minutes from now if no expiry set)
	expiresAt := time.Now().Add(2 * time.Minute)
	if session.QRCodeExpiresAt != nil {
		expiresAt = *session.QRCodeExpiresAt
	}

	return &QRCodeResponse{
		QRCode:      session.QRCode,
		QRCodeImage: qrCodeImage,
		ExpiresAt:   expiresAt,
		Timeout:     120,
	}, nil
}

func (s *Service) PairPhone(ctx context.Context, id string, req *PairPhoneRequest) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	if err := s.Wameow.PairPhone(id, req.PhoneNumber); err != nil {
		return errors.Wrap(err, "failed to pair phone")
	}

	session.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session")
	}

	return nil
}

func (s *Service) SetProxy(ctx context.Context, id string, config *ProxyConfig) error {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return errors.ErrNotFound
	}

	if err := s.Wameow.SetProxy(id, config); err != nil {
		return errors.Wrap(err, "failed to set proxy")
	}

	session.ProxyConfig = config
	session.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, session); err != nil {
		return errors.Wrap(err, "failed to update session")
	}

	return nil
}

func (s *Service) GetProxy(ctx context.Context, id string) (*ProxyConfig, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	if session == nil {
		return nil, errors.ErrNotFound
	}

	return session.ProxyConfig, nil
}
