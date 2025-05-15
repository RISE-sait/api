package game

import (
	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	repo "api/internal/domains/game/persistence"
	values "api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Service acts as the business logic layer for game operations.
// It coordinates between the repository and audit logging.
type Service struct {
	repo                     *repo.Repository           // Game repository
	staffActivityLogsService *staffActivityLogs.Service // Service to log staff activities
	db                       *sql.DB                    // Database connection for transactions
}

// NewService constructs a new Game Service using the DI container.
func NewService(container *di.Container) *Service {
	return &Service{
		repo:                     repo.NewGameRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		db:                       container.DB,
	}
}

// executeInTx wraps a function execution in a database transaction.
// It ensures atomic updates to both the game and audit logs.
func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx)) // Pass a transaction-aware repository
	})
}

// GetGameById retrieves a single game by its UUID.
func (s *Service) GetGameById(ctx context.Context, id uuid.UUID) (values.ReadGameValue, *errLib.CommonError) {
	return s.repo.GetGameById(ctx, id)
}

// GetGames retrieves a list of all games from the database.
func (s *Service) GetGames(ctx context.Context, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	return s.repo.GetGames(ctx, limit, offset)
}

// GetUpcomingGames retrieves a list of upcoming games.
func (s *Service) GetUpcomingGames(ctx context.Context, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	return s.repo.GetUpcomingGames(ctx, limit, offset)
}

// GetPastGames retrieves a list of past games.
func (s *Service) GetPastGames(ctx context.Context, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	return s.repo.GetPastGames(ctx, limit, offset)
}

// CreateGame adds a new game to the database and logs the activity.
func (s *Service) CreateGame(ctx context.Context, details values.CreateGameValue) *errLib.CommonError {
	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		// Create the game record
		err := txRepo.CreateGame(ctx, details)
		if err != nil {
			return err
		}

		// Get staff user ID from context
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		// Log staff activity for auditing
		if err = s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Created game with details: %+v", details),
		); err != nil {
			return err
		}

		return nil
	})
}

// UpdateGame updates an existing game and logs the modification.
func (s *Service) UpdateGame(ctx context.Context, details values.UpdateGameValue) *errLib.CommonError {
	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		// Update the game record
		if err := txRepo.UpdateGame(ctx, details); err != nil {
			return err
		}

		// Get staff user ID from context
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		// Log the update activity
		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Updated game with ID and new details: %+v", details),
		)
	})
}

// DeleteGame removes a game by ID and logs the deletion.
func (s *Service) DeleteGame(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		// Delete the game
		if err := txRepo.DeleteGame(ctx, id); err != nil {
			return err
		}

		// Get staff user ID from context
		staffID, err := contextUtils.GetUserID(ctx)
		if err != nil {
			return err
		}

		// Log the deletion
		return s.staffActivityLogsService.InsertStaffActivity(
			ctx,
			txRepo.GetTx(),
			staffID,
			fmt.Sprintf("Deleted game with ID: %s", id),
		)
	})
}
