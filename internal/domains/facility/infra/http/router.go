package facility

import (
	facility "api/internal/domains/facility/application"
	repository "api/internal/domains/facility/infra/persistence"
	db "api/internal/domains/facility/infra/persistence/sqlc/generated"

	"github.com/go-chi/chi"
)

func RegisterFacilityRoutes(r chi.Router, queries *db.Queries) {
	facilitiesHandler := NewHandler(facility.NewFacilityService(
		&repository.FacilityRepository{
			Queries: queries,
		},
	))

	r.Route("/facilities", func(auth chi.Router) {
		auth.Get("/", facilitiesHandler.GetAllFacilities)
		auth.Get("/{id}", facilitiesHandler.GetFacilityById)
		auth.Post("/", facilitiesHandler.CreateFacility)
		auth.Put("/", facilitiesHandler.UpdateFacility)
		auth.Delete("/{id}", facilitiesHandler.DeleteFacility)
	})
}
