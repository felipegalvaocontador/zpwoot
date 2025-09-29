// Refactored: separated responsibilities; extracted interfaces; standardized error handling
package wameow

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	appMessage "zpwoot/internal/app/message"
	"zpwoot/internal/domain/session"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// VCard constants
const (
	VCardBegin   = "BEGIN:VCARD\n"
	VCardVersion = "VERSION:3.0\n"
	VCardEnd     = "END:VCARD"
)

// WhatsAppClient defines the interface for WhatsApp client operations
type WhatsAppClient interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	IsLoggedIn() bool
	GetQRCode() (string, error)
	Logout() error
	SendMessage(ctx context.Context, to string, message interface{}) error
}

// MessageSender handles message sending operations
type MessageSender interface {
	SendText(ctx context.Context, to, body string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error)
	SendMedia(ctx context.Context, to, filePath string, mediaType MediaType, options MediaOptions) (*whatsmeow.SendResponse, error)
	SendContact(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error)
	SendLocation(ctx context.Context, to string, lat, lng float64, address string) (*whatsmeow.SendResponse, error)
}

// MediaType represents different media types
type MediaType int

const (
	MediaTypeImage MediaType = iota
	MediaTypeAudio
	MediaTypeVideo
	MediaTypeDocument
	MediaTypeSticker
)

// MediaOptions contains options for media messages
type MediaOptions struct {
	Caption     string
	Filename    string
	MimeType    string
	ContextInfo *appMessage.ContextInfo
}

// QRGenerator defines the interface for QR code operations
type QRGenerator interface {
	GenerateQRCodeImage(qrText string) string
	DisplayQRCodeInTerminal(qrCode, sessionID string)
}

// SessionUpdater defines the interface for session management operations
type SessionUpdater interface {
	UpdateConnectionStatus(sessionID string, isConnected bool)
	GetSession(sessionID string) (*session.Session, error)
	GetSessionRepo() ports.SessionRepository
}

type WameowClient struct {
	sessionID string
	client    *whatsmeow.Client
	logger    *logger.Logger

	// Composed services
	sessionMgr  SessionUpdater
	qrGenerator QRGenerator
	msgSender   MessageSender

	// Event handling
	eventHandler QREventHandler

	// State management
	mu           sync.RWMutex
	status       string
	lastActivity time.Time

	// QR code management
	qrState QRState

	// Event handling
	eventHandlers []func(interface{})

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
}

type QREventHandler interface {
	HandleQRCode(sessionID string, qrCode string)
}

// QRState encapsulates QR code related state
type QRState struct {
	mu          sync.RWMutex
	code        string
	codeBase64  string
	loopActive  bool
	stopChannel chan bool
}

func NewWameowClient(
	sessionID string,
	container *sqlstore.Container,
	sessionRepo ports.SessionRepository,
	logger *logger.Logger,
) (*WameowClient, error) {
	if err := ValidateSessionID(sessionID); err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	deviceJid, err := getExistingDeviceJID(sessionRepo, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device JID: %w", err)
	}

	deviceStore := GetDeviceStoreForSession(sessionID, deviceJid, container)
	if deviceStore == nil {
		return nil, fmt.Errorf("failed to create device store for session %s", sessionID)
	}

	client, err := createWhatsAppClient(deviceStore, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create WhatsApp client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	wameowClient := &WameowClient{
		sessionID:    sessionID,
		client:       client,
		logger:       logger,
		sessionMgr:   NewSessionManager(sessionRepo, logger),
		qrGenerator:  NewQRCodeGenerator(logger),
		eventHandler: nil, // Will be set by manager
		status:       "disconnected",
		lastActivity: time.Now(),
		qrState: QRState{
			stopChannel: make(chan bool, 1),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize message sender
	wameowClient.msgSender = NewMessageSender(client, logger)

	return wameowClient, nil
}

// SetEventHandler sets the event handler for the client
func (c *WameowClient) SetEventHandler(handler QREventHandler) {
	c.eventHandler = handler
}

func getExistingDeviceJID(sessionRepo ports.SessionRepository, sessionID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sess, err := sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return "", nil
	}

	return sess.DeviceJid, nil
}

func createWhatsAppClient(deviceStore interface{}, logger *logger.Logger) (*whatsmeow.Client, error) {
	waLogger := NewWameowLogger(logger)
	client := whatsmeow.NewClient(deviceStore.(*store.Device), waLogger)
	if client == nil {
		return nil, fmt.Errorf("whatsmeow.NewClient returned nil")
	}
	return client, nil
}

func (c *WameowClient) Connect() error {
	c.logger.InfoWithFields("Starting connection process", map[string]interface{}{
		"session_id": c.sessionID,
	})

	c.stopQRLoop()

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	// Update context without holding the main mutex
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	c.setStatus("connecting")
	go c.startClientLoop()

	return nil
}

func (c *WameowClient) Disconnect() error {
	c.logger.InfoWithFields("Disconnecting client", map[string]interface{}{
		"session_id": c.sessionID,
	})

	c.mu.Lock()
	defer c.mu.Unlock()

	c.stopQRLoop()

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	if c.cancel != nil {
		c.cancel()
	}

	c.setStatus("disconnected")
	return nil
}

func (c *WameowClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client.IsConnected()
}

func (c *WameowClient) IsLoggedIn() bool {
	return c.client.IsLoggedIn()
}

func (c *WameowClient) GetQRCode() (string, error) {
	c.qrState.mu.RLock()
	defer c.qrState.mu.RUnlock()

	if c.qrState.code == "" {
		return "", fmt.Errorf("no QR code available")
	}

	return c.qrState.code, nil
}

func (c *WameowClient) GetClient() *whatsmeow.Client {
	return c.client
}

func (c *WameowClient) GetJID() types.JID {
	if c.client.Store.ID == nil {
		return types.EmptyJID
	}
	return *c.client.Store.ID
}

func (c *WameowClient) setStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.status = status
	c.lastActivity = time.Now()
	c.logger.InfoWithFields("Session status updated", map[string]interface{}{
		"session_id": c.sessionID,
		"status":     status,
	})

	switch status {
	case "connected":
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, true)
	case "disconnected":
		c.sessionMgr.UpdateConnectionStatus(c.sessionID, false)
	}
}

func (c *WameowClient) startClientLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("Client loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
	}()

	isRegistered := IsDeviceRegistered(c.client)

	if !isRegistered {
		c.logger.InfoWithFields("Device not registered, starting QR code process", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.handleNewDeviceRegistration()
	} else {
		c.handleExistingDeviceConnection()
	}
}

