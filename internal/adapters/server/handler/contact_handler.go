package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/shared"
	"zpwoot/internal/core/contact"
	"zpwoot/internal/services"
	"zpwoot/internal/services/shared/dto"
	"zpwoot/platform/logger"
)

// ContactHandler implementa handlers REST para contatos integrados ao WhatsApp
type ContactHandler struct {
	*shared.BaseHandler
	contactService *contact.Service
	sessionService *services.SessionService
}

// NewContactHandler cria nova instância do handler de contatos
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

// ===== REQUEST/RESPONSE TYPES =====

// CheckWhatsAppRequest representa request para verificar números no WhatsApp
type CheckWhatsAppRequest struct {
	PhoneNumbers []string `json:"phoneNumbers" validate:"required,min=1,max=50"`
}

// CheckWhatsAppResponse representa response da verificação
type CheckWhatsAppResponse struct {
	Results []CheckWhatsAppResult `json:"results"`
	Total   int                   `json:"total"`
}

// CheckWhatsAppResult representa resultado individual da verificação
type CheckWhatsAppResult struct {
	PhoneNumber  string `json:"phoneNumber"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	JID          string `json:"jid,omitempty"`
	Error        string `json:"error,omitempty"`
}

// GetProfilePictureResponse representa response da foto de perfil
type GetProfilePictureResponse struct {
	JID        string `json:"jid"`
	PictureURL string `json:"pictureUrl,omitempty"`
	PictureID  string `json:"pictureId,omitempty"`
	HasPicture bool   `json:"hasPicture"`
}

// GetUserInfoRequest representa request para buscar informações de usuário
type GetUserInfoRequest struct {
	JIDs []string `json:"jids" validate:"required,min=1,max=20"`
}

// GetUserInfoResponse representa response das informações de usuário
type GetUserInfoResponse struct {
	Results []UserInfoResult `json:"results"`
	Total   int              `json:"total"`
}

// UserInfoResult representa informações de um usuário
type UserInfoResult struct {
	JID          string `json:"jid"`
	Name         string `json:"name,omitempty"`
	Status       string `json:"status,omitempty"`
	PictureID    string `json:"pictureId,omitempty"`
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	IsBusiness   bool   `json:"isBusiness"`
	Error        string `json:"error,omitempty"`
}

// ListContactsResponse representa response da listagem de contatos
type ListContactsResponse struct {
	Contacts []ContactInfo `json:"contacts"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

// ContactInfo representa informações de um contato
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

// SyncContactsResponse representa response da sincronização de contatos
type SyncContactsResponse struct {
	SyncedContacts int    `json:"syncedContacts"`
	TotalContacts  int    `json:"totalContacts"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// BusinessProfileResponse representa response do perfil business
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

// ===== CONTACT ENDPOINTS =====

// CheckWhatsApp verifica se números de telefone estão no WhatsApp
// @Summary Check WhatsApp numbers
// @Description Check if phone numbers are registered on WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body CheckWhatsAppRequest true "Phone numbers to check"
// @Success 200 {object} shared.APIResponse{data=CheckWhatsAppResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/check [post]
func (h *ContactHandler) CheckWhatsApp(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "check WhatsApp numbers")

	// Extrair session ID da URL
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req dto.CheckWhatsAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Chamar service
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

// GetProfilePicture busca foto de perfil de um contato
// @Summary Get profile picture
// @Description Get profile picture of a contact
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param jid query string true "Contact JID"
// @Success 200 {object} shared.APIResponse{data=GetProfilePictureResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/avatar [get]
func (h *ContactHandler) GetProfilePicture(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get profile picture")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		h.GetWriter().WriteBadRequest(w, "JID is required")
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar integração com wameow para buscar foto de perfil
	// Por enquanto, simular resposta
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

// GetUserInfo busca informações de usuários
// @Summary Get user info
// @Description Get information about WhatsApp users
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body GetUserInfoRequest true "User JIDs to get info"
// @Success 200 {object} shared.APIResponse{data=GetUserInfoResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/info [post]
func (h *ContactHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get user info")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse request body
	var req dto.GetUserInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Invalid request body")
		return
	}

	// Validar request
	if err := h.GetValidator().ValidateStruct(&req); err != nil {
		h.GetWriter().WriteBadRequest(w, "Validation failed", err.Error())
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Chamar service
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

// ListContacts lista contatos com paginação e filtros
// @Summary List contacts
// @Description List contacts with pagination and filters
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Param search query string false "Search term"
// @Success 200 {object} shared.APIResponse{data=ListContactsResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts [get]
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "list contacts")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Parse query parameters
	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)
	search := r.URL.Query().Get("search")

	// Validar parâmetros
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Criar request para o service
	req := &dto.ListContactsRequest{
		Limit:  limit,
		Offset: offset,
	}

	// Chamar service
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

// SyncContacts sincroniza contatos do dispositivo
// @Summary Sync contacts
// @Description Sync contacts from the device
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse{data=SyncContactsResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/sync [post]
func (h *ContactHandler) SyncContacts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "sync contacts")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// Criar request para o service
	req := &dto.SyncContactsRequest{
		Force: false, // Por padrão, não forçar
	}

	// Chamar service
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

// GetBusinessProfile busca perfil business de um contato
// @Summary Get business profile
// @Description Get business profile of a contact
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param jid query string true "Contact JID"
// @Success 200 {object} shared.APIResponse{data=BusinessProfileResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/business [get]
func (h *ContactHandler) GetBusinessProfile(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get business profile")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		h.GetWriter().WriteBadRequest(w, "JID is required")
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar integração com wameow para buscar perfil business
	// Por enquanto, simular resposta
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

// IsOnWhatsApp verifica se números estão no WhatsApp (batch)
// @Summary Check if numbers are on WhatsApp (batch)
// @Description Check if multiple phone numbers are registered on WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body CheckWhatsAppRequest true "Phone numbers to check"
// @Success 200 {object} shared.APIResponse{data=CheckWhatsAppResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/is-on-whatsapp [post]
func (h *ContactHandler) IsOnWhatsApp(w http.ResponseWriter, r *http.Request) {
	// Reutilizar a lógica do CheckWhatsApp
	h.CheckWhatsApp(w, r)
}

// GetAllContacts busca todos os contatos sem paginação
// @Summary Get all contacts
// @Description Get all contacts without pagination
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Success 200 {object} shared.APIResponse{data=[]ContactInfo}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/all [get]
func (h *ContactHandler) GetAllContacts(w http.ResponseWriter, r *http.Request) {
	h.LogRequest(r, "get all contacts")

	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		h.GetWriter().WriteBadRequest(w, "Session ID is required")
		return
	}

	// Verificar se sessão existe
	session, err := h.sessionService.GetSession(r.Context(), sessionID)
	if err != nil {
		h.GetWriter().WriteNotFound(w, "Session not found")
		return
	}

	// TODO: Implementar integração com wameow para buscar todos os contatos
	// Por enquanto, simular resposta
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

// GetProfilePictureInfo busca informações da foto de perfil
// @Summary Get profile picture info
// @Description Get profile picture information of a contact
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param jid query string true "Contact JID"
// @Success 200 {object} shared.APIResponse{data=GetProfilePictureResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/profile-picture-info [get]
func (h *ContactHandler) GetProfilePictureInfo(w http.ResponseWriter, r *http.Request) {
	// Reutilizar a lógica do GetProfilePicture
	h.GetProfilePicture(w, r)
}

// GetDetailedUserInfo busca informações detalhadas de usuários (batch)
// @Summary Get detailed user info (batch)
// @Description Get detailed information about WhatsApp users
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID"
// @Param request body GetUserInfoRequest true "User JIDs to get detailed info"
// @Success 200 {object} shared.APIResponse{data=GetUserInfoResponse}
// @Failure 400 {object} shared.APIResponse
// @Failure 404 {object} shared.APIResponse
// @Failure 500 {object} shared.APIResponse
// @Router /sessions/{sessionId}/contacts/detailed-info [post]
func (h *ContactHandler) GetDetailedUserInfo(w http.ResponseWriter, r *http.Request) {
	// Reutilizar a lógica do GetUserInfo
	h.GetUserInfo(w, r)
}


