package chatwoot

import (
	"context"
	"fmt"

	"zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type UseCase interface {
	CreateConfig(ctx context.Context, sessionID string, req *CreateChatwootConfigRequest) (*CreateChatwootConfigResponse, error)
	GetConfig(ctx context.Context) (*ChatwootConfigResponse, error)
	UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ChatwootConfigResponse, error)
	DeleteConfig(ctx context.Context) error
	SyncContact(ctx context.Context, req *SyncContactRequest) (*SyncContactResponse, error)
	SyncConversation(ctx context.Context, req *SyncConversationRequest) (*SyncConversationResponse, error)
	SendMessageToChatwoot(ctx context.Context, req *SendMessageToChatwootRequest) (*SendMessageToChatwootResponse, error)
	ProcessWebhook(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error
	TestConnection(ctx context.Context) (*TestChatwootConnectionResponse, error)
	GetStats(ctx context.Context) (*ChatwootStatsResponse, error)
	AutoCreateInbox(ctx context.Context, sessionID, inboxName, webhookURL string) error
}

type useCaseImpl struct {
	chatwootRepo        ports.ChatwootRepository
	chatwootIntegration ports.ChatwootIntegration
	chatwootManager     ports.ChatwootManager
	chatwootService     *chatwoot.Service
	logger              *logger.Logger
}

func NewUseCase(
	chatwootRepo ports.ChatwootRepository,
	chatwootIntegration ports.ChatwootIntegration,
	chatwootManager ports.ChatwootManager,
	chatwootService *chatwoot.Service,
	logger *logger.Logger,
) UseCase {
	return &useCaseImpl{
		chatwootRepo:        chatwootRepo,
		chatwootIntegration: chatwootIntegration,
		chatwootManager:     chatwootManager,
		chatwootService:     chatwootService,
		logger:              logger,
	}
}

