package router

import (
	"api/internal/di"
	enrollmentService "api/internal/domains/event/service"
	"api/internal/domains/game"
	gameRepo "api/internal/domains/game/persistence"
	haircutRepo "api/internal/domains/haircut/persistence/repository"
	locationRepo "api/internal/domains/location/persistence"
	programHandler "api/internal/domains/program"
	programRepo "api/internal/domains/program/persistence"

	teamsHandler "api/internal/domains/team"
	teamsRepo "api/internal/domains/team/persistence"

	userHandler "api/internal/domains/user/handler"

	eventHandler "api/internal/domains/event/handler"
	eventRepo "api/internal/domains/event/persistence/repository"
	barberServicesHandler "api/internal/domains/haircut/handler/barber_services"
	haircutEvents "api/internal/domains/haircut/handler/events"
	haircut "api/internal/domains/haircut/handler/haircuts"
	"api/internal/domains/identity/handler/authentication"
	"api/internal/domains/identity/handler/registration"
	locationsHandler "api/internal/domains/location/handler"
	membership "api/internal/domains/membership/handler"
	purchase "api/internal/domains/payment/handler"
	"api/internal/middlewares"

	"github.com/go-chi/chi"
)

type RouteConfig struct {
	Path      string
	Configure func(chi.Router)
}

const (
	RoleAdmin      = "ADMIN"
	RoleInstructor = "INSTRUCTOR"
	RoleParent     = "PARENT"
	RoleChild      = "CHILD"
	RoleAthlete    = "ATHLETE"
	RoleCoach      = "COACH"
	RoleBarber     = "BARBER"
)

var allowAdmin = middlewares.JWTAuthMiddleware(false, RoleAdmin)
var allowInstructor = middlewares.JWTAuthMiddleware(false, RoleInstructor)
var allowParent = middlewares.JWTAuthMiddleware(false, RoleParent)
var allowChild = middlewares.JWTAuthMiddleware(false, RoleChild)
var allowAthlete = middlewares.JWTAuthMiddleware(false, RoleAthlete)
var allowCoach = middlewares.JWTAuthMiddleware(false, RoleCoach)

var allowAnyoneWithValidToken = middlewares.JWTAuthMiddleware(true)

func RegisterRoutes(router *chi.Mux, container *di.Container) {
	routeMappings := map[string]func(*di.Container) func(chi.Router){

		// Membership-related routes
		"/memberships": RegisterMembershipRoutes,

		// Authentication & Registration routes
		"/auth":     RegisterAuthRoutes,
		"/register": RegisterRegistrationRoutes,

		// Core functionalities
		"/programs":  RegisterProgramRoutes,
		"/events":    RegisterEventRoutes,
		"/locations": RegisterLocationsRoutes,
		"/games":     RegisterGamesRoutes,
		"/teams":     RegisterTeamsRoutes,

		// Users & Staff routes
		"/customers": RegisterCustomerRoutes,
		"/staffs":    RegisterStaffRoutes,

		// Haircut routes
		"/haircuts":         RegisterHaircutRoutes,
		"/barbers/services": RegisterBarberServicesRoutes,

		// Purchase-related routes
		"/checkout": RegisterCheckoutRoutes,

		// Webhooks
		"/webhooks": RegisterWebhooksRoutes,
	}

	for path, handler := range routeMappings {
		router.Route(path, handler(container))
	}
}

func RegisterCustomerRoutes(container *di.Container) func(chi.Router) {

	h := userHandler.NewCustomersHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetCustomers)
		r.Get("/id/{id}", h.GetCustomerByID)
		r.Get("/email/{email}", h.GetCustomerByEmail)
		r.With(allowAdmin).Patch("/{customer_id}/athlete", h.UpdateCustomerStats)
	}
}

