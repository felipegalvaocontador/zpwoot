package chatwoot

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type IntegrationManager struct {
	logger          *logger.Logger
	chatwootManager ports.ChatwootManager
	messageMapper   *MessageMapper
	contactSync     *ContactSync
	conversationMgr *ConversationManager
	formatter       *MessageFormatter
}

func NewIntegrationManager(
	logger *logger.Logger,
	chatwootManager ports.ChatwootManager,
	messageMapper *MessageMapper,
	contactSync *ContactSync,
	conversationMgr *ConversationManager,
	formatter *MessageFormatter,
) *IntegrationManager {
	return &IntegrationManager{
		logger:          logger,
		chatwootManager: chatwootManager,
		messageMapper:   messageMapper,
		contactSync:     contactSync,
		conversationMgr: conversationMgr,
		formatter:       formatter,
	}
}

func (im *IntegrationManager) IsEnabled(sessionID string) bool {
	return im.chatwootManager.IsEnabled(sessionID)
}

func (im *IntegrationManager) ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time, fromMe bool) error {
	ctx := context.Background()

	if im.messageMapper.IsMessageMapped(ctx, sessionID, messageID) {
		return nil
	}

	if err := im.createMessageMapping(ctx, sessionID, messageID, from, messageType, content, timestamp, fromMe); err != nil {
		return err
	}

	return im.processMessageToChatwoot(ctx, sessionID, messageID, from, content, messageType, fromMe)
}

func (im *IntegrationManager) createMessageMapping(ctx context.Context, sessionID, messageID, from, messageType, content string, timestamp time.Time, fromMe bool) error {
	chatJID := im.extractChatJID(from)

	_, err := im.messageMapper.CreateMapping(ctx, sessionID, messageID, from, chatJID, messageType, content, timestamp, fromMe)
	if err != nil {
		return fmt.Errorf("failed to create message mapping: %w", err)
	}

	return nil
}

func (im *IntegrationManager) extractChatJID(from string) string {
	return from
}

func (im *IntegrationManager) processMessageToChatwoot(ctx context.Context, sessionID, messageID, from, content, messageType string, fromMe bool) error {
	client, phoneNumber, err := im.setupChatwootClient(ctx, sessionID, messageID, from)
	if err != nil {
		return err
	}

	inboxID, err := im.getInboxID(ctx, sessionID, messageID)
	if err != nil {
		return err
	}

	conversation, err := im.setupContactAndConversation(ctx, client, phoneNumber, sessionID, messageID, inboxID)
	if err != nil {
		return err
	}

	chatwootMessage, err := im.sendMessageToChatwoot(client, conversation.ID, content, messageType, fromMe, ctx, sessionID, messageID)
	if err != nil {
		return err
	}

	return im.finalizeMessageProcessing(ctx, sessionID, messageID, chatwootMessage.ID, conversation.ID)
}

func (im *IntegrationManager) setupChatwootClient(ctx context.Context, sessionID, messageID, from string) (ports.ChatwootClient, string, error) {
	client, err := im.chatwootManager.GetClient(sessionID)
	if err != nil {
		if markErr := im.messageMapper.MarkAsFailed(ctx, sessionID, messageID); markErr != nil {
			im.logger.WarnWithFields("Failed to mark message as failed", map[string]interface{}{
				"session_id": sessionID,
				"message_id": messageID,
				"error":      markErr.Error(),
			})
		}
		return nil, "", fmt.Errorf("failed to get Chatwoot client: %w", err)
	}

	phoneNumber := im.extractPhoneFromJID(from)
	if phoneNumber == "" {
		if markErr := im.messageMapper.MarkAsFailed(ctx, sessionID, messageID); markErr != nil {
			im.logger.WarnWithFields("Failed to mark message as failed", map[string]interface{}{
				"session_id": sessionID,
				"message_id": messageID,
				"error":      markErr.Error(),
			})
		}
		return nil, "", fmt.Errorf("failed to extract phone number from JID: %s", from)
	}

	return client, phoneNumber, nil
}

