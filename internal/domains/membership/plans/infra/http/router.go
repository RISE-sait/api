package membershipPlan

import (
	membershipPlan "api/internal/domains/membership/plans/application"
	repo "api/internal/domains/membership/plans/infra/persistence"

	db "api/internal/domains/membership/plans/infra/persistence/sqlc/generated"

	"github.com/go-chi/chi"
)

func RegisterMembershipPlansRoutes(r chi.Router, queries *db.Queries) {
	membershipPlansHandler := NewHandler(membershipPlan.NewFacilityManager(
		&repo.MembershipPlansRepository{
			Queries: queries,
		},
	))

	r.Route("/plans", func(auth chi.Router) {
		auth.Get("/", membershipPlansHandler.GetMembershipPlansByMembershipId)
		auth.Post("/", membershipPlansHandler.CreateMembershipPlan)
		auth.Put("/", membershipPlansHandler.UpdateMembershipPlan)
		auth.Delete("/", membershipPlansHandler.DeleteMembershipPlan)
	})
}
