package routes

import (
	"api/internal/di"
	"api/internal/domains/course"
	facility "api/internal/domains/facility/controllers"
	identity "api/internal/domains/identity/controllers"
	membership "api/internal/domains/membership/controllers"
	"api/internal/domains/schedule"
	"api/internal/middlewares"

	"github.com/go-chi/chi"
)

type RouteConfig struct {
	Path      string
	Configure func(chi.Router)
}

const (
	RoleAdmin = "ADMIN"
)

var allowAdminOnly = middlewares.JWTAuthMiddleware(RoleAdmin)

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

		r.With(allowAdminOnly).Post("/", ctrl.CreateMembership)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateMembership)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteMembership)

		// plans as subroute
		r.Route("/{membershipId}/plans", RegisterMembershipPlansRoutes(r, container))
	}
}

func RegisterMembershipPlansRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := membership.NewMembershipPlansController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetMembershipPlansByMembershipId)

		r.With(allowAdminOnly).Post("/", ctrl.CreateMembershipPlan)
		r.With(allowAdminOnly).Put("/{planId}", ctrl.UpdateMembershipPlan)
		r.With(allowAdminOnly).Delete("/{planId}", ctrl.DeleteMembershipPlan)
	}
}

func RegisterFacilityRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := facility.NewFacilitiesController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetFacilities)
		r.Get("/{id}", ctrl.GetFacilityById)
		r.With(allowAdminOnly).Post("/", ctrl.CreateFacility)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateFacility)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteFacility)

		r.Route("/types", RegisterFacilityTypesRoutes(r, container))
	}
}

func RegisterFacilityTypesRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := facility.NewFacilityTypesController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllFacilityTypes)
		r.Get("/{id}", ctrl.GetFacilityTypeById)
		r.With(allowAdminOnly).Post("/", ctrl.CreateFacilityType)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateFacilityType)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteFacilityType)
	}
}

func RegisterCourseRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := course.NewCourseController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetCourses)
		r.Get("/{id}", ctrl.GetCourseById)
		r.With(allowAdminOnly).Post("/", ctrl.CreateCourse)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateCourse)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteCourse)
	}
}

func RegisterScheduleRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := schedule.NewSchedulesController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetSchedules)
		r.With(allowAdminOnly).Post("/", ctrl.CreateSchedule)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateSchedule)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteSchedule)
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