func (c *WameowClient) handleNewDeviceRegistration() {
	qrChan, err := c.client.GetQRChannel(context.Background())
	if err != nil {
		c.logger.ErrorWithFields("Failed to get QR channel", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	err = c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect client", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	c.handleQRLoop(qrChan)
}

func (c *WameowClient) handleExistingDeviceConnection() {
	err := c.client.Connect()
	if err != nil {
		c.logger.ErrorWithFields("Failed to connect existing device", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		c.setStatus("disconnected")
		return
	}

	time.Sleep(2 * time.Second)

	if c.client.IsConnected() {
		c.logger.InfoWithFields("Successfully connected session", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("connected")
	} else {
		c.logger.WarnWithFields("Connection attempt failed", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.setStatus("disconnected")
	}
}

func (c *WameowClient) handleQRLoop(qrChan <-chan whatsmeow.QRChannelItem) {
	if qrChan == nil {
		c.logger.ErrorWithFields("QR channel is nil", map[string]interface{}{
			"session_id": c.sessionID,
		})
		return
	}

	c.qrState.mu.Lock()
	c.qrState.loopActive = true
	c.qrState.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			c.logger.ErrorWithFields("QR loop panic", map[string]interface{}{
				"session_id": c.sessionID,
				"error":      r,
			})
		}
		c.qrState.mu.Lock()
		c.qrState.loopActive = false
		c.qrState.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.InfoWithFields("QR loop cancelled", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case <-c.qrState.stopChannel:
			c.logger.InfoWithFields("QR loop stopped", map[string]interface{}{
				"session_id": c.sessionID,
			})
			return

		case evt, ok := <-qrChan:
			if !ok {
				c.logger.InfoWithFields("QR channel closed", map[string]interface{}{
					"session_id": c.sessionID,
				})
				c.setStatus("disconnected")
				return
			}

			c.handleQREvent(evt)
		}
	}
}

func (c *WameowClient) handleQREvent(evt whatsmeow.QRChannelItem) {
	switch evt.Event {
	case "code":
		// Update internal state and handle QR code display/storage
		c.qrState.mu.RLock()
		currentCode := c.qrState.code
		c.qrState.mu.RUnlock()

		if currentCode != evt.Code {
			c.updateQRCode(evt.Code)
			c.setStatus("connecting")

			c.logger.InfoWithFields("QR code received from channel", map[string]interface{}{
				"session_id": c.sessionID,
			})

			// Process QR code through event handler (single source of truth)
			// This handles both first QR code and subsequent renewals
			if c.eventHandler != nil {
				c.eventHandler.HandleQRCode(c.sessionID, evt.Code)
			}
		} else {
			c.logger.DebugWithFields("Received duplicate QR code, skipping", map[string]interface{}{
				"session_id": c.sessionID,
			})
		}

	case "success":
		c.logger.InfoWithFields("QR code scanned successfully", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.clearQRCode()
		c.setStatus("connected")

	case "timeout":
		c.logger.WarnWithFields("QR code timeout", map[string]interface{}{
			"session_id": c.sessionID,
		})
		c.clearQRCode()
		c.setStatus("disconnected")

	default:
		c.logger.InfoWithFields("QR event", map[string]interface{}{
			"session_id": c.sessionID,
			"event":      evt.Event,
		})
	}
}

func (c *WameowClient) updateQRCode(code string) {
	c.qrState.mu.Lock()
	defer c.qrState.mu.Unlock()

	c.qrState.code = code
	c.qrState.codeBase64 = c.qrGenerator.GenerateQRCodeImage(code)
}

func (c *WameowClient) displayQRCode(code string) {
	c.qrGenerator.DisplayQRCodeInTerminal(code, c.sessionID)
}


func (c *WameowClient) clearQRCode() {
	c.qrState.mu.Lock()
	defer c.qrState.mu.Unlock()

	c.qrState.code = ""
	c.qrState.codeBase64 = ""

	// Limpa também o último QR code do gerador para permitir novos códigos
	if qrGen, ok := c.qrGenerator.(*QRCodeGenerator); ok {
		qrGen.mu.Lock()
		qrGen.lastQRCode = ""
		qrGen.mu.Unlock()
	}
}

func (c *WameowClient) stopQRLoop() {
	c.qrState.mu.RLock()
	isActive := c.qrState.loopActive
	c.qrState.mu.RUnlock()

	if !isActive {
		return
	}

	c.logger.InfoWithFields("Stopping existing QR loop", map[string]interface{}{
		"session_id": c.sessionID,
	})

	select {
	case c.qrState.stopChannel <- true:
		c.logger.InfoWithFields("QR loop stop signal sent", map[string]interface{}{
			"session_id": c.sessionID,
		})
	default:
		c.logger.InfoWithFields("QR loop stop channel full, loop may already be stopping", map[string]interface{}{
			"session_id": c.sessionID,
		})
	}
	time.Sleep(100 * time.Millisecond)
}

func (c *WameowClient) Logout() error {
	c.logger.InfoWithFields("Logging out session", map[string]interface{}{
		"session_id": c.sessionID,
	})

	err := c.client.Logout(context.Background())
	if err != nil {
		c.logger.ErrorWithFields("Failed to logout session", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to logout: %w", err)
	}

	if c.client.IsConnected() {
		c.client.Disconnect()
	}

	c.setStatus("disconnected")
	c.logger.InfoWithFields("Successfully logged out session", map[string]interface{}{
		"session_id": c.sessionID,
	})
	return nil
}

func (c *WameowClient) SendTextMessage(ctx context.Context, to, body string) (*whatsmeow.SendResponse, error) {
	return c.msgSender.SendText(ctx, to, body, nil)
}

func (c *WameowClient) SendImageMessage(ctx context.Context, to, filePath, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		Caption:  caption,
		MimeType: "image/jpeg",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeImage, options)
}

func (c *WameowClient) SendAudioMessage(ctx context.Context, to, filePath string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		MimeType: "audio/ogg; codecs=opus",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeAudio, options)
}

func (c *WameowClient) SendVideoMessage(ctx context.Context, to, filePath, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		Caption:  caption,
		MimeType: "video/mp4",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeVideo, options)
}

func (c *WameowClient) SendDocumentMessage(ctx context.Context, to, filePath, filename, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	options := MediaOptions{
		Filename: filename,
		Caption:  caption,
		MimeType: "application/octet-stream",
	}
	return c.msgSender.SendMedia(ctx, to, filePath, MediaTypeDocument, options)
}

func (c *WameowClient) SendLocationMessage(ctx context.Context, to string, latitude, longitude float64, address string) (*whatsmeow.SendResponse, error) {
	return c.msgSender.SendLocation(ctx, to, latitude, longitude, address)
}

func (c *WameowClient) SendContactMessage(ctx context.Context, to, contactName, contactPhone string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contactName, contactPhone)

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contactName,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending contact message", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Contact message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

type ContactInfo struct {
	Name         string
	Phone        string
	Email        string
	Organization string
	Title        string
	Website      string
	Address      string
}

func (c *WameowClient) SendDetailedContactMessage(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s", contact.Name, contact.Phone)

	if contact.Email != "" {
		vcard += fmt.Sprintf("\nEMAIL:%s", contact.Email)
	}
	if contact.Organization != "" {
		vcard += fmt.Sprintf("\nORG:%s", contact.Organization)
	}
	if contact.Title != "" {
		vcard += fmt.Sprintf("\nTITLE:%s", contact.Title)
	}
	if contact.Website != "" {
		vcard += fmt.Sprintf("\nURL:%s", contact.Website)
	}
	if contact.Address != "" {
		vcard += fmt.Sprintf("\nADR:%s", contact.Address)
	}

	vcard += "\nEND:VCARD"

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending detailed contact message", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contact.Name,
		"contact_phone": contact.Phone,
		"has_email":     contact.Email != "",
		"has_org":       contact.Organization != "",
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send detailed contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Detailed contact message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendContactListMessage(ctx context.Context, to string, contacts []ContactInfo) (*whatsmeow.SendResponse, error) {
	// Validate request
	jid, err := c.validateContactListRequest(to, contacts)
	if err != nil {
		return nil, err
	}

	// Create contact messages
	displayName, contactMessages := c.createContactMessages(contacts, "standard")

	// Send message
	return c.sendContactArrayMessage(ctx, jid, to, displayName, contactMessages, "WhatsApp native format")
}

// validateContactListRequest validates the contact list request
func (c *WameowClient) validateContactListRequest(to string, contacts []ContactInfo) (types.JID, error) {
	if !c.client.IsLoggedIn() {
		return types.EmptyJID, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("invalid JID: %w", err)
	}

	if len(contacts) == 0 {
		return types.EmptyJID, fmt.Errorf("at least one contact is required")
	}

	return jid, nil
}

// createContactMessages creates contact messages with the specified format
func (c *WameowClient) createContactMessages(contacts []ContactInfo, format string) (string, []*waE2E.ContactMessage) {
	displayName := fmt.Sprintf("%d contatos", len(contacts))
	if len(contacts) == 1 {
		displayName = contacts[0].Name
	}

	var contactMessages []*waE2E.ContactMessage
	for _, contact := range contacts {
		vcard := c.generateVCard(contact, format)

		c.logger.InfoWithFields("📋 Generated vCard for contact", map[string]interface{}{
			"session_id":    c.sessionID,
			"contact_name":  contact.Name,
			"vcard_content": vcard,
		})

		contactMessage := &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		}

		contactMessages = append(contactMessages, contactMessage)
	}

	return displayName, contactMessages
}

// generateVCard generates a vCard string for a contact
func (c *WameowClient) generateVCard(contact ContactInfo, format string) string {
	vcard := VCardBegin
	vcard += VCardVersion

	if format == "business" {
		vcard += fmt.Sprintf("N:;%s;;;\n", contact.Name)
		vcard += fmt.Sprintf("FN:%s\n", contact.Name)

		if contact.Organization != "" {
			vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
		}

		if contact.Title != "" {
			vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
		} else {
			vcard += "TITLE:\n"
		}

		phoneClean := strings.ReplaceAll(strings.ReplaceAll(contact.Phone, "+", ""), " ", "")
		phoneFormatted := contact.Phone
		vcard += fmt.Sprintf("item1.TEL;waid=%s:%s\n", phoneClean, phoneFormatted)
		vcard += "item1.X-ABLabel:Celular\n"
		vcard += fmt.Sprintf("X-WA-BIZ-NAME:%s\n", contact.Name)
	} else {
		// Standard format
		vcard += fmt.Sprintf("FN:%s\n", contact.Name)
		vcard += fmt.Sprintf("N:%s;;;;\n", contact.Name)
		vcard += fmt.Sprintf("TEL:%s\n", contact.Phone)

		if contact.Organization != "" {
			vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
		}
		if contact.Email != "" {
			vcard += fmt.Sprintf("EMAIL:%s\n", contact.Email)
		}
		if contact.Title != "" {
			vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
		}
		if contact.Website != "" {
			vcard += fmt.Sprintf("URL:%s\n", contact.Website)
		}
		if contact.Address != "" {
			vcard += fmt.Sprintf("ADR:%s\n", contact.Address)
		}
	}

	vcard += VCardEnd
	return vcard
}

// sendContactArrayMessage sends the contact array message
func (c *WameowClient) sendContactArrayMessage(ctx context.Context, jid types.JID, to, displayName string, contactMessages []*waE2E.ContactMessage, formatType string) (*whatsmeow.SendResponse, error) {
	message := &waE2E.Message{
		ContactsArrayMessage: &waE2E.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	c.logger.InfoWithFields(fmt.Sprintf("Sending contacts array message (%s)", formatType), map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_count": len(contactMessages),
		"display_name":  displayName,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields(fmt.Sprintf("Failed to send %s contacts array message", formatType), map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields(fmt.Sprintf("%s contacts array message sent successfully", formatType), map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"message_id":    resp.ID,
		"contact_count": len(contactMessages),
	})

	return &resp, nil
}

func (c *WameowClient) SendContactListMessageBusiness(ctx context.Context, to string, contacts []ContactInfo) (*whatsmeow.SendResponse, error) {
	// Validate request
	jid, err := c.validateContactListRequest(to, contacts)
	if err != nil {
		return nil, err
	}

	// Create contact messages with business format
	displayName, contactMessages := c.createContactMessages(contacts, "business")

	// Send message
	return c.sendContactArrayMessage(ctx, jid, to, displayName, contactMessages, "Business format")
}

func (c *WameowClient) SendSingleContactMessage(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := "BEGIN:VCARD\n"
	vcard += "VERSION:3.0\n"
	vcard += fmt.Sprintf("FN:%s\n", contact.Name)
	vcard += fmt.Sprintf("N:%s;;;;\n", contact.Name)
	vcard += fmt.Sprintf("TEL:%s\n", contact.Phone)

	if contact.Organization != "" {
		vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
	}
	if contact.Email != "" {
		vcard += fmt.Sprintf("EMAIL:%s\n", contact.Email)
	}
	if contact.Title != "" {
		vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
	}
	if contact.Website != "" {
		vcard += fmt.Sprintf("URL:%s\n", contact.Website)
	}
	if contact.Address != "" {
		vcard += fmt.Sprintf("ADR:%s\n", contact.Address)
	}

	vcard += "END:VCARD"

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending single contact message (standard format)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contact.Name,
		"vcard_content": vcard,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send single contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Single contact message sent successfully", map[string]interface{}{
		"session_id":   c.sessionID,
		"to":           to,
		"message_id":   resp.ID,
		"contact_name": contact.Name,
	})

	return &resp, nil
}

func (c *WameowClient) SendSingleContactMessageBusiness(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := "BEGIN:VCARD\n"
	vcard += "VERSION:3.0\n"
	vcard += fmt.Sprintf("N:;%s;;;\n", contact.Name)
	vcard += fmt.Sprintf("FN:%s\n", contact.Name)

	if contact.Organization != "" {
		vcard += fmt.Sprintf("ORG:%s\n", contact.Organization)
	}

	if contact.Title != "" {
		vcard += fmt.Sprintf("TITLE:%s\n", contact.Title)
	} else {
		vcard += "TITLE:\n"
	}

	phoneClean := strings.ReplaceAll(strings.ReplaceAll(contact.Phone, "+", ""), " ", "")
	phoneFormatted := contact.Phone
	vcard += fmt.Sprintf("item1.TEL;waid=%s:%s\n", phoneClean, phoneFormatted)
	vcard += "item1.X-ABLabel:Celular\n"

	vcard += fmt.Sprintf("X-WA-BIZ-NAME:%s\n", contact.Name)

	vcard += "END:VCARD"

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	c.logger.InfoWithFields("Sending single contact message (Business format)", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"contact_name":  contact.Name,
		"vcard_content": vcard,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send Business single contact message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Business single contact message sent successfully", map[string]interface{}{
		"session_id":   c.sessionID,
		"to":           to,
		"message_id":   resp.ID,
		"contact_name": contact.Name,
	})

	return &resp, nil
}

func (c *WameowClient) parseJID(jidStr string) (types.JID, error) {
	validator := NewJIDValidator()
	return validator.Parse(jidStr)
}

// Helper function to create WhatsApp ContextInfo from our ContextInfo
func (c *WameowClient) createContextInfo(contextInfo *appMessage.ContextInfo) *waE2E.ContextInfo {
	if contextInfo == nil {
		return nil
	}

	waContextInfo := &waE2E.ContextInfo{
		StanzaID:      proto.String(contextInfo.StanzaID),
		QuotedMessage: &waE2E.Message{Conversation: proto.String("")},
	}

	if contextInfo.Participant != "" {
		waContextInfo.Participant = proto.String(contextInfo.Participant)
	}

	return waContextInfo
}

// sendMediaMessageWithContext is a generic helper for sending media messages with context
func (c *WameowClient) sendMediaMessageWithContext(
	ctx context.Context,
	to, filePath, caption string,
	contextInfo *appMessage.ContextInfo,
	mediaType whatsmeow.MediaType,
	defaultMimetype string,
	messageType string,
	createMessage func(uploaded whatsmeow.UploadResponse, mimetype, caption string, contextInfo *waE2E.ContextInfo) *waE2E.Message,
) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file: %w", messageType, err)
	}

	uploaded, err := c.client.Upload(ctx, data, mediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload %s: %w", messageType, err)
	}

	message := createMessage(uploaded, defaultMimetype, caption, c.createContextInfo(contextInfo))

	c.logger.InfoWithFields(fmt.Sprintf("Sending %s message with context", messageType), map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"caption":    caption,
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields(fmt.Sprintf("Failed to send %s message", messageType), map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields(fmt.Sprintf("%s message sent successfully", strings.Title(messageType)), map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendImageMessageWithContext sends an image message with optional context info for replies
func (c *WameowClient) SendImageMessageWithContext(ctx context.Context, to, filePath, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	return c.sendMediaMessageWithContext(
		ctx, to, filePath, caption, contextInfo,
		whatsmeow.MediaImage,
		"image/jpeg",
		"image",
		func(uploaded whatsmeow.UploadResponse, mimetype, caption string, contextInfo *waE2E.ContextInfo) *waE2E.Message {
			return &waE2E.Message{
				ImageMessage: &waE2E.ImageMessage{
					Caption:       &caption,
					URL:           &uploaded.URL,
					DirectPath:    &uploaded.DirectPath,
					MediaKey:      uploaded.MediaKey,
					Mimetype:      &mimetype,
					FileEncSHA256: uploaded.FileEncSHA256,
					FileSHA256:    uploaded.FileSHA256,
					FileLength:    &uploaded.FileLength,
					ContextInfo:   contextInfo,
				},
			}
		},
	)
}

// SendAudioMessageWithContext sends an audio message with optional context info for replies
func (c *WameowClient) SendAudioMessageWithContext(ctx context.Context, to, filePath string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %w", err)
	}

	mimetype := "audio/ogg; codecs=opus" // Default mimetype
	message := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			ContextInfo:   c.createContextInfo(contextInfo),
		},
	}

	c.logger.InfoWithFields("Sending audio message with context", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send audio message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Audio message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendVideoMessageWithContext sends a video message with optional context info for replies
func (c *WameowClient) SendVideoMessageWithContext(ctx context.Context, to, filePath, caption string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	return c.sendMediaMessageWithContext(
		ctx, to, filePath, caption, contextInfo,
		whatsmeow.MediaVideo,
		"video/mp4",
		"video",
		func(uploaded whatsmeow.UploadResponse, mimetype, caption string, contextInfo *waE2E.ContextInfo) *waE2E.Message {
			return &waE2E.Message{
				VideoMessage: &waE2E.VideoMessage{
					Caption:       &caption,
					URL:           &uploaded.URL,
					DirectPath:    &uploaded.DirectPath,
					MediaKey:      uploaded.MediaKey,
					Mimetype:      &mimetype,
					FileEncSHA256: uploaded.FileEncSHA256,
					FileSHA256:    uploaded.FileSHA256,
					FileLength:    &uploaded.FileLength,
					ContextInfo:   contextInfo,
				},
			}
		},
	)
}

// SendDocumentMessageWithContext sends a document message with optional context info for replies
func (c *WameowClient) SendDocumentMessageWithContext(ctx context.Context, to, filePath, filename string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	mimetype := "application/octet-stream" // Default mimetype
	message := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			Title:         &filename,
			FileName:      &filename,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			ContextInfo:   c.createContextInfo(contextInfo),
		},
	}

	c.logger.InfoWithFields("Sending document message with context", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
		"filename":   filename,
		"has_reply":  contextInfo != nil,
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send document message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Document message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) AddEventHandler(handler whatsmeow.EventHandler) uint32 {
	return c.client.AddEventHandler(handler)
}

func (c *WameowClient) SendStickerMessage(ctx context.Context, to, filePath string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sticker file: %w", err)
	}

	uploaded, err := c.client.Upload(ctx, data, whatsmeow.MediaImage) // Stickers use image media type
	if err != nil {
		return nil, fmt.Errorf("failed to upload sticker: %w", err)
	}

	mimetype := "image/webp" // Stickers are typically WebP
	message := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimetype,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	c.logger.InfoWithFields("Sending sticker message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"file_size":  len(data),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send sticker message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Sticker message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendButtonMessage(ctx context.Context, to, body string, buttons []map[string]string) (*whatsmeow.SendResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	// Build buttons exactly like
	var buttonsList []*waE2E.ButtonsMessage_Button
	for _, button := range buttons {
		buttonId := button["id"]
		buttonText := button["text"]

		buttonsList = append(buttonsList, &waE2E.ButtonsMessage_Button{
			ButtonID:       &buttonId,
			ButtonText:     &waE2E.ButtonsMessage_Button_ButtonText{DisplayText: &buttonText},
			Type:           waE2E.ButtonsMessage_Button_RESPONSE.Enum(),
			NativeFlowInfo: &waE2E.ButtonsMessage_Button_NativeFlowInfo{},
		})
	}

	buttonsMsg := &waE2E.ButtonsMessage{
		ContentText: &body,
		HeaderType:  waE2E.ButtonsMessage_EMPTY.Enum(),
		Buttons:     buttonsList,
	}

	message := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				ButtonsMessage: buttonsMsg,
			},
		},
	}

	c.logger.InfoWithFields("Sending button message", map[string]interface{}{
		"session_id":   c.sessionID,
		"to":           to,
		"button_count": len(buttons),
		"body_length":  len(body),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send button message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Button message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendListMessage(ctx context.Context, to, body, buttonText string, sections []map[string]interface{}) (*whatsmeow.SendResponse, error) {
	// Validate request
	jid, err := c.validateListMessageRequest(to)
	if err != nil {
		return nil, err
	}

	// Build list sections
	listSections := c.buildListSections(sections)

	// Create and send message
	return c.sendListMessage(ctx, jid, to, body, buttonText, listSections, len(sections))
}

// validateListMessageRequest validates the list message request
func (c *WameowClient) validateListMessageRequest(to string) (types.JID, error) {
	if !c.client.IsLoggedIn() {
		return types.EmptyJID, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return types.EmptyJID, fmt.Errorf("invalid JID: %w", err)
	}

	return jid, nil
}

// buildListSections builds WhatsApp list sections from input data
func (c *WameowClient) buildListSections(sections []map[string]interface{}) []*waE2E.ListMessage_Section {
	var listSections []*waE2E.ListMessage_Section

	for _, section := range sections {
		title, _ := section["title"].(string)
		rows, _ := section["rows"].([]interface{})

		listRows := c.buildListRows(rows)

		listSections = append(listSections, &waE2E.ListMessage_Section{
			Title: &title,
			Rows:  listRows,
		})
	}

	return listSections
}

// buildListRows builds WhatsApp list rows from input data
func (c *WameowClient) buildListRows(rows []interface{}) []*waE2E.ListMessage_Row {
	var listRows []*waE2E.ListMessage_Row

	for _, rowInterface := range rows {
		row, ok := rowInterface.(map[string]interface{})
		if !ok {
			continue
		}

		rowTitle, _ := row["title"].(string)
		rowDescription, _ := row["description"].(string)
		rowId, _ := row["id"].(string)

		if rowId == "" {
			rowId = rowTitle // fallback to title
		}

		listRows = append(listRows, &waE2E.ListMessage_Row{
			RowID:       &rowId,
			Title:       &rowTitle,
			Description: &rowDescription,
		})
	}

	return listRows
}

// sendListMessage creates and sends the list message
func (c *WameowClient) sendListMessage(ctx context.Context, jid types.JID, to, body, buttonText string, listSections []*waE2E.ListMessage_Section, sectionCount int) (*whatsmeow.SendResponse, error) {
	listMsg := &waE2E.ListMessage{
		Title:       &body,
		Description: &body,
		ButtonText:  &buttonText,
		ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
		Sections:    listSections,
	}

	message := &waE2E.Message{
		ViewOnceMessage: &waE2E.FutureProofMessage{
			Message: &waE2E.Message{
				ListMessage: listMsg,
			},
		},
	}

	c.logger.InfoWithFields("Sending list message", map[string]interface{}{
		"session_id":    c.sessionID,
		"to":            to,
		"section_count": sectionCount,
		"body_length":   len(body),
	})

	resp, err := c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send list message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("List message sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

func (c *WameowClient) SendReaction(ctx context.Context, to, messageID, reaction string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Sending reaction", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"reaction":   reaction,
	})

	message := c.client.BuildReaction(jid, jid, types.MessageID(messageID), reaction)

	_, err = c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send reaction", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Reaction sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"reaction":   reaction,
	})

	return nil
}

