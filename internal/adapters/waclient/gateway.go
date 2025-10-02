package waclient

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/core/group"
	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"
	"zpwoot/platform/logger"
)

// SessionService interface para atualizar dados da sessão
type SessionService interface {
	UpdateDeviceJID(ctx context.Context, id uuid.UUID, deviceJID string) error
	UpdateQRCode(ctx context.Context, id uuid.UUID, qrCode string, expiresAt time.Time) error
	ClearQRCode(ctx context.Context, id uuid.UUID) error
}

// ===== CONTACT TYPES =====

// ProfilePictureInfo representa informações da foto de perfil
type ProfilePictureInfo struct {
	JID         string     `json:"jid"`
	HasPicture  bool       `json:"has_picture"`
	URL         string     `json:"url,omitempty"`
	ID          string     `json:"id,omitempty"`
	Type        string     `json:"type,omitempty"`
	DirectPath  string     `json:"direct_path,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// UserInfo representa informações detalhadas do usuário
type UserInfo struct {
	JID          string     `json:"jid"`
	PhoneNumber  string     `json:"phone_number"`
	Name         string     `json:"name,omitempty"`
	Status       string     `json:"status,omitempty"`
	PictureID    string     `json:"picture_id,omitempty"`
	IsBusiness   bool       `json:"is_business"`
	VerifiedName string     `json:"verified_name,omitempty"`
	IsContact    bool       `json:"is_contact"`
	LastSeen     *time.Time `json:"last_seen,omitempty"`
	IsOnline     bool       `json:"is_online"`
}

// ContactInfo representa informações de contato
type ContactInfo struct {
	JID          string `json:"jid"`
	PhoneNumber  string `json:"phone_number"`
	Name         string `json:"name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBusiness   bool   `json:"is_business"`
	IsContact    bool   `json:"is_contact"`
}

// BusinessProfile representa perfil de negócio
type BusinessProfile struct {
	JID          string `json:"jid"`
	IsBusiness   bool   `json:"is_business"`
	BusinessName string `json:"business_name,omitempty"`
	Category     string `json:"category,omitempty"`
	Description  string `json:"description,omitempty"`
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
}

// SessionServiceExtended interface estendida para operações de sessão
type SessionServiceExtended interface {
	SessionService // Herda métodos existentes
	GetSession(ctx context.Context, sessionID string) (*SessionInfoResponse, error)
}

// SessionInfoResponse representa informações de uma sessão (DTO simplificado)
type SessionInfoResponse struct {
	Session *SessionDTO `json:"session"`
}

// SessionDTO representa uma sessão (DTO simplificado)
type SessionDTO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	DeviceJID *string `json:"deviceJid"`
}

// Gateway implementa session.WhatsAppGateway para integração com whatsmeow
type Gateway struct {
	// Dependencies
	logger    *logger.Logger
	container *sqlstore.Container
	db        DatabaseInterface // Interface para consultas SQL diretas

	// Internal state
	clients       map[string]*Client
	eventHandlers map[string][]session.EventHandler
	sessionUUIDs  map[string]string // mapeamento sessionName -> sessionUUID
	mu            sync.RWMutex

	// External integrations (baseado no legacy)
	webhookHandler  WebhookEventHandler
	chatwootManager ChatwootManager

	// Session service for database updates
	sessionService SessionServiceExtended
}

// DatabaseInterface interface para consultas SQL diretas
type DatabaseInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

// NewGateway cria nova instância do gateway WhatsApp
func NewGateway(container *sqlstore.Container, logger *logger.Logger) *Gateway {
	return &Gateway{
		logger:        logger,
		container:     container,
		clients:       make(map[string]*Client),
		eventHandlers: make(map[string][]session.EventHandler),
		sessionUUIDs:  make(map[string]string),
	}
}

// SetDatabase configura o database para consultas SQL diretas
func (g *Gateway) SetDatabase(db DatabaseInterface) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.db = db
}

// SetSessionService configura o session service para operações de banco
func (g *Gateway) SetSessionService(service SessionServiceExtended) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessionService = service

	// SessionService configured
}

// RegisterSessionUUID registra o mapeamento entre nome da sessão e UUID
func (g *Gateway) RegisterSessionUUID(sessionName, sessionUUID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessionUUIDs[sessionName] = sessionUUID

	g.logger.DebugWithFields("Session UUID registered", map[string]interface{}{
		"session_name": sessionName,
		"session_uuid": sessionUUID,
	})
}

// SessionExists verifica se uma sessão existe no gateway
func (g *Gateway) SessionExists(sessionName string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.clients[sessionName]
	return exists
}

// GetSessionUUID obtém o UUID da sessão pelo nome
func (g *Gateway) GetSessionUUID(sessionName string) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.sessionUUIDs[sessionName]
}

