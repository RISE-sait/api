package userOptionalInfo

import (
	db "api/sqlc"
	"database/sql"
)

type GetUserOptionalInfoRequest struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"notwhitespace"`
}

func (r *GetUserOptionalInfoRequest) ToDBParams() *db.GetUserOptionalInfoParams {

	return &db.GetUserOptionalInfoParams{
		Email: r.Email,
		HashedPassword: sql.NullString{
			String: r.Password,
			Valid:  r.Password != "",
		},
	}
}

type CreateUserOptionalInfoRequest struct {
	Name           string `json:"name" validate:"omitempty,notwhitespace"`
	Email          string `json:"email" validate:"email"`
	HashedPassword string `json:"password" validate:"omitempty,notwhitespace"`
}

func (r *CreateUserOptionalInfoRequest) ToDBParams() *db.CreateUserOptionalInfoParams {

	return &db.CreateUserOptionalInfoParams{
		Name: sql.NullString{
			String: r.Name,
			Valid:  r.Name != "",
		},
		Email: r.Email,
		HashedPassword: sql.NullString{
			String: r.HashedPassword,
			Valid:  r.HashedPassword != "",
		},
	}
}

type UpdateUserNameRequest struct {
	Name  string `json:"name" validate:"notwhitespace"`
	Email string `json:"email" validate:"required,email"`
}

func (r *UpdateUserNameRequest) ToDBParams() *db.UpdateUsernameParams {

	return &db.UpdateUsernameParams{
		Name: sql.NullString{
			String: r.Name,
			Valid:  r.Name != "",
		},
		Email: r.Email,
	}
}

type UpdateUserPasswordRequest struct {
	Email          string `json:"email" validate:"required,email"`
	HashedPassword string `json:"hashed_password" validate:"omitempty,notwhitespace"`
}

func (r *UpdateUserPasswordRequest) ToDBParams() *db.UpdateUserPasswordParams {

	return &db.UpdateUserPasswordParams{
		Email: r.Email,
		HashedPassword: sql.NullString{
			String: r.HashedPassword,
			Valid:  r.HashedPassword != "",
		},
	}
}
