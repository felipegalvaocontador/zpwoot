package wameow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	appMessage "zpwoot/internal/app/message"
	"zpwoot/internal/domain/message"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type SessionStats struct {
	MessagesSent     int64
	MessagesReceived int64
	LastActivity     int64
	StartTime        int64
}

type EventHandlerInfo struct {
	Handler ports.EventHandler
	ID      string
}

type Manager struct {
	sessionMgr      SessionUpdater
	chatwootManager ChatwootManager
	webhookHandler  WebhookEventHandler
	logger          *logger.Logger
	qrGenerator     *QRCodeGenerator
	connectionMgr   *ConnectionManager
	clients         map[string]*WameowClient
	sessionStats    map[string]*SessionStats
	eventHandlers   map[string]map[string]*EventHandlerInfo
	container       *sqlstore.Container
	statsMutex      sync.RWMutex
	handlersMutex   sync.RWMutex
	clientsMutex    sync.RWMutex
}

func NewManager(
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) *Manager {
	return &Manager{
		clients:       make(map[string]*WameowClient),
		container:     container,
		connectionMgr: NewConnectionManager(logger),
		qrGenerator:   NewQRCodeGenerator(logger),
		sessionMgr:    NewSessionManager(sessionRepo, logger),
		logger:        logger,
		sessionStats:  make(map[string]*SessionStats),
		eventHandlers: make(map[string]map[string]*EventHandlerInfo),
	}
}

func (m *Manager) CreateSession(sessionID string, config *session.ProxyConfig) error {
	if err := ValidateSessionID(sessionID); err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	if _, exists := m.clients[sessionID]; exists {
		return fmt.Errorf("session %s already exists", sessionID)
	}

	client, err := m.createWameowClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to create WameowClient for session %s: %w", sessionID, err)
	}

	if err := m.configureSession(client, sessionID, config); err != nil {
		return fmt.Errorf("failed to configure session %s: %w", sessionID, err)
	}

	m.clients[sessionID] = client
	m.initSessionStats(sessionID)

	return nil
}

func (m *Manager) createWameowClient(sessionID string) (*WameowClient, error) {
	return NewWameowClient(sessionID, m.container, m.sessionMgr.GetSessionRepo(), m.logger)
}

func (m *Manager) configureSession(client *WameowClient, sessionID string, config *session.ProxyConfig) error {
	m.setupEventHandlers(client.GetClient(), sessionID)

	eventHandler := NewEventHandler(m, m.sessionMgr, m.qrGenerator, m.logger)
	if m.webhookHandler != nil {
		eventHandler.SetWebhookHandler(m.webhookHandler)
	}
	if m.chatwootManager != nil {
		eventHandler.SetChatwootManager(m.chatwootManager)
	}
	client.SetEventHandler(eventHandler)

	if config != nil {
		if err := m.applyProxyConfig(client.GetClient(), config); err != nil {
			m.logger.WarnWithFields("Failed to apply proxy config", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}
	}

	return nil
}

func (m *Manager) ConnectSession(sessionID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		sess, err := m.sessionMgr.GetSession(sessionID)
		if err != nil {
			return fmt.Errorf("session %s not found", sessionID)
		}

		if err := m.CreateSession(sessionID, sess.ProxyConfig); err != nil {
			return fmt.Errorf("failed to initialize Wameow client for session %s: %w", sessionID, err)
		}

		client = m.getClient(sessionID)
		if client == nil {
			return fmt.Errorf("failed to create Wameow client for session %s", sessionID)
		}
	}

	err := client.Connect()
	if err != nil {
		m.sessionMgr.UpdateConnectionStatus(sessionID, false)
		return fmt.Errorf("failed to connect session %s: %w", sessionID, err)
	}

	return nil
}

func (m *Manager) DisconnectSession(sessionID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	err := client.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect session %s: %w", sessionID, err)
	}

	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	m.clientsMutex.Lock()
	delete(m.clients, sessionID)
	m.clientsMutex.Unlock()

	return nil
}

