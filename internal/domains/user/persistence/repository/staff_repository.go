package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	values "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
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
