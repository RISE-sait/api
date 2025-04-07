package identity

import (
	databaseErrors "api/internal/constants"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	identityValues "api/internal/domains/identity/values"
	userValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type StaffRepository struct {
	IdentityQueries *dbIdentity.Queries
}

func NewStaffRepository(identityDb *dbIdentity.Queries) *StaffRepository {
	return &StaffRepository{
		IdentityQueries: identityDb,
	}
}

func (r *StaffRepository) ApproveStaff(ctx context.Context, id uuid.UUID) *errLib.CommonError {

	approvedStaff, err := r.IdentityQueries.ApproveStaff(ctx, id)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("Staff with the ID already exists", http.StatusConflict)
		}

		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if approvedStaff.ID == uuid.Nil {
		return errLib.New("Staff not created", http.StatusInternalServerError)
	}

	return nil
}

func (r *StaffRepository) CreatePendingStaff(ctx context.Context, input identityValues.StaffRegistrationRequestInfo) *errLib.CommonError {

	args := dbIdentity.CreatePendingStaffParams{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Age:       input.Age,
		Phone: sql.NullString{
			String: input.Phone,
			Valid:  input.Phone != "",
		},
		CountryAlpha2Code: input.CountryCode,
		RoleName:          input.RoleName,
	}

	if input.Gender != "" {
		args.Gender = sql.NullString{
			String: input.Gender,
			Valid:  true,
		}
	}

	registeredStaff, err := r.IdentityQueries.CreatePendingStaff(ctx, args)

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				return errLib.New("Staff with the email already exists", http.StatusConflict)
			}
			if pqErr.Constraint == "pending_staff_gender_check" {
				return errLib.New("Invalid gender value", http.StatusBadRequest)
			}
			if pqErr.Constraint == "country_alpha2_code_check" {
				return errLib.New("Invalid country code", http.StatusBadRequest)
			}
		}

		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error creating pending staff: %v", err)
			return errLib.New("Staff not registered", http.StatusInternalServerError)
		}
		log.Printf("Error inserting staff rows: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if registeredStaff.ID == uuid.Nil {
		log.Printf("Error creating pending staff: %v", err)
		return errLib.New("Staff not registered", http.StatusInternalServerError)
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

func (r *StaffRepository) GetStaffRolesTx(ctx context.Context) ([]string, *errLib.CommonError) {

	dbRoles, err := r.IdentityQueries.GetStaffRoles(ctx)

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
