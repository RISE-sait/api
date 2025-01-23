package _interface

import (
	courseDb "api/internal/domains/course/infra/sqlc"
	identityDb "api/internal/domains/identity/authentication/infra/sqlc/generated"
	membershipDb "api/internal/domains/membership/infra/sqlc/generated"
	membershipPlanDb "api/internal/domains/membership/plans/infra/sqlc/generated"
)

type QueriesType struct {
	IdentityDb       *identityDb.Queries
	CoursesDb        *courseDb.Queries
	MembershipDb     *membershipDb.Queries
	MembershipPlanDb *membershipPlanDb.Queries
}
