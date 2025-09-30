package wameow

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/domain/newsletter"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type NewsletterAdapter struct {
	wameowManager ports.WameowManager
	logger        logger.Logger
}

func NewNewsletterAdapter(wameowManager ports.WameowManager, logger logger.Logger) ports.NewsletterManager {
	return &NewsletterAdapter{
		wameowManager: wameowManager,
		logger:        logger,
	}
}

func (na *NewsletterAdapter) CreateNewsletter(ctx context.Context, sessionID string, name, description string) (*newsletter.NewsletterInfo, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	metadata, err := client.CreateNewsletter(ctx, name, description)
	if err != nil {
		return nil, err
	}

	return convertNewsletterMetadata(metadata), nil
}

func (na *NewsletterAdapter) GetNewsletterInfo(ctx context.Context, sessionID string, jid string) (*newsletter.NewsletterInfo, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	metadata, err := client.GetNewsletterInfo(ctx, jid)
	if err != nil {
		return nil, err
	}

	return convertNewsletterMetadata(metadata), nil
}

func (na *NewsletterAdapter) GetNewsletterInfoWithInvite(ctx context.Context, sessionID string, inviteKey string) (*newsletter.NewsletterInfo, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	metadata, err := client.GetNewsletterInfoWithInvite(ctx, inviteKey)
	if err != nil {
		return nil, err
	}

	return convertNewsletterMetadata(metadata), nil
}

func (na *NewsletterAdapter) FollowNewsletter(ctx context.Context, sessionID string, jid string) error {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.FollowNewsletter(ctx, jid)
}

func (na *NewsletterAdapter) UnfollowNewsletter(ctx context.Context, sessionID string, jid string) error {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.UnfollowNewsletter(ctx, jid)
}

func (na *NewsletterAdapter) GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]*newsletter.NewsletterInfo, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	metadataList, err := client.GetSubscribedNewsletters(ctx)
	if err != nil {
		return nil, err
	}

	newsletters := make([]*newsletter.NewsletterInfo, len(metadataList))
	for i, metadata := range metadataList {
		newsletters[i] = convertNewsletterMetadata(metadata)
	}

	return newsletters, nil
}

func (na *NewsletterAdapter) GetNewsletterMessages(ctx context.Context, sessionID string, jid string, count int, before string) ([]*newsletter.NewsletterMessage, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	messages, err := client.GetNewsletterMessages(ctx, jid, count, before)
	if err != nil {
		return nil, err
	}

	domainMessages := make([]*newsletter.NewsletterMessage, len(messages))
	for i, msg := range messages {
		domainMessages[i] = convertNewsletterMessage(msg)
	}

	return domainMessages, nil
}

func (na *NewsletterAdapter) GetNewsletterMessageUpdates(ctx context.Context, sessionID string, jid string, count int, since string, after string) ([]*newsletter.NewsletterMessage, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	updates, err := client.GetNewsletterMessageUpdates(ctx, jid, count, since, after)
	if err != nil {
		return nil, err
	}

	domainUpdates := make([]*newsletter.NewsletterMessage, len(updates))
	for i, update := range updates {
		domainUpdates[i] = convertNewsletterMessage(update)
	}

	return domainUpdates, nil
}

func (na *NewsletterAdapter) NewsletterMarkViewed(ctx context.Context, sessionID string, jid string, serverIDs []string) error {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.NewsletterMarkViewed(ctx, jid, serverIDs)
}

func (na *NewsletterAdapter) NewsletterSendReaction(ctx context.Context, sessionID string, jid string, serverID string, reaction string, messageID string) error {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.NewsletterSendReaction(ctx, jid, serverID, reaction, messageID)
}

func (na *NewsletterAdapter) NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID string, jid string) (int64, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return 0, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return 0, fmt.Errorf("session %s not found", sessionID)
	}

	return client.NewsletterSubscribeLiveUpdates(ctx, jid)
}

func (na *NewsletterAdapter) NewsletterToggleMute(ctx context.Context, sessionID string, jid string, mute bool) error {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.NewsletterToggleMute(ctx, jid, mute)
}

func (na *NewsletterAdapter) AcceptTOSNotice(ctx context.Context, sessionID string, noticeID string, stage string) error {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return client.AcceptTOSNotice(ctx, noticeID, stage)
}

func (na *NewsletterAdapter) UploadNewsletter(ctx context.Context, sessionID string, data []byte, mimeType string, mediaType string) (*newsletter.UploadNewsletterResponse, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	uploaded, err := client.UploadNewsletter(ctx, data, mimeType, mediaType)
	if err != nil {
		return nil, err
	}

	return convertUploadResponse(uploaded), nil
}

func (na *NewsletterAdapter) UploadNewsletterReader(ctx context.Context, sessionID string, data []byte, mimeType string, mediaType string) (*newsletter.UploadNewsletterResponse, error) {
	manager, ok := na.wameowManager.(*Manager)
	if !ok {
		return nil, fmt.Errorf("wameow manager is not a Manager instance")
	}

	client := manager.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	uploaded, err := client.UploadNewsletterReader(ctx, data, mimeType, mediaType)
	if err != nil {
		return nil, err
	}

	return convertUploadResponse(uploaded), nil
}

