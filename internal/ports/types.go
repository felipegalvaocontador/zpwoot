package ports

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// UnixTime is a custom type that can unmarshal both Unix timestamps and RFC3339 strings
type UnixTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for UnixTime
func (ut *UnixTime) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as Unix timestamp (number)
	if len(data) > 0 && data[0] != '"' {
		// Try as float64 first (for timestamps with decimals)
		var timestampFloat float64
		if err := json.Unmarshal(data, &timestampFloat); err == nil {
			seconds := int64(timestampFloat)
			nanoseconds := int64((timestampFloat - float64(seconds)) * 1e9)
			ut.Time = time.Unix(seconds, nanoseconds)
			return nil
		}

		// Fallback to int64
		var timestamp int64
		if err := json.Unmarshal(data, &timestamp); err == nil {
			ut.Time = time.Unix(timestamp, 0)
			return nil
		}
	}

	// Try to unmarshal as string (RFC3339 or other formats)
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("cannot unmarshal %s into UnixTime", data)
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			ut.Time = t
			return nil
		}
	}

	// Try as Unix timestamp string
	if timestamp, err := strconv.ParseInt(str, 10, 64); err == nil {
		ut.Time = time.Unix(timestamp, 0)
		return nil
	}

	return fmt.Errorf("cannot parse %s as time", str)
}

// MarshalJSON implements json.Marshaler for UnixTime
func (ut UnixTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ut.Unix())
}

// Errors
var (
	ErrConfigNotFound  = errors.New("chatwoot config not found")
	ErrContactNotFound = errors.New("chatwoot contact not found")
)

// ChatwootConfig represents the Chatwoot configuration
type ChatwootConfig struct {
	UpdatedAt      time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedAt      time.Time `json:"createdAt" db:"createdAt"`
	InboxName      *string   `json:"inboxName,omitempty" db:"inboxName"`
	Number         *string   `json:"number,omitempty" db:"number"`
	Logo           *string   `json:"logo,omitempty" db:"logo"`
	InboxID        *string   `json:"inboxId,omitempty" db:"inboxId"`
	Organization   *string   `json:"organization,omitempty" db:"organization"`
	SignDelimiter  string    `json:"signDelimiter" db:"signDelimiter"`
	URL            string    `json:"url" db:"url"`
	Token          string    `json:"token" db:"token"`
	AccountID      string    `json:"accountId" db:"accountId"`
	IgnoreJids     []string  `json:"ignoreJids,omitempty" db:"ignoreJids"`
	ImportDays     int       `json:"importDays" db:"importDays"`
	SessionID      uuid.UUID `json:"sessionId" db:"sessionId"`
	ID             uuid.UUID `json:"id" db:"id"`
	ImportMessages bool      `json:"importMessages" db:"importMessages"`
	MergeBrazil    bool      `json:"mergeBrazil" db:"mergeBrazil"`
	Enabled        bool      `json:"enabled" db:"enabled"`
	AutoCreate     bool      `json:"autoCreate" db:"autoCreate"`
	ImportContacts bool      `json:"importContacts" db:"importContacts"`
	ConvPending    bool      `json:"convPending" db:"convPending"`
	ReopenConv     bool      `json:"reopenConv" db:"reopenConv"`
	SignMsg        bool      `json:"signMsg" db:"signMsg"`
}

// ChatwootContact represents a contact in Chatwoot
type ChatwootContact struct {
	CreatedAt            UnixTime               `json:"created_at"`
	UpdatedAt            UnixTime               `json:"updated_at"`
	CustomAttributes     map[string]interface{} `json:"custom_attributes,omitempty"`
	AdditionalAttributes map[string]interface{} `json:"additional_attributes,omitempty"`
	Name                 string                 `json:"name"`
	PhoneNumber          string                 `json:"phone_number"`
	Email                string                 `json:"email,omitempty"`
	Identifier           string                 `json:"identifier,omitempty"`
	ID                   int                    `json:"id"`
}

// ChatwootConversation represents a conversation in Chatwoot
type ChatwootConversation struct {
	CreatedAt UnixTime `json:"created_at"`
	UpdatedAt UnixTime `json:"updated_at"`
	Status    string   `json:"status"`
	ID        int      `json:"id"`
	ContactID int      `json:"contact_id"`
	InboxID   int      `json:"inbox_id"`
}

// ChatwootMessage represents a message in Chatwoot
type ChatwootMessage struct {
	CreatedAt         UnixTime               `json:"created_at"`
	ContentAttributes map[string]interface{} `json:"content_attributes,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	Content           string                 `json:"content"`
	ContentType       string                 `json:"content_type"`
	SourceID          string                 `json:"source_id,omitempty"`
	Attachments       []ChatwootAttachment   `json:"attachments,omitempty"`
	ID                int                    `json:"id"`
	ConversationID    int                    `json:"conversation_id"`
	MessageType       int                    `json:"message_type"`
	Private           bool                   `json:"private"`
}

// ChatwootAttachment represents an attachment in Chatwoot
type ChatwootAttachment struct {
	FileType string `json:"file_type"`
	FileName string `json:"file_name"`
	DataURL  string `json:"data_url"`
	ThumbURL string `json:"thumb_url,omitempty"`
	ID       int    `json:"id"`
	FileSize int    `json:"file_size,omitempty"`
}

// JIDValidator defines the interface for JID validation
type JIDValidator interface {
	IsValid(jid string) bool
	Normalize(jid string) string
	IsValidJID(jid string) bool
	IsNewsletterJID(jid string) bool
	ParseJID(jid string) (string, error)
}
