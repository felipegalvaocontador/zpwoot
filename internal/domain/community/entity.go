package community

import (
	"time"

	"zpwoot/internal/constants"
)

// Community represents a WhatsApp community
type Community struct {
	ID          string    `json:"id"`
	JID         string    `json:"jid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Community metadata
	ParticipantCount int  `json:"participantCount"`
	GroupCount       int  `json:"groupCount"`
	IsOwner          bool `json:"isOwner"`
	IsAdmin          bool `json:"isAdmin"`

	// Community settings
	IsMuted        bool `json:"isMuted"`
	IsAnnouncement bool `json:"isAnnouncement"`

	// Linked groups
	LinkedGroups []*LinkedGroup `json:"linkedGroups,omitempty"`
}

// LinkedGroup represents a group linked to a community
type LinkedGroup struct {
	JID         string `json:"jid"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Group metadata
	ParticipantCount int  `json:"participantCount"`
	IsOwner          bool `json:"isOwner"`
	IsAdmin          bool `json:"isAdmin"`

	// Link metadata
	LinkedAt time.Time `json:"linkedAt"`
}

// CommunityInfo represents basic community information
type CommunityInfo struct {
	ID               string `json:"id"`
	JID              string `json:"jid"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	ParticipantCount int    `json:"participantCount"`
	GroupCount       int    `json:"groupCount"`
	IsOwner          bool   `json:"isOwner"`
	IsAdmin          bool   `json:"isAdmin"`
	IsMuted          bool   `json:"isMuted"`
	IsAnnouncement   bool   `json:"isAnnouncement"`
}

// GroupLinkInfo represents information about linking a group to a community
type GroupLinkInfo struct {
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	LinkedAt     time.Time `json:"linkedAt"`
	Success      bool      `json:"success"`
	Message      string    `json:"message,omitempty"`
}

// CommunityParticipant represents a participant in a community
type CommunityParticipant struct {
	JID      string    `json:"jid"`
	Name     string    `json:"name"`
	IsOwner  bool      `json:"isOwner"`
	IsAdmin  bool      `json:"isAdmin"`
	JoinedAt time.Time `json:"joinedAt"`
}

// CommunitySettings represents community settings
type CommunitySettings struct {
	IsAnnouncement bool `json:"isAnnouncement"`
	IsMuted        bool `json:"isMuted"`

	// Permission settings
	WhoCanAddGroups    string `json:"whoCanAddGroups"`    // "admins", "all"
	WhoCanSendMessages string `json:"whoCanSendMessages"` // "admins", "all"
	WhoCanEditInfo     string `json:"whoCanEditInfo"`     // "admins", "all"
}

// CommunityInviteInfo represents community invite information
type CommunityInviteInfo struct {
	InviteCode string    `json:"inviteCode"`
	InviteLink string    `json:"inviteLink"`
	ExpiresAt  time.Time `json:"expiresAt,omitempty"`
	CreatedBy  string    `json:"createdBy"`
	CreatedAt  time.Time `json:"createdAt"`
}

// CommunityStats represents community statistics
type CommunityStats struct {
	TotalParticipants int `json:"totalParticipants"`
	TotalGroups       int `json:"totalGroups"`
	ActiveGroups      int `json:"activeGroups"`
	RecentActivity    int `json:"recentActivity"`
}

// CommunityEvent represents a community event
type CommunityEvent struct {
	ID           string                 `json:"id"`
	Type         CommunityEventType     `json:"type"`
	CommunityJID string                 `json:"communityJid"`
	ActorJID     string                 `json:"actorJid"`
	TargetJID    string                 `json:"targetJid,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// CommunityEventType represents the type of community event
type CommunityEventType string

const (
	CommunityEventGroupLinked       CommunityEventType = "group_linked"
	CommunityEventGroupUnlinked     CommunityEventType = "group_unlinked"
	CommunityEventParticipantJoined CommunityEventType = "participant_joined"
	CommunityEventParticipantLeft   CommunityEventType = "participant_left"
	CommunityEventSettingsChanged   CommunityEventType = "settings_changed"
	CommunityEventInfoUpdated       CommunityEventType = "info_updated"
)

// Validation methods

// IsValidCommunityJID checks if a JID is a valid community JID
func IsValidCommunityJID(jid string) bool {
	// Community JIDs typically end with @g.us but have specific patterns
	// This is a basic validation - may need to be refined based on WhatsApp's actual format
	return len(jid) > 0 && (jid[len(jid)-len(constants.GroupJIDSuffix):] == constants.GroupJIDSuffix ||
		jid[len(jid)-len(constants.NewsletterJIDSuffix):] == constants.NewsletterJIDSuffix)
}

// IsValidGroupJID checks if a JID is a valid group JID
func IsValidGroupJID(jid string) bool {
	return len(jid) > 0 && jid[len(jid)-len(constants.GroupJIDSuffix):] == constants.GroupJIDSuffix
}

// Helper methods

// GetCommunityID extracts the community ID from a JID
func GetCommunityID(jid string) string {
	if len(jid) > 5 && jid[len(jid)-5:] == "@g.us" {
		return jid[:len(jid)-5]
	}
	return jid
}

// FormatCommunityJID formats a community ID to a full JID
func FormatCommunityJID(id string) string {
	if len(id) > 5 && id[len(id)-5:] == "@g.us" {
		return id
	}
	return id + "@g.us"
}

// GetGroupID extracts the group ID from a JID
func GetGroupID(jid string) string {
	if len(jid) > 5 && jid[len(jid)-5:] == "@g.us" {
		return jid[:len(jid)-5]
	}
	return jid
}

// FormatGroupJID formats a group ID to a full JID
func FormatGroupJID(id string) string {
	if len(id) > 5 && id[len(id)-5:] == "@g.us" {
		return id
	}
	return id + "@g.us"
}
