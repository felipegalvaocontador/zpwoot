package newsletter

import (
	"fmt"
	"time"

	"zpwoot/internal/domain/newsletter"
)

// CreateNewsletterRequest - Request para criar newsletter
type CreateNewsletterRequest struct {
	Name        string `json:"name" validate:"required,max=25"`
	Description string `json:"description,omitempty" validate:"max=512"`
}

// CreateNewsletterResponse - Response da criação de newsletter
type CreateNewsletterResponse struct {
	CreatedAt   time.Time `json:"createdAt"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	InviteCode  string    `json:"inviteCode"`
	State       string    `json:"state"`
	Role        string    `json:"role"`
}

// GetNewsletterInfoRequest - Request para obter info de newsletter
type GetNewsletterInfoRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

// GetNewsletterInfoWithInviteRequest - Request para obter info via convite
type GetNewsletterInfoWithInviteRequest struct {
	InviteKey string `json:"inviteKey" validate:"required"`
}

// NewsletterInfoResponse - Response com informações do newsletter
type NewsletterInfoResponse struct {
	CreationTime      time.Time           `json:"creationTime"`
	UpdateTime        time.Time           `json:"updateTime"`
	Preview           *ProfilePictureInfo `json:"preview,omitempty"`
	Picture           *ProfilePictureInfo `json:"picture,omitempty"`
	Role              string              `json:"role"`
	State             string              `json:"state"`
	ID                string              `json:"id"`
	MuteState         string              `json:"muteState"`
	VerificationState string              `json:"verificationState"`
	InviteCode        string              `json:"inviteCode"`
	Description       string              `json:"description"`
	Name              string              `json:"name"`
	SubscriberCount   int                 `json:"subscriberCount"`
	Muted             bool                `json:"muted"`
	Verified          bool                `json:"verified"`
}

// ProfilePictureInfo - Informações da foto do perfil
type ProfilePictureInfo struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Direct string `json:"direct"`
}

// FollowNewsletterRequest - Request para seguir newsletter
type FollowNewsletterRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

// GetNewsletterJID returns the newsletter JID for interface compliance
func (req *FollowNewsletterRequest) GetNewsletterJID() string {
	return req.NewsletterJID
}

// UnfollowNewsletterRequest - Request para deixar de seguir newsletter
type UnfollowNewsletterRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

// GetNewsletterJID returns the newsletter JID for interface compliance
func (req *UnfollowNewsletterRequest) GetNewsletterJID() string {
	return req.NewsletterJID
}

// SubscribedNewslettersResponse - Response com newsletters seguidos
type SubscribedNewslettersResponse struct {
	Newsletters []NewsletterInfoResponse `json:"newsletters"`
	Total       int                      `json:"total"`
}

// NewsletterActionResponse - Response genérica para ações
type NewsletterActionResponse struct {
	Timestamp time.Time `json:"timestamp"`
	JID       string    `json:"jid"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}

// Conversion methods

// ToDomain converts CreateNewsletterRequest to domain entity
func (req *CreateNewsletterRequest) ToDomain() *newsletter.CreateNewsletterRequest {
	return &newsletter.CreateNewsletterRequest{
		Name:        req.Name,
		Description: req.Description,
	}
}

// FromDomain converts domain NewsletterInfo to NewsletterInfoResponse
func (resp *NewsletterInfoResponse) FromDomain(info *newsletter.NewsletterInfo) {
	resp.ID = info.ID
	resp.Name = info.Name
	resp.Description = info.Description
	resp.InviteCode = info.InviteCode
	resp.SubscriberCount = info.SubscriberCount
	resp.State = string(info.State)
	resp.Role = string(info.Role)
	resp.Muted = info.Muted
	resp.MuteState = string(info.MuteState)
	resp.Verified = info.Verified
	resp.VerificationState = string(info.VerificationState)
	resp.CreationTime = info.CreationTime
	resp.UpdateTime = info.UpdateTime

	if info.Picture != nil {
		resp.Picture = &ProfilePictureInfo{
			URL:    info.Picture.URL,
			ID:     info.Picture.ID,
			Type:   info.Picture.Type,
			Direct: info.Picture.Direct,
		}
	}

	if info.Preview != nil {
		resp.Preview = &ProfilePictureInfo{
			URL:    info.Preview.URL,
			ID:     info.Preview.ID,
			Type:   info.Preview.Type,
			Direct: info.Preview.Direct,
		}
	}
}

