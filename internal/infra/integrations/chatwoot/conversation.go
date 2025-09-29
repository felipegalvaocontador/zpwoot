package chatwoot

import (
	"fmt"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// ConversationManager handles conversation operations between WhatsApp and Chatwoot
type ConversationManager struct {
	logger *logger.Logger
	client ports.ChatwootClient
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(logger *logger.Logger, client ports.ChatwootClient) *ConversationManager {
	return &ConversationManager{
		logger: logger,
		client: client,
	}
}

// CreateOrGetConversation creates or gets an existing conversation
func (cm *ConversationManager) CreateOrGetConversation(contactID, inboxID int) (*ports.ChatwootConversation, error) {
	// Try to get existing conversation
	conversation, err := cm.client.GetConversation(contactID, inboxID)
	if err == nil {
		return conversation, nil
	}

	// Create new conversation
	conversation, err = cm.client.CreateConversation(contactID, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// ReopenConversation reopens a resolved conversation
func (cm *ConversationManager) ReopenConversation(conversationID int) error {
	err := cm.client.UpdateConversationStatus(conversationID, "open")
	if err != nil {
		return fmt.Errorf("failed to reopen conversation: %w", err)
	}

	return nil
}

// SetConversationPending sets conversation status to pending
func (cm *ConversationManager) SetConversationPending(conversationID int) error {
	cm.logger.InfoWithFields("Setting conversation to pending", map[string]interface{}{
		"conversation_id": conversationID,
	})

	err := cm.client.UpdateConversationStatus(conversationID, "pending")
	if err != nil {
		return fmt.Errorf("failed to set conversation pending: %w", err)
	}

	return nil
}

// ResolveConversation resolves a conversation
func (cm *ConversationManager) ResolveConversation(conversationID int) error {
	cm.logger.InfoWithFields("Resolving conversation", map[string]interface{}{
		"conversation_id": conversationID,
	})

	err := cm.client.UpdateConversationStatus(conversationID, "resolved")
	if err != nil {
		return fmt.Errorf("failed to resolve conversation: %w", err)
	}

	return nil
}

// GetConversationByID gets a conversation by ID
func (cm *ConversationManager) GetConversationByID(conversationID int) (*ports.ChatwootConversation, error) {
	cm.logger.InfoWithFields("Getting conversation by ID", map[string]interface{}{
		"conversation_id": conversationID,
	})

	conversation, err := cm.client.GetConversationByID(conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversation, nil
}

// HandleConversationStatusChange handles conversation status changes based on Evolution API logic
func (cm *ConversationManager) HandleConversationStatusChange(conversationID int, newStatus string, reopenConversation, conversationPending bool) error {
	cm.logger.InfoWithFields("Handling conversation status change", map[string]interface{}{
		"conversation_id":      conversationID,
		"new_status":           newStatus,
		"reopen_conversation":  reopenConversation,
		"conversation_pending": conversationPending,
	})

	switch newStatus {
	case "resolved":
		if reopenConversation {
			// If reopen is enabled, reopen the conversation when it gets resolved
			return cm.ReopenConversation(conversationID)
		}
	case "open":
		if conversationPending {
			// If conversation pending is enabled, set to pending instead of open
			return cm.SetConversationPending(conversationID)
		}
	}

	return nil
}

// SendMessage sends a message to a conversation
func (cm *ConversationManager) SendMessage(conversationID int, content string) (*ports.ChatwootMessage, error) {
	cm.logger.InfoWithFields("Sending message to conversation", map[string]interface{}{
		"conversation_id": conversationID,
		"content_length":  len(content),
	})

	message, err := cm.client.SendMessage(conversationID, content)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return message, nil
}

// GetConversationMessages gets messages from a conversation
func (cm *ConversationManager) GetConversationMessages(conversationID int, before int) ([]ports.ChatwootMessage, error) {
	messages, err := cm.client.GetMessages(conversationID, before)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// IsConversationActive checks if a conversation is active (not resolved)
func (cm *ConversationManager) IsConversationActive(conversation *ports.ChatwootConversation) bool {
	return conversation.Status != "resolved"
}

// GetConversationStats returns statistics about conversations
func (cm *ConversationManager) GetConversationStats(conversations []ports.ChatwootConversation) ConversationStats {
	stats := ConversationStats{}

	for _, conv := range conversations {
		stats.Total++
		switch conv.Status {
		case "open":
			stats.Open++
		case "pending":
			stats.Pending++
		case "resolved":
			stats.Resolved++
		default:
			stats.Other++
		}
	}

	return stats
}

// ConversationStats represents conversation statistics
type ConversationStats struct {
	Total    int `json:"total"`
	Open     int `json:"open"`
	Pending  int `json:"pending"`
	Resolved int `json:"resolved"`
	Other    int `json:"other"`
}

// ConversationConfig represents configuration for conversation management
type ConversationConfig struct {
	ReopenConversation  bool `json:"reopen_conversation"`
	ConversationPending bool `json:"conversation_pending"`
	AutoResolveTimeout  int  `json:"auto_resolve_timeout"` // minutes
}

// ApplyConversationConfig applies conversation configuration
func (cm *ConversationManager) ApplyConversationConfig(conversationID int, config ConversationConfig) error {
	conversation, err := cm.GetConversationByID(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// Apply status based on config
	if conversation.Status == "resolved" && config.ReopenConversation {
		return cm.ReopenConversation(conversationID)
	}

	if conversation.Status == "open" && config.ConversationPending {
		return cm.SetConversationPending(conversationID)
	}

	return nil
}

// CreateConversationWithContact creates a conversation for a specific contact
func (cm *ConversationManager) CreateConversationWithContact(contact *ports.ChatwootContact, inboxID int) (*ports.ChatwootConversation, error) {
	return cm.CreateOrGetConversation(contact.ID, inboxID)
}

// UpdateConversationAttributes updates conversation custom attributes
func (cm *ConversationManager) UpdateConversationAttributes(conversationID int, attributes map[string]interface{}) error {
	cm.logger.InfoWithFields("Updating conversation attributes", map[string]interface{}{
		"conversation_id": conversationID,
		"attributes":      attributes,
	})

	// TODO: Implement conversation attribute updates
	// This would require extending the ChatwootClient interface
	cm.logger.WarnWithFields("Conversation attribute updates not implemented", map[string]interface{}{
		"conversation_id": conversationID,
	})

	return nil
}
