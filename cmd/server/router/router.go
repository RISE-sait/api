package router

import (
	"api/internal/di"
	adminHandler "api/internal/domains/admin/handler"
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
	"api/internal/domains/identity/handler/email_verification"
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
	subsidyHandler "api/internal/domains/subsidy/handler"
	teamsHandler "api/internal/domains/team"
	uploadHandler "api/internal/domains/upload/handler"
	userHandler "api/internal/domains/user/handler"
	waiverHandler "api/internal/domains/waiver/handler"
	websitePromoHandler "api/internal/domains/website_promo/handler"
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
		"/subsidies": RegisterSubsidyRoutes,

		// Payment & Reporting routes
		"/admin/payments": RegisterPaymentReportsRoutes,

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

		// Waiver routes
		"/waivers": RegisterWaiverRoutes,

		// Website promo routes (public + admin)
		"/website": RegisterWebsitePromoRoutes,
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
	suspensionHandler := userHandler.NewSuspensionHandler(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/", h.GetCustomers)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/id/{id}", h.GetCustomerByID)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/email/{email}", h.GetCustomerByEmail)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/checkin/{id}", h.CheckinCustomer)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/{id}/memberships", h.GetMembershipHistory)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/{id}/archive", h.ArchiveCustomer)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/{id}/unarchive", h.UnarchiveCustomer)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/archived", h.ListArchivedCustomers)
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/delete-account", h.DeleteMyAccount)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Put("/{id}/notes", h.UpdateCustomerNotes)

		// Suspension routes
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/{id}/suspend", suspensionHandler.SuspendUser)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/{id}/unsuspend", suspensionHandler.UnsuspendUser)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/{id}/collect-arrears", suspensionHandler.CollectArrears)
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/{id}/suspension", suspensionHandler.GetSuspensionInfo)
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
		// IMPORTANT: Specific routes must come before wildcard routes like /{id}

		// Search route
		r.Get("/search", h.SearchTeams)

		// External teams routes
		r.Get("/external", h.GetExternalTeams)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Post("/external", h.CreateExternalTeam)

		// Public routes
		r.Get("/", h.GetTeams)
		r.Get("/{id}", h.GetTeamByID)

		// Team management - coaches and admins can create/manage teams
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Post("/", h.CreateTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Put("/{id}", h.UpdateTeam)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleCoach)).Delete("/{id}", h.DeleteTeam)
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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/", h.CreateProgram)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Put("/{id}", h.UpdateProgram)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{id}", h.DeleteProgram)
	}
}

func RegisterStaffRoutes(container *di.Container) func(chi.Router) {
	staffHandlers := userHandler.NewStaffHandlers(container)
	staffLogsHandlers := staff_activity_logs.NewHandler(container)

	return func(r chi.Router) {
		r.Get("/", staffHandlers.GetStaffs) // Public endpoint for website to display coaches
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/logs", staffLogsHandlers.GetStaffActivityLogs)

		r.With(middlewares.JWTAuthMiddleware(false)).Put("/{id}", staffHandlers.UpdateStaff)
		r.With(middlewares.JWTAuthMiddleware(true)).Patch("/{id}/profile", staffHandlers.UpdateStaffProfile)
		r.With(middlewares.JWTAuthMiddleware(false)).Delete("/{id}", staffHandlers.DeleteStaff)
	}
}

func RegisterEventRoutes(container *di.Container) func(chi.Router) {
	handler := eventHandler.NewEventsHandler(container)
	notificationHandler := eventHandler.NewEventNotificationHandler(container)

	return func(r chi.Router) {
		r.Get("/", handler.GetEvents)
		r.Get("/{id}", handler.GetEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/one-time", handler.CreateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Put("/{id}", handler.UpdateEvent)
		r.With(middlewares.JWTAuthMiddleware(true)).Delete("/", handler.DeleteEvents)

		r.Route("/{event_id}/staffs", RegisterEventStaffRoutes(container))

		// Get enrolled customers for event notifications
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleCoach, contextUtils.RoleReceptionist)).Get("/{event_id}/customers", notificationHandler.GetEventCustomers)

		// Customer enrollment management (remove from event)
		r.Route("/{event_id}/customers", RegisterEventCustomerRoutes(container))

		// Event notification routes
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleCoach)).Post("/{event_id}/notifications", notificationHandler.SendNotification)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleCoach)).Get("/{event_id}/notifications", notificationHandler.GetNotificationHistory)

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

