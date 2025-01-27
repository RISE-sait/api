package router

import (
	di "api/cmd/server/di"
	course "api/internal/domains/course/infra/http"
	"api/internal/domains/customer"
	facility "api/internal/domains/facility/infra/http"
	identity "api/internal/domains/identity/infra"
	membership "api/internal/domains/membership/infra/http"

	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, container *di.Container) {

	identity.RegisterIdentityRoutes(r, container)
	course.RegisterCourseRoutes(r, container)
	membership.RegisterMembershipRoutes(r, container)
	facility.RegisterFacilityRoutes(r, container)
	customer.RegisterCustomerRoutes(r, container.HubspotService)
}
