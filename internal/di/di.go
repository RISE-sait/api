package di

import (
	"api/config"
	eventDb "api/internal/domains/event/persistence/sqlc/generated"
	gameDb "api/internal/domains/game/persistence/sqlc/generated"
	barberDb "api/internal/domains/haircut/persistence/sqlc/generated"
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	locationDb "api/internal/domains/location/persistence/sqlc/generated"
	membershipDb "api/internal/domains/membership/persistence/sqlc/generated"
	purchaseDb "api/internal/domains/payment/persistence/sqlc/generated"
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
	DbConnString    string
	Queries         *QueriesType
	HubspotService  *hubspot.Service
	FirebaseService *gcp.Service
	SquareClient    *squareClient.Client
}

type QueriesType struct {
	IdentityDb   *identityDb.Queries
	PurchasesDb  *purchaseDb.Queries
	ProgramDb    *programDb.Queries
	MembershipDb *membershipDb.Queries
	LocationDb   *locationDb.Queries
	EventDb      *eventDb.Queries
	BarberDb     *barberDb.Queries
	GameDb       *gameDb.Queries
	TeamDb       *teamDb.Queries
	UserDb       *userDb.Queries
	OutboxDb     *outboxDb.Queries
}

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

func initializeQueries(db *sql.DB) *QueriesType {
	return &QueriesType{
		IdentityDb:   identityDb.New(db),
		UserDb:       userDb.New(db),
		PurchasesDb:  purchaseDb.New(db),
		ProgramDb:    programDb.New(db),
		MembershipDb: membershipDb.New(db),
		LocationDb:   locationDb.New(db),
		EventDb:      eventDb.New(db),
		BarberDb:     barberDb.New(db),
		GameDb:       gameDb.New(db),
		TeamDb:       teamDb.New(db),
		OutboxDb:     outboxDb.New(db),
	}
}

func (c *Container) Cleanup() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}