func RegisterEventCustomerRoutes(container *di.Container) func(chi.Router) {
	h := enrollmentHandler.NewCustomerEnrollmentHandler(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/{customer_id}", h.RemoveCustomerFromEvent)
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

		// Checkout verification endpoint - called by frontend after redirect from Stripe
		// This ensures enrollment is complete even if webhooks fail
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/verify/{session_id}", h.VerifyCheckoutSession)
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
	verificationHandler := email_verification.NewEmailVerificationHandler(container)

	return func(r chi.Router) {
		r.Post("/", h.Login)
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/child/{id}", h.LoginAsChild)

		// Email verification routes (public endpoints)
		r.Post("/verify-email", verificationHandler.VerifyEmail)
		r.Post("/resend-verification", verificationHandler.ResendVerificationEmail)
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
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/staff/pending", staffHandler.GetPendingStaffs)
		r.With(middlewares.JWTAuthMiddleware(false)).Post("/staff/approve/{id}", staffHandler.ApproveStaff)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Delete("/staff/reject/{id}", staffHandler.DeletePendingStaff)
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
	firebaseCleanupHandler := adminHandler.NewFirebaseCleanupHandler(container)

	return func(r chi.Router) {
		// Credit management routes - receptionist can view
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/customers/{id}/credits", creditHandler.GetAnyCustomerCredits)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/customers/{id}/credits/transactions", creditHandler.GetAnyCustomerCreditTransactions)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/customers/{id}/credits/weekly-usage", creditHandler.GetAnyCustomerWeeklyUsage)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/customers/{id}/credits/add", creditHandler.AddCustomerCredits)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/customers/{id}/credits/deduct", creditHandler.DeductCustomerCredits)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist)).Get("/events/{id}/credit-transactions", creditHandler.GetEventCreditTransactions)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Put("/events/{id}/credit-cost", creditHandler.UpdateEventCreditCost)

		// Firebase cleanup - IT and SuperAdmin only (sensitive operation)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/firebase/cleanup", firebaseCleanupHandler.CleanupOrphanedFirebaseUsers)

		// Firebase recovery - IT and SuperAdmin only (recreates missing Firebase users from DB)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/firebase/recover", firebaseCleanupHandler.RecoverMissingFirebaseUsers)
	}
}

func RegisterUploadRoutes(container *di.Container) func(chi.Router) {
	uploadHandlers := uploadHandler.NewUploadHandler(container)

	return func(r chi.Router) {
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/image", uploadHandlers.UploadImage)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin)).Post("/program-photo", uploadHandlers.UploadProgramPhoto)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/promo-image", uploadHandlers.UploadPromoImage)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/promo-video", uploadHandlers.UploadPromoVideo)
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

func RegisterSubsidyRoutes(container *di.Container) func(chi.Router) {
	h := subsidyHandler.NewSubsidyHandler(container)

	return func(r chi.Router) {
		// Customer routes - check balance and history
		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTAuthMiddleware(true)) // Require auth
			r.Use(middlewares.RateLimitMiddleware(10, 20, 1*time.Minute)) // 10 rps, burst 20
			r.Get("/me", h.GetMySubsidies)
			r.Get("/me/balance", h.GetMyBalance)
			r.Get("/me/usage", h.GetMyUsageHistory)
		})

		// Admin routes - manage providers and subsidies
		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist))
			r.Use(middlewares.RateLimitMiddleware(5, 10, 1*time.Minute)) // 5 rps, burst 10

			// Provider management - receptionist can only view
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/providers", h.CreateProvider)
			r.Get("/providers", h.ListProviders)
			r.Get("/providers/{id}", h.GetProvider)
			r.Get("/providers/{id}/stats", h.GetProviderStats)

			// Subsidy management - receptionist can only view
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/", h.CreateSubsidy)
			r.Get("/", h.ListSubsidies)
			r.Get("/{id}", h.GetSubsidy)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/{id}/deactivate", h.DeactivateSubsidy)

			// Summary/reports
			r.Get("/summary", h.GetSubsidySummary)
		})
	}
}

