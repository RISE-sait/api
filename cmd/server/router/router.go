package router

import (
	"api/internal/di"
	staff_activity_logs "api/internal/domains/audit/staff_activity_logs/handler"
	haircutEvents "api/internal/domains/haircut/event/handler"
	barberServicesHandler "api/internal/domains/haircut/haircut_service"
	haircut "api/internal/domains/haircut/portfolio"

	enrollmentHandler "api/internal/domains/enrollment/handler"

	courtHandler "api/internal/domains/court/handler"
	discountHandler "api/internal/domains/discount/handler"
	eventHandler "api/internal/domains/event/handler"
	"api/internal/domains/game"
	"api/internal/domains/identity/handler/authentication"
	"api/internal/domains/identity/handler/registration"
	locationsHandler "api/internal/domains/location/handler"
	membership "api/internal/domains/membership/handler"
	payment "api/internal/domains/payment/handler"
	playground "api/internal/domains/playground/handler"
	practice "api/internal/domains/practice/handler"
	programHandler "api/internal/domains/program"
	teamsHandler "api/internal/domains/team"
	userHandler "api/internal/domains/user/handler"
	"api/internal/middlewares"
	contextUtils "api/utils/context"

	aiHandler "api/internal/domains/ai/handler"
	contactHandler "api/internal/domains/contact/handler"
	"time"

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
		"/programs":   RegisterProgramRoutes,
		"/events":     RegisterEventRoutes,
		"/locations":  RegisterLocationsRoutes,
		"/games":      RegisterGamesRoutes,
		"/teams":      RegisterTeamsRoutes,
		"/playground": RegisterPlaygroundRoutes,
		"/discounts":  RegisterDiscountRoutes,
		"/courts":     RegisterCourtsRoutes,
		"/practices":  RegisterPracticesRoutes,

		// Users & Staff routes
		"/users":     RegisterUserRoutes,
		"/customers": RegisterCustomerRoutes,
		"/athletes":  RegisterAthleteRoutes,
		"/staffs":    RegisterStaffRoutes,

		// Haircut routes
		"/haircuts":         RegisterHaircutRoutes,
		"/barbers/services": RegisterBarberServicesRoutes,

		// Purchase-related routes
		"/checkout": RegisterCheckoutRoutes,

		// Webhooks
		"/webhooks": RegisterWebhooksRoutes,

		// Contact routes
		"/contact": RegisterContactRoutes,

		//Secure routes
		"/secure": RegisterSecureRoutes,

		// AI proxy route
		"/ai": RegisterAIRoutes,
	}

	for path, handler := range routeMappings {
		router.Route(path, handler(container))
	}
}

func RegisterUserRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewUsersHandlers(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Put("/{id}", h.UpdateUser)
	}
}

func RegisterCustomerRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewCustomersHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetCustomers)
		r.Get("/id/{id}", h.GetCustomerByID)
		r.Get("/email/{email}", h.GetCustomerByEmail)
		r.Get("/checkin/{id}", h.CheckinCustomer)
		r.Get("/{id}/memberships", h.GetMembershipHistory)
		r.Post("/{id}/archive", h.ArchiveCustomer)
		r.Post("/{id}/unarchive", h.UnarchiveCustomer)
		r.Get("/archived", h.ListArchivedCustomers)
	}
}

func RegisterAthleteRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewCustomersHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetAthletes)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Patch("/{id}/stats", h.UpdateAthleteStats)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{athlete_id}/team/{team_id}", h.UpdateAthletesTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{athlete_id}/team", h.RemoveAthleteFromTeam)
	}
}

func RegisterHaircutRoutes(container *di.Container) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", haircut.GetHaircutImages)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleBarber)).Post("/", haircut.UploadHaircutImage)

		r.Route("/events", RegisterHaircutEventsRoutes(container))
		r.Route("/services", RegisterBarberServicesRoutes(container))
	}
}

func RegisterBarberServicesRoutes(container *di.Container) func(chi.Router) {
	h := barberServicesHandler.NewBarberServicesHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetBarberServices)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleBarber)).Post("/", h.CreateBarberService)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleBarber)).Delete("/{id}", h.DeleteBarberService)
	}
}

