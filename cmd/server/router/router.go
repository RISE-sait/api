package router

import (
	_interface "api/cmd/server/interface"
	"api/internal/domains/course"
	"api/internal/domains/identity"
	"api/internal/domains/membership"
	membership_plan "api/internal/domains/membership/plans"

	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, queries _interface.QueriesType) {

	identity.RegisterIdentityRoutes(r, queries.IdentityDb)
	course.RegisterCourseRoutes(r, queries.CoursesDb)
	membership.RegisterMembershipRoutes(r, queries.MembershipDb)
	membership_plan.RegisterMembershipRoutes(r, queries.MembershipPlanDb)
}
