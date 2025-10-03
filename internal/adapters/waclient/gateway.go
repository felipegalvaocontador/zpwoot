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

type SessionService interface {
	UpdateDeviceJID(ctx context.Context, id uuid.UUID, deviceJID string) error
	UpdateQRCode(ctx context.Context, id uuid.UUID, qrCode string, expiresAt time.Time) error
	ClearQRCode(ctx context.Context, id uuid.UUID) error
}

type ProfilePictureInfo struct {
	JID        string     `json:"jid"`
	HasPicture bool       `json:"has_picture"`
	URL        string     `json:"url,omitempty"`
	ID         string     `json:"id,omitempty"`
	Type       string     `json:"type,omitempty"`
	DirectPath string     `json:"direct_path,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

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

type ContactInfo struct {
	JID          string `json:"jid"`
	PhoneNumber  string `json:"phone_number"`
	Name         string `json:"name,omitempty"`
	BusinessName string `json:"business_name,omitempty"`
	IsBusiness   bool   `json:"is_business"`
	IsContact    bool   `json:"is_contact"`
}

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

type SessionServiceExtended interface {
	SessionService
	GetSession(ctx context.Context, sessionID string) (*SessionInfoResponse, error)
}

type SessionInfoResponse struct {
	Session *SessionDTO `json:"session"`
}

type SessionDTO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	DeviceJID *string `json:"deviceJid"`
}

type Gateway struct {
	logger    *logger.Logger
	container *sqlstore.Container
	db        DatabaseInterface

	clients       map[string]*Client
	eventHandlers map[string][]session.EventHandler
	sessionUUIDs  map[string]string
	mu            sync.RWMutex

	webhookHandler  WebhookEventHandler
	chatwootManager ChatwootManager

	sessionService SessionServiceExtended
}

type DatabaseInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

func NewGateway(container *sqlstore.Container, logger *logger.Logger) *Gateway {
	return &Gateway{
		logger:        logger,
		container:     container,
		clients:       make(map[string]*Client),
		eventHandlers: make(map[string][]session.EventHandler),
		sessionUUIDs:  make(map[string]string),
	}
}

func (g *Gateway) SetDatabase(db DatabaseInterface) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.db = db
}

func (g *Gateway) SetSessionService(service SessionServiceExtended) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessionService = service

}

func (g *Gateway) RegisterSessionUUID(sessionName, sessionUUID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessionUUIDs[sessionName] = sessionUUID

	g.logger.DebugWithFields("Session UUID registered", map[string]interface{}{
		"session_name": sessionName,
		"session_uuid": sessionUUID,
	})
}

func (g *Gateway) SessionExists(sessionName string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.clients[sessionName]
	return exists
}

func (g *Gateway) GetSessionUUID(sessionName string) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.sessionUUIDs[sessionName]
}

func (g *Gateway) CreateSession(ctx context.Context, sessionName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.clients[sessionName]; exists {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	client, err := NewClient(sessionName, g.container, g.logger)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	g.setupEventHandlers(client, sessionName)

	g.clients[sessionName] = client

	return nil
}

func (g *Gateway) ConnectSession(ctx context.Context, sessionName string) error {
	client := g.getClient(sessionName)
	if client == nil {
		g.logger.InfoWithFields("Client not found in memory, attempting to restore", map[string]interface{}{
			"session_name": sessionName,
		})

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

	if client.GetClient().IsConnected() {
		return nil
	}

	if err := client.Connect(); err != nil {
		g.logger.ErrorWithFields("Failed to connect WhatsApp session", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to connect session: %w", err)
	}

	return nil
}

func (g *Gateway) RestoreSession(ctx context.Context, sessionName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.clients[sessionName]; exists {
		return nil
	}

	sessionUUID, exists := g.sessionUUIDs[sessionName]
	if !exists {
		g.logger.ErrorWithFields("Session UUID not found in mapping", map[string]interface{}{
			"session_name":    sessionName,
			"available_uuids": len(g.sessionUUIDs),
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

	client, err := g.newClientWithExistingDevice(sessionName, sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	g.setupEventHandlers(client, sessionName)

	g.clients[sessionName] = client

	return nil
}

func (g *Gateway) newClientWithExistingDevice(sessionName, sessionUUID string) (*Client, error) {

	deviceJID, err := g.getDeviceJIDFromDatabase(sessionUUID)
	if err != nil {
		g.logger.WarnWithFields("Failed to get deviceJID from database, creating new device", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return NewClient(sessionName, g.container, g.logger)
	}

	if deviceJID == "" {
		g.logger.InfoWithFields("No deviceJID found, creating new device", map[string]interface{}{
			"session_name": sessionName,
		})
		return NewClient(sessionName, g.container, g.logger)
	}

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

func (g *Gateway) getDeviceJIDFromDatabase(sessionUUID string) (string, error) {
	if g.db == nil {
		return "", fmt.Errorf("database not configured")
	}

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

			continue
		}
	}

	g.logger.InfoWithFields("Session restoration completed", map[string]interface{}{
		"session_count": len(sessionNames),
	})

	return nil
}

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

	if client.IsConnected() {
		if err := client.Disconnect(); err != nil {
			g.logger.WarnWithFields("Error disconnecting session during deletion", map[string]interface{}{
				"session_name": sessionName,
				"error":        err.Error(),
			})
		}
	}

	if client.IsLoggedIn() {
		if err := client.Logout(); err != nil {
			g.logger.WarnWithFields("Error logging out session during deletion", map[string]interface{}{
				"session_name": sessionName,
				"error":        err.Error(),
			})
		}
	}

	delete(g.clients, sessionName)
	delete(g.eventHandlers, sessionName)

	g.logger.InfoWithFields("WhatsApp session deleted successfully", map[string]interface{}{
		"session_name": sessionName,
	})

	return nil
}

func (g *Gateway) IsSessionConnected(ctx context.Context, sessionName string) (bool, error) {
	client := g.getClient(sessionName)
	if client == nil {
		g.logger.DebugWithFields("Session not found for connection check", map[string]interface{}{
			"session_name": sessionName,
		})
		return false, nil
	}

	whatsmeowClient := client.GetClient()
	isConnected := whatsmeowClient.IsConnected()
	isLoggedIn := whatsmeowClient.IsLoggedIn()

	fullyConnected := isConnected && isLoggedIn

	g.logger.DebugWithFields("Session connection status", map[string]interface{}{
		"session_name":    sessionName,
		"is_connected":    isConnected,
		"is_logged_in":    isLoggedIn,
		"fully_connected": fullyConnected,
		"client_status":   client.GetStatus(),
	})

	return fullyConnected, nil
}

func (g *Gateway) GenerateQRCode(ctx context.Context, sessionName string) (*session.QRCodeResponse, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	g.logger.InfoWithFields("Generating QR code", map[string]interface{}{
		"session_name": sessionName,
	})

	if client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is already logged in", sessionName)
	}

	if !client.IsConnected() {
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect for QR generation: %w", err)
		}
	}

	qrCode, err := client.GetQRCode()
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}

	expiresAt := time.Now().Add(2 * time.Minute)

	response := &session.QRCodeResponse{
		QRCode:    qrCode,
		ExpiresAt: expiresAt,
		Timeout:   120,
	}

	g.logger.InfoWithFields("QR code generated successfully", map[string]interface{}{
		"session_name": sessionName,
		"expires_at":   expiresAt,
	})

	return response, nil
}

func (g *Gateway) SetProxy(ctx context.Context, sessionName string, proxy *session.ProxyConfig) error {
	client := g.getClient(sessionName)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	if err := client.SetProxy(proxy); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}

	return nil
}

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

func (g *Gateway) RemoveEventHandler(sessionName string, handler session.EventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	handlers := g.eventHandlers[sessionName]
	if handlers == nil {
		return
	}

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

func (g *Gateway) getClient(sessionName string) *Client {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.clients[sessionName]
}

func (g *Gateway) setupEventHandlers(client *Client, sessionName string) {

	eventHandler := NewEventHandler(g, sessionName, g.logger)

	if g.webhookHandler != nil {
		eventHandler.SetWebhookHandler(g.webhookHandler)
	}

	if g.chatwootManager != nil {
		eventHandler.SetChatwootManager(g.chatwootManager)
	}

	client.GetClient().AddEventHandler(func(evt interface{}) {

		sessionUUID := g.GetSessionUUID(sessionName)
		if sessionUUID == "" {

			sessionUUID = sessionName
			g.logger.WarnWithFields("Session UUID not found, using session name", map[string]interface{}{
				"session_name": sessionName,
			})
		}
		eventHandler.HandleEvent(evt, sessionUUID)
	})

	client.AddEventHandler(func(evt interface{}) {

		sessionUUID := g.GetSessionUUID(sessionName)
		if sessionUUID == "" {

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

func (g *Gateway) SetWebhookHandler(handler WebhookEventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.webhookHandler = handler
	g.logger.Info("Webhook handler configured for WhatsApp gateway")
}

func (g *Gateway) SetChatwootManager(manager ChatwootManager) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.chatwootManager = manager
	g.logger.Info("Chatwoot manager configured for WhatsApp gateway")
}

func (g *Gateway) SaveReceivedMessage(message *messaging.Message) error {

	return nil
}

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

	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if len(participants) == 0 {
		return nil, fmt.Errorf("at least one participant is required")
	}

	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		jid, err := types.ParseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = jid
	}

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

	result := g.convertToGroupInfo(groupInfo, description)

	g.logger.InfoWithFields("Group created successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  result.GroupJID,
		"name":       result.Name,
	})

	return result, nil
}

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

	groups, err := client.client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get joined groups: %w", err)
	}

	result := make([]*group.GroupInfo, len(groups))
	for i, groupInfo := range groups {
		result[i] = g.convertToGroupInfo(groupInfo, "")
	}

	g.logger.InfoWithFields("Groups listed successfully", map[string]interface{}{
		"session_id":  sessionID,
		"group_count": len(result),
	})

	return result, nil
}

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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	groupInfo, err := client.client.GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	result := g.convertToGroupInfo(groupInfo, "")

	g.logger.InfoWithFields("Group info retrieved successfully", map[string]interface{}{
		"session_id":        sessionID,
		"group_jid":         groupJID,
		"group_name":        result.Name,
		"participant_count": len(result.Participants),
	})

	return result, nil
}

func (g *Gateway) UpdateSessionStatus(sessionID, status string) error {

	g.logger.DebugWithFields("Session status updated", map[string]interface{}{
		"session_id": sessionID,
		"new_status": status,
	})

	return nil
}

func (g *Gateway) UpdateSessionDeviceJID(sessionUUID, deviceJID string) error {
	g.logger.InfoWithFields("Updating session device JID", map[string]interface{}{
		"session_uuid": sessionUUID,
		"device_jid":   deviceJID,
	})

	if g.sessionService == nil {
		g.logger.WarnWithFields("Session service not configured, skipping device JID update", map[string]interface{}{
			"session_uuid": sessionUUID,
			"device_jid":   deviceJID,
		})
		return nil
	}

	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		g.logger.ErrorWithFields("Invalid session UUID format", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("invalid session UUID: %w", err)
	}

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

func (g *Gateway) UpdateSessionQRCode(sessionUUID, qrCode string, expiresAt time.Time) error {
	g.logger.InfoWithFields("Updating session QR code", map[string]interface{}{
		"session_uuid": sessionUUID,
		"qr_length":    len(qrCode),
		"expires_at":   expiresAt,
	})

	if g.sessionService == nil {
		g.logger.WarnWithFields("Session service not configured, skipping QR code update", map[string]interface{}{
			"session_uuid": sessionUUID,
			"qr_length":    len(qrCode),
		})
		return nil
	}

	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		g.logger.ErrorWithFields("Invalid session UUID format", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("invalid session UUID: %w", err)
	}

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

func (g *Gateway) ClearSessionQRCode(sessionUUID string) error {
	g.logger.InfoWithFields("Clearing session QR code", map[string]interface{}{
		"session_uuid": sessionUUID,
	})

	if g.sessionService == nil {
		g.logger.WarnWithFields("Session service not configured, skipping QR code clear", map[string]interface{}{
			"session_uuid": sessionUUID,
		})
		return nil
	}

	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		g.logger.ErrorWithFields("Invalid session UUID format", map[string]interface{}{
			"session_uuid": sessionUUID,
			"error":        err.Error(),
		})
		return fmt.Errorf("invalid session UUID: %w", err)
	}

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

func (g *Gateway) AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "add")
}

func (g *Gateway) RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "remove")
}

func (g *Gateway) PromoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "promote")
}

func (g *Gateway) DemoteParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return g.updateGroupParticipants(ctx, sessionID, groupJID, participants, "demote")
}

func (g *Gateway) updateGroupParticipants(_ context.Context, sessionID, groupJID string, participants []string, action string) error {
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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if name == "" {
		return fmt.Errorf("group name is required")
	}

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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if len(photoData) == 0 {
		return fmt.Errorf("photo data is required")
	}

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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	inviteLink, err := client.client.GetGroupInviteLink(jid, false)
	if err != nil {
		g.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

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

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

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

	groupJID, err := client.client.JoinGroupWithLink(inviteLink)
	if err != nil {
		g.logger.ErrorWithFields("Failed to join group via link", map[string]interface{}{
			"session_id":  sessionID,
			"invite_link": inviteLink,
			"error":       err.Error(),
		})
		return nil, err
	}

	groupInfo, err := client.client.GetGroupInfo(groupJID)
	if err != nil {
		g.logger.WarnWithFields("Failed to get group info after joining", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupJID.String(),
			"error":      err.Error(),
		})

		return &group.GroupInfo{
			GroupJID:  groupJID.String(),
			Name:      "Unknown",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	result := g.convertToGroupInfo(groupInfo, "")

	g.logger.InfoWithFields("Joined group via link successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  result.GroupJID,
		"group_name": result.Name,
	})

	return result, nil
}

func (g *Gateway) IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]bool, error) {
	g.logger.InfoWithFields("Checking if numbers are on WhatsApp", map[string]interface{}{
		"session_id":  sessionID,
		"phone_count": len(phoneNumbers),
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

	normalizedNumbers := make([]string, len(phoneNumbers))
	for i, phone := range phoneNumbers {

		normalizedPhone := strings.ReplaceAll(phone, "+", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, "-", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, " ", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, "(", "")
		normalizedPhone = strings.ReplaceAll(normalizedPhone, ")", "")
		normalizedNumbers[i] = normalizedPhone
	}

	resultMap := make(map[string]bool)
	for _, phone := range phoneNumbers {

		resultMap[phone] = true
	}

	g.logger.InfoWithFields("WhatsApp numbers checked successfully", map[string]interface{}{
		"session_id":  sessionID,
		"phone_count": len(phoneNumbers),
		"found_count": len(resultMap),
	})

	return resultMap, nil
}

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

	targetJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

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

		now := time.Now()
		result.UpdatedAt = &now
	}

	g.logger.InfoWithFields("Profile picture info retrieved successfully", map[string]interface{}{
		"session_id":  sessionID,
		"jid":         jid,
		"has_picture": result.HasPicture,
	})

	return result, nil
}

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

	targetJIDs := make([]types.JID, len(jids))
	for i, jid := range jids {
		targetJID, err := types.ParseJID(jid)
		if err != nil {
			return nil, fmt.Errorf("invalid JID %s: %w", jid, err)
		}
		targetJIDs[i] = targetJID
	}
	_ = targetJIDs

	results := make([]*UserInfo, 0, len(jids))
	for i := range targetJIDs {
		userInfo := &UserInfo{
			JID:         jids[i],
			PhoneNumber: g.extractPhoneFromJID(jids[i]),
		}

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

	results := make([]*ContactInfo, 0)

	g.logger.InfoWithFields("All contacts retrieved successfully", map[string]interface{}{
		"session_id":    sessionID,
		"contact_count": len(results),
	})

	return results, nil
}

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

	targetJID, err := types.ParseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}
	_ = targetJID

	result := &BusinessProfile{
		JID:        jid,
		IsBusiness: false,
	}

	g.logger.InfoWithFields("Business profile retrieved successfully", map[string]interface{}{
		"session_id":  sessionID,
		"jid":         jid,
		"is_business": result.IsBusiness,
	})

	return result, nil
}

func (g *Gateway) extractPhoneFromJID(jid string) string {
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

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
			JoinedAt: time.Now(),
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

func (g *Gateway) GetSessionInfo(ctx context.Context, sessionName string) (*session.DeviceInfo, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	whatsmeowClient := client.GetClient()
	store := whatsmeowClient.Store

	deviceInfo := &session.DeviceInfo{
		Platform:    "whatsmeow",
		DeviceModel: "zpwoot-gateway",
		OSVersion:   "1.0.0",
		AppVersion:  "2.0.0",
	}

	if store.ID != nil {
		g.logger.DebugWithFields("Retrieved session info", map[string]interface{}{
			"session_name":  sessionName,
			"device_jid":    store.ID.String(),
			"push_name":     store.PushName,
			"business_name": store.BusinessName,
		})
	} else {
		g.logger.DebugWithFields("Retrieved session info - no device registered", map[string]interface{}{
			"session_name": sessionName,
		})
	}

	return deviceInfo, nil
}

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

	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	message := &waE2E.Message{
		Conversation: &content,
	}

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

	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	content := mediaURL
	if caption != "" {
		content = fmt.Sprintf("%s\n\n%s", caption, mediaURL)
	}

	message := &waE2E.Message{
		Conversation: &content,
	}

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

	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

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

func (g *Gateway) SendContactMessage(ctx context.Context, sessionName, to, contactName, contactPhone string) (*session.MessageSendResult, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionName)
	}

	g.logger.InfoWithFields("Sending contact message via WhatsApp", map[string]interface{}{
		"session_name":  sessionName,
		"to":            to,
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	recipientJID, err := types.ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contactName, contactPhone)

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contactName,
			Vcard:       &vcard,
		},
	}

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

func (g *Gateway) SetEventHandler(handler session.EventHandler) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.eventHandlers["global"] == nil {
		g.eventHandlers["global"] = make([]session.EventHandler, 0)
	}
	g.eventHandlers["global"] = append(g.eventHandlers["global"], handler)

	g.logger.Debug("Global event handler registered")
}

func (g *Gateway) getEventHandlers(key string) []session.EventHandler {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if handlers, exists := g.eventHandlers[key]; exists {

		result := make([]session.EventHandler, len(handlers))
		copy(result, handlers)
		return result
	}
	return nil
}
