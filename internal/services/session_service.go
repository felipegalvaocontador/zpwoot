package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/internal/core/session"
	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)



type SessionService struct {

	coreService *session.Service


	repository session.Repository
	gateway    session.WhatsAppGateway
	qrGen      session.QRCodeGenerator


	logger    *logger.Logger
	validator *validation.Validator
}


func NewSessionService(
	coreService *session.Service,
	repository session.Repository,
	gateway session.WhatsAppGateway,
	qrGen session.QRCodeGenerator,
	logger *logger.Logger,
	validator *validation.Validator,
) *SessionService {
	return &SessionService{
		coreService: coreService,
		repository:  repository,
		gateway:     gateway,
		qrGen:       qrGen,
		logger:      logger,
		validator:   validator,
	}
}


func (s *SessionService) CreateSession(ctx context.Context, req *contracts.CreateSessionRequest) (*contracts.CreateSessionResponse, error) {

	s.logger.InfoWithFields("Creating session", map[string]interface{}{
		"name":     req.Name,
		"qr_code":  req.QRCode,
		"has_proxy": req.ProxyConfig != nil,
	})


	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.WarnWithFields("Invalid create session request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}


	coreReq := &session.CreateSessionRequest{
		Name:        req.Name,
		AutoConnect: req.QRCode,
	}


	if req.ProxyConfig != nil {
		coreReq.ProxyConfig = &session.ProxyConfig{
			Type:     req.ProxyConfig.Type,
			Host:     req.ProxyConfig.Host,
			Port:     req.ProxyConfig.Port,
			Username: req.ProxyConfig.Username,
			Password: req.ProxyConfig.Password,
		}
	}


	sess, err := s.coreService.CreateSession(ctx, coreReq)
	if err != nil {
		s.logger.ErrorWithFields("Failed to create session", map[string]interface{}{
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create session: %w", err)
	}


	response := &contracts.CreateSessionResponse{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		CreatedAt:   sess.CreatedAt,
	}


	if sess.ProxyConfig != nil {
		response.ProxyConfig = &contracts.ProxyConfig{
			Type:     sess.ProxyConfig.Type,
			Host:     sess.ProxyConfig.Host,
			Port:     sess.ProxyConfig.Port,
			Username: sess.ProxyConfig.Username,
			Password: sess.ProxyConfig.Password,
		}
	}


	if req.QRCode {

		time.Sleep(500 * time.Millisecond)

		qrResponse, err := s.coreService.GetQRCode(ctx, sess.ID)
		if err == nil && qrResponse != nil {
			response.QRCode = qrResponse.QRCode
			response.QRCodeImage = qrResponse.QRCode
		} else {
			s.logger.WarnWithFields("Failed to get QR code after session creation", map[string]interface{}{
				"session_id": sess.ID.String(),
				"error":      err.Error(),
			})
		}
	}

	s.logger.InfoWithFields("Session created successfully", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"name":         sess.Name,
		"is_connected": sess.IsConnected,
		"has_qr_code":  response.QRCode != "",
	})

	return response, nil
}


func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*contracts.SessionInfoResponse, error) {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}


	sess, err := s.coreService.GetSession(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get session", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get session: %w", err)
	}


	response := &contracts.SessionInfoResponse{
		Session: s.sessionToDTO(sess),
	}






	return response, nil
}


func (s *SessionService) ResolveSessionID(ctx context.Context, idOrName string) (uuid.UUID, error) {

	if id, err := uuid.Parse(idOrName); err == nil {

		_, err := s.coreService.GetSession(ctx, id)
		if err != nil {
			return uuid.Nil, fmt.Errorf("session with ID %s not found: %w", idOrName, err)
		}
		return id, nil
	}


	sess, err := s.coreService.GetSessionByName(ctx, idOrName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("session with name '%s' not found: %w", idOrName, err)
	}

	return sess.ID, nil
}


