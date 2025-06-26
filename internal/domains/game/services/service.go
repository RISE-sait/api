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
	"net/http"

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
func (s *Service) GetGames(ctx context.Context, filter values.GetGamesFilter) ([]values.ReadGameValue, *errLib.CommonError) {
	return s.repo.GetGames(ctx, filter)
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

// GetUserGames retrieves games for a specific user based on their role.
// Coaches receive games for teams they coach, while athletes receive games
// for the team they belong to.
func (s *Service) GetUserGames(ctx context.Context, userID uuid.UUID, role contextUtils.CtxRole, limit, offset int32) ([]values.ReadGameValue, *errLib.CommonError) {
	teamIDs, err := s.getUserTeamIDs(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	if len(teamIDs) == 0 {
		return []values.ReadGameValue{}, nil
	}

	games, err := s.repo.GetGamesByTeams(ctx, teamIDs, limit, offset)
	if err != nil {
		return nil, err
	}
	return games, nil
}

func (s *Service) getUserTeamIDs(ctx context.Context, userID uuid.UUID, role contextUtils.CtxRole) ([]uuid.UUID, *errLib.CommonError) {
	switch role {
	case contextUtils.RoleCoach:
		rows, err := s.db.QueryContext(ctx, `SELECT id FROM athletic.teams WHERE coach_id = $1`, userID)
		if err != nil {
			return nil, errLib.New("failed to get coach teams", http.StatusInternalServerError)
		}
		defer rows.Close()
		var ids []uuid.UUID
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return nil, errLib.New("failed to scan team id", http.StatusInternalServerError)
			}
			ids = append(ids, id)
		}
		return ids, nil
	case contextUtils.RoleAthlete:
		var id uuid.UUID
		err := s.db.QueryRowContext(ctx, `SELECT team_id FROM athletic.athletes WHERE id = $1`, userID).Scan(&id)
		if err == sql.ErrNoRows {
			return []uuid.UUID{}, nil
		}
		if err != nil {
			return nil, errLib.New("failed to get athlete team", http.StatusInternalServerError)
		}
		if id == uuid.Nil {
			return []uuid.UUID{}, nil
		}
		return []uuid.UUID{id}, nil
	default:
		return nil, errLib.New("role not supported", http.StatusForbidden)
	}
}
