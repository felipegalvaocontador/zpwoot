package chatwoot

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"

	"github.com/google/uuid"
)

type Service struct {
	logger        *logger.Logger
	repository    ports.ChatwootRepository
	wameowManager ports.WameowManager
	messageMapper ports.ChatwootMessageMapper // Optional - for storing outgoing messages
}

func NewService(logger *logger.Logger, repository ports.ChatwootRepository, wameowManager ports.WameowManager) *Service {
	return &Service{
		logger:        logger,
		repository:    repository,
		wameowManager: wameowManager,
	}
}

// SetMessageMapper sets the message mapper for storing outgoing messages
func (s *Service) SetMessageMapper(messageMapper ports.ChatwootMessageMapper) {
	s.messageMapper = messageMapper
}

// ============================================================================
// CONFIGURATION MANAGEMENT
// ============================================================================

func (s *Service) CreateConfig(ctx context.Context, req *CreateChatwootConfigRequest) (*ports.ChatwootConfig, error) {
	// Apply defaults to request
	defaults := s.applyConfigDefaults(req)

	// Build configuration
	config := s.buildChatwootConfig(req, defaults)

	// Persist to repository
	if err := s.repository.CreateConfig(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// configDefaults holds default values for configuration
type configDefaults struct {
	enabled        bool
	autoCreate     bool
	signMsg        bool
	signDelimiter  string
	reopenConv     bool
	convPending    bool
	importContacts bool
	importMessages bool
	importDays     int
	mergeBrazil    bool
	ignoreJids     []string
}

// applyConfigDefaults applies default values to configuration request
func (s *Service) applyConfigDefaults(req *CreateChatwootConfigRequest) *configDefaults {
	defaults := &configDefaults{
		enabled:        true,
		autoCreate:     false,
		signMsg:        false,
		signDelimiter:  "\n\n",
		reopenConv:     true,
		convPending:    false,
		importContacts: false,
		importMessages: false,
		importDays:     60,
		mergeBrazil:    true,
		ignoreJids:     []string{},
	}

	// Override defaults with request values
	if req.Enabled != nil {
		defaults.enabled = *req.Enabled
	}
	if req.AutoCreate != nil {
		defaults.autoCreate = *req.AutoCreate
	}
	if req.SignMsg != nil {
		defaults.signMsg = *req.SignMsg
	}
	if req.SignDelimiter != nil {
		defaults.signDelimiter = *req.SignDelimiter
	}
	if req.ReopenConv != nil {
		defaults.reopenConv = *req.ReopenConv
	}
	if req.ConvPending != nil {
		defaults.convPending = *req.ConvPending
	}
	if req.ImportContacts != nil {
		defaults.importContacts = *req.ImportContacts
	}
	if req.ImportMessages != nil {
		defaults.importMessages = *req.ImportMessages
	}
	if req.ImportDays != nil {
		defaults.importDays = *req.ImportDays
	}
	if req.MergeBrazil != nil {
		defaults.mergeBrazil = *req.MergeBrazil
	}
	if req.IgnoreJids != nil {
		defaults.ignoreJids = req.IgnoreJids
	}

	return defaults
}

// buildChatwootConfig builds the final configuration object
func (s *Service) buildChatwootConfig(req *CreateChatwootConfigRequest, defaults *configDefaults) *ports.ChatwootConfig {
	return &ports.ChatwootConfig{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		URL:       req.URL,
		Token:     req.Token,
		AccountID: req.AccountID,
		InboxID:   req.InboxID,
		Enabled:   defaults.enabled,

		// Advanced configuration
		InboxName:      req.InboxName,
		AutoCreate:     defaults.autoCreate,
		SignMsg:        defaults.signMsg,
		SignDelimiter:  defaults.signDelimiter,
		ReopenConv:     defaults.reopenConv,
		ConvPending:    defaults.convPending,
		ImportContacts: defaults.importContacts,
		ImportMessages: defaults.importMessages,
		ImportDays:     defaults.importDays,
		MergeBrazil:    defaults.mergeBrazil,
		Organization:   req.Organization,
		Logo:           req.Logo,
		Number:         req.Number,
		IgnoreJids:     defaults.ignoreJids,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *Service) GetConfig(ctx context.Context) (*ports.ChatwootConfig, error) {
	config, err := s.repository.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Service) GetConfigBySessionID(ctx context.Context, sessionID string) (*ports.ChatwootConfig, error) {
	config, err := s.repository.GetConfigBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Service) UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ports.ChatwootConfig, error) {
	// Get existing config first
	existingConfig, err := s.repository.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Update config with request values
	config := s.updateConfigFields(existingConfig, req)

	// Persist changes
	if err := s.repository.UpdateConfig(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

// updateConfigFields updates configuration fields from request
func (s *Service) updateConfigFields(existingConfig *ports.ChatwootConfig, req *UpdateChatwootConfigRequest) *ports.ChatwootConfig {
	// Copy existing config
	config := *existingConfig
	config.UpdatedAt = time.Now()

	// Update basic fields
	s.updateBasicConfigFields(&config, req)

	// Update advanced fields
	s.updateAdvancedConfigFields(&config, req)

	// Update optional fields
	s.updateOptionalConfigFields(&config, req)

	return &config
}

// updateBasicConfigFields updates basic configuration fields
func (s *Service) updateBasicConfigFields(config *ports.ChatwootConfig, req *UpdateChatwootConfigRequest) {
	if req.URL != nil {
		config.URL = *req.URL
	}
	if req.Token != nil {
		config.Token = *req.Token
	}
	if req.AccountID != nil {
		config.AccountID = *req.AccountID
	}
	if req.InboxID != nil {
		config.InboxID = req.InboxID
	}
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.InboxName != nil {
		config.InboxName = req.InboxName
	}
}

// updateAdvancedConfigFields updates advanced configuration fields
func (s *Service) updateAdvancedConfigFields(config *ports.ChatwootConfig, req *UpdateChatwootConfigRequest) {
	if req.AutoCreate != nil {
		config.AutoCreate = *req.AutoCreate
	}
	if req.SignMsg != nil {
		config.SignMsg = *req.SignMsg
	}
	if req.SignDelimiter != nil {
		config.SignDelimiter = *req.SignDelimiter
	}
	if req.ReopenConv != nil {
		config.ReopenConv = *req.ReopenConv
	}
	if req.ConvPending != nil {
		config.ConvPending = *req.ConvPending
	}
	if req.ImportContacts != nil {
		config.ImportContacts = *req.ImportContacts
	}
	if req.ImportMessages != nil {
		config.ImportMessages = *req.ImportMessages
	}
	if req.ImportDays != nil {
		config.ImportDays = *req.ImportDays
	}
	if req.MergeBrazil != nil {
		config.MergeBrazil = *req.MergeBrazil
	}
}

// updateOptionalConfigFields updates optional configuration fields
func (s *Service) updateOptionalConfigFields(config *ports.ChatwootConfig, req *UpdateChatwootConfigRequest) {
	if req.Organization != nil {
		config.Organization = req.Organization
	}
	if req.Logo != nil {
		config.Logo = req.Logo
	}
	if req.Number != nil {
		config.Number = req.Number
	}
	if req.IgnoreJids != nil {
		config.IgnoreJids = req.IgnoreJids
	}
}

func (s *Service) DeleteConfig(ctx context.Context) error {
	if err := s.repository.DeleteConfig(ctx); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// SYNC OPERATIONS (MOCK IMPLEMENTATIONS)
// ============================================================================

func (s *Service) SyncContact(ctx context.Context, req *SyncContactRequest) (*ChatwootContact, error) {
	contact := &ChatwootContact{
		ID:               1, // This would be assigned by Chatwoot
		Name:             req.Name,
		PhoneNumber:      req.PhoneNumber,
		Email:            req.Email,
		CustomAttributes: req.Attributes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return contact, nil
}

func (s *Service) SyncConversation(ctx context.Context, req *SyncConversationRequest) (*ChatwootConversation, error) {
	conversation := &ChatwootConversation{
		ID:        1, // This would be assigned by Chatwoot
		ContactID: req.ContactID,
		Status:    "open",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return conversation, nil
}

func (s *Service) SendMessage(ctx context.Context, req *SendMessageToChatwootRequest) (*ChatwootMessage, error) {
	message := &ChatwootMessage{
		ID:             1, // This would be assigned by Chatwoot
		ConversationID: req.ConversationID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		ContentType:    req.ContentType,
		Metadata:       req.Metadata,
		CreatedAt:      time.Now(),
	}

	return message, nil
}

// ============================================================================
// WEBHOOK PROCESSING
// ============================================================================

func (s *Service) ProcessWebhook(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	// Delay 500ms to avoid race conditions (based on Evolution API)
	time.Sleep(500 * time.Millisecond)

	// Skip private messages
	if payload.Message != nil && payload.Message.Private {
		return nil
	}

	// Skip message updates without deletion
	if payload.Event == "message_updated" {
		// TODO: Handle message deletion if needed
		s.logger.DebugWithFields("Skipping message update", map[string]interface{}{
			"session_id": sessionID,
			"event":      payload.Event,
		})
		return nil
	}

	// Process conversation status changes
	if payload.Event == "conversation_status_changed" {
		// TODO: Handle conversation status changes if needed
		s.logger.DebugWithFields("Processing conversation status change", map[string]interface{}{
			"session_id": sessionID,
			"event":      payload.Event,
		})
		return nil
	}

	// Handle typing events (ignore them as they don't require action)
	if payload.Event == "conversation_typing_on" || payload.Event == "conversation_typing_off" {
		return nil
	}

	// Process new messages (main functionality)
	if payload.Event == "message_created" {
		return s.handleMessageCreated(ctx, sessionID, payload)
	}

	return nil
}

// handleMessageCreated processes new messages from Chatwoot
func (s *Service) handleMessageCreated(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	// Extract and validate message details
	content, messageType, _, isPrivate := s.extractMessageDetails(payload)

	// Apply message filters
	if s.shouldSkipMessage(content, messageType, isPrivate, payload) {
		return nil
	}

	return s.sendToWhatsApp(ctx, sessionID, payload, content)
}

// extractMessageDetails extracts message information from webhook payload
func (s *Service) extractMessageDetails(payload *ChatwootWebhookPayload) (content, messageType string, messageID int, isPrivate bool) {
	if payload.Message != nil {
		// Legacy format
		content = payload.Message.Content
		messageType = payload.Message.MessageType
		messageID = payload.Message.ID
		isPrivate = payload.Message.Private
	} else {
		// Current Chatwoot format
		content = payload.Content
		messageType = payload.MessageType
		messageID = payload.ID
		isPrivate = payload.Private
	}

	// Try to extract content from alternative sources if empty
	if content == "" {
		content = s.extractContentFromAlternativeSources(payload)
	}

	return content, messageType, messageID, isPrivate
}

// extractContentFromAlternativeSources tries to find content in other payload fields
func (s *Service) extractContentFromAlternativeSources(payload *ChatwootWebhookPayload) string {
	// Check content_attributes
	if payload.ContentAttributes != nil {
		if textContent, ok := payload.ContentAttributes["text"].(string); ok {
			return textContent
		}
	}

	// Check conversation messages
	if len(payload.Conversation.Messages) > 0 {
		lastMessage := payload.Conversation.Messages[len(payload.Conversation.Messages)-1]
		if lastMessage.Content != "" {
			return lastMessage.Content
		}
	}

	return ""
}

// shouldSkipMessage determines if a message should be skipped based on various filters
func (s *Service) shouldSkipMessage(content, messageType string, isPrivate bool, payload *ChatwootWebhookPayload) bool {
	// Skip empty outgoing messages
	if content == "" && messageType == "outgoing" {
		return true
	}

	// Skip incoming messages (already processed from WhatsApp)
	if messageType == "incoming" {
		return true
	}

	// Skip empty outgoing messages
	if messageType == "outgoing" && content == "" {
		return true
	}

	// Skip private messages
	if isPrivate {
		return true
	}

	// Skip bot messages (source_id starting with WAID:)
	return s.isBotMessage(payload)
}

// isBotMessage checks if message is from a bot based on source_id
func (s *Service) isBotMessage(payload *ChatwootWebhookPayload) bool {
	var sourceID string
	if payload.Message != nil {
		sourceID = payload.Message.SourceID
	} else if payload.SourceID != nil {
		sourceID = *payload.SourceID
	}

	return sourceID != "" && len(sourceID) >= 5 && sourceID[:5] == "WAID:"
}

// sendToWhatsApp sends a message from Chatwoot to WhatsApp
func (s *Service) sendToWhatsApp(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload, content string) error {
	// Extract recipient phone number
	phoneNumber, err := s.extractRecipientPhone(payload)
	if err != nil {
		return err
	}

	// Format content and extract message ID
	formattedContent := s.formatContentForWhatsApp(content)
	messageID := s.extractMessageID(payload)

	// Send message to WhatsApp
	result, err := s.wameowManager.SendMessage(sessionID, phoneNumber, "text", formattedContent, "", "", "", 0, 0, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to send message to WhatsApp: %w", err)
	}

	// Store message for tracking (non-blocking)
	_ = s.storeOutgoingMessage(ctx, sessionID, result.MessageID, phoneNumber, formattedContent, result.Timestamp, messageID, payload.Conversation.ID)

	return nil
}

// extractRecipientPhone determines the recipient phone number from payload
func (s *Service) extractRecipientPhone(payload *ChatwootWebhookPayload) (string, error) {
	messageType := "outgoing" // Default assumption
	if payload.Message != nil {
		messageType = payload.Message.MessageType
	}

	var phoneNumber string
	if messageType == "outgoing" {
		// Outgoing: agent → client, recipient is Contact
		phoneNumber = payload.Contact.PhoneNumber
	} else {
		// Incoming: client → agent, recipient is Sender
		phoneNumber = payload.Sender.PhoneNumber
	}

	if phoneNumber == "" {
		return "", fmt.Errorf("no valid recipient phone number found for %s message", messageType)
	}

	return phoneNumber, nil
}

// extractMessageID extracts message ID from payload
func (s *Service) extractMessageID(payload *ChatwootWebhookPayload) int {
	if payload.Message != nil {
		return payload.Message.ID
	}
	return payload.ID
}

// storeOutgoingMessage stores an outgoing message in the zpMessage table
func (s *Service) storeOutgoingMessage(ctx context.Context, sessionID, whatsappMessageID, phoneNumber, content string, timestamp time.Time, chatwootMessageID, chatwootConversationID int) error {
	// Check if we have a message mapper available
	if s.messageMapper == nil {
		s.logger.WarnWithFields("Message mapper not available, cannot store outgoing message", map[string]interface{}{
			"session_id":      sessionID,
			"whatsapp_msg_id": whatsappMessageID,
		})
		return nil
	}

	// Create mapping for outgoing message
	mapping, err := s.messageMapper.CreateMapping(ctx, sessionID, whatsappMessageID, phoneNumber, phoneNumber, "text", content, timestamp, true)
	if err != nil {
		return fmt.Errorf("failed to create mapping for outgoing message: %w", err)
	}

	// Update mapping with Chatwoot IDs
	err = s.messageMapper.UpdateMapping(ctx, sessionID, whatsappMessageID, chatwootMessageID, chatwootConversationID)
	if err != nil {
		return fmt.Errorf("failed to update mapping with Chatwoot IDs: %w", err)
	}

	s.logger.InfoWithFields("Outgoing message stored in zpMessage table", map[string]interface{}{
		"session_id":               sessionID,
		"whatsapp_msg_id":          whatsappMessageID,
		"chatwoot_msg_id":          chatwootMessageID,
		"chatwoot_conversation_id": chatwootConversationID,
		"mapping_id":               mapping.ID,
	})

	return nil
}

// formatContentForWhatsApp formats message content for WhatsApp
func (s *Service) formatContentForWhatsApp(content string) string {
	// TODO: Use MessageFormatter for consistent formatting across the application
	// For now, return as-is to avoid code duplication
	return content
}

// ============================================================================
// UTILITY METHODS & TYPES
// ============================================================================

// TestConnectionResult represents the result of a connection test
type TestConnectionResult struct {
	Success     bool
	AccountName string
	InboxName   string
	Error       error
}

// ChatwootStats represents Chatwoot integration statistics
type ChatwootStats struct {
	TotalContacts       int
	TotalConversations  int
	ActiveConversations int
	MessagesSent        int64
	MessagesReceived    int64
}

// TestConnection tests the connection to Chatwoot (mock implementation)
func (s *Service) TestConnection(ctx context.Context) (*TestConnectionResult, error) {
	return &TestConnectionResult{
		Success:     true,
		AccountName: "Test Account",
		InboxName:   "Wameow Inbox",
	}, nil
}

// GetStats returns Chatwoot integration statistics (mock implementation)
func (s *Service) GetStats(ctx context.Context) (*ChatwootStats, error) {
	return &ChatwootStats{
		TotalContacts:       100,
		TotalConversations:  50,
		ActiveConversations: 10,
		MessagesSent:        500,
		MessagesReceived:    300,
	}, nil
}

// ============================================================================
// IMPORT VALIDATION
// ============================================================================

// ValidateImportRequest validates an import request according to business rules
func (s *Service) ValidateImportRequest(sessionID string, daysLimit int, inboxID int) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	if daysLimit <= 0 || daysLimit > 365 {
		return fmt.Errorf("days_limit must be between 1 and 365")
	}

	if inboxID <= 0 {
		return fmt.Errorf("inbox_id must be positive")
	}

	return nil
}

// ValidateContactImportRequest validates a contact import request
func (s *Service) ValidateContactImportRequest(sessionID string, inboxID int) error {
	if sessionID == "" {
		return fmt.Errorf("session_id is required")
	}

	if inboxID <= 0 {
		return fmt.Errorf("inbox_id must be positive")
	}

	return nil
}

// ValidateBotContactCreation validates bot contact creation parameters
func (s *Service) ValidateBotContactCreation(inboxID int) error {
	if inboxID <= 0 {
		return fmt.Errorf("inbox_id must be positive")
	}

	return nil
}

// CreateBotContactIfNeeded creates a bot contact if it doesn't exist
// This encapsulates the business logic for bot contact creation
func (s *Service) CreateBotContactIfNeeded(ctx context.Context, client ports.ChatwootClient, inboxID int) error {
	// Validate parameters using domain rules
	if err := s.ValidateBotContactCreation(inboxID); err != nil {
		return err
	}

	// Business rules for bot contact
	botPhone := "123456"
	botName := "Bot"

	// Try to find existing bot contact
	_, err := client.FindContact(botPhone, inboxID)
	if err == nil {
		// Bot contact already exists - this is expected behavior
		s.logger.DebugWithFields("Bot contact already exists", map[string]interface{}{
			"phone":    botPhone,
			"inbox_id": inboxID,
		})
		return nil
	}

	// Create bot contact according to business rules
	s.logger.InfoWithFields("Creating bot contact", map[string]interface{}{
		"phone":    botPhone,
		"name":     botName,
		"inbox_id": inboxID,
	})

	_, err = client.CreateContact(botPhone, botName, inboxID)
	if err != nil {
		return fmt.Errorf("failed to create bot contact: %w", err)
	}

	s.logger.InfoWithFields("Bot contact created successfully", map[string]interface{}{
		"phone":    botPhone,
		"name":     botName,
		"inbox_id": inboxID,
	})

	return nil
}

// ============================================================================
// INBOX MANAGEMENT
// ============================================================================

// ShouldCreateInbox determines if an inbox should be auto-created based on business rules
func (s *Service) ShouldCreateInbox(autoCreate bool, existingInboxes []ports.ChatwootInbox, inboxName string) (bool, *ports.ChatwootInbox) {
	if !autoCreate {
		return false, nil
	}

	// Business rule: Check if inbox already exists
	for _, inbox := range existingInboxes {
		if inbox.Name == inboxName {
			s.logger.DebugWithFields("Inbox already exists", map[string]interface{}{
				"inbox_name": inboxName,
				"inbox_id":   inbox.ID,
			})
			return false, &inbox
		}
	}

	// Business rule: Create inbox if it doesn't exist and auto-create is enabled
	return true, nil
}

// ValidateInboxCreation validates inbox creation parameters
func (s *Service) ValidateInboxCreation(inboxName, webhookURL string) error {
	if inboxName == "" {
		return fmt.Errorf("inbox_name is required")
	}

	if webhookURL == "" {
		return fmt.Errorf("webhook_url is required")
	}

	return nil
}

// ProcessInboxInitialization handles the business logic for inbox initialization
func (s *Service) ProcessInboxInitialization(ctx context.Context, sessionID, inboxName, webhookURL string, autoCreate bool, client ports.ChatwootClient) (*ports.ChatwootInbox, error) {
	// Validate parameters
	if err := s.ValidateInboxCreation(inboxName, webhookURL); err != nil {
		return nil, err
	}

	// Get existing inboxes
	inboxes, err := client.ListInboxes()
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	// Apply business rules to determine if inbox should be created
	shouldCreate, existingInbox := s.ShouldCreateInbox(autoCreate, inboxes, inboxName)

	if existingInbox != nil {
		// Use existing inbox
		return existingInbox, nil
	}

	if shouldCreate {
		// Create new inbox
		s.logger.InfoWithFields("Creating new inbox", map[string]interface{}{
			"inbox_name": inboxName,
			"session_id": sessionID,
		})

		createdInbox, err := client.CreateInbox(inboxName, webhookURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create inbox: %w", err)
		}

		return createdInbox, nil
	}

	return nil, fmt.Errorf("inbox '%s' not found and auto-create is disabled", inboxName)
}

// UpdateConfigWithInbox updates session config with inbox information
func (s *Service) UpdateConfigWithInbox(ctx context.Context, sessionID string, inbox *ports.ChatwootInbox) error {
	if inbox == nil {
		return fmt.Errorf("inbox cannot be nil")
	}

	if inbox.ID <= 0 {
		return fmt.Errorf("invalid inbox ID: %d", inbox.ID)
	}

	// Get current config
	config, err := s.GetConfigBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Update with inbox ID
	inboxIDStr := fmt.Sprintf("%d", inbox.ID)
	config.InboxID = &inboxIDStr

	// Save updated config
	err = s.repository.UpdateConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	s.logger.InfoWithFields("Config updated with inbox", map[string]interface{}{
		"session_id": sessionID,
		"inbox_id":   inbox.ID,
		"inbox_name": inbox.Name,
	})

	return nil
}

// ============================================================================
// INTEGRATION ORCHESTRATION
// ============================================================================

// InitializeChatwootIntegration orchestrates the complete Chatwoot integration setup
func (s *Service) InitializeChatwootIntegration(ctx context.Context, sessionID, inboxName, webhookURL string, autoCreate bool, client ports.ChatwootClient) error {
	// Process inbox initialization using business rules
	inbox, err := s.ProcessInboxInitialization(ctx, sessionID, inboxName, webhookURL, autoCreate, client)
	if err != nil {
		return fmt.Errorf("failed to process inbox initialization: %w", err)
	}

	// Update config with inbox information
	err = s.UpdateConfigWithInbox(ctx, sessionID, inbox)
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Create bot contact if needed (business rule)
	err = s.CreateBotContactIfNeeded(ctx, client, inbox.ID)
	if err != nil {
		s.logger.WarnWithFields("Failed to create bot contact", map[string]interface{}{
			"session_id": sessionID,
			"inbox_id":   inbox.ID,
			"error":      err.Error(),
		})
		// Don't fail the entire initialization if bot contact creation fails
	}

	s.logger.InfoWithFields("Chatwoot integration initialized successfully", map[string]interface{}{
		"session_id": sessionID,
		"inbox_id":   inbox.ID,
		"inbox_name": inbox.Name,
	})

	return nil
}
