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

// WhatsmeowLogger adapta logger para whatsmeow
type WhatsmeowLogger struct {
	logger *logger.Logger
}

// NewWhatsmeowLogger cria novo logger para whatsmeow
func NewWhatsmeowLogger(logger *logger.Logger) waLog.Logger {
	return &WhatsmeowLogger{logger: logger}
}

func (w *WhatsmeowLogger) Errorf(msg string, args ...interface{}) {
	w.logger.ErrorWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": "whatsmeow",
	})
}

func (w *WhatsmeowLogger) Warnf(msg string, args ...interface{}) {
	w.logger.WarnWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": "whatsmeow",
	})
}

func (w *WhatsmeowLogger) Infof(msg string, args ...interface{}) {
	w.logger.InfoWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": "whatsmeow",
	})
}

func (w *WhatsmeowLogger) Debugf(msg string, args ...interface{}) {
	w.logger.DebugWithFields(fmt.Sprintf(msg, args...), map[string]interface{}{
		"module": "whatsmeow",
	})
}

func (w *WhatsmeowLogger) Sub(module string) waLog.Logger {
	return &WhatsmeowLogger{logger: w.logger}
}

// QRCodeEvent representa evento de QR code
type QRCodeEvent struct {
	SessionName string
	QRCode      string
	ExpiresAt   time.Time
}

// Client encapsula um cliente whatsmeow para uma sessão específica
type Client struct {
	// Identificação
	sessionName string

	// WhatsApp client
	client *whatsmeow.Client
	device *store.Device

	// Dependencies
	logger      *logger.Logger
	qrGenerator *QRGenerator

	// State management
	mu            sync.RWMutex
	isConnected   bool
	isLoggedIn    bool
	status        string
	lastActivity  time.Time

	// QR Code management
	qrCode        string
	qrCodeExpires time.Time
	qrChannel     <-chan whatsmeow.QRChannelItem
	qrContext     context.Context
	qrCancel      context.CancelFunc

	// Event handling
	eventHandler  func(interface{})
	eventHandlers []func(interface{})

	// Configuration
	proxyConfig *session.ProxyConfig

	// Connection management
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClient cria novo cliente WhatsApp baseado no legacy funcional
func NewClient(sessionName string, container *sqlstore.Container, logger *logger.Logger) (*Client, error) {
	if sessionName == "" {
		return nil, fmt.Errorf("session name cannot be empty")
	}

	if container == nil {
		return nil, fmt.Errorf("sqlstore container cannot be nil")
	}

	// Criar novo device store para esta sessão (baseado no legacy)
	deviceStore := container.NewDevice()
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store")
	}

	// Criar logger compatível com whatsmeow
	waLogger := NewWhatsmeowLogger(logger)

	// Criar cliente whatsmeow
	whatsmeowClient := whatsmeow.NewClient(deviceStore, waLogger)

	// Criar contexto para gerenciamento de lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		sessionName:   sessionName,
		client:        whatsmeowClient,
		device:        deviceStore,
		logger:        logger,
		qrGenerator:   NewQRGenerator(logger),
		status:        "disconnected",
		lastActivity:  time.Now(),
		eventHandlers: make([]func(interface{}), 0),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Configurar event handlers
	client.setupEventHandlers()

	// WhatsApp client created

	return client, nil
}

// NewClientWithDevice cria cliente WhatsApp com device existente
func NewClientWithDevice(sessionName string, deviceStore *store.Device, container *sqlstore.Container, logger *logger.Logger) (*Client, error) {
	if sessionName == "" {
		return nil, fmt.Errorf("session name cannot be empty")
	}

	if deviceStore == nil {
		return nil, fmt.Errorf("device store cannot be nil")
	}

	if container == nil {
		return nil, fmt.Errorf("sqlstore container cannot be nil")
	}

	// Criar logger compatível com whatsmeow
	waLogger := NewWhatsmeowLogger(logger)

	// Criar cliente whatsmeow com device existente
	whatsmeowClient := whatsmeow.NewClient(deviceStore, waLogger)

	// Criar contexto para gerenciamento de lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		sessionName:   sessionName,
		client:        whatsmeowClient,
		device:        deviceStore,
		logger:        logger,
		qrGenerator:   NewQRGenerator(logger),
		status:        "disconnected",
		lastActivity:  time.Now(),
		eventHandlers: make([]func(interface{}), 0),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Configurar event handlers
	client.setupEventHandlers()

	logger.InfoWithFields("WhatsApp client created with existing device", map[string]interface{}{
		"module":  "client",
		"session": sessionName,
	})

	return client, nil
}