func (im *IntegrationManager) getInboxID(ctx context.Context, sessionID, messageID string) (int, error) {
	config, err := im.chatwootManager.GetConfig(sessionID)
	if err != nil {
		if markErr := im.messageMapper.MarkAsFailed(ctx, sessionID, messageID); markErr != nil {
			im.logger.WarnWithFields("Failed to mark message as failed", map[string]interface{}{
				"session_id": sessionID,
				"message_id": messageID,
				"error":      markErr.Error(),
			})
		}
		return 0, fmt.Errorf("failed to get Chatwoot config: %w", err)
	}

	inboxID := 1
	if config.InboxID != nil {
		if id, err := strconv.Atoi(*config.InboxID); err == nil {
			inboxID = id
		}
	}

	return inboxID, nil
}

func (im *IntegrationManager) setupContactAndConversation(ctx context.Context, client ports.ChatwootClient, phoneNumber, sessionID, messageID string, inboxID int) (*ports.ChatwootConversation, error) {
	contact, err := im.getOrCreateContact(client, phoneNumber, sessionID, inboxID)
	if err != nil {
		if markErr := im.messageMapper.MarkAsFailed(ctx, sessionID, messageID); markErr != nil {
			im.logger.WarnWithFields("Failed to mark message as failed", map[string]interface{}{
				"session_id": sessionID,
				"message_id": messageID,
				"error":      markErr.Error(),
			})
		}
		return nil, fmt.Errorf("failed to get or create contact: %w", err)
	}

	conversation, err := im.getOrCreateConversation(client, contact.ID, sessionID, inboxID)
	if err != nil {
		if markErr := im.messageMapper.MarkAsFailed(ctx, sessionID, messageID); markErr != nil {
			im.logger.WarnWithFields("Failed to mark message as failed", map[string]interface{}{
				"session_id": sessionID,
				"message_id": messageID,
				"error":      markErr.Error(),
			})
		}
		return nil, fmt.Errorf("failed to get or create conversation: %w", err)
	}

	return conversation, nil
}

func (im *IntegrationManager) sendMessageToChatwoot(client ports.ChatwootClient, conversationID int, content, messageType string, fromMe bool, ctx context.Context, sessionID, messageID string) (*ports.ChatwootMessage, error) {
	formattedContent := im.formatContentForChatwoot(content, messageType)

	chatwootMessageType := "incoming"
	if fromMe {
		chatwootMessageType = "outgoing"
	}

	chatwootMessage, err := client.SendMessageWithType(conversationID, formattedContent, chatwootMessageType)
	if err != nil {
		if markErr := im.messageMapper.MarkAsFailed(ctx, sessionID, messageID); markErr != nil {
			im.logger.WarnWithFields("Failed to mark message as failed", map[string]interface{}{
				"session_id": sessionID,
				"message_id": messageID,
				"error":      markErr.Error(),
			})
		}
		return nil, fmt.Errorf("failed to send message to Chatwoot: %w", err)
	}

	return chatwootMessage, nil
}

func (im *IntegrationManager) finalizeMessageProcessing(ctx context.Context, sessionID, messageID string, chatwootMessageID, conversationID int) error {
	err := im.messageMapper.UpdateMapping(ctx, sessionID, messageID, chatwootMessageID, conversationID)
	if err != nil {
		im.logger.WarnWithFields("Failed to update mapping", map[string]interface{}{
			"message_id":         messageID,
			"cw_message_id":      chatwootMessageID,
			"cw_conversation_id": conversationID,
			"error":              err.Error(),
		})
	}

	im.logger.InfoWithFields("WhatsApp message processed successfully", map[string]interface{}{
		"session_id":         sessionID,
		"message_id":         messageID,
		"cw_message_id":      chatwootMessageID,
		"cw_conversation_id": conversationID,
	})

	return nil
}

