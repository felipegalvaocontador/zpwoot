package contracts

import (
	"time"
)

type CreateMessageRequest struct {
	ZpMessageID string `json:"zp_message_id" validate:"required" example:"3EB0C767D71D"`
	ZpSender    string `json:"zp_sender" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ZpChat      string `json:"zp_chat" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ZpTimestamp string `json:"zp_timestamp" validate:"required" example:"2024-01-01T12:00:00Z"`
	ZpFromMe    bool   `json:"zp_from_me" example:"false"`
	ZpType      string `json:"zp_type" validate:"required" example:"text"`
	Content     string `json:"content,omitempty" example:"Hello World"`
} // @name CreateMessageRequest

type ListMessagesRequest struct {
	PaginationRequest
	ChatJID     string `json:"chat_jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	MessageType string `json:"message_type,omitempty" example:"text"`
	FromMe      *bool  `json:"from_me,omitempty" example:"false"`
	DateFrom    string `json:"date_from,omitempty" example:"2024-01-01"`
	DateTo      string `json:"date_to,omitempty" example:"2024-01-31"`
} // @name ListMessagesRequest

type SendTextMessageRequest struct {
	RemoteJID   string       `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body        string       `json:"body" validate:"required" example:"Hello, World!"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} // @name SendTextMessageRequest

type ContextInfo struct {
	StanzaID    string `json:"stanzaId" validate:"required" example:"ABCD1234abcd"`
	Participant string `json:"participant,omitempty" example:"5511999999999@s.whatsapp.net"`
} // @name ContextInfo

type SendMediaMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MediaURL string `json:"media_url" validate:"required,url" example:"https://example.com/image.jpg"`
	Type     string `json:"type" validate:"required,oneof=image audio video document" example:"image"`
	Caption  string `json:"caption,omitempty" example:"Check this out!"`
	Filename string `json:"filename,omitempty" example:"image.jpg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendMediaMessageRequest

type UpdateSyncStatusRequest struct {
	SyncStatus       string `json:"sync_status" validate:"required,oneof=pending synced failed" example:"synced"`
	CwMessageID      *int   `json:"cw_message_id,omitempty" example:"123"`
	CwConversationID *int   `json:"cw_conversation_id,omitempty" example:"456"`
} // @name UpdateSyncStatusRequest

type SendImageMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_image_data"`
	Caption  string `json:"caption,omitempty" example:"Check this image!"`
	Filename string `json:"filename,omitempty" example:"image.jpg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendImageMessageRequest

type SendAudioMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_audio_data"`
	Caption  string `json:"caption,omitempty" example:"Audio message"`
	Filename string `json:"filename,omitempty" example:"audio.mp3"`
	MimeType string `json:"mime_type,omitempty" example:"audio/mpeg"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendAudioMessageRequest

type SendVideoMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_video_data"`
	Caption  string `json:"caption,omitempty" example:"Check this video!"`
	Filename string `json:"filename,omitempty" example:"video.mp4"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendVideoMessageRequest

type SendDocumentMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_document_data"`
	Caption  string `json:"caption,omitempty" example:"Document"`
	Filename string `json:"filename" validate:"required" example:"document.pdf"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendDocumentMessageRequest

type SendStickerMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File     string `json:"file" validate:"required" example:"base64_sticker_data"`
	MimeType string `json:"mime_type,omitempty" example:"image/webp"`
	ReplyTo  string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendStickerMessageRequest

type SendLocationMessageRequest struct {
	To        string  `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Latitude  float64 `json:"latitude" validate:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" validate:"required" example:"-46.6333"`
	Name      string  `json:"name,omitempty" example:"São Paulo"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP, Brazil"`
	ReplyTo   string  `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendLocationMessageRequest

type SendContactMessageRequest struct {
	To           string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Name         string `json:"name" validate:"required" example:"John Doe"`
	Phone        string `json:"phone" validate:"required" example:"+5511888888888"`
	ContactName  string `json:"contact_name,omitempty" example:"John Doe"`
	ContactPhone string `json:"contact_phone,omitempty" example:"+5511888888888"`
	ReplyTo      string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendContactMessageRequest

type CreateMessageResponse struct {
	BaseResponse
	Message *MessageInfo `json:"message"`
} // @name CreateMessageResponse

type ListMessagesResponse struct {
	ListResponse
	Messages []MessageInfo `json:"messages"`
} // @name ListMessagesResponse

type SendMessageResponse struct {
	BaseResponse
	To          string     `json:"to" example:"5511999999999@s.whatsapp.net"`
	MessageID   string     `json:"message_id" example:"3EB0C767D71D"`
	Timestamp   time.Time  `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Status      string     `json:"status" example:"sent"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty" example:"2024-01-01T12:00:05Z"`
	ReadAt      *time.Time `json:"read_at,omitempty" example:"2024-01-01T12:00:10Z"`
} // @name SendMessageResponse