func (s *SessionService) RestoreAllSessions(ctx context.Context) error {
	s.logger.Info("Starting session restoration process")


	sessions, err := s.coreService.ListSessions(ctx, 1000, 0)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get sessions for restoration", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	if len(sessions) == 0 {
		s.logger.Info("No sessions found to restore")
		return nil
	}


	for _, sess := range sessions {
		s.gateway.RegisterSessionUUID(sess.Name, sess.ID.String())
	}


	sessionNames := make([]string, len(sessions))
	for i, sess := range sessions {
		sessionNames[i] = sess.Name
	}


	err = s.gateway.RestoreAllSessions(ctx, sessionNames)
	if err != nil {
		s.logger.ErrorWithFields("Failed to restore sessions in gateway", map[string]interface{}{
			"session_count": len(sessionNames),
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to restore sessions: %w", err)
	}

	s.logger.InfoWithFields("Session restoration completed successfully", map[string]interface{}{
		"restored_sessions": len(sessionNames),
	})

	return nil
}


func (s *SessionService) DeleteSessionByNameOrID(ctx context.Context, idOrName string) error {

	sessionID, err := s.ResolveSessionID(ctx, idOrName)
	if err != nil {
		return err
	}


	return s.DeleteSession(ctx, sessionID.String())
}


func (s *SessionService) GetSessionByNameOrID(ctx context.Context, identifier string) (*contracts.SessionInfoResponse, error) {

	if id, err := uuid.Parse(identifier); err == nil {
		return s.GetSession(ctx, id.String())
	}


	sess, err := s.coreService.GetSessionByName(ctx, identifier)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get session by name", map[string]interface{}{
			"session_name": identifier,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to get session: %w", err)
	}


	response := &contracts.SessionInfoResponse{
		Session: s.sessionToDTO(sess),
	}

	return response, nil
}


func (s *SessionService) ListSessions(ctx context.Context, req *contracts.ListSessionsRequest) (*contracts.ListSessionsResponse, error) {

	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}


	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}


	sessions, err := s.coreService.ListSessions(ctx, limit, offset)
	if err != nil {
		s.logger.ErrorWithFields("Failed to list sessions", map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}


	sessionResponses := make([]contracts.SessionInfoResponse, len(sessions))
	for i, sess := range sessions {
		sessionResponses[i] = contracts.SessionInfoResponse{
			Session: s.sessionToDTO(sess),
		}
	}


	total := len(sessions)

	response := &contracts.ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	return response, nil
}


func (s *SessionService) ConnectSession(ctx context.Context, sessionID string) (*contracts.ConnectSessionResponse, error) {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}




	err = s.coreService.ConnectSession(ctx, id)

	response := &contracts.ConnectSessionResponse{
		Success: true,
	}


	if err != nil {
		if err == session.ErrSessionAlreadyConnected {
			response.Message = "Session is already connected and active"
		} else {
			s.logger.ErrorWithFields("Failed to connect session", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			return nil, fmt.Errorf("failed to connect session: %w", err)
		}
	} else {
		response.Message = "Session connection initiated successfully"
	}


	qrResponse, qrErr := s.coreService.GetQRCode(ctx, id)
	if qrErr == nil && qrResponse != nil {
		response.QRCode = qrResponse.QRCode
		response.QRCodeImage = qrResponse.QRCode

		if err != nil && response.Message == "Session is already connected and active" {
			response.Message = "Session is connected"
		} else {
			response.Message = "QR code generated - scan with WhatsApp to connect"
		}
	}



	return response, nil
}


func (s *SessionService) DisconnectSession(ctx context.Context, sessionID string) error {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Disconnecting session", map[string]interface{}{
		"session_id": sessionID,
	})


	if err := s.coreService.DisconnectSession(ctx, id); err != nil {
		s.logger.ErrorWithFields("Failed to disconnect session", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to disconnect session: %w", err)
	}

	s.logger.InfoWithFields("Session disconnected successfully", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}


func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Deleting session", map[string]interface{}{
		"session_id": sessionID,
	})


	if err := s.coreService.DeleteSession(ctx, id); err != nil {
		s.logger.ErrorWithFields("Failed to delete session", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to delete session: %w", err)
	}

	s.logger.InfoWithFields("Session deleted successfully", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}


func (s *SessionService) GetQRCode(ctx context.Context, sessionID string) (*contracts.QRCodeResponse, error) {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}


	qrResponse, err := s.coreService.GetQRCode(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}


	response := &contracts.QRCodeResponse{
		QRCode:    qrResponse.QRCode,
		ExpiresAt: qrResponse.ExpiresAt,
		Timeout:   qrResponse.Timeout,
	}




	return response, nil
}