func (m *Manager) LogoutSession(sessionID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	err := client.Logout()
	if err != nil {
		m.logger.WarnWithFields("Error during logout", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}

	m.sessionMgr.UpdateConnectionStatus(sessionID, false)

	m.clientsMutex.Lock()
	delete(m.clients, sessionID)
	m.clientsMutex.Unlock()

	return nil
}

func (m *Manager) GetQRCode(sessionID string) (*session.QRCodeResponse, error) {
	m.logger.InfoWithFields("Getting QR code for session", map[string]interface{}{
		"session_id": sessionID,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is already logged in", sessionID)
	}

	qrCode, err := client.GetQRCode()
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code for session %s: %w", sessionID, err)
	}

	qrCodeImage := m.qrGenerator.GenerateQRCodeImage(qrCode)

	return &session.QRCodeResponse{
		QRCode:      qrCode,
		QRCodeImage: qrCodeImage,
		ExpiresAt:   time.Now().Add(2 * time.Minute),
		Timeout:     120,
	}, nil
}

func (m *Manager) PairPhone(sessionID, phoneNumber string) error {
	m.logger.InfoWithFields("Pairing phone number", map[string]interface{}{
		"session_id":   sessionID,
		"phone_number": phoneNumber,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return fmt.Errorf("phone pairing not implemented yet")
}

func (m *Manager) IsConnected(sessionID string) bool {
	client := m.getClient(sessionID)
	if client == nil {
		return false
	}
	return client.IsConnected()
}

func (m *Manager) GetDeviceInfo(sessionID string) (*session.DeviceInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	return &session.DeviceInfo{
		Platform:    "web",
		DeviceModel: "Chrome",
		OSVersion:   "Unknown",
		AppVersion:  "2.2412.54",
	}, nil
}

func (m *Manager) SetProxy(sessionID string, config *session.ProxyConfig) error {
	m.logger.InfoWithFields("Setting proxy for session", map[string]interface{}{
		"session_id": sessionID,
		"proxy_type": config.Type,
		"proxy_host": config.Host,
	})

	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return m.applyProxyConfig(client.GetClient(), config)
}

func (m *Manager) initSessionStats(sessionID string) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	if _, exists := m.sessionStats[sessionID]; !exists {
		m.sessionStats[sessionID] = &SessionStats{
			StartTime: time.Now().Unix(),
		}
	}
}

func (m *Manager) incrementMessagesSent(sessionID string) {
	m.statsMutex.RLock()
	stats, exists := m.sessionStats[sessionID]
	m.statsMutex.RUnlock()

	if exists {
		atomic.AddInt64(&stats.MessagesSent, 1)
		atomic.StoreInt64(&stats.LastActivity, time.Now().Unix())
	}
}

func (m *Manager) incrementMessagesReceived(sessionID string) {
	m.statsMutex.RLock()
	stats, exists := m.sessionStats[sessionID]
	m.statsMutex.RUnlock()

	if exists {
		atomic.AddInt64(&stats.MessagesReceived, 1)
		atomic.StoreInt64(&stats.LastActivity, time.Now().Unix())
	}
}

func (m *Manager) getSessionStats(sessionID string) *SessionStats {
	m.statsMutex.RLock()
	defer m.statsMutex.RUnlock()

	stats, exists := m.sessionStats[sessionID]
	if !exists {
		return &SessionStats{
			StartTime: time.Now().Unix(),
		}
	}

	return stats
}

func (m *Manager) GetProxy(sessionID string) (*session.ProxyConfig, error) {
	return nil, nil
}

func (m *Manager) GetSessionStats(sessionID string) (*ports.SessionStats, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	stats := m.getSessionStats(sessionID)

	uptime := int64(0)
	if stats.StartTime > 0 {
		uptime = time.Now().Unix() - stats.StartTime
	}

	return &ports.SessionStats{
		MessagesSent:     atomic.LoadInt64(&stats.MessagesSent),
		MessagesReceived: atomic.LoadInt64(&stats.MessagesReceived),
		LastActivity:     atomic.LoadInt64(&stats.LastActivity),
		Uptime:           uptime,
	}, nil
}

func (m *Manager) GetSession(sessionID string) (*session.Session, error) {
	return m.sessionMgr.GetSession(sessionID)
}

func (m *Manager) SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error {
	client, recipientJID, err := m.validateMediaMessageRequest(sessionID, to)
	if err != nil {
		return err
	}

	uploaded, err := m.uploadMedia(client, media, mediaType, sessionID, to)
	if err != nil {
		return err
	}

	msg, err := m.createMediaMessage(mediaType, uploaded, caption)
	if err != nil {
		return err
	}

	return m.sendMediaMessageAndLog(client, recipientJID, msg, sessionID, to, mediaType)
}

func (m *Manager) validateMediaMessageRequest(sessionID, to string) (*WameowClient, types.JID, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, types.EmptyJID, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, types.EmptyJID, fmt.Errorf("session %s is not logged in", sessionID)
	}

	recipientJID, err := ParseJID(to)
	if err != nil {
		return nil, types.EmptyJID, fmt.Errorf("invalid recipient JID %s: %w", to, err)
	}

	return client, recipientJID, nil
}

func (m *Manager) uploadMedia(client *WameowClient, media []byte, mediaType, sessionID, to string) (whatsmeow.UploadResponse, error) {
	uploaded, err := client.GetClient().Upload(context.Background(), media, whatsmeow.MediaType(mediaType))
	if err != nil {
		m.logger.ErrorWithFields("Failed to upload media", map[string]interface{}{
			"session_id": sessionID,
			"to":         to,
			"media_type": mediaType,
			"error":      err.Error(),
		})
		return whatsmeow.UploadResponse{}, fmt.Errorf("failed to upload media: %w", err)
	}
	return uploaded, nil
}

func (m *Manager) createMediaMessage(mediaType string, uploaded whatsmeow.UploadResponse, caption string) (*waE2E.Message, error) {
	switch mediaType {
	case "image":
		return &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}, nil
	case "video":
		return &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}, nil
	case "audio":
		return &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
			},
		}, nil
	case "document":
		return &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				URL:           &uploaded.URL,
				DirectPath:    &uploaded.DirectPath,
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    &uploaded.FileLength,
				Caption:       &caption,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mediaType)
	}
}

