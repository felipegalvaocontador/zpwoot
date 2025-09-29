package message

import (
	"time"

	"zpwoot/internal/domain/message"
)

type SendMessageRequest struct {
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Type      string `json:"type" validate:"required,oneof=text image audio video document sticker location contact" example:"text"`
	Body      string `json:"body,omitempty" example:"Hello World!"`
	Caption   string `json:"caption,omitempty" example:"Image caption"`
	File      string `json:"file,omitempty" example:"https://example.com/image.jpg"`
	Filename  string `json:"filename,omitempty" example:"document.pdf"` // Only used for document type, not for audio
	MimeType  string `json:"mimeType,omitempty" example:"image/jpeg"`

	Latitude  float64 `json:"latitude,omitempty" example:"-23.5505"`
	Longitude float64 `json:"longitude,omitempty" example:"-46.6333"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP"`

	ContactName  string       `json:"contactName,omitempty" example:"John Doe"`
	ContactPhone string       `json:"contactPhone,omitempty" example:"+5511999999999"`
	ContextInfo  *ContextInfo `json:"contextInfo,omitempty"`
} //@name SendMessageRequest

type SendMessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name SendMessageResponse

func FromDomainRequest(req *message.SendMessageRequest) *SendMessageRequest {
	return &SendMessageRequest{
		RemoteJID:    req.To,
		Type:         string(req.Type),
		Body:         req.Body,
		Caption:      req.Caption,
		File:         req.File,
		Filename:     req.Filename,
		MimeType:     req.MimeType,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Address:      req.Address,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
	}
}

func (r *SendMessageRequest) ToDomainRequest() *message.SendMessageRequest {
	var contextInfo *message.ContextInfo
	if r.ContextInfo != nil {
		contextInfo = &message.ContextInfo{
			StanzaID:    r.ContextInfo.StanzaID,
			Participant: r.ContextInfo.Participant,
		}
	}

	return &message.SendMessageRequest{
		To:           r.RemoteJID,
		Type:         message.MessageType(r.Type),
		Body:         r.Body,
		Caption:      r.Caption,
		File:         r.File,
		Filename:     r.Filename,
		MimeType:     r.MimeType,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		Address:      r.Address,
		ContactName:  r.ContactName,
		ContactPhone: r.ContactPhone,
		ContextInfo:  contextInfo,
	}
}

func FromDomainResponse(resp *message.SendMessageResponse) *SendMessageResponse {
	return &SendMessageResponse{
		ID:        resp.MessageID,
		Status:    resp.Status,
		Timestamp: resp.Timestamp,
	}
}

func (r *SendMessageResponse) ToDomainResponse() *message.SendMessageResponse {
	return &message.SendMessageResponse{
		MessageID: r.ID,
		Status:    r.Status,
		Timestamp: r.Timestamp,
	}
}

type ButtonMessageRequest struct {
	RemoteJID string   `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body      string   `json:"body" validate:"required" example:"Please choose one of the options below:"`
	Buttons   []Button `json:"buttons" validate:"required,min=1,max=3"`
} //@name ButtonMessageRequest

type Button struct {
	ID   string `json:"id" example:"btn_yes"`
	Text string `json:"text" validate:"required" example:"Yes, I agree"`
} //@name Button

type ListMessageRequest struct {
	RemoteJID  string    `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Body       string    `json:"body" validate:"required" example:"Please select one of the available options:"`
	ButtonText string    `json:"buttonText" validate:"required" example:"Select Option"`
	Sections   []Section `json:"sections" validate:"required,min=1"`
} //@name ListMessageRequest

type Section struct {
	Title string `json:"title" example:"Available Services"`
	Rows  []Row  `json:"rows" validate:"required,min=1,max=10"`
} //@name Section

type Row struct {
	ID          string `json:"id" example:"service_support"`
	Title       string `json:"title" validate:"required" example:"Customer Support"`
	Description string `json:"description" example:"Get help from our support team"`
} //@name Row

type MediaMessageRequest struct {
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File      string `json:"file" validate:"required" example:"https://example.com/media.file"`
	Caption   string `json:"caption" example:"Media caption"`
	MimeType  string `json:"mimeType" example:"application/octet-stream"`
	Filename  string `json:"filename" example:"media.file"`
} //@name MediaMessageRequest

