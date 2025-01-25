package router

import (
	_interface "api/cmd/server/interface"
	course "api/internal/domains/course/infra/http"
	"api/internal/domains/customer"
	facility "api/internal/domains/facility/infra/http"
	"api/internal/domains/identity"
	membership "api/internal/domains/membership/infra/http"
	"api/internal/services/hubspot"

	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, queries _interface.QueriesType) {

	hubspotService := hubspot.GetHubSpotService()

	identity.RegisterIdentityRoutes(r, queries.IdentityDb)
	course.RegisterCourseRoutes(r, queries.CoursesDb)
	membership.RegisterMembershipRoutes(r, queries.MembershipDb, queries.MembershipPlanDb)
	facility.RegisterFacilityRoutes(r, queries.FacilityDb)
	customer.RegisterCustomerRoutes(r, hubspotService)
}
