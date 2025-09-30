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

// IntegrationManager handles the integration between WhatsApp and Chatwoot
type IntegrationManager struct {
	logger          *logger.Logger
	chatwootManager ports.ChatwootManager
	messageMapper   *MessageMapper
	contactSync     *ContactSync
	conversationMgr *ConversationManager
	formatter       *MessageFormatter
}

// NewIntegrationManager creates a new integration manager
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

// IsEnabled checks if Chatwoot integration is enabled for a session
func (im *IntegrationManager) IsEnabled(sessionID string) bool {
	return im.chatwootManager.IsEnabled(sessionID)
}

// ProcessWhatsAppMessage processes a WhatsApp message for Chatwoot integration
func (im *IntegrationManager) ProcessWhatsAppMessage(sessionID, messageID, from, content, messageType string, timestamp time.Time, fromMe bool) error {
	ctx := context.Background()

	// Skip if message is already mapped (originated from Chatwoot)
	if im.messageMapper.IsMessageMapped(ctx, sessionID, messageID) {
		return nil
	}

	// Create message mapping
	if err := im.createMessageMapping(ctx, sessionID, messageID, from, messageType, content, timestamp, fromMe); err != nil {
		return err
	}

	// Process message through Chatwoot
	return im.processMessageToChatwoot(ctx, sessionID, messageID, from, content, messageType, fromMe)
}

// createMessageMapping creates initial message mapping
func (im *IntegrationManager) createMessageMapping(ctx context.Context, sessionID, messageID, from, messageType, content string, timestamp time.Time, fromMe bool) error {
	chatJID := im.extractChatJID(from)

	_, err := im.messageMapper.CreateMapping(ctx, sessionID, messageID, from, chatJID, messageType, content, timestamp, fromMe)
	if err != nil {
		return fmt.Errorf("failed to create message mapping: %w", err)
	}

	return nil
}

// extractChatJID extracts chat JID from sender
func (im *IntegrationManager) extractChatJID(from string) string {
	// For both groups and individual messages, use the from field as chatJID
	return from
}

// processMessageToChatwoot handles the Chatwoot integration flow
func (im *IntegrationManager) processMessageToChatwoot(ctx context.Context, sessionID, messageID, from, content, messageType string, fromMe bool) error {
	// Setup Chatwoot client and extract phone number
	client, phoneNumber, err := im.setupChatwootClient(ctx, sessionID, messageID, from)
	if err != nil {
		return err
	}

	// Get inbox ID from configuration
	inboxID, err := im.getInboxID(ctx, sessionID, messageID)
	if err != nil {
		return err
	}

	// Get or create contact and conversation
	conversation, err := im.setupContactAndConversation(ctx, client, phoneNumber, sessionID, messageID, inboxID)
	if err != nil {
		return err
	}

	// Send message to Chatwoot
	chatwootMessage, err := im.sendMessageToChatwoot(client, conversation.ID, content, messageType, fromMe, ctx, sessionID, messageID)
	if err != nil {
		return err
	}

	// Update mapping and log success
	return im.finalizeMessageProcessing(ctx, sessionID, messageID, chatwootMessage.ID, conversation.ID)
}

