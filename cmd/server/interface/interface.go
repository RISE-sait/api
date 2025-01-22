package _interface

import (
	courseDb "api/internal/domains/course/infra/sqlc"
	identityDb "api/internal/domains/identity/authentication/infra/sqlc/generated"
)

type QueriesType struct {
	IdentityDb *identityDb.Queries
	CoursesDb  *courseDb.Queries
}