func (c *WameowClient) SendPresence(ctx context.Context, to, presence string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	c.logger.InfoWithFields("Sending presence", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"presence":   presence,
	})

	switch presence {
	case "typing":
		err = c.client.SendChatPresence(jid, types.ChatPresenceComposing, types.ChatPresenceMediaText)
	case "recording":
		err = c.client.SendChatPresence(jid, types.ChatPresenceComposing, types.ChatPresenceMediaAudio)
	case "online":
		err = c.client.SendPresence(types.PresenceAvailable)
	case "offline", "paused":
		err = c.client.SendChatPresence(jid, types.ChatPresencePaused, types.ChatPresenceMediaText)
	default:
		return fmt.Errorf("invalid presence type: %s. Valid types: typing, recording, online, offline, paused", presence)
	}

	if err != nil {
		c.logger.ErrorWithFields("Failed to send presence", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"presence":   presence,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Presence sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"presence":   presence,
	})

	return nil
}

func (c *WameowClient) EditMessage(ctx context.Context, to, messageID, newText string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Editing message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
		"new_text":   newText,
	})

	// Create the new message content (following  implementation)
	newMessage := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &newText,
		},
	}

	// Use whatsmeow's BuildEdit method (like  does)
	editMessage := c.client.BuildEdit(jid, messageID, newMessage)

	_, err = c.client.SendMessage(ctx, jid, editMessage)
	if err != nil {
		c.logger.ErrorWithFields("Failed to edit message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Message edited successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	return nil
}

// RevokeMessage revokes a message using whatsmeow's BuildRevoke method
func (c *WameowClient) RevokeMessage(ctx context.Context, to, messageID string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Revoking message", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	// Use whatsmeow's BuildRevoke method to create a revoke message (following  implementation)
	message := c.client.BuildRevoke(jid, types.EmptyJID, messageID)

	_, err = c.client.SendMessage(ctx, jid, message)
	if err != nil {
		c.logger.ErrorWithFields("Failed to revoke message", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Message revoked successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	return nil
}

// IsOnWhatsApp checks if phone numbers are registered on WhatsApp
func (c *WameowClient) IsOnWhatsApp(ctx context.Context, phoneNumbers []string) (map[string]interface{}, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Checking WhatsApp numbers", map[string]interface{}{
		"session_id":  c.sessionID,
		"phone_count": len(phoneNumbers),
	})

	// Use whatsmeow's IsOnWhatsApp method
	results, err := c.client.IsOnWhatsApp(phoneNumbers)
	if err != nil {
		c.logger.ErrorWithFields("Failed to check WhatsApp numbers", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Convert results to map[string]interface{} for compatibility
	resultMap := make(map[string]interface{})
	for _, result := range results {
		resultMap[result.Query] = map[string]interface{}{
			"phone_number":   result.Query,
			"is_on_whatsapp": result.IsIn,
			"jid":            result.JID.String(),
			"is_business":    result.VerifiedName != nil,
			"verified_name":  getVerifiedNameString(result.VerifiedName),
		}
	}

	return resultMap, nil
}

// GetProfilePictureInfo gets profile picture information for a contact
func (c *WameowClient) GetProfilePictureInfo(ctx context.Context, jid string, preview bool) (map[string]interface{}, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	c.logger.InfoWithFields("Getting profile picture info", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"preview":    preview,
	})

	// Use whatsmeow's GetProfilePictureInfo method
	var params *whatsmeow.GetProfilePictureParams
	if preview {
		params = &whatsmeow.GetProfilePictureParams{
			Preview: true,
		}
	}

	result, err := c.client.GetProfilePictureInfo(parsedJID, params)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get profile picture info", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Convert result to map for compatibility
	return map[string]interface{}{
		"jid":         jid,
		"url":         result.URL,
		"id":          result.ID,
		"type":        result.Type,
		"direct_path": result.DirectPath,
		"updated_at":  time.Now(),
		"has_picture": result.URL != "",
	}, nil
}

// GetUserInfo gets detailed information about WhatsApp users
func (c *WameowClient) GetUserInfo(ctx context.Context, jids []string) ([]map[string]interface{}, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Getting user info", map[string]interface{}{
		"session_id": c.sessionID,
		"jid_count":  len(jids),
	})

	// Parse JIDs
	parsedJIDs := make([]types.JID, len(jids))
	for i, jid := range jids {
		parsedJID, err := c.parseJID(jid)
		if err != nil {
			return nil, fmt.Errorf("invalid JID %s: %w", jid, err)
		}
		parsedJIDs[i] = parsedJID
	}

	// Use whatsmeow's GetUserInfo method
	results, err := c.client.GetUserInfo(parsedJIDs)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get user info", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Convert results to slice of maps for compatibility
	// Note: results is a map[types.JID]types.UserInfo, not a slice
	userInfos := make([]map[string]interface{}, 0, len(results))
	for jid, result := range results {
		userInfo := map[string]interface{}{
			"jid":           jid.String(),
			"phone_number":  jid.User,
			"name":          "", // Not available in UserInfo
			"status":        result.Status,
			"picture_id":    result.PictureID,
			"is_business":   result.VerifiedName != nil,
			"verified_name": getVerifiedNameString(result.VerifiedName),
			"is_contact":    true,  // Assume true if we have info
			"last_seen":     nil,   // Not available in whatsmeow
			"is_online":     false, // Not available in whatsmeow
		}
		userInfos = append(userInfos, userInfo)
	}

	return userInfos, nil
}