// FromDomainList converts a list of domain NewsletterInfo to NewsletterInfoResponse list
func FromDomainList(infos []*newsletter.NewsletterInfo) []NewsletterInfoResponse {
	responses := make([]NewsletterInfoResponse, len(infos))
	for i, info := range infos {
		responses[i].FromDomain(info)
	}
	return responses
}

// ToDomain converts GetNewsletterInfoRequest to domain entity
func (req *GetNewsletterInfoRequest) ToDomain() *newsletter.GetNewsletterInfoRequest {
	return &newsletter.GetNewsletterInfoRequest{
		JID: req.NewsletterJID,
	}
}

// ToDomain converts GetNewsletterInfoWithInviteRequest to domain entity
func (req *GetNewsletterInfoWithInviteRequest) ToDomain() *newsletter.GetNewsletterInfoWithInviteRequest {
	return &newsletter.GetNewsletterInfoWithInviteRequest{
		InviteKey: req.InviteKey,
	}
}

// ToDomain converts FollowNewsletterRequest to domain entity
func (req *FollowNewsletterRequest) ToDomain() *newsletter.FollowNewsletterRequest {
	return &newsletter.FollowNewsletterRequest{
		JID: req.NewsletterJID,
	}
}

// ToDomain converts UnfollowNewsletterRequest to domain entity
func (req *UnfollowNewsletterRequest) ToDomain() *newsletter.UnfollowNewsletterRequest {
	return &newsletter.UnfollowNewsletterRequest{
		JID: req.NewsletterJID,
	}
}

// Validation methods

