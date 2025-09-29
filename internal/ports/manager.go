package ports

import (
	"context"
	"time"

	"go.mau.fi/whatsmeow/types"
	"zpwoot/internal/domain/message"
	"zpwoot/internal/domain/session"
)

// WameowManager defines the interface for WhatsApp session management operations
type WameowManager interface {
	CreateSession(sessionID string, config *session.ProxyConfig) error
	ConnectSession(sessionID string) error
	DisconnectSession(sessionID string) error
	LogoutSession(sessionID string) error

	GetQRCode(sessionID string) (*session.QRCodeResponse, error)
	PairPhone(sessionID, phoneNumber string) error
	IsConnected(sessionID string) bool
	GetDeviceInfo(sessionID string) (*session.DeviceInfo, error)

	SetProxy(sessionID string, config *session.ProxyConfig) error
	GetProxy(sessionID string) (*session.ProxyConfig, error)
	GetUserJID(sessionID string) (string, error)

	// Message operations
	SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string, contextInfo *message.ContextInfo) (*message.SendResult, error)
	SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error
	SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error)
	SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error)
	SendReaction(sessionID, to, messageID, reaction string) error
	SendPresence(sessionID, to, presence string) error
	EditMessage(sessionID, to, messageID, newText string) error
	MarkRead(sessionID, to, messageID string) error
	RevokeMessage(sessionID, to, messageID string) (*message.SendResult, error)

	// Contact operations
	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error)
	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error)
	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error)
	GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error)
	GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error)

	// Group management methods
	CreateGroup(sessionID, name string, participants []string, description string) (*GroupInfo, error)
	GetGroupInfo(sessionID, groupJID string) (*GroupInfo, error)
	ListJoinedGroups(sessionID string) ([]*GroupInfo, error)
	UpdateGroupParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error)
	SetGroupName(sessionID, groupJID, name string) error
	SetGroupDescription(sessionID, groupJID, description string) error
	SetGroupPhoto(sessionID, groupJID string, photo []byte) error
	GetGroupInviteLink(sessionID, groupJID string, reset bool) (string, error)
	JoinGroupViaLink(sessionID, inviteLink string) (*GroupInfo, error)
	LeaveGroup(sessionID, groupJID string) error
	UpdateGroupSettings(sessionID, groupJID string, announce, locked *bool) error
	GetGroupRequestParticipants(sessionID, groupJID string) ([]types.GroupParticipantRequest, error)
	UpdateGroupRequestParticipants(sessionID, groupJID string, participants []string, action string) ([]string, []string, error)
	SetGroupJoinApprovalMode(sessionID, groupJID string, requireApproval bool) error
	SetGroupMemberAddMode(sessionID, groupJID string, mode string) error

	// Advanced group methods
	GetGroupInfoFromLink(sessionID string, inviteLink string) (*types.GroupInfo, error)
	GetGroupInfoFromInvite(sessionID string, jid, inviter, code string, expiration int64) (*types.GroupInfo, error)
	JoinGroupWithInvite(sessionID string, jid, inviter, code string, expiration int64) error

	// Session statistics and event handling
	GetSessionStats(sessionID string) (*SessionStats, error)
	RegisterEventHandler(sessionID string, handler EventHandler) error
	UnregisterEventHandler(sessionID string, handlerID string) error
}

// GroupInfo represents information about a WhatsApp group
type GroupInfo struct {
	GroupJID     string             `json:"groupJid"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Owner        string             `json:"owner"`
	Participants []GroupParticipant `json:"participants"`
	Settings     GroupSettings      `json:"settings"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// GroupParticipant represents a participant in a WhatsApp group
type GroupParticipant struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"isAdmin"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

// GroupSettings represents settings for a WhatsApp group
type GroupSettings struct {
	Announce bool `json:"announce"`
	Locked   bool `json:"locked"`
}

// SessionStats represents statistics for a WhatsApp session
type SessionStats struct {
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	LastActivity     int64 `json:"last_activity"`
	Uptime           int64 `json:"uptime"`
}

// EventHandler defines the interface for handling WhatsApp events
type EventHandler interface {
	HandleMessage(sessionID string, message *WameowMessage) error
	HandleConnection(sessionID string, connected bool) error
	HandleQRCode(sessionID string, qrCode string) error
	HandlePairSuccess(sessionID string) error
	HandleError(sessionID string, err error) error
}

// WameowMessage represents a message in the Wameow system
type WameowMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	MediaURL  string `json:"media_url,omitempty"`
	Caption   string `json:"caption,omitempty"`
}

// MessageInfo represents basic information about a message
type MessageInfo struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Chat      string    `json:"chat"`
}
