package router

import (
	"api/internal/di"
	staff_activity_logs "api/internal/domains/audit/staff_activity_logs/handler"
	haircutEvents "api/internal/domains/haircut/event/handler"
	barberServicesHandler "api/internal/domains/haircut/haircut_service"
	haircut "api/internal/domains/haircut/portfolio"

	bookingsHandler "api/internal/domains/booking/handler"
	courtHandler "api/internal/domains/court/handler"
	creditPackageHandler "api/internal/domains/credit_package/handler"
	discountHandler "api/internal/domains/discount/handler"
	enrollmentHandler "api/internal/domains/enrollment/handler"
	eventHandler "api/internal/domains/event/handler"
	"api/internal/domains/game"
	"api/internal/domains/identity/handler/authentication"
	"api/internal/domains/identity/handler/registration"
	locationsHandler "api/internal/domains/location/handler"
	membership "api/internal/domains/membership/handler"
	notificationHandler "api/internal/domains/notification/handler"
	payment "api/internal/domains/payment/handler"
	paymentMiddleware "api/internal/domains/payment/middleware"
	playground "api/internal/domains/playground/handler"
	practice "api/internal/domains/practice/handler"
	programHandler "api/internal/domains/program"
	schedule "api/internal/domains/schedule/handler"
	teamsHandler "api/internal/domains/team"
	uploadHandler "api/internal/domains/upload/handler"
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
		"/bookings":   RegisterBookingsRoutes,

		// Users & Staff routes
		"/users":     RegisterUserRoutes,
		"/customers": RegisterCustomerRoutes,
		"/athletes":  RegisterAthleteRoutes,
		"/staffs":    RegisterStaffRoutes,
		"/upload":    RegisterUploadRoutes,

		// Haircut routes
		"/haircuts":         RegisterHaircutRoutes,
		"/barbers/services": RegisterBarberServicesRoutes,

		// Purchase-related routes
		"/checkout": RegisterCheckoutRoutes,
		"/subscriptions": RegisterSubscriptionRoutes,
		"/credit_packages": RegisterCreditPackageRoutes,

		// Webhooks
		"/webhooks": RegisterWebhooksRoutes,

		// Contact routes
		"/contact": RegisterContactRoutes,

		//Secure routes
		"/secure": RegisterSecureRoutes,

		// AI proxy route
		"/ai": RegisterAIRoutes,

		// Admin routes  
		"/admin": RegisterAdminRoutes,
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
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/delete-account", h.DeleteMyAccount)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}/notes", h.UpdateCustomerNotes)
	}
}

func RegisterAthleteRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewCustomersHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetAthletes)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Patch("/{id}/stats", h.UpdateAthleteStats)
		r.With(middlewares.JWTAuthMiddleware(true)).Patch("/{id}/profile", h.UpdateAthleteProfile)
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
		r.Route("/barbers", RegisterBarberAvailabilityRoutes(container))
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

func RegisterBarberAvailabilityRoutes(container *di.Container) func(chi.Router) {
	h := haircutEvents.NewEventsHandler(container)

	return func(r chi.Router) {
		// Public endpoint - get available time slots for any barber
		r.Get("/{barber_id}/availability", h.GetAvailableTimeSlots)
		
		// Authenticated barber endpoints - manage own availability
		r.Route("/me", func(r chi.Router) {
			r.Use(middlewares.JWTAuthMiddleware(false, contextUtils.RoleBarber))
			r.Get("/availability", h.GetMyAvailability)
			r.Post("/availability", h.SetMyAvailability)
			r.Post("/availability/bulk", h.BulkSetMyAvailability)
			r.Put("/availability/{id}", h.UpdateMyAvailability)
			r.Delete("/availability/{id}", h.DeleteMyAvailability)
		})
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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Patch("/{id}/visibility", h.ToggleMembershipPlanVisibility)
	}
}

func RegisterGamesRoutes(container *di.Container) func(chi.Router) {
	h := game.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", h.GetGames)
		r.Get("/{id}", h.GetGameById)

		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Post("/", h.CreateGame)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Put("/{id}", h.UpdateGame)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Delete("/{id}", h.DeleteGame)
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
		r.With(middlewares.JWTAuthMiddleware(true)).Patch("/{id}/profile", staffHandlers.UpdateStaffProfile)
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
	creditPkgHandler := creditPackageHandler.NewCreditPackageHandler(container)
	securityMw := paymentMiddleware.NewSecurityMiddleware()
	pciMw := paymentMiddleware.NewPCIComplianceMiddleware()

	return func(r chi.Router) {
		// Apply comprehensive security to all checkout endpoints
		r.Use(securityMw.SecurePaymentEndpoints)
		r.Use(pciMw.EnforcePCICompliance)
		r.Use(pciMw.DataMaskingMiddleware)

		// Rate limit checkout endpoints - 10 requests per minute per user
		r.Use(middlewares.RateLimitMiddleware(10.0/60.0, 3, time.Minute))

		r.With(middlewares.JWTAuthMiddleware(true)).Post("/membership_plans/{id}", h.CheckoutMembership)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/credit_packages/{id}", creditPkgHandler.CheckoutCreditPackage)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/programs/{id}", h.CheckoutProgram)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/events/{id}", h.CheckoutEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/events/{id}/options", h.GetEventEnrollmentOptions)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/events/{id}/enhanced", h.CheckoutEventEnhanced)
	}
}

