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

type WhatsmeowLogger struct {
	logger *logger.Logger
}

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

type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateLoggedIn
	StateError
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "disconnected"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateLoggedIn:
		return "logged_in"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

type QRCodeEvent struct {
	SessionName string
	QRCode      string
	ExpiresAt   time.Time
}

type ClientConfig struct {
	SessionName string
	Device      *store.Device
	Container   *sqlstore.Container
	Logger      *logger.Logger
	ProxyConfig *session.ProxyConfig
}

type Client struct {
	sessionName string
	client      *whatsmeow.Client
	device      *store.Device
	logger      *logger.Logger

	mu           sync.RWMutex
	state        ConnectionState
	lastActivity time.Time
	errorMessage string

	qrGenerator *QRGenerator

	eventHandlers []func(interface{})

	ctx    context.Context
	cancel context.CancelFunc

	proxyConfig *session.ProxyConfig
}

func validateConfig(config ClientConfig) error {
	if config.SessionName == "" {
		return fmt.Errorf("session name cannot be empty")
	}
	if config.Container == nil {
		return fmt.Errorf("container cannot be nil")
	}
	if config.Logger == nil {
		return fmt.Errorf("logger cannot be nil")
	}
	return nil
}

func NewClient(config ClientConfig) (*Client, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	device := config.Device
	if device == nil {
		deviceStore := config.Container.NewDevice()
		if deviceStore == nil {
			return nil, fmt.Errorf("failed to create device store")
		}
		device = deviceStore
	}

	waLogger := NewWhatsmeowLogger(config.Logger)
	whatsmeowClient := whatsmeow.NewClient(device, waLogger)

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		sessionName:   config.SessionName,
		client:        whatsmeowClient,
		device:        device,
		logger:        config.Logger,
		state:         StateDisconnected,
		lastActivity:  time.Now(),
		qrGenerator:   NewQRGenerator(config.Logger),
		eventHandlers: make([]func(interface{}), 0),
		ctx:           ctx,
		cancel:        cancel,
		proxyConfig:   config.ProxyConfig,
	}

	client.setupEventHandlers()

	config.Logger.InfoWithFields("WhatsApp client created", map[string]interface{}{
		"session_name": config.SessionName,
		"has_device":   device.ID != nil,
	})

	return client, nil
}

func NewClientWithDevice(sessionName string, deviceStore *store.Device, container *sqlstore.Container, logger *logger.Logger) (*Client, error) {
	config := ClientConfig{
		SessionName: sessionName,
		Device:      deviceStore,
		Container:   container,
		Logger:      logger,
	}
	return NewClient(config)
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state == StateConnecting || c.state == StateLoggedIn {
		return nil
	}

	c.setState(StateConnecting)
	c.clearError()

	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	go c.performConnection()

	return nil
}

func (c *Client) performConnection() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("Connection panic", map[string]interface{}{
				"session_name": c.sessionName,
				"error":        r,
			})
			c.setError(fmt.Sprintf("connection panic: %v", r))
		}
	}()

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	if c.isDeviceRegistered() {
		c.connectExistingDevice()
	} else {
		c.connectNewDevice()
	}
}

func (c *Client) isDeviceRegistered() bool {
	return c.device.ID != nil
}

func (c *Client) connectExistingDevice() {
	c.logger.InfoWithFields("Connecting existing device", map[string]interface{}{
		"session_name": c.sessionName,
	})

	if err := c.client.Connect(); err != nil {
		c.logger.ErrorWithFields("Failed to connect existing device", map[string]interface{}{
			"session_name": c.sessionName,
			"error":        err.Error(),
		})
		c.setError(fmt.Sprintf("connection failed: %v", err))
		return
	}

	// Wait for authentication to complete
	c.waitForAuthentication()
}

func (c *Client) connectNewDevice() {
	c.logger.InfoWithFields("Connecting new device", map[string]interface{}{
		"session_name": c.sessionName,
	})

	qrChan, err := c.client.GetQRChannel(c.ctx)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get QR channel", map[string]interface{}{
			"session_name": c.sessionName,
			"error":        err.Error(),
		})
		c.setError(fmt.Sprintf("QR channel failed: %v", err))
		return
	}

	if err := c.client.Connect(); err != nil {
		c.logger.ErrorWithFields("Failed to connect new device", map[string]interface{}{
			"session_name": c.sessionName,
			"error":        err.Error(),
		})
		c.setError(fmt.Sprintf("connection failed: %v", err))
		return
	}

	c.qrGenerator.StartQRLoop(c.ctx, qrChan, c.sessionName)
}

