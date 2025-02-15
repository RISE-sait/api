package di

import (
	"api/config"
	courseDb "api/internal/domains/course/persistence/sqlc/generated"
	customerDb "api/internal/domains/customer/persistence/sqlc/generated"
	eventDb "api/internal/domains/events/persistence/sqlc/generated"
	facilityDb "api/internal/domains/facility/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	membershipDb "api/internal/domains/membership/persistence/sqlc/generated"
	staffDb "api/internal/domains/staff/persistence/sqlc/generated"

	"api/internal/services/hubspot"
	"database/sql"
)

type Container struct {
	DB             *sql.DB
	Queries        *QueriesType
	HubspotService *hubspot.HubSpotService
}

type QueriesType struct {
	IdentityDb   *identityDb.Queries
	CoursesDb    *courseDb.Queries
	MembershipDb *membershipDb.Queries
	FacilityDb   *facilityDb.Queries
	EventDb      *eventDb.Queries
	CustomerDb   *customerDb.Queries
	StaffDb      *staffDb.Queries
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
		IdentityDb:   identityDb.New(db),
		CustomerDb:   customerDb.New(db),
		CoursesDb:    courseDb.New(db),
		MembershipDb: membershipDb.New(db),
		FacilityDb:   facilityDb.New(db),
		EventDb:      eventDb.New(db),
		StaffDb:      staffDb.New(db),
	}
}

func (c *Container) Cleanup() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
