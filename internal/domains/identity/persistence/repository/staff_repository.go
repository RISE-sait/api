package identity

import (
	databaseErrors "api/internal/constants"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	identityValues "api/internal/domains/identity/values"
	userValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	dbOutbox "api/internal/services/outbox/generated"
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
	IdentityQueries *dbIdentity.Queries
	OutboxQueries   *dbOutbox.Queries
}

func NewStaffRepository(identityDb *dbIdentity.Queries, outboxDb *dbOutbox.Queries) *StaffRepository {
	return &StaffRepository{
		IdentityQueries: identityDb,
		OutboxQueries:   outboxDb,
	}
}

func (r *StaffRepository) CreateApprovedStaff(ctx context.Context, input identityValues.ApprovedStaffRegistrationRequestInfo) *errLib.CommonError {

	args := dbIdentity.CreateApprovedStaffParams{
		ID:       input.UserID,
		RoleName: input.RoleName,
		IsActive: input.IsActive,
	}

	createdStaff, err := r.IdentityQueries.CreateApprovedStaff(ctx, args)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Staff with the ID already exists", http.StatusConflict)
		}

		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if createdStaff.ID == uuid.Nil {
		return errLib.New("Staff not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *StaffRepository) CreatePendingStaff(ctx context.Context, tx *sql.Tx, input identityValues.PendingStaffRegistrationRequestInfo) *errLib.CommonError {

	q := r.OutboxQueries

	if tx != nil {
		q = r.OutboxQueries.WithTx(tx)
	}

	sqlStatement := fmt.Sprintf(
		"CREATE staff (first_name, last_name, age, email, phone, role_name, is_active, country) VALUES ('%s', '%s', '%v', '%s', '%s', '%s', '%v', '%v')",
		input.FirstName, input.LastName, input.Age, input.StaffRegistrationRequestInfo.Email, input.StaffRegistrationRequestInfo.Phone,
		input.RoleName, input.IsActive, input.CountryCode,
	)

	args := dbOutbox.InsertIntoOutboxParams{
		Status:       dbOutbox.AuditStatusPENDING,
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
	dbStaff, err := r.IdentityQueries.GetStaffById(ctx, id)

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
	txQueries := r.IdentityQueries.WithTx(tx)

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
