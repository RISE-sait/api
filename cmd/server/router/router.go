package router

import (
	"api/internal/di"
	"api/internal/domains/course"
	customer "api/internal/domains/customer"
	"api/internal/domains/staff"

	courseRepo "api/internal/domains/course/persistence"
	"api/internal/domains/events"
	facility "api/internal/domains/facility/controllers"
	auth "api/internal/domains/identity/controllers/auth"
	registration "api/internal/domains/identity/controllers/registration"
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
			{Path: "/memberships", Configure: RegisterMembershipRoutes(container)},
			{Path: "/auth", Configure: RegisterAuthRoutes(container)},
			{Path: "/register", Configure: RegisterRegistrationRoutes(container)},
			{Path: "/courses", Configure: RegisterCourseRoutes(container)},
			{Path: "/events", Configure: RegisterEventRoutes(container)},
			{Path: "/facilities", Configure: RegisterFacilityRoutes(container)},
			{Path: "/customers", Configure: RegisterCustomerRoutes(container)},
			{Path: "/staffs", Configure: RegisterStaffRoutes(container)},
		}

		for _, route := range routes {
			r.Route(route.Path, route.Configure)
		}
	})
}

func RegisterCustomerRoutes(container *di.Container) func(chi.Router) {

	ctrl := customer.NewCustomersController(container)

	return func(r chi.Router) {

		r.Get("/", ctrl.GetCustomers)
		r.Get("/{email}", ctrl.GetCustomerByEmail)
		r.Post("/", ctrl.CreateCustomer)
	}
}

func RegisterMembershipRoutes(container *di.Container) func(chi.Router) {
	ctrl := membership.NewMembershipController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllMemberships)
		r.Get("/{id}", ctrl.GetMembershipById)

		r.With(allowAdminOnly).Post("/", ctrl.CreateMembership)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateMembership)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteMembership)

		// plans as subroute
		r.Route("/{membershipId}/plans", RegisterMembershipPlansRoutes(container))
	}
}

func RegisterMembershipPlansRoutes(container *di.Container) func(chi.Router) {
	ctrl := membership.NewMembershipPlansController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetMembershipPlansByMembershipId)

		r.With(allowAdminOnly).Post("/", ctrl.CreateMembershipPlan)
		r.With(allowAdminOnly).Put("/{planId}", ctrl.UpdateMembershipPlan)
		r.With(allowAdminOnly).Delete("/{planId}", ctrl.DeleteMembershipPlan)
	}
}

func RegisterFacilityRoutes(container *di.Container) func(chi.Router) {
	ctrl := facility.NewFacilitiesController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetFacilities)
		r.Get("/{id}", ctrl.GetFacilityById)
		r.Post("/", ctrl.CreateFacility)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateFacility)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteFacility)

		r.Route("/types", RegisterFacilityTypesRoutes(container))
	}
}

func RegisterFacilityTypesRoutes(container *di.Container) func(chi.Router) {
	ctrl := facility.NewFacilityTypesController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllFacilityTypes)
		r.Get("/{id}", ctrl.GetFacilityTypeById)
		r.Post("/", ctrl.CreateFacilityType)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateFacilityType)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteFacilityType)
	}
}

func RegisterCourseRoutes(container *di.Container) func(chi.Router) {

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

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	ctrl := events.NewEventsController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetEvents)

		r.With(allowAdminOnly).Post("/", ctrl.CreateEvent)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateEvent)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteEvent)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	ctrl := staff.NewStaffController(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetStaffs)

		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateStaff)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteStaff)
	}
}

func RegisterAuthRoutes(container *di.Container) func(chi.Router) {

	authController := auth.NewAuthenticationController(container)
	OauthController := auth.NewOauthController(container)

	return func(r chi.Router) {
		r.Post("/login", authController.Login)

		r.Post("/oauth/google", OauthController.HandleOAuthCallback)
	}
}

func RegisterRegistrationRoutes(container *di.Container) func(chi.Router) {

	customerRegistrationCtrl := registration.NewCustomerRegistrationController(container)

	childRegistrationCtrl := registration.NewCreatePendingChildAccountController(container)

	confirmChildCtrl := registration.NewChildAccountConfirmationController(container)

	staffRegistrationCtrl := registration.NewStaffRegistrationController(container)

	return func(r chi.Router) {

		r.Post("/customer", customerRegistrationCtrl.CreateCustomer)
		r.Post("/child/pending", childRegistrationCtrl.CreatePendingChildAccount)
		r.Post("/staff", staffRegistrationCtrl.CreateStaff)

		r.Get("/confirm-child", confirmChildCtrl.ConfirmChild)
	}
}
