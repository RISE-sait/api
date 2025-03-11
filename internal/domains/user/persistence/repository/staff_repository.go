package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type StaffRepository struct {
	Queries *db.Queries
}

func NewStaffRepository(container *di.Container) *StaffRepository {
	return &StaffRepository{
		Queries: container.Queries.UserDb,
	}
}

func (r *StaffRepository) GetByID(c context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	staff, err := r.Queries.GetStaffByID(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Staff not found", http.StatusNotFound)
		}
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := values.ReadValues{
		ID:          staff.ID,
		FirstName:   staff.FirstName,
		LastName:    staff.LastName,
		IsActive:    staff.IsActive,
		CreatedAt:   staff.CreatedAt,
		UpdatedAt:   staff.UpdatedAt,
		RoleName:    staff.RoleName,
		CountryCode: staff.CountryAlpha2Code,
	}

	if staff.Email.Valid {
		response.Email = staff.Email.String
	}

	if staff.Phone.Valid {
		response.Phone = staff.Phone.String
	}

	return response, nil
}

func (r *StaffRepository) List(ctx context.Context, rolePtr *string) ([]values.ReadValues, *errLib.CommonError) {

	var arg sql.NullString

	if rolePtr != nil {
		arg = sql.NullString{String: *rolePtr, Valid: true}
	}

	dbStaffs, err := r.Queries.GetStaffs(ctx, arg)

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
			IsActive:    dbStaff.IsActive,
			CreatedAt:   dbStaff.CreatedAt,
			UpdatedAt:   dbStaff.UpdatedAt,
			RoleName:    dbStaff.RoleName,
			CountryCode: dbStaff.CountryAlpha2Code,
		}

		if dbStaff.Email.Valid {
			response.Email = dbStaff.Email.String
		}

		if dbStaff.Phone.Valid {
			response.Phone = dbStaff.Phone.String
		}

		staffs[i] = response
	}

	return staffs, nil
}

func (r *StaffRepository) Update(c context.Context, staffFields values.UpdateValues) *errLib.CommonError {

	dbStaffParams := db.UpdateStaffParams{
		ID:       staffFields.ID,
		RoleName: staffFields.RoleName,
		IsActive: staffFields.IsActive,
	}

	if affectedRows, err := r.Queries.UpdateStaff(c, dbStaffParams); err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	} else if affectedRows == 0 {
		return errLib.New("Staff not found", http.StatusNotFound)
	}

	return nil

}

func (r *StaffRepository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteStaff(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Staff not found", http.StatusNotFound)
	}

	return nil
}
