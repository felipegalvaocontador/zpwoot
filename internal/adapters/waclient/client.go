package waclient

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"

	"zpwoot/internal/core/session"
	"zpwoot/platform/logger"
)

// Client encapsula um cliente whatsmeow para uma sessão específica
type Client struct {
	// Identificação
	sessionName string

	// WhatsApp client
	client *whatsmeow.Client
	device *store.Device

	// Dependencies
	logger *logger.Logger

	// State management
	mu            sync.RWMutex
	isConnected   bool
	isLoggedIn    bool
	qrCode        string
	qrCodeExpires time.Time
	eventHandler  func(interface{})

	// Configuration
	proxyConfig *session.ProxyConfig
}

// NewClient cria novo cliente WhatsApp
func NewClient(sessionName string, container *sqlstore.Container, logger *logger.Logger) (*Client, error) {
	ctx := context.Background()

	// Obter ou criar device store
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		// Se não existe, criar novo device
		deviceStore = container.NewDevice()
	}

	// Criar logger compatível com whatsmeow
	waLogger := waLog.Noop // Por enquanto usar noop, depois implementar adapter

	// Criar cliente whatsmeow
	whatsmeowClient := whatsmeow.NewClient(deviceStore, waLogger)

	client := &Client{
		sessionName: sessionName,
		client:      whatsmeowClient,
		device:      deviceStore,
		logger:      logger,
	}

	// Configurar event handler interno
	whatsmeowClient.AddEventHandler(client.handleEvent)

	return client, nil
}

// Connect conecta o cliente ao WhatsApp
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnected {
		return fmt.Errorf("client is already connected")
	}

	c.logger.InfoWithFields("Connecting WhatsApp client", map[string]interface{}{
		"session_name": c.sessionName,
	})

	// Configurar proxy se necessário
	if c.proxyConfig != nil {
		if err := c.configureProxy(); err != nil {
			return fmt.Errorf("failed to configure proxy: %w", err)
		}
	}

	// Conectar
	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WhatsApp: %w", err)
	}

	c.isConnected = true

	// Verificar se já está logado
	if c.client.IsLoggedIn() {
		c.isLoggedIn = true
		c.logger.InfoWithFields("WhatsApp client connected and logged in", map[string]interface{}{
			"session_name": c.sessionName,
			"jid":          c.client.Store.ID.String(),
		})
	} else {
		c.logger.InfoWithFields("WhatsApp client connected but not logged in", map[string]interface{}{
			"session_name": c.sessionName,
		})
	}

	return nil
}

// Disconnect desconecta o cliente
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isConnected {
		return fmt.Errorf("client is not connected")
	}

	c.logger.InfoWithFields("Disconnecting WhatsApp client", map[string]interface{}{
		"session_name": c.sessionName,
	})

	c.client.Disconnect()
	c.isConnected = false

	return nil
}

// Logout faz logout da sessão
func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isLoggedIn {
		return fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Logging out WhatsApp client", map[string]interface{}{
		"session_name": c.sessionName,
	})

	if err := c.client.Logout(context.Background()); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	c.isLoggedIn = false
	c.qrCode = ""
	c.qrCodeExpires = time.Time{}

	return nil
}

// IsConnected verifica se cliente está conectado
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isConnected
}

// IsLoggedIn verifica se cliente está logado
func (c *Client) IsLoggedIn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isLoggedIn
}

// GetQRCode obtém QR code para pareamento
func (c *Client) GetQRCode(ctx context.Context) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isLoggedIn {
		return "", fmt.Errorf("client is already logged in")
	}

	if !c.isConnected {
		return "", fmt.Errorf("client is not connected")
	}

	// Se já temos QR code válido, retornar
	if c.qrCode != "" && time.Now().Before(c.qrCodeExpires) {
		return c.qrCode, nil
	}

	// Aguardar QR code ser gerado pelos eventos
	// TODO: Implementar timeout e canal para aguardar QR code
	return c.qrCode, nil
}

// SetProxy configura proxy para o cliente
func (c *Client) SetProxy(proxy *session.ProxyConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.proxyConfig = proxy

	// Se já conectado, aplicar proxy imediatamente
	if c.isConnected {
		if err := c.configureProxy(); err != nil {
			return fmt.Errorf("failed to apply proxy configuration: %w", err)
		}
	}

	return nil
}

// SetEventHandler configura handler de eventos
func (c *Client) SetEventHandler(handler func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandler = handler
}

// GetJID retorna JID do cliente
func (c *Client) GetJID() types.JID {
	if c.client.Store.ID == nil {
		return types.EmptyJID
	}
	return *c.client.Store.ID
}