func (m *Manager) sendMediaMessageAndLog(client *WameowClient, recipientJID types.JID, msg *waE2E.Message, sessionID, to, mediaType string) error {
	_, err := client.GetClient().SendMessage(context.Background(), recipientJID, msg)
	if err != nil {
		m.logger.ErrorWithFields("Failed to send media message", map[string]interface{}{
			"session_id": sessionID,
			"to":         to,
			"media_type": mediaType,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to send media message: %w", err)
	}

	m.incrementMessagesSent(sessionID)

	m.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"media_type": mediaType,
	})

	return nil
}

func (m *Manager) RegisterEventHandler(sessionID string, handler ports.EventHandler) error {
	handlerID := m.registerHandlerInRegistry(sessionID, handler)

	m.attachHandlerToClient(sessionID, handler)

	m.logger.InfoWithFields("Event handler registered", map[string]interface{}{
		"session_id": sessionID,
		"handler_id": handlerID,
	})

	return nil
}

func (m *Manager) registerHandlerInRegistry(sessionID string, handler ports.EventHandler) string {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	if m.eventHandlers[sessionID] == nil {
		m.eventHandlers[sessionID] = make(map[string]*EventHandlerInfo)
	}

	handlerID := fmt.Sprintf("handler_%d", time.Now().UnixNano())

	m.eventHandlers[sessionID][handlerID] = &EventHandlerInfo{
		ID:      handlerID,
		Handler: handler,
	}

	return handlerID
}

func (m *Manager) attachHandlerToClient(sessionID string, handler ports.EventHandler) {
	client := m.getClient(sessionID)
	if client != nil {
		client.GetClient().AddEventHandler(func(evt interface{}) {
			m.processEventForHandler(evt, sessionID, handler)
		})
	}
}

func (m *Manager) processEventForHandler(evt interface{}, sessionID string, handler ports.EventHandler) {
	switch e := evt.(type) {
	case *events.Message:
		m.handleMessageEvent(e, sessionID, handler)
	case *events.Connected:
		m.handleConnectionEvent(sessionID, handler, true)
	case *events.Disconnected:
		m.handleConnectionEvent(sessionID, handler, false)
	case *events.QR:
		m.handleQREvent(e, sessionID, handler)
	case *events.PairSuccess:
		m.handlePairSuccessEvent(sessionID, handler)
	}
}