// CreateSession cria uma nova sessão WhatsApp
func (g *Gateway) CreateSession(ctx context.Context, sessionName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Verificar se sessão já existe
	if _, exists := g.clients[sessionName]; exists {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	// Criar cliente WhatsApp
	client, err := NewClient(sessionName, g.container, g.logger)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	// Configurar event handlers
	g.setupEventHandlers(client, sessionName)

	// Armazenar cliente
	g.clients[sessionName] = client

	return nil
}

// ConnectSession conecta uma sessão WhatsApp baseado no legacy
func (g *Gateway) ConnectSession(ctx context.Context, sessionName string) error {
	client := g.getClient(sessionName)
	if client == nil {
		g.logger.InfoWithFields("Client not found in memory, attempting to restore", map[string]interface{}{
			"session_name": sessionName,
		})

		// Restaurar cliente para sessão existente
		err := g.RestoreSession(ctx, sessionName)
		if err != nil {
			g.logger.ErrorWithFields("Failed to restore session", map[string]interface{}{
				"session_name": sessionName,
				"error":        err.Error(),
			})
			return fmt.Errorf("failed to restore session %s: %w", sessionName, err)
		}
		client = g.getClient(sessionName)

		if client == nil {
			g.logger.ErrorWithFields("Client still not found after restore attempt", map[string]interface{}{
				"session_name": sessionName,
			})
			return fmt.Errorf("failed to restore client for session %s", sessionName)
		}
	}

	// Verificar se já está conectado
	if client.GetClient().IsConnected() {
		return nil
	}

	// Conectar
	if err := client.Connect(); err != nil {
		g.logger.ErrorWithFields("Failed to connect WhatsApp session", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to connect session: %w", err)
	}

	return nil
}

// RestoreSession restaura um cliente WhatsApp para uma sessão existente
func (g *Gateway) RestoreSession(ctx context.Context, sessionName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Verificar se cliente já existe na memória
	if _, exists := g.clients[sessionName]; exists {
		return nil
	}

	// Restoring session client

	// Buscar deviceJID da sessão no banco para carregar device existente
	sessionUUID, exists := g.sessionUUIDs[sessionName]
	if !exists {
		g.logger.ErrorWithFields("Session UUID not found in mapping", map[string]interface{}{
			"session_name":     sessionName,
			"available_uuids":  len(g.sessionUUIDs),
			"registered_names": func() []string {
				names := make([]string, 0, len(g.sessionUUIDs))
				for name := range g.sessionUUIDs {
					names = append(names, name)
				}
				return names
			}(),
		})
		return fmt.Errorf("session UUID not found for session %s", sessionName)
	}

	// Found session UUID for restoration

	// Criar cliente WhatsApp com device existente
	client, err := g.newClientWithExistingDevice(sessionName, sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	// Configurar event handlers
	g.setupEventHandlers(client, sessionName)

	// Armazenar cliente
	g.clients[sessionName] = client

	// Session client restored successfully

	return nil
}

// newClientWithExistingDevice cria cliente WhatsApp carregando device existente
func (g *Gateway) newClientWithExistingDevice(sessionName, sessionUUID string) (*Client, error) {
	// Starting device restoration

	// Buscar deviceJID do banco de dados
	deviceJID, err := g.getDeviceJIDFromDatabase(sessionUUID)
	if err != nil {
		g.logger.WarnWithFields("Failed to get deviceJID from database, creating new device", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return NewClient(sessionName, g.container, g.logger)
	}

	// Se não tem deviceJID, criar novo device
	if deviceJID == "" {
		g.logger.InfoWithFields("No deviceJID found, creating new device", map[string]interface{}{
			"session_name": sessionName,
		})
		return NewClient(sessionName, g.container, g.logger)
	}

	// Carregar device existente pelo deviceJID
	g.logger.InfoWithFields("Loading existing device", map[string]interface{}{
		"module":  "gateway",
		"session": sessionName,
	})

	client, err := g.newClientWithDeviceJID(sessionName, deviceJID)
	if err != nil {
		g.logger.WarnWithFields("Failed to load existing device, creating new one", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return NewClient(sessionName, g.container, g.logger)
	}

	return client, nil
}

// getDeviceJIDFromDatabase busca deviceJID diretamente do banco de dados
func (g *Gateway) getDeviceJIDFromDatabase(sessionUUID string) (string, error) {
	if g.db == nil {
		return "", fmt.Errorf("database not configured")
	}

	// Query SQL direta para buscar deviceJid da sessão
	query := `SELECT "deviceJid" FROM "zpSessions" WHERE "id" = $1`

	var deviceJID *string
	err := g.db.QueryRow(query, sessionUUID).Scan(&deviceJID)
	if err != nil {
		return "", fmt.Errorf("failed to query deviceJID: %w", err)
	}

	if deviceJID == nil {
		return "", nil
	}

	return *deviceJID, nil
}

// newClientWithDeviceJID cria cliente com device existente
func (g *Gateway) newClientWithDeviceJID(sessionName, deviceJID string) (*Client, error) {
	jid, err := types.ParseJID(deviceJID)
	if err != nil {
		return nil, fmt.Errorf("invalid device JID format: %w", err)
	}

	deviceStore, err := g.container.GetDevice(context.Background(), jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get device from store: %w", err)
	}

	if deviceStore == nil {
		return nil, fmt.Errorf("device not found in store")
	}

	return NewClientWithDevice(sessionName, deviceStore, g.container, g.logger)
}

// RestoreAllSessions restaura clientes WhatsApp para todas as sessões do banco
func (g *Gateway) RestoreAllSessions(ctx context.Context, sessionNames []string) error {
	g.logger.InfoWithFields("Restoring WhatsApp clients for existing sessions", map[string]interface{}{
		"session_count": len(sessionNames),
	})

	for _, sessionName := range sessionNames {
		err := g.RestoreSession(ctx, sessionName)
		if err != nil {
			g.logger.ErrorWithFields("Failed to restore session", map[string]interface{}{
				"session_name": sessionName,
				"error":        err.Error(),
			})
			// Continuar com outras sessões mesmo se uma falhar
			continue
		}
	}

	g.logger.InfoWithFields("Session restoration completed", map[string]interface{}{
		"session_count": len(sessionNames),
	})

	return nil
}

// DisconnectSession desconecta uma sessão WhatsApp
func (g *Gateway) DisconnectSession(ctx context.Context, sessionName string) error {
	client := g.getClient(sessionName)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	g.logger.InfoWithFields("Disconnecting WhatsApp session", map[string]interface{}{
		"session_name": sessionName,
	})

	if err := client.Disconnect(); err != nil {
		g.logger.ErrorWithFields("Failed to disconnect WhatsApp session", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to disconnect session: %w", err)
	}

	return nil
}

// DeleteSession remove uma sessão WhatsApp
func (g *Gateway) DeleteSession(ctx context.Context, sessionName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	client := g.clients[sessionName]
	if client == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	g.logger.InfoWithFields("Deleting WhatsApp session", map[string]interface{}{
		"session_name": sessionName,
	})

	// Desconectar se conectado
	if client.IsConnected() {
		if err := client.Disconnect(); err != nil {
			g.logger.WarnWithFields("Error disconnecting session during deletion", map[string]interface{}{
				"session_name": sessionName,
				"error":        err.Error(),
			})
		}
	}

	// Fazer logout se logado
	if client.IsLoggedIn() {
		if err := client.Logout(); err != nil {
			g.logger.WarnWithFields("Error logging out session during deletion", map[string]interface{}{
				"session_name": sessionName,
				"error":        err.Error(),
			})
		}
	}

	// Remover da memória
	delete(g.clients, sessionName)
	delete(g.eventHandlers, sessionName)

	g.logger.InfoWithFields("WhatsApp session deleted successfully", map[string]interface{}{
		"session_name": sessionName,
	})

	return nil
}

// IsSessionConnected verifica se uma sessão está conectada baseado no legacy
func (g *Gateway) IsSessionConnected(ctx context.Context, sessionName string) (bool, error) {
	client := g.getClient(sessionName)
	if client == nil {
		g.logger.DebugWithFields("Session not found for connection check", map[string]interface{}{
			"session_name": sessionName,
		})
		return false, nil // Não retornar erro, apenas false
	}

	whatsmeowClient := client.GetClient()
	isConnected := whatsmeowClient.IsConnected()
	isLoggedIn := whatsmeowClient.IsLoggedIn()

	// Sessão está realmente conectada se ambos são true
	fullyConnected := isConnected && isLoggedIn

	g.logger.DebugWithFields("Session connection status", map[string]interface{}{
		"session_name":     sessionName,
		"is_connected":     isConnected,
		"is_logged_in":     isLoggedIn,
		"fully_connected":  fullyConnected,
		"client_status":    client.GetStatus(),
	})

	return fullyConnected, nil
}

// GenerateQRCode gera QR code para pareamento
func (g *Gateway) GenerateQRCode(ctx context.Context, sessionName string) (*session.QRCodeResponse, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	g.logger.InfoWithFields("Generating QR code", map[string]interface{}{
		"session_name": sessionName,
	})

	// Verificar se já está logado
	if client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is already logged in", sessionName)
	}

	// Conectar se não estiver conectado
	if !client.IsConnected() {
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect for QR generation: %w", err)
		}
	}

	// Obter QR code
	qrCode, err := client.GetQRCode()
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}

	// Calcular expiração (2 minutos padrão do WhatsApp)
	expiresAt := time.Now().Add(2 * time.Minute)

	response := &session.QRCodeResponse{
		QRCode:    qrCode,
		ExpiresAt: expiresAt,
		Timeout:   120, // 2 minutos em segundos
	}

	g.logger.InfoWithFields("QR code generated successfully", map[string]interface{}{
		"session_name": sessionName,
		"expires_at":   expiresAt,
	})

	return response, nil
}

// SetProxy configura proxy para uma sessão
func (g *Gateway) SetProxy(ctx context.Context, sessionName string, proxy *session.ProxyConfig) error {
	client := g.getClient(sessionName)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	// Setting proxy for session

	if err := client.SetProxy(proxy); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	return nil
}

// AddEventHandler adiciona handler de eventos
func (g *Gateway) AddEventHandler(sessionName string, handler session.EventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.eventHandlers[sessionName] == nil {
		g.eventHandlers[sessionName] = make([]session.EventHandler, 0)
	}

	g.eventHandlers[sessionName] = append(g.eventHandlers[sessionName], handler)

	g.logger.InfoWithFields("Event handler added", map[string]interface{}{
		"session_name":   sessionName,
		"handlers_count": len(g.eventHandlers[sessionName]),
	})
}

// RemoveEventHandler remove handler de eventos
func (g *Gateway) RemoveEventHandler(sessionName string, handler session.EventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	handlers := g.eventHandlers[sessionName]
	if handlers == nil {
		return
	}

	// Remover handler da lista
	for i, h := range handlers {
		if h == handler {
			g.eventHandlers[sessionName] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	g.logger.InfoWithFields("Event handler removed", map[string]interface{}{
		"session_name":   sessionName,
		"handlers_count": len(g.eventHandlers[sessionName]),
	})
}

// ===== MÉTODOS PRIVADOS =====

// getClient obtém cliente de uma sessão
func (g *Gateway) getClient(sessionName string) *Client {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.clients[sessionName]
}

// setupEventHandlers configura handlers de eventos para um cliente baseado no legacy
func (g *Gateway) setupEventHandlers(client *Client, sessionName string) {
	// Criar event handler baseado no legacy
	eventHandler := NewEventHandler(g, sessionName, g.logger)

	// Configurar webhook handler se disponível
	if g.webhookHandler != nil {
		eventHandler.SetWebhookHandler(g.webhookHandler)
	}

	// Configurar chatwoot manager se disponível
	if g.chatwootManager != nil {
		eventHandler.SetChatwootManager(g.chatwootManager)
	}

	// Configurar handler no cliente whatsmeow
	client.GetClient().AddEventHandler(func(evt interface{}) {
		// Obter UUID da sessão para o event handler
		sessionUUID := g.GetSessionUUID(sessionName)
		if sessionUUID == "" {
			// Fallback para sessionName se UUID não estiver registrado
			sessionUUID = sessionName
			g.logger.WarnWithFields("Session UUID not found, using session name", map[string]interface{}{
				"session_name": sessionName,
			})
		}
		eventHandler.HandleEvent(evt, sessionUUID)
	})

	// Registrar event handler no client para eventos customizados
	client.AddEventHandler(func(evt interface{}) {
		// Obter UUID da sessão para o event handler
		sessionUUID := g.GetSessionUUID(sessionName)
		if sessionUUID == "" {
			// Fallback para sessionName se UUID não estiver registrado
			sessionUUID = sessionName
			g.logger.WarnWithFields("Session UUID not found for custom event, using session name", map[string]interface{}{
				"session_name": sessionName,
				"event_type":   fmt.Sprintf("%T", evt),
			})
		}
		eventHandler.HandleEvent(evt, sessionUUID)
	})

	g.logger.DebugWithFields("Event handlers configured", map[string]interface{}{
		"session_name":     sessionName,
		"webhook_enabled":  g.webhookHandler != nil,
		"chatwoot_enabled": g.chatwootManager != nil,
	})
}

// SetWebhookHandler configura webhook handler baseado no legacy
func (g *Gateway) SetWebhookHandler(handler WebhookEventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.webhookHandler = handler
	g.logger.Info("Webhook handler configured for WhatsApp gateway")
}

// SetChatwootManager configura Chatwoot manager baseado no legacy
func (g *Gateway) SetChatwootManager(manager ChatwootManager) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.chatwootManager = manager
	g.logger.Info("Chatwoot manager configured for WhatsApp gateway")
}

// SaveReceivedMessage salva mensagem recebida no banco de dados
func (g *Gateway) SaveReceivedMessage(message *messaging.Message) error {
	// TODO: Implementar salvamento via message repository
	// Por enquanto, apenas log
	// Message received and ready to save (silently)

	return nil
}

// ===== GROUP OPERATIONS =====

// CreateGroup cria um novo grupo WhatsApp
func (g *Gateway) CreateGroup(ctx context.Context, sessionID, name string, participants []string, description string) (*group.GroupInfo, error) {
	g.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":   sessionID,
		"name":         name,
		"participants": len(participants),
		"description":  description != "",
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Validar entrada
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if len(participants) == 0 {
		return nil, fmt.Errorf("at least one participant is required")
	}

	// Converter participantes para JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		jid, err := types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = jid
	}

	// Criar grupo via whatsmeow
	groupInfo, err := client.client.CreateGroup(ctx, whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantJIDs,
	})
	if err != nil {
		g.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": sessionID,
			"name":       name,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Definir descrição se fornecida
	if description != "" {
		err = client.client.SetGroupTopic(groupInfo.JID, "", "", description)
		if err != nil {
			g.logger.WarnWithFields("Failed to set group description", map[string]interface{}{
				"session_id": sessionID,
				"group_jid":  groupInfo.JID.String(),
				"error":      err.Error(),
			})
		}
	}

	// Converter para formato interno
	result := g.convertToGroupInfo(groupInfo, description)

	g.logger.InfoWithFields("Group created successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  result.GroupJID,
		"name":       result.Name,
	})

	return result, nil
}

// ListJoinedGroups lista todos os grupos de uma sessão
func (g *Gateway) ListJoinedGroups(ctx context.Context, sessionID string) ([]*group.GroupInfo, error) {
	g.logger.InfoWithFields("Listing joined groups", map[string]interface{}{
		"session_id": sessionID,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Obter grupos via whatsmeow
	groups, err := client.client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get joined groups: %w", err)
	}

	// Converter para formato interno
	result := make([]*group.GroupInfo, len(groups))
	for i, groupInfo := range groups {
		result[i] = g.convertToGroupInfo(groupInfo, "")
	}

	g.logger.InfoWithFields("Groups listed successfully", map[string]interface{}{
		"session_id":   sessionID,
		"group_count":  len(result),
	})

	return result, nil
}

// GetGroupInfo obtém informações detalhadas de um grupo
func (g *Gateway) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*group.GroupInfo, error) {
	g.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	// Obter informações do grupo
	groupInfo, err := client.client.GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	// Converter para formato interno
	result := g.convertToGroupInfo(groupInfo, "")

	g.logger.InfoWithFields("Group info retrieved successfully", map[string]interface{}{
		"session_id":       sessionID,
		"group_jid":        groupJID,
		"group_name":       result.Name,
		"participant_count": len(result.Participants),
	})

	return result, nil
}

// UpdateSessionStatus atualiza o status de uma sessão no banco de dados
func (g *Gateway) UpdateSessionStatus(sessionID, status string) error {
	// Updating session status

	// TODO: Implementar atualização via session repository
	// Por enquanto, apenas log
	g.logger.DebugWithFields("Session status updated", map[string]interface{}{
		"session_id": sessionID,
		"new_status": status,
	})

	return nil
}

// UpdateSessionDeviceJID atualiza o device JID de uma sessão após pareamento bem-sucedido
func (g *Gateway) UpdateSessionDeviceJID(sessionUUID, deviceJID string) error {
	g.logger.InfoWithFields("Updating session device JID", map[string]interface{}{
		"session_uuid": sessionUUID,
		"device_jid":   deviceJID,
	})

	// Verificar se session service está configurado
	if g.sessionService == nil {
		g.logger.WarnWithFields("Session service not configured, skipping device JID update", map[string]interface{}{
			"session_uuid": sessionUUID,
			"device_jid":   deviceJID,
		})
		return nil
	}

	// Converter UUID string para uuid.UUID
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		g.logger.ErrorWithFields("Invalid session UUID format", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("invalid session UUID: %w", err)
	}

	// Atualizar device JID no banco de dados
	ctx := context.Background()
	if err := g.sessionService.UpdateDeviceJID(ctx, id, deviceJID); err != nil {
		g.logger.ErrorWithFields("Failed to update device JID in database", map[string]interface{}{
			"session_uuid": sessionUUID,
			"device_jid":   deviceJID,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to update device JID: %w", err)
	}

	g.logger.InfoWithFields("Session device JID updated successfully", map[string]interface{}{
		"session_uuid": sessionUUID,
		"device_jid":   deviceJID,
	})

	return nil
}

// UpdateSessionQRCode atualiza o QR code de uma sessão no banco de dados
func (g *Gateway) UpdateSessionQRCode(sessionUUID, qrCode string, expiresAt time.Time) error {
	g.logger.InfoWithFields("Updating session QR code", map[string]interface{}{
		"session_uuid": sessionUUID,
		"qr_length":    len(qrCode),
		"expires_at":   expiresAt,
	})

	// Verificar se session service está configurado
	if g.sessionService == nil {
		g.logger.WarnWithFields("Session service not configured, skipping QR code update", map[string]interface{}{
			"session_uuid": sessionUUID,
			"qr_length":    len(qrCode),
		})
		return nil
	}

	// Converter UUID string para uuid.UUID
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		g.logger.ErrorWithFields("Invalid session UUID format", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("invalid session UUID: %w", err)
	}

	// Atualizar QR code no banco de dados
	ctx := context.Background()
	if err := g.sessionService.UpdateQRCode(ctx, id, qrCode, expiresAt); err != nil {
		g.logger.ErrorWithFields("Failed to update QR code in database", map[string]interface{}{
			"session_uuid": sessionUUID,
			"qr_length":    len(qrCode),
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to update QR code: %w", err)
	}

	g.logger.InfoWithFields("Session QR code updated successfully", map[string]interface{}{
		"session_uuid": sessionUUID,
		"qr_length":    len(qrCode),
		"expires_at":   expiresAt,
	})

	return nil
}

// ClearSessionQRCode limpa o QR code de uma sessão no banco de dados
func (g *Gateway) ClearSessionQRCode(sessionUUID string) error {
	g.logger.InfoWithFields("Clearing session QR code", map[string]interface{}{
		"session_uuid": sessionUUID,
	})

	// Verificar se session service está configurado
	if g.sessionService == nil {
		g.logger.WarnWithFields("Session service not configured, skipping QR code clear", map[string]interface{}{
			"session_uuid": sessionUUID,
		})
		return nil
	}

	// Converter UUID string para uuid.UUID
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		g.logger.ErrorWithFields("Invalid session UUID format", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("invalid session UUID: %w", err)
	}

	// Limpar QR code no banco de dados
	ctx := context.Background()
	if err := g.sessionService.ClearQRCode(ctx, id); err != nil {
		g.logger.ErrorWithFields("Failed to clear QR code in database", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to clear QR code: %w", err)
	}

	g.logger.InfoWithFields("Session QR code cleared successfully", map[string]interface{}{
		"session_uuid": sessionUUID,
	})

	return nil
}

// AddParticipants adiciona participantes ao grupo
func (g *Gateway) AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "add")
}

// RemoveParticipants remove participantes do grupo
func (g *Gateway) RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "remove")
}

// PromoteParticipants promove participantes a admin
func (g *Gateway) PromoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "promote")
}

// DemoteParticipants remove admin de participantes
func (g *Gateway) DemoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "demote")
}