// Connect conecta o cliente ao WhatsApp baseado no legacy
func (c *Client) Connect() error {
	// Starting connection process

	// Parar qualquer processo de QR code ativo
	c.stopQRProcess()

	// Se já conectado, desconectar primeiro
	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	// Resetar contexto
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	c.setStatus("connecting")

	// Iniciar processo de conexão em goroutine
	go c.startConnectionLoop()

	return nil
}

// startConnectionLoop inicia o loop de conexão baseado no legacy
func (c *Client) startConnectionLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("Connection loop panic", map[string]interface{}{
				"session_name": c.sessionName,
				"error":        r,
			})
		}
	}()

	// Verificar se device está registrado
	isRegistered := c.isDeviceRegistered()

	if !isRegistered {
		c.handleNewDeviceRegistration()
	} else {
		c.handleExistingDeviceConnection()
	}
}

// isDeviceRegistered verifica se o device está registrado
func (c *Client) isDeviceRegistered() bool {
	return c.device.ID != nil
}

// handleNewDeviceRegistration lida com registro de novo device
func (c *Client) handleNewDeviceRegistration() {
	qrChan, err := c.client.GetQRChannel(c.ctx)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get QR channel", map[string]interface{}{
			"session_name": c.sessionName,
			"error":        err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	err = c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect client", map[string]interface{}{
			"session_name": c.sessionName,
			"error":        err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	c.handleQRLoop(qrChan)
}

// handleExistingDeviceConnection lida com conexão de device existente
func (c *Client) handleExistingDeviceConnection() {
	err := c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect existing device", map[string]interface{}{
			"session_name": c.sessionName,
			"error":        err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	// Existing device connected successfully
}

// handleQRLoop processa o loop de QR code
func (c *Client) handleQRLoop(qrChan <-chan whatsmeow.QRChannelItem) {
	c.mu.Lock()
	c.qrChannel = qrChan
	c.qrContext, c.qrCancel = context.WithCancel(c.ctx)
	c.mu.Unlock()

	for {
		select {
		case <-c.qrContext.Done():
			c.logger.InfoWithFields("QR loop cancelled", map[string]interface{}{
				"session_name": c.sessionName,
			})
			return

		case evt, ok := <-qrChan:
			if !ok {
				c.logger.InfoWithFields("QR channel closed", map[string]interface{}{
					"session_name": c.sessionName,
				})
				return
			}

			if evt.Event == "code" {
				c.handleQRCode(evt.Code)
			} else if evt.Event == "timeout" {
				c.logger.WarnWithFields("QR code timeout", map[string]interface{}{
					"session_name": c.sessionName,
				})
				c.setStatus("disconnected")
				return
			}
		}
	}
}

// handleQRCodeEvent processa evento de QR code
func (c *Client) handleQRCodeEvent(evt *events.QR) {
	c.handleQRCode(evt.Codes[0])
}

// handleQRCode processa novo QR code
func (c *Client) handleQRCode(qrCode string) {
	c.mu.Lock()
	c.qrCode = qrCode
	c.qrCodeExpires = time.Now().Add(30 * time.Second) // QR codes expiram em 30s
	c.mu.Unlock()

	c.logger.InfoWithFields("QR code generated", map[string]interface{}{
		"session_name": c.sessionName,
		"qr_code":      qrCode,
	})

	// Notificar event handlers sobre novo QR code
	// A exibição do QR code será feita pelo EventHandler
	c.notifyEventHandlers(&QRCodeEvent{
		SessionName: c.sessionName,
		QRCode:      qrCode,
		ExpiresAt:   c.qrCodeExpires,
	})
}

// stopQRProcess para o processo de QR code
func (c *Client) stopQRProcess() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.qrCancel != nil {
		c.qrCancel()
		c.qrCancel = nil
	}

	c.qrCode = ""
	c.qrCodeExpires = time.Time{}
}

// setStatus atualiza o status do cliente
func (c *Client) setStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = status
	c.lastActivity = time.Now()

	c.logger.DebugWithFields("Client status updated", map[string]interface{}{
		"session_name": c.sessionName,
		"status":       status,
	})
}

// setupEventHandlers configura os event handlers do whatsmeow
func (c *Client) setupEventHandlers() {
	c.client.AddEventHandler(c.handleEvent)
}

