package session

import (
	"context"
	"time"

	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/pkg/errors"
	"zpwoot/platform/logger"
)

type UseCase interface {
	CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error)
	ListSessions(ctx context.Context, req *ListSessionsRequest) (*ListSessionsResponse, error)
	GetSessionInfo(ctx context.Context, sessionID string) (*SessionInfoResponse, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ConnectSession(ctx context.Context, sessionID string) (*ConnectSessionResponse, error)
	LogoutSession(ctx context.Context, sessionID string) error
	GetQRCode(ctx context.Context, sessionID string) (*QRCodeResponse, error)
	PairPhone(ctx context.Context, sessionID string, req *PairPhoneRequest) error
	SetProxy(ctx context.Context, sessionID string, req *SetProxyRequest) error
	GetProxy(ctx context.Context, sessionID string) (*ProxyResponse, error)
}

type useCaseImpl struct {
	sessionRepo    ports.SessionRepository
	WameowMgr      ports.WameowManager
	sessionService *session.Service
	logger         *logger.Logger
}

func NewUseCase(
	sessionRepo ports.SessionRepository,
	wameowMgr ports.WameowManager,
	sessionService *session.Service,
	logger *logger.Logger,
) UseCase {
	return &useCaseImpl{
		sessionRepo:    sessionRepo,
		WameowMgr:      wameowMgr,
		sessionService: sessionService,
		logger:         logger,
	}
}

func (uc *useCaseImpl) CreateSession(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	domainReq := req.ToCreateSessionRequest()

	sess, err := uc.sessionService.CreateSession(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	var proxyConfig *ProxyConfig
	if sess.ProxyConfig != nil {
		proxyConfig = &ProxyConfig{
			Type:     sess.ProxyConfig.Type,
			Host:     sess.ProxyConfig.Host,
			Port:     sess.ProxyConfig.Port,
			Username: sess.ProxyConfig.Username,
			Password: sess.ProxyConfig.Password,
		}
	}

	response := &CreateSessionResponse{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		ProxyConfig: proxyConfig,
		CreatedAt:   sess.CreatedAt,
	}

	// If QR code was requested during creation, try to get it from database
	if req.QrCode {
		// Give a small delay for the QR code to be generated and saved by events
		time.Sleep(500 * time.Millisecond)

		qrResponse, err := uc.sessionService.GetQRCode(ctx, sess.ID.String())
		if err == nil && qrResponse != nil {
			response.QrCode = qrResponse.QRCodeImage
			response.Code = qrResponse.QRCode
		}
		// Don't fail creation if QR code is not ready yet
	}

	return response, nil
}

func (uc *useCaseImpl) ListSessions(ctx context.Context, req *ListSessionsRequest) (*ListSessionsResponse, error) {
	domainReq := &session.ListSessionsRequest{
		IsConnected: req.IsConnected,
		DeviceJid:   req.DeviceJid,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	if domainReq.Limit == 0 {
		domainReq.Limit = 20
	}

	sessions, total, err := uc.sessionService.ListSessions(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	sessionResponses := make([]SessionInfoResponse, len(sessions))
	for i, sess := range sessions {
		sessionInfo := &session.SessionInfo{
			Session: sess,
		}
		sessionResponses[i] = *FromSessionInfo(sessionInfo)
	}

	response := &ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    total,
		Limit:    domainReq.Limit,
		Offset:   domainReq.Offset,
	}

	return response, nil
}

func (uc *useCaseImpl) GetSessionInfo(ctx context.Context, sessionID string) (*SessionInfoResponse, error) {
	sess, err := uc.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	sessionInfo := &session.SessionInfo{
		Session: sess,
	}

	response := FromSessionInfo(sessionInfo)
	return response, nil
}

func (uc *useCaseImpl) DeleteSession(ctx context.Context, sessionID string) error {
	return uc.sessionService.DeleteSession(ctx, sessionID)
}

func (uc *useCaseImpl) ConnectSession(ctx context.Context, sessionID string) (*ConnectSessionResponse, error) {
	err := uc.sessionService.ConnectSession(ctx, sessionID)

	var response *ConnectSessionResponse

	if err != nil {
		// Check if it's an "already connected" error
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == 409 {
			response = &ConnectSessionResponse{
				Success: true,
				Message: "Session is already connected and active",
			}
		} else {
			return nil, err
		}
	} else {
		response = &ConnectSessionResponse{
			Success: true,
			Message: "Session connection initiated successfully",
		}
	}

	// Always try to get QR code if available (for both connected and connecting sessions)
	qrResponse, qrErr := uc.sessionService.GetQRCode(ctx, sessionID)
	if qrErr == nil && qrResponse != nil {
		response.QrCode = qrResponse.QRCodeImage
		response.Code = qrResponse.QRCode

		// Update message based on session state
		if err != nil && response.Message == "Session is already connected and active" {
			// Session is connected but has QR code (shouldn't happen normally)
			response.Message = "Session is connected"
		} else {
			response.Message = "QR code generated - scan with WhatsApp to connect"
		}
	}

	return response, nil
}

func (uc *useCaseImpl) LogoutSession(ctx context.Context, sessionID string) error {
	err := uc.sessionService.LogoutSession(ctx, sessionID)
	if err != nil {
		// Check if it's an "already disconnected" error
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code == 409 {
			// Return success for already disconnected sessions
			return nil
		}
		return err
	}
	return nil
}

func (uc *useCaseImpl) GetQRCode(ctx context.Context, sessionID string) (*QRCodeResponse, error) {
	qrCode, err := uc.sessionService.GetQRCode(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	response := FromQRCodeResponse(qrCode)
	return response, nil
}

func (uc *useCaseImpl) PairPhone(ctx context.Context, sessionID string, req *PairPhoneRequest) error {
	return nil
}

func (uc *useCaseImpl) SetProxy(ctx context.Context, sessionID string, req *SetProxyRequest) error {
	domainProxyConfig := &session.ProxyConfig{
		Type:     req.ProxyConfig.Type,
		Host:     req.ProxyConfig.Host,
		Port:     req.ProxyConfig.Port,
		Username: req.ProxyConfig.Username,
		Password: req.ProxyConfig.Password,
	}
	return uc.sessionService.SetProxy(ctx, sessionID, domainProxyConfig)
}

func (uc *useCaseImpl) GetProxy(ctx context.Context, sessionID string) (*ProxyResponse, error) {
	proxyConfig, err := uc.sessionService.GetProxy(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var appProxyConfig *ProxyConfig
	if proxyConfig != nil {
		appProxyConfig = &ProxyConfig{
			Type:     proxyConfig.Type,
			Host:     proxyConfig.Host,
			Port:     proxyConfig.Port,
			Username: proxyConfig.Username,
			Password: proxyConfig.Password,
		}
	}

	response := &ProxyResponse{
		ProxyConfig: appProxyConfig,
	}

	return response, nil
}
