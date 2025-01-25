package main

import (
	"api/config"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

func main() {
	// Print the current working directory

	fmt.Println("Remember to delete the migration containers as docker creates a new container for each migration.")

	fmt.Println("Starting migration...")

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get current working directory: %v", err))
	}

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

		fmt.Println("Migrations applied successfully.")

	case "down":
		// Rollback migration (Down)
		if err := goose.Down(dbConn, migrationsPath); err != nil {
			log.Fatal(fmt.Errorf("failed to roll back migration: %v", err))
		}
		fmt.Println("Migrate down successfully.")

	case "reset":
		// Rollback all migrations (Down)
		if err := goose.DownTo(dbConn, migrationsPath, 0); err != nil {
			log.Fatal(fmt.Errorf("failed to roll back all migrations: %v", err))
		}

		fmt.Println("All migrations rolled back successfully.")
		migrateUp(dbConn, migrationsPath)

		fmt.Println("All migrations applied successfully.")

	default:
		log.Fatal(fmt.Errorf("invalid command: %s. Please pick up or down or reset", command))
	}
}

func migrateUp(dbConn *sql.DB, migrationsPath string) {
	if err := goose.Up(dbConn, migrationsPath); err != nil {
		fmt.Println(fmt.Errorf("failed to apply migrations: %v", err))
	}
	fmt.Println("Migrations applied successfully.")
}