func (c *Client) waitForAuthentication() {
	// Wait up to 10 seconds for authentication
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			c.logger.WarnWithFields("Authentication timeout", map[string]interface{}{
				"session_name": c.sessionName,
			})
			return
		case <-ticker.C:
			if c.client.IsConnected() && c.client.IsLoggedIn() {
				c.setState(StateLoggedIn)
				c.logger.InfoWithFields("Authentication successful", map[string]interface{}{
					"session_name": c.sessionName,
				})
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// State management methods
func (c *Client) setState(state ConnectionState) {
	c.state = state
	c.lastActivity = time.Now()

	c.logger.DebugWithFields("State changed", map[string]interface{}{
		"session_name": c.sessionName,
		"new_state":    state.String(),
	})
}

func (c *Client) setError(message string) {
	c.state = StateError
	c.errorMessage = message
	c.lastActivity = time.Now()
}

func (c *Client) clearError() {
	c.errorMessage = ""
}

func (c *Client) setupEventHandlers() {
	c.client.AddEventHandler(c.handleEvent)
}

func (c *Client) handleEvent(evt interface{}) {

	c.mu.Lock()
	c.lastActivity = time.Now()
	c.mu.Unlock()

	switch v := evt.(type) {
	case *events.Connected:
		c.handleConnectedEvent(v)
	case *events.Disconnected:
		c.handleDisconnectedEvent(v)
	case *events.LoggedOut:
		c.handleLoggedOutEvent(v)
	case *events.QR:

	case *events.PairSuccess:
		c.handlePairSuccessEvent(v)
	case *events.PairError:
		c.handlePairErrorEvent(v)
	default:

		c.logger.DebugWithFields("WhatsApp event received", map[string]interface{}{
			"session_name": c.sessionName,
			"event_type":   fmt.Sprintf("%T", evt),
		})
	}

	c.notifyEventHandlers(evt)
}

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

func (c *Client) handleConnectedEvent(_ *events.Connected) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setState(StateConnected)
	c.logger.InfoWithFields("WhatsApp connected", map[string]interface{}{
		"session_name": c.sessionName,
	})
}

func (c *Client) handleDisconnectedEvent(_ *events.Disconnected) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setState(StateDisconnected)
	c.logger.WarnWithFields("WhatsApp disconnected", map[string]interface{}{
		"session_name": c.sessionName,
	})
}

func (c *Client) handleLoggedOutEvent(evt *events.LoggedOut) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setState(StateDisconnected)
	c.logger.WarnWithFields("WhatsApp logged out", map[string]interface{}{
		"session_name": c.sessionName,
		"reason":       evt.Reason,
	})
}

func (c *Client) handlePairSuccessEvent(evt *events.PairSuccess) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setState(StateLoggedIn)
	c.logger.InfoWithFields("WhatsApp pairing successful", map[string]interface{}{
		"session_name": c.sessionName,
		"device_jid":   evt.ID.String(),
	})
}

func (c *Client) handlePairErrorEvent(evt *events.PairError) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setError(fmt.Sprintf("pairing failed: %v", evt.Error))
	c.logger.ErrorWithFields("WhatsApp pairing failed", map[string]interface{}{
		"session_name": c.sessionName,
		"error":        evt.Error.Error(),
	})
}

// Disconnect gracefully disconnects the client
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.InfoWithFields("Disconnecting client", map[string]interface{}{
		"session_name": c.sessionName,
	})

	// Stop QR process
	c.qrGenerator.Stop()

	// Disconnect WhatsApp client
	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	// Cancel context
	if c.cancel != nil {
		c.cancel()
	}

	c.setState(StateDisconnected)
	return nil
}

func (c *Client) Logout() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Logging out WhatsApp client", map[string]interface{}{
		"session_name": c.sessionName,
	})

	if err := c.client.Logout(context.Background()); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	// State is managed through setState method now

	return nil
}

// Public getters
func (c *Client) GetState() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == StateConnected || c.state == StateLoggedIn
}

func (c *Client) IsLoggedIn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state == StateLoggedIn
}

func (c *Client) GetQRCode() (string, error) {
	if c.IsLoggedIn() {
		return "", fmt.Errorf("client is already logged in")
	}

	if !c.IsConnected() {
		return "", fmt.Errorf("client is not connected")
	}

	// Use QRGenerator to get current QR code
	if qrCode, valid := c.qrGenerator.GetQRCode(); valid {
		return qrCode, nil
	}

	return "", fmt.Errorf("no QR code available")
}

func (c *Client) SetProxy(proxy *session.ProxyConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.proxyConfig = proxy

	if c.IsConnected() {
		if err := c.configureProxy(); err != nil {
			return fmt.Errorf("failed to apply proxy configuration: %w", err)
		}
	}

	return nil
}

func (c *Client) SetEventHandler(handler func(interface{})) {
	c.AddEventHandler(handler)
}

func (c *Client) GetJID() types.JID {
	if c.client.Store.ID == nil {
		return types.EmptyJID
	}
	return *c.client.Store.ID
}

func (c *Client) GetClient() *whatsmeow.Client {
	return c.client
}

func (c *Client) configureProxy() error {
	if c.proxyConfig == nil {
		return nil
	}

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

	c.client.SetProxyAddress(proxyURL.String())

	c.logger.InfoWithFields("Proxy configured", map[string]interface{}{
		"session_name": c.sessionName,
		"proxy_type":   c.proxyConfig.Type,
		"proxy_host":   c.proxyConfig.Host,
		"proxy_port":   c.proxyConfig.Port,
	})

	return nil
}

func (c *Client) GetStatus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == StateError {
		return fmt.Sprintf("error: %s", c.errorMessage)
	}
	return c.state.String()
}

func (c *Client) GetSessionName() string {
	return c.sessionName
}

func (c *Client) AddEventHandler(handler func(interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers = append(c.eventHandlers, handler)
}
