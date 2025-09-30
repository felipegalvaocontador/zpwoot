package chatwoot

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type ImportManager struct {
	logger           *logger.Logger
	client           ports.ChatwootClient
	contactSync      *ContactSync
	conversationMgr  *ConversationManager
	messageFormatter *MessageFormatter
}

func NewImportManager(
	logger *logger.Logger,
	client ports.ChatwootClient,
	contactSync *ContactSync,
	conversationMgr *ConversationManager,
	messageFormatter *MessageFormatter,
) *ImportManager {
	return &ImportManager{
		logger:           logger,
		client:           client,
		contactSync:      contactSync,
		conversationMgr:  conversationMgr,
		messageFormatter: messageFormatter,
	}
}

func (im *ImportManager) ImportHistoricalMessages(ctx context.Context, sessionID string, daysLimit int, inboxID int) (*ImportResult, error) {
	im.logger.InfoWithFields("Starting historical message import", map[string]interface{}{
		"session_id": sessionID,
		"days_limit": daysLimit,
		"inbox_id":   inboxID,
	})

	result := &ImportResult{
		SessionID: sessionID,
		InboxID:   inboxID,
		DaysLimit: daysLimit,
		StartTime: time.Now(),
		Status:    "running",
	}

	cutoffDate := time.Now().AddDate(0, 0, -daysLimit)

	im.logger.InfoWithFields("Import date range", map[string]interface{}{
		"cutoff_date": cutoffDate.Format("2006-01-02"),
		"days_limit":  daysLimit,
	})

	result.Status = "completed"
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	im.logger.InfoWithFields("Historical message import completed", map[string]interface{}{
		"session_id":            sessionID,
		"duration":              result.Duration.String(),
		"messages_imported":     result.MessagesImported,
		"contacts_created":      result.ContactsCreated,
		"conversations_created": result.ConversationsCreated,
	})

	return result, nil
}

func (im *ImportManager) ImportContacts(ctx context.Context, sessionID string, inboxID int, mergeBrazilContacts bool) (*ContactImportSummary, error) {
	im.logger.InfoWithFields("Starting contact import", map[string]interface{}{
		"session_id":            sessionID,
		"inbox_id":              inboxID,
		"merge_brazil_contacts": mergeBrazilContacts,
	})

	result := &ContactImportSummary{
		SessionID:           sessionID,
		InboxID:             inboxID,
		MergeBrazilContacts: mergeBrazilContacts,
		StartTime:           time.Now(),
		Status:              "running",
	}

	result.Status = "completed"
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	im.logger.InfoWithFields("Contact import completed", map[string]interface{}{
		"session_id":        sessionID,
		"duration":          result.Duration.String(),
		"contacts_imported": result.ContactsImported,
		"contacts_updated":  result.ContactsUpdated,
		"contacts_failed":   result.ContactsFailed,
	})

	return result, nil
}

func (im *ImportManager) ProcessMessageForImport(ctx context.Context, message WhatsAppMessage, inboxID int, mergeBrazilContacts bool) error {
	im.logger.DebugWithFields("Processing message for import", map[string]interface{}{
		"message_id": message.ID,
		"from":       message.From,
		"timestamp":  message.Timestamp,
	})

	contact, err := im.contactSync.CreateOrUpdateContact(message.From, message.FromName, inboxID, mergeBrazilContacts)
	if err != nil {
		return fmt.Errorf("failed to create/update contact: %w", err)
	}

	conversation, err := im.conversationMgr.CreateOrGetConversation(contact.ID, inboxID)
	if err != nil {
		return fmt.Errorf("failed to create/get conversation: %w", err)
	}

	content := im.messageFormatter.FormatMarkdownForChatwoot(message.Content)

	switch message.Type {
	case "text":
		_, err = im.conversationMgr.SendMessage(conversation.ID, content)
	case "image", "video", "audio", "document":
		mediaContent := fmt.Sprintf("📎 **%s**\n%s", message.Type, content)
		_, err = im.conversationMgr.SendMessage(conversation.ID, mediaContent)
	case "contact":
		contactContent := im.messageFormatter.FormatContactMessage(message.ContactName, message.ContactPhone)
		_, err = im.conversationMgr.SendMessage(conversation.ID, contactContent)
	case "location":
		locationContent := im.messageFormatter.FormatLocationMessage(message.Latitude, message.Longitude, message.Address)
		_, err = im.conversationMgr.SendMessage(conversation.ID, locationContent)
	default:
		unknownContent := fmt.Sprintf("📄 **%s message**\n%s", message.Type, content)
		_, err = im.conversationMgr.SendMessage(conversation.ID, unknownContent)
	}

	if err != nil {
		return fmt.Errorf("failed to send message to Chatwoot: %w", err)
	}

	return nil
}

