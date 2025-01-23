package _interface

import (
	courseDb "api/internal/domains/course/infra/persistence/sqlc/generated"
	facilityDb "api/internal/domains/facility/infra/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/authentication/infra/sqlc/generated"
	membershipDb "api/internal/domains/membership/infra/persistence/sqlc/generated"
	membershipPlanDb "api/internal/domains/membership/plans/infra/persistence/sqlc/generated"
)

type QueriesType struct {
	IdentityDb       *identityDb.Queries
	CoursesDb        *courseDb.Queries
	MembershipDb     *membershipDb.Queries
	MembershipPlanDb *membershipPlanDb.Queries
	FacilityDb       *facilityDb.Queries
}
