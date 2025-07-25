package identity

import (
	databaseErrors "api/internal/constants"
	"api/internal/di"
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

func NewStaffRepository(container *di.Container) *StaffRepository {
	return &StaffRepository{
		IdentityQueries: container.Queries.IdentityDb,
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
		Dob:       input.DOB,
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

	_, err := r.IdentityQueries.CreatePendingStaff(ctx, args)

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

func (r *StaffRepository) GetStaffRoles(ctx context.Context) ([]string, *errLib.CommonError) {

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

func (r *StaffRepository) GetPendingStaffs(ctx context.Context) ([]identityValues.PendingStaffInfo, *errLib.CommonError) {

	dbStaffs, err := r.IdentityQueries.GetPendingStaffs(ctx)
	if err != nil {
		log.Printf("Error fetching pending staffs: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	staffs := make([]identityValues.PendingStaffInfo, len(dbStaffs))
	for i, s := range dbStaffs {
		var genderPtr *string
		if s.Gender.Valid {
			gender := s.Gender.String
			genderPtr = &gender
		}

		var phonePtr *string
		if s.Phone.Valid {
			phone := s.Phone.String
			phonePtr = &phone
		}

		staffs[i] = identityValues.PendingStaffInfo{
			ID:          s.ID,
			FirstName:   s.FirstName,
			LastName:    s.LastName,
			Email:       s.Email,
			Gender:      genderPtr,
			Phone:       phonePtr,
			CountryCode: s.CountryAlpha2Code,
			RoleID:      s.RoleID,
			CreatedAt:   s.CreatedAt.Time,
			UpdatedAt:   s.UpdatedAt.Time,
			Dob:         s.Dob,
		}
	}

	return staffs, nil
}
