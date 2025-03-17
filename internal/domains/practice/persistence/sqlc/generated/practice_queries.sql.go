// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: practice_queries.sql

package db_practice

import (
	"context"

	"github.com/google/uuid"
)

const createPractice = `-- name: CreatePractice :one
INSERT INTO practices (name, description, level, capacity)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, level, capacity, created_at, updated_at
`

type CreatePracticeParams struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Level       PracticeLevel `json:"level"`
	Capacity    int32         `json:"capacity"`
}

func (q *Queries) CreatePractice(ctx context.Context, arg CreatePracticeParams) (Practice, error) {
	row := q.db.QueryRowContext(ctx, createPractice,
		arg.Name,
		arg.Description,
		arg.Level,
		arg.Capacity,
	)
	var i Practice
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deletePractice = `-- name: DeletePractice :execrows
DELETE FROM practices WHERE id = $1
`

func (q *Queries) DeletePractice(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deletePractice, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getPracticeById = `-- name: GetPracticeById :one
SELECT id, name, description, level, capacity, created_at, updated_at FROM practices WHERE id = $1
`

func (q *Queries) GetPracticeById(ctx context.Context, id uuid.UUID) (Practice, error) {
	row := q.db.QueryRowContext(ctx, getPracticeById, id)
	var i Practice
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPracticeByName = `-- name: GetPracticeByName :one
SELECT id, name, description, level, capacity, created_at, updated_at FROM practices WHERE name = $1 LIMIT 1
`

func (q *Queries) GetPracticeByName(ctx context.Context, name string) (Practice, error) {
	row := q.db.QueryRowContext(ctx, getPracticeByName, name)
	var i Practice
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPractices = `-- name: GetPractices :many
SELECT id, name, description, level, capacity, created_at, updated_at FROM practices
`

func (q *Queries) GetPractices(ctx context.Context) ([]Practice, error) {
	rows, err := q.db.QueryContext(ctx, getPractices)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Practice
	for rows.Next() {
		var i Practice
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Level,
			&i.Capacity,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
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

const updatePractice = `-- name: UpdatePractice :execrows
UPDATE practices
SET
    name = $1,
    description = $2,
    level = $3,
    capacity = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $5
`

type UpdatePracticeParams struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Level       PracticeLevel `json:"level"`
	Capacity    int32         `json:"capacity"`
	ID          uuid.UUID     `json:"id"`
}

func (q *Queries) UpdatePractice(ctx context.Context, arg UpdatePracticeParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updatePractice,
		arg.Name,
		arg.Description,
		arg.Level,
		arg.Capacity,
		arg.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
