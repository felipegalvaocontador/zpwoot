package router

import (
	"github.com/go-chi/chi/v5"

	"zpwoot/internal/adapters/server/handler"
	"zpwoot/internal/services"
	"zpwoot/platform/logger"
)

func setupGroupRoutes(r chi.Router, groupService *services.GroupService, sessionService *services.SessionService, appLogger *logger.Logger) {
	groupHandler := handler.NewGroupHandler(groupService, sessionService, appLogger)

	r.Route("/{sessionId}/groups", func(r chi.Router) {

		r.Post("/", groupHandler.CreateGroup)
		r.Get("/", groupHandler.ListGroups)
		r.Get("/info", groupHandler.GetGroupInfo)

		r.Post("/participants", groupHandler.UpdateGroupParticipants)

		r.Put("/name", groupHandler.SetGroupName)
		r.Put("/description", groupHandler.SetGroupDescription)
		r.Put("/photo", groupHandler.SetGroupPhoto)

		r.Get("/invite-link", groupHandler.GetGroupInviteLink)
		r.Post("/join-via-link", groupHandler.JoinGroupViaLink)
		r.Post("/leave", groupHandler.LeaveGroup)

		r.Put("/settings", groupHandler.UpdateGroupSettings)
		r.Get("/request-participants", groupHandler.GetGroupRequestParticipants)
		r.Post("/request-participants", groupHandler.UpdateGroupRequestParticipants)
		r.Put("/join-approval-mode", groupHandler.SetGroupJoinApprovalMode)
		r.Put("/member-add-mode", groupHandler.SetGroupMemberAddMode)

		r.Get("/info-from-link", groupHandler.GetGroupInfoFromLink)
		r.Post("/info-from-invite", groupHandler.GetGroupInfoFromInvite)
		r.Post("/join-with-invite", groupHandler.JoinGroupWithInvite)
	})
}