func (m *Manager) handleMessageEvent(e *events.Message, sessionID string, handler ports.EventHandler) {
	m.incrementMessagesReceived(sessionID)
	msg := &ports.WameowMessage{
		ID:   e.Info.ID,
		From: e.Info.Sender.String(),
		To:   e.Info.Chat.String(),
		Body: e.Message.GetConversation(),
	}
	if err := handler.HandleMessage(sessionID, msg); err != nil {
		m.logger.ErrorWithFields("Failed to handle message event", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

func (m *Manager) handleConnectionEvent(sessionID string, handler ports.EventHandler, connected bool) {
	if err := handler.HandleConnection(sessionID, connected); err != nil {
		m.logger.ErrorWithFields("Failed to handle connection event", map[string]interface{}{
			"session_id": sessionID,
			"connected":  connected,
			"error":      err.Error(),
		})
	}
}

func (m *Manager) handleQREvent(e *events.QR, sessionID string, handler ports.EventHandler) {
	if err := handler.HandleQRCode(sessionID, e.Codes[0]); err != nil {
		m.logger.ErrorWithFields("Failed to handle QR code event", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

func (m *Manager) handlePairSuccessEvent(sessionID string, handler ports.EventHandler) {
	if err := handler.HandlePairSuccess(sessionID); err != nil {
		m.logger.ErrorWithFields("Failed to handle pair success event", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
	}
}

func (m *Manager) UnregisterEventHandler(sessionID string, handlerID string) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	sessionHandlers, exists := m.eventHandlers[sessionID]
	if !exists {
		return fmt.Errorf("no event handlers found for session %s", sessionID)
	}

	_, exists = sessionHandlers[handlerID]
	if !exists {
		return fmt.Errorf("event handler %s not found for session %s", handlerID, sessionID)
	}

	delete(sessionHandlers, handlerID)

	if len(sessionHandlers) == 0 {
		delete(m.eventHandlers, sessionID)
	}

	m.logger.InfoWithFields("Event handler unregistered", map[string]interface{}{
		"session_id": sessionID,
		"handler_id": handlerID,
	})

	return nil
}

func (m *Manager) getClient(sessionID string) *WameowClient {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return m.clients[sessionID]
}

func (m *Manager) applyProxyConfig(client *whatsmeow.Client, config *session.ProxyConfig) error {
	if client == nil {
		return fmt.Errorf("cannot apply proxy config to nil client")
	}

	if config == nil {
		return fmt.Errorf("proxy configuration is nil")
	}

	switch config.Type {
	case "http", "https", "socks5":
	default:
		return fmt.Errorf("unsupported proxy type: %s", config.Type)
	}

	return nil
}

func (m *Manager) SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	resp, err := client.SendButtonMessage(ctx, to, body, buttons)
	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	return &message.SendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

func (m *Manager) SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	resp, err := client.SendListMessage(ctx, to, body, buttonText, sections)
	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	return &message.SendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

func (m *Manager) SendReaction(sessionID, to, messageID, reaction string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SendReaction(ctx, to, messageID, reaction)
}

func (m *Manager) SendPresence(sessionID, to, presence string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SendPresence(ctx, to, presence)
}

func (m *Manager) EditMessage(sessionID, to, messageID, newText string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.EditMessage(ctx, to, messageID, newText)
}

func (m *Manager) MarkRead(sessionID, to, messageID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.MarkRead(ctx, to, messageID)
}

func (m *Manager) RevokeMessage(sessionID, to, messageID string) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	err := client.RevokeMessage(ctx, to, messageID)
	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	return &message.SendResult{
		MessageID: messageID,
		Status:    "revoked",
		Timestamp: time.Now(),
	}, nil
}

func (m *Manager) CreateGroup(sessionID, name string, participants []string, description string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	groupInfo, err := client.CreateGroup(ctx, name, participants, description)
	if err != nil {
		return nil, err
	}

	return convertToPortsGroupInfo(groupInfo), nil
}

func (m *Manager) GetGroupInfo(sessionID, groupJID string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	groupInfo, err := client.GetGroupInfo(ctx, groupJID)
	if err != nil {
		return nil, err
	}

	return convertToPortsGroupInfo(groupInfo), nil
}

func (m *Manager) ListJoinedGroups(sessionID string) ([]*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	groups, err := client.ListJoinedGroups(ctx)
	if err != nil {
		return nil, err
	}

	var result []*ports.GroupInfo
	for _, group := range groups {
		result = append(result, convertToPortsGroupInfo(group))
	}

	return result, nil
}

func (m *Manager) UpdateGroupParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.UpdateGroupParticipants(ctx, groupJID, participants, action)
}

func (m *Manager) SetGroupName(sessionID, groupJID, name string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SetGroupName(ctx, groupJID, name)
}

func (m *Manager) SetGroupDescription(sessionID, groupJID, description string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SetGroupDescription(ctx, groupJID, description)
}

func (m *Manager) SetGroupPhoto(sessionID, groupJID string, photo []byte) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	return client.SetGroupPhotoFromBytes(context.Background(), groupJID, photo)
}

func (m *Manager) GetGroupInviteLink(sessionID, groupJID string, reset bool) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return "", fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.GetGroupInviteLink(ctx, groupJID, reset)
}

func (m *Manager) JoinGroupViaLink(sessionID, inviteLink string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	groupInfo, err := client.JoinGroupViaLink(ctx, inviteLink)
	if err != nil {
		return nil, err
	}

	return convertToPortsGroupInfo(groupInfo), nil
}

func (m *Manager) LeaveGroup(sessionID, groupJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.LeaveGroup(ctx, groupJID)
}

func (m *Manager) UpdateGroupSettings(sessionID, groupJID string, announce, locked *bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.UpdateGroupSettings(ctx, groupJID, announce, locked)
}

func (m *Manager) GetGroupRequestParticipants(sessionID, groupJID string) ([]types.GroupParticipantRequest, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.GetGroupRequestParticipants(ctx, groupJID)
}

func (m *Manager) UpdateGroupRequestParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.UpdateGroupRequestParticipants(ctx, groupJID, participants, action)
}

func (m *Manager) SetGroupJoinApprovalMode(sessionID, groupJID string, requireApproval bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SetGroupJoinApprovalMode(ctx, groupJID, requireApproval)
}

func (m *Manager) SetGroupMemberAddMode(sessionID, groupJID string, mode string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	return client.SetGroupMemberAddMode(ctx, groupJID, mode)
}

func (m *Manager) GetGroupInfoFromLink(sessionID string, inviteLink string) (*types.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.GetGroupInfoFromLink(context.Background(), inviteLink)
}

func (m *Manager) GetGroupInfoFromInvite(sessionID string, jid, inviter, code string, expiration int64) (*types.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.GetGroupInfoFromInvite(context.Background(), jid, inviter, code, expiration)
}

func (m *Manager) JoinGroupWithInvite(sessionID string, jid, inviter, code string, expiration int64) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.JoinGroupWithInvite(context.Background(), jid, inviter, code, expiration)
}