// handleEvent processa eventos do whatsmeow
func (c *Client) handleEvent(evt interface{}) {
	// Atualizar última atividade
	c.mu.Lock()
	c.lastActivity = time.Now()
	c.mu.Unlock()

	// Processar eventos específicos
	switch v := evt.(type) {
	case *events.Connected:
		c.handleConnectedEvent(v)
	case *events.Disconnected:
		c.handleDisconnectedEvent(v)
	case *events.LoggedOut:
		c.handleLoggedOutEvent(v)
	case *events.QR:
		// QR code já é tratado pelo loop de QR code, não processar aqui
	case *events.PairSuccess:
		c.handlePairSuccessEvent(v)
	case *events.PairError:
		c.handlePairErrorEvent(v)
	default:
		// Log outros eventos para debug
		c.logger.DebugWithFields("WhatsApp event received", map[string]interface{}{
			"session_name": c.sessionName,
			"event_type":   fmt.Sprintf("%T", evt),
		})
	}

	// Notificar event handlers externos
	c.notifyEventHandlers(evt)
}

// notifyEventHandlers notifica todos os event handlers registrados
func (c *Client) notifyEventHandlers(evt interface{}) {
	c.mu.RLock()
	handlers := make([]func(interface{}), len(c.eventHandlers))
	copy(handlers, c.eventHandlers)
	c.mu.RUnlock()

	for _, handler := range handlers {
		go func(h func(interface{})) {
			defer func() {
				if r := recover(); r != nil {
					c.logger.ErrorWithFields("Event handler panic", map[string]interface{}{
						"session_name": c.sessionName,
						"error":        r,
					})
				}
			}()
			h(evt)
		}(handler)
	}
}

// handleConnectedEvent processa evento de conexão
func (c *Client) handleConnectedEvent(evt *events.Connected) {
	c.mu.Lock()
	c.isConnected = true
	c.mu.Unlock()

	c.setStatus("connected")

	c.logger.InfoWithFields("WhatsApp client connected", map[string]interface{}{
		"module":  "client",
		"session": c.sessionName,
	})
}

// handleDisconnectedEvent processa evento de desconexão
func (c *Client) handleDisconnectedEvent(evt *events.Disconnected) {
	c.mu.Lock()
	c.isConnected = false
	c.isLoggedIn = false
	c.mu.Unlock()

	c.setStatus("disconnected")

	c.logger.WarnWithFields("WhatsApp client disconnected", map[string]interface{}{
		"session_name": c.sessionName,
	})
}

// handleLoggedOutEvent processa evento de logout
func (c *Client) handleLoggedOutEvent(evt *events.LoggedOut) {
	c.mu.Lock()
	c.isLoggedIn = false
	c.mu.Unlock()

	c.setStatus("logged_out")

	c.logger.WarnWithFields("WhatsApp client logged out", map[string]interface{}{
		"session_name": c.sessionName,
		"reason":       evt.Reason,
	})
}

// handlePairSuccessEvent processa evento de pareamento bem-sucedido
func (c *Client) handlePairSuccessEvent(evt *events.PairSuccess) {
	c.mu.Lock()
	c.isLoggedIn = true
	c.mu.Unlock()

	c.setStatus("logged_in")
	c.stopQRProcess()

	c.logger.InfoWithFields("WhatsApp pairing successful", map[string]interface{}{
		"session_name": c.sessionName,
		"jid":          evt.ID.String(),
	})
}

// handlePairErrorEvent processa evento de erro de pareamento
func (c *Client) handlePairErrorEvent(evt *events.PairError) {
	c.setStatus("pair_error")

	c.logger.ErrorWithFields("WhatsApp pairing failed", map[string]interface{}{
		"session_name": c.sessionName,
		"error":        evt.Error.Error(),
	})
}

// Disconnect desconecta o cliente baseado no legacy
func (c *Client) Disconnect() error {
	c.logger.InfoWithFields("Starting client disconnection", map[string]interface{}{
		"session_name": c.sessionName,
	})

	c.mu.Lock()
	defer c.mu.Unlock()

	// Parar processo de QR code
	c.stopQRProcess()

	// Desconectar cliente whatsmeow
	if c.client.IsConnected() {
		c.logger.InfoWithFields("Disconnecting whatsmeow client", map[string]interface{}{
			"session_name": c.sessionName,
		})
		c.client.Disconnect()
	}

	// Cancelar contexto
	if c.cancel != nil {
		c.logger.InfoWithFields("Canceling client context", map[string]interface{}{
			"session_name": c.sessionName,
		})
		c.cancel()
	}

	c.setStatus("disconnected")

	c.logger.InfoWithFields("Client disconnection completed", map[string]interface{}{
		"session_name": c.sessionName,
	})

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
func (c *Client) GetQRCode() (string, error) {
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

	if c.qrCode == "" {
		return "", fmt.Errorf("no QR code available")
	}

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



// GetStatus retorna status atual do cliente
func (c *Client) GetStatus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// AddEventHandler adiciona um event handler
func (c *Client) AddEventHandler(handler func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers = append(c.eventHandlers, handler)
}