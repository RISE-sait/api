package test_utils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/pressly/goose"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
)

var (
	dbInstance *sql.DB
	once       sync.Once
	cleanup    func()
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {

	once.Do(func() {
		ctx := context.Background()

		// Start a PostgreSQL container
		req := testcontainers.ContainerRequest{
			Image:        "postgres:13",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "postgres",
				"POSTGRES_PASSWORD": "root",
				"POSTGRES_DB":       "testdb",
			},
			//HostConfigModifier:
			HostConfigModifier: func(config *container.HostConfig) {
				config.Memory = 512 * 1024 * 1024 // 512MB memory limit
				config.CPUCount = 2               // Use 2 CPU cores
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithStartupTimeout(60 * time.Second)}

		_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
		_ = os.Setenv("TESTCONTAINERS_CHECK_CONTAINERS", "false") // Reduce resource usage

		// Create the container
		postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(t, err)

		host, err := postgresC.Host(ctx)
		require.NoError(t, err)

		port, err := postgresC.MappedPort(ctx, "5432")
		require.NoError(t, err)

		dsn := fmt.Sprintf("postgresql://postgres:root@%s:%s/testdb?sslmode=disable", host, port.Port())

		// Open DB connection
		dbInstance, err = sql.Open("postgres", dsn)
		require.NoError(t, err)

		require.NoError(t, retryConnection(dbInstance, 5, 2*time.Second))

		// Return cleanup function to stop and remove the container after tests
		cleanup = func() {

			postgresC.Terminate(ctx)
			dbInstance.Close()
		}
	})

	if dbInstance == nil {
		t.Fatal("Database connection is nil. SetupTestDB failed.")
	}

	return dbInstance, cleanup
}

func SetupTestDbQueries(t *testing.T, path string) (
	testDb *sql.DB,
	cleanup func(),
) {
	// Initialize test database
	testDb, _ = setupTestDB(t)

	// Run migrations
	err := goose.Up(testDb, path)
	require.NoError(t, err)

	// Setup cleanup function
	cleanup = func() {
		// Clean tables in reverse dependency order
		_, err = testDb.Exec(`
		    TRUNCATE 
		        athletic.teams,
        program.customer_enrollment,
        events.customer_enrollment,
        membership.memberships,
        program.programs, 
        events.events,
		playground.sessions,
		playground.systems,
        location.locations,
        users.users
    CASCADE;
		`)
		require.NoError(t, err)
	}

	return
}

func retryConnection(db *sql.DB, retries int, delay time.Duration) error {
	for i := 0; i < retries; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to connect to database after %d retries", retries)
}