func RegisterHaircutRoutes(container *di.Container) func(chi.Router) {

	return func(r chi.Router) {

		r.Get("/", haircut.GetHaircutImages)
		r.With(middlewares.JWTAuthMiddleware(false, RoleBarber)).Post("/", haircut.UploadHaircutImage)

		r.Route("/events", RegisterHaircutEventsRoutes(container))
	}
}

func RegisterBarberServicesRoutes(container *di.Container) func(chi.Router) {

	repo := haircutRepo.NewBarberServiceRepository(container.Queries.BarberDb)
	h := barberServicesHandler.NewBarberServicesHandler(repo)

	return func(r chi.Router) {

		r.Get("/", h.GetBarberServices)
		r.Post("/", h.CreateBarberService)
		r.Delete("/{id}", h.DeleteBarberService)
	}
}

func RegisterHaircutEventsRoutes(container *di.Container) func(chi.Router) {

	repo := haircutRepo.NewEventsRepository(container.Queries.BarberDb)
	h := haircutEvents.NewEventsHandler(repo)

	return func(r chi.Router) {

		r.Get("/", h.GetEvents)
		r.Get("/{id}", h.GetEvent)
		r.With(allowAnyoneWithValidToken).Post("/", h.CreateEvent)
		r.Delete("/{id}", h.DeleteEvent)
	}
}

func RegisterMembershipRoutes(container *di.Container) func(chi.Router) {
	membershipsHandler := membership.NewHandlers(container)
	membershipPlansHandler := membership.NewPlansHandlers(container)

	return func(r chi.Router) {
		r.Get("/", membershipsHandler.GetMemberships)
		r.Get("/{id}", membershipsHandler.GetMembershipById)

		r.With(allowAdmin).Post("/", membershipsHandler.CreateMembership)
		r.With(allowAdmin).Put("/{id}", membershipsHandler.UpdateMembership)
		r.With(allowAdmin).Delete("/{id}", membershipsHandler.DeleteMembership)

		r.Get("/{id}/plans", membershipPlansHandler.GetMembershipPlans)
		r.Route("/plans", RegisterMembershipPlansRoutes(container))
	}
}

func RegisterMembershipPlansRoutes(container *di.Container) func(chi.Router) {
	h := membership.NewPlansHandlers(container)

	return func(r chi.Router) {

		r.Get("/payment-frequencies", h.GetMembershipPlanPaymentFrequencies)
		r.With(allowAdmin).Post("/", h.CreateMembershipPlan)
		r.Put("/{id}", h.UpdateMembershipPlan)
		r.With(allowAdmin).Delete("/{id}", h.DeleteMembershipPlan)
	}
}

func RegisterGamesRoutes(container *di.Container) func(chi.Router) {

	repo := gameRepo.NewGameRepository(container.Queries.GameDb)
	h := game.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetGames)
		r.Get("/{id}", h.GetGameById)

		r.With(allowAdmin).Post("/", h.CreateGame)
		r.With(allowAdmin).Put("/{id}", h.UpdateGame)
		r.With(allowAdmin).Delete("/{id}", h.DeleteGame)
	}
}

func RegisterTeamsRoutes(container *di.Container) func(chi.Router) {

	repo := teamsRepo.NewTeamRepository(container.Queries.TeamDb)
	h := teamsHandler.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetTeams)

		r.With(allowAdmin).Post("/", h.CreateTeam)
		r.With(allowAdmin).Put("/{id}", h.UpdateTeam)
		r.With(allowAdmin).Delete("/{id}", h.DeleteTeam)
	}
}

func RegisterLocationsRoutes(container *di.Container) func(chi.Router) {

	repo := locationRepo.NewLocationRepository(container.Queries.LocationDb)
	h := locationsHandler.NewLocationsHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetLocations)
		r.Get("/{id}", h.GetLocationById)
		r.Post("/", h.CreateLocation)
		r.With(allowAdmin).Put("/{id}", h.UpdateLocation)
		r.With(allowAdmin).Delete("/{id}", h.DeleteLocation)
	}
}

