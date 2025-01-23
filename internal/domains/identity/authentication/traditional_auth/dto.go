package traditional_auth

import (
	db "api/internal/domains/identity/authentication/infra/sqlc/generated"
	"database/sql"
)

type GetUserRequest struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"notwhitespace"`
}

func (r *GetUserRequest) ToDBParams() *db.GetUserByEmailPasswordParams {

	return &db.GetUserByEmailPasswordParams{
		Email: r.Email,
		HashedPassword: sql.NullString{
			String: r.Password,
			Valid:  r.Password != "",
		},
	}
}
