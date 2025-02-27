package router

import (
	"api/internal/di"
	barber "api/internal/domains/barber/handler/events"
	haircut "api/internal/domains/barber/handler/haircuts"
	barberEventRepo "api/internal/domains/barber/persistence/repository/event"
	courseHandler "api/internal/domains/course/handler"
	courseRepo "api/internal/domains/course/persistence/repository"
	"api/internal/domains/customer"
	eventHandler "api/internal/domains/event/handler"
	eventRepo "api/internal/domains/event/persistence/repository"
	eventStaffHandler "api/internal/domains/event_staff/handler"
	eventStaffRepo "api/internal/domains/event_staff/persistence/repository"
	game "api/internal/domains/game/handler"
	gameRepo "api/internal/domains/game/persistence/repository"

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

var allowAdminOnly = middlewares.JWTAuthMiddleware(false, RoleAdmin)
var allowAnyoneWithValidToken = middlewares.JWTAuthMiddleware(true)

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
		r.Route("/games", RegisterGamesRoutes(container))

		// Users & Staff
		r.Route("/customers", RegisterCustomerRoutes(container))
		r.Route("/staffs", RegisterStaffRoutes(container))
		r.Route("/event-staff", RegisterEventStaffRoutes(container))

		// Barber
		r.Route("/barbers", RegisterBarberRoutes(container))

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

func RegisterBarberRoutes(container *di.Container) func(chi.Router) {

	//h := customer.NewCustomersHandler(container)

	return func(r chi.Router) {

		r.Post("/haircuts", haircut.UploadHaircutImage)

		r.Route("/events", RegisterBarberEventsRoutes(container))
	}
}

func RegisterBarberEventsRoutes(container *di.Container) func(chi.Router) {

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
		r.With(allowAdminOnly).Put("/{id}", h.UpdateMembershipPlan)
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
	h := staff.NewStaffHandlers(container)

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

func RegisterPurchasesRoutes(container *di.Container) func(chi.Router) {

	//repo := eventStaffRepo.NewEventStaffsRepository(container.Queries.EventStaffDb)
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
