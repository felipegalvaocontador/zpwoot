package chatwoot

import (
	"fmt"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

const (
	StatusResolved = "resolved"
	StatusOpen     = "open"
	StatusPending  = "pending"
)

type ConversationManager struct {
	logger *logger.Logger
	client ports.ChatwootClient
}

func NewConversationManager(logger *logger.Logger, client ports.ChatwootClient) *ConversationManager {
	return &ConversationManager{
		logger: logger,
		client: client,
	}
}

func (cm *ConversationManager) CreateOrGetConversation(contactID, inboxID int) (*ports.ChatwootConversation, error) {
	conversation, err := cm.client.GetConversation(contactID, inboxID)
	if err == nil {
		return conversation, nil
	}

	conversation, err = cm.client.CreateConversation(contactID, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

func (cm *ConversationManager) ReopenConversation(conversationID int) error {
	err := cm.client.UpdateConversationStatus(conversationID, "open")
	if err != nil {
		return fmt.Errorf("failed to reopen conversation: %w", err)
	}

	return nil
}

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

func (cm *ConversationManager) ResolveConversation(conversationID int) error {
	cm.logger.InfoWithFields("Resolving conversation", map[string]interface{}{
		"conversation_id": conversationID,
	})

	err := cm.client.UpdateConversationStatus(conversationID, StatusResolved)
	if err != nil {
		return fmt.Errorf("failed to resolve conversation: %w", err)
	}

	return nil
}

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

func (cm *ConversationManager) HandleConversationStatusChange(conversationID int, newStatus string, reopenConversation, conversationPending bool) error {
	cm.logger.InfoWithFields("Handling conversation status change", map[string]interface{}{
		"conversation_id":      conversationID,
		"new_status":           newStatus,
		"reopen_conversation":  reopenConversation,
		"conversation_pending": conversationPending,
	})

	switch newStatus {
	case StatusResolved:
		if reopenConversation {
			return cm.ReopenConversation(conversationID)
		}
	case "open":
		if conversationPending {
			return cm.SetConversationPending(conversationID)
		}
	}

	return nil
}

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

func (cm *ConversationManager) GetConversationMessages(conversationID, before int) ([]ports.ChatwootMessage, error) {
	messages, err := cm.client.GetMessages(conversationID, before)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

func (cm *ConversationManager) IsConversationActive(conversation *ports.ChatwootConversation) bool {
	return conversation.Status != StatusResolved
}

func (cm *ConversationManager) GetConversationStats(conversations []ports.ChatwootConversation) ConversationStats {
	stats := ConversationStats{}

	for _, conv := range conversations {
		stats.Total++
		switch conv.Status {
		case "open":
			stats.Open++
		case "pending":
			stats.Pending++
		case StatusResolved:
			stats.Resolved++
		default:
			stats.Other++
		}
	}

	return stats
}

type ConversationStats struct {
	Total    int `json:"total"`
	Open     int `json:"open"`
	Pending  int `json:"pending"`
	Resolved int `json:"resolved"`
	Other    int `json:"other"`
}

type ConversationConfig struct {
	ReopenConversation  bool `json:"reopen_conversation"`
	ConversationPending bool `json:"conversation_pending"`
	AutoResolveTimeout  int  `json:"auto_resolve_timeout"`
}

func (cm *ConversationManager) ApplyConversationConfig(conversationID int, config ConversationConfig) error {
	conversation, err := cm.GetConversationByID(conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	if conversation.Status == StatusResolved && config.ReopenConversation {
		return cm.ReopenConversation(conversationID)
	}

	if conversation.Status == "open" && config.ConversationPending {
		return cm.SetConversationPending(conversationID)
	}

	return nil
}

func (cm *ConversationManager) CreateConversationWithContact(contact *ports.ChatwootContact, inboxID int) (*ports.ChatwootConversation, error) {
	return cm.CreateOrGetConversation(contact.ID, inboxID)
}

func (cm *ConversationManager) UpdateConversationAttributes(conversationID int, attributes map[string]interface{}) error {
	cm.logger.InfoWithFields("Updating conversation attributes", map[string]interface{}{
		"conversation_id": conversationID,
		"attributes":      attributes,
	})

	cm.logger.WarnWithFields("Conversation attribute updates not implemented", map[string]interface{}{
		"conversation_id": conversationID,
	})

	return nil
}
