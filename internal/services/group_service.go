package services

import (
	"context"
	"fmt"

	"zpwoot/internal/core/group"
	"zpwoot/internal/services/shared/dto"
	"zpwoot/internal/services/shared/validation"
	"zpwoot/platform/logger"
)

// GroupService implementa a lógica de aplicação para grupos
type GroupService struct {
	groupCore       group.Service
	groupRepo       group.Repository
	whatsappGateway group.WhatsAppGateway
	logger          *logger.Logger
	validator       *validation.Validator
}

// NewGroupService cria uma nova instância do GroupService
func NewGroupService(
	groupCore group.Service,
	groupRepo group.Repository,
	whatsappGateway group.WhatsAppGateway,
	logger *logger.Logger,
	validator *validation.Validator,
) *GroupService {
	return &GroupService{
		groupCore:       groupCore,
		groupRepo:       groupRepo,
		whatsappGateway: whatsappGateway,
		logger:          logger,
		validator:       validator,
	}
}

// CreateGroup cria um novo grupo WhatsApp
func (s *GroupService) CreateGroup(ctx context.Context, sessionID string, req *dto.CreateGroupRequest) (*dto.CreateGroupResponse, error) {
	s.logger.InfoWithFields("Creating group", map[string]interface{}{
		"session_id":      sessionID,
		"group_name":      req.Name,
		"participants":    len(req.Participants),
		"has_description": req.Description != "",
	})

	// Validar request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Converter para domain request
	domainReq := &group.CreateGroupRequest{
		Name:         req.Name,
		Description:  req.Description,
		Participants: req.Participants,
	}

	// Validar no core domain
	if err := s.groupCore.ValidateGroupCreation(domainReq); err != nil {
		return nil, fmt.Errorf("group validation failed: %w", err)
	}

	// Criar grupo via WhatsApp Gateway
	groupInfo, err := s.whatsappGateway.CreateGroup(ctx, sessionID, req.Name, req.Participants, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create group in WhatsApp: %w", err)
	}

	// Salvar no banco de dados
	groupModel := s.convertGroupInfoToModel(groupInfo, sessionID)
	if err := s.groupRepo.Create(ctx, groupModel); err != nil {
		s.logger.ErrorWithFields("Failed to save group to database", map[string]interface{}{
			"session_id": sessionID,
			"group_jid":  groupInfo.GroupJID,
			"error":      err.Error(),
		})
		// Não retornar erro aqui pois o grupo foi criado no WhatsApp
	}

	response := &dto.CreateGroupResponse{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Participants: req.Participants,
		CreatedAt:    groupInfo.CreatedAt,
		Success:      true,
		Message:      "Group created successfully",
	}

	s.logger.InfoWithFields("Group created successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupInfo.GroupJID,
		"group_name": groupInfo.Name,
	})

	return response, nil
}

// ListGroups lista todos os grupos de uma sessão
func (s *GroupService) ListGroups(ctx context.Context, sessionID string) (*dto.ListGroupsResponse, error) {
	s.logger.InfoWithFields("Listing groups", map[string]interface{}{
		"session_id": sessionID,
	})

	// Buscar grupos via WhatsApp Gateway
	groupInfos, err := s.whatsappGateway.ListJoinedGroups(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups from WhatsApp: %w", err)
	}

	// Converter para response
	groups := make([]dto.GroupInfo, len(groupInfos))
	for i, groupInfo := range groupInfos {
		groups[i] = dto.GroupInfo{
			GroupJID:     groupInfo.GroupJID,
			Name:         groupInfo.Name,
			Description:  groupInfo.Description,
			Owner:        groupInfo.Owner,
			Participants: len(groupInfo.Participants),
			CreatedAt:    groupInfo.CreatedAt,
		}
	}

	response := &dto.ListGroupsResponse{
		Groups:  groups,
		Count:   len(groups),
		Success: true,
		Message: "Groups retrieved successfully",
	}

	s.logger.InfoWithFields("Groups listed successfully", map[string]interface{}{
		"session_id":   sessionID,
		"group_count":  len(groups),
	})

	return response, nil
}

