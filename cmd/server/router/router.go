package router

import (
	"api/internal/di"
	courseHandler "api/internal/domains/course/handler"
	courseRepo "api/internal/domains/course/persistence/repository"
	enrollment "api/internal/domains/enrollment/handler"
	enrollmentRepo "api/internal/domains/enrollment/persistence"
	enrollmentService "api/internal/domains/enrollment/service"
	"api/internal/domains/game"
	gameRepo "api/internal/domains/game/persistence"
	userHandler "api/internal/domains/user/handler"

	eventHandler "api/internal/domains/event/handler"
	eventRepo "api/internal/domains/event/persistence/repository"
	eventStaffHandler "api/internal/domains/event_staff/handler"
	eventStaffRepo "api/internal/domains/event_staff/persistence/repository"
	"api/internal/domains/facility/handler"
	barber "api/internal/domains/haircut/handler/events"
	haircut "api/internal/domains/haircut/handler/haircuts"
	barberEventRepo "api/internal/domains/haircut/persistence/repository/event"
	"api/internal/domains/identity/handler/authentication"
	"api/internal/domains/identity/handler/registration"
	"api/internal/domains/membership/handler"
	practiceHandler "api/internal/domains/practice/handler"
	practiceRepo "api/internal/domains/practice/persistence/repository"
	purchase "api/internal/domains/purchase/handler"
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

var allowAdminOnly = middlewares.JWTAuthMiddleware(false, RoleAdmin)
var allowAnyoneWithValidToken = middlewares.JWTAuthMiddleware(true)

func RegisterRoutes(router *chi.Mux, container *di.Container) {
	routeMappings := map[string]func(*di.Container) func(chi.Router){

		// Membership-related routes
		"/memberships":     RegisterMembershipRoutes,
		"/membership-plan": RegisterMembershipPlansRoutes,

		// Authentication & Registration routes
		"/auth":     RegisterAuthRoutes,
		"/register": RegisterRegistrationRoutes,

		// Core functionalities
		"/courses":     RegisterCourseRoutes,
		"/practices":   RegisterPracticeRoutes,
		"/events":      RegisterEventRoutes,
		"/facilities":  RegisterFacilityRoutes,
		"/games":       RegisterGamesRoutes,
		"/enrollments": RegisterEnrollmentRoutes,

		// Users & Staff routes
		"/users":       RegisterUserRoutes,
		"/customers":   RegisterCustomerRoutes,
		"/staffs":      RegisterStaffRoutes,
		"/event-staff": RegisterEventStaffRoutes,

		// Haircut routes
		"/haircuts": RegisterHaircutRoutes,

		// Purchase-related routes
		"/purchases": RegisterPurchasesRoutes,
	}

	for path, handler := range routeMappings {
		router.Route(path, handler(container))
	}
}

func RegisterUserRoutes(container *di.Container) func(chi.Router) {

	h := userHandler.NewUsersHandler(container)

	return func(r chi.Router) {

		r.Get("/{email}/children", h.GetChildrenByParentEmail)
		r.Get("/{email}", h.GetUserByEmail)
	}
}

func RegisterCustomerRoutes(container *di.Container) func(chi.Router) {

	h := userHandler.NewCustomersHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetCustomers)
		r.Patch("/customers/{customer_id}/stats", h.UpdateCustomerStats)
	}
}

func RegisterHaircutRoutes(container *di.Container) func(chi.Router) {

	return func(r chi.Router) {

		r.Get("/", haircut.GetHaircutImages)
		r.Post("/", haircut.UploadHaircutImage)

		r.Route("/events", RegisterHaircutEventsRoutes(container))
	}
}

func RegisterHaircutEventsRoutes(container *di.Container) func(chi.Router) {

	repo := barberEventRepo.NewEventsRepository(container.Queries.BarberDb)
	h := barber.NewEventsHandler(repo)

	return func(r chi.Router) {

		r.Get("/", h.GetEvents)
		r.Get("/{id}", h.GetEventDetails)
		r.Post("/", h.CreateEvent)
		r.Put("/{id}", h.UpdateEvent)
		r.Delete("/{id}", h.DeleteEvent)
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
		r.Put("/{id}", h.UpdateMembershipPlan)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteMembershipPlan)
	}
}