func (im *IntegrationManager) extractPhoneFromJID(jid string) string {
	im.logger.DebugWithFields("Extracting phone from JID", map[string]interface{}{
		"original_jid": jid,
	})

	phone := strings.Split(jid, "@")[0]

	phone = regexp.MustCompile(`:\d+`).ReplaceAllString(phone, "")

	if strings.Contains(phone, "-") {
		parts := strings.Split(phone, "-")
		if len(parts) > 0 {
			phone = parts[0]
		}
	}

	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	phone = im.formatBrazilianPhone(phone)

	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	im.logger.DebugWithFields("Extracted phone from JID", map[string]interface{}{
		"original_jid": jid,
		"final_phone":  phone,
	})

	return phone
}

func (im *IntegrationManager) getBrazilianNumbers(query string) []string {
	numbers := []string{query}

	if strings.HasPrefix(query, "+55") {
		if len(query) == 14 {
			withoutNine := query[:5] + query[6:]
			numbers = append(numbers, withoutNine)
		} else if len(query) == 13 {
			withNine := query[:5] + "9" + query[5:]
			numbers = append(numbers, withNine)
		}
	}

	return numbers
}

func (im *IntegrationManager) mergeBrazilianContacts(client ports.ChatwootClient, contacts []*ports.ChatwootContact, sessionID string) (*ports.ChatwootContact, error) {
	if len(contacts) != 2 {
		return nil, fmt.Errorf("expected exactly 2 contacts for merge, got %d", len(contacts))
	}

	var contact14, contact13 *ports.ChatwootContact

	for _, contact := range contacts {
		if len(contact.PhoneNumber) == 14 {
			contact14 = contact
		} else if len(contact.PhoneNumber) == 13 {
			contact13 = contact
		}
	}

	if contact14 != nil && contact13 != nil {
		im.logger.InfoWithFields("Merging Brazilian contacts", map[string]interface{}{
			"session_id":       sessionID,
			"base_contact_id":  contact14.ID,
			"base_phone":       contact14.PhoneNumber,
			"merge_contact_id": contact13.ID,
			"merge_phone":      contact13.PhoneNumber,
		})

		err := im.mergeContacts(client, contact14.ID, contact13.ID, sessionID)
		if err != nil {
			im.logger.ErrorWithFields("Failed to merge Brazilian contacts", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
		}

		return contact14, nil
	}

	return contacts[0], nil
}

func (im *IntegrationManager) mergeContacts(client ports.ChatwootClient, baseContactID, mergeContactID int, sessionID string) error {
	im.logger.InfoWithFields("Merging contacts using Chatwoot API", map[string]interface{}{
		"session_id":       sessionID,
		"base_contact_id":  baseContactID,
		"merge_contact_id": mergeContactID,
	})

	err := client.MergeContacts(baseContactID, mergeContactID)
	if err != nil {
		return fmt.Errorf("failed to merge contacts via Chatwoot API: %w", err)
	}

	return nil
}

func (im *IntegrationManager) formatBrazilianPhone(phone string) string {
	if !strings.HasPrefix(phone, "55") {
		return phone
	}

	localNumber := phone[2:]

	if len(localNumber) == 10 {
		areaCode := localNumber[:2]
		number := localNumber[2:]

		formatted := "55" + areaCode + "9" + number

		im.logger.DebugWithFields("Added 9 to Brazilian mobile", map[string]interface{}{
			"original":  phone,
			"formatted": formatted,
		})

		return formatted
	}

	return phone
}

func (im *IntegrationManager) getOrCreateContact(client ports.ChatwootClient, phoneNumber, sessionID string, inboxID int) (*ports.ChatwootContact, error) {
	phoneNumbers := im.getBrazilianNumbers(phoneNumber)

	var foundContacts []*ports.ChatwootContact
	for _, phone := range phoneNumbers {
		contact, err := client.FindContact(phone, inboxID)
		if err == nil {
			foundContacts = append(foundContacts, contact)
		}
	}

	if len(foundContacts) > 0 {
		if len(foundContacts) == 2 && strings.HasPrefix(phoneNumber, "+55") {
			mergedContact, err := im.mergeBrazilianContacts(client, foundContacts, sessionID)
			if err == nil && mergedContact != nil {
				return mergedContact, nil
			}
		}

		contact := foundContacts[0]
		im.logger.InfoWithFields("Found existing contact", map[string]interface{}{
			"session_id":     sessionID,
			"original_phone": phoneNumber,
			"contact_id":     contact.ID,
			"contacts_found": len(foundContacts),
		})
		return contact, nil
	}

	contactName := phoneNumber
	contact, err := client.CreateContact(phoneNumber, contactName, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	im.logger.InfoWithFields("Created new Chatwoot contact", map[string]interface{}{
		"contact_id":   contact.ID,
		"phone_number": phoneNumber,
		"session_id":   sessionID,
	})

	return contact, nil
}

func (im *IntegrationManager) getOrCreateConversation(client ports.ChatwootClient, contactID int, sessionID string, inboxID int) (*ports.ChatwootConversation, error) {
	im.logger.InfoWithFields("Getting or creating conversation", map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
		"session_id": sessionID,
	})

	conversations, err := client.ListContactConversations(contactID)
	if err != nil {
		im.logger.WarnWithFields("Failed to list contact conversations", map[string]interface{}{
			"contact_id": contactID,
			"error":      err.Error(),
		})
	} else {
		for _, conv := range conversations {
			if conv.InboxID == inboxID && conv.Status != "resolved" {
				im.logger.InfoWithFields("Found existing active conversation", map[string]interface{}{
					"conversation_id": conv.ID,
					"status":          conv.Status,
					"inbox_id":        conv.InboxID,
				})
				return &conv, nil
			}
		}
	}

	im.logger.InfoWithFields("Creating new conversation", map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
	})

	conversation, err := client.CreateConversation(contactID, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	im.logger.InfoWithFields("Created new Chatwoot conversation", map[string]interface{}{
		"conversation_id": conversation.ID,
		"contact_id":      contactID,
		"session_id":      sessionID,
	})

	return conversation, nil
}

func (im *IntegrationManager) formatContentForChatwoot(content, messageType string) string {
	switch messageType {
	case "text":
		return im.formatter.FormatMarkdownForChatwoot(content)
	case "image":
		return "🖼️ **Image**\n" + content
	case "video":
		return "🎥 **Video**\n" + content
	case "audio":
		return "🎵 **Audio**\n" + content
	case "document":
		return "📄 **Document**\n" + content
	case "contact":
		return "👤 **Contact**\n" + content
	case "contacts":
		return "👥 **Contacts**\n" + content
	case "location":
		return "📍 **Location**\n" + content
	case "sticker":
		return "😊 **Sticker**"
	default:
		return content
	}
}

func (im *IntegrationManager) GetMappingStats(sessionID string) (*MappingStats, error) {
	ctx := context.Background()
	return im.messageMapper.GetMappingStats(ctx, sessionID)
}

func (im *IntegrationManager) ProcessPendingMessages(sessionID string, limit int) error {
	ctx := context.Background()

	im.logger.InfoWithFields("Processing pending messages", map[string]interface{}{
		"session_id": sessionID,
		"limit":      limit,
	})

	pendingMappings, err := im.messageMapper.GetPendingMappings(ctx, sessionID, limit)
	if err != nil {
		return fmt.Errorf("failed to get pending mappings: %w", err)
	}

	if len(pendingMappings) == 0 {
		im.logger.DebugWithFields("No pending messages to process", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil
	}

	processed := 0
	failed := 0

	for _, mapping := range pendingMappings {

		err := im.messageMapper.MarkAsFailed(ctx, sessionID, mapping.ZpMessageID)
		if err != nil {
			im.logger.WarnWithFields("Failed to mark mapping as failed", map[string]interface{}{
				"message_id": mapping.ZpMessageID,
				"error":      err.Error(),
			})
			failed++
		} else {
			processed++
		}
	}

	im.logger.InfoWithFields("Processed pending messages", map[string]interface{}{
		"session_id": sessionID,
		"processed":  processed,
		"failed":     failed,
		"total":      len(pendingMappings),
	})

	return nil
}

func (im *IntegrationManager) CleanupOldMappings(sessionID string, olderThanDays int) (int, error) {
	ctx := context.Background()
	return im.messageMapper.CleanupOldMappings(ctx, sessionID, olderThanDays)
}