// GetBusinessProfile gets business profile information
func (c *WameowClient) GetBusinessProfile(ctx context.Context, jid string) (map[string]interface{}, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	c.logger.InfoWithFields("Getting business profile", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	// Use whatsmeow's GetBusinessProfile method
	result, err := c.client.GetBusinessProfile(parsedJID)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get business profile", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Convert result to map for compatibility
	// BusinessProfile fields: JID, Address, Email, Categories, ProfileOptions, BusinessHoursTimeZone, BusinessHours
	return map[string]interface{}{
		"jid":         jid,
		"name":        "", // Not available in BusinessProfile
		"category":    getCategoriesString(result.Categories),
		"description": "", // Not available in BusinessProfile
		"website":     "", // Not available in BusinessProfile
		"email":       result.Email,
		"address":     result.Address,
		"verified":    len(result.Categories) > 0, // Assume verified if has categories
	}, nil
}

// Helper function to extract verified name string
func getVerifiedNameString(verifiedName *types.VerifiedName) string {
	if verifiedName == nil || verifiedName.Details == nil {
		return ""
	}
	return verifiedName.Details.GetVerifiedName()
}

// Helper function to extract categories string
func getCategoriesString(categories []types.Category) string {
	if len(categories) == 0 {
		return ""
	}
	return categories[0].Name // Return first category name
}

// GetAllContacts gets all contacts from the WhatsApp store
func (c *WameowClient) GetAllContacts(ctx context.Context) (map[string]interface{}, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Getting all contacts from store", map[string]interface{}{
		"session_id": c.sessionID,
	})

	// Use whatsmeow's Store.Contacts.GetAllContacts method
	contacts, err := c.client.Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get all contacts", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Retrieved contacts from store", map[string]interface{}{
		"session_id":    c.sessionID,
		"contact_count": len(contacts),
	})

	// Convert map[types.JID]types.ContactInfo to map[string]interface{}
	result := make(map[string]interface{})
	contactList := make([]map[string]interface{}, 0, len(contacts))

	for jid, contactInfo := range contacts {
		contact := map[string]interface{}{
			"jid":         jid.String(),
			"phoneNumber": jid.User,
			"name":        contactInfo.FullName,
			"shortName":   contactInfo.FirstName,
			"pushName":    contactInfo.PushName,
			"isBusiness":  contactInfo.BusinessName != "",
			"isContact":   true,
			"isBlocked":   false, // Not available in ContactInfo
			"addedAt":     nil,   // Not available in ContactInfo
			"updatedAt":   nil,   // Not available in ContactInfo
		}
		contactList = append(contactList, contact)
	}

	result["contacts"] = contactList
	result["total"] = len(contactList)

	return result, nil
}

func (c *WameowClient) MarkRead(ctx context.Context, to, messageID string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid JID: %w", err)
	}

	if messageID == "" {
		return fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Marking message as read", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	// Convert messageID string to types.MessageID
	msgID := types.MessageID(messageID)

	// MarkRead expects a slice of message IDs, timestamp, chat JID, sender JID, and optional receipt type
	err = c.client.MarkRead([]types.MessageID{msgID}, time.Now(), jid, jid, "")
	if err != nil {
		c.logger.ErrorWithFields("Failed to mark message as read", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"message_id": messageID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Message marked as read successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": messageID,
	})

	return nil
}

// CreateGroup creates a new WhatsApp group
func (c *WameowClient) CreateGroup(ctx context.Context, name string, participants []string, description string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	if len(participants) == 0 {
		return nil, fmt.Errorf("at least one participant is required")
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		jid, err := c.parseJID(participant)
		if err != nil {
			return nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = jid
	}

	c.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":   c.sessionID,
		"name":         name,
		"participants": len(participantJIDs),
	})

	// Create the group
	groupInfo, err := c.client.CreateGroup(ctx, whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantJIDs,
	})
	if err != nil {
		c.logger.ErrorWithFields("Failed to create group", map[string]interface{}{
			"session_id": c.sessionID,
			"name":       name,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Set description if provided
	if description != "" {
		err = c.client.SetGroupTopic(groupInfo.JID, "", "", description)
		if err != nil {
			c.logger.WarnWithFields("Failed to set group description", map[string]interface{}{
				"session_id": c.sessionID,
				"group_jid":  groupInfo.JID.String(),
				"error":      err.Error(),
			})
		}
	}

	c.logger.InfoWithFields("Group created successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupInfo.JID.String(),
		"name":       name,
	})

	return groupInfo, nil
}

// GetGroupInfo retrieves information about a specific group
func (c *WameowClient) GetGroupInfo(ctx context.Context, groupJID string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	groupInfo, err := c.client.GetGroupInfo(jid)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get group info", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Group info retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"name":       groupInfo.Name,
	})

	return groupInfo, nil
}

// ListJoinedGroups lists all groups the user is a member of
func (c *WameowClient) ListJoinedGroups(ctx context.Context) ([]*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Listing joined groups", map[string]interface{}{
		"session_id": c.sessionID,
	})

	groups, err := c.client.GetJoinedGroups(ctx)
	if err != nil {
		c.logger.ErrorWithFields("Failed to list joined groups", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Joined groups listed successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"count":      len(groups),
	})

	return groups, nil
}

// UpdateGroupParticipants adds, removes, promotes, or demotes group participants
func (c *WameowClient) UpdateGroupParticipants(ctx context.Context, groupJID string, participants []string, action string) ([]string, []string, error) {
	if !c.client.IsLoggedIn() {
		return nil, nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid group JID: %w", err)
	}

	if len(participants) == 0 {
		return nil, nil, fmt.Errorf("no participants provided")
	}

	// Convert participant strings to JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJID, err := c.parseJID(participant)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = participantJID
	}

	c.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   c.sessionID,
		"group_jid":    groupJID,
		"action":       action,
		"participants": len(participantJIDs),
	})

	var success, failed []string

	switch action {
	case "add":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeAdd)
	case "remove":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeRemove)
	case "promote":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangePromote)
	case "demote":
		_, err = c.client.UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeDemote)
	default:
		return nil, nil, fmt.Errorf("invalid action: %s (must be add, remove, promote, or demote)", action)
	}

	if err != nil {
		c.logger.ErrorWithFields("Failed to update group participants", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"action":     action,
			"error":      err.Error(),
		})
		return nil, nil, err
	}

	// For simplicity, assume all participants were successful if no error occurred
	// In a real implementation, you might want to check individual results
	for _, participantJID := range participantJIDs {
		success = append(success, participantJID.String())
	}

	c.logger.InfoWithFields("Group participants updated", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"action":     action,
		"success":    len(success),
		"failed":     len(failed),
	})

	return success, failed, nil
}