type ImageMessageRequest struct {
	RemoteJID   string       `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/image.jpg"`
	Caption     string       `json:"caption" example:"Beautiful sunset photo"`
	MimeType    string       `json:"mimeType" example:"image/jpeg"`
	Filename    string       `json:"filename" example:"sunset.jpg"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name ImageMessageRequest

type VideoMessageRequest struct {
	RemoteJID   string       `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/video.mp4"`
	Caption     string       `json:"caption" example:"Check out this amazing video!"`
	MimeType    string       `json:"mimeType" example:"video/mp4"`
	Filename    string       `json:"filename" example:"amazing_video.mp4"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name VideoMessageRequest

type AudioMessageRequest struct {
	RemoteJID   string       `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/audio.ogg"`
	Caption     string       `json:"caption" example:"Voice message"`
	MimeType    string       `json:"mimeType" example:"audio/ogg"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
	// Note: AudioMessage in WhatsApp protocol doesn't support filename field
	// Use DocumentMessage for files that need a filename
} //@name AudioMessageRequest

type DocumentMessageRequest struct {
	RemoteJID   string       `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	File        string       `json:"file" validate:"required" example:"https://example.com/document.pdf"`
	Caption     string       `json:"caption" example:"Important document"`
	MimeType    string       `json:"mimeType" example:"application/pdf"`
	Filename    string       `json:"filename" validate:"required" example:"important_document.pdf"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name DocumentMessageRequest

type LocationMessageRequest struct {
	RemoteJID string  `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Latitude  float64 `json:"latitude" validate:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" validate:"required" example:"-46.6333"`
	Address   string  `json:"address" example:"Avenida Paulista, 1578 - Bela Vista, São Paulo - SP, Brazil"`
} //@name LocationMessageRequest

type ContactMessageRequest struct {
	RemoteJID    string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	ContactName  string `json:"contactName" validate:"required" example:"Maria Silva"`
	ContactPhone string `json:"contactPhone" validate:"required" example:"+5511987654321"`
} //@name ContactMessageRequest

type ContactInfo struct {
	Name         string `json:"name" validate:"required" example:"João Santos"`
	Phone        string `json:"phone" validate:"required" example:"+5511987654321"`
	Email        string `json:"email,omitempty" example:"joao.santos@email.com"`                 // Not displayed in WhatsApp
	Organization string `json:"organization,omitempty" example:"Tech Solutions Ltda"`            // Displayed in WhatsApp
	Title        string `json:"title,omitempty" example:"Software Engineer"`                     // Not displayed in WhatsApp
	Website      string `json:"website,omitempty" example:"https://joaosantos.dev"`              // Not displayed in WhatsApp
	Address      string `json:"address,omitempty" example:"Rua das Flores, 123 - São Paulo, SP"` // Not displayed in WhatsApp
} //@name ContactInfo

type ContactListMessageRequest struct {
	RemoteJID string        `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Contacts  []ContactInfo `json:"contacts" validate:"required,min=1,max=10"`
} //@name ContactListMessageRequest

type ContactListMessageResponse struct {
	TotalContacts int                 `json:"totalContacts" example:"3"`
	SuccessCount  int                 `json:"successCount" example:"3"`
	FailureCount  int                 `json:"failureCount" example:"0"`
	Results       []ContactSendResult `json:"results"`
	Timestamp     string              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
} //@name ContactListMessageResponse

type ContactSendResult struct {
	ContactName string `json:"contactName" example:"João Santos"`
	MessageID   string `json:"messageId,omitempty" example:"3EB07F264CA1B4AD714A3F"`
	Status      string `json:"status" example:"sent"`
	Error       string `json:"error,omitempty"`
} //@name ContactSendResult

type ReactionMessageRequest struct {
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	Reaction  string `json:"reaction" validate:"required" example:"👍"`
} //@name ReactionMessageRequest

type PresenceMessageRequest struct {
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Presence  string `json:"presence" validate:"required,oneof=typing recording online offline paused" example:"typing"`
} //@name PresenceMessageRequest

type EditMessageRequest struct {
	SessionID string `json:"sessionId,omitempty" example:"mySession"`
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
	NewBody   string `json:"newBody" validate:"required" example:"Updated message text"`
} //@name EditMessageRequest

type EditMessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"edited"`
	NewBody   string    `json:"newBody" example:"Updated message text"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name EditMessageResponse

type RevokeMessageRequest struct {
	SessionID string `json:"sessionId" validate:"required" example:"mySession"`
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C767D71D"`
} //@name RevokeMessageRequest

type RevokeMessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"revoked"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name RevokeMessageResponse