type MessageInfo struct {
	ID               string     `json:"id" example:"1b2e424c-a2a0-41a4-b992-15b7ec06b9bc"`
	SessionID        string     `json:"session_id" example:"session-123"`
	ZpMessageID      string     `json:"zp_message_id" example:"3EB0C767D71D"`
	ZpSender         string     `json:"zp_sender" example:"5511999999999@s.whatsapp.net"`
	ZpChat           string     `json:"zp_chat" example:"5511999999999@s.whatsapp.net"`
	ZpTimestamp      time.Time  `json:"zp_timestamp" example:"2024-01-01T12:00:00Z"`
	ZpFromMe         bool       `json:"zp_from_me" example:"false"`
	ZpType           string     `json:"zp_type" example:"text"`
	Content          string     `json:"content,omitempty" example:"Hello World"`
	MediaURL         string     `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	MediaType        string     `json:"media_type,omitempty" example:"image"`
	CwMessageID      *int       `json:"cw_message_id,omitempty" example:"123"`
	CwConversationID *int       `json:"cw_conversation_id,omitempty" example:"456"`
	SyncStatus       string     `json:"sync_status" example:"synced"`
	SyncedAt         *time.Time `json:"synced_at,omitempty" example:"2024-01-01T12:00:05Z"`
	CreatedAt        time.Time  `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" example:"2024-01-01T12:00:00Z"`
} // @name MessageInfo

type MessageDTO = MessageInfo

type MessageStats struct {
	TotalMessages     int64            `json:"total_messages" example:"1000"`
	MessagesByType    map[string]int64 `json:"messages_by_type"`
	MessagesByStatus  map[string]int64 `json:"messages_by_status"`
	SyncedMessages    int64            `json:"synced_messages" example:"950"`
	PendingMessages   int64            `json:"pending_messages" example:"30"`
	FailedMessages    int64            `json:"failed_messages" example:"20"`
	MessagesToday     int64            `json:"messages_today" example:"50"`
	MessagesThisWeek  int64            `json:"messages_this_week" example:"300"`
	MessagesThisMonth int64            `json:"messages_this_month" example:"1000"`
	AveragePerDay     float64          `json:"average_per_day" example:"33.3"`
	PeakHour          int              `json:"peak_hour" example:"14"`
} // @name MessageStats