func (uc *useCaseImpl) CreateConfig(ctx context.Context, sessionID string, req *CreateChatwootConfigRequest) (*CreateChatwootConfigResponse, error) {
	domainReq, err := req.ToCreateChatwootConfigRequest(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	config, err := uc.chatwootService.CreateConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &CreateChatwootConfigResponse{
		ID:        config.ID.String(),
		URL:       config.URL,
		AccountID: config.AccountID,
		InboxID:   config.InboxID,
		Active:    config.Enabled,
		CreatedAt: config.CreatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) GetConfig(ctx context.Context) (*ChatwootConfigResponse, error) {
	config, err := uc.chatwootService.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	response := FromChatwootConfig(config)
	return response, nil
}

func (uc *useCaseImpl) UpdateConfig(ctx context.Context, req *UpdateChatwootConfigRequest) (*ChatwootConfigResponse, error) {
	domainReq := req.ToUpdateChatwootConfigRequest()

	config, err := uc.chatwootService.UpdateConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := FromChatwootConfig(config)
	return response, nil
}

func (uc *useCaseImpl) DeleteConfig(ctx context.Context) error {
	return uc.chatwootService.DeleteConfig(ctx)
}

func (uc *useCaseImpl) SyncContact(ctx context.Context, req *SyncContactRequest) (*SyncContactResponse, error) {
	domainReq := &chatwoot.SyncContactRequest{
		PhoneNumber: req.PhoneNumber,
		Name:        req.Name,
		Email:       req.Email,
		Attributes:  req.Attributes,
	}

	contact, err := uc.chatwootService.SyncContact(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &SyncContactResponse{
		ID:          contact.ID,
		PhoneNumber: contact.PhoneNumber,
		Name:        contact.Name,
		Email:       contact.Email,
		Attributes:  contact.CustomAttributes,
		CreatedAt:   contact.CreatedAt,
		UpdatedAt:   contact.UpdatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) SyncConversation(ctx context.Context, req *SyncConversationRequest) (*SyncConversationResponse, error) {
	domainReq := &chatwoot.SyncConversationRequest{
		ContactID: req.ContactID,
		SessionID: req.SessionID,
	}

	conversation, err := uc.chatwootService.SyncConversation(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &SyncConversationResponse{
		ID:          conversation.ID,
		ContactID:   conversation.ContactID,
		SessionID:   req.SessionID, // Use from request since domain doesn't have it
		PhoneNumber: "",            // This would need to be retrieved from contact
		Status:      conversation.Status,
		CreatedAt:   conversation.CreatedAt,
		UpdatedAt:   conversation.UpdatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) SendMessageToChatwoot(ctx context.Context, req *SendMessageToChatwootRequest) (*SendMessageToChatwootResponse, error) {
	domainReq := &chatwoot.SendMessageToChatwootRequest{
		ConversationID: req.ConversationID,
		Content:        req.Content,
		MessageType:    req.MessageType,
		ContentType:    req.ContentType,
		Attachments:    convertAttachments(req.Attachments),
		Metadata:       req.Metadata,
	}

	message, err := uc.chatwootService.SendMessage(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &SendMessageToChatwootResponse{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		ContentType:    message.ContentType,
		Metadata:       message.Metadata,
		CreatedAt:      message.CreatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) ProcessWebhook(ctx context.Context, sessionID string, payload *ChatwootWebhookPayload) error {
	// Convert app-layer payload to domain-layer payload
	domainPayload := uc.convertToDomainPayload(payload)

	// Resolve sender phone number if missing
	if err := uc.resolveSenderPhoneNumber(ctx, domainPayload); err != nil {
		uc.logger.WarnWithFields("Failed to resolve sender phone number", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		// Continue processing even if phone resolution fails
	}

	return uc.chatwootService.ProcessWebhook(ctx, sessionID, domainPayload)
}

// convertToDomainPayload converts app-layer webhook payload to domain-layer payload
func (uc *useCaseImpl) convertToDomainPayload(payload *ChatwootWebhookPayload) *chatwoot.ChatwootWebhookPayload {
	domainPayload := &chatwoot.ChatwootWebhookPayload{
		Event:   payload.Event,
		Account: chatwoot.ChatwootAccount{ID: payload.Account.ID, Name: payload.Account.Name},
		Conversation: chatwoot.ChatwootConversation{
			ID:        payload.Conversation.ID,
			ContactID: payload.Conversation.ContactID,
			InboxID:   payload.Conversation.InboxID,
			Status:    payload.Conversation.Status,
		},
	}

	// Map message data from nested or top-level fields
	uc.mapMessageData(payload, domainPayload)

	// Map sender data
	uc.mapSenderData(payload, domainPayload)

	return domainPayload
}

// mapMessageData maps message data from nested or top-level fields
func (uc *useCaseImpl) mapMessageData(payload *ChatwootWebhookPayload, domainPayload *chatwoot.ChatwootWebhookPayload) {
	// Message from nested field (if present)
	if payload.Message.ID != 0 || payload.Message.Content != "" {
		m := payload.Message
		domainPayload.Message = &chatwoot.ChatwootMessage{
			ID:          m.ID,
			Content:     m.Content,
			MessageType: m.MessageType,
			ContentType: m.ContentType,
			Private:     m.Private,
			SourceID:    m.SourceID,
		}
		// Set shortcuts
		domainPayload.ID = m.ID
		domainPayload.Content = m.Content
		domainPayload.MessageType = m.MessageType
		domainPayload.ContentType = m.ContentType
		domainPayload.Private = m.Private
		return
	}

	// Or from top-level (real Chatwoot webhook format)
	if payload.Content != "" || payload.ID != 0 {
		domainPayload.ID = payload.ID
		domainPayload.Content = payload.Content
		domainPayload.MessageType = payload.MessageType
		domainPayload.ContentType = payload.ContentType
		domainPayload.Private = payload.Private
		if payload.SourceID != nil {
			domainPayload.SourceID = payload.SourceID
		}
	}
}

// mapSenderData maps sender data from root or nested message sender
func (uc *useCaseImpl) mapSenderData(payload *ChatwootWebhookPayload, domainPayload *chatwoot.ChatwootWebhookPayload) {
	// Prefer root sender, fallback to nested message.sender
	if payload.Sender.PhoneNumber != "" || payload.Sender.Name != "" {
		domainPayload.Sender.Name = payload.Sender.Name
		domainPayload.Sender.PhoneNumber = payload.Sender.PhoneNumber
	} else if payload.Message.Sender != nil {
		domainPayload.Sender.Name = payload.Message.Sender.Name
		domainPayload.Sender.PhoneNumber = payload.Message.Sender.PhoneNumber
	}
}

// resolveSenderPhoneNumber attempts to resolve sender phone number via Chatwoot API
func (uc *useCaseImpl) resolveSenderPhoneNumber(ctx context.Context, domainPayload *chatwoot.ChatwootWebhookPayload) error {
	// Skip if phone number already available or no conversation ID
	if domainPayload.Sender.PhoneNumber != "" || domainPayload.Conversation.ID == 0 {
		return nil
	}

	// Get Chatwoot client for the session (assuming we can get sessionID from context or payload)
	// For now, we'll need to pass sessionID - this is a limitation we need to address
	sessionID := "default" // TODO: Get sessionID from context or payload
	client, err := uc.chatwootManager.GetClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get chatwoot client: %w", err)
	}

	// Try to resolve phone number via conversation -> contact
	if phone := uc.resolvePhoneViaContact(client, domainPayload.Conversation.ID); phone != "" {
		domainPayload.Sender.PhoneNumber = phone
		domainPayload.Contact.PhoneNumber = phone
		return nil
	}

	// Fallback to conversation sender phone
	if phone := uc.resolvePhoneViaConversation(client, domainPayload.Conversation.ID); phone != "" {
		domainPayload.Sender.PhoneNumber = phone
		domainPayload.Contact.PhoneNumber = phone
		return nil
	}

	return fmt.Errorf("could not resolve sender phone number")
}

// resolvePhoneViaContact resolves phone number via conversation -> contact
func (uc *useCaseImpl) resolvePhoneViaContact(client ports.ChatwootClient, conversationID int) string {
	conv, err := client.GetConversationByID(conversationID)
	if err != nil || conv == nil || conv.ContactID == 0 {
		return ""
	}

	contact, err := client.GetContact(conv.ContactID)
	if err != nil || contact == nil {
		return ""
	}

	return contact.PhoneNumber
}

// resolvePhoneViaConversation resolves phone number via conversation meta
func (uc *useCaseImpl) resolvePhoneViaConversation(client ports.ChatwootClient, conversationID int) string {
	phone, err := client.GetConversationSenderPhone(conversationID)
	if err != nil {
		return ""
	}
	return phone
}

func (uc *useCaseImpl) TestConnection(ctx context.Context) (*TestChatwootConnectionResponse, error) {
	result, err := uc.chatwootService.TestConnection(ctx)
	if err != nil {
		return &TestChatwootConnectionResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	response := &TestChatwootConnectionResponse{
		Success:     result.Success,
		AccountName: result.AccountName,
		InboxName:   result.InboxName,
	}

	if result.Error != nil {
		response.Error = result.Error.Error()
	}

	return response, nil
}

func (uc *useCaseImpl) GetStats(ctx context.Context) (*ChatwootStatsResponse, error) {
	stats, err := uc.chatwootService.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	response := &ChatwootStatsResponse{
		TotalContacts:       stats.TotalContacts,
		TotalConversations:  stats.TotalConversations,
		ActiveConversations: stats.ActiveConversations,
		MessagesSent:        int(stats.MessagesSent),
		MessagesReceived:    int(stats.MessagesReceived),
	}

	return response, nil
}

func convertAttachments(attachments []ChatwootAttachment) []chatwoot.ChatwootAttachment {
	domainAttachments := make([]chatwoot.ChatwootAttachment, len(attachments))
	for i, att := range attachments {
		domainAttachments[i] = chatwoot.ChatwootAttachment{
			FileType: att.FileType,
			FileName: att.FileName,
		}
	}
	return domainAttachments
}

func (uc *useCaseImpl) AutoCreateInbox(ctx context.Context, sessionID, inboxName, webhookURL string) error {
	// Get Chatwoot client for the session
	client, err := uc.chatwootManager.GetClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get chatwoot client for session %s: %w", sessionID, err)
	}

	// Create inbox in Chatwoot
	inbox, err := client.CreateInbox(inboxName, webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create inbox in Chatwoot: %w", err)
	}

	// Update configuration with the new inbox ID
	inboxIDStr := fmt.Sprintf("%d", inbox.ID)
	updateReq := &chatwoot.UpdateChatwootConfigRequest{
		InboxID: &inboxIDStr,
	}

	_, err = uc.chatwootService.UpdateConfig(ctx, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update config with inbox ID: %w", err)
	}

	return nil
}
