package router

import (
	"api/internal/di"
	"api/internal/domains/course"
	customer "api/internal/domains/customer"
	"api/internal/domains/staff"

	courseRepo "api/internal/domains/course/persistence"
	"api/internal/domains/events"
	facility "api/internal/domains/facility/controllers"
	identity "api/internal/domains/identity/controllers"
	membership "api/internal/domains/membership/controllers"

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
			{Path: "/auth", Configure: RegisterAuthRoutes(r, container)},
			{Path: "/courses", Configure: RegisterCourseRoutes(r, container)},
			{Path: "/events", Configure: RegisterEventRoutes(r, container)},
			{Path: "/facilities", Configure: RegisterFacilityRoutes(r, container)},
			{Path: "/customers", Configure: RegisterCustomerRoutes(r, container)},
			{Path: "/staffs", Configure: RegisterStaffRoutes(r, container)},
		}

		for _, route := range routes {
			r.Route(route.Path, route.Configure)
		}
	})
}

func RegisterCustomerRoutes(r chi.Router, container *di.Container) func(chi.Router) {

	ctrl := customer.NewCustomersController(container)

	return func(r chi.Router) {

		r.Get("/", ctrl.GetCustomers)
		r.Get("/{email}", ctrl.GetCustomerByEmail)
		r.Post("/", ctrl.CreateCustomer)
	}
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
		r.Post("/", ctrl.CreateFacility)
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
		r.Post("/", ctrl.CreateFacilityType)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateFacilityType)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteFacilityType)
	}
}

func RegisterCourseRoutes(r chi.Router, container *di.Container) func(chi.Router) {

	courseRepo := courseRepo.NewCourseRepository(container)
	courseService := course.NewCourseService(courseRepo)
	ctrl := course.NewCourseController(courseService)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetCourses)
		r.Get("/{id}", ctrl.GetCourseById)
		r.Post("/", ctrl.CreateCourse)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateCourse)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteCourse)
	}
}

func RegisterEventRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := events.NewEventsController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetEvents)

		r.With(allowAdminOnly).Post("/", ctrl.CreateEvent)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateEvent)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteEvent)
	}
}

func RegisterStaffRoutes(r chi.Router, container *di.Container) func(chi.Router) {
	ctrl := staff.NewStaffController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetStaffs)

		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateStaff)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteStaff)
	}
}

func RegisterAuthRoutes(r chi.Router, container *di.Container) func(chi.Router) {

	authController := identity.NewAuthenticationController(container)
	tokenValidationCtrl := identity.NewTokenValidationController(container)
	OauthController := identity.NewOauthController(container)

	return func(r chi.Router) {
		r.Post("/login", authController.Login)

		r.Post("/oauth/google", OauthController.HandleOAuthCallback)
		r.Get("/validate-jwt", tokenValidationCtrl.ValidateToken)
	}
}

func RegisterIdentityRoutes(r chi.Router, container *di.Container) func(chi.Router) {

	customerRegistrationCtrl := identity.NewCustomerRegistrationController(container)

	childRegistrationCtrl := identity.NewCreatePendingChildAccountController(container)

	confirmChildCtrl := identity.NewChildAccountConfirmationController(container)

	return func(r chi.Router) {

		r.Route("/register", func(registration chi.Router) {
			registration.Post("/", customerRegistrationCtrl.CreateCustomer)
			registration.Post("/child", childRegistrationCtrl.CreatePendingChildAccount)
		})

		r.Get("/confirm-child", confirmChildCtrl.ConfirmChild)
	}
}