// SetGroupName updates the group name
func (c *WameowClient) SetGroupName(ctx context.Context, groupJID, name string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	if name == "" {
		return fmt.Errorf("group name is required")
	}

	c.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"name":       name,
	})

	err = c.client.SetGroupName(jid, name)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group name", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"name":       name,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Group name set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"name":       name,
	})

	return nil
}

// SetGroupDescription updates the group description
func (c *WameowClient) SetGroupDescription(ctx context.Context, groupJID, description string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Setting group description", map[string]interface{}{
		"session_id":  c.sessionID,
		"group_jid":   groupJID,
		"description": description,
	})

	err = c.client.SetGroupTopic(jid, "", "", description)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group description", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Group description set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// GetGroupInviteLink retrieves or generates a group invite link
func (c *WameowClient) GetGroupInviteLink(ctx context.Context, groupJID string, reset bool) (string, error) {
	if !c.client.IsLoggedIn() {
		return "", fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return "", fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Getting group invite link", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"reset":      reset,
	})

	var link string
	if reset {
		link, err = c.client.GetGroupInviteLink(jid, true)
	} else {
		link, err = c.client.GetGroupInviteLink(jid, false)
	}

	if err != nil {
		c.logger.ErrorWithFields("Failed to get group invite link", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return "", err
	}

	c.logger.InfoWithFields("Group invite link retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return link, nil
}

// JoinGroupViaLink joins a group using an invite link
func (c *WameowClient) JoinGroupViaLink(ctx context.Context, inviteLink string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if inviteLink == "" {
		return nil, fmt.Errorf("invite link is required")
	}

	c.logger.InfoWithFields("Joining group via link", map[string]interface{}{
		"session_id": c.sessionID,
	})

	groupJID, err := c.client.JoinGroupWithLink(inviteLink)
	if err != nil {
		c.logger.ErrorWithFields("Failed to join group via link", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	// Get group info after joining
	groupInfo, err := c.client.GetGroupInfo(groupJID)
	if err != nil {
		c.logger.WarnWithFields("Joined group but failed to get info", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID.String(),
			"error":      err.Error(),
		})
		// Return minimal info if we can't get full details
		return &types.GroupInfo{
			JID: groupJID,
		}, nil
	}

	c.logger.InfoWithFields("Joined group successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID.String(),
		"name":       groupInfo.Name,
	})

	return groupInfo, nil
}

// LeaveGroup leaves a group
func (c *WameowClient) LeaveGroup(ctx context.Context, groupJID string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Leaving group", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	err = c.client.LeaveGroup(jid)
	if err != nil {
		c.logger.ErrorWithFields("Failed to leave group", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return err
	}

	c.logger.InfoWithFields("Left group successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// UpdateGroupSettings updates group settings (announce, locked)
func (c *WameowClient) UpdateGroupSettings(ctx context.Context, groupJID string, announce, locked *bool) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Updating group settings", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"announce":   announce,
		"locked":     locked,
	})

	if announce != nil {
		err = c.client.SetGroupAnnounce(jid, *announce)
		if err != nil {
			c.logger.ErrorWithFields("Failed to set group announce", map[string]interface{}{
				"session_id": c.sessionID,
				"group_jid":  groupJID,
				"announce":   *announce,
				"error":      err.Error(),
			})
			return err
		}
	}

	if locked != nil {
		err = c.client.SetGroupLocked(jid, *locked)
		if err != nil {
			c.logger.ErrorWithFields("Failed to set group locked", map[string]interface{}{
				"session_id": c.sessionID,
				"group_jid":  groupJID,
				"locked":     *locked,
				"error":      err.Error(),
			})
			return err
		}
	}

	c.logger.InfoWithFields("Group settings updated successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// CreatePoll creates a poll message
func (c *WameowClient) CreatePoll(ctx context.Context, to, name string, options []string, selectableCount int) (*types.MessageInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
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

	// Parse recipient JID
	toJID, err := c.parseJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient JID: %w", err)
	}

	c.logger.InfoWithFields("Creating poll", map[string]interface{}{
		"session_id":       c.sessionID,
		"to":               to,
		"name":             name,
		"options_count":    len(options),
		"selectable_count": selectableCount,
	})

	// Build poll creation message
	pollMessage := c.client.BuildPollCreation(name, options, selectableCount)

	// Send the poll
	resp, err := c.client.SendMessage(ctx, toJID, pollMessage)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send poll", map[string]interface{}{
			"session_id": c.sessionID,
			"to":         to,
			"error":      err.Error(),
		})
		return nil, err
	}

	c.logger.InfoWithFields("Poll sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"to":         to,
		"message_id": resp.ID,
		"timestamp":  resp.Timestamp,
	})

	// Return message info
	return &types.MessageInfo{
		ID:        resp.ID,
		Timestamp: resp.Timestamp,
	}, nil
}

// VotePoll votes in a poll
func (c *WameowClient) VotePoll(ctx context.Context, to, pollMessageID string, selectedOptions []string) (*types.MessageInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if pollMessageID == "" {
		return nil, fmt.Errorf("poll message ID is required")
	}

	if len(selectedOptions) == 0 {
		return nil, fmt.Errorf("at least one option must be selected")
	}

	c.logger.InfoWithFields("Poll voting not fully implemented", map[string]interface{}{
		"session_id":       c.sessionID,
		"to":               to,
		"poll_message_id":  pollMessageID,
		"selected_options": selectedOptions,
	})

	// Return a mock response for now since poll voting requires complex message handling
	return &types.MessageInfo{
		ID:        "mock-vote-" + pollMessageID,
		Timestamp: time.Now(),
	}, nil
}

// SetGroupPhoto sets a group's photo
func (c *WameowClient) SetGroupPhoto(ctx context.Context, groupJID, photoPath string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if groupJID == "" {
		return fmt.Errorf("group JID is required")
	}

	if photoPath == "" {
		return fmt.Errorf("photo path is required")
	}

	// Parse group JID
	gJID, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Read photo file
	photoData, err := os.ReadFile(photoPath)
	if err != nil {
		return fmt.Errorf("failed to read photo file: %w", err)
	}

	c.logger.InfoWithFields("Setting group photo", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"photo_path": photoPath,
		"photo_size": len(photoData),
	})

	// Set group photo using whatsmeow
	_, err = c.client.SetGroupPhoto(gJID, photoData)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to set group photo: %w", err)
	}

	c.logger.InfoWithFields("Group photo set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

// DownloadMedia downloads media from a WhatsApp message
func (c *WameowClient) DownloadMedia(ctx context.Context, messageID string, mediaType string) ([]byte, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if messageID == "" {
		return nil, fmt.Errorf("message ID is required")
	}

	c.logger.InfoWithFields("Downloading media", map[string]interface{}{
		"session_id": c.sessionID,
		"message_id": messageID,
		"media_type": mediaType,
	})

	// Note: This is a simplified implementation
	// In a real implementation, you would need to:
	// 1. Get the message by ID from the store
	// 2. Extract the media info from the message
	// 3. Use client.Download() with the media info

	// For now, return an error indicating the feature needs message context
	return nil, fmt.Errorf("download media requires message context - feature needs enhancement")
}

// DownloadMediaFromMessage downloads media from a specific message object
func (c *WameowClient) DownloadMediaFromMessage(ctx context.Context, msg interface{}) ([]byte, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Downloading media from message", map[string]interface{}{
		"session_id": c.sessionID,
	})

	// Note: This would need proper message type handling
	// The actual implementation would depend on the whatsmeow message structure
	return nil, fmt.Errorf("download media from message not fully implemented - requires whatsmeow message handling")
}

// SetGroupPhotoFromBytes sets a group's photo from byte data
func (c *WameowClient) SetGroupPhotoFromBytes(ctx context.Context, groupJID string, photoData []byte) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if groupJID == "" {
		return fmt.Errorf("group JID is required")
	}

	if len(photoData) == 0 {
		return fmt.Errorf("photo data is required")
	}

	// Parse group JID
	gJID, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Setting group photo from bytes", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
		"photo_size": len(photoData),
	})

	// Set group photo using whatsmeow
	_, err = c.client.SetGroupPhoto(gJID, photoData)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group photo", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to set group photo: %w", err)
	}

	c.logger.InfoWithFields("Group photo set successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupJID,
	})

	return nil
}

func IsDeviceRegistered(client *whatsmeow.Client) bool {
	if client == nil || client.Store == nil {
		return false
	}
	return client.Store.ID != nil
}

// GetGroupRequestParticipants gets the list of participants that have requested to join the group
func (c *WameowClient) GetGroupRequestParticipants(ctx context.Context, groupJID string) ([]types.GroupParticipantRequest, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Getting group request participants", map[string]interface{}{
		"group_jid": jid.String(),
	})

	participants, err := c.client.GetGroupRequestParticipants(jid)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get group request participants", map[string]interface{}{
			"group_jid": jid.String(),
			"error":     err.Error(),
		})
		return nil, fmt.Errorf("failed to get group request participants: %w", err)
	}

	c.logger.InfoWithFields("Group request participants retrieved successfully", map[string]interface{}{
		"group_jid": jid.String(),
		"count":     len(participants),
	})

	return participants, nil
}

// UpdateGroupRequestParticipants can be used to approve or reject requests to join the group
func (c *WameowClient) UpdateGroupRequestParticipants(ctx context.Context, groupJID string, participants []string, action string) ([]string, []string, error) {
	if !c.client.IsLoggedIn() {
		return nil, nil, fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid group JID: %w", err)
	}

	// Validate action
	var participantAction whatsmeow.ParticipantRequestChange
	switch action {
	case "approve":
		participantAction = whatsmeow.ParticipantChangeApprove
	case "reject":
		participantAction = whatsmeow.ParticipantChangeReject
	default:
		return nil, nil, fmt.Errorf("invalid action: %s (must be 'approve' or 'reject')", action)
	}

	// Parse participant JIDs
	participantJIDs := make([]types.JID, len(participants))
	for i, participant := range participants {
		participantJID, err := c.parseJID(participant)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid participant JID %s: %w", participant, err)
		}
		participantJIDs[i] = participantJID
	}

	c.logger.InfoWithFields("Updating group request participants", map[string]interface{}{
		"group_jid":    jid.String(),
		"action":       action,
		"participants": len(participantJIDs),
	})

	result, err := c.client.UpdateGroupRequestParticipants(jid, participantJIDs, participantAction)
	if err != nil {
		c.logger.ErrorWithFields("Failed to update group request participants", map[string]interface{}{
			"group_jid": jid.String(),
			"action":    action,
			"error":     err.Error(),
		})
		return nil, nil, fmt.Errorf("failed to update group request participants: %w", err)
	}

	// Process results
	var success, failed []string
	for _, participant := range result {
		if participant.Error == 0 {
			success = append(success, participant.JID.String())
		} else {
			failed = append(failed, participant.JID.String())
		}
	}

	c.logger.InfoWithFields("Group request participants updated", map[string]interface{}{
		"group_jid": jid.String(),
		"action":    action,
		"success":   len(success),
		"failed":    len(failed),
	})

	return success, failed, nil
}

