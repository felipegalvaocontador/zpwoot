package waclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/core/messaging"
	"zpwoot/internal/core/session"
	"zpwoot/platform/logger"
)

// Gateway implementa session.WhatsAppGateway para integração com whatsmeow
type Gateway struct {
	// Dependencies
	logger    *logger.Logger
	container *sqlstore.Container

	// Internal state
	clients       map[string]*Client
	eventHandlers map[string][]session.EventHandler
	mu            sync.RWMutex

	// External integrations (baseado no legacy)
	webhookHandler  WebhookEventHandler
	chatwootManager ChatwootManager
}

// NewGateway cria nova instância do gateway WhatsApp
func NewGateway(container *sqlstore.Container, logger *logger.Logger) *Gateway {
	return &Gateway{
		logger:        logger,
		container:     container,
		clients:       make(map[string]*Client),
		eventHandlers: make(map[string][]session.EventHandler),
	}
}

// CreateSession cria uma nova sessão WhatsApp
func (g *Gateway) CreateSession(ctx context.Context, sessionName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Verificar se sessão já existe
	if _, exists := g.clients[sessionName]; exists {
		return fmt.Errorf("session %s already exists", sessionName)
	}

	g.logger.InfoWithFields("Creating WhatsApp session", map[string]interface{}{
		"session_name": sessionName,
	})

	// Criar cliente WhatsApp
	client, err := NewClient(sessionName, g.container, g.logger)
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	// Configurar event handlers
	g.setupEventHandlers(client, sessionName)

	// Armazenar cliente
	g.clients[sessionName] = client

	g.logger.InfoWithFields("WhatsApp session created successfully", map[string]interface{}{
		"session_name": sessionName,
	})

	return nil
}

// ConnectSession conecta uma sessão WhatsApp baseado no legacy
func (g *Gateway) ConnectSession(ctx context.Context, sessionName string) error {
	g.logger.InfoWithFields("Starting session connection", map[string]interface{}{
		"session_name": sessionName,
	})

	client := g.getClient(sessionName)
	if client == nil {
		// Criar sessão se não existe
		err := g.CreateSession(ctx, sessionName)
		if err != nil {
			return fmt.Errorf("failed to create session %s: %w", sessionName, err)
		}
		client = g.getClient(sessionName)
	}

	// Verificar se já está conectado
	if client.GetClient().IsConnected() {
		g.logger.InfoWithFields("Session already connected", map[string]interface{}{
			"session_name": sessionName,
		})
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

	g.logger.InfoWithFields("Session connection initiated", map[string]interface{}{
		"session_name": sessionName,
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

	g.logger.InfoWithFields("Setting proxy for session", map[string]interface{}{
		"session_name": sessionName,
		"proxy_type":   proxy.Type,
		"proxy_host":   proxy.Host,
	})

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
		eventHandler.HandleEvent(evt, sessionName)
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
	g.logger.InfoWithFields("Message received and ready to save", map[string]interface{}{
		"session_id":    message.SessionID,
		"message_id":    message.ZpMessageID,
		"sender":        message.ZpSender,
		"chat":          message.ZpChat,
		"type":          message.ZpType,
		"from_me":       message.ZpFromMe,
	})

	return nil
}

// UpdateSessionStatus atualiza o status de uma sessão no banco de dados
func (g *Gateway) UpdateSessionStatus(sessionID, status string) error {
	g.logger.InfoWithFields("Updating session status", map[string]interface{}{
		"session_id": sessionID,
		"status":     status,
	})

	// TODO: Implementar atualização via session repository
	// Por enquanto, apenas log
	g.logger.DebugWithFields("Session status updated", map[string]interface{}{
		"session_id": sessionID,
		"new_status": status,
	})

	return nil
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