// Validate validates the CreateNewsletterRequest
func (req *CreateNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the GetNewsletterInfoRequest
func (req *GetNewsletterInfoRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the GetNewsletterInfoWithInviteRequest
func (req *GetNewsletterInfoWithInviteRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the FollowNewsletterRequest
func (req *FollowNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Validate validates the UnfollowNewsletterRequest
func (req *UnfollowNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

// Helper functions

// NewCreateNewsletterResponse creates a new CreateNewsletterResponse from domain data
func NewCreateNewsletterResponse(info *newsletter.NewsletterInfo) *CreateNewsletterResponse {
	return &CreateNewsletterResponse{
		ID:          info.ID,
		Name:        info.Name,
		Description: info.Description,
		InviteCode:  info.InviteCode,
		State:       string(info.State),
		Role:        string(info.Role),
		CreatedAt:   info.CreationTime,
	}
}

// NewNewsletterInfoResponse creates a new NewsletterInfoResponse from domain data
func NewNewsletterInfoResponse(info *newsletter.NewsletterInfo) *NewsletterInfoResponse {
	resp := &NewsletterInfoResponse{}
	resp.FromDomain(info)
	return resp
}

// NewSubscribedNewslettersResponse creates a new SubscribedNewslettersResponse
func NewSubscribedNewslettersResponse(infos []*newsletter.NewsletterInfo) *SubscribedNewslettersResponse {
	newsletters := FromDomainList(infos)
	return &SubscribedNewslettersResponse{
		Newsletters: newsletters,
		Total:       len(newsletters),
	}
}

// NewNewsletterActionResponse creates a new NewsletterActionResponse
func NewNewsletterActionResponse(jid, status, message string) *NewsletterActionResponse {
	return &NewsletterActionResponse{
		JID:       jid,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// Success response helpers

// NewSuccessFollowResponse creates a success response for follow action
func NewSuccessFollowResponse(jid string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "success", "Newsletter followed successfully")
}

// NewSuccessUnfollowResponse creates a success response for unfollow action
func NewSuccessUnfollowResponse(jid string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "success", "Newsletter unfollowed successfully")
}

// Error response helpers

// NewErrorResponse creates an error response
func NewErrorResponse(jid, message string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "error", message)
}

// GetNewsletterMessagesRequest represents the request for getting newsletter messages
type GetNewsletterMessagesRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	Before        string `json:"before,omitempty"`
	Count         int    `json:"count,omitempty"`
}

// Validate validates the GetNewsletterMessagesRequest
func (req *GetNewsletterMessagesRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if req.Count < 0 {
		return fmt.Errorf("count cannot be negative")
	}
	return nil
}

// NewsletterMessageDTO represents a newsletter message in the API
type NewsletterMessageDTO struct {
	Timestamp   time.Time `json:"timestamp"`
	ID          string    `json:"id"`
	ServerID    string    `json:"serverId"`
	FromJID     string    `json:"fromJid"`
	Type        string    `json:"type"`
	Body        string    `json:"body,omitempty"`
	Reactions   []string  `json:"reactions,omitempty"`
	ViewsCount  int       `json:"viewsCount"`
	SharesCount int       `json:"sharesCount"`
}

// FromDomain converts domain NewsletterMessage to DTO
func (dto *NewsletterMessageDTO) FromDomain(msg *newsletter.NewsletterMessage) {
	dto.ID = msg.ID
	dto.ServerID = msg.ServerID
	dto.FromJID = msg.FromJID
	dto.Timestamp = msg.Timestamp
	dto.Type = msg.Type
	dto.Body = msg.Body
	dto.ViewsCount = msg.ViewsCount
	dto.SharesCount = msg.SharesCount
	dto.Reactions = msg.Reactions
}

// GetNewsletterMessagesResponse represents the response for getting newsletter messages
type GetNewsletterMessagesResponse struct {
	Messages []*NewsletterMessageDTO `json:"messages"`
	Total    int                     `json:"total"`
	HasMore  bool                    `json:"hasMore"`
}

// NewGetNewsletterMessagesResponse creates a new GetNewsletterMessagesResponse from domain data
func NewGetNewsletterMessagesResponse(messages []*newsletter.NewsletterMessage) *GetNewsletterMessagesResponse {
	dtoMessages := make([]*NewsletterMessageDTO, len(messages))
	for i, msg := range messages {
		dto := &NewsletterMessageDTO{}
		dto.FromDomain(msg)
		dtoMessages[i] = dto
	}

	return &GetNewsletterMessagesResponse{
		Messages: dtoMessages,
		Total:    len(messages),
		HasMore:  len(messages) > 0, // Simple logic, can be improved
	}
}

// GetNewsletterMessageUpdatesRequest represents the request for getting newsletter message updates
type GetNewsletterMessageUpdatesRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	Since         string `json:"since,omitempty"`
	After         string `json:"after,omitempty"`
	Count         int    `json:"count,omitempty"`
}

// Validate validates the GetNewsletterMessageUpdatesRequest
func (req *GetNewsletterMessageUpdatesRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if req.Count < 0 {
		return fmt.Errorf("count cannot be negative")
	}
	return nil
}

// GetNewsletterMessageUpdatesResponse represents the response for getting newsletter message updates
type GetNewsletterMessageUpdatesResponse struct {
	Updates []*NewsletterMessageDTO `json:"updates"`
	Total   int                     `json:"total"`
	HasMore bool                    `json:"hasMore"`
}

// NewGetNewsletterMessageUpdatesResponse creates a new GetNewsletterMessageUpdatesResponse from domain data
func NewGetNewsletterMessageUpdatesResponse(updates []*newsletter.NewsletterMessage) *GetNewsletterMessageUpdatesResponse {
	dtoUpdates := make([]*NewsletterMessageDTO, len(updates))
	for i, update := range updates {
		dto := &NewsletterMessageDTO{}
		dto.FromDomain(update)
		dtoUpdates[i] = dto
	}

	return &GetNewsletterMessageUpdatesResponse{
		Updates: dtoUpdates,
		Total:   len(updates),
		HasMore: len(updates) > 0, // Simple logic, can be improved
	}
}

// NewsletterMarkViewedRequest represents the request for marking newsletter messages as viewed
type NewsletterMarkViewedRequest struct {
	NewsletterJID string   `json:"newsletterJid" validate:"required"`
	ServerIDs     []string `json:"serverIds" validate:"required,min=1"`
}

// Validate validates the NewsletterMarkViewedRequest
func (req *NewsletterMarkViewedRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if len(req.ServerIDs) == 0 {
		return fmt.Errorf("serverIds is required and cannot be empty")
	}
	return nil
}

// NewsletterSendReactionRequest represents the request for sending a reaction to a newsletter message
type NewsletterSendReactionRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	ServerID      string `json:"serverId,omitempty"`            // Optional - will be looked up from MessageID if not provided
	Reaction      string `json:"reaction"`                      // Empty string to remove reaction
	MessageID     string `json:"messageId" validate:"required"` // Required - used to find ServerID if not provided
}

// Validate validates the NewsletterSendReactionRequest
func (req *NewsletterSendReactionRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if req.MessageID == "" {
		return fmt.Errorf("messageId is required")
	}
	// ServerID is optional - will be looked up if not provided
	return nil
}

// NewsletterToggleMuteRequest represents the request for toggling newsletter mute status
type NewsletterToggleMuteRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	Mute          bool   `json:"mute"`
}

