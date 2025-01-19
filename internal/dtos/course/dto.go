package course

import (
	db "api/sqlc"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CreateCourseRequestBody struct {
	Name        string    `json:"name" validate:"required_and_notwhitespace"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" `
	EndDate     time.Time `json:"end_date" validate:"required,enddate"`
}

func (r *CreateCourseRequestBody) ToDBParams() *db.CreateCourseParams {

	dbParams := db.CreateCourseParams{
		Name: r.Name,
		Description: sql.NullString{
			String: r.Description,
			Valid:  r.Description != "",
		},
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
	}

	return &dbParams
}

type UpdateCourseRequest struct {
	Name        string    `json:"name" validate:"required_and_notwhitespace"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required"`
	ID          uuid.UUID `json:"id" validate:"required"`
}

func (r *UpdateCourseRequest) ToDBParams() *db.UpdateCourseParams {

	dbParams := db.UpdateCourseParams{
		Name: r.Name,
		Description: sql.NullString{
			String: r.Description,
			Valid:  r.Description != "",
		},
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
	}

	return &dbParams
}
