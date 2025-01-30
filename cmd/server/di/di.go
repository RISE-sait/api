package di

import (
	"api/config"
	courseDb "api/internal/domains/course/infra/persistence/sqlc/generated"
	facilityDb "api/internal/domains/facility/infra/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	membershipDb "api/internal/domains/membership/infra/persistence/sqlc/generated"
	membershipPlanDb "api/internal/domains/membership/plans/infra/persistence/sqlc/generated"

	"api/internal/services/hubspot"
	"database/sql"
)

type Container struct {
	DB             *sql.DB
	Queries        *QueriesType
	HubspotService *hubspot.HubSpotService
}

type QueriesType struct {
	IdentityDb       *identityDb.Queries
	CoursesDb        *courseDb.Queries
	MembershipDb     *membershipDb.Queries
	MembershipPlanDb *membershipPlanDb.Queries
	FacilityDb       *facilityDb.Queries
}

func NewContainer() *Container {
	db := config.GetDBConnection()
	queries := initializeQueries(db)
	hubspotService := hubspot.GetHubSpotService()

	return &Container{
		DB:             db,
		Queries:        queries,
		HubspotService: hubspotService,
	}
}

func initializeQueries(db *sql.DB) *QueriesType {
	return &QueriesType{
		IdentityDb:       identityDb.New(db),
		CoursesDb:        courseDb.New(db),
		MembershipDb:     membershipDb.New(db),
		MembershipPlanDb: membershipPlanDb.New(db),
		FacilityDb:       facilityDb.New(db),
	}
}

func (c *Container) Cleanup() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
