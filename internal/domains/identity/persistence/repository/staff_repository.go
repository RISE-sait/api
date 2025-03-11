package identity

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	identityValues "api/internal/domains/identity/values"
	userValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"log"
	"net/http"
)

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(db *db.Queries) *StaffRepository {
	return &StaffRepository{
		Queries: db,
	}
}

func (r *StaffRepository) CreateApprovedStaff(ctx context.Context, input identityValues.ApprovedStaffRegistrationRequestInfo) *errLib.CommonError {

	args := db.CreateApprovedStaffParams{
		ID:       input.UserID,
		RoleName: input.RoleName,
		IsActive: input.IsActive,
	}

	rows, err := r.Queries.CreateApprovedStaff(ctx, args)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Staff with the ID already exists", http.StatusConflict)
		}
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if rows == 0 {
		return errLib.New("Error: Staff not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *StaffRepository) CreatePendingStaff(ctx context.Context, tx *sql.Tx, input identityValues.PendingStaffRegistrationRequestInfo) *errLib.CommonError {

	q := r.Queries

	if tx != nil {
		q = r.Queries.WithTx(tx)
	}

	sqlStatement := fmt.Sprintf(
		"CREATE staff (first_name, last_name, age, email, phone, role_name, is_active, country) VALUES ('%s', '%s', '%v', '%s', '%s', '%s', '%v', '%v')",
		input.FirstName, input.LastName, input.Age, input.StaffRegistrationRequestInfo.Email, input.StaffRegistrationRequestInfo.Phone,
		input.RoleName, input.IsActive, input.CountryCode,
	)

	args := db.InsertIntoOutboxParams{
		Status:       db.AuditStatusPENDING,
		SqlStatement: sqlStatement,
	}

	rows, err := q.InsertIntoOutbox(ctx, args)

	if err != nil {
		log.Printf("Error inserting staff rows: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if rows == 0 {
		return errLib.New("Error: Staff not registered", http.StatusInternalServerError)
	}

	return nil
}

func (r *StaffRepository) GetStaffByUserId(ctx context.Context, id uuid.UUID) (userValues.ReadValues, *errLib.CommonError) {
	dbStaff, err := r.Queries.GetStaffById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userValues.ReadValues{}, errLib.New("Staff not found", http.StatusNotFound)
		}
		log.Printf("Error fetching staff by id: %v", err)
		return userValues.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return userValues.ReadValues{
		ID:        dbStaff.ID,
		HubspotID: dbStaff.HubspotID.String,
		IsActive:  dbStaff.IsActive,
		CreatedAt: dbStaff.CreatedAt,
		UpdatedAt: dbStaff.UpdatedAt,
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
