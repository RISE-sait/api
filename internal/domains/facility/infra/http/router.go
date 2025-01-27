package facility

import (
	"api/cmd/server/di"
	facility "api/internal/domains/facility/application"
	repository "api/internal/domains/facility/infra/persistence"

	"github.com/go-chi/chi"
)

func RegisterFacilityRoutes(r chi.Router, container *di.Container) {
	facilitiesHandler := NewHandler(facility.NewFacilityService(
		&repository.FacilityRepository{
			Queries: container.Queries.FacilityDb,
		},
	))

	r.Route("/facilities", func(auth chi.Router) {
		auth.Get("/", facilitiesHandler.GetAllFacilities)
		auth.Get("/{id}", facilitiesHandler.GetFacilityById)
		auth.Post("/", facilitiesHandler.CreateFacility)
		auth.Put("/{id}", facilitiesHandler.UpdateFacility)
		auth.Delete("/{id}", facilitiesHandler.DeleteFacility)
	})
}
