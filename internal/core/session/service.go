package session

import (
	"context"
	"fmt"
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

// GetAllSessionNames retorna nomes de todas as sessões para restauração
func (s *Service) GetAllSessionNames(ctx context.Context) ([]string, error) {
	sessions, err := s.repository.List(ctx, 1000, 0) // Buscar todas as sessões
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	names := make([]string, len(sessions))
	for i, session := range sessions {
		names[i] = session.Name
	}

	return names, nil
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

// DeleteSession remove uma sessão completamente
func (s *Service) DeleteSession(ctx context.Context, id uuid.UUID) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Desconectar se estiver conectada
	if session.IsConnected {
		if err := s.DisconnectSession(ctx, id); err != nil {
			// Log do erro mas continua com a exclusão
		}
	}

	// Remover do gateway WhatsApp
	if err := s.gateway.DeleteSession(ctx, session.Name); err != nil {
		// Log do erro mas continua com a exclusão
	}

	// Remover do repositório
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GenerateQRCode gera QR code para pareamento
func (s *Service) GenerateQRCode(ctx context.Context, id uuid.UUID) (*QRCodeResponse, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Verificar se já está conectada
	if session.IsConnected {
		return nil, ErrSessionAlreadyConnected
	}

	// Verificar se QR code atual ainda é válido
	if session.QRCode != nil && !session.IsQRCodeExpired() {
		return &QRCodeResponse{
			QRCode:    *session.QRCode,
			ExpiresAt: *session.QRCodeExpiresAt,
			Timeout:   120,
		}, nil
	}

	// Gerar novo QR code via gateway
	qrResponse, err := s.gateway.GenerateQRCode(ctx, session.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Atualizar sessão com novo QR code
	session.SetQRCode(qrResponse.QRCode, qrResponse.ExpiresAt)
	if err := s.repository.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session with QR code: %w", err)
	}

	return qrResponse, nil
}

// GetQRCode retorna QR code atual da sessão
func (s *Service) GetQRCode(ctx context.Context, id uuid.UUID) (*QRCodeResponse, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Verificar se tem QR code
	if session.QRCode == nil {
		return nil, ErrQRCodeNotAvailable
	}

	// Verificar se expirou
	if session.IsQRCodeExpired() {
		return nil, ErrQRCodeExpired
	}

	return &QRCodeResponse{
		QRCode:    *session.QRCode,
		ExpiresAt: *session.QRCodeExpiresAt,
		Timeout:   120,
	}, nil
}

// SetProxy configura proxy para uma sessão
func (s *Service) SetProxy(ctx context.Context, id uuid.UUID, proxy *ProxyConfig) error {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Validar configuração de proxy
	if err := s.validateProxyConfig(proxy); err != nil {
		return err
	}

	// Configurar no gateway
	if err := s.gateway.SetProxy(ctx, session.Name, proxy); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	// Atualizar sessão
	session.ProxyConfig = proxy
	session.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// GetProxy retorna configuração de proxy da sessão
func (s *Service) GetProxy(ctx context.Context, id uuid.UUID) (*ProxyConfig, error) {
	session, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session.ProxyConfig, nil
}

// UpdateLastSeen atualiza timestamp de último acesso
func (s *Service) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := s.repository.UpdateLastSeen(ctx, id, now); err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	return nil
}

// GetSessionStats retorna estatísticas das sessões
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

// SessionStats estatísticas das sessões
type SessionStats struct {
	Total     int `json:"total"`
	Connected int `json:"connected"`
	Offline   int `json:"offline"`
}

// ===== MÉTODOS AUXILIARES PRIVADOS =====

// validateCreateRequest valida dados de criação de sessão
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

	// Validar caracteres do nome (apenas alfanuméricos, hífen e underscore)
	if !isValidSessionName(req.Name) {
		return fmt.Errorf("session name contains invalid characters (only alphanumeric, dash and underscore allowed)")
	}

	// Validar proxy se fornecido
	if req.ProxyConfig != nil {
		if err := s.validateProxyConfig(req.ProxyConfig); err != nil {
			return err
		}
	}

	return nil
}

// validateProxyConfig valida configuração de proxy
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

// initiateConnection inicia processo de conexão
func (s *Service) initiateConnection(ctx context.Context, session *Session) error {
	// Configurar proxy se necessário
	if session.ProxyConfig != nil {
		if err := s.gateway.SetProxy(ctx, session.Name, session.ProxyConfig); err != nil {
			return fmt.Errorf("failed to set proxy: %w", err)
		}
	}

	// Iniciar conexão no gateway
	if err := s.gateway.ConnectSession(ctx, session.Name); err != nil {
		// Atualizar com erro de conexão
		session.SetConnectionError(err.Error())
		_ = s.repository.Update(ctx, session)
		return fmt.Errorf("failed to connect session: %w", err)
	}

	return nil
}

// syncSessionStatus sincroniza status da sessão com o gateway
func (s *Service) syncSessionStatus(ctx context.Context, session *Session) error {
	connected, err := s.gateway.IsSessionConnected(ctx, session.Name)
	if err != nil {
		return fmt.Errorf("failed to check session status: %w", err)
	}

	// Atualizar status se diferente
	if session.IsConnected != connected {
		session.UpdateConnectionStatus(connected)
		if err := s.repository.Update(ctx, session); err != nil {
			return fmt.Errorf("failed to update session status: %w", err)
		}
	}

	return nil
}

// isValidSessionName verifica se nome da sessão é válido
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

// ===== EVENT HANDLER IMPLEMENTATION =====

// SessionEventHandler implementa EventHandler para receber eventos do gateway
type SessionEventHandler struct {
	service *Service
}

// NewSessionEventHandler cria um novo handler de eventos
func NewSessionEventHandler(service *Service) *SessionEventHandler {
	return &SessionEventHandler{
		service: service,
	}
}

// OnSessionConnected chamado quando sessão conecta
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

// OnSessionDisconnected chamado quando sessão desconecta
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

// OnQRCodeGenerated chamado quando QR code é gerado
func (h *SessionEventHandler) OnQRCodeGenerated(sessionName string, qrCode string, expiresAt time.Time) {
	ctx := context.Background()

	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.SetQRCode(qrCode, expiresAt)
	_ = h.service.repository.Update(ctx, session)
}

// OnConnectionError chamado quando há erro de conexão
func (h *SessionEventHandler) OnConnectionError(sessionName string, err error) {
	ctx := context.Background()

	session, err2 := h.service.repository.GetByName(ctx, sessionName)
	if err2 != nil {
		return
	}

	session.SetConnectionError(err.Error())
	_ = h.service.repository.Update(ctx, session)
}

// OnMessageReceived chamado quando mensagem é recebida (stub)
func (h *SessionEventHandler) OnMessageReceived(sessionName string, message *WhatsAppMessage) {
	// Atualizar last seen
	ctx := context.Background()
	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.UpdateLastSeen()
	_ = h.service.repository.Update(ctx, session)
}

// OnMessageSent chamado quando mensagem é enviada (stub)
func (h *SessionEventHandler) OnMessageSent(sessionName string, messageID string, status string) {
	// Atualizar last seen
	ctx := context.Background()
	session, err := h.service.repository.GetByName(ctx, sessionName)
	if err != nil {
		return
	}

	session.UpdateLastSeen()
	_ = h.service.repository.Update(ctx, session)
}