type MarkAsReadRequest struct {
	SessionID  string   `json:"sessionId" validate:"required" example:"mySession"`
	RemoteJID  string   `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageIDs []string `json:"messageIds" validate:"required,min=1" example:"3EB0C767D71D,3EB0C767D71E"`
} //@name MarkAsReadRequest

type MarkAsReadResponse struct {
	MessageIDs []string  `json:"messageIds" example:"3EB0C767D71D,3EB0C767D71E"`
	Status     string    `json:"status" example:"read"`
	Timestamp  time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name MarkAsReadResponse

type MessageResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name MessageResponse

type ReactionResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Reaction  string    `json:"reaction" example:"👍"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name ReactionResponse

type PresenceResponse struct {
	Status    string    `json:"status" example:"sent"`
	Presence  string    `json:"presence" example:"typing"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name PresenceResponse

type EditResponse struct {
	ID        string    `json:"id" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"edited"`
	NewBody   string    `json:"newBody" example:"Updated message text"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name EditResponse

type BusinessProfileRequest struct {
	RemoteJID    string `json:"remoteJid" validate:"required" example:"5511987654321@s.whatsapp.net"`
	Name         string `json:"name" validate:"required" example:"Empresa Teste Ltda"`
	PhoneNumber  string `json:"phone" validate:"required" example:"+5511987654321"`
	Email        string `json:"email,omitempty" example:"contato@empresateste.com.br"`
	Organization string `json:"organization,omitempty" example:"Empresa Teste Ltda"`
	Title        string `json:"title,omitempty" example:"Atendimento ao Cliente"`
	Website      string `json:"website,omitempty" example:"https://www.empresateste.com.br"`
	Address      string `json:"address,omitempty" example:"Rua Teste, 123 - São Paulo, SP"`
} //@name BusinessProfileRequest

type TextMessageRequest struct {
	RemoteJID   string       `json:"remoteJid" validate:"required" example:"5511987654321@s.whatsapp.net"`
	Body        string       `json:"body" validate:"required" example:"Hello, this is a text message"`
	ContextInfo *ContextInfo `json:"contextInfo,omitempty"`
} //@name TextMessageRequest

type ContextInfo struct {
	StanzaID    string `json:"stanzaId" validate:"required" example:"ABCD1234abcd"`
	Participant string `json:"participant,omitempty" example:"5511999999999@s.whatsapp.net"`
} //@name ContextInfo

// Poll-related DTOs

// CreatePollRequest represents a request to create a poll
type CreatePollRequest struct {
	RemoteJID             string   `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Name                  string   `json:"name" validate:"required,min=1,max=100" example:"What's your favorite color?"`
	Options               []string `json:"options" validate:"required,min=2,max=12,dive,required,min=1,max=100" example:"Red,Blue,Green"`
	SelectableOptionCount int      `json:"selectableOptionCount" validate:"min=1" example:"1"`
	AllowMultipleAnswers  bool     `json:"allowMultipleAnswers" example:"false"`
} //@name CreatePollRequest

// CreatePollResponse represents the response after creating a poll
type CreatePollResponse struct {
	MessageID string    `json:"messageId" example:"3EB0C767D71D"`
	PollName  string    `json:"pollName" example:"What's your favorite color?"`
	Options   []string  `json:"options" example:"Red,Blue,Green"`
	RemoteJID string    `json:"remoteJid" example:"5511999999999@s.whatsapp.net"`
	Status    string    `json:"status" example:"sent"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name CreatePollResponse

// VotePollRequest represents a request to vote in a poll
type VotePollRequest struct {
	RemoteJID       string   `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	PollMessageID   string   `json:"pollMessageId" validate:"required" example:"3EB0C767D71D"`
	SelectedOptions []string `json:"selectedOptions" validate:"required,min=1,dive,required" example:"Red"`
} //@name VotePollRequest

// VotePollResponse represents the response after voting in a poll
type VotePollResponse struct {
	PollMessageID   string    `json:"pollMessageId" example:"3EB0C767D71D"`
	SelectedOptions []string  `json:"selectedOptions" example:"Red"`
	RemoteJID       string    `json:"remoteJid" example:"5511999999999@s.whatsapp.net"`
	Status          string    `json:"status" example:"sent"`
	Timestamp       time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
} //@name VotePollResponse

// GetPollResultsRequest represents a request to get poll results
type GetPollResultsRequest struct {
	RemoteJID     string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	PollMessageID string `json:"pollMessageId" validate:"required" example:"3EB0C767D71D"`
} //@name GetPollResultsRequest

// PollOption represents a poll option with vote count
type PollOption struct {
	Name      string   `json:"name" example:"Red"`
	VoteCount int      `json:"voteCount" example:"5"`
	Voters    []string `json:"voters,omitempty" example:"5511999999999@s.whatsapp.net"`
} //@name PollOption

// GetPollResultsResponse represents poll results
type GetPollResultsResponse struct {
	PollMessageID         string       `json:"pollMessageId" example:"3EB0C767D71D"`
	PollName              string       `json:"pollName" example:"What's your favorite color?"`
	Options               []PollOption `json:"options"`
	TotalVotes            int          `json:"totalVotes" example:"10"`
	SelectableOptionCount int          `json:"selectableOptionCount" example:"1"`
	AllowMultipleAnswers  bool         `json:"allowMultipleAnswers" example:"false"`
	CreatedAt             time.Time    `json:"createdAt" example:"2024-01-01T12:00:00Z"`
	RemoteJID             string       `json:"remoteJid" example:"5511999999999@s.whatsapp.net"`
} //@name GetPollResultsResponse

// MarkReadRequest represents a request to mark a message as read
type MarkReadRequest struct {
	RemoteJID string `json:"remoteJid" validate:"required" example:"5511999999999@s.whatsapp.net"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C431C26A1916E07E"`
} //@name MarkReadRequest

// MarkReadResponse represents the response for marking a message as read
type MarkReadResponse struct {
	MessageID string    `json:"messageId" example:"3EB0C431C26A1916E07E"`
	Success   bool      `json:"success" example:"true"`
	MarkedAt  time.Time `json:"markedAt" example:"2024-01-01T12:00:00Z"`
	Message   string    `json:"message" example:"Message marked as read successfully"`
} //@name MarkReadResponse
