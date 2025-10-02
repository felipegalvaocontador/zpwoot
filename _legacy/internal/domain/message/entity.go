package message

import (
	"errors"
	"strings"
	"time"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
	MessageTypeSticker  MessageType = "sticker"
	MessageTypeLocation MessageType = "location"
	MessageTypeContact  MessageType = "contact"
	MessageTypePoll     MessageType = "poll"
	MessageTypePollVote MessageType = "poll_vote"
)

type MediaSource string

const (
	MediaSourceURL    MediaSource = "url"
	MediaSourceBase64 MediaSource = "base64"
	MediaSourceFile   MediaSource = "file"
)

type SendResult struct {
	Timestamp time.Time `json:"timestamp"`
	MessageID string    `json:"messageId"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

type SendMessageRequest struct {
	ContextInfo  *ContextInfo `json:"contextInfo,omitempty"`
	MimeType     string       `json:"mimeType,omitempty" example:"image/jpeg"`
	Body         string       `json:"body,omitempty" example:"Hello World!"`
	Caption      string       `json:"caption,omitempty" example:"Image caption"`
	File         string       `json:"file,omitempty" example:"https://example.com/image.jpg"`
	Filename     string       `json:"filename,omitempty" example:"document.pdf"`
	To           string       `json:"to" validate:"required" example:"5511999999999@s.whatsapp.net"`
	Address      string       `json:"address,omitempty" example:"São Paulo, SP"`
	ContactName  string       `json:"contactName,omitempty" example:"John Doe"`
	ContactPhone string       `json:"contactPhone,omitempty" example:"+5511999999999"`
	Type         MessageType  `json:"type" validate:"required,oneof=text image audio video document sticker location contact" example:"text"`
	Latitude     float64      `json:"latitude,omitempty" example:"-23.5505"`
	Longitude    float64      `json:"longitude,omitempty" example:"-46.6333"`
}

type ContextInfo struct {
	StanzaID    string `json:"stanzaId" validate:"required" example:"ABCD1234abcd"`
	Participant string `json:"participant,omitempty" example:"5511999999999@s.whatsapp.net"`
}

type SendMessageResponse struct {
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	MessageID string    `json:"messageId" example:"3EB0C767D71D"`
	Status    string    `json:"status" example:"sent"`
}

type MediaInfo struct {
	MimeType string `json:"mimeType"`
	FileSize int64  `json:"fileSize"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

type LocationMessage struct {
	Address   string  `json:"address,omitempty"`
	Name      string  `json:"name,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ContactMessage struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

func (req *SendMessageRequest) IsMediaMessage() bool {
	return req.Type != MessageTypeText && req.Type != MessageTypeLocation && req.Type != MessageTypeContact
}

func (req *SendMessageRequest) GetMediaSource() MediaSource {
	if req.File == "" {
		return ""
	}

	if strings.HasPrefix(req.File, "data:") {
		return MediaSourceBase64
	}

	if strings.HasPrefix(req.File, "http://") || strings.HasPrefix(req.File, "https://") {
		return MediaSourceURL
	}

	return MediaSourceFile
}

var (
	ErrInvalidPollName        = errors.New("invalid poll name")
	ErrPollNameTooLong        = errors.New("poll name too long (max 100 characters)")
	ErrInsufficientOptions    = errors.New("poll must have at least 2 options")
	ErrTooManyOptions         = errors.New("poll cannot have more than 12 options")
	ErrInvalidOptionName      = errors.New("invalid option name")
	ErrOptionNameTooLong      = errors.New("option name too long (max 100 characters)")
	ErrInvalidSelectableCount = errors.New("invalid selectable option count")
	ErrInvalidPollMessageID   = errors.New("invalid poll message ID")
	ErrNoOptionsSelected      = errors.New("no options selected")
	ErrPollNotFound           = errors.New("poll not found")
	ErrDuplicateVote          = errors.New("duplicate vote not allowed")
	ErrTooManyOptionsSelected = errors.New("too many options selected")
	ErrInvalidOptionSelected  = errors.New("invalid option selected")
	ErrInvalidRecipient       = errors.New("invalid recipient")
)

type CreatePollRequest struct {
	To                    string
	Name                  string
	Options               []string
	SelectableOptionCount int
	AllowMultipleAnswers  bool
}

type VotePollRequest struct {
	To              string
	PollMessageID   string
	SelectedOptions []string
}

type GetPollResultsRequest struct {
	To            string
	PollMessageID string
}

type PollInfo struct {
	CreatedAt             time.Time
	MessageID             string
	Name                  string
	To                    string
	Options               []string
	SelectableOptionCount int
	AllowMultipleAnswers  bool
}

type PollOption struct {
	Name      string
	Voters    []string
	VoteCount int
}

type PollResults struct {
	CreatedAt             time.Time
	PollMessageID         string
	PollName              string
	To                    string
	Options               []PollOption
	TotalVotes            int
	SelectableOptionCount int
	AllowMultipleAnswers  bool
}

type PollVote struct {
	VotedAt         time.Time
	PollMessageID   string
	VoterJID        string
	SelectedOptions []string
}

type Poll struct {
	CreatedAt             time.Time
	MessageID             string
	Name                  string
	To                    string
	Options               []string
	Votes                 []PollVote
	SelectableOptionCount int
	AllowMultipleAnswers  bool
}

func ValidateCreatePollRequest(req *CreatePollRequest) error {
	if req.Name == "" {
		return ErrInvalidPollName
	}

	if len(req.Name) > 100 {
		return ErrPollNameTooLong
	}

	if len(req.Options) < 2 {
		return ErrInsufficientOptions
	}

	if len(req.Options) > 12 {
		return ErrTooManyOptions
	}

	optionMap := make(map[string]bool)
	for _, option := range req.Options {
		if option == "" {
			return ErrInvalidOptionName
		}
		if len(option) > 100 {
			return ErrOptionNameTooLong
		}
		if optionMap[option] {
			return errors.New("duplicate option names not allowed")
		}
		optionMap[option] = true
	}

	if req.SelectableOptionCount < 1 || req.SelectableOptionCount > len(req.Options) {
		return ErrInvalidSelectableCount
	}

	if req.To == "" {
		return ErrInvalidRecipient
	}

	return nil
}

func ValidateVotePollRequest(req *VotePollRequest) error {
	if req.PollMessageID == "" {
		return ErrInvalidPollMessageID
	}

	if len(req.SelectedOptions) == 0 {
		return ErrNoOptionsSelected
	}

	for _, option := range req.SelectedOptions {
		if option == "" {
			return ErrInvalidOptionName
		}
	}

	if req.To == "" {
		return ErrInvalidRecipient
	}

	return nil
}
