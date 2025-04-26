package main

/*
This file provides a database migration tool that uses Goose to manage database migrations.

Usage:
    migrate [command]

Commands:
    up      - Apply all pending migrations
    down    - Rollback the last applied migration
    reset   - Rollback all migrations and reapply them from scratch

The tool expects migrations to be located in the 'db/migrations' directory relative to the
current working directory. Migration files should follow Goose naming conventions:
    YYYYMMDDHHMMSS_description.sql

Run:

`goose create <migration_name> sql -dir ./db/migrations`
in the root directory to create a new migration file.

eg: `goose create event_is_date_time_modified_to_is_modified sql -dir ./db/migrations`

Example:
    migrate up     # Apply all pending migrations
    migrate down   # Rollback the most recent migration
    migrate reset  # Reset and reapply all migrations

Environment:
    The database connection is configured through environment variables using the config package.
    See config.GetDBConnection() for required environment variables.

Error Handling:
    - Validates migrations directory exists before execution
    - Reports detailed error messages for common failures
    - Ensures proper command argument is provided
*/

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"api/config"

	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

func main() {
	// Print the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get current working directory: %v", err))
	}
	fmt.Println("Current directory:", dir)

	// Get the database connection
	dbConn := config.GetDBConnection()

	// Define the migrations path
	migrationsPath := "db/migrations"

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		log.Fatal(fmt.Errorf("migrations directory '%s' does not exist in the current directory '%s'", migrationsPath, dir))
	}

	if len(os.Args) < 2 {
		log.Fatal("Please specify an argument: 'up' or 'down'.")
	}

	command := os.Args[1]

	switch command {
	case "up":
		migrateUp(dbConn, migrationsPath)
	case "down":
		// Rollback migration (Down)
		if err := goose.Down(dbConn, migrationsPath); err != nil {
			log.Fatal(fmt.Errorf("failed to roll back migration: %v", err))
			return
		}
		fmt.Println("Migrate down successfully.")

	case "reset":
		// Rollback all migrations (Down)
		if err := goose.DownTo(dbConn, migrationsPath, 0); err != nil {
			log.Fatal(fmt.Errorf("failed to roll back all migrations: %v", err))
		}

		log.Println("All migrations rolled back successfully.")
		migrateUp(dbConn, migrationsPath)

		fmt.Println("All migrations applied successfully.")

	default:
		log.Fatal(fmt.Errorf("invalid command: %s. Please pick up or down or reset", command))
	}
}

func migrateUp(dbConn *sql.DB, migrationsPath string) {
	if err := goose.Up(dbConn, migrationsPath); err != nil {
		log.Fatal(fmt.Errorf("failed to apply migrations: %v", err))
		return
	}
	fmt.Println("Migrations applied successfully.")
}
