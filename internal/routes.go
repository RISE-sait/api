package routes

import (
	"api/internal/controllers"
	"api/internal/controllers/auth/oauth"
	traditional_login "api/internal/controllers/auth/traditional"
	"api/internal/controllers/facilities"
	"api/internal/controllers/memberships"
	"api/internal/repositories"
	"api/internal/types"
	"net/http"

	"github.com/go-chi/chi"
)

type RouteConfig struct {
	Path      string
	Configure func(chi.Router)
}

func RegisterRoutes(router *chi.Mux, dependencies *types.Dependencies) {

	repositories := repositories.NewRepositories(dependencies)

	registrationController := controllers.AccountRegistrationController{
		UsersRespository:           repositories.Users,
		StaffRepository:            repositories.Staff,
		UserOptionalInfoRepository: repositories.UserAccounts,
		DB:                         dependencies.DB,
	}

	ctrls := initiateControllers(repositories)

	router.Route("/api", func(r chi.Router) {
		routes := []RouteConfig{
			{Path: "/facilities", Configure: configureFacilitiesRoutes(ctrls.Facilities)},
			{Path: "/facilities/types", Configure: configureFacilitiesTypesRoutes(ctrls.FacilityTypes)},
			// {Path: "/schedules", Configure: configureSchedulesRoutes(ctrls.Schedules)},
			{Path: "/memberships", Configure: configureMembershipsRoutes(ctrls.Memberships)},
			{Path: "/courses", Configure: configureCoursesRoutes(ctrls.Courses)},
			{Path: "/memberships/plans", Configure: configureMembershipPlansRoutes(ctrls.MembershipPlans)},
			{Path: "/waivers", Configure: configureWaiversRoutes(ctrls.Waivers)},
			{Path: "/customers", Configure: configureCustomersRoutes(ctrls.Customers)},
		}

		for _, route := range routes {
			r.Route(route.Path, route.Configure)
		}

		configureAuthRoutes(r, repositories, ctrls)

		r.Route("/auth/traditional/register", func(r chi.Router) {

			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				registrationController.CreateTraditionalAccount(w, r)
			})
		})
	})
}

func initiateControllers(repos *repositories.Repositories) *controllers.Controllers {
	return &controllers.Controllers{
		Facilities:    facilities.NewFacilitiesController(repos.Facility),
		FacilityTypes: facilities.NewFacilityTypesController(repos.FacilityTypes),
		// Schedules:       controllers.NewSchedulesController(repos.Schedules),
		Memberships:     memberships.NewMembershipsController(repos.Memberships),
		Courses:         controllers.NewCoursesController(repos.Course),
		MembershipPlans: memberships.NewMembershipPlansController(repos.MembershipPlans),
		Waivers:         controllers.NewWaiversController(repos.Waivers),
		Customers:       controllers.NewCustomersController(repos.Customer),
		TraditionalLogin: traditional_login.NewTraditionalLoginController(
			repos.UserAccounts,
			repos.Staff,
		),
	}
}

func configureFacilitiesRoutes(ctrl *facilities.FacilitiesController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllFacilities)
		r.Post("/", ctrl.CreateFacility)
		r.Get("/{id}", ctrl.GetFacility)
		r.Put("/", ctrl.UpdateFacility)
		r.Delete("/{id}", ctrl.DeleteFacility)
	}
}

func configureFacilitiesTypesRoutes(ctrl *facilities.FacilityTypesController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllFacilityTypes)
		r.Post("/", ctrl.CreateFacilityType)
		r.Get("/{id}", ctrl.GetFacilityTypeByID)
		r.Put("/", ctrl.UpdateFacilityType)
	}
}

func configureCustomersRoutes(ctrl *controllers.CustomersController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetCustomers)
		r.Get("/email/{email}", ctrl.GetCustomerByEmail)
		r.Get("/id/{id}", ctrl.GetCustomerById)
		r.Post("/", ctrl.CreateCustomer)
	}
}

func configureMembershipsRoutes(ctrl *memberships.MembershipsController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllMemberships)
		r.Get("/{id}", ctrl.GetMembershipById)
		r.Post("/", ctrl.CreateMembership)
		r.Put("/", ctrl.UpdateMembership)
		r.Delete("/{id}", ctrl.DeleteMembership)
	}
}

func configureMembershipPlansRoutes(ctrl *memberships.MembershipPlansController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetMembershipPlanDetails)
		r.Post("/", ctrl.CreateMembershipPlan)
		r.Put("/", ctrl.UpdateMembershipPlan)
		r.Delete("/", ctrl.DeleteMembershipPlan)
	}
}

func configureCoursesRoutes(ctrl *controllers.CoursesController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllCourses)
		r.Get("/{id}", ctrl.GetCourseById)
		r.Post("/", ctrl.CreateCourse)
		r.Put("/{id}", ctrl.UpdateCourse)
		r.Delete("/{id}", ctrl.DeleteCourse)
	}
}

// func configureSchedulesRoutes(ctrl *controllers.SchedulesController) func(chi.Router) {
// 	return func(r chi.Router) {
// 		r.Get("/", ctrl.GetAllSchedules)
// 		r.Post("/", ctrl.CreateSchedule)
// 		r.Get("/{id}", ctrl.GetScheduleByID)
// 		r.Put("/{id}", ctrl.UpdateSchedule)
// 		r.Delete("/{id}", ctrl.DeleteSchedule)
// 	}
// }

func configureWaiversRoutes(ctrl *controllers.WaiversController) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", ctrl.GetAllUniqueWaivers)
		r.Get("/signed", ctrl.GetWaiverByEmailAndDocLink)
		r.Patch("/signed", ctrl.UpdateWaiverStatus)
	}
}

func configureAuthRoutes(r chi.Router, repos *repositories.Repositories, ctrls *controllers.Controllers) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/google/callback", func(w http.ResponseWriter, r *http.Request) {
			oauth.HandleOAuthCallback(w, r, repos.Staff)
		})
		r.Post("/traditional/login", ctrls.TraditionalLogin.GetUser)
		// r.Post("/traditional/register", ctrls.Registration.CreateTraditionalAccount)
	})
}
