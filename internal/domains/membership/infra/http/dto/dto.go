package dto

import (
	db "api/internal/domains/membership/infra/persistence/sqlc/generated"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CreateMembershipRequest struct {
	Name        string    `json:"name" validate:"required_and_notwhitespace"`
	Description string    `json:"description" validate:"omitempty,notwhitespace"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required,enddate"`
}

func (r *CreateMembershipRequest) ToDBParams() *db.CreateMembershipParams {

	dbParams := db.CreateMembershipParams{

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

type UpdateMembershipRequest struct {
	Name        string    `json:"name" validate:"required_and_notwhitespace"`
	Description string    `json:"description" validate:"notwhitespace"`
	StartDate   time.Time `json:"start_date" validate:"required"`
	EndDate     time.Time `json:"end_date" validate:"required,enddate"`
}

func (r *UpdateMembershipRequest) ToDBParams() *db.UpdateMembershipParams {

	dbParams := db.UpdateMembershipParams{

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

type MembershipResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}
