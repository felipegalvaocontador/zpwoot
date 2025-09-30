package community

import (
	"time"
)

// Community represents a WhatsApp community
type Community struct {
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	JID              string         `json:"jid"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	ID               string         `json:"id"`
	LinkedGroups     []*LinkedGroup `json:"linkedGroups,omitempty"`
	ParticipantCount int            `json:"participantCount"`
	GroupCount       int            `json:"groupCount"`
	IsOwner          bool           `json:"isOwner"`
	IsAdmin          bool           `json:"isAdmin"`
	IsMuted          bool           `json:"isMuted"`
	IsAnnouncement   bool           `json:"isAnnouncement"`
}

// LinkedGroup represents a group linked to a community
type LinkedGroup struct {
	LinkedAt         time.Time `json:"linkedAt"`
	JID              string    `json:"jid"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	ParticipantCount int       `json:"participantCount"`
	IsOwner          bool      `json:"isOwner"`
	IsAdmin          bool      `json:"isAdmin"`
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
	LinkedAt     time.Time `json:"linkedAt"`
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	Message      string    `json:"message,omitempty"`
	Success      bool      `json:"success"`
}

// CommunityParticipant represents a participant in a community
type CommunityParticipant struct {
	JoinedAt time.Time `json:"joinedAt"`
	JID      string    `json:"jid"`
	Name     string    `json:"name"`
	IsOwner  bool      `json:"isOwner"`
	IsAdmin  bool      `json:"isAdmin"`
}

// CommunitySettings represents community settings
type CommunitySettings struct {
	WhoCanAddGroups    string `json:"whoCanAddGroups"`
	WhoCanSendMessages string `json:"whoCanSendMessages"`
	WhoCanEditInfo     string `json:"whoCanEditInfo"`
	IsAnnouncement     bool   `json:"isAnnouncement"`
	IsMuted            bool   `json:"isMuted"`
}

// CommunityInviteInfo represents community invite information
type CommunityInviteInfo struct {
	ExpiresAt  time.Time `json:"expiresAt,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	InviteCode string    `json:"inviteCode"`
	InviteLink string    `json:"inviteLink"`
	CreatedBy  string    `json:"createdBy"`
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
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ID           string                 `json:"id"`
	Type         CommunityEventType     `json:"type"`
	CommunityJID string                 `json:"communityJid"`
	ActorJID     string                 `json:"actorJid"`
	TargetJID    string                 `json:"targetJid,omitempty"`
}

// CommunityEventType represents the type of community event
type CommunityEventType string

const (
	// JID suffixes for WhatsApp entities
	GroupJIDSuffix      = "@g.us"
	NewsletterJIDSuffix = "@newsletter"

	// Community event types
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
	return len(jid) > 0 && (jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix ||
		jid[len(jid)-len(NewsletterJIDSuffix):] == NewsletterJIDSuffix)
}

// IsValidGroupJID checks if a JID is a valid group JID
func IsValidGroupJID(jid string) bool {
	return len(jid) > 0 && jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix
}

// Helper methods

// GetCommunityID extracts the community ID from a JID
func GetCommunityID(jid string) string {
	if len(jid) > len(GroupJIDSuffix) && jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return jid[:len(jid)-len(GroupJIDSuffix)]
	}
	return jid
}

// FormatCommunityJID formats a community ID to a full JID
func FormatCommunityJID(id string) string {
	if len(id) > len(GroupJIDSuffix) && id[len(id)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return id
	}
	return id + GroupJIDSuffix
}

// GetGroupID extracts the group ID from a JID
func GetGroupID(jid string) string {
	if len(jid) > len(GroupJIDSuffix) && jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return jid[:len(jid)-len(GroupJIDSuffix)]
	}
	return jid
}

// FormatGroupJID formats a group ID to a full JID
func FormatGroupJID(id string) string {
	if len(id) > len(GroupJIDSuffix) && id[len(id)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return id
	}
	return id + GroupJIDSuffix
}
