package staff

import (
	"api/internal/domains/staff/values"
	"api/internal/libs/errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

// RepositoryInterface defines the methods that the Repository must implement
type RepositoryInterface interface {
	// GetStaffByUserId fetches staff details by user ID
	GetStaffByUserId(ctx context.Context, id uuid.UUID) (*staff.Details, *errLib.CommonError)

	// GetStaffRolesTx fetches staff roles within a transaction
	GetStaffRolesTx(ctx context.Context, tx *sql.Tx) ([]string, *errLib.CommonError)

	// AssignStaffRoleAndStatusTx assigns a role and status to staff within a transaction
	AssignStaffRoleAndStatusTx(ctx context.Context, tx *sql.Tx, id uuid.UUID, role string, isActive bool) *errLib.CommonError
}
