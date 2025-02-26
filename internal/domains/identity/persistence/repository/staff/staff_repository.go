package staff

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	"api/internal/domains/identity/persistence/sqlc/generated"
	values "api/internal/domains/staff/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
}

func NewStaffRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.IdentityDb,
	}
}

var _ RepositoryInterface = (*Repository)(nil)

func (r *Repository) GetStaffByUserId(ctx context.Context, id uuid.UUID) (*values.Details, *errLib.CommonError) {
	dbStaff, err := r.Queries.GetStaffById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Staff not found", http.StatusNotFound)
		}
		log.Printf("Error fetching staff by email: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &values.Details{
		RoleName: dbStaff.RoleName,
		IsActive: dbStaff.IsActive,
	}, nil
}

func (r *Repository) GetStaffRolesTx(ctx context.Context, tx *sql.Tx) ([]string, *errLib.CommonError) {
	txQueries := r.Queries.WithTx(tx)

	dbRoles, err := txQueries.GetStaffRoles(ctx)

	if err != nil {
		log.Printf("Error fetching staff roles: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var roles []string

	for _, dbRole := range dbRoles {
		roles = append(roles, dbRole.RoleName)
	}

	return roles, nil
}

func (r *Repository) AssignStaffRoleAndStatusTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, role string, isActive bool) *errLib.CommonError {

	params := db.CreateStaffParams{
		ID:       id,
		RoleName: role,
		IsActive: isActive,
	}

	txQueries := r.Queries.WithTx(tx)

	rows, err := txQueries.CreateStaff(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle PostgreSQL unique violation errors (e.g., duplicate staff emails)
			if pqErr.Code == databaseErrors.UniqueViolation { // Unique violation
				return errLib.New("Staff with this email already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating staff: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if rows == 0 {
		log.Println("Error creating staff ", err)
		return errLib.New("Failed to create staff", 500)
	}

	return nil
}