// GetGroupInfo obtém informações detalhadas de um grupo
func (s *GroupService) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*dto.GetGroupInfoResponse, error) {
	s.logger.InfoWithFields("Getting group info", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  groupJID,
	})

	// Buscar informações via WhatsApp Gateway
	groupInfo, err := s.whatsappGateway.GetGroupInfo(ctx, sessionID, groupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info from WhatsApp: %w", err)
	}

	// Converter participantes
	participants := make([]dto.ParticipantInfo, len(groupInfo.Participants))
	for i, p := range groupInfo.Participants {
		participants[i] = dto.ParticipantInfo{
			JID:      p.JID,
			Role:     string(p.Role),
			JoinedAt: p.JoinedAt,
			Status:   string(p.Status),
		}
	}

	response := &dto.GetGroupInfoResponse{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Owner:        groupInfo.Owner,
		Participants: participants,
		Settings: dto.GroupSettings{
			Announce:         groupInfo.Settings.Announce,
			Restrict:         groupInfo.Settings.Restrict,
			JoinApprovalMode: groupInfo.Settings.JoinApprovalMode,
			MemberAddMode:    groupInfo.Settings.MemberAddMode,
			Locked:           groupInfo.Settings.Locked,
		},
		CreatedAt: groupInfo.CreatedAt,
		UpdatedAt: groupInfo.UpdatedAt,
		Success:   true,
		Message:   "Group info retrieved successfully",
	}

	s.logger.InfoWithFields("Group info retrieved successfully", map[string]interface{}{
		"session_id":      sessionID,
		"group_jid":       groupJID,
		"group_name":      groupInfo.Name,
		"participant_count": len(participants),
	})

	return response, nil
}

// UpdateGroupParticipants gerencia participantes do grupo
func (s *GroupService) UpdateGroupParticipants(ctx context.Context, sessionID string, req *dto.UpdateParticipantsRequest) (*dto.UpdateParticipantsResponse, error) {
	s.logger.InfoWithFields("Updating group participants", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	// Validar request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Buscar informações do grupo para validação
	groupInfo, err := s.whatsappGateway.GetGroupInfo(ctx, sessionID, req.GroupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	// Validar mudanças no core domain
	domainReq := &group.UpdateParticipantsRequest{
		GroupJID:     req.GroupJID,
		Action:       req.Action,
		Participants: req.Participants,
	}

	if err := s.groupCore.ProcessParticipantChanges(domainReq, groupInfo); err != nil {
		return nil, fmt.Errorf("participant changes validation failed: %w", err)
	}

	// Executar ação via WhatsApp Gateway
	switch req.Action {
	case "add":
		err = s.whatsappGateway.AddParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	case "remove":
		err = s.whatsappGateway.RemoveParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	case "promote":
		err = s.whatsappGateway.PromoteParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	case "demote":
		err = s.whatsappGateway.DemoteParticipants(ctx, sessionID, req.GroupJID, req.Participants)
	default:
		return nil, fmt.Errorf("invalid action: %s", req.Action)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to %s participants: %w", req.Action, err)
	}

	response := &dto.UpdateParticipantsResponse{
		GroupJID:     req.GroupJID,
		Action:       req.Action,
		Participants: req.Participants,
		Success:      true,
		Message:      fmt.Sprintf("Participants %s successfully", req.Action),
	}

	s.logger.InfoWithFields("Group participants updated successfully", map[string]interface{}{
		"session_id":   sessionID,
		"group_jid":    req.GroupJID,
		"action":       req.Action,
		"participants": len(req.Participants),
	})

	return response, nil
}

// SetGroupName altera o nome do grupo
func (s *GroupService) SetGroupName(ctx context.Context, sessionID string, req *dto.SetGroupNameRequest) (*dto.SetGroupNameResponse, error) {
	s.logger.InfoWithFields("Setting group name", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"new_name":   req.Name,
	})

	// Validar request
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validar nome no core domain
	if err := s.groupCore.ValidateGroupName(req.Name); err != nil {
		return nil, fmt.Errorf("group name validation failed: %w", err)
	}

	// Alterar nome via WhatsApp Gateway
	if err := s.whatsappGateway.SetGroupName(ctx, sessionID, req.GroupJID, req.Name); err != nil {
		return nil, fmt.Errorf("failed to set group name: %w", err)
	}

	response := &dto.SetGroupNameResponse{
		GroupJID: req.GroupJID,
		Name:     req.Name,
		Success:  true,
		Message:  "Group name updated successfully",
	}

	s.logger.InfoWithFields("Group name updated successfully", map[string]interface{}{
		"session_id": sessionID,
		"group_jid":  req.GroupJID,
		"new_name":   req.Name,
	})

	return response, nil
}

// convertGroupInfoToModel converte GroupInfo para modelo de domínio
func (s *GroupService) convertGroupInfoToModel(groupInfo *group.GroupInfo, sessionID string) *group.Group {
	// TODO: Implementar conversão completa
	// Por enquanto, retorna um modelo básico
	return &group.Group{
		GroupJID:     groupInfo.GroupJID,
		Name:         groupInfo.Name,
		Description:  groupInfo.Description,
		Owner:        groupInfo.Owner,
		Settings:     groupInfo.Settings,
		Participants: groupInfo.Participants,
		CreatedAt:    groupInfo.CreatedAt,
		UpdatedAt:    groupInfo.UpdatedAt,
	}
}