func RegisterProgramRoutes(container *di.Container) func(chi.Router) {

	repo := programRepo.NewProgramRepository(container.Queries.ProgramDb)
	h := programHandler.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetPrograms)
		r.Get("/{id}", h.GetProgram)
		r.Get("/levels", h.GetProgramLevels)
		r.Post("/", h.CreateProgram)
		r.Put("/{id}", h.UpdateProgram)
		r.With(allowAdmin).Delete("/{id}", h.DeleteProgram)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewStaffHandlers(container)

	return func(r chi.Router) {
		r.Get("/", h.GetStaffs)

		r.With(allowAdmin).Put("/{id}", h.UpdateStaff)
		r.With(allowAdmin).Delete("/{id}", h.DeleteStaff)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	repo := eventRepo.NewEventsRepository(container.Queries.EventDb)
	handler := eventHandler.NewEventsHandler(repo)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)
		r.Get("/{id}", handler.GetEvent)
		r.With(allowAdmin).Post("/", handler.CreateEvent)
		r.With(allowAdmin).Put("/{id}", handler.UpdateEvent)
		r.With(allowAdmin).Delete("/{id}", handler.DeleteEvent)

		r.Route("/{event_id}/staffs", RegisterEventStaffRoutes(container))
		r.Route("/{event_id}/customers", RegisterEventCustomerRoutes(container))
	}
}

func RegisterEventCustomerRoutes(container *di.Container) func(chi.Router) {

	repo := eventRepo.NewEnrollmentRepository(container.Queries.EventDb)
	service := enrollmentService.NewEnrollmentService(repo)
	h := eventHandler.NewCustomerEnrollmentHandler(service)

	return func(r chi.Router) {
		r.With(allowAdmin).Post("/{customer_id}", h.EnrollCustomer)
		r.With(allowAdmin).Delete("/{customer_id}", h.UnenrollCustomer)
	}
}

func RegisterEventStaffRoutes(container *di.Container) func(chi.Router) {

	repo := eventRepo.NewEventStaffsRepository(container.Queries.EventDb)
	h := eventHandler.NewEventStaffsHandler(repo)

	return func(r chi.Router) {
		r.With(allowAdmin).Post("/{staff_id}", h.AssignStaffToEvent)
		r.With(allowAdmin).Delete("/{staff_id}", h.UnassignStaffFromEvent)
	}
}

func RegisterCheckoutRoutes(container *di.Container) func(chi.Router) {

	h := purchase.NewPaymentHandlers(container)

	return func(r chi.Router) {
		r.With(allowAnyoneWithValidToken).Post("/membership_plans/{id}", h.CheckoutMembership)
		r.Post("/programs/{id}", h.CheckoutProgram)
	}
}

func RegisterWebhooksRoutes(container *di.Container) func(chi.Router) {

	h := purchase.NewPaymentHandlers(container)

	return func(r chi.Router) {
		r.Post("/stripe", h.HandleStripeWebhook)
	}
}

func RegisterAuthRoutes(container *di.Container) func(chi.Router) {

	h := authentication.NewHandlers(container)

	return func(r chi.Router) {
		r.Post("/", h.Login)
		r.With(allowAnyoneWithValidToken).Post("/child/{id}", h.LoginAsChild)
	}
}

func RegisterRegistrationRoutes(container *di.Container) func(chi.Router) {

	athleteHandler := registration.NewAthleteRegistrationHandlers(container)
	staffHandler := registration.NewStaffRegistrationHandlers(container)

	childRegistrationHandler := registration.NewChildRegistrationHandlers(container)
	parentRegistrationHandler := registration.NewParentRegistrationHandlers(container)

	return func(r chi.Router) {

		r.Post("/athlete", athleteHandler.RegisterAthlete)

		r.Post("/staff", staffHandler.RegisterStaff)
		r.Post("/child", childRegistrationHandler.RegisterChild)
		r.Post("/parent", parentRegistrationHandler.RegisterParent)
	}
}
