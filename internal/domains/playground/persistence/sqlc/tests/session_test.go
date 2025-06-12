package playground_tests

import (
	identityDb "api/internal/domains/identity/persistence/sqlc/generated"
	repo "api/internal/domains/playground/persistence"
	dbPlayground "api/internal/domains/playground/persistence/sqlc/generated"
	values "api/internal/domains/playground/values"
	dbTestUtils "api/utils/test_utils"
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createSystem(t *testing.T, db *sql.DB, name string) uuid.UUID {
	id := uuid.New()
	_, err := db.Exec(`INSERT INTO playground.systems (id, name) VALUES ($1, $2)`, id, name)
	require.NoError(t, err)
	return id
}

func createUser(t *testing.T, q *identityDb.Queries) uuid.UUID {
	user, err := q.CreateUser(context.Background(), identityDb.CreateUserParams{
		CountryAlpha2Code:        "CA",
		Email:                    sql.NullString{String: "test@example.com", Valid: true},
		Dob:                      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		Phone:                    sql.NullString{String: "+12312341234", Valid: true},
		HasMarketingEmailConsent: false,
		HasSmsConsent:            false,
		FirstName:                "Test",
		LastName:                 "Customer",
		HubspotID:                sql.NullString{},
		ParentEmail:              sql.NullString{},
	})
	require.NoError(t, err)
	return user.ID
}

func setupRepo(t *testing.T) (*repo.Repository, *identityDb.Queries, *sql.DB, func()) {
	dbConn, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")
	playgroundQueries := dbPlayground.New(dbConn)
	identityQueries := identityDb.New(dbConn)
	repository := &repo.Repository{Queries: playgroundQueries}
	return repository, identityQueries, dbConn, cleanup
}

func TestCreateSession(t *testing.T) {
	repo, identityQ, dbConn, cleanup := setupRepo(t)
	defer cleanup()

	systemID := createSystem(t, dbConn, "Test 1")
	customerID := createUser(t, identityQ)

	start := time.Now().Truncate(time.Second).UTC()
	end := start.Add(time.Hour)

	session, err := repo.CreateSession(context.Background(), values.CreateSessionValue{
		SystemID:   systemID,
		CustomerID: customerID,
		StartTime:  start,
		EndTime:    end,
	})

	require.Nil(t, err)
	require.Equal(t, systemID, session.SystemID)
	require.Equal(t, customerID, session.CustomerID)
	require.WithinDuration(t, start, session.StartTime, time.Second)
	require.WithinDuration(t, end, session.EndTime, time.Second)
}

func TestGetSessions(t *testing.T) {
	repo, identityQ, dbConn, cleanup := setupRepo(t)
	defer cleanup()

	systemID := createSystem(t, dbConn, "Test 1")
	customerID := createUser(t, identityQ)

	start := time.Now().Truncate(time.Second).UTC()
	end := start.Add(time.Hour)

	created, err := repo.CreateSession(context.Background(), values.CreateSessionValue{
		SystemID:   systemID,
		CustomerID: customerID,
		StartTime:  start,
		EndTime:    end,
	})
	require.Nil(t, err)

	sessions, err2 := repo.GetSessions(context.Background())
	require.Nil(t, err2)
	require.Len(t, sessions, 1)
	require.Equal(t, created.ID, sessions[0].ID)
}

func TestGetSession(t *testing.T) {
	repo, identityQ, dbConn, cleanup := setupRepo(t)
	defer cleanup()

	systemID := createSystem(t, dbConn, "Test 1")
	customerID := createUser(t, identityQ)

	start := time.Now().Truncate(time.Second).UTC()
	end := start.Add(time.Hour)

	created, err := repo.CreateSession(context.Background(), values.CreateSessionValue{
		SystemID:   systemID,
		CustomerID: customerID,
		StartTime:  start,
		EndTime:    end,
	})
	require.Nil(t, err)

	got, err2 := repo.GetSession(context.Background(), created.ID)
	require.Nil(t, err2)
	require.Equal(t, created.ID, got.ID)
	require.Equal(t, created.SystemID, got.SystemID)
	require.Equal(t, created.CustomerID, got.CustomerID)
}

func TestDeleteSession(t *testing.T) {
	repo, identityQ, dbConn, cleanup := setupRepo(t)
	defer cleanup()

	systemID := createSystem(t, dbConn, "Test 1")
	customerID := createUser(t, identityQ)

	start := time.Now().Truncate(time.Second).UTC()
	end := start.Add(time.Hour)

	created, err := repo.CreateSession(context.Background(), values.CreateSessionValue{
		SystemID:   systemID,
		CustomerID: customerID,
		StartTime:  start,
		EndTime:    end,
	})
	require.Nil(t, err)

	err2 := repo.DeleteSession(context.Background(), created.ID)
	require.Nil(t, err2)

	_, err3 := repo.GetSession(context.Background(), created.ID)
	require.NotNil(t, err3)
	require.Equal(t, http.StatusNotFound, err3.HTTPCode)
}
