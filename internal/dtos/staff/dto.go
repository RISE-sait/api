package staff

import (
	db "api/sqlc"
)

type CreateStaffRequest struct {
	Email    string `json:"email" validate:"email"`
	Role     string `json:"role" validate:"notwhitespace"`
	IsActive bool   `json:"is_active" validate:"notwhitespace"`
}

func (r *CreateStaffRequest) ToCreateStaffParams() *db.CreateStaffParams {

	return &db.CreateStaffParams{

		Email:    r.Email,
		Role:     db.StaffRoleEnum(r.Role),
		IsActive: r.IsActive,
	}
}

type UpdateStaffRequest struct {
	IsActive bool   `json:"is_active" validate:"notwhitespace"`
	Role     string `json:"role" validate:"notwhitespace"`
	Email    string `json:"email" validate:"email"`
}

func (r *UpdateStaffRequest) ToDBParams() *db.UpdateStaffParams {

	return &db.UpdateStaffParams{

		Role:     db.StaffRoleEnum(r.Role),
		IsActive: r.IsActive,
		Email:    r.Email,
	}
}
