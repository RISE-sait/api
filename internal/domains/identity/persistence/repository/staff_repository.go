package identity

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
	"api/internal/domains/identity/persistence/sqlc/generated"
	values "api/internal/domains/user/values/staff"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(container *di.Container) *StaffRepository {
	return &StaffRepository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *StaffRepository) GetStaffByUserId(ctx context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	dbStaff, err := r.Queries.GetStaffById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Staff not found", http.StatusNotFound)
		}
		log.Printf("Error fetching staff by id: %v", err)
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValues{
		ID:        dbStaff.ID,
		HubspotID: dbStaff.HubspotID,
		IsActive:  dbStaff.IsActive,
		CreatedAt: dbStaff.CreatedAt,
		UpdatedAt: dbStaff.UpdatedAt,
		RoleID:    dbStaff.RoleID,
		RoleName:  dbStaff.RoleName,
	}, nil
}

func (r *StaffRepository) GetStaffRolesTx(ctx context.Context, tx *sql.Tx) ([]string, *errLib.CommonError) {
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

func (r *StaffRepository) AssignStaffRoleAndStatusTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, role string, isActive bool) *errLib.CommonError {

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
