package router

import (
	"api/internal/di"

	enrollmentHandler "api/internal/domains/enrollment/handler"

	"api/internal/domains/game"
	programHandler "api/internal/domains/program"
	contextUtils "api/utils/context"

	teamsHandler "api/internal/domains/team"

	userHandler "api/internal/domains/user/handler"

	eventHandler "api/internal/domains/event/handler"

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

	h := barberServicesHandler.NewBarberServicesHandler(container)

	return func(r chi.Router) {

		r.Get("/", h.GetBarberServices)
		r.Post("/", h.CreateBarberService)
		r.Delete("/{id}", h.DeleteBarberService)
	}
}

func RegisterHaircutEventsRoutes(container *di.Container) func(chi.Router) {

	h := haircutEvents.NewEventsHandler(container)

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

	h := game.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetGames)
		r.Get("/{id}", h.GetGameById)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateGame)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateGame)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteGame)
	}
}

func RegisterTeamsRoutes(container *di.Container) func(chi.Router) {

	h := teamsHandler.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetTeams)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteTeam)
	}
}

func RegisterLocationsRoutes(container *di.Container) func(chi.Router) {

	h := locationsHandler.NewLocationsHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetLocations)
		r.Get("/{id}", h.GetLocationById)
		r.Post("/", h.CreateLocation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateLocation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteLocation)
	}
}

func RegisterProgramRoutes(container *di.Container) func(chi.Router) {

	h := programHandler.NewHandler(container)

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
	handler := eventHandler.NewSchedulesHandler(container)

	return func(r chi.Router) {
		r.Get("/", handler.GetSchedules)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	handler := eventHandler.NewEventsHandler(container)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)
		r.Get("/{id}", handler.GetEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/", handler.CreateEvents)
		r.With(middlewares.JWTAuthMiddleware(true)).Put("/{id}", handler.UpdateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/", handler.DeleteEvents)

		r.Route("/{event_id}/staffs", RegisterEventStaffRoutes(container))
	}
}

func RegisterEventStaffRoutes(container *di.Container) func(chi.Router) {

	h := enrollmentHandler.NewEventStaffsHandler(container)

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
