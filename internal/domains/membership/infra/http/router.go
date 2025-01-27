package membership

import (
	"api/cmd/server/di"
	membership "api/internal/domains/membership/application"

	membershipRepo "api/internal/domains/membership/infra/persistence"
	membershipPlan "api/internal/domains/membership/plans/infra/http"

	"github.com/go-chi/chi"
)

func RegisterMembershipRoutes(r chi.Router, container *di.Container) {
	membershipsHandler := NewHandler(membership.NewMembershipService(
		&membershipRepo.MembershipsRepository{
			Queries: container.Queries.MembershipDb,
		},
	))

	r.Route("/memberships", func(r chi.Router) {
		r.Get("/", membershipsHandler.GetAllMemberships)
		r.Get("/{id}", membershipsHandler.GetMembershipById)
		r.Post("/", membershipsHandler.CreateMembership)
		r.Put("/{id}", membershipsHandler.UpdateMembership)
		r.Delete("/{id}", membershipsHandler.DeleteMembership)

		membershipPlan.RegisterMembershipPlansRoutes(r, container.Queries.MembershipPlanDb)
	})
}