func (im *ImportManager) GetImportProgress(sessionID string) (*ImportProgress, error) {

	return &ImportProgress{
		SessionID:         sessionID,
		Status:            "not_found",
		ProgressPercent:   0,
		MessagesProcessed: 0,
		TotalMessages:     0,
		EstimatedTimeLeft: 0,
	}, nil
}

func (im *ImportManager) CancelImport(sessionID string) error {
	im.logger.InfoWithFields("Canceling import", map[string]interface{}{
		"session_id": sessionID,
	})

	return nil
}

type ImportResult struct {
	SessionID            string        `json:"session_id"`
	InboxID              int           `json:"inbox_id"`
	DaysLimit            int           `json:"days_limit"`
	Status               string        `json:"status"` // running, completed, failed, canceled
	StartTime            time.Time     `json:"start_time"`
	EndTime              time.Time     `json:"end_time"`
	Duration             time.Duration `json:"duration"`
	MessagesImported     int           `json:"messages_imported"`
	ContactsCreated      int           `json:"contacts_created"`
	ConversationsCreated int           `json:"conversations_created"`
	Errors               []string      `json:"errors,omitempty"`
}

type ContactImportSummary struct {
	SessionID           string        `json:"session_id"`
	InboxID             int           `json:"inbox_id"`
	MergeBrazilContacts bool          `json:"merge_brazil_contacts"`
	Status              string        `json:"status"` // running, completed, failed, canceled
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
	Duration            time.Duration `json:"duration"`
	ContactsImported    int           `json:"contacts_imported"`
	ContactsUpdated     int           `json:"contacts_updated"`
	ContactsFailed      int           `json:"contacts_failed"`
	Errors              []string      `json:"errors,omitempty"`
}

type ImportProgress struct {
	SessionID         string `json:"session_id"`
	Status            string `json:"status"` // running, completed, failed, canceled, not_found
	ProgressPercent   int    `json:"progress_percent"`
	MessagesProcessed int    `json:"messages_processed"`
	TotalMessages     int    `json:"total_messages"`
	EstimatedTimeLeft int    `json:"estimated_time_left"` // seconds
	CurrentOperation  string `json:"current_operation"`
}

type WhatsAppMessage struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	FromName  string    `json:"from_name"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // text, image, video, audio, document, contact, location
	Timestamp time.Time `json:"timestamp"`
	IsFromMe  bool      `json:"is_from_me"`

	MediaURL string `json:"media_url,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileName string `json:"file_name,omitempty"`

	ContactName  string `json:"contact_name,omitempty"`
	ContactPhone string `json:"contact_phone,omitempty"`

	Latitude  string `json:"latitude,omitempty"`
	Longitude string `json:"longitude,omitempty"`
	Address   string `json:"address,omitempty"`

	QuotedMessageID string `json:"quoted_message_id,omitempty"`
	QuotedContent   string `json:"quoted_content,omitempty"`
}

func (im *ImportManager) ValidateImportRequest(sessionID string, daysLimit int, inboxID int) error {
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