func (m *Manager) CreatePoll(sessionID, to, name string, options []string, selectableCount int) (*ports.MessageInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	messageInfo, err := client.CreatePoll(ctx, to, name, options, selectableCount)
	if err != nil {
		return nil, err
	}

	return &ports.MessageInfo{
		ID:        messageInfo.ID,
		Timestamp: messageInfo.Timestamp,
		Chat:      to,
	}, nil
}

func (m *Manager) SendPoll(sessionID, to, name string, options []string, selectableCount int) (*MessageResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	toJID, err := ParseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	if name == "" {
		return nil, fmt.Errorf("poll name is required")
	}

	if len(options) < 2 {
		return nil, fmt.Errorf("at least 2 options are required")
	}

	if len(options) > 12 {
		return nil, fmt.Errorf("maximum 12 options allowed")
	}

	if selectableCount < 1 {
		selectableCount = 1 // Default to single selection
	}

	if selectableCount > len(options) {
		return nil, fmt.Errorf("selectable count cannot exceed number of options")
	}

	msgID := client.GetClient().GenerateMessageID()

	pollMessage := client.GetClient().BuildPollCreation(name, options, selectableCount)

	resp, err := client.GetClient().SendMessage(context.Background(), toJID, pollMessage, whatsmeow.SendRequestExtra{ID: msgID})
	if err != nil {
		return nil, fmt.Errorf("failed to send poll: %w", err)
	}

	return &MessageResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

func (m *Manager) GetUserJID(sessionID string) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return "", fmt.Errorf("session %s is not logged in", sessionID)
	}

	userJID := client.GetClient().Store.ID
	if userJID == nil {
		return "", fmt.Errorf("user JID not available")
	}

	return userJID.String(), nil
}

func (m *Manager) VotePoll(sessionID, to, pollMessageID string, selectedOptions []string) (*ports.MessageInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	messageInfo, err := client.VotePoll(ctx, to, pollMessageID, selectedOptions)
	if err != nil {
		return nil, err
	}

	return &ports.MessageInfo{
		ID:        messageInfo.ID,
		Timestamp: messageInfo.Timestamp,
		Chat:      to,
	}, nil
}

func (m *Manager) setupEventHandlers(client *whatsmeow.Client, sessionID string) {
	m.SetupEventHandlers(client, sessionID)
}

type MessageResult struct {
	Timestamp time.Time
	MessageID string
	Status    string
}

type ContactListResult struct {
	Timestamp     time.Time
	Results       []ContactResult
	TotalContacts int
	SuccessCount  int
	FailureCount  int
}

