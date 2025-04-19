package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var constraintErrors = map[string]struct {
	Message string
	Status  int
}{
	"staff_id_fkey": {
		Message: "The user doesn't exist for the staff",
		Status:  http.StatusNotFound,
	},
	"staff_role_id_fkey": {
		Message: "The role doesn't exist",
		Status:  http.StatusNotFound,
	},
}

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(container *di.Container) *StaffRepository {
	return &StaffRepository{
		Queries: container.Queries.UserDb,
	}
}

func (r *StaffRepository) List(ctx context.Context, role string) ([]values.ReadValues, *errLib.CommonError) {

	dbStaffs, err := r.Queries.GetStaffs(ctx, sql.NullString{
		String: role,
		Valid:  role != "",
	})

	if err != nil {
		log.Println("Failed to get staffs: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	staffs := make([]values.ReadValues, len(dbStaffs))
	for i, dbStaff := range dbStaffs {
		response := values.ReadValues{
			ID:          dbStaff.ID,
			FirstName:   dbStaff.FirstName,
			LastName:    dbStaff.LastName,
			Email:       dbStaff.Email.String,
			Phone:       dbStaff.Phone.String,
			IsActive:    dbStaff.IsActive,
			CreatedAt:   dbStaff.CreatedAt,
			UpdatedAt:   dbStaff.UpdatedAt,
			RoleName:    dbStaff.RoleName,
			CountryCode: dbStaff.CountryAlpha2Code,
		}

		if dbStaff.Wins.Valid && dbStaff.Losses.Valid {
			response.CoachStatsReadValues = &values.CoachStatsReadValues{
				Wins:   dbStaff.Wins.Int32,
				Losses: dbStaff.Losses.Int32,
			}
		}

		staffs[i] = response
	}

	return staffs, nil
}

func (r *StaffRepository) GetAvailableStaffRoles(ctx context.Context) ([]values.StaffRoleReadValue, *errLib.CommonError) {

	dbRoles, err := r.Queries.GetAvailableStaffRoles(ctx)

	if err != nil {
		log.Println("Failed to get staff roles: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	roles := make([]values.StaffRoleReadValue, len(dbRoles))
	for i, dbRole := range dbRoles {
		response := values.StaffRoleReadValue{
			ID:        dbRole.ID,
			CreatedAt: dbRole.CreatedAt,
			UpdatedAt: dbRole.UpdatedAt,
			RoleName:  dbRole.RoleName,
		}

		roles[i] = response
	}

	return roles, nil
}

func (r *StaffRepository) Update(ctx context.Context, staffFields values.UpdateStaffValues) *errLib.CommonError {

	dbStaffParams := db.UpdateStaffParams{
		ID:       staffFields.ID,
		RoleName: staffFields.RoleName,
		IsActive: staffFields.IsActive,
	}

	if affectedRows, err := r.Queries.UpdateStaff(ctx, dbStaffParams); err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to update staff %+v. Error: %v", staffFields.ID, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	} else if affectedRows == 0 {
		return errLib.New("Staff not found", http.StatusNotFound)
	}

	return nil

}

func (r *StaffRepository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	impactedRows, err := r.Queries.DeleteStaff(c, id)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) {

			if errInfo, found := constraintErrors[pqErr.Constraint]; found {
				return errLib.New(errInfo.Message, errInfo.Status)
			}
		}

		log.Printf("Failed to delete staff with ID: %s. Error: %v", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if impactedRows == 0 {
		return errLib.New("Staff not found", http.StatusNotFound)
	}

	return nil
}
