package router

import (
	"api/internal/di"
	courseHandler "api/internal/domains/course/handler"
	courseRepo "api/internal/domains/course/persistence/repository"
	"api/internal/domains/customer"
	eventHandler "api/internal/domains/event/handler"
	eventRepo "api/internal/domains/event/persistence/repository"
	eventStaffHandler "api/internal/domains/event_staff/handler"
	eventStaffRepo "api/internal/domains/event_staff/persistence/repository"
	"api/internal/domains/identity/handler/authentication"
	"api/internal/domains/practice"
	practiceHandler "api/internal/domains/practice/handler"
	practiceRepo "api/internal/domains/practice/persistence/repository"
	purchase "api/internal/domains/purchase/handler"
	"api/internal/domains/staff"

	"api/internal/domains/facility/handler"
	"api/internal/domains/identity/handler/registration"
	"api/internal/domains/membership/handler"

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

		// Membership-related
		r.Route("/memberships", RegisterMembershipRoutes(container))
		r.Route("/membership-plan", RegisterMembershipPlansRoutes(container))

		// Authentication & Registration
		r.Route("/auth", RegisterAuthRoutes(container))
		r.Route("/register", RegisterRegistrationRoutes(container))

		// Core Functionalities
		r.Route("/courses", RegisterCourseRoutes(container))
		r.Route("/practices", RegisterPracticeRoutes(container))
		r.Route("/events", RegisterEventRoutes(container))
		r.Route("/facilities", RegisterFacilityRoutes(container))
		r.Route("/facility-categories", RegisterFacilityCategoriesRoutes(container))

		// Users & Staff
		r.Route("/customers", RegisterCustomerRoutes(container))
		r.Route("/staffs", RegisterStaffRoutes(container))
		r.Route("/event-staff", RegisterEventStaffRoutes(container))

		// Purchases
		r.Route("/purchases", RegisterPurchasesRoutes(container))
	})
}

func RegisterCustomerRoutes(container *di.Container) func(chi.Router) {

	h := customer.NewCustomersHandler(container)

	return func(r chi.Router) {

		r.Get("/", h.GetCustomers)
		r.Get("/{email}/children", h.GetChildrenByParentEmail)
		r.Get("/{email}", h.GetCustomerByEmail)
	}
}

func RegisterMembershipRoutes(container *di.Container) func(chi.Router) {
	h := membership.NewHandlers(container)

	return func(r chi.Router) {
		r.Get("/", h.GetMemberships)
		r.Get("/{id}", h.GetMembershipById)

		r.With(allowAdminOnly).Post("/", h.CreateMembership)
		r.With(allowAdminOnly).Put("/{id}", h.UpdateMembership)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteMembership)
	}
}

func RegisterMembershipPlansRoutes(container *di.Container) func(chi.Router) {
	h := membership.NewPlansHandlers(container)

	return func(r chi.Router) {
		r.Get("/", h.GetMembershipPlans)

		r.With(allowAdminOnly).Post("/", h.CreateMembershipPlan)
		r.With(allowAdminOnly).Put("/{id}", h.UpdateMembershipPlan)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteMembershipPlan)
	}
}

func RegisterFacilityRoutes(container *di.Container) func(chi.Router) {
	ctrl := facility.NewFacilitiesHandler(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetFacilities)
		r.Get("/{id}", ctrl.GetFacilityById)
		r.Post("/", ctrl.CreateFacility)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateFacility)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteFacility)
	}
}

func RegisterFacilityCategoriesRoutes(container *di.Container) func(chi.Router) {
	ctrl := facility.NewFacilityCategoriesHandler(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.List)
		r.Get("/{id}", ctrl.GetById)
		r.Post("/", ctrl.Create)
		r.With(allowAdminOnly).Put("/{id}", ctrl.Update)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.Delete)
	}
}

func RegisterCourseRoutes(container *di.Container) func(chi.Router) {

	courseRepository := courseRepo.NewCourseRepository(container.Queries.CoursesDb)
	ctrl := courseHandler.NewHandler(courseRepository)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetCourses)
		r.Get("/{id}", ctrl.GetCourseById)
		r.With(allowAdminOnly).Post("/", ctrl.CreateCourse)
		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateCourse)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteCourse)
	}
}

func RegisterPracticeRoutes(container *di.Container) func(chi.Router) {

	practiceRepository := practiceRepo.NewPracticeRepository(container.Queries.PracticesDb)
	service := practice.NewPracticeService(practiceRepository)

	h := practiceHandler.NewHandler(service)

	return func(r chi.Router) {
		r.Get("/", h.GetPractices)
		r.Post("/", h.CreatePractice)
		r.With(allowAdminOnly).Put("/{id}", h.UpdatePractice)
		r.With(allowAdminOnly).Delete("/{id}", h.DeletePractice)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	repo := eventRepo.NewEventsRepository(container.Queries.EventDb)
	handler := eventHandler.NewEventsHandler(repo)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)

		r.Post("/", handler.CreateEvent)
		r.With(allowAdminOnly).Put("/{id}", handler.UpdateEvent)
		r.With(allowAdminOnly).Delete("/{id}", handler.DeleteEvent)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	ctrl := staff.NewStaffHandlers(container)

	return func(r chi.Router) {
		r.Get("/", ctrl.GetStaffs)

		r.With(allowAdminOnly).Put("/{id}", ctrl.UpdateStaff)
		r.With(allowAdminOnly).Delete("/{id}", ctrl.DeleteStaff)
	}
}

func RegisterEventStaffRoutes(container *di.Container) func(chi.Router) {

	repo := eventStaffRepo.NewEventStaffsRepository(container.Queries.EventStaffDb)
	h := eventStaffHandler.NewEventStaffsHandler(repo)

	return func(r chi.Router) {
		r.Get("/{id}", h.GetStaffsAssignedToEvent)

		r.With(allowAdminOnly).Post("/", h.AssignStaffToEvent)
		r.With(allowAdminOnly).Delete("/", h.UnassignStaffFromEvent)
	}
}

func RegisterPurchasesRoutes(container *di.Container) func(chi.Router) {

	//repo := eventStaffRepo.NewEventStaffsRepository(container.Queries.EventStaffDb)
	h := purchase.NewPurchaseHandlers(container)

	return func(r chi.Router) {
		r.With(allowAdminOnly).Post("/memberships", h.PurchaseMembership)
	}
}

func RegisterAuthRoutes(container *di.Container) func(chi.Router) {

	authController := authentication.NewHandlers(container)

	return func(r chi.Router) {
		r.Post("/login", authController.Login)
	}
}

func RegisterRegistrationRoutes(container *di.Container) func(chi.Router) {

	customerRegistrationCtrl := registration.NewCustomerRegistrationHandlers(container)

	childRegistrationCtrl := registration.NewChildRegistrationHandlers(container)

	staffRegistrationCtrl := registration.NewStaffRegistrationHandlers(container)

	return func(r chi.Router) {

		r.Post("/customer", customerRegistrationCtrl.RegisterCustomer)
		r.Post("/staff", staffRegistrationCtrl.CreateStaff)
		//
		r.Post("/child", childRegistrationCtrl.RegisterChild)
	}
}
