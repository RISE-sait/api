package router

import (
	_interface "api/cmd/server/interface"
	course "api/internal/domains/course/infra/http"
	"api/internal/domains/customer"
	"api/internal/domains/customer/hubspot"
	facility "api/internal/domains/facility/infra/http"
	"api/internal/domains/identity"
	membership "api/internal/domains/membership/infra/http"
	"api/internal/services"

	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, queries _interface.QueriesType) {

	hubspotService := services.GetHubSpotService()

	identity.RegisterIdentityRoutes(r, queries.IdentityDb)
	course.RegisterCourseRoutes(r, queries.CoursesDb)
	membership.RegisterMembershipRoutes(r, queries.MembershipDb, queries.MembershipPlanDb)
	facility.RegisterFacilityRoutes(r, queries.FacilityDb)
	customer.RegisterCustomerRoutes(r, &hubspot.HubSpotCustomersService{HubSpotService: hubspotService})
}