func RegisterWebhooksRoutes(container *di.Container) func(chi.Router) {
	h := payment.NewWebhookHandlers(container)
	securityMw := paymentMiddleware.NewSecurityMiddleware()

	return func(r chi.Router) {
		// Apply webhook-specific security
		r.Use(securityMw.WebhookSecurityMiddleware)
		
		// Rate limit webhook endpoints - 1000 per hour (Stripe can send many)
		r.Use(middlewares.RateLimitMiddleware(1000.0/3600.0, 10, time.Hour))
		
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
		r.Route("/teams", RegisterSecureTeamRoutes(container))
		r.Route("/schedule", RegisterSecureScheduleRoutes(container))
		r.Route("/credits", RegisterSecureCreditRoutes(container))
		r.Route("/notifications", RegisterSecureNotificationRoutes(container))
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
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/delete-account", h.DeleteMyAccount)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/recover-account", h.RecoverAccount)
	}
}

// RegisterSecureTeamRoutes registers secure team routes that require authentication.
func RegisterSecureTeamRoutes(container *di.Container) func(chi.Router) {
	h := teamsHandler.NewHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/", h.GetMyTeams)
	}
}

// RegisterSecureScheduleRoutes registers the consolidated schedule route.
func RegisterSecureScheduleRoutes(container *di.Container) func(chi.Router) {
	h := schedule.NewHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/", h.GetMySchedule)
	}
}

// RegisterAIRoutes registers the AI proxy route with rate limiting.
func RegisterAIRoutes(_ *di.Container) func(chi.Router) {
	h := aiHandler.NewHandler()
	return func(r chi.Router) {
		r.With(middlewares.RateLimitMiddleware(1.0, 4, time.Minute)).Post("/chat", h.ProxyMessage)
	}
}

// RegisterSecureCreditRoutes registers secure credit routes for customers
func RegisterSecureCreditRoutes(container *di.Container) func(chi.Router) {
	h := userHandler.NewCreditHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/", h.GetCustomerCredits)
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/transactions", h.GetCustomerCreditTransactions)
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/weekly-usage", h.GetWeeklyUsage)
	}
}

// RegisterBookingsRoutes registers the bookings routes.
func RegisterBookingsRoutes(container *di.Container) func(chi.Router) {
	h := bookingsHandler.NewHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/upcoming", h.GetMyUpcomingBookings)
	}
}

// RegisterSubscriptionRoutes registers subscription management routes.
func RegisterSubscriptionRoutes(container *di.Container) func(chi.Router) {
	h := payment.NewSubscriptionHandlers(container)
	securityMw := paymentMiddleware.NewSecurityMiddleware()
	pciMw := paymentMiddleware.NewPCIComplianceMiddleware()

	return func(r chi.Router) {
		// Apply security to subscription endpoints
		r.Use(securityMw.SecurePaymentEndpoints)
		r.Use(pciMw.EnforcePCICompliance)
		
		// Rate limit subscription endpoints - 30 requests per minute per user
		r.Use(middlewares.RateLimitMiddleware(30.0/60.0, 5, time.Minute))
		
		// All subscription endpoints require authentication
		r.Use(middlewares.JWTAuthMiddleware(true))

		// Get all user subscriptions
		r.Get("/", h.GetCustomerSubscriptions)
		
		// Get specific subscription
		r.Get("/{id}", h.GetSubscription)
		
		// Cancel subscription
		r.Post("/{id}/cancel", h.CancelSubscription)
		
		// Pause subscription
		r.Post("/{id}/pause", h.PauseSubscription)
		
		// Resume subscription  
		r.Post("/{id}/resume", h.ResumeSubscription)
		
		// Create customer portal session
		r.Post("/portal", h.CreatePortalSession)
	}
}

// RegisterAdminRoutes registers admin routes requiring admin authentication
func RegisterAdminRoutes(container *di.Container) func(chi.Router) {
	creditHandler := userHandler.NewCreditHandler(container)
	
	return func(r chi.Router) {
		// Apply admin authentication to all routes
		r.Use(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin))
		
		// Credit management routes
		r.Get("/customers/{id}/credits", creditHandler.GetAnyCustomerCredits)
		r.Get("/customers/{id}/credits/transactions", creditHandler.GetAnyCustomerCreditTransactions)
		r.Get("/customers/{id}/credits/weekly-usage", creditHandler.GetAnyCustomerWeeklyUsage)
		r.Post("/customers/{id}/credits/add", creditHandler.AddCustomerCredits)
		r.Post("/customers/{id}/credits/deduct", creditHandler.DeductCustomerCredits)
		r.Get("/events/{id}/credit-transactions", creditHandler.GetEventCreditTransactions)
		r.Put("/events/{id}/credit-cost", creditHandler.UpdateEventCreditCost)
	}
}

func RegisterUploadRoutes(container *di.Container) func(chi.Router) {
	uploadHandlers := uploadHandler.NewUploadHandler(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/image", uploadHandlers.UploadImage)
	}
}

// RegisterSecureNotificationRoutes registers secure notification routes that require authentication.
func RegisterSecureNotificationRoutes(container *di.Container) func(chi.Router) {
	h := notificationHandler.NewNotificationHandler(container)
	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/register", h.RegisterPushToken)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/send", h.SendTeamNotification)
	}
}

// RegisterCreditPackageRoutes registers credit package routes (public viewing, admin management)
func RegisterCreditPackageRoutes(container *di.Container) func(chi.Router) {
	h := creditPackageHandler.NewCreditPackageHandler(container)
	return func(r chi.Router) {
		// Public routes - anyone can view available credit packages
		r.Get("/", h.GetAllCreditPackages)
		r.Get("/{id}", h.GetCreditPackageByID)

		// Admin routes - managing credit packages
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateCreditPackage)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateCreditPackage)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteCreditPackage)
	}
}
