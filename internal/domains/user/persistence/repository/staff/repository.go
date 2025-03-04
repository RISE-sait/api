package staff

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values/staff"

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

var _ RepositoryInterface = (*Repository)(nil)

func NewStaffRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.UserDb,
	}
}

func (r *Repository) GetByID(c context.Context, id uuid.UUID) (values.ReadValues, *errLib.CommonError) {
	staff, err := r.Queries.GetStaffByID(c, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return values.ReadValues{}, errLib.New("Staff not found", http.StatusNotFound)
		}
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValues{
		ID:        staff.ID,
		IsActive:  staff.IsActive,
		CreatedAt: staff.CreatedAt,
		UpdatedAt: staff.UpdatedAt,
		RoleID:    staff.RoleID,
		RoleName:  staff.RoleName,
	}, nil
}

func (r *Repository) List(ctx context.Context, role *string, hubspotIds []string) ([]values.ReadValues, *errLib.CommonError) {

	arg := db.GetStaffsParams{
		HubspotIds: hubspotIds,
	}

	if role != nil {
		arg.Role = sql.NullString{String: *role, Valid: true}
	}

	dbStaffs, err := r.Queries.GetStaffs(ctx, arg)

	if err != nil {
		log.Println("Failed to get staffs: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	staffs := make([]values.ReadValues, len(dbStaffs))
	for i, dbStaff := range dbStaffs {
		staffs[i] = values.ReadValues{
			ID:        dbStaff.ID,
			HubspotID: dbStaff.HubspotID,
			IsActive:  dbStaff.IsActive,
			CreatedAt: dbStaff.CreatedAt,
			UpdatedAt: dbStaff.UpdatedAt,
			RoleID:    dbStaff.RoleID,
			RoleName:  dbStaff.RoleName,
		}
	}

	return staffs, nil
}

func (r *Repository) Update(c context.Context, staffFields values.UpdateValues) (values.ReadValues, *errLib.CommonError) {

	dbStaffParams := db.UpdateStaffParams{
		ID:       staffFields.ID,
		RoleName: staffFields.RoleName,
		IsActive: staffFields.IsActive,
	}

	staff, err := r.Queries.UpdateStaff(c, dbStaffParams)

	if err != nil {
		return values.ReadValues{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return values.ReadValues{
		ID:        staff.ID,
		IsActive:  staff.IsActive,
		CreatedAt: staff.CreatedAt,
		UpdatedAt: staff.UpdatedAt,
		RoleID:    staff.RoleID,
		RoleName:  staff.RoleName,
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
