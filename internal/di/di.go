package di

import (
	"api/config"
	courseDb "api/internal/domains/course/persistence/sqlc/generated"
	enrollmentDb "api/internal/domains/enrollment/persistence/sqlc/generated"
	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	eventStaffDb "api/internal/domains/event_staff/persistence/sqlc/generated"
	facilityDb "api/internal/domains/facility/persistence/sqlc/generated"
	gameDb "api/internal/domains/game/persistence/sqlc/generated"
	barberDb "api/internal/domains/haircut/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	membershipDb "api/internal/domains/membership/persistence/sqlc/generated"
	practiceDb "api/internal/domains/practice/persistence/sqlc/generated"
	purchaseDb "api/internal/domains/purchase/persistence/sqlc/generated"
	staffDb "api/internal/domains/staff/persistence/sqlc/generated"
	customerDb "api/internal/domains/user/persistence/sqlc/generated"

	"api/internal/services/firebase"
	"api/internal/services/hubspot"
	"database/sql"
)

type Container struct {
	DB              *sql.DB
	DbConnString    string
	Queries         *QueriesType
	HubspotService  *hubspot.Service
	FirebaseService *firebase.Service
}

type QueriesType struct {
	IdentityDb   *identityDb.Queries
	CustomerDb   *customerDb.Queries
	PurchasesDb  *purchaseDb.Queries
	CoursesDb    *courseDb.Queries
	PracticesDb  *practiceDb.Queries
	MembershipDb *membershipDb.Queries
	FacilityDb   *facilityDb.Queries
	EventDb      *eventDb.Queries
	StaffDb      *staffDb.Queries
	EnrollmentDb *enrollmentDb.Queries
	EventStaffDb *eventStaffDb.Queries
	BarberDb     *barberDb.Queries
	GameDb       *gameDb.Queries
}

func NewContainer() *Container {
	db := config.GetDBConnection()
	queries := initializeQueries(db)
	hubspotService := hubspot.GetHubSpotService(nil)
	firebaseService, err := firebase.NewFirebaseService()

	if err != nil {
		panic("Failed to get firebase auth client")
	}

	return &Container{
		DB:              db,
		Queries:         queries,
		HubspotService:  hubspotService,
		FirebaseService: firebaseService,
	}
}

func initializeQueries(db *sql.DB) *QueriesType {
	return &QueriesType{
		IdentityDb:   identityDb.New(db),
		CustomerDb:   customerDb.New(db),
		PurchasesDb:  purchaseDb.New(db),
		CoursesDb:    courseDb.New(db),
		PracticesDb:  practiceDb.New(db),
		MembershipDb: membershipDb.New(db),
		FacilityDb:   facilityDb.New(db),
		EventDb:      eventDb.New(db),
		StaffDb:      staffDb.New(db),
		EnrollmentDb: enrollmentDb.New(db),
		EventStaffDb: eventStaffDb.New(db),
		BarberDb:     barberDb.New(db),
		GameDb:       gameDb.New(db),
	}
}

func (c *Container) Cleanup() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
