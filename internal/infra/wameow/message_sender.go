// Refactored: extracted message sending logic; reduced duplication; improved error handling
package wameow

import (
	"context"
	"fmt"
	"os"

	appMessage "zpwoot/internal/app/message"
	"zpwoot/platform/logger"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

// messageSender implements MessageSender interface
type messageSender struct {
	client    *whatsmeow.Client
	logger    *logger.Logger
	validator *JIDValidator
}

// NewMessageSender creates a new message sender
func NewMessageSender(client *whatsmeow.Client, logger *logger.Logger) MessageSender {
	return &messageSender{
		client:    client,
		logger:    logger,
		validator: NewJIDValidator(),
	}
}

// SendText sends a text message with optional context info
func (ms *messageSender) SendText(ctx context.Context, to, body string, contextInfo *appMessage.ContextInfo) (*whatsmeow.SendResponse, error) {
	if !ms.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := ms.validator.Parse(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	message := &waE2E.Message{
		Conversation: &body,
	}

	if contextInfo != nil {
		message.ExtendedTextMessage = &waE2E.ExtendedTextMessage{
			Text:        &body,
			ContextInfo: ms.createContextInfo(contextInfo),
		}
		message.Conversation = nil
	}

	ms.logger.InfoWithFields("Sending text message", map[string]interface{}{
		"to":        to,
		"body_len":  len(body),
		"has_reply": contextInfo != nil,
	})

	resp, err := ms.client.SendMessage(ctx, jid, message)
	if err != nil {
		ms.logger.ErrorWithFields("Failed to send text message", map[string]interface{}{
			"to":    to,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	ms.logger.InfoWithFields("Text message sent successfully", map[string]interface{}{
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendMedia sends a media message with the specified type and options
func (ms *messageSender) SendMedia(ctx context.Context, to, filePath string, mediaType MediaType, options MediaOptions) (*whatsmeow.SendResponse, error) {
	if !ms.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := ms.validator.Parse(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	whatsmeowMediaType := ms.convertMediaType(mediaType)
	uploaded, err := ms.client.Upload(ctx, data, whatsmeowMediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}

	message := ms.createMediaMessage(mediaType, uploaded, options)

	ms.logger.InfoWithFields("Sending media message", map[string]interface{}{
		"to":        to,
		"type":      mediaType,
		"file_size": len(data),
		"has_reply": options.ContextInfo != nil,
	})

	resp, err := ms.client.SendMessage(ctx, jid, message)
	if err != nil {
		ms.logger.ErrorWithFields("Failed to send media message", map[string]interface{}{
			"to":    to,
			"type":  mediaType,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to send media message: %w", err)
	}

	ms.logger.InfoWithFields("Media message sent successfully", map[string]interface{}{
		"to":         to,
		"type":       mediaType,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendContact sends a contact message
func (ms *messageSender) SendContact(ctx context.Context, to string, contact ContactInfo) (*whatsmeow.SendResponse, error) {
	if !ms.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := ms.validator.Parse(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", contact.Name, contact.Phone)

	message := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		},
	}

	ms.logger.InfoWithFields("Sending contact message", map[string]interface{}{
		"to":           to,
		"contact_name": contact.Name,
	})

	resp, err := ms.client.SendMessage(ctx, jid, message)
	if err != nil {
		ms.logger.ErrorWithFields("Failed to send contact message", map[string]interface{}{
			"to":    to,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to send contact message: %w", err)
	}

	ms.logger.InfoWithFields("Contact message sent successfully", map[string]interface{}{
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// SendLocation sends a location message
func (ms *messageSender) SendLocation(ctx context.Context, to string, lat, lng float64, address string) (*whatsmeow.SendResponse, error) {
	if !ms.client.IsLoggedIn() {
		return nil, fmt.Errorf("client is not logged in")
	}

	jid, err := ms.validator.Parse(to)
	if err != nil {
		return nil, fmt.Errorf("invalid JID: %w", err)
	}

	message := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  &lat,
			DegreesLongitude: &lng,
			Name:             &address,
		},
	}

	ms.logger.InfoWithFields("Sending location message", map[string]interface{}{
		"to":        to,
		"latitude":  lat,
		"longitude": lng,
		"address":   address,
	})

	resp, err := ms.client.SendMessage(ctx, jid, message)
	if err != nil {
		ms.logger.ErrorWithFields("Failed to send location message", map[string]interface{}{
			"to":    to,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to send location message: %w", err)
	}

	ms.logger.InfoWithFields("Location message sent successfully", map[string]interface{}{
		"to":         to,
		"message_id": resp.ID,
	})

	return &resp, nil
}

// createContextInfo creates WhatsApp ContextInfo from app ContextInfo
func (ms *messageSender) createContextInfo(contextInfo *appMessage.ContextInfo) *waE2E.ContextInfo {
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

// convertMediaType converts our MediaType to whatsmeow MediaType
func (ms *messageSender) convertMediaType(mediaType MediaType) whatsmeow.MediaType {
	switch mediaType {
	case MediaTypeImage:
		return whatsmeow.MediaImage
	case MediaTypeAudio:
		return whatsmeow.MediaAudio
	case MediaTypeVideo:
		return whatsmeow.MediaVideo
	case MediaTypeDocument:
		return whatsmeow.MediaDocument
	default:
		return whatsmeow.MediaImage
	}
}

// createMediaMessage creates the appropriate media message based on type
func (ms *messageSender) createMediaMessage(mediaType MediaType, uploaded whatsmeow.UploadResponse, options MediaOptions) *waE2E.Message {
	contextInfo := ms.createContextInfo(options.ContextInfo)

	switch mediaType {
	case MediaTypeImage:
		return ms.createImageMessage(uploaded, options, contextInfo)
	case MediaTypeAudio:
		return ms.createAudioMessage(uploaded, options, contextInfo)
	case MediaTypeVideo:
		return ms.createVideoMessage(uploaded, options, contextInfo)
	case MediaTypeDocument:
		return ms.createDocumentMessage(uploaded, options, contextInfo)
	case MediaTypeSticker:
		return ms.createStickerMessage(uploaded, options)
	default:
		return ms.createImageMessage(uploaded, options, contextInfo)
	}
}

// createImageMessage creates an image message
func (ms *messageSender) createImageMessage(uploaded whatsmeow.UploadResponse, options MediaOptions, contextInfo *waE2E.ContextInfo) *waE2E.Message {
	mimetype := options.MimeType
	if mimetype == "" {
		mimetype = "image/jpeg"
	}

	return &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       &options.Caption,
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
}

// createAudioMessage creates an audio message
func (ms *messageSender) createAudioMessage(uploaded whatsmeow.UploadResponse, options MediaOptions, contextInfo *waE2E.ContextInfo) *waE2E.Message {
	mimetype := options.MimeType
	if mimetype == "" {
		mimetype = "audio/ogg; codecs=opus"
	}

	return &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
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
}

// createVideoMessage creates a video message
func (ms *messageSender) createVideoMessage(uploaded whatsmeow.UploadResponse, options MediaOptions, contextInfo *waE2E.ContextInfo) *waE2E.Message {
	mimetype := options.MimeType
	if mimetype == "" {
		mimetype = "video/mp4"
	}

	return &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption:       &options.Caption,
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
}

// createDocumentMessage creates a document message
func (ms *messageSender) createDocumentMessage(uploaded whatsmeow.UploadResponse, options MediaOptions, contextInfo *waE2E.ContextInfo) *waE2E.Message {
	mimetype := options.MimeType
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}

	filename := options.Filename
	if filename == "" {
		filename = "document"
	}

	documentMessage := &waE2E.DocumentMessage{
		Title:         &filename,
		FileName:      &filename,
		URL:           &uploaded.URL,
		DirectPath:    &uploaded.DirectPath,
		MediaKey:      uploaded.MediaKey,
		Mimetype:      &mimetype,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    &uploaded.FileLength,
		ContextInfo:   contextInfo,
	}

	// Add caption if provided
	if options.Caption != "" {
		documentMessage.Caption = &options.Caption
	}

	return &waE2E.Message{
		DocumentMessage: documentMessage,
	}
}

// createStickerMessage creates a sticker message
func (ms *messageSender) createStickerMessage(uploaded whatsmeow.UploadResponse, options MediaOptions) *waE2E.Message {
	mimetype := options.MimeType
	if mimetype == "" {
		mimetype = "image/webp"
	}

	return &waE2E.Message{
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
}
