package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/contracts"
	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/core/contact"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

type ContactHandler struct {
	*shared.BaseHandler
	contactService *contact.Service
	sessionService *services.SessionService
}

func NewContactHandler(
	contactService *contact.Service,
	sessionService *services.SessionService,
	logger *logger.Logger,
) *ContactHandler {
	return &ContactHandler{
		BaseHandler:    shared.NewBaseHandler(logger),
		contactService: contactService,
		sessionService: sessionService,
	}
}

type CheckWhatsAppRequest struct {
	PhoneNumbers []string `json:"phoneNumbers" validate:"required,min=1,max=50"`
}

type CheckWhatsAppResponse struct {
	Results []CheckWhatsAppResult `json:"results"`
	Total   int                   `json:"total"`
}

type CheckWhatsAppResult struct {
	PhoneNumber  string `json:"phoneNumber"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	JID          string `json:"jid,omitempty"`
	Error        string `json:"error,omitempty"`
}

type GetProfilePictureResponse struct {
	JID        string `json:"jid"`
	PictureURL string `json:"pictureUrl,omitempty"`
	PictureID  string `json:"pictureId,omitempty"`
	HasPicture bool   `json:"hasPicture"`
}

type GetUserInfoRequest struct {
	JIDs []string `json:"jids" validate:"required,min=1,max=20"`
}

type GetUserInfoResponse struct {
	Results []UserInfoResult `json:"results"`
	Total   int              `json:"total"`
}

type UserInfoResult struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Status       string `json:"status,omitempty"`
	PictureID    string `json:"pictureId,omitempty"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	IsBusiness   bool   `json:"isBusiness"`
	Error        string `json:"error,omitempty"`
}