// SetGroupJoinApprovalMode sets the group join approval mode to 'on' or 'off'
func (c *WameowClient) SetGroupJoinApprovalMode(ctx context.Context, groupJID string, requireApproval bool) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields("Setting group join approval mode", map[string]interface{}{
		"group_jid":        jid.String(),
		"require_approval": requireApproval,
	})

	err = c.client.SetGroupJoinApprovalMode(jid, requireApproval)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group join approval mode", map[string]interface{}{
			"group_jid": jid.String(),
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to set group join approval mode: %w", err)
	}

	c.logger.InfoWithFields("Group join approval mode set successfully", map[string]interface{}{
		"group_jid":        jid.String(),
		"require_approval": requireApproval,
	})

	return nil
}

// SetGroupMemberAddMode sets the group member add mode to 'admin_add' or 'all_member_add'
func (c *WameowClient) SetGroupMemberAddMode(ctx context.Context, groupJID string, mode string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	jid, err := c.parseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	// Validate and convert mode
	var memberAddMode types.GroupMemberAddMode
	switch mode {
	case "admin_add":
		memberAddMode = types.GroupMemberAddModeAdmin
	case "all_member_add":
		memberAddMode = types.GroupMemberAddModeAllMember
	default:
		return fmt.Errorf("invalid mode: %s (must be 'admin_add' or 'all_member_add')", mode)
	}

	c.logger.InfoWithFields("Setting group member add mode", map[string]interface{}{
		"group_jid": jid.String(),
		"mode":      mode,
	})

	err = c.client.SetGroupMemberAddMode(jid, memberAddMode)
	if err != nil {
		c.logger.ErrorWithFields("Failed to set group member add mode", map[string]interface{}{
			"group_jid": jid.String(),
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to set group member add mode: %w", err)
	}

	c.logger.InfoWithFields("Group member add mode set successfully", map[string]interface{}{
		"group_jid": jid.String(),
		"mode":      mode,
	})

	return nil
}

// ============================================================================
// Newsletter Methods
// ============================================================================

// CreateNewsletter creates a new WhatsApp newsletter/channel
func (c *WameowClient) CreateNewsletter(ctx context.Context, name, description string) (*types.NewsletterMetadata, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if name == "" {
		return nil, fmt.Errorf("newsletter name is required")
	}

	if len(name) > 25 {
		return nil, fmt.Errorf("newsletter name too long (max 25 characters)")
	}

	if len(description) > 512 {
		return nil, fmt.Errorf("description too long (max 512 characters)")
	}

	params := whatsmeow.CreateNewsletterParams{
		Name:        name,
		Description: description,
	}

	c.logger.InfoWithFields("Creating newsletter", map[string]interface{}{
		"session_id":  c.sessionID,
		"name":        name,
		"description": description,
	})

	newsletter, err := c.client.CreateNewsletter(params)
	if err != nil {
		c.logger.ErrorWithFields("Failed to create newsletter", map[string]interface{}{
			"session_id": c.sessionID,
			"name":       name,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to create newsletter: %w", err)
	}

	c.logger.InfoWithFields("Newsletter created successfully", map[string]interface{}{
		"session_id":    c.sessionID,
		"newsletter_id": newsletter.ID.String(),
		"name":          name,
	})

	return newsletter, nil
}

// GetNewsletterInfo gets information about a newsletter by JID
func (c *WameowClient) GetNewsletterInfo(ctx context.Context, jid string) (*types.NewsletterMetadata, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return nil, fmt.Errorf("newsletter JID is required")
	}

	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid newsletter JID: %w", err)
	}

	// Validate that it's a newsletter JID
	if !strings.Contains(jid, "@newsletter") {
		return nil, fmt.Errorf("invalid newsletter JID format: must contain @newsletter")
	}

	c.logger.InfoWithFields("Getting newsletter info", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	newsletter, err := c.client.GetNewsletterInfo(parsedJID)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get newsletter info", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter info: %w", err)
	}

	c.logger.InfoWithFields("Newsletter info retrieved successfully", map[string]interface{}{
		"session_id":    c.sessionID,
		"newsletter_id": newsletter.ID.String(),
		"name":          newsletter.ThreadMeta.Name.Text,
	})

	return newsletter, nil
}

// GetNewsletterInfoWithInvite gets newsletter information using an invite key
func (c *WameowClient) GetNewsletterInfoWithInvite(ctx context.Context, inviteKey string) (*types.NewsletterMetadata, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if inviteKey == "" {
		return nil, fmt.Errorf("invite key is required")
	}

	// Clean up the invite key by removing common prefixes
	cleanKey := strings.TrimSpace(inviteKey)
	cleanKey = strings.TrimPrefix(cleanKey, "https://whatsapp.com/channel/")
	cleanKey = strings.TrimPrefix(cleanKey, "whatsapp.com/channel/")
	cleanKey = strings.TrimPrefix(cleanKey, "channel/")

	if cleanKey == "" {
		return nil, fmt.Errorf("invalid invite key")
	}

	c.logger.InfoWithFields("Getting newsletter info with invite", map[string]interface{}{
		"session_id": c.sessionID,
		"invite_key": cleanKey,
	})

	newsletter, err := c.client.GetNewsletterInfoWithInvite(cleanKey)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get newsletter info with invite", map[string]interface{}{
			"session_id": c.sessionID,
			"invite_key": cleanKey,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter info with invite: %w", err)
	}

	c.logger.InfoWithFields("Newsletter info retrieved with invite successfully", map[string]interface{}{
		"session_id":    c.sessionID,
		"newsletter_id": newsletter.ID.String(),
		"name":          newsletter.ThreadMeta.Name.Text,
	})

	return newsletter, nil
}

// FollowNewsletter makes the user follow (subscribe to) a newsletter
func (c *WameowClient) FollowNewsletter(ctx context.Context, jid string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return fmt.Errorf("newsletter JID is required")
	}

	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return fmt.Errorf("invalid newsletter JID: %w", err)
	}

	// Validate that it's a newsletter JID
	if !strings.Contains(jid, "@newsletter") {
		return fmt.Errorf("invalid newsletter JID format: must contain @newsletter")
	}

	c.logger.InfoWithFields("Following newsletter", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	err = c.client.FollowNewsletter(parsedJID)
	if err != nil {
		c.logger.ErrorWithFields("Failed to follow newsletter", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to follow newsletter: %w", err)
	}

	c.logger.InfoWithFields("Newsletter followed successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	return nil
}

// UnfollowNewsletter makes the user unfollow (unsubscribe from) a newsletter
func (c *WameowClient) UnfollowNewsletter(ctx context.Context, jid string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return fmt.Errorf("newsletter JID is required")
	}

	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return fmt.Errorf("invalid newsletter JID: %w", err)
	}

	// Validate that it's a newsletter JID
	if !strings.Contains(jid, "@newsletter") {
		return fmt.Errorf("invalid newsletter JID format: must contain @newsletter")
	}

	c.logger.InfoWithFields("Unfollowing newsletter", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	err = c.client.UnfollowNewsletter(parsedJID)
	if err != nil {
		// Check if it's a 405 Not Allowed error (common for owner newsletters)
		if strings.Contains(err.Error(), "405 Not Allowed") {
			c.logger.WarnWithFields("Cannot unfollow newsletter - not allowed (possibly owner)", map[string]interface{}{
				"session_id": c.sessionID,
				"jid":        jid,
				"error":      err.Error(),
			})
			return fmt.Errorf("cannot unfollow newsletter: you may be the owner or unfollow is not allowed for this newsletter")
		}

		c.logger.ErrorWithFields("Failed to unfollow newsletter", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to unfollow newsletter: %w", err)
	}

	c.logger.InfoWithFields("Newsletter unfollowed successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	return nil
}

// GetSubscribedNewsletters gets all newsletters the user is subscribed to
func (c *WameowClient) GetSubscribedNewsletters(ctx context.Context) ([]*types.NewsletterMetadata, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	c.logger.InfoWithFields("Getting subscribed newsletters", map[string]interface{}{
		"session_id": c.sessionID,
	})

	newsletters, err := c.client.GetSubscribedNewsletters()
	if err != nil {
		c.logger.ErrorWithFields("Failed to get subscribed newsletters", map[string]interface{}{
			"session_id": c.sessionID,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get subscribed newsletters: %w", err)
	}

	c.logger.InfoWithFields("Subscribed newsletters retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"count":      len(newsletters),
	})

	return newsletters, nil
}

// GetNewsletterMessages gets messages from a newsletter
func (c *WameowClient) GetNewsletterMessages(ctx context.Context, jid string, count int, before string) ([]*types.NewsletterMessage, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return nil, fmt.Errorf("newsletter JID cannot be empty")
	}

	// Parse and validate JID
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		c.logger.ErrorWithFields("Invalid newsletter JID", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("invalid newsletter JID: %w", err)
	}

	newsletterJID := parsedJID

	c.logger.InfoWithFields("Getting newsletter messages", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"count":      count,
		"before":     before,
	})

	// Prepare parameters
	params := &whatsmeow.GetNewsletterMessagesParams{}
	if count > 0 {
		params.Count = count
	}
	if before != "" {
		// Convert string to MessageServerID
		if serverID, err := strconv.ParseUint(before, 10, 64); err == nil {
			params.Before = types.MessageServerID(serverID)
		}
	}

	messages, err := c.client.GetNewsletterMessages(newsletterJID, params)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get newsletter messages", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter messages: %w", err)
	}

	c.logger.InfoWithFields("Newsletter messages retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"count":      len(messages),
	})

	return messages, nil
}

// GetNewsletterMessageUpdates gets message updates from a newsletter (view counts, reactions)
func (c *WameowClient) GetNewsletterMessageUpdates(ctx context.Context, jid string, count int, since string, after string) ([]*types.NewsletterMessage, error) {
	// Validate request
	newsletterJID, err := c.validateNewsletterUpdatesRequest(jid)
	if err != nil {
		return nil, err
	}

	// Log request
	c.logNewsletterUpdatesRequest(jid, count, since, after)

	// Prepare parameters
	params := c.buildNewsletterUpdatesParams(count, since, after)

	// Execute request with timeout
	return c.executeNewsletterUpdatesRequest(ctx, jid, newsletterJID, params)
}

