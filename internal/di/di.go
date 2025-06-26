package di

import (
	"api/config"
	staffActivityLogsDb "api/internal/domains/audit/staff_activity_logs/persistence/sqlc/generated"
	discountDb "api/internal/domains/discount/persistence/sqlc/generated"
	enrollmentDb "api/internal/domains/enrollment/persistence/sqlc/generated"
	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	gameDb "api/internal/domains/game/persistence/sqlc/generated"
	courtDb "api/internal/domains/court/persistence/sqlc/generated"
	haircutEventDb "api/internal/domains/haircut/event/persistence/sqlc/generated"
	haircutServiceDb "api/internal/domains/haircut/haircut_service/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	locationDb "api/internal/domains/location/persistence/sqlc/generated"
	membershipDb "api/internal/domains/membership/persistence/sqlc/generated"
	purchaseDb "api/internal/domains/payment/persistence/sqlc/generated"
	playgroundDb "api/internal/domains/playground/persistence/sqlc/generated"
	programDb "api/internal/domains/program/persistence/sqlc/generated"
	teamDb "api/internal/domains/team/persistence/sqlc/generated"

	"api/internal/services/square"

	outboxDb "api/internal/services/outbox/generated"

	userDb "api/internal/domains/user/persistence/sqlc/generated"

	"api/internal/services/gcp"
	"api/internal/services/hubspot"
	"database/sql"

	squareClient "github.com/square/square-go-sdk/client"
)

type Container struct {
	DB              *sql.DB
	Queries         *QueriesType
	HubspotService  *hubspot.Service
	FirebaseService *gcp.Service
	SquareClient    *squareClient.Client
}

type QueriesType struct {
	IdentityDb          *identityDb.Queries
	PurchasesDb         *purchaseDb.Queries
	ProgramDb           *programDb.Queries
	MembershipDb        *membershipDb.Queries
	LocationDb          *locationDb.Queries
	EventDb             *eventDb.Queries
	EnrollmentDb        *enrollmentDb.Queries
	HaircutServiceDb    *haircutServiceDb.Queries
	HaircutEventDb      *haircutEventDb.Queries
	GameDb              *gameDb.Queries
	TeamDb              *teamDb.Queries
	PlaygroundDb        *playgroundDb.Queries
	UserDb              *userDb.Queries
	OutboxDb            *outboxDb.Queries
	StaffActivityLogsDb *staffActivityLogsDb.Queries
	DiscountDb          *discountDb.Queries
	CourtDb             *courtDb.Queries
}

// NewContainer initializes and returns a Container with database, queries, HubSpot, Firebase, and Square services.
// Panics if any initialization fails.
//
// Returns:
//   - *Container: The initialized Container.
//
// Example usage:
//
//	container := NewContainer()  // Initializes the container.
func NewContainer() *Container {
	db := config.GetDBConnection()
	queries := initializeQueries(db)
	hubspotService := hubspot.GetHubSpotService(nil)
	firebaseService, err := gcp.NewFirebaseService()

	if err != nil {
		panic(err.Error())
	}

	localSquareClient, err := square.GetSquareClient()

	if err != nil {
		panic(err.Error())
	}

	return &Container{
		DB:              db,
		Queries:         queries,
		HubspotService:  hubspotService,
		FirebaseService: firebaseService,
		SquareClient:    localSquareClient,
	}
}

// initializeQueries initializes and returns a QueriesType struct with database connections for various services.
//
// Returns:
//   - *QueriesType: The initialized QueriesType containing all database connections.
//
// Example usage:
//
//	queries := initializeQueries(db)  // Initializes the queries for all services.
func initializeQueries(db *sql.DB) *QueriesType {
	return &QueriesType{
		IdentityDb:          identityDb.New(db),
		UserDb:              userDb.New(db),
		PurchasesDb:         purchaseDb.New(db),
		ProgramDb:           programDb.New(db),
		MembershipDb:        membershipDb.New(db),
		LocationDb:          locationDb.New(db),
		EventDb:             eventDb.New(db),
		EnrollmentDb:        enrollmentDb.New(db),
		HaircutServiceDb:    haircutServiceDb.New(db),
		HaircutEventDb:      haircutEventDb.New(db),
		GameDb:              gameDb.New(db),
		PlaygroundDb:        playgroundDb.New(db),
		TeamDb:              teamDb.New(db),
		OutboxDb:            outboxDb.New(db),
		StaffActivityLogsDb: staffActivityLogsDb.New(db),
		DiscountDb:          discountDb.New(db),
		CourtDb:             courtDb.New(db),
	}
}

// Cleanup closes the database connection in the Container if it exists.
//
// Returns:
//   - error: Any error that occurs while closing the database connection, or nil if successful.
//
// Example usage:
//
//	err := container.Cleanup()  // Closes the database connections
func (c *Container) Cleanup() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