// setupChatwootClient sets up the Chatwoot client and extracts phone number
func (im *IntegrationManager) setupChatwootClient(ctx context.Context, sessionID, messageID, from string) (ports.ChatwootClient, string, error) {
	// Get Chatwoot client
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

	// Extract phone number from JID
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

// getInboxID retrieves the inbox ID from Chatwoot configuration
func (im *IntegrationManager) getInboxID(ctx context.Context, sessionID, messageID string) (int, error) {
	// Get Chatwoot configuration to get inbox ID
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

	// Convert inbox ID from string to int
	inboxID := 1 // Default fallback
	if config.InboxID != nil {
		if id, err := strconv.Atoi(*config.InboxID); err == nil {
			inboxID = id
		}
	}

	return inboxID, nil
}

// setupContactAndConversation gets or creates contact and conversation
func (im *IntegrationManager) setupContactAndConversation(ctx context.Context, client ports.ChatwootClient, phoneNumber, sessionID, messageID string, inboxID int) (*ports.ChatwootConversation, error) {
	// Get or create contact
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

	// Get or create conversation
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

// sendMessageToChatwoot sends the formatted message to Chatwoot
func (im *IntegrationManager) sendMessageToChatwoot(client ports.ChatwootClient, conversationID int, content, messageType string, fromMe bool, ctx context.Context, sessionID, messageID string) (*ports.ChatwootMessage, error) {
	// Format content for Chatwoot
	formattedContent := im.formatContentForChatwoot(content, messageType)

	// Determine message type based on from_me flag
	chatwootMessageType := "incoming" // default for client messages
	if fromMe {
		chatwootMessageType = "outgoing" // messages sent by agent/phone
	}

	// Send message to Chatwoot with correct type
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

// finalizeMessageProcessing updates mapping and logs success
func (im *IntegrationManager) finalizeMessageProcessing(ctx context.Context, sessionID, messageID string, chatwootMessageID, conversationID int) error {
	// Update mapping with Chatwoot IDs
	err := im.messageMapper.UpdateMapping(ctx, sessionID, messageID, chatwootMessageID, conversationID)
	if err != nil {
		im.logger.WarnWithFields("Failed to update mapping", map[string]interface{}{
			"message_id":         messageID,
			"cw_message_id":      chatwootMessageID,
			"cw_conversation_id": conversationID,
			"error":              err.Error(),
		})
		// Don't return error here as the message was sent successfully
	}

	im.logger.InfoWithFields("WhatsApp message processed successfully", map[string]interface{}{
		"session_id":         sessionID,
		"message_id":         messageID,
		"cw_message_id":      chatwootMessageID,
		"cw_conversation_id": conversationID,
	})

	return nil
}

// extractPhoneFromJID extracts phone number from WhatsApp JID and formats to E164
// Following Evolution API logic exactly
func (im *IntegrationManager) extractPhoneFromJID(jid string) string {
	im.logger.DebugWithFields("Extracting phone from JID", map[string]interface{}{
		"original_jid": jid,
	})

	// Step 1: Remove @s.whatsapp.net or @g.us suffix (like Evolution API line 36)
	phone := strings.Split(jid, "@")[0]

	// Step 2: Remove :XX suffix (like Evolution API line 36: number.replace(/:\d+/, ''))
	phone = regexp.MustCompile(`:\d+`).ReplaceAllString(phone, "")

	// Step 3: For group JIDs, extract the creator's phone
	if strings.Contains(phone, "-") {
		parts := strings.Split(phone, "-")
		if len(parts) > 0 {
			phone = parts[0]
		}
	}

	// Step 4: Remove any non-digit characters (like Evolution API line 59)
	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Step 5: Format Brazilian numbers (like Evolution API formatBRNumber)
	phone = im.formatBrazilianPhone(phone)

	// Step 6: Add + prefix for E164 format (required by Chatwoot)
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	im.logger.DebugWithFields("Extracted phone from JID", map[string]interface{}{
		"original_jid": jid,
		"final_phone":  phone,
	})

	return phone
}

// getBrazilianNumbers returns all possible formats for a Brazilian number (like Evolution API)
func (im *IntegrationManager) getBrazilianNumbers(query string) []string {
	numbers := []string{query}

	// Check if it's a Brazilian number with +55
	if strings.HasPrefix(query, "+55") {
		if len(query) == 14 { // +55 XX 9XXXXXXXX (14 digits)
			// Create version without the 9: +55 XX XXXXXXXX (13 digits)
			withoutNine := query[:5] + query[6:]
			numbers = append(numbers, withoutNine)
		} else if len(query) == 13 { // +55 XX XXXXXXXX (13 digits)
			// Create version with the 9: +55 XX 9XXXXXXXX (14 digits)
			withNine := query[:5] + "9" + query[5:]
			numbers = append(numbers, withNine)
		}
	}

	return numbers
}

// mergeBrazilianContacts merges two Brazilian contacts with different formats (like Evolution API)
func (im *IntegrationManager) mergeBrazilianContacts(client ports.ChatwootClient, contacts []*ports.ChatwootContact, sessionID string) (*ports.ChatwootContact, error) {
	if len(contacts) != 2 {
		return nil, fmt.Errorf("expected exactly 2 contacts for merge, got %d", len(contacts))
	}

	// Find the contact with 14 digits (with 9) and 13 digits (without 9)
	var contact14, contact13 *ports.ChatwootContact

	for _, contact := range contacts {
		if len(contact.PhoneNumber) == 14 { // +55 XX 9XXXXXXXX
			contact14 = contact
		} else if len(contact.PhoneNumber) == 13 { // +55 XX XXXXXXXX
			contact13 = contact
		}
	}

	// If we have both formats, merge them (keep the 14-digit as base, like Evolution API)
	if contact14 != nil && contact13 != nil {
		im.logger.InfoWithFields("Merging Brazilian contacts", map[string]interface{}{
			"session_id":       sessionID,
			"base_contact_id":  contact14.ID,
			"base_phone":       contact14.PhoneNumber,
			"merge_contact_id": contact13.ID,
			"merge_phone":      contact13.PhoneNumber,
		})

		// Use the 14-digit contact as base (Evolution API logic)
		err := im.mergeContacts(client, contact14.ID, contact13.ID, sessionID)
		if err != nil {
			im.logger.ErrorWithFields("Failed to merge Brazilian contacts", map[string]interface{}{
				"session_id": sessionID,
				"error":      err.Error(),
			})
			// Return base contact even if merge fails - this is intentional fallback behavior
			// The merge failure is logged but doesn't prevent using the base contact
		}

		return contact14, nil
	}

	// If we don't have both formats, return the first contact
	return contacts[0], nil
}

// mergeContacts merges two contacts in Chatwoot (like Evolution API)
func (im *IntegrationManager) mergeContacts(client ports.ChatwootClient, baseContactID, mergeContactID int, sessionID string) error {
	im.logger.InfoWithFields("Merging contacts using Chatwoot API", map[string]interface{}{
		"session_id":       sessionID,
		"base_contact_id":  baseContactID,
		"merge_contact_id": mergeContactID,
	})

	// Call the actual Chatwoot merge API (same as Evolution API)
	err := client.MergeContacts(baseContactID, mergeContactID)
	if err != nil {
		return fmt.Errorf("failed to merge contacts via Chatwoot API: %w", err)
	}

	return nil
}

// formatBrazilianPhone formats Brazilian phone numbers according to Evolution API logic
func (im *IntegrationManager) formatBrazilianPhone(phone string) string {
	// Check if it's a Brazilian number (starts with 55)
	if !strings.HasPrefix(phone, "55") {
		return phone
	}

	// Remove country code for processing
	localNumber := phone[2:]

	// Brazilian mobile numbers should have 11 digits (including area code)
	if len(localNumber) == 10 {
		// Old format without the 9 - add it for mobile numbers
		areaCode := localNumber[:2]
		number := localNumber[2:]

		// Add 9 for mobile numbers (Evolution API logic)
		formatted := "55" + areaCode + "9" + number

		im.logger.DebugWithFields("Added 9 to Brazilian mobile", map[string]interface{}{
			"original":  phone,
			"formatted": formatted,
		})

		return formatted
	}

	// Return as is if already correct format
	return phone
}

// getOrCreateContact gets or creates a contact in Chatwoot
func (im *IntegrationManager) getOrCreateContact(client ports.ChatwootClient, phoneNumber, sessionID string, inboxID int) (*ports.ChatwootContact, error) {
	// Get all possible Brazilian number formats (like Evolution API)
	phoneNumbers := im.getBrazilianNumbers(phoneNumber)

	// Try to find existing contacts with all possible formats
	var foundContacts []*ports.ChatwootContact
	for _, phone := range phoneNumbers {
		contact, err := client.FindContact(phone, inboxID)
		if err == nil {
			foundContacts = append(foundContacts, contact)
		}
	}

	// If we found contacts, handle them according to Evolution API logic
	if len(foundContacts) > 0 {
		// If we found exactly 2 contacts and it's a Brazilian number, merge them (like Evolution API)
		if len(foundContacts) == 2 && strings.HasPrefix(phoneNumber, "+55") {
			mergedContact, err := im.mergeBrazilianContacts(client, foundContacts, sessionID)
			if err == nil && mergedContact != nil {
				return mergedContact, nil
			}
		}

		// Return the first found contact
		contact := foundContacts[0]
		im.logger.InfoWithFields("Found existing contact", map[string]interface{}{
			"session_id":     sessionID,
			"original_phone": phoneNumber,
			"contact_id":     contact.ID,
			"contacts_found": len(foundContacts),
		})
		return contact, nil
	}

	// Create new contact with the original phone number
	contactName := phoneNumber // Use phone as name initially
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

// getOrCreateConversation gets or creates a conversation in Chatwoot following Evolution API logic
func (im *IntegrationManager) getOrCreateConversation(client ports.ChatwootClient, contactID int, sessionID string, inboxID int) (*ports.ChatwootConversation, error) {
	im.logger.InfoWithFields("Getting or creating conversation", map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
		"session_id": sessionID,
	})

	// Step 1: List all conversations for this contact (like Evolution API line 736-740)
	conversations, err := client.ListContactConversations(contactID)
	if err != nil {
		im.logger.WarnWithFields("Failed to list contact conversations", map[string]interface{}{
			"contact_id": contactID,
			"error":      err.Error(),
		})
		// Continue to create new conversation if listing fails
	} else {
		// Step 2: Find conversation for this inbox that is not resolved (like Evolution API line 747-768)
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

	// Step 3: No active conversation found, create new one (like Evolution API line 794-797)
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

// formatContentForChatwoot formats message content for Chatwoot
func (im *IntegrationManager) formatContentForChatwoot(content, messageType string) string {
	switch messageType {
	case "text":
		// Format markdown for Chatwoot
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

// GetMappingStats returns statistics about message mappings
func (im *IntegrationManager) GetMappingStats(sessionID string) (*MappingStats, error) {
	ctx := context.Background()
	return im.messageMapper.GetMappingStats(ctx, sessionID)
}

// ProcessPendingMessages processes pending message mappings
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
		// For pending mappings, we would need to get the original message data
		// This is a simplified version that just marks them as failed
		// In a real implementation, you'd store more message data or retrieve it from WhatsApp

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

// CleanupOldMappings removes old message mappings
func (im *IntegrationManager) CleanupOldMappings(sessionID string, olderThanDays int) (int, error) {
	ctx := context.Background()
	return im.messageMapper.CleanupOldMappings(ctx, sessionID, olderThanDays)
}