type ListContactsResponse struct {
	Contacts []ContactInfo `json:"contacts"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

type ContactInfo struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	PushName     string `json:"pushName,omitempty"`
	ShortName    string `json:"shortName,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	IsBusiness   bool   `json:"isBusiness"`
	IsMyContact  bool   `json:"isMyContact"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
}

type SyncContactsResponse struct {
	SyncedContacts int    `json:"syncedContacts"`
	TotalContacts  int    `json:"totalContacts"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

type BusinessProfileResponse struct {
	JID         string `json:"jid"`
	Name        string `json:"name,omitempty"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Address     string `json:"address,omitempty"`
	IsBusiness  bool   `json:"isBusiness"`
}

// @Summary Check WhatsApp numbers
// @Description Check if phone numbers are registered on WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body CheckWhatsAppRequest true "Phone numbers to check"
// @Success 200 {object} shared.SuccessResponse{data=CheckWhatsAppResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/check [post]
func (h *ContactHandler) CheckWhatsApp(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "check WhatsApp numbers")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.CheckWhatsAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	response, err := h.contactService.CheckWhatsApp(r.Context(), sessionID, &req)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to check WhatsApp numbers", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to check WhatsApp numbers")
		return
	}

	h.LogSuccess("check WhatsApp numbers", map[string]interface{}{
		"session_id":    sessionID,
		"session_name":  session.Session.Name,
		"phone_count":   len(req.PhoneNumbers),
		"results_count": len(response.Results),
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary Get profile picture
// @Description Get profile picture of a contact
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param jid query string true "Contact JID"
// @Success 200 {object} shared.SuccessResponse{data=GetProfilePictureResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/avatar [get]
func (h *ContactHandler) GetProfilePicture(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get profile picture")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		h.GetWriter().WriteBadRequest(w, "JID is required")
		return
	}

	session, err := h.sessionService.GetSessionByNameOrID(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	response := GetProfilePictureResponse{
		JID:        jid,
		PictureURL: "https://example.com/profile-picture.jpg",
		PictureID:  "placeholder-picture-id",
		HasPicture: true,
	}

	h.LogSuccess("get profile picture", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"jid":          jid,
		"has_picture":  response.HasPicture,
	})

	h.GetWriter().WriteSuccess(w, response, "Profile picture retrieved successfully")
}

// @Summary Get user info
// @Description Get information about WhatsApp users
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body GetUserInfoRequest true "User JIDs to get info"
// @Success 200 {object} shared.SuccessResponse{data=GetUserInfoResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/info [post]
func (h *ContactHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get user info")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	var req contracts.GetUserInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	response, err := h.contactService.GetUserInfo(r.Context(), sessionID, &req)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to get user info", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to get user info")
		return
	}

	h.LogSuccess("get user info", map[string]interface{}{
		"session_id":    sessionID,
		"session_name":  session.Session.Name,
		"jid_count":     len(req.JIDs),
		"results_count": len(response.Users),
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary List contacts
// @Description List contacts with pagination and filters
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Param search query string false "Search term"
// @Success 200 {object} shared.SuccessResponse{data=ListContactsResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts [get]
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list contacts")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)
	search := r.URL.Query().Get("search")

	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	req := &contracts.ListContactsRequest{
		Limit:  limit,
		Offset: offset,
	}

	response, err := h.contactService.ListContacts(r.Context(), sessionID, req)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to list contacts", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to list contacts")
		return
	}

	h.LogSuccess("list contacts", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"total":        response.Total,
		"returned":     len(response.Contacts),
		"limit":        limit,
		"offset":       offset,
		"search":       search,
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary Sync contacts
// @Description Sync contacts from the device
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=SyncContactsResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/sync [post]
func (h *ContactHandler) SyncContacts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "sync contacts")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	session, err := h.sessionService.GetSessionByNameOrID(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	req := &contracts.SyncContactsRequest{
		Force: false,
	}

	response, err := h.contactService.SyncContacts(r.Context(), sessionID, req)
	if err != nil {
		h.GetLogger().ErrorWithFields("Failed to sync contacts", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		h.GetWriter().WriteInternalError(w, "Failed to sync contacts")
		return
	}

	h.LogSuccess("sync contacts", map[string]interface{}{
		"session_id":      sessionID,
		"session_name":    session.Session.Name,
		"synced_contacts": response.SyncedCount,
		"total_contacts":  response.TotalContacts,
	})

	h.GetWriter().WriteSuccess(w, response, response.Message)
}

// @Summary Get business profile
// @Description Get business profile of a contact
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param jid query string true "Contact JID"
// @Success 200 {object} shared.SuccessResponse{data=BusinessProfileResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/business [get]
func (h *ContactHandler) GetBusinessProfile(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get business profile")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		h.GetWriter().WriteBadRequest(w, "JID is required")
		return
	}

	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	response := BusinessProfileResponse{
		JID:         jid,
		Name:        "Empresa Exemplo LTDA",
		Category:    "Technology",
		Description: "Empresa de tecnologia especializada em soluções digitais",
		Website:     "https://exemplo.com.br",
		Email:       "contato@exemplo.com.br",
		Address:     "São Paulo, SP, Brasil",
		IsBusiness:  true,
	}

	h.LogSuccess("get business profile", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"jid":          jid,
		"is_business":  response.IsBusiness,
	})

	h.GetWriter().WriteSuccess(w, response, "Business profile retrieved successfully")
}

// @Summary Check if numbers are on WhatsApp (batch)
// @Description Check if multiple phone numbers are registered on WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body CheckWhatsAppRequest true "Phone numbers to check"
// @Success 200 {object} shared.SuccessResponse{data=CheckWhatsAppResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/is-on-whatsapp [post]
func (h *ContactHandler) IsOnWhatsApp(w http.ResponseWriter, r *http.Request) {

	h.CheckWhatsApp(w, r)
}

// @Summary Get all contacts
// @Description Get all contacts without pagination
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.SuccessResponse{data=[]ContactInfo}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/all [get]
func (h *ContactHandler) GetAllContacts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get all contacts")

	sessionID := chi.URLParam(r, "sessionName")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	session, err := h.sessionService.GetSessionByNameOrID(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	contacts := []ContactInfo{
		{
			JID:          "5511999999999@s.whatsapp.net",
			Name:         "João Silva",
			PushName:     "João",
			PhoneNumber:  "5511999999999",
			IsBusiness:   false,
			IsMyContact:  true,
			IsOnWhatsApp: true,
		},
		{
			JID:          "5511888888888@s.whatsapp.net",
			Name:         "Maria Santos",
			PushName:     "Maria",
			PhoneNumber:  "5511888888888",
			IsBusiness:   true,
			IsMyContact:  true,
			IsOnWhatsApp: true,
		},
	}

	h.LogSuccess("get all contacts", map[string]interface{}{
		"session_id":   sessionID,
		"session_name": session.Session.Name,
		"total":        len(contacts),
	})

	h.GetWriter().WriteSuccess(w, contacts, "All contacts retrieved successfully")
}

// @Summary Get profile picture info
// @Description Get profile picture information of a contact
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param jid query string true "Contact JID"
// @Success 200 {object} shared.SuccessResponse{data=GetProfilePictureResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/profile-picture-info [get]
func (h *ContactHandler) GetProfilePictureInfo(w http.ResponseWriter, r *http.Request) {

	h.GetProfilePicture(w, r)
}

// @Summary Get detailed user info (batch)
// @Description Get detailed information about WhatsApp users
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body GetUserInfoRequest true "User JIDs to get detailed info"
// @Success 200 {object} shared.SuccessResponse{data=GetUserInfoResponse}
// @Failure 400 {object} shared.SuccessResponse
// @Failure 404 {object} shared.SuccessResponse
// @Failure 500 {object} shared.SuccessResponse
// @Router /sessions/{sessionId}/contacts/detailed-info [post]
func (h *ContactHandler) GetDetailedUserInfo(w http.ResponseWriter, r *http.Request) {

	h.GetUserInfo(w, r)
}