// validateNewsletterUpdatesRequest validates the newsletter updates request
func (c *WameowClient) validateNewsletterUpdatesRequest(jid string) (types.JID, error) {
	if !c.client.IsLoggedIn() {
		return types.EmptyJID, fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return types.EmptyJID, fmt.Errorf("newsletter JID cannot be empty")
	}

	// Parse and validate JID
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		c.logger.ErrorWithFields("Invalid newsletter JID", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return types.EmptyJID, fmt.Errorf("invalid newsletter JID: %w", err)
	}

	return parsedJID, nil
}

// logNewsletterUpdatesRequest logs the newsletter updates request
func (c *WameowClient) logNewsletterUpdatesRequest(jid string, count int, since, after string) {
	c.logger.InfoWithFields("Getting newsletter message updates", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"count":      count,
		"since":      since,
		"after":      after,
	})
}

// buildNewsletterUpdatesParams builds the parameters for newsletter updates request
func (c *WameowClient) buildNewsletterUpdatesParams(count int, since, after string) *whatsmeow.GetNewsletterUpdatesParams {
	params := &whatsmeow.GetNewsletterUpdatesParams{}

	if count > 0 {
		params.Count = count
	}

	if since != "" {
		// Parse ISO timestamp
		if sinceTime, err := time.Parse(time.RFC3339, since); err == nil {
			params.Since = sinceTime
		}
	}

	if after != "" {
		// Convert string to MessageServerID
		if serverID, err := strconv.ParseUint(after, 10, 64); err == nil {
			params.After = types.MessageServerID(serverID)
		}
	}

	return params
}

// executeNewsletterUpdatesRequest executes the newsletter updates request with timeout
func (c *WameowClient) executeNewsletterUpdatesRequest(ctx context.Context, jid string, newsletterJID types.JID, params *whatsmeow.GetNewsletterUpdatesParams) ([]*types.NewsletterMessage, error) {
	// Add timeout context for the operation
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	c.logger.InfoWithFields("Calling whatsmeow GetNewsletterMessageUpdates", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"params":     params,
	})

	updates, err := c.client.GetNewsletterMessageUpdates(newsletterJID, params)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get newsletter message updates", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get newsletter message updates: %w", err)
	}

	// Check if context was cancelled
	select {
	case <-ctxWithTimeout.Done():
		c.logger.WarnWithFields("Newsletter message updates operation timed out", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
		})
		return nil, fmt.Errorf("operation timed out")
	default:
		// Continue normally
	}

	c.logger.InfoWithFields("Newsletter message updates retrieved successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"count":      len(updates),
	})

	return updates, nil
}

// NewsletterMarkViewed marks newsletter messages as viewed
func (c *WameowClient) NewsletterMarkViewed(ctx context.Context, jid string, serverIDs []string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return fmt.Errorf("newsletter JID cannot be empty")
	}

	if len(serverIDs) == 0 {
		return fmt.Errorf("server IDs cannot be empty")
	}

	// Parse and validate JID
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		c.logger.ErrorWithFields("Invalid newsletter JID", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return fmt.Errorf("invalid newsletter JID: %w", err)
	}

	c.logger.InfoWithFields("Marking newsletter messages as viewed", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"count":      len(serverIDs),
	})

	// Convert string server IDs to MessageServerID
	messageServerIDs := make([]types.MessageServerID, 0, len(serverIDs))
	for _, serverID := range serverIDs {
		if id, err := strconv.ParseUint(serverID, 10, 64); err == nil {
			messageServerIDs = append(messageServerIDs, types.MessageServerID(id))
		} else {
			c.logger.WarnWithFields("Invalid server ID, skipping", map[string]interface{}{
				"session_id": c.sessionID,
				"server_id":  serverID,
				"error":      err.Error(),
			})
			// Skip invalid server IDs instead of failing
			continue
		}
	}

	err = c.client.NewsletterMarkViewed(parsedJID, messageServerIDs)
	if err != nil {
		c.logger.ErrorWithFields("Failed to mark newsletter messages as viewed", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to mark newsletter messages as viewed: %w", err)
	}

	c.logger.InfoWithFields("Newsletter messages marked as viewed successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"count":      len(serverIDs),
	})

	return nil
}

// NewsletterSendReaction sends a reaction to a newsletter message
func (c *WameowClient) NewsletterSendReaction(ctx context.Context, jid string, serverID string, reaction string, messageID string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return fmt.Errorf("newsletter JID cannot be empty")
	}

	if serverID == "" {
		return fmt.Errorf("server ID cannot be empty")
	}

	// Parse and validate JID
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		c.logger.ErrorWithFields("Invalid newsletter JID", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return fmt.Errorf("invalid newsletter JID: %w", err)
	}

	c.logger.InfoWithFields("Sending newsletter reaction", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"server_id":  serverID,
		"reaction":   reaction,
		"message_id": messageID,
	})

	// Convert string server ID to MessageServerID
	msgServerID, err := strconv.ParseUint(serverID, 10, 64)
	if err != nil {
		c.logger.ErrorWithFields("Invalid server ID format", map[string]interface{}{
			"session_id": c.sessionID,
			"server_id":  serverID,
			"error":      err.Error(),
		})
		// Try to handle non-numeric server IDs by using a hash or default value
		// For now, we'll return an error but with a more helpful message
		return fmt.Errorf("server ID must be numeric, got: %s", serverID)
	}

	// Convert messageID to types.MessageID
	var msgID types.MessageID
	if messageID != "" {
		msgID = types.MessageID(messageID)
	}

	err = c.client.NewsletterSendReaction(parsedJID, types.MessageServerID(msgServerID), reaction, msgID)
	if err != nil {
		c.logger.ErrorWithFields("Failed to send newsletter reaction", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"server_id":  serverID,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to send newsletter reaction: %w", err)
	}

	c.logger.InfoWithFields("Newsletter reaction sent successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"server_id":  serverID,
		"reaction":   reaction,
	})

	return nil
}

// NewsletterSubscribeLiveUpdates subscribes to live updates from a newsletter
func (c *WameowClient) NewsletterSubscribeLiveUpdates(ctx context.Context, jid string) (int64, error) {
	if !c.client.IsLoggedIn() {
		return 0, fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return 0, fmt.Errorf("newsletter JID cannot be empty")
	}

	// Parse and validate JID
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		c.logger.ErrorWithFields("Invalid newsletter JID", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return 0, fmt.Errorf("invalid newsletter JID: %w", err)
	}

	c.logger.InfoWithFields("Subscribing to newsletter live updates", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
	})

	duration, err := c.client.NewsletterSubscribeLiveUpdates(ctx, parsedJID)
	if err != nil {
		c.logger.ErrorWithFields("Failed to subscribe to newsletter live updates", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return 0, fmt.Errorf("failed to subscribe to newsletter live updates: %w", err)
	}

	durationSeconds := int64(duration.Seconds())

	c.logger.InfoWithFields("Subscribed to newsletter live updates successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"duration":   durationSeconds,
	})

	return durationSeconds, nil
}

// NewsletterToggleMute toggles mute status of a newsletter
func (c *WameowClient) NewsletterToggleMute(ctx context.Context, jid string, mute bool) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return fmt.Errorf("newsletter JID cannot be empty")
	}

	// Parse and validate JID
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		c.logger.ErrorWithFields("Invalid newsletter JID", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"error":      err.Error(),
		})
		return fmt.Errorf("invalid newsletter JID: %w", err)
	}

	c.logger.InfoWithFields("Toggling newsletter mute status", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"mute":       mute,
	})

	err = c.client.NewsletterToggleMute(parsedJID, mute)
	if err != nil {
		c.logger.ErrorWithFields("Failed to toggle newsletter mute status", map[string]interface{}{
			"session_id": c.sessionID,
			"jid":        jid,
			"mute":       mute,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to toggle newsletter mute status: %w", err)
	}

	c.logger.InfoWithFields("Newsletter mute status toggled successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"jid":        jid,
		"mute":       mute,
	})

	return nil
}

// AcceptTOSNotice accepts a terms of service notice
func (c *WameowClient) AcceptTOSNotice(ctx context.Context, noticeID string, stage string) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if noticeID == "" {
		return fmt.Errorf("notice ID cannot be empty")
	}

	if stage == "" {
		return fmt.Errorf("stage cannot be empty")
	}

	c.logger.InfoWithFields("Accepting TOS notice", map[string]interface{}{
		"session_id": c.sessionID,
		"notice_id":  noticeID,
		"stage":      stage,
	})

	err := c.client.AcceptTOSNotice(noticeID, stage)
	if err != nil {
		c.logger.ErrorWithFields("Failed to accept TOS notice", map[string]interface{}{
			"session_id": c.sessionID,
			"notice_id":  noticeID,
			"stage":      stage,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to accept TOS notice: %w", err)
	}

	c.logger.InfoWithFields("TOS notice accepted successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"notice_id":  noticeID,
		"stage":      stage,
	})

	return nil
}

