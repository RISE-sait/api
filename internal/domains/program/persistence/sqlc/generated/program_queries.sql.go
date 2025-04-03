// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: program_queries.sql

package db_program

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createProgram = `-- name: CreateProgram :exec
INSERT INTO program.programs (name, description, level, type, capacity)
VALUES ($1, $2, $3, $4, $5)
`

type CreateProgramParams struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Level       ProgramProgramLevel `json:"level"`
	Type        ProgramProgramType  `json:"type"`
	Capacity    sql.NullInt32       `json:"capacity"`
}

// Active: 1739459832645@@127.0.0.1@5432@postgres
func (q *Queries) CreateProgram(ctx context.Context, arg CreateProgramParams) error {
	_, err := q.db.ExecContext(ctx, createProgram,
		arg.Name,
		arg.Description,
		arg.Level,
		arg.Type,
		arg.Capacity,
	)
	return err
}

const deleteProgram = `-- name: DeleteProgram :execrows
DELETE FROM program.programs WHERE id = $1
`

func (q *Queries) DeleteProgram(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteProgram, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getProgram = `-- name: GetProgram :one
SELECT id, name, description, level, type, capacity, created_at, updated_at
FROM program.programs
WHERE id = $1
`

func (q *Queries) GetProgram(ctx context.Context, id uuid.UUID) (ProgramProgram, error) {
	row := q.db.QueryRowContext(ctx, getProgram, id)
	var i ProgramProgram
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.Type,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getProgramById = `-- name: GetProgramById :one
SELECT id, name, description, level, type, capacity, created_at, updated_at FROM program.programs WHERE id = $1
`

func (q *Queries) GetProgramById(ctx context.Context, id uuid.UUID) (ProgramProgram, error) {
	row := q.db.QueryRowContext(ctx, getProgramById, id)
	var i ProgramProgram
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.Type,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getPrograms = `-- name: GetPrograms :many
SELECT id, name, description, level, type, capacity, created_at, updated_at
FROM program.programs
WHERE type = $1
   OR $1 IS NULL
`

func (q *Queries) GetPrograms(ctx context.Context, type_ NullProgramProgramType) ([]ProgramProgram, error) {
	rows, err := q.db.QueryContext(ctx, getPrograms, type_)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ProgramProgram
	for rows.Next() {
		var i ProgramProgram
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Level,
			&i.Type,
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

const updateProgram = `-- name: UpdateProgram :exec
UPDATE program.programs
SET
    name = $1,
    description = $2,
    level = $3,
    type = $4,
    capacity = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $6
`

type UpdateProgramParams struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Level       ProgramProgramLevel `json:"level"`
	Type        ProgramProgramType  `json:"type"`
	Capacity    sql.NullInt32       `json:"capacity"`
	ID          uuid.UUID           `json:"id"`
}

func (q *Queries) UpdateProgram(ctx context.Context, arg UpdateProgramParams) error {
	_, err := q.db.ExecContext(ctx, updateProgram,
		arg.Name,
		arg.Description,
		arg.Level,
		arg.Type,
		arg.Capacity,
		arg.ID,
	)
	return err
}
