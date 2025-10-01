package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"zpwoot/internal/core/session"
	"zpwoot/internal/services/shared/dto"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)

// SessionService implementa a camada de aplicação para sessões
// Responsável por orquestrar entre core business logic e adapters externos
type SessionService struct {
	// Core business logic
	coreService *session.Service

	// External dependencies (injected via interfaces)
	repository session.Repository
	gateway    session.WhatsAppGateway
	qrGen      session.QRCodeGenerator

	// Platform dependencies
	logger    *logger.Logger
	validator *validation.Validator
}

// NewSessionService cria nova instância do serviço de aplicação
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

// CreateSession cria uma nova sessão com validação e orquestração
func (s *SessionService) CreateSession(ctx context.Context, req *dto.CreateSessionRequest) (*dto.CreateSessionResponse, error) {
	// Log da operação
	s.logger.InfoWithFields("Creating session", map[string]interface{}{
		"name":     req.Name,
		"qr_code":  req.QRCode,
		"has_proxy": req.ProxyConfig != nil,
	})

	// Validar entrada
	if err := s.validator.ValidateStruct(req); err != nil {
		s.logger.WarnWithFields("Invalid create session request", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Converter DTO para request do core
	coreReq := &session.CreateSessionRequest{
		Name:        req.Name,
		AutoConnect: req.QRCode, // Se QR code solicitado, auto-conectar
	}

	// Converter proxy config se fornecido
	if req.ProxyConfig != nil {
		coreReq.ProxyConfig = &session.ProxyConfig{
			Type:     req.ProxyConfig.Type,
			Host:     req.ProxyConfig.Host,
			Port:     req.ProxyConfig.Port,
			Username: req.ProxyConfig.Username,
			Password: req.ProxyConfig.Password,
		}
	}

	// Executar lógica de negócio no core
	sess, err := s.coreService.CreateSession(ctx, coreReq)
	if err != nil {
		s.logger.ErrorWithFields("Failed to create session", map[string]interface{}{
			"name":  req.Name,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Converter entidade para DTO de resposta
	response := &dto.CreateSessionResponse{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		CreatedAt:   sess.CreatedAt,
	}

	// Adicionar proxy config se presente
	if sess.ProxyConfig != nil {
		response.ProxyConfig = &dto.ProxyConfig{
			Type:     sess.ProxyConfig.Type,
			Host:     sess.ProxyConfig.Host,
			Port:     sess.ProxyConfig.Port,
			Username: sess.ProxyConfig.Username,
			Password: sess.ProxyConfig.Password,
		}
	}

	// Se QR code foi solicitado, tentar obter
	if req.QRCode {
		// Aguardar um pouco para o gateway processar
		time.Sleep(500 * time.Millisecond)

		qrResponse, err := s.coreService.GetQRCode(ctx, sess.ID)
		if err == nil && qrResponse != nil {
			response.QRCode = qrResponse.QRCode
			response.QRCodeImage = qrResponse.QRCode // TODO: Implementar geração de imagem
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

// GetSession busca informações de uma sessão
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*dto.SessionInfoResponse, error) {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	// Buscar no core
	sess, err := s.coreService.GetSession(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get session", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Converter para DTO
	response := &dto.SessionInfoResponse{
		Session: s.sessionToDTO(sess),
	}

	// TODO: Adicionar device info se disponível
	// if deviceInfo, err := s.gateway.GetDeviceInfo(ctx, sess.Name); err == nil {
	//     response.DeviceInfo = s.deviceInfoToDTO(deviceInfo)
	// }

	return response, nil
}

// ListSessions lista sessões com paginação
func (s *SessionService) ListSessions(ctx context.Context, req *dto.ListSessionsRequest) (*dto.ListSessionsResponse, error) {
	// Validar entrada
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Definir valores padrão
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

	// Buscar no core
	sessions, err := s.coreService.ListSessions(ctx, limit, offset)
	if err != nil {
		s.logger.ErrorWithFields("Failed to list sessions", map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"error":  err.Error(),
		})
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Converter para DTOs
	sessionResponses := make([]dto.SessionInfoResponse, len(sessions))
	for i, sess := range sessions {
		sessionResponses[i] = dto.SessionInfoResponse{
			Session: s.sessionToDTO(sess),
		}
	}

	// TODO: Obter total count do repositório
	total := len(sessions) // Placeholder

	response := &dto.ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	return response, nil
}

// ConnectSession inicia conexão de uma sessão
func (s *SessionService) ConnectSession(ctx context.Context, sessionID string) (*dto.ConnectSessionResponse, error) {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Connecting session", map[string]interface{}{
		"session_id": sessionID,
	})

	// Executar conexão no core
	err = s.coreService.ConnectSession(ctx, id)

	response := &dto.ConnectSessionResponse{
		Success: true,
	}

	// Tratar diferentes tipos de erro
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

	// Tentar obter QR code
	qrResponse, qrErr := s.coreService.GetQRCode(ctx, id)
	if qrErr == nil && qrResponse != nil {
		response.QRCode = qrResponse.QRCode
		response.QRCodeImage = qrResponse.QRCode // TODO: Implementar geração de imagem

		if err != nil && response.Message == "Session is already connected and active" {
			response.Message = "Session is connected"
		} else {
			response.Message = "QR code generated - scan with WhatsApp to connect"
		}
	}

	s.logger.InfoWithFields("Session connection processed", map[string]interface{}{
		"session_id": sessionID,
		"success":    response.Success,
		"has_qr":     response.QRCode != "",
	})

	return response, nil
}

// DisconnectSession desconecta uma sessão
func (s *SessionService) DisconnectSession(ctx context.Context, sessionID string) error {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Disconnecting session", map[string]interface{}{
		"session_id": sessionID,
	})

	// Executar no core
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

// DeleteSession remove uma sessão
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Deleting session", map[string]interface{}{
		"session_id": sessionID,
	})

	// Executar no core
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

// GetQRCode obtém QR code de uma sessão
func (s *SessionService) GetQRCode(ctx context.Context, sessionID string) (*dto.QRCodeResponse, error) {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	// Buscar no core
	qrResponse, err := s.coreService.GetQRCode(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}

	// Converter para DTO
	response := &dto.QRCodeResponse{
		QRCode:    qrResponse.QRCode,
		ExpiresAt: qrResponse.ExpiresAt,
		Timeout:   qrResponse.Timeout,
	}

	// TODO: Gerar imagem do QR code
	// response.QRCodeImage = s.qrGen.GenerateQRCodeImage(qrResponse.QRCode)

	return response, nil
}

// GenerateQRCode gera novo QR code para uma sessão
func (s *SessionService) GenerateQRCode(ctx context.Context, sessionID string) (*dto.QRCodeResponse, error) {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	s.logger.InfoWithFields("Generating QR code", map[string]interface{}{
		"session_id": sessionID,
	})

	// Gerar no core
	qrResponse, err := s.coreService.GenerateQRCode(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to generate QR code", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Converter para DTO
	response := &dto.QRCodeResponse{
		QRCode:    qrResponse.QRCode,
		ExpiresAt: qrResponse.ExpiresAt,
		Timeout:   qrResponse.Timeout,
	}

	// TODO: Gerar imagem do QR code
	// response.QRCodeImage = s.qrGen.GenerateQRCodeImage(qrResponse.QRCode)

	s.logger.InfoWithFields("QR code generated successfully", map[string]interface{}{
		"session_id": sessionID,
		"expires_at": qrResponse.ExpiresAt,
	})

	return response, nil
}

// SetProxy configura proxy para uma sessão
func (s *SessionService) SetProxy(ctx context.Context, sessionID string, req *dto.SetProxyRequest) error {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	// Validar entrada
	if err := s.validator.ValidateStruct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	s.logger.InfoWithFields("Setting proxy for session", map[string]interface{}{
		"session_id": sessionID,
		"proxy_type": req.ProxyConfig.Type,
		"proxy_host": req.ProxyConfig.Host,
	})

	// Converter DTO para entidade do core
	proxyConfig := &session.ProxyConfig{
		Type:     req.ProxyConfig.Type,
		Host:     req.ProxyConfig.Host,
		Port:     req.ProxyConfig.Port,
		Username: req.ProxyConfig.Username,
		Password: req.ProxyConfig.Password,
	}

	// Executar no core
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

// GetProxy obtém configuração de proxy de uma sessão
func (s *SessionService) GetProxy(ctx context.Context, sessionID string) (*dto.ProxyResponse, error) {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	// Buscar no core
	proxyConfig, err := s.coreService.GetProxy(ctx, id)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get proxy", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get proxy: %w", err)
	}

	response := &dto.ProxyResponse{}

	// Converter se proxy existe
	if proxyConfig != nil {
		response.ProxyConfig = &dto.ProxyConfig{
			Type:     proxyConfig.Type,
			Host:     proxyConfig.Host,
			Port:     proxyConfig.Port,
			Username: proxyConfig.Username,
			Password: proxyConfig.Password,
		}
	}

	return response, nil
}

// GetSessionStats obtém estatísticas das sessões
func (s *SessionService) GetSessionStats(ctx context.Context) (*dto.SessionStatsResponse, error) {
	// Buscar no core
	stats, err := s.coreService.GetSessionStats(ctx)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get session stats", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to get session stats: %w", err)
	}

	// Converter para DTO
	response := &dto.SessionStatsResponse{
		Total:     stats.Total,
		Connected: stats.Connected,
		Offline:   stats.Offline,
	}

	return response, nil
}

// UpdateLastSeen atualiza timestamp de último acesso
func (s *SessionService) UpdateLastSeen(ctx context.Context, sessionID string) error {
	// Validar UUID
	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	// Executar no core
	if err := s.coreService.UpdateLastSeen(ctx, id); err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

// ===== MÉTODOS AUXILIARES PRIVADOS =====

// sessionToDTO converte entidade Session para DTO
func (s *SessionService) sessionToDTO(sess *session.Session) *dto.SessionResponse {
	response := &dto.SessionResponse{
		ID:          sess.ID.String(),
		Name:        sess.Name,
		IsConnected: sess.IsConnected,
		CreatedAt:   sess.CreatedAt,
		UpdatedAt:   sess.UpdatedAt,
	}

	// Campos opcionais
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
		response.ProxyConfig = &dto.ProxyConfig{
			Type:     sess.ProxyConfig.Type,
			Host:     sess.ProxyConfig.Host,
			Port:     sess.ProxyConfig.Port,
			Username: sess.ProxyConfig.Username,
			Password: sess.ProxyConfig.Password,
		}
	}

	return response
}

// deviceInfoToDTO converte DeviceInfo para DTO (placeholder)
// func (s *SessionService) deviceInfoToDTO(deviceInfo *session.DeviceInfo) *dto.DeviceInfoResponse {
//     return &dto.DeviceInfoResponse{
//         Platform:    deviceInfo.Platform,
//         DeviceModel: deviceInfo.DeviceModel,
//         OSVersion:   deviceInfo.OSVersion,
//         AppVersion:  deviceInfo.AppVersion,
//     }
// }