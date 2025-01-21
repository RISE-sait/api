package dto

import (
	db "api/sqlc"
)

type UpdateWaiverRequest struct {
	Email    string `json:"email" validate:"required,email"`
	IsSigned bool   `json:"signed_status" validate:"required"`
}

func (r *UpdateWaiverRequest) ToDBParams() *db.UpdateWaiverSignedStatusByEmailParams {

	return &db.UpdateWaiverSignedStatusByEmailParams{
		Email:    r.Email,
		IsSigned: r.IsSigned,
	}
}
