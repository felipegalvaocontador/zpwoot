package newsletter

import (
	"fmt"
	"time"

	"zpwoot/internal/domain/newsletter"
)

type CreateNewsletterRequest struct {
	Name        string `json:"name" validate:"required,max=25"`
	Description string `json:"description,omitempty" validate:"max=512"`
}

type CreateNewsletterResponse struct {
	CreatedAt   time.Time `json:"createdAt"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	InviteCode  string    `json:"inviteCode"`
	State       string    `json:"state"`
	Role        string    `json:"role"`
}

type GetNewsletterInfoRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

type GetNewsletterInfoWithInviteRequest struct {
	InviteKey string `json:"inviteKey" validate:"required"`
}

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

type ProfilePictureInfo struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Direct string `json:"direct"`
}

type FollowNewsletterRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

func (req *FollowNewsletterRequest) GetNewsletterJID() string {
	return req.NewsletterJID
}

type UnfollowNewsletterRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

func (req *UnfollowNewsletterRequest) GetNewsletterJID() string {
	return req.NewsletterJID
}

type SubscribedNewslettersResponse struct {
	Newsletters []NewsletterInfoResponse `json:"newsletters"`
	Total       int                      `json:"total"`
}

type NewsletterActionResponse struct {
	Timestamp time.Time `json:"timestamp"`
	JID       string    `json:"jid"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
}


func (req *CreateNewsletterRequest) ToDomain() *newsletter.CreateNewsletterRequest {
	return &newsletter.CreateNewsletterRequest{
		Name:        req.Name,
		Description: req.Description,
	}
}

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

func FromDomainList(infos []*newsletter.NewsletterInfo) []NewsletterInfoResponse {
	responses := make([]NewsletterInfoResponse, len(infos))
	for i, info := range infos {
		responses[i].FromDomain(info)
	}
	return responses
}

func (req *GetNewsletterInfoRequest) ToDomain() *newsletter.GetNewsletterInfoRequest {
	return &newsletter.GetNewsletterInfoRequest{
		JID: req.NewsletterJID,
	}
}

func (req *GetNewsletterInfoWithInviteRequest) ToDomain() *newsletter.GetNewsletterInfoWithInviteRequest {
	return &newsletter.GetNewsletterInfoWithInviteRequest{
		InviteKey: req.InviteKey,
	}
}

func (req *FollowNewsletterRequest) ToDomain() *newsletter.FollowNewsletterRequest {
	return &newsletter.FollowNewsletterRequest{
		JID: req.NewsletterJID,
	}
}

func (req *UnfollowNewsletterRequest) ToDomain() *newsletter.UnfollowNewsletterRequest {
	return &newsletter.UnfollowNewsletterRequest{
		JID: req.NewsletterJID,
	}
}


func (req *CreateNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

func (req *GetNewsletterInfoRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

func (req *GetNewsletterInfoWithInviteRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

func (req *FollowNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}

func (req *UnfollowNewsletterRequest) Validate() error {
	domainReq := req.ToDomain()
	return domainReq.Validate()
}


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

func NewNewsletterInfoResponse(info *newsletter.NewsletterInfo) *NewsletterInfoResponse {
	resp := &NewsletterInfoResponse{}
	resp.FromDomain(info)
	return resp
}

func NewSubscribedNewslettersResponse(infos []*newsletter.NewsletterInfo) *SubscribedNewslettersResponse {
	newsletters := FromDomainList(infos)
	return &SubscribedNewslettersResponse{
		Newsletters: newsletters,
		Total:       len(newsletters),
	}
}

func NewNewsletterActionResponse(jid, status, message string) *NewsletterActionResponse {
	return &NewsletterActionResponse{
		JID:       jid,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}
}


func NewSuccessFollowResponse(jid string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "success", "Newsletter followed successfully")
}

func NewSuccessUnfollowResponse(jid string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "success", "Newsletter unfollowed successfully")
}


func NewErrorResponse(jid, message string) *NewsletterActionResponse {
	return NewNewsletterActionResponse(jid, "error", message)
}

type GetNewsletterMessagesRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	Before        string `json:"before,omitempty"`
	Count         int    `json:"count,omitempty"`
}

func (req *GetNewsletterMessagesRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if req.Count < 0 {
		return fmt.Errorf("count cannot be negative")
	}
	return nil
}

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

type GetNewsletterMessagesResponse struct {
	Messages []*NewsletterMessageDTO `json:"messages"`
	Total    int                     `json:"total"`
	HasMore  bool                    `json:"hasMore"`
}

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

type GetNewsletterMessageUpdatesRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	Since         string `json:"since,omitempty"`
	After         string `json:"after,omitempty"`
	Count         int    `json:"count,omitempty"`
}

func (req *GetNewsletterMessageUpdatesRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if req.Count < 0 {
		return fmt.Errorf("count cannot be negative")
	}
	return nil
}

type GetNewsletterMessageUpdatesResponse struct {
	Updates []*NewsletterMessageDTO `json:"updates"`
	Total   int                     `json:"total"`
	HasMore bool                    `json:"hasMore"`
}

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

type NewsletterMarkViewedRequest struct {
	NewsletterJID string   `json:"newsletterJid" validate:"required"`
	ServerIDs     []string `json:"serverIds" validate:"required,min=1"`
}

func (req *NewsletterMarkViewedRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if len(req.ServerIDs) == 0 {
		return fmt.Errorf("serverIds is required and cannot be empty")
	}
	return nil
}

type NewsletterSendReactionRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	ServerID      string `json:"serverId,omitempty"`            // Optional - will be looked up from MessageID if not provided
	Reaction      string `json:"reaction"`                      // Empty string to remove reaction
	MessageID     string `json:"messageId" validate:"required"` // Required - used to find ServerID if not provided
}

func (req *NewsletterSendReactionRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	if req.MessageID == "" {
		return fmt.Errorf("messageId is required")
	}
	return nil
}

type NewsletterToggleMuteRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
	Mute          bool   `json:"mute"`
}

func (req *NewsletterToggleMuteRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	return nil
}

type NewsletterSubscribeLiveUpdatesRequest struct {
	NewsletterJID string `json:"newsletterJid" validate:"required"`
}

func (req *NewsletterSubscribeLiveUpdatesRequest) Validate() error {
	if req.NewsletterJID == "" {
		return fmt.Errorf("newsletterJid is required")
	}
	return nil
}

type NewsletterSubscribeLiveUpdatesResponse struct {
	Duration int64 `json:"duration"` // Duration in seconds
}

type AcceptTOSNoticeRequest struct {
	NoticeID string `json:"noticeId" validate:"required"`
	Stage    string `json:"stage" validate:"required"`
}

func (req *AcceptTOSNoticeRequest) Validate() error {
	if req.NoticeID == "" {
		return fmt.Errorf("noticeId is required")
	}
	if req.Stage == "" {
		return fmt.Errorf("stage is required")
	}
	return nil
}

type UploadNewsletterRequest struct {
	MimeType  string `json:"mimeType" validate:"required"`
	MediaType string `json:"mediaType" validate:"required"`
	Data      []byte `json:"data" validate:"required"`
}

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

type UploadNewsletterResponse struct {
	URL        string `json:"url"`
	DirectPath string `json:"directPath"`
	Handle     string `json:"handle"`
	ObjectID   string `json:"objectId"`
	FileSHA256 string `json:"fileSha256"`
	FileLength uint64 `json:"fileLength"`
}