type ContactResult struct {
	ContactName string
	MessageID   string
	Status      string
	Error       string
}

type TextMessageResult struct {
	Timestamp time.Time
	MessageID string
	Status    string
}

func (m *Manager) SendTextMessage(sessionID, to, text string, contextInfo *appMessage.ContextInfo) (*TextMessageResult, error) {
	client, recipientJID, err := m.validateTextMessageRequest(sessionID, to)
	if err != nil {
		return nil, err
	}

	messageID, msg := m.createTextMessage(client, text, contextInfo)

	resp, _, err := m.sendTextMessageWithFallback(client, recipientJID, msg, messageID, sessionID, to)
	if err != nil {
		return nil, err
	}

	return m.logAndReturnTextResult(sessionID, to, messageID, contextInfo, resp)
}

func (m *Manager) validateTextMessageRequest(sessionID, to string) (*WameowClient, types.JID, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, types.EmptyJID, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsConnected() {
		return nil, types.EmptyJID, fmt.Errorf("session %s is not connected", sessionID)
	}

	recipientJID, err := ParseJID(to)
	if err != nil {
		return nil, types.EmptyJID, fmt.Errorf("invalid recipient JID: %w", err)
	}

	m.logger.InfoWithFields("Sending text message - JID details", map[string]interface{}{
		"session_id":  sessionID,
		"original_to": to,
		"parsed_jid":  recipientJID.String(),
		"jid_user":    recipientJID.User,
		"jid_server":  recipientJID.Server,
	})

	return client, recipientJID, nil
}

func (m *Manager) createTextMessage(client *WameowClient, text string, contextInfo *appMessage.ContextInfo) (string, *waE2E.Message) {
	messageID := client.GetClient().GenerateMessageID()

	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	if contextInfo != nil {
		waContextInfo := &waE2E.ContextInfo{
			StanzaID:      proto.String(contextInfo.StanzaID),
			QuotedMessage: &waE2E.Message{Conversation: proto.String("")},
		}

		if contextInfo.Participant != "" {
			waContextInfo.Participant = proto.String(contextInfo.Participant)
		}

		msg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text:        proto.String(text),
				ContextInfo: waContextInfo,
			},
		}
	}

	return messageID, msg
}

func (m *Manager) sendTextMessageWithFallback(client *WameowClient, recipientJID types.JID, msg *waE2E.Message, messageID, sessionID, to string) (whatsmeow.SendResponse, types.JID, error) {
	resp, err := client.GetClient().SendMessage(context.Background(), recipientJID, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		if altResp, altJID, altErr := m.tryBrazilianAlternative(client, msg, messageID, sessionID, to); altErr == nil {
			return altResp, altJID, nil
		}
		return whatsmeow.SendResponse{}, types.EmptyJID, fmt.Errorf("failed to send text message: %w", err)
	}

	return resp, recipientJID, nil
}

func (m *Manager) tryBrazilianAlternative(client *WameowClient, msg *waE2E.Message, messageID, sessionID, to string) (whatsmeow.SendResponse, types.JID, error) {
	alternativeNumber := GetBrazilianAlternativeNumber(to)
	if alternativeNumber == "" {
		return whatsmeow.SendResponse{}, types.EmptyJID, fmt.Errorf("no alternative number available")
	}

	m.logger.InfoWithFields("Trying Brazilian alternative number format", map[string]interface{}{
		"session_id":         sessionID,
		"original_number":    to,
		"alternative_number": alternativeNumber,
	})

	altRecipientJID, altErr := ParseJID(alternativeNumber)
	if altErr != nil {
		return whatsmeow.SendResponse{}, types.EmptyJID, altErr
	}

	resp, err := client.GetClient().SendMessage(context.Background(), altRecipientJID, msg, whatsmeow.SendRequestExtra{ID: messageID})
	if err != nil {
		return whatsmeow.SendResponse{}, types.EmptyJID, err
	}

	m.logger.InfoWithFields("Message sent successfully with alternative Brazilian format", map[string]interface{}{
		"session_id":         sessionID,
		"original_number":    to,
		"alternative_number": alternativeNumber,
		"used_jid":           altRecipientJID.String(),
	})

	return resp, altRecipientJID, nil
}