type SendContactListMessageRequest struct {
	To       string        `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Contacts []ContactInfo `json:"contacts" validate:"required,min=1"`
	ReplyTo  string        `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendContactListMessageRequest

type ContactInfo struct {
	Name  string `json:"name" validate:"required" example:"John Doe"`
	Phone string `json:"phone" validate:"required" example:"+5511888888888"`
} // @name ContactInfo

type ContactResult struct {
	Name        string `json:"name" example:"John Doe"`
	Phone       string `json:"phone" example:"+5511888888888"`
	ContactName string `json:"contact_name,omitempty" example:"John Doe"`
	MessageID   string `json:"message_id,omitempty" example:"3EB0C767D71D"`
	Status      string `json:"status" example:"sent"`
	Success     bool   `json:"success" example:"true"`
	Error       string `json:"error,omitempty" example:""`
} // @name ContactResult

type SendContactListResponse struct {
	BaseResponse
	SessionID      string          `json:"session_id,omitempty" example:"session-123"`
	RemoteJID      string          `json:"remote_jid,omitempty" example:"5511999999999@s.whatsapp.net"`
	ContactCount   int             `json:"contact_count" example:"3"`
	ContactResults []ContactResult `json:"contact_results"`
	Results        []ContactResult `json:"results"`
	SentAt         time.Time       `json:"sent_at,omitempty" example:"2024-01-01T12:00:00Z"`
} // @name SendContactListResponse

type SendBusinessProfileMessageRequest struct {
	To          string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	BusinessJID string `json:"business_jid,omitempty" example:"5511888888888@s.whatsapp.net"`
	ReplyTo     string `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendBusinessProfileMessageRequest

type SendButtonMessageRequest struct {
	To      string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Text    string       `json:"text" validate:"required" example:"Choose an option:"`
	Footer  string       `json:"footer,omitempty" example:"Powered by ZPWoot"`
	Buttons []ButtonInfo `json:"buttons" validate:"required,min=1,max=3"`
	ReplyTo string       `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendButtonMessageRequest

type SendListMessageRequest struct {
	To         string            `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body       string            `json:"body" validate:"required" example:"Choose from the list:"`
	ButtonText string            `json:"button_text" validate:"required" example:"View Options"`
	Sections   []ListSectionInfo `json:"sections" validate:"required,min=1,max=10"`
	ReplyTo    string            `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendListMessageRequest

type SendPollMessageRequest struct {
	To                string           `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Name              string           `json:"name" validate:"required" example:"Favorite Color Poll"`
	Question          string           `json:"question" validate:"required" example:"What's your favorite color?"`
	Options           []PollOptionInfo `json:"options" validate:"required,min=2,max=12"`
	SelectableCount   int              `json:"selectable_count" validate:"min=1" example:"1"`
	AllowMultipleVote bool             `json:"allow_multiple_vote" example:"false"`
	ReplyTo           string           `json:"reply_to,omitempty" example:"3EB0C767D71D"`
} // @name SendPollMessageRequest

type SendReactionMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"message_id" validate:"required" example:"3EB0C767D71D"`
	Reaction  string `json:"reaction" validate:"required" example:"👍"`
} // @name SendReactionMessageRequest

type SendPresenceMessageRequest struct {
	To       string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Presence string `json:"presence" validate:"required,oneof=typing recording online offline paused" example:"typing"`
} // @name SendPresenceMessageRequest

type ButtonInfo struct {
	ID   string `json:"id" validate:"required" example:"btn-1"`
	Text string `json:"text" validate:"required" example:"Option 1"`
	Type string `json:"type,omitempty" example:"reply"`
} // @name ButtonInfo

type ListSectionInfo struct {
	Title string        `json:"title" validate:"required" example:"Section 1"`
	Rows  []ListRowInfo `json:"rows" validate:"required,min=1,max=10"`
} // @name ListSectionInfo

type ListRowInfo struct {
	ID          string `json:"id" validate:"required" example:"row-1"`
	Title       string `json:"title" validate:"required" example:"Option 1"`
	Description string `json:"description,omitempty" example:"Description of option 1"`
} // @name ListRowInfo

type PollOptionInfo struct {
	Name string `json:"name" validate:"required" example:"Option 1"`
} // @name PollOptionInfo

type EditMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"message_id" validate:"required" example:"3EB0C767D71D"`
	NewText   string `json:"new_text" validate:"required" example:"Updated message"`
	NewBody   string `json:"new_body,omitempty" example:"Updated message"`
} // @name EditMessageRequest

type RevokeMessageRequest struct {
	To        string `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"message_id" validate:"required" example:"3EB0C767D71D"`
} // @name RevokeMessageRequest

type MarkAsReadRequest struct {
	ChatJID    string   `json:"chat_jid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageIDs []string `json:"message_ids" validate:"required,min=1" example:"[\"3EB0C767D71D\"]"`
} // @name MarkAsReadRequest

type PollVoteInfo struct {
	OptionName string   `json:"option_name" example:"Option 1"`
	Voters     []string `json:"voters" example:"[\"5511888888888@s.whatsapp.net\"]"`
	VoteCount  int      `json:"vote_count" example:"5"`
} // @name PollVoteInfo

type GetPollResultsResponse struct {
	BaseResponse
	MessageID   string         `json:"message_id" example:"3EB0C767D71D"`
	PollID      string         `json:"poll_id" example:"3EB0C767D71D"`
	PollName    string         `json:"poll_name" example:"Favorite Color Poll"`
	Question    string         `json:"question" example:"What's your favorite color?"`
	Votes       []PollVoteInfo `json:"votes"`
	VoteResults []PollVoteInfo `json:"vote_results"`
	TotalVotes  int            `json:"total_votes" example:"15"`
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-01T12:00:00Z"`
} // @name GetPollResultsResponse

type MarkAsReadResponse struct {
	BaseResponse
	ChatJID      string    `json:"chat_jid" example:"5511999999999@s.whatsapp.net"`
	MessagesRead int       `json:"messages_read" example:"3"`
	MarkedCount  int       `json:"marked_count" example:"3"`
	Status       string    `json:"status" example:"success"`
	LastReadAt   time.Time `json:"last_read_at" example:"2024-01-01T12:00:00Z"`
} // @name MarkAsReadResponse
