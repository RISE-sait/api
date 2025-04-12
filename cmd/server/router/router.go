package router

import (
	"api/internal/di"

	enrollmentHandler "api/internal/domains/enrollment/handler"
	enrollmentRepo "api/internal/domains/enrollment/persistence/repository"

	"api/internal/domains/game"
	gameRepo "api/internal/domains/game/persistence"
	haircutRepo "api/internal/domains/haircut/persistence/repository"
	locationRepo "api/internal/domains/location/persistence"
	programHandler "api/internal/domains/program"
	programRepo "api/internal/domains/program/persistence"
	contextUtils "api/utils/context"

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
	payment "api/internal/domains/payment/handler"
	"api/internal/middlewares"

	"github.com/go-chi/chi"
)

type RouteConfig struct {
	Path      string
	Configure func(chi.Router)
}

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
		"/schedules": RegisterScheduleRoutes,
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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Patch("/{customer_id}/athlete", h.UpdateCustomerStats)
	}
}

func RegisterHaircutRoutes(container *di.Container) func(chi.Router) {

	return func(r chi.Router) {

		r.Get("/", haircut.GetHaircutImages)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleBarber)).Post("/", haircut.UploadHaircutImage)

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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleInstructor, contextUtils.RoleCoach, contextUtils.RoleAdmin)).Post("/", h.CreateEvent)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteEvent)
	}
}

func RegisterMembershipRoutes(container *di.Container) func(chi.Router) {
	membershipsHandler := membership.NewHandlers(container)
	membershipPlansHandler := membership.NewPlansHandlers(container)

	return func(r chi.Router) {
		r.Get("/", membershipsHandler.GetMemberships)
		r.Get("/{id}", membershipsHandler.GetMembershipById)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", membershipsHandler.CreateMembership)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", membershipsHandler.UpdateMembership)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", membershipsHandler.DeleteMembership)

		r.Get("/{id}/plans", membershipPlansHandler.GetMembershipPlans)
		r.Route("/plans", RegisterMembershipPlansRoutes(container))
	}
}

func RegisterMembershipPlansRoutes(container *di.Container) func(chi.Router) {
	h := membership.NewPlansHandlers(container)

	return func(r chi.Router) {

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateMembershipPlan)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateMembershipPlan)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteMembershipPlan)
	}
}

func RegisterGamesRoutes(container *di.Container) func(chi.Router) {

	repo := gameRepo.NewGameRepository(container.Queries.GameDb)
	h := game.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetGames)
		r.Get("/{id}", h.GetGameById)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateGame)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateGame)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteGame)
	}
}

func RegisterTeamsRoutes(container *di.Container) func(chi.Router) {

	repo := teamsRepo.NewTeamRepository(container.Queries.TeamDb)
	h := teamsHandler.NewHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetTeams)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteTeam)
	}
}

func RegisterLocationsRoutes(container *di.Container) func(chi.Router) {

	repo := locationRepo.NewLocationRepository(container.Queries.LocationDb)
	h := locationsHandler.NewLocationsHandler(repo)

	return func(r chi.Router) {
		r.Get("/", h.GetLocations)
		r.Get("/{id}", h.GetLocationById)
		r.Post("/", h.CreateLocation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateLocation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteLocation)
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
		r.Delete("/{id}", h.DeleteProgram)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewStaffHandlers(container)

	return func(r chi.Router) {
		r.Get("/", h.GetStaffs)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateStaff)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteStaff)
	}
}

func RegisterScheduleRoutes(container *di.Container) func(chi.Router) {
	repo := eventRepo.NewSchedulesRepository(container.Queries.EventDb)
	handler := eventHandler.NewSchedulesHandler(repo)

	return func(r chi.Router) {
		r.Get("/", handler.GetSchedules)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	repo := eventRepo.NewEventsRepository(container.Queries.EventDb)
	handler := eventHandler.NewEventsHandler(repo)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)
		r.Get("/{id}", handler.GetEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/", handler.CreateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Put("/{id}", handler.UpdateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/{id}", handler.DeleteEvent)

		r.Route("/{event_id}/staffs", RegisterEventStaffRoutes(container))
	}
}

func RegisterEventStaffRoutes(container *di.Container) func(chi.Router) {

	repo := enrollmentRepo.NewEventStaffsRepository(container.Queries.EnrollmentDb)
	h := enrollmentHandler.NewEventStaffsHandler(repo)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/{staff_id}", h.AssignStaffToEvent)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{staff_id}", h.UnassignStaffFromEvent)

	}
}

func RegisterCheckoutRoutes(container *di.Container) func(chi.Router) {

	h := payment.NewCheckoutHandlers(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/membership_plans/{id}", h.CheckoutMembership)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/programs/{id}", h.CheckoutProgram)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/events/{id}", h.CheckoutEvent)
	}
}

func RegisterWebhooksRoutes(container *di.Container) func(chi.Router) {

	h := payment.NewWebhookHandlers(container)

	return func(r chi.Router) {
		r.Post("/stripe", h.HandleStripeWebhook)
	}
}

func RegisterAuthRoutes(container *di.Container) func(chi.Router) {

	h := authentication.NewHandlers(container)

	return func(r chi.Router) {
		r.Post("/", h.Login)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/child/{id}", h.LoginAsChild)
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
		r.With(middlewares.JWTAuthMiddleware(false)).Post("/staff/approve/{id}", staffHandler.ApproveStaff)
		r.Post("/child", childRegistrationHandler.RegisterChild)
		r.Post("/parent", parentRegistrationHandler.RegisterParent)
	}
}
