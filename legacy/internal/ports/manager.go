package ports

import (
	"context"
	"time"

	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/domain/message"
	"zpwoot/internal/domain/session"
)

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

	SendMessage(sessionID, to, messageType, body, caption, file, filename string, latitude, longitude float64, contactName, contactPhone string, contextInfo *message.ContextInfo) (*message.SendResult, error)
	SendMediaMessage(sessionID, to string, media []byte, mediaType, caption string) error
	SendButtonMessage(sessionID, to, body string, buttons []map[string]string) (*message.SendResult, error)
	SendListMessage(sessionID, to, body, buttonText string, sections []map[string]interface{}) (*message.SendResult, error)
	SendReaction(sessionID, to, messageID, reaction string) error
	SendPresence(sessionID, to, presence string) error
	EditMessage(sessionID, to, messageID, newText string) error
	MarkRead(sessionID, to, messageID string) error
	RevokeMessage(sessionID, to, messageID string) (*message.SendResult, error)

	IsOnWhatsApp(ctx context.Context, sessionID string, phoneNumbers []string) (map[string]interface{}, error)
	GetProfilePictureInfo(ctx context.Context, sessionID, jid string, preview bool) (map[string]interface{}, error)
	GetUserInfo(ctx context.Context, sessionID string, jids []string) ([]map[string]interface{}, error)
	GetBusinessProfile(ctx context.Context, sessionID, jid string) (map[string]interface{}, error)
	GetAllContacts(ctx context.Context, sessionID string) (map[string]interface{}, error)

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

	GetGroupInfoFromLink(sessionID string, inviteLink string) (*types.GroupInfo, error)
	GetGroupInfoFromInvite(sessionID string, jid, inviter, code string, expiration int64) (*types.GroupInfo, error)
	JoinGroupWithInvite(sessionID string, jid, inviter, code string, expiration int64) error

	GetSessionStats(sessionID string) (*SessionStats, error)
	RegisterEventHandler(sessionID string, handler EventHandler) error
	UnregisterEventHandler(sessionID string, handlerID string) error
}

type GroupInfo struct {
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
	GroupJID     string             `json:"groupJid"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Owner        string             `json:"owner"`
	Participants []GroupParticipant `json:"participants"`
	Settings     GroupSettings      `json:"settings"`
}

type GroupParticipant struct {
	JID          string `json:"jid"`
	IsAdmin      bool   `json:"isAdmin"`
	IsSuperAdmin bool   `json:"isSuperAdmin"`
}

type GroupSettings struct {
	Announce bool `json:"announce"`
	Locked   bool `json:"locked"`
}

type SessionStats struct {
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	LastActivity     int64 `json:"last_activity"`
	Uptime           int64 `json:"uptime"`
}

type EventHandler interface {
	HandleMessage(sessionID string, message *WameowMessage) error
	HandleConnection(sessionID string, connected bool) error
	HandleQRCode(sessionID string, qrCode string) error
	HandlePairSuccess(sessionID string) error
	HandleError(sessionID string, err error) error
}

type WameowMessage struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	MediaURL  string `json:"media_url,omitempty"`
	Caption   string `json:"caption,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type MessageInfo struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Chat      string    `json:"chat"`
}
