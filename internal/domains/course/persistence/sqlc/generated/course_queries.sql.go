// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: course_queries.sql

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createCourse = `-- name: CreateCourse :execrows
INSERT INTO courses (name, description, start_date, end_date)
VALUES ($1, $2, $3, $4)
`

type CreateCourseParams struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
}

func (q *Queries) CreateCourse(ctx context.Context, arg CreateCourseParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createCourse,
		arg.Name,
		arg.Description,
		arg.StartDate,
		arg.EndDate,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteCourse = `-- name: DeleteCourse :execrows
DELETE FROM courses WHERE id = $1
`

func (q *Queries) DeleteCourse(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteCourse, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getCourseById = `-- name: GetCourseById :one
SELECT id, name, description, start_date, end_date, created_at, updated_at FROM courses WHERE id = $1
`

func (q *Queries) GetCourseById(ctx context.Context, id uuid.UUID) (Course, error) {
	row := q.db.QueryRowContext(ctx, getCourseById, id)
	var i Course
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.StartDate,
		&i.EndDate,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourses = `-- name: GetCourses :many
SELECT id, name, description, start_date, end_date, created_at, updated_at FROM courses
WHERE (name ILIKE '%' || $1 || '%' OR $1 IS NULL)
AND (description ILIKE '%' || $2|| '%' OR $2 IS NULL)
`

type GetCoursesParams struct {
	Name        sql.NullString `json:"name"`
	Description sql.NullString `json:"description"`
}

func (q *Queries) GetCourses(ctx context.Context, arg GetCoursesParams) ([]Course, error) {
	rows, err := q.db.QueryContext(ctx, getCourses, arg.Name, arg.Description)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Course
	for rows.Next() {
		var i Course
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.StartDate,
			&i.EndDate,
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

const updateCourse = `-- name: UpdateCourse :execrows
UPDATE courses
SET name = $1, description = $2, start_date = $3, end_date = $4
WHERE id = $5
`

type UpdateCourseParams struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	ID          uuid.UUID      `json:"id"`
}

func (q *Queries) UpdateCourse(ctx context.Context, arg UpdateCourseParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateCourse,
		arg.Name,
		arg.Description,
		arg.StartDate,
		arg.EndDate,
		arg.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
