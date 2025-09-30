package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"zpwoot/internal/app/common"
	"zpwoot/internal/app/contact"
	domainSession "zpwoot/internal/domain/session"
	"zpwoot/internal/infra/http/helpers"
	"zpwoot/platform/logger"
)

type ContactHandler struct {
	logger          *logger.Logger
	contactUC       contact.UseCase
	sessionResolver *helpers.SessionResolver
}

func NewContactHandler(appLogger *logger.Logger, contactUC contact.UseCase, sessionRepo helpers.SessionRepository) *ContactHandler {
	return &ContactHandler{
		logger:          appLogger,
		contactUC:       contactUC,
		sessionResolver: helpers.NewSessionResolver(appLogger, sessionRepo),
	}
}

// handleContactAction handles common contact action logic
func (h *ContactHandler) handleContactAction(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	parseFunc func(*http.Request, *domainSession.Session) (interface{}, error),
	actionFunc func(context.Context, interface{}) (interface{}, error),
) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	req, err := parseFunc(r, sess)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body"))
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(r.Context(), req)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to %s: %s", actionName, err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse(fmt.Sprintf("Failed to %s", actionName)))
		return
	}

	response := common.NewSuccessResponse(result, successMessage)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Check if phone numbers are on WhatsApp
// @Description Check if one or more phone numbers are registered on WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body contact.CheckWhatsAppRequest true "Phone numbers to check"
// @Success 200 {object} common.SuccessResponse{data=contact.CheckWhatsAppResponse} "Phone numbers checked successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/check [post]
func (h *ContactHandler) CheckWhatsApp(w http.ResponseWriter, r *http.Request) {
	h.handleContactAction(
		w,
		r,
		"Checking WhatsApp numbers",
		"Phone numbers checked successfully",
		func(r *http.Request, sess *domainSession.Session) (interface{}, error) {
			var req contact.CheckWhatsAppRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			req.SessionID = sess.ID.String()
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleContactAction wrapper
			return h.contactUC.CheckWhatsApp(ctx, req.(*contact.CheckWhatsAppRequest))
		},
	)
}

// @Summary Get profile picture
// @Description Get profile picture URL and metadata for a WhatsApp user
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param jid query string true "WhatsApp JID" example("5511999999999@s.whatsapp.net")
// @Param preview query bool false "Get preview (low resolution) image" default(false)
// @Success 200 {object} common.SuccessResponse{data=contact.ProfilePictureResponse} "Profile picture retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or user not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/avatar [get]
func (h *ContactHandler) GetProfilePicture(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("JID is required"))
		return
	}

	preview := r.URL.Query().Get("preview") == "true"

	h.logger.InfoWithFields("Getting profile picture", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
		"preview":      preview,
	})

	req := &contact.GetProfilePictureRequest{
		SessionID: sess.ID.String(),
		JID:       jid,
		Preview:   preview,
	}

	result, err := h.contactUC.GetProfilePicture(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get profile picture: " + err.Error())
		if err.Error() == "user not found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(common.NewErrorResponse("User not found"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get profile picture"))
		return
	}

	response := common.NewSuccessResponse(result, "Profile picture retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get user information
// @Description Get detailed information about WhatsApp users
// @Tags Contacts
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param request body contact.GetUserInfoRequest true "User JIDs to get info for"
// @Success 200 {object} common.SuccessResponse{data=contact.GetUserInfoResponse} "User information retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/info [post]
func (h *ContactHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	h.handleContactAction(
		w,
		r,
		"Getting user info",
		"User information retrieved successfully",
		func(r *http.Request, sess *domainSession.Session) (interface{}, error) {
			var req contact.GetUserInfoRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return nil, err
			}
			req.SessionID = sess.ID.String()
			return &req, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			//nolint:errcheck // Error is handled by handleContactAction wrapper
			return h.contactUC.GetUserInfo(ctx, req.(*contact.GetUserInfoRequest))
		},
	)
}