func (m *Manager) logAndReturnTextResult(sessionID, to, messageID string, contextInfo *appMessage.ContextInfo, resp whatsmeow.SendResponse) (*TextMessageResult, error) {
	m.logger.InfoWithFields("Text message sent", map[string]interface{}{
		"session_id": sessionID,
		"to":         to,
		"message_id": messageID,
		"has_reply":  contextInfo != nil,
		"timestamp":  resp.Timestamp,
	})

	return &TextMessageResult{
		MessageID: messageID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

func (m *Manager) SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string, contextInfo *message.ContextInfo) (*message.SendResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	var resp *whatsmeow.SendResponse
	var err error

	var appContextInfo *appMessage.ContextInfo
	if contextInfo != nil {
		appContextInfo = &appMessage.ContextInfo{
			StanzaID:    contextInfo.StanzaID,
			Participant: contextInfo.Participant,
		}
	}

	switch messageType {
	case "text":
		textResult, err := m.SendTextMessage(sessionID, to, body, appContextInfo)
		if err != nil {
			return nil, err
		}
		return &message.SendResult{
			MessageID: textResult.MessageID,
			Status:    textResult.Status,
			Timestamp: textResult.Timestamp,
		}, nil
	case "image":
		resp, err = client.SendImageMessage(ctx, to, file, caption, appContextInfo)
	case "audio":
		resp, err = client.SendAudioMessage(ctx, to, file, appContextInfo)
	case "video":
		resp, err = client.SendVideoMessage(ctx, to, file, caption, appContextInfo)
	case "document":
		resp, err = client.SendDocumentMessage(ctx, to, file, filename, caption, appContextInfo)
	case "location":
		resp, err = client.SendLocationMessage(ctx, to, latitude, longitude, body)
	case "contact":
		resp, err = client.SendContactMessage(ctx, to, contactName, contactPhone)
	case "sticker":
		resp, err = client.SendStickerMessage(ctx, to, file)
	default:
		return nil, fmt.Errorf("unsupported message type: %s", messageType)
	}

	if err != nil {
		return &message.SendResult{
			Status:    "failed",
			Error:     err.Error(),
			Timestamp: time.Now(),
		}, err
	}

	m.incrementMessagesSent(sessionID)

	return &message.SendResult{
		MessageID: resp.ID,
		Status:    "sent",
		Timestamp: resp.Timestamp,
	}, nil
}

func (m *Manager) SendContactList(sessionID, to string, contacts []ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if !client.IsLoggedIn() {
		return nil, fmt.Errorf("session %s is not logged in", sessionID)
	}

	ctx := context.Background()
	result := &ContactListResult{
		TotalContacts: len(contacts),
		Results:       make([]ContactResult, 0, len(contacts)),
		Timestamp:     time.Now(),
	}

	var wameowContacts []ContactInfo
	for _, contact := range contacts {
		wameowContacts = append(wameowContacts, ContactInfo{
			Name:         contact.Name,
			Phone:        contact.Phone,
			Email:        contact.Email,
			Organization: contact.Organization,
			Title:        contact.Title,
			Website:      contact.Website,
			Address:      contact.Address,
		})
	}

	resp, err := client.SendContactListMessage(ctx, to, wameowContacts)
	if err != nil {
		for _, contact := range contacts {
			result.Results = append(result.Results, ContactResult{
				ContactName: contact.Name,
				Status:      "failed",
				Error:       err.Error(),
			})
		}
		result.FailureCount = len(contacts)
		return result, err
	}

	for _, contact := range contacts {
		result.Results = append(result.Results, ContactResult{
			ContactName: contact.Name,
			MessageID:   resp.ID,
			Status:      "sent",
		})
	}
	result.SuccessCount = len(contacts)

	return result, nil
}

func (m *Manager) SendContactListBusiness(sessionID, to string, contacts []ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	var wameowContacts []ContactInfo
	for _, contact := range contacts {
		wameowContacts = append(wameowContacts, ContactInfo{
			Name:         contact.Name,
			Phone:        contact.Phone,
			Email:        contact.Email,
			Organization: contact.Organization,
			Title:        contact.Title,
			Website:      contact.Website,
			Address:      contact.Address,
		})
	}

	resp, err := client.SendContactListMessageBusiness(context.Background(), to, wameowContacts)
	if err != nil {
		return nil, fmt.Errorf("failed to send WhatsApp Business contact list: %w", err)
	}

	result := &ContactListResult{
		TotalContacts: len(contacts),
		SuccessCount:  len(contacts),
		FailureCount:  0,
		Results:       make([]ContactResult, len(contacts)),
		Timestamp:     time.Now(),
	}

	for i, contact := range contacts {
		result.Results[i] = ContactResult{
			ContactName: contact.Name,
			MessageID:   resp.ID,
			Status:      "sent",
		}
	}

	return result, nil
}

func (m *Manager) sendSingleContactGeneric(
	sessionID, to string,
	contact ContactInfo,
	sendFunc func(context.Context, string, ContactInfo) (*whatsmeow.SendResponse, error),
	errorMsg string,
) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("session %s is not connected", sessionID)
	}

	resp, err := sendFunc(context.Background(), to, ContactInfo{
		Name:         contact.Name,
		Phone:        contact.Phone,
		Email:        contact.Email,
		Organization: contact.Organization,
		Title:        contact.Title,
		Website:      contact.Website,
		Address:      contact.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorMsg, err)
	}

	result := &ContactListResult{
		TotalContacts: 1,
		SuccessCount:  1,
		FailureCount:  0,
		Results:       make([]ContactResult, 1),
		Timestamp:     time.Now(),
	}

	result.Results[0] = ContactResult{
		ContactName: contact.Name,
		MessageID:   resp.ID,
		Status:      "sent",
	}

	return result, nil
}