func RegisterHaircutEventsRoutes(container *di.Container) func(chi.Router) {
	h := haircutEvents.NewEventsHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetEvents)
		r.Get("/{id}", h.GetEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/", h.CreateEvent)
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

func RegisterPracticesRoutes(container *di.Container) func(chi.Router) {
	h := practice.NewHandler(container)
	return func(r chi.Router) {
		r.Get("/", h.GetPractices)
		r.Get("/{id}", h.GetPractice)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Post("/", h.CreatePractice)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Put("/{id}", h.UpdatePractice)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Delete("/{id}", h.DeletePractice)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Post("/recurring", h.CreateRecurringPractices)
	}
}

func RegisterPlaygroundRoutes(container *di.Container) func(chi.Router) {
	h := playground.NewHandler(container)
	systemHandlers := playground.NewSystemsHandlers(container)
	return func(r chi.Router) {
		r.Get("/", h.GetSessions)
		r.Get("/{id}", h.GetSession)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/", h.CreateSession)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteSession)

		r.Route("/systems", func(r chi.Router) {
			r.Get("/", systemHandlers.GetSystems)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", systemHandlers.CreateSystem)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", systemHandlers.UpdateSystem)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", systemHandlers.DeleteSystem)
		})
	}
}
func RegisterTeamsRoutes(container *di.Container) func(chi.Router) {
	h := teamsHandler.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetTeams)
		r.Get("/{id}", h.GetTeamByID)
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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateLocation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateLocation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteLocation)
	}
}
func RegisterCourtsRoutes(container *di.Container) func(chi.Router) {
	h := courtHandler.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetCourts)
		r.Get("/{id}", h.GetCourt)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateCourt)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateCourt)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteCourt)
	}
}
func RegisterProgramRoutes(container *di.Container) func(chi.Router) {
	h := programHandler.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetPrograms)
		r.Get("/{id}", h.GetProgram)
		r.Get("/levels", h.GetProgramLevels)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateProgram)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateProgram)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteProgram)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	staffHandlers := userHandler.NewStaffHandlers(container)
	staffLogsHandlers := staff_activity_logs.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", staffHandlers.GetStaffs)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Get("/logs", staffLogsHandlers.GetStaffActivityLogs)

		r.With(middlewares.JWTAuthMiddleware(false)).Put("/{id}", staffHandlers.UpdateStaff)
		r.With(middlewares.JWTAuthMiddleware(false)).Delete("/{id}", staffHandlers.DeleteStaff)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	handler := eventHandler.NewEventsHandler(container)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)
		r.Get("/{id}", handler.GetEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/one-time", handler.CreateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Put("/{id}", handler.UpdateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/", handler.DeleteEvents)

		r.Route("/{event_id}/staffs", RegisterEventStaffRoutes(container))

		r.Route("/recurring", RegisterRecurringEventRoutes(container))
	}
}

func RegisterRecurringEventRoutes(container *di.Container) func(chi.Router) {
	handler := eventHandler.NewEventsHandler(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", handler.CreateRecurrences)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", handler.UpdateRecurrences)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", handler.DeleteRecurrence)
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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Get("/staff/pending", staffHandler.GetPendingStaffs)
		r.With(middlewares.JWTAuthMiddleware(false)).Post("/staff/approve/{id}", staffHandler.ApproveStaff)
		r.Post("/child", childRegistrationHandler.RegisterChild)
		r.Post("/parent", parentRegistrationHandler.RegisterParent)
	}
}
func RegisterDiscountRoutes(container *di.Container) func(chi.Router) {
	h := discountHandler.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetDiscounts)
		r.Get("/{id}", h.GetDiscount)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/apply", h.ApplyDiscount)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateDiscount)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateDiscount)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteDiscount)
	}
}
func RegisterContactRoutes(container *di.Container) func(chi.Router) {
	h := contactHandler.NewContactHandler(container)

	return func(r chi.Router) {
		// Contact form with rate limit
		r.With(middlewares.RateLimitMiddleware(0.1, 1, 10*time.Minute)).Post("/", h.SendContactEmail)

		// Newsletter subscription endpoint
		r.Post("/newsletter", h.SubscribeNewsletter)
	}
}

// RegisterSecureRoutes registers all secure routes that require authentication.
func RegisterSecureRoutes(container *di.Container) func(chi.Router) {
	return func(r chi.Router) {
		r.Route("/events", RegisterSecureEventRoutes(container))
		r.Route("/games", RegisterSecureGameRoutes(container))
		r.Route("/customers", RegisterSecureCustomerRoutes(container))
	}
}

// RegisterSecureEventRoutes registers secure event routes that require authentication.
func RegisterSecureEventRoutes(container *di.Container) func(chi.Router) {
	h := eventHandler.NewEventsHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/", h.GetRoleEvents)
	}
}

// RegisterSecureGameRoutes registers secure game routes that require authentication.
func RegisterSecureGameRoutes(container *di.Container) func(chi.Router) {
	h := game.NewHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/", h.GetRoleGames)
	}
}

// RegisterSecureCustomerRoutes registers secure customer routes that require authentication.
func RegisterSecureCustomerRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewCustomersHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/memberships", h.GetUserMembershipHistory)
	}
}

// RegisterAIRoutes registers the AI proxy route with rate limiting.
func RegisterAIRoutes(_ *di.Container) func(chi.Router) {
	h := aiHandler.NewHandler()
	return func(r chi.Router) {
		r.With(middlewares.RateLimitMiddleware(1.0, 5, time.Minute)).Post("/chat", h.ProxyMessage)
	}
}