// Validate validates the NewsletterToggleMuteRequest
func (req *NewsletterToggleMuteRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	return nil
}

// NewsletterSubscribeLiveUpdatesRequest represents the request for subscribing to live updates
type NewsletterSubscribeLiveUpdatesRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

// Validate validates the NewsletterSubscribeLiveUpdatesRequest
func (req *NewsletterSubscribeLiveUpdatesRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	return nil
}

// NewsletterSubscribeLiveUpdatesResponse represents the response for subscribing to live updates
type NewsletterSubscribeLiveUpdatesResponse struct {
	Duration int64 `json:"duration"` // Duration in seconds
}

// AcceptTOSNoticeRequest represents the request for accepting terms of service notice
type AcceptTOSNoticeRequest struct {
	NoticeID string `json:"noticeId" validate:"required"`
	Stage    string `json:"stage" validate:"required"`
}

// Validate validates the AcceptTOSNoticeRequest
func (req *AcceptTOSNoticeRequest) Validate() error {
	if req.NoticeID == "" {
		return fmt.Errorf("noticeId is required")
	}
	if req.Stage == "" {
		return fmt.Errorf("stage is required")
	}
	return nil
}

// UploadNewsletterRequest represents the request for uploading newsletter media
type UploadNewsletterRequest struct {
	MimeType  string `json:"mimeType" validate:"required"`
	MediaType string `json:"mediaType" validate:"required"`
	Data      []byte `json:"data" validate:"required"`
}

// Validate validates the UploadNewsletterRequest
func (req *UploadNewsletterRequest) Validate() error {
	if len(req.Data) == 0 {
		return fmt.Errorf("data is required and cannot be empty")
	}
	if req.MimeType == "" {
		return fmt.Errorf("mimeType is required")
	}
	if req.MediaType == "" {
		return fmt.Errorf("mediaType is required")
	}
	// Validate media type
	validMediaTypes := []string{"image", "video", "audio", "document"}
	isValid := false
	for _, validType := range validMediaTypes {
		if req.MediaType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("mediaType must be one of: %v", validMediaTypes)
	}
	return nil
}

// UploadNewsletterResponse represents the response for uploading newsletter media
type UploadNewsletterResponse struct {
	URL        string `json:"url"`
	DirectPath string `json:"directPath"`
	Handle     string `json:"handle"`
	ObjectID   string `json:"objectId"`
	FileSHA256 string `json:"fileSha256"`
	FileLength uint64 `json:"fileLength"`
}
