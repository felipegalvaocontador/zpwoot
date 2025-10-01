package waclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"

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

// ConnectSession conecta uma sessão WhatsApp
func (g *Gateway) ConnectSession(ctx context.Context, sessionName string) error {
	client := g.getClient(sessionName)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionName)
	}

	g.logger.InfoWithFields("Connecting WhatsApp session", map[string]interface{}{
		"session_name": sessionName,
	})

	if err := client.Connect(ctx); err != nil {
		g.logger.ErrorWithFields("Failed to connect WhatsApp session", map[string]interface{}{
			"session_name": sessionName,
			"error":        err.Error(),
		})
		return fmt.Errorf("failed to connect session: %w", err)
	}

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

// IsSessionConnected verifica se uma sessão está conectada
func (g *Gateway) IsSessionConnected(ctx context.Context, sessionName string) (bool, error) {
	client := g.getClient(sessionName)
	if client == nil {
		return false, fmt.Errorf("session %s not found", sessionName)
	}

	return client.IsConnected(), nil
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
		if err := client.Connect(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect for QR generation: %w", err)
		}
	}

	// Obter QR code
	qrCode, err := client.GetQRCode(ctx)
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

// setupEventHandlers configura handlers de eventos para um cliente
func (g *Gateway) setupEventHandlers(client *Client, sessionName string) {
	// Configurar handler interno para eventos do whatsmeow
	client.SetEventHandler(func(evt interface{}) {
		g.handleWhatsmeowEvent(evt, sessionName)
	})
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

// GetSessionInfo implementa session.WhatsAppGateway.GetSessionInfo
// TODO: Implementar busca real de informações da sessão
func (g *Gateway) GetSessionInfo(ctx context.Context, sessionName string) (*session.DeviceInfo, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	_, exists := g.clients[sessionName]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionName)
	}

	// Por enquanto retorna informações básicas
	// TODO: Implementar busca real de device info do whatsmeow
	return &session.DeviceInfo{
		Platform:    "whatsmeow",
		DeviceModel: "zpwoot-gateway",
		OSVersion:   "1.0.0",
		AppVersion:  "2.0.0",
	}, nil
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