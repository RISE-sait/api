package staff

import (
	"api/internal/di"
	entity "api/internal/domains/staff/entity"
	db "api/internal/domains/staff/persistence/sqlc/generated"
	values "api/internal/domains/staff/values"

	// values "api/internal/domains/staff/values/memberships"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Repository struct {
	Queries *db.Queries
}

func NewStaffRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.StaffDb,
	}
}

func (r *Repository) GetByID(c context.Context, id uuid.UUID) (*entity.Staff, *errLib.CommonError) {
	staff, err := r.Queries.GetStaffByID(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Staff not found", http.StatusNotFound)
		}
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entity.Staff{
		ID: staff.ID,
		Details: values.Details{
			IsActive:  staff.IsActive,
			CreatedAt: staff.CreatedAt,
			UpdatedAt: staff.UpdatedAt,
			RoleID:    staff.RoleID,
			RoleName:  staff.RoleName,
		},
	}, nil
}

func (r *Repository) List(ctx context.Context, roleIdPtr *uuid.UUID) ([]entity.Staff, *errLib.CommonError) {

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
		return []entity.Staff{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	staffs := make([]entity.Staff, len(dbStaffs))
	for i, dbStaff := range dbStaffs {
		staffs[i] = entity.Staff{
			ID: dbStaff.ID,
			Details: values.Details{
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

func (r *Repository) Update(c context.Context, staffFields *entity.Staff) (entity.Staff, *errLib.CommonError) {

	dbStaffParams := db.UpdateStaffParams{
		ID:       staffFields.ID,
		RoleID:   staffFields.RoleID,
		IsActive: staffFields.IsActive,
	}

	staff, err := r.Queries.UpdateStaff(c, dbStaffParams)

	if err != nil {
		return entity.Staff{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return entity.Staff{
		ID: staff.ID,
		Details: values.Details{
			IsActive:  staff.IsActive,
			CreatedAt: staff.CreatedAt,
			UpdatedAt: staff.UpdatedAt,
			RoleID:    staff.RoleID,
			RoleName:  staff.RoleName,
		},
	}, nil

}

func (r *Repository) Delete(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteStaff(c, id)

	if err != nil {
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Staff not found", http.StatusNotFound)
	}

	return nil
}
