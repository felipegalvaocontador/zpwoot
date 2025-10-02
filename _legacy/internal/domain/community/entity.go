package community

import (
	"time"
)

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

type LinkedGroup struct {
	LinkedAt         time.Time `json:"linkedAt"`
	JID              string    `json:"jid"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	ParticipantCount int       `json:"participantCount"`
	IsOwner          bool      `json:"isOwner"`
	IsAdmin          bool      `json:"isAdmin"`
}

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

type GroupLinkInfo struct {
	LinkedAt     time.Time `json:"linkedAt"`
	CommunityJID string    `json:"communityJid"`
	GroupJID     string    `json:"groupJid"`
	Message      string    `json:"message,omitempty"`
	Success      bool      `json:"success"`
}

type CommunityParticipant struct {
	JoinedAt time.Time `json:"joinedAt"`
	JID      string    `json:"jid"`
	Name     string    `json:"name"`
	IsOwner  bool      `json:"isOwner"`
	IsAdmin  bool      `json:"isAdmin"`
}

type CommunitySettings struct {
	WhoCanAddGroups    string `json:"whoCanAddGroups"`
	WhoCanSendMessages string `json:"whoCanSendMessages"`
	WhoCanEditInfo     string `json:"whoCanEditInfo"`
	IsAnnouncement     bool   `json:"isAnnouncement"`
	IsMuted            bool   `json:"isMuted"`
}

type CommunityInviteInfo struct {
	ExpiresAt  time.Time `json:"expiresAt,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	InviteCode string    `json:"inviteCode"`
	InviteLink string    `json:"inviteLink"`
	CreatedBy  string    `json:"createdBy"`
}

type CommunityStats struct {
	TotalParticipants int `json:"totalParticipants"`
	TotalGroups       int `json:"totalGroups"`
	ActiveGroups      int `json:"activeGroups"`
	RecentActivity    int `json:"recentActivity"`
}

type CommunityEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ID           string                 `json:"id"`
	Type         CommunityEventType     `json:"type"`
	CommunityJID string                 `json:"communityJid"`
	ActorJID     string                 `json:"actorJid"`
	TargetJID    string                 `json:"targetJid,omitempty"`
}

type CommunityEventType string

const (
	GroupJIDSuffix      = "@g.us"
	NewsletterJIDSuffix = "@newsletter"

	CommunityEventGroupLinked       CommunityEventType = "group_linked"
	CommunityEventGroupUnlinked     CommunityEventType = "group_unlinked"
	CommunityEventParticipantJoined CommunityEventType = "participant_joined"
	CommunityEventParticipantLeft   CommunityEventType = "participant_left"
	CommunityEventSettingsChanged   CommunityEventType = "settings_changed"
	CommunityEventInfoUpdated       CommunityEventType = "info_updated"
)

func IsValidCommunityJID(jid string) bool {
	return jid != "" && (jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix ||
		jid[len(jid)-len(NewsletterJIDSuffix):] == NewsletterJIDSuffix)
}

func IsValidGroupJID(jid string) bool {
	return jid != "" && jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix
}

func GetCommunityID(jid string) string {
	if len(jid) > len(GroupJIDSuffix) && jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return jid[:len(jid)-len(GroupJIDSuffix)]
	}
	return jid
}

func FormatCommunityJID(id string) string {
	if len(id) > len(GroupJIDSuffix) && id[len(id)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return id
	}
	return id + GroupJIDSuffix
}

func GetGroupID(jid string) string {
	if len(jid) > len(GroupJIDSuffix) && jid[len(jid)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return jid[:len(jid)-len(GroupJIDSuffix)]
	}
	return jid
}

func FormatGroupJID(id string) string {
	if len(id) > len(GroupJIDSuffix) && id[len(id)-len(GroupJIDSuffix):] == GroupJIDSuffix {
		return id
	}
	return id + GroupJIDSuffix
}