func RegisterGamesRoutes(container *di.Container) func(chi.Router) {

	repo := gameRepo.NewGameRepository(container.Queries.GameDb)
	h := game.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetGames)
		r.Get("/{id}", h.GetGameById)

		r.With(allowAdminOnly).Post("/", h.CreateGame)
		r.With(allowAdminOnly).Put("/{id}", h.UpdateGame)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteGame)
	}
}

func RegisterFacilityRoutes(container *di.Container) func(chi.Router) {
	h := facility.NewFacilitiesHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetFacilities)
		r.Get("/{id}", h.GetFacilityById)
		r.Post("/", h.CreateFacility)
		r.With(allowAdminOnly).Put("/{id}", h.UpdateFacility)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteFacility)

		r.Route("/categories", RegisterFacilityCategoriesRoutes(container))
	}
}

func RegisterFacilityCategoriesRoutes(container *di.Container) func(chi.Router) {
	h := facility.NewFacilityCategoriesHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("/{id}", h.GetById)
		r.Post("/", h.Create)
		r.With(allowAdminOnly).Put("/{id}", h.Update)
		r.With(allowAdminOnly).Delete("/{id}", h.Delete)
	}
}

func RegisterCourseRoutes(container *di.Container) func(chi.Router) {

	courseRepository := courseRepo.NewCourseRepository(container.Queries.CoursesDb)
	h := courseHandler.NewHandler(courseRepository)

	return func(r chi.Router) {
		r.Get("/", h.GetCourses)
		r.Get("/{id}", h.GetCourseById)
		r.With(allowAdminOnly).Post("/", h.CreateCourse)
		r.With(allowAdminOnly).Put("/{id}", h.UpdateCourse)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteCourse)
	}
}

func RegisterPracticeRoutes(container *di.Container) func(chi.Router) {

	repo := practiceRepo.NewPracticeRepository(container.Queries.PracticesDb)
	h := practiceHandler.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetPractices)
		r.Post("/", h.CreatePractice)
		r.Put("/{id}", h.UpdatePractice)
		r.With(allowAdminOnly).Delete("/{id}", h.DeletePractice)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	repo := eventRepo.NewEventsRepository(container.Queries.EventDb)
	handler := eventHandler.NewEventsHandler(repo)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)
		r.Get("/{id}", handler.GetEvent)
		r.Post("/", handler.CreateEvent)
		r.With(allowAdminOnly).Put("/{id}", handler.UpdateEvent)
		r.With(allowAdminOnly).Delete("/{id}", handler.DeleteEvent)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewStaffHandlers(container)

	return func(r chi.Router) {
		r.Get("/", h.GetStaffs)

		r.With(allowAdminOnly).Put("/{id}", h.UpdateStaff)
		r.With(allowAdminOnly).Delete("/{id}", h.DeleteStaff)
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

func RegisterEnrollmentRoutes(container *di.Container) func(chi.Router) {

	repo := enrollmentRepo.NewEnrollmentRepository(container.Queries.EnrollmentDb)

	service := enrollmentService.NewEnrollmentService(repo)
	h := enrollment.NewHandler(service)

	return func(r chi.Router) {
		r.Get("/", h.GetEnrollments)
		r.Post("/", h.CreateEnrollment)
		r.Delete("/{id}", h.DeleteEnrollment)
	}
}

func RegisterPurchasesRoutes(container *di.Container) func(chi.Router) {

	h := purchase.NewPurchaseHandlers(container)

	return func(r chi.Router) {
		r.With(allowAdminOnly).Post("/memberships", h.PurchaseMembership)
	}
}

func RegisterAuthRoutes(container *di.Container) func(chi.Router) {

	h := authentication.NewHandlers(container)

	return func(r chi.Router) {
		r.Post("/", h.Login)
		r.With(allowAnyoneWithValidToken).Post("/child/{hubspot_id}", h.LoginAsChild)
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
