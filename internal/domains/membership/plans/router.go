package membership_plan

import (
	plans "api/internal/domains/membership/plans/infra"
	db "api/internal/domains/membership/plans/infra/sqlc/generated"

	"github.com/go-chi/chi"
)

func RegisterMembershipRoutes(r chi.Router, queries *db.Queries) {
	membershipPlansHandler := NewHandler(NewService(
		&plans.Repo{
			Queries: queries,
		},
	))

	r.Route("/memberships/plans", func(auth chi.Router) {
		auth.Get("/", membershipPlansHandler.GetMembershipPlanDetails)
		auth.Post("/", membershipPlansHandler.CreateMembershipPlan)
		auth.Put("/", membershipPlansHandler.UpdateMembershipPlan)
		auth.Delete("/", membershipPlansHandler.DeleteMembershipPlan)
	})
}
