package chatwoot

import (
	"context"
	"fmt"
	"time"

	chatwootdomain "zpwoot/internal/domain/chatwoot"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type WebhookHandler struct {
	logger        *logger.Logger
	manager       ports.ChatwootManager
	wameowManager ports.WameowManager // For sending messages to WhatsApp
}

func NewWebhookHandler(logger *logger.Logger, manager ports.ChatwootManager, wameowManager ports.WameowManager) *WebhookHandler {
	return &WebhookHandler{
		logger:        logger,
		manager:       manager,
		wameowManager: wameowManager,
	}
}

func (h *WebhookHandler) ProcessWebhook(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	time.Sleep(500 * time.Millisecond)

	if h.isPrivateMessage(webhook) {
		return nil
	}

	if webhook.Event == "message_updated" && !h.isMessageDeleted(webhook) {
		return nil
	}

	if webhook.Event == "conversation_status_changed" {
		return h.handleConversationStatusChanged(webhook, sessionID)
	}

	if webhook.Event == "message_updated" && h.isMessageDeleted(webhook) {
		return h.handleMessageDeleted(webhook, sessionID)
	}

	if webhook.Event == "message_created" {
		return h.handleMessageCreated(ctx, webhook, sessionID)
	}

	h.logger.DebugWithFields("Unhandled webhook event", map[string]interface{}{
		"event":      webhook.Event,
		"session_id": sessionID,
	})

	return nil
}

func (h *WebhookHandler) handleConversationStatusChanged(webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	if webhook.Conversation.Status == "resolved" {
		h.logger.InfoWithFields("Conversation resolved", map[string]interface{}{
			"conversation_id": webhook.Conversation.ID,
			"session_id":      sessionID,
		})

	}

	return nil
}

func (h *WebhookHandler) handleMessageDeleted(webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	h.logger.InfoWithFields("Message deleted", map[string]interface{}{
		"message_id": webhook.Message.ID,
		"session_id": sessionID,
	})

	return nil
}

func (h *WebhookHandler) handleMessageCreated(ctx context.Context, webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	if webhook.Message == nil {
		return fmt.Errorf("message is nil in webhook payload")
	}

	if h.isBotMessage(webhook) {
		h.logger.DebugWithFields("Skipping bot message", map[string]interface{}{
			"message_id": webhook.Message.ID,
			"session_id": sessionID,
		})
		return nil
	}

	if webhook.Message.MessageType != "outgoing" {
		return nil
	}

	return h.sendToWhatsApp(webhook, sessionID)
}

func (h *WebhookHandler) sendToWhatsApp(webhook *chatwootdomain.ChatwootWebhookPayload, sessionID string) error {
	h.logger.InfoWithFields("Sending message to WhatsApp", map[string]interface{}{
		"message_id":      webhook.Message.ID,
		"conversation_id": webhook.Conversation.ID,
		"session_id":      sessionID,
	})

	phoneNumber := webhook.Contact.PhoneNumber
	if phoneNumber == "" {
		return fmt.Errorf("contact phone number is empty")
	}

	content := h.formatContentForWhatsApp(webhook.Message.Content)

	_, err := h.wameowManager.SendMessage(sessionID, phoneNumber, "text", content, "", "", "", 0, 0, "", "", nil)
	if err != nil {
		return fmt.Errorf("failed to send message to WhatsApp: %w", err)
	}

	return nil
}

func (h *WebhookHandler) formatContentForWhatsApp(content string) string {
	return content
}

func (h *WebhookHandler) isPrivateMessage(webhook *chatwootdomain.ChatwootWebhookPayload) bool {
	if webhook.Message == nil {
		return false
	}
	return webhook.Message.Private
}

func (h *WebhookHandler) isMessageDeleted(webhook *chatwootdomain.ChatwootWebhookPayload) bool {
	if webhook.Message == nil {
		return false
	}

	if webhook.Message.ContentAttributes != nil {
		if deleted, exists := webhook.Message.ContentAttributes["deleted"]; exists && deleted != nil {
			return true
		}
	}

	return false
}

func (h *WebhookHandler) isBotMessage(webhook *chatwootdomain.ChatwootWebhookPayload) bool {
	if webhook.Message == nil || webhook.Message.Sender == nil {
		return false
	}

	if webhook.Message.Sender.Type == "agent_bot" {
		return true
	}

	if webhook.Message.Sender.Identifier == "123456" {
		return true
	}

	return false
}