// GetClient retorna cliente whatsmeow subjacente
func (c *Client) GetClient() *whatsmeow.Client {
	return c.client
}

// ===== MÉTODOS PRIVADOS =====

// configureProxy configura proxy HTTP para o cliente
func (c *Client) configureProxy() error {
	if c.proxyConfig == nil {
		return nil
	}

	// Construir URL do proxy
	var proxyURL *url.URL
	var err error

	switch c.proxyConfig.Type {
	case "http":
		if c.proxyConfig.Username != "" && c.proxyConfig.Password != "" {
			proxyURL, err = url.Parse(fmt.Sprintf("http://%s:%s@%s:%d",
				c.proxyConfig.Username, c.proxyConfig.Password,
				c.proxyConfig.Host, c.proxyConfig.Port))
		} else {
			proxyURL, err = url.Parse(fmt.Sprintf("http://%s:%d",
				c.proxyConfig.Host, c.proxyConfig.Port))
		}
	case "socks5":
		if c.proxyConfig.Username != "" && c.proxyConfig.Password != "" {
			proxyURL, err = url.Parse(fmt.Sprintf("socks5://%s:%s@%s:%d",
				c.proxyConfig.Username, c.proxyConfig.Password,
				c.proxyConfig.Host, c.proxyConfig.Port))
		} else {
			proxyURL, err = url.Parse(fmt.Sprintf("socks5://%s:%d",
				c.proxyConfig.Host, c.proxyConfig.Port))
		}
	default:
		return fmt.Errorf("unsupported proxy type: %s", c.proxyConfig.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	// Configurar proxy no cliente HTTP
	c.client.SetProxyAddress(proxyURL.String())

	c.logger.InfoWithFields("Proxy configured", map[string]interface{}{
		"session_name": c.sessionName,
		"proxy_type":   c.proxyConfig.Type,
		"proxy_host":   c.proxyConfig.Host,
		"proxy_port":   c.proxyConfig.Port,
	})

	return nil
}

// handleEvent processa eventos do whatsmeow
func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		c.handleConnected(v)
	case *events.Disconnected:
		c.handleDisconnected(v)
	case *events.LoggedOut:
		c.handleLoggedOut(v)
	case *events.QR:
		c.handleQRCode(v)
	case *events.PairSuccess:
		c.handlePairSuccess(v)
	default:
		// Log outros eventos para debug
		c.logger.DebugWithFields("WhatsApp event received", map[string]interface{}{
			"session_name": c.sessionName,
			"event_type":   fmt.Sprintf("%T", evt),
		})
	}

	// Repassar evento para handler externo se configurado
	if c.eventHandler != nil {
		c.eventHandler(evt)
	}
}

// handleConnected processa evento de conexão
func (c *Client) handleConnected(evt *events.Connected) {
	c.mu.Lock()
	c.isConnected = true
	c.mu.Unlock()

	c.logger.InfoWithFields("WhatsApp connected", map[string]interface{}{
		"session_name": c.sessionName,
	})
}

// handleDisconnected processa evento de desconexão
func (c *Client) handleDisconnected(evt *events.Disconnected) {
	c.mu.Lock()
	c.isConnected = false
	c.mu.Unlock()

	c.logger.InfoWithFields("WhatsApp disconnected", map[string]interface{}{
		"session_name": c.sessionName,
	})
}

// handleLoggedOut processa evento de logout
func (c *Client) handleLoggedOut(evt *events.LoggedOut) {
	c.mu.Lock()
	c.isLoggedIn = false
	c.qrCode = ""
	c.qrCodeExpires = time.Time{}
	c.mu.Unlock()

	c.logger.InfoWithFields("WhatsApp logged out", map[string]interface{}{
		"session_name": c.sessionName,
		"reason":       evt.Reason,
	})
}

// handleQRCode processa evento de QR code
func (c *Client) handleQRCode(evt *events.QR) {
	c.mu.Lock()
	c.qrCode = evt.Codes[0]
	c.qrCodeExpires = time.Now().Add(2 * time.Minute)
	c.mu.Unlock()

	c.logger.InfoWithFields("QR code generated", map[string]interface{}{
		"session_name": c.sessionName,
		"expires_at":   c.qrCodeExpires,
	})
}

// handlePairSuccess processa evento de pareamento bem-sucedido
func (c *Client) handlePairSuccess(evt *events.PairSuccess) {
	c.mu.Lock()
	c.isLoggedIn = true
	c.qrCode = ""
	c.qrCodeExpires = time.Time{}
	c.mu.Unlock()

	c.logger.InfoWithFields("WhatsApp paired successfully", map[string]interface{}{
		"session_name": c.sessionName,
		"jid":          evt.ID.String(),
	})
}