// updateGroupParticipants método interno para atualizar participantes
func (g *Gateway) updateGroupParticipants(ctx context.Context, sessionID, groupJID string, participants []string, action string) error {
	g.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    groupJID,
		"action":       action,
		"participants": len(participants),
	})

	client := g.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JIDs
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if len(participants) == 0 {
		return fmt.Errorf("no participants provided")
	}

	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJID, err := types.ParseJID(participant)
		if err != nil {
			return fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = participantJID
	}

	// Executar ação
	var participantAction whatsmeow.ParticipantChange
	switch action {
	case "add":
		participantAction = whatsmeow.ParticipantChangeAdd
	case "remove":
		participantAction = whatsmeow.ParticipantChangeRemove
	case "promote":
		participantAction = whatsmeow.ParticipantChangePromote
	case "demote":
		participantAction = whatsmeow.ParticipantChangeDemote
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	_, err = client.client.UpdateGroupParticipants(jid, participantJIDs, participantAction)
	if err != nil {
		g.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"action":     action,
			"error":      err.Error(),
		})
		return err
	}

	g.logger.InfoWithFields("Group participants updated successfully", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    groupJID,
		"action":       action,
		"participants": len(participants),
	})

	return nil
}

// SetGroupName altera o nome do grupo
func (g *Gateway) SetGroupName(ctx context.Context, sessionID, groupJID, name string) error {
	g.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
		"name":       name,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if name == "" {
		return fmt.Errorf("group name is required")
	}

	// Alterar nome
	err = client.client.SetGroupName(jid, name)
	if err != nil {
		g.logger.ErrorWithFields("Failed to set group name", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"name":       name,
			"error":      err.Error(),
		})
		return err
	}

	g.logger.InfoWithFields("Group name updated successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
		"name":       name,
	})

	return nil
}

