package repository

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(container *di.Container) *StaffRepository {
	return &StaffRepository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *StaffRepository) GetStaffByEmail(ctx context.Context, email string) (*db.GetStaffByEmailRow, *errLib.CommonError) {
	staff, err := r.Queries.GetStaffByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Staff not found", http.StatusNotFound)
		}
		log.Printf("Error fetching staff by email: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &staff, nil
}

func (r *StaffRepository) GetStaffRolesTx(ctx context.Context, tx *sql.Tx) ([]db.StaffRole, *errLib.CommonError) {
	txQueries := r.Queries.WithTx(tx)

	roles, err := txQueries.GetStaffRoles(ctx)

	if err != nil {
		log.Printf("Error fetching staff roles: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return roles, nil
}

func (r *StaffRepository) AssignStaffRoleAndStatusTx(ctx context.Context, tx *sql.Tx, email, role string, isActive bool) *errLib.CommonError {

	params := db.CreateStaffParams{
		Email:    email,
		RoleName: role,
		IsActive: isActive,
	}

	txQueries := r.Queries.WithTx(tx)

	rows, err := txQueries.CreateStaff(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle PostgreSQL unique violation errors (e.g., duplicate staff emails)
			if pqErr.Code == database_errors.UniqueViolation { // Unique violation
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