func convertUploadResponse(uploaded *whatsmeow.UploadResponse) *newsletter.UploadNewsletterResponse {
	if uploaded == nil {
		return nil
	}

	return &newsletter.UploadNewsletterResponse{
		URL:        uploaded.URL,
		DirectPath: uploaded.DirectPath,
		Handle:     uploaded.Handle,
		ObjectID:   uploaded.ObjectID,
		FileSHA256: fmt.Sprintf("%x", uploaded.FileSHA256),
		FileLength: uploaded.FileLength,
	}
}

func convertNewsletterMessage(msg *types.NewsletterMessage) *newsletter.NewsletterMessage {
	if msg == nil {
		return nil
	}

	domainMsg := &newsletter.NewsletterMessage{
		ID:        msg.MessageID,
		ServerID:  fmt.Sprintf("%d", msg.MessageServerID), // Convert numeric ID to string properly
		FromJID:   "",                                     // Newsletter messages don't have a specific sender JID
		Timestamp: msg.Timestamp,
		Type:      msg.Type,
	}

	if msg.Message != nil && msg.Message.GetConversation() != "" {
		domainMsg.Body = msg.Message.GetConversation()
	}

	domainMsg.ViewsCount = msg.ViewsCount

	if len(msg.ReactionCounts) > 0 {
		reactions := make([]string, 0, len(msg.ReactionCounts))
		for reaction := range msg.ReactionCounts {
			reactions = append(reactions, reaction)
		}
		domainMsg.Reactions = reactions
	}

	return domainMsg
}

func convertNewsletterMetadata(metadata *types.NewsletterMetadata) *newsletter.NewsletterInfo {
	if metadata == nil {
		return nil
	}

	info := &newsletter.NewsletterInfo{
		ID:              metadata.ID.String(),
		CreationTime:    metadata.ThreadMeta.CreationTime.Time,
		UpdateTime:      time.Now(), // WhatsApp doesn't provide update time, use current time
		SubscriberCount: metadata.ThreadMeta.SubscriberCount,
	}

	info.Name = metadata.ThreadMeta.Name.Text

	info.Description = metadata.ThreadMeta.Description.Text

	info.InviteCode = metadata.ThreadMeta.InviteCode

	switch metadata.State.Type {
	case types.NewsletterStateActive:
		info.State = newsletter.NewsletterStateActive
	case types.NewsletterStateSuspended:
		info.State = newsletter.NewsletterStateSuspended
	case types.NewsletterStateGeoSuspended:
		info.State = newsletter.NewsletterStateGeoSuspended
	default:
		info.State = newsletter.NewsletterStateActive
	}

	if metadata.ViewerMeta != nil {
		switch metadata.ViewerMeta.Role {
		case types.NewsletterRoleSubscriber:
			info.Role = newsletter.NewsletterRoleSubscriber
		case types.NewsletterRoleGuest:
			info.Role = newsletter.NewsletterRoleGuest
		case types.NewsletterRoleAdmin:
			info.Role = newsletter.NewsletterRoleAdmin
		case types.NewsletterRoleOwner:
			info.Role = newsletter.NewsletterRoleOwner
		default:
			info.Role = newsletter.NewsletterRoleGuest
		}

		switch metadata.ViewerMeta.Mute {
		case types.NewsletterMuteOn:
			info.Muted = true
			info.MuteState = newsletter.NewsletterMuteOn
		case types.NewsletterMuteOff:
			info.Muted = false
			info.MuteState = newsletter.NewsletterMuteOff
		default:
			info.Muted = false
			info.MuteState = newsletter.NewsletterMuteOff
		}
	} else {
		info.Role = newsletter.NewsletterRoleGuest
		info.Muted = false
		info.MuteState = newsletter.NewsletterMuteOff
	}

	if metadata.ThreadMeta.VerificationState == types.NewsletterVerificationStateVerified {
		info.Verified = true
		info.VerificationState = newsletter.NewsletterVerificationStateVerified
	} else {
		info.Verified = false
		info.VerificationState = newsletter.NewsletterVerificationStateUnverified
	}

	if metadata.ThreadMeta.Picture != nil {
		info.Picture = &newsletter.ProfilePictureInfo{
			URL:    metadata.ThreadMeta.Picture.URL,
			ID:     metadata.ThreadMeta.Picture.ID,
			Type:   metadata.ThreadMeta.Picture.Type,
			Direct: metadata.ThreadMeta.Picture.DirectPath,
		}
	}

	info.Preview = &newsletter.ProfilePictureInfo{
		URL:    metadata.ThreadMeta.Preview.URL,
		ID:     metadata.ThreadMeta.Preview.ID,
		Type:   metadata.ThreadMeta.Preview.Type,
		Direct: metadata.ThreadMeta.Preview.DirectPath,
	}

	return info
}