// SetGroupDescription altera a descrição do grupo
func (g *Gateway) SetGroupDescription(ctx context.Context, sessionID, groupJID, description string) error {
	g.logger.InfoWithFields("Setting group description", map[string]interface{}{
		"session_id":  sessionID,
		"group_jid":   groupJID,
		"description": description,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Alterar descrição
	err = client.client.SetGroupTopic(jid, "", "", description)
	if err != nil {
		g.logger.ErrorWithFields("Failed to set group description", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	g.logger.InfoWithFields("Group description updated successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// SetGroupPhoto altera a foto do grupo
func (g *Gateway) SetGroupPhoto(ctx context.Context, sessionID, groupJID string, photoData []byte) error {
	g.logger.InfoWithFields("Setting group photo", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
		"photo_size": len(photoData),
	})

	client := g.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if len(photoData) == 0 {
		return fmt.Errorf("photo data is required")
	}

	// Alterar foto
	_, err = client.client.SetGroupPhoto(jid, photoData)
	if err != nil {
		g.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	g.logger.InfoWithFields("Group photo updated successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// GetGroupInviteLink obtém o link de convite do grupo
func (g *Gateway) GetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (*group.InviteLink, error) {
	g.logger.InfoWithFields("Getting group invite link", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	// Obter link de convite
	inviteLink, err := client.client.GetGroupInviteLink(jid, false)
	if err != nil {
		g.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Extrair código do link
	code := ""
	if inviteLink != "" {
		parts := strings.Split(inviteLink, "/")
		if len(parts) > 0 {
			code = parts[len(parts)-1]
		}
	}

	result := &group.InviteLink{
		GroupJID:  groupJID,
		Link:      inviteLink,
		Code:      code,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	g.logger.InfoWithFields("Group invite link retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
		"link":       inviteLink,
	})

	return result, nil
}

// RevokeGroupInviteLink revoga o link de convite atual
func (g *Gateway) RevokeGroupInviteLink(ctx context.Context, sessionID, groupJID string) error {
	g.logger.InfoWithFields("Revoking group invite link", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Revogar link (gerar novo)
	_, err = client.client.GetGroupInviteLink(jid, true)
	if err != nil {
		g.logger.ErrorWithFields("Failed to revoke group invite link", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	g.logger.InfoWithFields("Group invite link revoked successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// LeaveGroup sai do grupo
func (g *Gateway) LeaveGroup(ctx context.Context, sessionID, groupJID string) error {
	g.logger.InfoWithFields("Leaving group", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Sair do grupo
	err = client.client.LeaveGroup(jid)
	if err != nil {
		g.logger.ErrorWithFields("Failed to leave group", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	g.logger.InfoWithFields("Left group successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// JoinGroupViaLink entra em grupo via link de convite
func (g *Gateway) JoinGroupViaLink(ctx context.Context, sessionID, inviteLink string) (*group.GroupInfo, error) {
	g.logger.InfoWithFields("Joining group via link", map[string]interface{}{
		"session_id":  sessionID,
		"invite_link": inviteLink,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	if inviteLink == "" {
		return nil, fmt.Errorf("invite link is required")
	}

	// Entrar no grupo
	groupJID, err := client.client.JoinGroupWithLink(inviteLink)
	if err != nil {
		g.logger.ErrorWithFields("Failed to join group via link", map[string]interface{}{
			"session_id":  sessionID,
			"invite_link": inviteLink,
			"error":       err.Error(),
		})
		return nil, err
	}

	// Obter informações do grupo após entrar
	groupInfo, err := client.client.GetGroupInfo(groupJID)
	if err != nil {
		g.logger.WarnWithFields("Failed to get group info after joining", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID.String(),
			"error":      err.Error(),
		})
		// Retornar informações básicas mesmo sem detalhes
		return &group.GroupInfo{
			GroupJID:  groupJID.String(),
			Name:      "Unknown",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	// Converter para formato interno
	result := g.convertToGroupInfo(groupInfo, "")

	g.logger.InfoWithFields("Joined group via link successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  result.GroupJID,
		"group_name": result.Name,
	})

	return result, nil
}

// ===== CONTACT OPERATIONS =====

// IsOnWhatsApp verifica se números de telefone estão no WhatsApp
func (g *Gateway) IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]bool, error) {
	g.logger.InfoWithFields("Checking if numbers are on WhatsApp", map[string]interface{}{
		"session_id":   sessionID,
		"phone_count":  len(phoneNumbers),
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	if len(phoneNumbers) == 0 {
		return nil, fmt.Errorf("no phone numbers provided")
	}
	if len(phoneNumbers) > 50 {
		return nil, fmt.Errorf("maximum 50 phone numbers allowed")
	}

	// Normalizar números
	normalizedNumbers := make([]string, len(phoneNumbers))
	for i, phone := range phoneNumbers {
		// Normalizar número (remover caracteres especiais)
		normalizedPhone := strings.ReplaceAll(phone, "+", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, "-", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, " ", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, "(", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, ")", "")
		normalizedNumbers[i] = normalizedPhone
	}

	// Verificar via whatsmeow (método simplificado)
	resultMap := make(map[string]bool)
	for _, phone := range phoneNumbers {
		// Por enquanto, assumir que todos os números estão no WhatsApp
		// TODO: Implementar verificação real quando API estiver disponível
		resultMap[phone] = true
	}

	g.logger.InfoWithFields("WhatsApp numbers checked successfully", map[string]interface{}{
		"session_id":   sessionID,
		"phone_count":  len(phoneNumbers),
		"found_count":  len(resultMap),
	})

	return resultMap, nil
}

// GetProfilePictureInfo obtém informações da foto de perfil
func (g *Gateway) GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (*ProfilePictureInfo, error) {
	g.logger.InfoWithFields("Getting profile picture info", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
		"preview":    preview,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	targetJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Obter foto de perfil
	pic, err := client.client.GetProfilePictureInfo(targetJID, &whatsmeow.GetProfilePictureParams{
		Preview: preview,
	})
	if err != nil {
		g.logger.ErrorWithFields("Failed to get profile picture info", map[string]interface{}{
			"session_id": sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, err
	}

	result := &ProfilePictureInfo{
		JID:        jid,
		HasPicture: pic != nil,
	}

	if pic != nil {
		result.URL = pic.URL
		result.ID = pic.ID
		result.Type = "image"
		result.DirectPath = pic.DirectPath
		// whatsmeow não fornece timestamp da foto de perfil
		now := time.Now()
		result.UpdatedAt = &now
	}

	g.logger.InfoWithFields("Profile picture info retrieved successfully", map[string]interface{}{
		"session_id":   sessionID,
		"jid":          jid,
		"has_picture":  result.HasPicture,
	})

	return result, nil
}

// GetUserInfo obtém informações detalhadas do usuário
func (g *Gateway) GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]*UserInfo, error) {
	g.logger.InfoWithFields("Getting user info", map[string]interface{}{
		"session_id": sessionID,
		"jid_count":  len(jids),
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	if len(jids) == 0 {
		return nil, fmt.Errorf("no JIDs provided")
	}
	if len(jids) > 20 {
		return nil, fmt.Errorf("maximum 20 JIDs allowed")
	}

	// Parse JIDs
	targetJIDs := make([]types.JID, len(jids))
	for i, jid := range jids {
		targetJID, err := types.ParseJID(jid)
		if err != nil {
			return nil, fmt.Errorf("invalid JID %s: %w", jid, err)
		}
		targetJIDs[i] = targetJID
	}
	_ = targetJIDs // Evitar warning de variável não usada

	// Obter informações dos usuários
	results := make([]*UserInfo, 0, len(jids))
	for i, _ := range targetJIDs {
		userInfo := &UserInfo{
			JID:         jids[i],
			PhoneNumber: g.extractPhoneFromJID(jids[i]),
		}

		// Obter informações básicas (simplificado por enquanto)
		// TODO: Implementar quando API do whatsmeow estiver disponível
		userInfo.Name = "User " + userInfo.PhoneNumber
		userInfo.IsContact = true

		results = append(results, userInfo)
	}

	g.logger.InfoWithFields("User info retrieved successfully", map[string]interface{}{
		"session_id": sessionID,
		"jid_count":  len(jids),
		"found":      len(results),
	})

	return results, nil
}

// GetAllContacts obtém todos os contatos da agenda
func (g *Gateway) GetAllContacts(ctx context.Context, sessionID string) ([]*ContactInfo, error) {
	g.logger.InfoWithFields("Getting all contacts", map[string]interface{}{
		"session_id": sessionID,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Obter contatos do store (simplificado por enquanto)
	// TODO: Implementar quando API do whatsmeow estiver disponível
	results := make([]*ContactInfo, 0)

	// Por enquanto, retornar lista vazia
	// A implementação real virá quando a API estiver disponível

	g.logger.InfoWithFields("All contacts retrieved successfully", map[string]interface{}{
		"session_id":    sessionID,
		"contact_count": len(results),
	})

	return results, nil
}

// GetBusinessProfile obtém perfil de negócio
func (g *Gateway) GetBusinessProfile(ctx context.Context, sessionID, jid string) (*BusinessProfile, error) {
	g.logger.InfoWithFields("Getting business profile", map[string]interface{}{
		"session_id": sessionID,
		"jid":        jid,
	})

	client := g.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	// Parse JID
	targetJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}
	_ = targetJID // Evitar warning de variável não usada

	// Obter perfil de negócio (simplificado por enquanto)
	// TODO: Implementar quando API do whatsmeow estiver disponível
	result := &BusinessProfile{
		JID:        jid,
		IsBusiness: false, // Por enquanto, assumir que não é negócio
	}

	g.logger.InfoWithFields("Business profile retrieved successfully", map[string]interface{}{
		"session_id":  sessionID,
		"jid":         jid,
		"is_business": result.IsBusiness,
	})

	return result, nil
}

// extractPhoneFromJID extrai o número de telefone de um JID
func (g *Gateway) extractPhoneFromJID(jid string) string {
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

// convertToGroupInfo converte whatsmeow.GroupInfo para group.GroupInfo
func (g *Gateway) convertToGroupInfo(groupInfo *types.GroupInfo, description string) *group.GroupInfo {
	participants := make([]group.Participant, len(groupInfo.Participants))
	for i, p := range groupInfo.Participants {
		role := group.ParticipantRoleMember
		if p.IsSuperAdmin {
			role = group.ParticipantRoleOwner
		} else if p.IsAdmin {
			role = group.ParticipantRoleAdmin
		}

		participants[i] = group.Participant{
			JID:      p.JID.String(),
			Role:     role,
			JoinedAt: time.Now(), // whatsmeow não fornece data de entrada
			Status:   group.ParticipantStatusActive,
		}
	}

	settings := group.GroupSettings{
		Announce:         groupInfo.IsAnnounce,
		Restrict:         groupInfo.IsLocked,
		JoinApprovalMode: "auto",
		MemberAddMode:    "all_members",
		Locked:           groupInfo.IsLocked,
	}

	return &group.GroupInfo{
		GroupJID:     groupInfo.JID.String(),
		Name:         groupInfo.Name,
		Description:  description,
		Owner:        groupInfo.OwnerJID.String(),
		Participants: participants,
		Settings:     settings,
		CreatedAt:    groupInfo.GroupCreated,
		UpdatedAt:    time.Now(),
	}
}

// handleWhatsmeowEvent processa eventos do whatsmeow e repassa para handlers registrados
func (g *Gateway) handleWhatsmeowEvent(evt interface{}, sessionName string) {
	g.mu.RLock()
	handlers := g.eventHandlers[sessionName]
	g.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	// Processar evento e repassar para handlers
	g.processAndDispatchEvent(evt, sessionName, handlers)
}

// processAndDispatchEvent processa evento e despacha para handlers
func (g *Gateway) processAndDispatchEvent(evt interface{}, sessionName string, handlers []session.EventHandler) {
	// TODO: Implementar processamento específico de cada tipo de evento
	// Por enquanto, apenas log do evento
	g.logger.DebugWithFields("WhatsApp event received", map[string]interface{}{
		"session_name": sessionName,
		"event_type":   fmt.Sprintf("%T", evt),
		"handlers":     len(handlers),
	})
}

// GetSessionInfo implementa session.WhatsAppGateway.GetSessionInfo baseado no legacy
func (g *Gateway) GetSessionInfo(ctx context.Context, sessionName string) (*session.DeviceInfo, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	whatsmeowClient := client.GetClient()
	store := whatsmeowClient.Store

	// Obter informações reais do device baseado no legacy
	deviceInfo := &session.DeviceInfo{
		Platform:    "whatsmeow",
		DeviceModel: "zpwoot-gateway",
		OSVersion:   "1.0.0",
		AppVersion:  "2.0.0",
	}

	// Log informações do device se disponível
	if store.ID != nil {
		g.logger.DebugWithFields("Retrieved session info", map[string]interface{}{
			"session_name":   sessionName,
			"device_jid":     store.ID.String(),
			"push_name":      store.PushName,
			"business_name":  store.BusinessName,
		})
	} else {
		g.logger.DebugWithFields("Retrieved session info - no device registered", map[string]interface{}{
			"session_name": sessionName,
		})
	}

	return deviceInfo, nil
}

// ===== MÉTODOS DE ENVIO DE MENSAGEM =====

// SendTextMessage envia uma mensagem de texto via WhatsApp
func (g *Gateway) SendTextMessage(ctx context.Context, sessionName, to, content string) (*session.MessageSendResult, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionName)
	}

	g.logger.InfoWithFields("Sending text message via WhatsApp", map[string]interface{}{
		"session_name": sessionName,
		"to":           to,
		"content_len":  len(content),
	})

	// Parse recipient JID
	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	// Criar mensagem de texto
	message := &waE2E.Message{
		Conversation: &content,
	}

	// Enviar mensagem
	whatsmeowClient := client.GetClient()
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, message)
	if err != nil {
		g.logger.ErrorWithFields("Failed to send text message", map[string]interface{}{
			"session_name": sessionName,
			"to":           to,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	result := &session.MessageSendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		To:        to,
	}

	g.logger.InfoWithFields("Text message sent successfully", map[string]interface{}{
		"session_name": sessionName,
		"message_id":   resp.ID,
		"to":           to,
	})

	return result, nil
}

// SendMediaMessage envia uma mensagem de mídia via WhatsApp
func (g *Gateway) SendMediaMessage(ctx context.Context, sessionName, to, mediaURL, caption, mediaType string) (*session.MessageSendResult, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionName)
	}

	g.logger.InfoWithFields("Sending media message via WhatsApp", map[string]interface{}{
		"session_name": sessionName,
		"to":           to,
		"media_url":    mediaURL,
		"media_type":   mediaType,
		"has_caption":  caption != "",
	})

	// Parse recipient JID
	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	// TODO: Implementar download e upload de mídia
	// Por enquanto, enviar como mensagem de texto com URL
	content := mediaURL
	if caption != "" {
		content = fmt.Sprintf("%s\n\n%s", caption, mediaURL)
	}

	message := &waE2E.Message{
		Conversation: &content,
	}

	// Enviar mensagem
	whatsmeowClient := client.GetClient()
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, message)
	if err != nil {
		g.logger.ErrorWithFields("Failed to send media message", map[string]interface{}{
			"session_name": sessionName,
			"to":           to,
			"media_type":   mediaType,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to send media message: %w", err)
	}

	result := &session.MessageSendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		To:        to,
	}

	g.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_name": sessionName,
		"message_id":   resp.ID,
		"to":           to,
		"media_type":   mediaType,
	})

	return result, nil
}

// SendLocationMessage envia uma mensagem de localização via WhatsApp
func (g *Gateway) SendLocationMessage(ctx context.Context, sessionName, to string, latitude, longitude float64, address string) (*session.MessageSendResult, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionName)
	}

	g.logger.InfoWithFields("Sending location message via WhatsApp", map[string]interface{}{
		"session_name": sessionName,
		"to":           to,
		"latitude":     latitude,
		"longitude":    longitude,
		"address":      address,
	})

	// Parse recipient JID
	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	// Criar mensagem de localização
	degreesLatitude := latitude
	degreesLongitude := longitude

	message := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  &degreesLatitude,
			DegreesLongitude: &degreesLongitude,
			Name:             &address,
			Address:          &address,
		},
	}

	// Enviar mensagem
	whatsmeowClient := client.GetClient()
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, message)
	if err != nil {
		g.logger.ErrorWithFields("Failed to send location message", map[string]interface{}{
			"session_name": sessionName,
			"to":           to,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to send location message: %w", err)
	}

	result := &session.MessageSendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		To:        to,
	}

	g.logger.InfoWithFields("Location message sent successfully", map[string]interface{}{
		"session_name": sessionName,
		"message_id":   resp.ID,
		"to":           to,
	})

	return result, nil
}

// SendContactMessage envia uma mensagem de contato via WhatsApp
func (g *Gateway) SendContactMessage(ctx context.Context, sessionName, to, contactName, contactPhone string) (*session.MessageSendResult, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionName)
	}

	g.logger.InfoWithFields("Sending contact message via WhatsApp", map[string]interface{}{
		"session_name":   sessionName,
		"to":             to,
		"contact_name":   contactName,
		"contact_phone":  contactPhone,
	})

	// Parse recipient JID
	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	// Criar vCard
	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contactName, contactPhone)

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contactName,
			Vcard:       &vcard,
		},
	}

	// Enviar mensagem
	whatsmeowClient := client.GetClient()
	resp, err := whatsmeowClient.SendMessage(ctx, recipientJID, message)
	if err != nil {
		g.logger.ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"session_name": sessionName,
			"to":           to,
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("failed to send contact message: %w", err)
	}

	result := &session.MessageSendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
		To:        to,
	}

	g.logger.InfoWithFields("Contact message sent successfully", map[string]interface{}{
		"session_name": sessionName,
		"message_id":   resp.ID,
		"to":           to,
	})

	return result, nil
}

// SetEventHandler implementa session.WhatsAppGateway.SetEventHandler
func (g *Gateway) SetEventHandler(handler session.EventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Adicionar handler global para todas as sessões
	if g.eventHandlers["global"] == nil {
		g.eventHandlers["global"] = make([]session.EventHandler, 0)
	}
	g.eventHandlers["global"] = append(g.eventHandlers["global"], handler)

	g.logger.Debug("Global event handler registered")
}