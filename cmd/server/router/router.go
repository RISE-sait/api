package routes

import (
	"api/cmd/server/di"
	"api/internal/domains/course"
	"api/internal/domains/facility"
	identity "api/internal/domains/identity/controllers"
	membership "api/internal/domains/membership"
	membership_plan "api/internal/domains/membership/plans"
	"api/internal/domains/schedule"

	"github.com/go-chi/chi"
)

type RouteConfig struct {
	Path      string
	Configure func(chi.Router)
}

func RegisterRoutes(router *chi.Mux, container *di.Container) {

	router.Route("/api", func(r chi.Router) {
		routes := []RouteConfig{
			{Path: "/memberships", Configure: RegisterMembershipRoutes(r, container)},
			{Path: "/identity", Configure: RegisterIdentityRoutes(r, container)},
			{Path: "/courses", Configure: RegisterCourseRoutes(r, container)},
			{Path: "/schedules", Configure: RegisterScheduleRoutes(r, container)},
			{Path: "/facilities", Configure: RegisterFacilityRoutes(r, container)},
		}

		for _, route := range routes {
			r.Route(route.Path, route.Configure)
		}
	})
}

func RegisterMembershipRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := membership.NewMembershipController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllMemberships)
		r.Get("/{id}", ctrl.GetMembershipById)
		r.Post("/", ctrl.CreateMembership)
		r.Put("/{id}", ctrl.UpdateMembership)
		r.Delete("/{id}", ctrl.DeleteMembership)

		// plans as subroute
		r.Route("/{membershipId}/plans", RegisterMembershipPlansRoutes(r, container))
	}
}

func RegisterMembershipPlansRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := membership_plan.NewMembershipPlansController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetMembershipPlansByMembershipId)
		r.Post("/", ctrl.CreateMembershipPlan)
		r.Put("/{planId}", ctrl.UpdateMembershipPlan)
		r.Delete("/{planId}", ctrl.DeleteMembershipPlan)
	}
}

func RegisterFacilityRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := facility.NewController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllFacilities)
		r.Get("/{id}", ctrl.GetFacilityById)
		r.Post("/", ctrl.CreateFacility)
		r.Put("/{id}", ctrl.UpdateFacility)
		r.Delete("/{id}", ctrl.DeleteFacility)
	}
}

func RegisterCourseRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := course.NewCourseController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllCourses)
		r.Get("/{id}", ctrl.GetCourseById)
		r.Post("/", ctrl.CreateCourse)
		r.Put("/{id}", ctrl.UpdateCourse)
		r.Delete("/{id}", ctrl.DeleteCourse)
	}
}

func RegisterScheduleRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := schedule.NewSchedulesController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetSchedules)
		r.Post("/", ctrl.CreateSchedule)
		r.Put("/{id}", ctrl.UpdateSchedule)
		r.Delete("/{id}", ctrl.DeleteSchedule)
	}
}

func RegisterIdentityRoutes(r chi.Router, container *di.Container) func(chi.Router) {

	authController := identity.NewAuthenticationController(container)

	OauthController := identity.NewOauthController(container)

	customerRegistrationCtrl := identity.NewCustomerRegistrationController(container)

	childRegistrationCtrl := identity.NewCreatePendingChildAccountController(container)

	confirmChildCtrl := identity.NewChildAccountConfirmationController(container)

	return func(r chi.Router) {

		r.Route("/auth", func(auth chi.Router) {
			auth.Post("/traditional", authController.Login)
			auth.Post("/oauth/google", OauthController.HandleOAuthCallback)
		})

		r.Route("/register", func(registration chi.Router) {
			registration.Post("/", customerRegistrationCtrl.CreateCustomer)
			registration.Post("/child", childRegistrationCtrl.CreatePendingChildAccount)
		})

		r.Get("/confirm-child", confirmChildCtrl.ConfirmChild)
	}
}
