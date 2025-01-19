// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :execrows
INSERT INTO users (email) VALUES ($1)
`

func (q *Queries) CreateUser(ctx context.Context, email string) (int64, error) {
	result, err := q.db.ExecContext(ctx, createUser, email)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteUser = `-- name: DeleteUser :execrows
DELETE FROM users 
WHERE email = $1
`

func (q *Queries) DeleteUser(ctx context.Context, email string) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteUser, email)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email 
FROM users 
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(&i.ID, &i.Email)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, email 
FROM users
`

func (q *Queries) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(&i.ID, &i.Email); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUserEmail = `-- name: UpdateUserEmail :execrows
UPDATE users
SET email = $2
WHERE id = $1
`

type UpdateUserEmailParams struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateUserEmail, arg.ID, arg.Email)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
