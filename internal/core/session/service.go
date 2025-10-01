package session

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service implementa a lógica de negócio pura para sessões WhatsApp
// Esta é a camada Core da Clean Architecture - sem dependências externas
type Service struct {
	repository Repository
	gateway    WhatsAppGateway
	qrGen      QRCodeGenerator
}

// NewService cria uma nova instância do serviço de sessões
func NewService(repo Repository, gateway WhatsAppGateway, qrGen QRCodeGenerator) *Service {
	return &Service{
		repository: repo,
		gateway:    gateway,
		qrGen:      qrGen,
	}
}

// CreateSessionRequest dados para criação de sessão
type CreateSessionRequest struct {
	Name        string       `json:"name" validate:"required,min=1,max=100"`
	ProxyConfig *ProxyConfig `json:"proxyConfig,omitempty"`
	AutoConnect bool         `json:"autoConnect,omitempty"`
}

// CreateSession cria uma nova sessão WhatsApp
func (s *Service) CreateSession(ctx context.Context, req *CreateSessionRequest) (*Session, error) {
	// Validação de entrada
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Verificar se já existe sessão com este nome
	exists, err := s.repository.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists {
		return nil, ErrSessionAlreadyExists
	}

	// Criar nova sessão
	session := NewSession(req.Name)
	session.ProxyConfig = req.ProxyConfig

	// Validar sessão criada
	if err := session.Validate(); err != nil {
		return nil, err
	}

	// Persistir no repositório
	if err := s.repository.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Inicializar no gateway WhatsApp
	if err := s.gateway.CreateSession(ctx, session.Name); err != nil {
		// Tentar reverter criação no repositório
		_ = s.repository.Delete(ctx, session.ID)
		return nil, fmt.Errorf("failed to initialize WhatsApp session: %w", err)
	}

	// Auto-conectar se solicitado
	if req.AutoConnect {
		if err := s.initiateConnection(ctx, session); err != nil {
			// Log do erro mas não falha a criação
			// A conexão pode ser tentada posteriormente
		}
	}

	return session, nil
}

// GetSession busca uma sessão por ID
func (s *Service) GetSession(ctx context.Context, id uuid.UUID) (*Session, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Sincronizar status com gateway se necessário
	if err := s.syncSessionStatus(ctx, session); err != nil {
		// Log do erro mas retorna a sessão mesmo assim
	}

	return session, nil
}

// GetSessionByName busca uma sessão por nome
func (s *Service) GetSessionByName(ctx context.Context, name string) (*Session, error) {
	if name == "" {
		return nil, ErrInvalidSessionName
	}

	session, err := s.repository.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get session by name: %w", err)
	}

	// Sincronizar status com gateway se necessário
	if err := s.syncSessionStatus(ctx, session); err != nil {
		// Log do erro mas retorna a sessão mesmo assim
	}

	return session, nil
}

// ListSessions retorna lista paginada de sessões
func (s *Service) ListSessions(ctx context.Context, limit, offset int) ([]*Session, error) {
	// Validar parâmetros de paginação
	if limit <= 0 {
		limit = 20 // default
	}
	if limit > 100 {
		limit = 100 // máximo
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

// ListConnectedSessions retorna apenas sessões conectadas
func (s *Service) ListConnectedSessions(ctx context.Context) ([]*Session, error) {
	sessions, err := s.repository.ListConnected(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list connected sessions: %w", err)
	}

	return sessions, nil
}

// ConnectSession inicia conexão de uma sessão
func (s *Service) ConnectSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Verificar se já está conectada
	if session.IsConnected {
		return ErrSessionAlreadyConnected
	}

	// Verificar status no gateway
	connected, err := s.gateway.IsSessionConnected(ctx, session.Name)
	if err != nil {
		return fmt.Errorf("failed to check session status: %w", err)
	}

	if connected {
		// Atualizar status local
		session.UpdateConnectionStatus(true)
		if err := s.repository.Update(ctx, session); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}
		return ErrSessionAlreadyConnected
	}

	// Iniciar conexão
	return s.initiateConnection(ctx, session)
}

// DisconnectSession desconecta uma sessão
func (s *Service) DisconnectSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Verificar se está conectada
	if !session.IsConnected {
		return ErrSessionNotConnected
	}

	// Desconectar no gateway
	if err := s.gateway.DisconnectSession(ctx, session.Name); err != nil {
		return fmt.Errorf("failed to disconnect session: %w", err)
	}

	// Atualizar status local
	session.UpdateConnectionStatus(false)
	session.ClearQRCode()

	if err := s.repository.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session status: %w", err)
	}

	return nil
}