// @Summary List contacts
// @Description List all contacts from the WhatsApp account
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Param search query string false "Search contacts by name or phone"
// @Success 200 {object} common.SuccessResponse{data=contact.ListContactsResponse} "Contacts retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts [get]
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	// Parse query parameters
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	search := r.URL.Query().Get("search")

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	h.logger.InfoWithFields("Listing contacts", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"limit":        limit,
		"offset":       offset,
		"search":       search,
	})

	req := &contact.ListContactsRequest{
		SessionID: sess.ID.String(),
		Limit:     limit,
		Offset:    offset,
		Search:    search,
	}

	result, err := h.contactUC.ListContacts(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list contacts: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to list contacts"))
		return
	}

	response := common.NewSuccessResponse(result, "Contacts retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Sync contacts
// @Description Synchronize contacts from the device with WhatsApp
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Success 200 {object} common.SuccessResponse{data=contact.SyncContactsResponse} "Contacts synchronized successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/sync [post]
func (h *ContactHandler) SyncContacts(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Syncing contacts", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	req := &contact.SyncContactsRequest{
		SessionID: sess.ID.String(),
	}

	result, err := h.contactUC.SyncContacts(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to sync contacts: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to sync contacts"))
		return
	}

	response := common.NewSuccessResponse(result, "Contacts synchronized successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get business profile
// @Description Get business profile information for a WhatsApp Business account
// @Tags Contacts
// @Security ApiKeyAuth
// @Produce json
// @Param sessionId path string true "Session ID or Name" example("mySession")
// @Param jid query string true "WhatsApp JID" example("5511999999999@s.whatsapp.net")
// @Success 200 {object} common.SuccessResponse{data=contact.BusinessProfileResponse} "Business profile retrieved successfully"
// @Failure 400 {object} object "Bad Request"
// @Failure 404 {object} object "Session or business not found"
// @Failure 500 {object} object "Internal Server Error"
// @Router /sessions/{sessionId}/contacts/business [get]
func (h *ContactHandler) GetBusinessProfile(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("JID is required"))
		return
	}

	h.logger.InfoWithFields("Getting business profile", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
	})

	req := &contact.GetBusinessProfileRequest{
		SessionID: sess.ID.String(),
		JID:       jid,
	}

	result, err := h.contactUC.GetBusinessProfile(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get business profile: " + err.Error())
		if err.Error() == "business not found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(common.NewErrorResponse("Business profile not found"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get business profile"))
		return
	}

	response := common.NewSuccessResponse(result, "Business profile retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ContactHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
	idOrName := chi.URLParam(r, "sessionId")

	sess, err := h.sessionResolver.ResolveSession(r.Context(), idOrName)
	if err != nil {
		h.logger.WarnWithFields("Failed to resolve session", map[string]interface{}{
			"identifier": idOrName,
			"error":      err.Error(),
			"path":       r.URL.Path,
		})

		return nil, err
	}

	return sess, nil
}

// IsOnWhatsApp checks if phone numbers are registered on WhatsApp
func (h *ContactHandler) IsOnWhatsApp(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req struct {
		PhoneNumbers []string `json:"phoneNumbers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if len(req.PhoneNumbers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Phone numbers are required"))
		return
	}

	if len(req.PhoneNumbers) > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Maximum 50 phone numbers allowed"))
		return
	}

	h.logger.InfoWithFields("Checking WhatsApp numbers", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"phone_count":  len(req.PhoneNumbers),
	})

	// For now, return placeholder response until implemented in use case
	results := make([]map[string]interface{}, len(req.PhoneNumbers))
	for i, phone := range req.PhoneNumbers {
		results[i] = map[string]interface{}{
			"phoneNumber": phone,
			"isOnWhatsApp": true, // placeholder
			"jid":         phone + "@s.whatsapp.net",
		}
	}

	response := map[string]interface{}{
		"results": results,
		"message": "IsOnWhatsApp functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Numbers checked successfully"))
}

// GetAllContacts gets all contacts
func (h *ContactHandler) GetAllContacts(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	h.logger.InfoWithFields("Getting all contacts", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	// For now, return placeholder response until implemented in use case
	response := map[string]interface{}{
		"contacts": []interface{}{},
		"count":    0,
		"message":  "GetAllContacts functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "All contacts retrieved successfully"))
}

// GetProfilePictureInfo gets profile picture information
func (h *ContactHandler) GetProfilePictureInfo(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("JID is required"))
		return
	}

	h.logger.InfoWithFields("Getting profile picture info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
	})

	// For now, return placeholder response until implemented in use case
	response := map[string]interface{}{
		"jid":     jid,
		"url":     "https://placeholder.com/avatar.jpg",
		"id":      "placeholder-id",
		"type":    "image",
		"message": "GetProfilePictureInfo functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Profile picture info retrieved successfully"))
}

// GetDetailedUserInfo gets detailed user information
func (h *ContactHandler) GetDetailedUserInfo(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error()))
		return
	}

	var req struct {
		JIDs []string `json:"jids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format"))
		return
	}

	if len(req.JIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("JIDs are required"))
		return
	}

	if len(req.JIDs) > 20 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.NewErrorResponse("Maximum 20 JIDs allowed"))
		return
	}

	h.logger.InfoWithFields("Getting detailed user info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid_count":    len(req.JIDs),
	})

	// For now, return placeholder response until implemented in use case
	results := make([]map[string]interface{}, len(req.JIDs))
	for i, jid := range req.JIDs {
		results[i] = map[string]interface{}{
			"jid":         jid,
			"name":        "Placeholder Name",
			"status":      "Available",
			"pictureId":   "placeholder-pic-id",
			"isOnWhatsApp": true,
		}
	}

	response := map[string]interface{}{
		"results": results,
		"message": "GetDetailedUserInfo functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Detailed user info retrieved successfully"))
}




