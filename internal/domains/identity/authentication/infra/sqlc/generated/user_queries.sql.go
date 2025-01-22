// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user_queries.sql

package db

import (
	"context"
	"database/sql"
)

const getUserByEmailPassword = `-- name: GetUserByEmailPassword :one
SELECT id, name, hashed_password FROM user_optional_info WHERE id = (SELECT id FROM users WHERE email = $1) and hashed_password = $2
`

type GetUserByEmailPasswordParams struct {
	Email          string         `json:"email"`
	HashedPassword sql.NullString `json:"hashed_password"`
}

func (q *Queries) GetUserByEmailPassword(ctx context.Context, arg GetUserByEmailPasswordParams) (UserOptionalInfo, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmailPassword, arg.Email, arg.HashedPassword)
	var i UserOptionalInfo
	err := row.Scan(&i.ID, &i.Name, &i.HashedPassword)
	return i, err
}
