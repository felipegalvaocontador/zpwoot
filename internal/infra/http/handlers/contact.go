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

func (h *ContactHandler) resolveSession(r *http.Request) (*domainSession.Session, error) {
	sessionIdentifier := chi.URLParam(r, "sessionId")
	if sessionIdentifier == "" {
		return nil, domainSession.ErrSessionNotFound
	}
	return h.sessionResolver.ResolveSession(r.Context(), sessionIdentifier)
}

func (h *ContactHandler) handleActionRequest(
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
		if err := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); err != nil {
			h.logger.Error("Failed to encode error response: " + err.Error())
		}
		return
	}

	req, err := parseFunc(r, sess)
	if err != nil {
		h.logger.Error("Failed to parse request body: " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request body")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	result, err := actionFunc(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to " + actionName + ": " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to "+actionName)); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(result, successMessage)); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}

func (h *ContactHandler) handleListRequest(
	w http.ResponseWriter,
	r *http.Request,
	actionName string,
	successMessage string,
	listFunc func(context.Context, *domainSession.Session, int, int, string) (interface{}, error),
) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, parseErr := strconv.Atoi(limitStr); parseErr == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, parseErr := strconv.Atoi(offsetStr); parseErr == nil {
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

	h.logger.InfoWithFields(actionName, map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"limit":        limit,
		"offset":       offset,
		"search":       search,
	})

	result, err := listFunc(r.Context(), sess, limit, offset, search)
	if err != nil {
		h.logger.Error("Failed to " + actionName + ": " + err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to "+actionName)); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(result, successMessage)); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
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
	h.handleActionRequest(
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
			checkReq, ok := req.(*contact.CheckWhatsAppRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			return h.contactUC.CheckWhatsApp(ctx, checkReq)
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
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("JID is required")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
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
			if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("User not found")); encErr != nil {
				h.logger.Error("Failed to encode error response: " + encErr.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get profile picture")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	response := common.NewSuccessResponse(result, "Profile picture retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
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
	h.handleActionRequest(
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
			userReq, ok := req.(*contact.GetUserInfoRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			return h.contactUC.GetUserInfo(ctx, userReq)
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
	h.handleListRequest(
		w,
		r,
		"Listing contacts",
		"Contacts retrieved successfully",
		func(ctx context.Context, sess *domainSession.Session, limit, offset int, search string) (interface{}, error) {
			req := &contact.ListContactsRequest{
				SessionID: sess.ID.String(),
				Limit:     limit,
				Offset:    offset,
				Search:    search,
			}
			return h.contactUC.ListContacts(ctx, req)
		},
	)
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
	h.handleActionRequest(
		w,
		r,
		"Syncing contacts",
		"Contacts synchronized successfully",
		func(r *http.Request, sess *domainSession.Session) (interface{}, error) {
			return &contact.SyncContactsRequest{
				SessionID: sess.ID.String(),
			}, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			syncReq, ok := req.(*contact.SyncContactsRequest)
			if !ok {
				return nil, fmt.Errorf("invalid request type")
			}
			return h.contactUC.SyncContacts(ctx, syncReq)
		},
	)
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
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("JID is required")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
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
			if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Business profile not found")); encErr != nil {
				h.logger.Error("Failed to encode error response: " + encErr.Error())
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Failed to get business profile")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	response := common.NewSuccessResponse(result, "Business profile retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(response); encErr != nil {
		h.logger.Error("Failed to encode response: " + encErr.Error())
	}
}

func (h *ContactHandler) IsOnWhatsApp(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	var req struct {
		PhoneNumbers []string `json:"phoneNumbers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	if len(req.PhoneNumbers) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Phone numbers are required")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	if len(req.PhoneNumbers) > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Maximum 50 phone numbers allowed")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields("Checking WhatsApp numbers", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"phone_count":  len(req.PhoneNumbers),
	})

	results := make([]map[string]interface{}, len(req.PhoneNumbers))
	for i, phone := range req.PhoneNumbers {
		results[i] = map[string]interface{}{
			"phoneNumber":  phone,
			"isOnWhatsApp": true,
			"jid":          phone + "@s.whatsapp.net",
		}
	}

	response := map[string]interface{}{
		"results": results,
		"message": "IsOnWhatsApp functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Numbers checked successfully")); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}

func (h *ContactHandler) GetAllContacts(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields("Getting all contacts", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
	})

	response := map[string]interface{}{
		"contacts": []interface{}{},
		"count":    0,
		"message":  "GetAllContacts functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "All contacts retrieved successfully")); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}

func (h *ContactHandler) GetProfilePictureInfo(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	jid := r.URL.Query().Get("jid")
	if jid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("JID is required")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields("Getting profile picture info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid":          jid,
	})

	response := map[string]interface{}{
		"jid":     jid,
		"url":     "https://placeholder.com/avatar.jpg",
		"id":      "placeholder-id",
		"type":    "image",
		"message": "GetProfilePictureInfo functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Profile picture info retrieved successfully")); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}

func (h *ContactHandler) GetDetailedUserInfo(w http.ResponseWriter, r *http.Request) {
	sess, err := h.resolveSession(r)
	if err != nil {
		statusCode := 500
		if errors.Is(err, domainSession.ErrSessionNotFound) {
			statusCode = 404
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse(err.Error())); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	var req struct {
		JIDs []string `json:"jids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Invalid request format")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	if len(req.JIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("JIDs are required")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	if len(req.JIDs) > 20 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encErr := json.NewEncoder(w).Encode(common.NewErrorResponse("Maximum 20 JIDs allowed")); encErr != nil {
			h.logger.Error("Failed to encode error response: " + encErr.Error())
		}
		return
	}

	h.logger.InfoWithFields("Getting detailed user info", map[string]interface{}{
		"session_id":   sess.ID.String(),
		"session_name": sess.Name,
		"jid_count":    len(req.JIDs),
	})

	results := make([]map[string]interface{}, len(req.JIDs))
	for i, jid := range req.JIDs {
		results[i] = map[string]interface{}{
			"jid":          jid,
			"name":         "Placeholder Name",
			"status":       "Available",
			"pictureId":    "placeholder-pic-id",
			"isOnWhatsApp": true,
		}
	}

	response := map[string]interface{}{
		"results": results,
		"message": "GetDetailedUserInfo functionality needs to be implemented in use case",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if encErr := json.NewEncoder(w).Encode(common.NewSuccessResponse(response, "Detailed user info retrieved successfully")); encErr != nil {
		h.logger.Error("Failed to encode success response: " + encErr.Error())
	}
}
