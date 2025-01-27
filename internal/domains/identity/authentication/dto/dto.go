package dto

import (
	db "api/internal/domains/identity/infra/persistence/sqlc/generated"
	"database/sql"
)

type GetUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