func (s *SessionService) GenerateQRCode(ctx context.Context, sessionID string) (*contracts.QRCodeResponse, error) {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Generating QR code", map[string]interface{}{
		"session_id": sessionID,
	})


	qrResponse, err := s.coreService.GenerateQRCode(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to generate QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}


	response := &contracts.QRCodeResponse{
		QRCode:    qrResponse.QRCode,
		ExpiresAt: qrResponse.ExpiresAt,
		Timeout:   qrResponse.Timeout,
	}




	s.logger.InfoWithFields("QR code generated successfully", map[string]interface{}{
		"session_id": sessionID,
		"expires_at": qrResponse.ExpiresAt,
	})

	return response, nil
}


func (s *SessionService) SetProxy(ctx context.Context, sessionID string, req *contracts.SetProxyRequest) error {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}


	if err := s.validator.ValidateStruct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	s.logger.InfoWithFields("Setting proxy for session", map[string]interface{}{
		"session_id": sessionID,
		"proxy_type": req.ProxyConfig.Type,
		"proxy_host": req.ProxyConfig.Host,
	})


	proxyConfig := &session.ProxyConfig{
		Type:     req.ProxyConfig.Type,
		Host:     req.ProxyConfig.Host,
		Port:     req.ProxyConfig.Port,
		Username: req.ProxyConfig.Username,
		Password: req.ProxyConfig.Password,
	}


	if err := s.coreService.SetProxy(ctx, id, proxyConfig); err != nil {
		s.logger.ErrorWithFields("Failed to set proxy", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	s.logger.InfoWithFields("Proxy set successfully", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}


func (s *SessionService) GetProxy(ctx context.Context, sessionID string) (*contracts.ProxyResponse, error) {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}


	proxyConfig, err := s.coreService.GetProxy(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get proxy", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get proxy: %w", err)
	}

	response := &contracts.ProxyResponse{}


	if proxyConfig != nil {
		response.ProxyConfig = &contracts.ProxyConfig{
			Type:     proxyConfig.Type,
			Host:     proxyConfig.Host,
			Port:     proxyConfig.Port,
			Username: proxyConfig.Username,
			Password: proxyConfig.Password,
		}
	}

	return response, nil
}


func (s *SessionService) GetSessionStats(ctx context.Context) (*contracts.SessionStatsResponse, error) {

	stats, err := s.coreService.GetSessionStats(ctx)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get session stats", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}


	response := &contracts.SessionStatsResponse{
		Total:     stats.Total,
		Connected: stats.Connected,
		Offline:   stats.Offline,
	}

	return response, nil
}


func (s *SessionService) UpdateLastSeen(ctx context.Context, sessionID string) error {

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}


	if err := s.coreService.UpdateLastSeen(ctx, id); err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}




func (s *SessionService) sessionToDTO(sess *session.Session) *contracts.SessionResponse {
	response := &contracts.SessionResponse{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		CreatedAt:   sess.CreatedAt,
		UpdatedAt:   sess.UpdatedAt,
	}


	if sess.DeviceJID != nil {
		response.DeviceJID = *sess.DeviceJID
	}

	if sess.ConnectionError != nil {
		response.ConnectionError = sess.ConnectionError
	}

	if sess.ConnectedAt != nil {
		response.ConnectedAt = sess.ConnectedAt
	}

	if sess.ProxyConfig != nil {
		response.ProxyConfig = &contracts.ProxyConfig{
			Type:     sess.ProxyConfig.Type,
			Host:     sess.ProxyConfig.Host,
			Port:     sess.ProxyConfig.Port,
			Username: sess.ProxyConfig.Username,
			Password: sess.ProxyConfig.Password,
		}
	}

	return response
}









