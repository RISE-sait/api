package membership

import (
	membership "api/internal/domains/membership/application"
	membershipDb "api/internal/domains/membership/infra/persistence/sqlc/generated"
	plansDb "api/internal/domains/membership/plans/infra/persistence/sqlc/generated"

	membershipRepo "api/internal/domains/membership/infra/persistence"
	membershipPlan "api/internal/domains/membership/plans/infra/http"

	"github.com/go-chi/chi"
)

func RegisterMembershipRoutes(r chi.Router, membershipQueries *membershipDb.Queries, planQueries *plansDb.Queries) {
	membershipsHandler := NewHandler(membership.NewMembershipService(
		&membershipRepo.MembershipsRepository{
			Queries: membershipQueries,
		},
	))

	r.Route("/memberships", func(r chi.Router) {
		r.Get("/", membershipsHandler.GetAllMemberships)
		r.Get("/{id}", membershipsHandler.GetMembershipById)
		r.Post("/", membershipsHandler.CreateMembership)
		r.Put("/{id}", membershipsHandler.UpdateMembership)
		r.Delete("/{id}", membershipsHandler.DeleteMembership)

		membershipPlan.RegisterMembershipPlansRoutes(r, planQueries)
	})
}
