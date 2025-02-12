package staff

import (
	"api/internal/di"
	db "api/internal/domains/staff/persistence/sqlc/generated"
	"api/internal/domains/staff/values"

	// values "api/internal/domains/staff/values/memberships"
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
		Queries: container.Queries.StaffDb,
	}
}

func (r *StaffRepository) GetByID(c context.Context, id uuid.UUID) (*values.StaffAllFields, *errLib.CommonError) {
	staff, err := r.Queries.GetStaffByID(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Staff not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &values.StaffAllFields{
		ID: staff.ID,
		StaffDetails: values.StaffDetails{
			IsActive:  staff.IsActive,
			CreatedAt: staff.CreatedAt,
			UpdatedAt: staff.UpdatedAt,
			RoleID:    staff.RoleID,
			RoleName:  staff.RoleName,
		},
	}, nil
}

func (r *StaffRepository) List(ctx context.Context, roleIdPtr *uuid.UUID) ([]values.StaffAllFields, *errLib.CommonError) {

	roleId := uuid.NullUUID{
		UUID:  uuid.Nil,
		Valid: false,
	}

	if roleIdPtr != nil {
		roleId = uuid.NullUUID{
			UUID:  *roleIdPtr,
			Valid: true,
		}
	}

	dbStaffs, err := r.Queries.GetStaffs(ctx, roleId)

	if err != nil {
		log.Println("Failed to get staffs: ", err.Error())
		return []values.StaffAllFields{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	staffs := make([]values.StaffAllFields, len(dbStaffs))
	for i, dbStaff := range dbStaffs {
		staffs[i] = values.StaffAllFields{
			ID: dbStaff.ID,
			StaffDetails: values.StaffDetails{
				IsActive:  dbStaff.IsActive,
				CreatedAt: dbStaff.CreatedAt,
				UpdatedAt: dbStaff.UpdatedAt,
				RoleID:    dbStaff.RoleID,
				RoleName:  dbStaff.RoleName,
			},
		}
	}

	return staffs, nil
}

func (r *StaffRepository) Update(c context.Context, staffFields *values.StaffAllFields) (values.StaffAllFields, *errLib.CommonError) {

	dbStaffParams := db.UpdateStaffParams{
		ID:       staffFields.ID,
		RoleID:   staffFields.RoleID,
		IsActive: staffFields.IsActive,
	}

	staff, err := r.Queries.UpdateStaff(c, dbStaffParams)

	if err != nil {
		return values.StaffAllFields{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.StaffAllFields{
		ID: staff.ID,
		StaffDetails: values.StaffDetails{
			IsActive:  staff.IsActive,
			CreatedAt: staff.CreatedAt,
			UpdatedAt: staff.UpdatedAt,
			RoleID:    staff.RoleID,
			RoleName:  staff.RoleName,
		},
	}, nil

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
