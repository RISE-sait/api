package controllers

import (
	traditional_login "api/internal/controllers/auth/traditional"
	"api/internal/controllers/facilities"
	"api/internal/controllers/memberships"
)

type Controllers struct {
	Facilities    *facilities.FacilitiesController
	FacilityTypes *facilities.FacilityTypesController
	// Schedules        *SchedulesController
	Memberships      *memberships.MembershipsController
	Courses          *CoursesController
	MembershipPlans  *memberships.MembershipPlansController
	Waivers          *WaiversController
	Customers        *CustomersController
	TraditionalLogin *traditional_login.TraditionalLoginController
	// Registration     *AccountRegistrationController
}