func (m *Manager) SendSingleContact(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}
	return m.sendSingleContactGeneric(sessionID, to, contact, client.SendSingleContactMessage, "failed to send single contact")
}

func (m *Manager) SendSingleContactBusiness(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}
	return m.sendSingleContactGeneric(sessionID, to, contact, client.SendSingleContactMessageBusiness, "failed to send Business single contact")
}

func (m *Manager) SendSingleContactBusinessFormat(sessionID, to string, contact ContactInfo) (*ContactListResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}
	return m.sendSingleContactGeneric(sessionID, to, contact, client.SendSingleContactMessageBusiness, "failed to send WhatsApp Business single contact")
}

func (m *Manager) IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.IsOnWhatsApp(ctx, phoneNumbers)
}

func (m *Manager) GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.GetAllContacts(ctx)
}

func (m *Manager) GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.GetProfilePictureInfo(ctx, jid, preview)
}

func (m *Manager) GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.GetUserInfo(ctx, jids)
}

func (m *Manager) GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return client.GetBusinessProfile(ctx, jid)
}

func (m *Manager) SetupEventHandlers(client *whatsmeow.Client, sessionID string) {
	eventHandler := NewEventHandler(m, m.sessionMgr, m.qrGenerator, m.logger)

	if m.webhookHandler != nil {
		eventHandler.SetWebhookHandler(m.webhookHandler)
	}

	if m.chatwootManager != nil {
		eventHandler.SetChatwootManager(m.chatwootManager)
	}

	client.AddEventHandler(func(evt interface{}) {
		eventHandler.HandleEvent(evt, sessionID)
	})
}

func (m *Manager) SetWebhookHandler(handler WebhookEventHandler) {
	m.webhookHandler = handler
	m.logger.Info("Webhook handler configured for wameow manager")
}

func (m *Manager) SetChatwootManager(manager ChatwootManager) {
	m.chatwootManager = manager
	m.logger.Info("Chatwoot manager configured for wameow manager")
}

func convertToPortsGroupInfo(groupInfo interface{}) *ports.GroupInfo {
	if gi, ok := groupInfo.(*types.GroupInfo); ok {
		participants := make([]ports.GroupParticipant, len(gi.Participants))
		for i, p := range gi.Participants {
			participants[i] = ports.GroupParticipant{
				JID:          p.JID.String(),
				IsAdmin:      p.IsAdmin,
				IsSuperAdmin: p.IsSuperAdmin,
			}
		}

		return &ports.GroupInfo{
			GroupJID:     gi.JID.String(),
			Name:         gi.Name,  // Usar gi.GroupName.Name
			Description:  gi.Topic, // Usar gi.GroupTopic.Topic
			Owner:        gi.OwnerJID.String(),
			Participants: participants,
			Settings: ports.GroupSettings{
				Announce: gi.IsAnnounce, // Usar gi.GroupAnnounce.IsAnnounce
				Locked:   gi.IsLocked,   // Usar gi.IsLocked
			},
			CreatedAt: gi.GroupCreated,
			UpdatedAt: time.Now(),
		}
	}

	return &ports.GroupInfo{
		GroupJID:     "unknown@g.us",
		Name:         "Unknown Group",
		Description:  "",
		Owner:        "",
		Participants: []ports.GroupParticipant{},
		Settings: ports.GroupSettings{
			Announce: false,
			Locked:   false,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
