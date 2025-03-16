package test_utils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

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

func SetupTestDB(t *testing.T) (*sql.DB, func()) {

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
			WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
		}

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
		dbInstance, err := sql.Open("postgres", dsn)
		require.NoError(t, err)

		require.NoError(t, dbInstance.Ping())

		// Return cleanup function to stop and remove the container after tests
		cleanup = func() {

			postgresC.Terminate(ctx)
		}
	})

	return dbInstance, cleanup
}
