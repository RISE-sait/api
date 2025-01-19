package repositories

import "api/internal/types"

type Repositories struct {
	Customer        *CustomerRepository
	Course          *CourseRepository
	Facility        *FacilityRepository
	FacilityTypes   *FacilityTypesRepository
	Memberships     *MembershipsRepository
	Schedules       *SchedulesRepository
	MembershipPlans *MembershipPlansRepository
	Waivers         *WaiverRepository
	UserAccounts    *UserOptionalInfoRepository
	Staff           *StaffRepository
	Users           *UsersRepository
}

func NewRepositories(deps *types.Dependencies) *Repositories {
	return &Repositories{
		Customer: &CustomerRepository{
			HubSpotService: deps.HubSpotService,
			Queries:        deps.Queries,
		},
		Course:          &CourseRepository{Queries: deps.Queries},
		Facility:        &FacilityRepository{Queries: deps.Queries},
		FacilityTypes:   &FacilityTypesRepository{Queries: deps.Queries},
		Memberships:     &MembershipsRepository{Queries: deps.Queries},
		Schedules:       &SchedulesRepository{Queries: deps.Queries},
		MembershipPlans: &MembershipPlansRepository{Queries: deps.Queries},
		Waivers:         &WaiverRepository{Queries: deps.Queries},
		UserAccounts:    &UserOptionalInfoRepository{Queries: deps.Queries},
		Staff:           &StaffRepository{Queries: deps.Queries},
		Users:           &UsersRepository{Queries: deps.Queries},
	}
}