// RegisterPaymentReportsRoutes registers payment reporting routes
func RegisterPaymentReportsRoutes(container *di.Container) func(chi.Router) {
	h := payment.NewPaymentReportsHandler(container)
	return func(r chi.Router) {
		// All routes require admin authentication - receptionist can view
		r.Use(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT, contextUtils.RoleReceptionist))
		r.Use(middlewares.RateLimitMiddleware(5, 10, 1*time.Minute)) // 5 rps, burst 10

		// Transaction listing and details
		r.Get("/transactions", h.ListPaymentTransactions)

		// Summary and analytics
		r.Get("/summary", h.GetPaymentSummary)
		r.Get("/summary/by-type", h.GetPaymentSummaryByType)
		r.Get("/subsidy-usage", h.GetSubsidyUsageSummary)

		// Export
		r.Get("/export", h.ExportPaymentTransactions)

		// Backfill URLs from Stripe (admin only, not receptionist)
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/backfill-urls", h.BackfillPaymentURLs)
	}
}

// RegisterWaiverRoutes registers waiver upload and management routes
func RegisterWaiverRoutes(container *di.Container) func(chi.Router) {
	h := waiverHandler.NewWaiverHandler(container)
	return func(r chi.Router) {
		// Upload waiver - authenticated users can upload for themselves, staff can upload for others
		r.With(middlewares.JWTAuthMiddleware(true)).Post("/upload", h.UploadWaiver)

		// Get waivers for a user - users can view their own, staff can view anyone's
		r.With(middlewares.JWTAuthMiddleware(true)).Get("/user/{user_id}", h.GetUserWaivers)

		// Delete waiver - admin only
		r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Delete("/{id}", h.DeleteWaiver)
	}
}

// RegisterWebsitePromoRoutes registers website promo routes
func RegisterWebsitePromoRoutes(container *di.Container) func(chi.Router) {
	h := websitePromoHandler.NewWebsitePromoHandler(container)
	return func(r chi.Router) {
		// Public routes - get active promos for website
		r.Get("/hero-promos/active", h.GetActiveHeroPromos)
		r.Get("/feature-cards/active", h.GetActiveFeatureCards)
		r.Get("/promo-videos/active", h.GetActivePromoVideos)

		// Admin routes - full CRUD
		r.Route("/hero-promos", func(r chi.Router) {
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Get("/", h.GetAllHeroPromos)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Get("/{id}", h.GetHeroPromoById)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/", h.CreateHeroPromo)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Put("/{id}", h.UpdateHeroPromo)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Delete("/{id}", h.DeleteHeroPromo)
		})

		r.Route("/feature-cards", func(r chi.Router) {
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Get("/", h.GetAllFeatureCards)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Get("/{id}", h.GetFeatureCardById)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/", h.CreateFeatureCard)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Put("/{id}", h.UpdateFeatureCard)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Delete("/{id}", h.DeleteFeatureCard)
		})

		r.Route("/promo-videos", func(r chi.Router) {
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Get("/", h.GetAllPromoVideos)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Get("/{id}", h.GetPromoVideoById)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Post("/", h.CreatePromoVideo)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Put("/{id}", h.UpdatePromoVideo)
			r.With(middlewares.JWTAuthMiddleware(false, contextUtils.RoleAdmin, contextUtils.RoleSuperAdmin, contextUtils.RoleIT)).Delete("/{id}", h.DeletePromoVideo)
		})
	}
}
