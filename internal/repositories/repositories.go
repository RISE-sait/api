package repositories

import (
	"api/internal/dependencies"
)

type Repositories struct {
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

func NewRepositories(deps *dependencies.Dependencies) *Repositories {
	return &Repositories{
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