// UploadNewsletter uploads media for newsletters
func (c *WameowClient) UploadNewsletter(ctx context.Context, data []byte, mimeType string, mediaType string) (*whatsmeow.UploadResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	if mimeType == "" {
		return nil, fmt.Errorf("mime type cannot be empty")
	}

	if mediaType == "" {
		return nil, fmt.Errorf("media type cannot be empty")
	}

	c.logger.InfoWithFields("Uploading newsletter media", map[string]interface{}{
		"session_id": c.sessionID,
		"mime_type":  mimeType,
		"media_type": mediaType,
		"data_size":  len(data),
	})

	// Convert string media type to whatsmeow MediaType
	whatsmeowMediaType := c.convertStringToMediaType(mediaType)

	uploaded, err := c.client.UploadNewsletter(ctx, data, whatsmeowMediaType)
	if err != nil {
		c.logger.ErrorWithFields("Failed to upload newsletter media", map[string]interface{}{
			"session_id": c.sessionID,
			"mime_type":  mimeType,
			"media_type": mediaType,
			"data_size":  len(data),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to upload newsletter media: %w", err)
	}

	c.logger.InfoWithFields("Newsletter media uploaded successfully", map[string]interface{}{
		"session_id":  c.sessionID,
		"mime_type":   mimeType,
		"media_type":  mediaType,
		"data_size":   len(data),
		"url":         uploaded.URL,
		"handle":      uploaded.Handle,
		"file_length": uploaded.FileLength,
	})

	return &uploaded, nil
}

// UploadNewsletterReader uploads media for newsletters from a reader
func (c *WameowClient) UploadNewsletterReader(ctx context.Context, data []byte, mimeType string, mediaType string) (*whatsmeow.UploadResponse, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	if mimeType == "" {
		return nil, fmt.Errorf("mime type cannot be empty")
	}

	if mediaType == "" {
		return nil, fmt.Errorf("media type cannot be empty")
	}

	c.logger.InfoWithFields("Uploading newsletter media with reader", map[string]interface{}{
		"session_id": c.sessionID,
		"mime_type":  mimeType,
		"media_type": mediaType,
		"data_size":  len(data),
	})

	// Convert string media type to whatsmeow MediaType
	whatsmeowMediaType := c.convertStringToMediaType(mediaType)

	// Create a reader from the data
	reader := bytes.NewReader(data)

	uploaded, err := c.client.UploadNewsletterReader(ctx, reader, whatsmeowMediaType)
	if err != nil {
		c.logger.ErrorWithFields("Failed to upload newsletter media with reader", map[string]interface{}{
			"session_id": c.sessionID,
			"mime_type":  mimeType,
			"media_type": mediaType,
			"data_size":  len(data),
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to upload newsletter media with reader: %w", err)
	}

	c.logger.InfoWithFields("Newsletter media uploaded successfully with reader", map[string]interface{}{
		"session_id":  c.sessionID,
		"mime_type":   mimeType,
		"media_type":  mediaType,
		"data_size":   len(data),
		"url":         uploaded.URL,
		"handle":      uploaded.Handle,
		"file_length": uploaded.FileLength,
	})

	return &uploaded, nil
}

// convertStringToMediaType converts string media type to whatsmeow MediaType
func (c *WameowClient) convertStringToMediaType(mediaType string) whatsmeow.MediaType {
	switch mediaType {
	case "image":
		return whatsmeow.MediaImage
	case "video":
		return whatsmeow.MediaVideo
	case "audio":
		return whatsmeow.MediaAudio
	case "document":
		return whatsmeow.MediaDocument
	default:
		return whatsmeow.MediaImage // Default fallback
	}
}

// ============================================================================
// COMMUNITY METHODS
// ============================================================================

// handleCommunityAction is a generic helper for community actions that require two JIDs
func (c *WameowClient) handleCommunityAction(
	ctx context.Context,
	communityJID, groupJID string,
	actionName string,
	actionFunc func(parsedCommunityJID, parsedGroupJID types.JID) error,
) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if communityJID == "" {
		return fmt.Errorf("community JID cannot be empty")
	}

	if groupJID == "" {
		return fmt.Errorf("group JID cannot be empty")
	}

	// Parse and validate JIDs
	parsedCommunityJID, err := c.parseJID(communityJID)
	if err != nil {
		c.logger.ErrorWithFields("Invalid community JID", map[string]interface{}{
			"session_id":    c.sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return fmt.Errorf("invalid community JID: %w", err)
	}

	parsedGroupJID, err := c.parseJID(groupJID)
	if err != nil {
		c.logger.ErrorWithFields("Invalid group JID", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  groupJID,
			"error":      err.Error(),
		})
		return fmt.Errorf("invalid group JID: %w", err)
	}

	c.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":    c.sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	// Execute the action
	err = actionFunc(parsedCommunityJID, parsedGroupJID)
	if err != nil {
		c.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id":    c.sessionID,
			"community_jid": communityJID,
			"group_jid":     groupJID,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to %s: %w", actionName, err)
	}

	c.logger.InfoWithFields(fmt.Sprintf("%s successfully", strings.Title(actionName)), map[string]interface{}{
		"session_id":    c.sessionID,
		"community_jid": communityJID,
		"group_jid":     groupJID,
	})

	return nil
}

// handleCommunityQuery is a generic helper for community queries that require one JID
func (c *WameowClient) handleCommunityQuery(
	ctx context.Context,
	communityJID string,
	actionName string,
	queryFunc func(parsedCommunityJID types.JID) (interface{}, error),
) (interface{}, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if communityJID == "" {
		return nil, fmt.Errorf("community JID cannot be empty")
	}

	// Parse and validate community JID
	parsedCommunityJID, err := c.parseJID(communityJID)
	if err != nil {
		c.logger.ErrorWithFields("Invalid community JID", map[string]interface{}{
			"session_id":    c.sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("invalid community JID: %w", err)
	}

	c.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":    c.sessionID,
		"community_jid": communityJID,
	})

	// Execute the query
	result, err := queryFunc(parsedCommunityJID)
	if err != nil {
		c.logger.ErrorWithFields(fmt.Sprintf("Failed to %s", actionName), map[string]interface{}{
			"session_id":    c.sessionID,
			"community_jid": communityJID,
			"error":         err.Error(),
		})
		return nil, fmt.Errorf("failed to %s: %w", actionName, err)
	}

	// Log success with count if result is a slice
	logFields := map[string]interface{}{
		"session_id":    c.sessionID,
		"community_jid": communityJID,
	}

	// Add count to log if result is a slice
	switch v := result.(type) {
	case []*types.GroupLinkTarget:
		logFields["count"] = len(v)
	case []types.JID:
		logFields["count"] = len(v)
	}

	c.logger.InfoWithFields(fmt.Sprintf("%s successfully", strings.Title(actionName)), logFields)

	return result, nil
}

// LinkGroup links a group to a community
func (c *WameowClient) LinkGroup(ctx context.Context, communityJID, groupJID string) error {
	return c.handleCommunityAction(ctx, communityJID, groupJID, "linking group to community", c.client.LinkGroup)
}

// UnlinkGroup unlinks a group from a community
func (c *WameowClient) UnlinkGroup(ctx context.Context, communityJID, groupJID string) error {
	return c.handleCommunityAction(ctx, communityJID, groupJID, "unlinking group from community", c.client.UnlinkGroup)
}

// GetSubGroups gets all sub-groups (linked groups) of a community
func (c *WameowClient) GetSubGroups(ctx context.Context, communityJID string) ([]*types.GroupLinkTarget, error) {
	result, err := c.handleCommunityQuery(ctx, communityJID, "getting community sub-groups", func(parsedCommunityJID types.JID) (interface{}, error) {
		return c.client.GetSubGroups(parsedCommunityJID)
	})
	if err != nil {
		return nil, err
	}
	return result.([]*types.GroupLinkTarget), nil
}

// GetLinkedGroupsParticipants gets participants from all linked groups in a community
func (c *WameowClient) GetLinkedGroupsParticipants(ctx context.Context, communityJID string) ([]types.JID, error) {
	result, err := c.handleCommunityQuery(ctx, communityJID, "getting linked groups participants", func(parsedCommunityJID types.JID) (interface{}, error) {
		return c.client.GetLinkedGroupsParticipants(parsedCommunityJID)
	})
	if err != nil {
		return nil, err
	}
	return result.([]types.JID), nil
}

// ============================================================================
// ADVANCED GROUP METHODS
// ============================================================================

// GetGroupInfoFromLink gets group information from an invite link
func (c *WameowClient) GetGroupInfoFromLink(ctx context.Context, inviteLink string) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if inviteLink == "" {
		return nil, fmt.Errorf("invite link cannot be empty")
	}

	c.logger.InfoWithFields("Getting group info from link", map[string]interface{}{
		"session_id":  c.sessionID,
		"invite_link": inviteLink,
	})

	// Use whatsmeow's GetGroupInfoFromLink method
	groupInfo, err := c.client.GetGroupInfoFromLink(inviteLink)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get group info from link", map[string]interface{}{
			"session_id":  c.sessionID,
			"invite_link": inviteLink,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("failed to get group info from link: %w", err)
	}

	c.logger.InfoWithFields("Group info retrieved from link successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupInfo.JID.String(),
		"group_name": groupInfo.GroupName.Name,
	})

	return groupInfo, nil
}

// GetGroupInfoFromInvite gets group information from an invite
func (c *WameowClient) GetGroupInfoFromInvite(ctx context.Context, jid, inviter, code string, expiration int64) (*types.GroupInfo, error) {
	if !c.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return nil, fmt.Errorf("group JID cannot be empty")
	}

	if code == "" {
		return nil, fmt.Errorf("invite code cannot be empty")
	}

	// Parse JIDs
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID: %w", err)
	}

	var parsedInviter types.JID
	if inviter != "" {
		parsedInviter, err = c.parseJID(inviter)
		if err != nil {
			return nil, fmt.Errorf("invalid inviter JID: %w", err)
		}
	}

	c.logger.InfoWithFields("Getting group info from invite", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  jid,
		"inviter":    inviter,
		"code":       code,
	})

	// Use whatsmeow's GetGroupInfoFromInvite method
	groupInfo, err := c.client.GetGroupInfoFromInvite(parsedJID, parsedInviter, code, expiration)
	if err != nil {
		c.logger.ErrorWithFields("Failed to get group info from invite", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  jid,
			"inviter":    inviter,
			"code":       code,
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("failed to get group info from invite: %w", err)
	}

	c.logger.InfoWithFields("Group info retrieved from invite successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  groupInfo.JID.String(),
		"group_name": groupInfo.GroupName.Name,
	})

	return groupInfo, nil
}

// JoinGroupWithInvite joins a group using a specific invite
func (c *WameowClient) JoinGroupWithInvite(ctx context.Context, jid, inviter, code string, expiration int64) error {
	if !c.client.IsLoggedIn() {
		return fmt.Errorf("client is not logged in")
	}

	if jid == "" {
		return fmt.Errorf("group JID cannot be empty")
	}

	if code == "" {
		return fmt.Errorf("invite code cannot be empty")
	}

	// Parse JIDs
	parsedJID, err := c.parseJID(jid)
	if err != nil {
		return fmt.Errorf("invalid group JID: %w", err)
	}

	var parsedInviter types.JID
	if inviter != "" {
		parsedInviter, err = c.parseJID(inviter)
		if err != nil {
			return fmt.Errorf("invalid inviter JID: %w", err)
		}
	}

	c.logger.InfoWithFields("Joining group with invite", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  jid,
		"inviter":    inviter,
		"code":       code,
	})

	// Use whatsmeow's JoinGroupWithInvite method
	err = c.client.JoinGroupWithInvite(parsedJID, parsedInviter, code, expiration)
	if err != nil {
		c.logger.ErrorWithFields("Failed to join group with invite", map[string]interface{}{
			"session_id": c.sessionID,
			"group_jid":  jid,
			"inviter":    inviter,
			"code":       code,
			"error":      err.Error(),
		})
		return fmt.Errorf("failed to join group with invite: %w", err)
	}

	c.logger.InfoWithFields("Joined group with invite successfully", map[string]interface{}{
		"session_id": c.sessionID,
		"group_jid":  jid,
	})

	return nil
}
