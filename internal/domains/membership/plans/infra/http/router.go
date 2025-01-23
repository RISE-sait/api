package membershipPlan

import (
	membershipPlan "api/internal/domains/membership/plans"
	repo "api/internal/domains/membership/plans/infra/persistence"

	db "api/internal/domains/membership/plans/infra/persistence/sqlc/generated"

	"github.com/go-chi/chi"
)

func RegisterMembershipPlansRoutes(r chi.Router, queries *db.Queries) {
	membershipPlansHandler := NewHandler(membershipPlan.NewService(
		&repo.Repo{
			Queries: queries,
		},
	))

	r.Route("/plans", func(auth chi.Router) {
		auth.Get("/", membershipPlansHandler.GetMembershipPlanDetails)
		auth.Post("/", membershipPlansHandler.CreateMembershipPlan)
		auth.Put("/", membershipPlansHandler.UpdateMembershipPlan)
		auth.Delete("/", membershipPlansHandler.DeleteMembershipPlan)
	})
}
