package membership

import (
	membership "api/internal/domains/membership/infra"
	db "api/internal/domains/membership/infra/sqlc/generated"

	"github.com/go-chi/chi"
)

func RegisterMembershipRoutes(r chi.Router, queries *db.Queries) {
	membershipsHandler := NewHandler(NewService(
		&membership.MembershipsRepository{
			Queries: queries,
		},
	))

	r.Route("/memberships", func(auth chi.Router) {
		auth.Get("/", membershipsHandler.GetAllMemberships)
		auth.Get("/{id}", membershipsHandler.GetMembershipById)
		auth.Post("/", membershipsHandler.CreateMembership)
		auth.Put("/{id}", membershipsHandler.UpdateMembership)
		auth.Delete("/{id}", membershipsHandler.DeleteMembership)
	})
}
