// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: course_queries.sql

package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createCourse = `-- name: CreateCourse :one
INSERT INTO course.courses (name, description, capacity)
VALUES ($1, $2, $3)
RETURNING id, name, description, capacity, created_at, updated_at
`

type CreateCourseParams struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Capacity    int32          `json:"capacity"`
}

func (q *Queries) CreateCourse(ctx context.Context, arg CreateCourseParams) (CourseCourse, error) {
	row := q.db.QueryRowContext(ctx, createCourse, arg.Name, arg.Description, arg.Capacity)
	var i CourseCourse
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteCourse = `-- name: DeleteCourse :execrows
DELETE FROM course.courses WHERE id = $1
`

func (q *Queries) DeleteCourse(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteCourse, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getCourseById = `-- name: GetCourseById :one
SELECT id, name, description, capacity, created_at, updated_at FROM course.courses WHERE id = $1
`

func (q *Queries) GetCourseById(ctx context.Context, id uuid.UUID) (CourseCourse, error) {
	row := q.db.QueryRowContext(ctx, getCourseById, id)
	var i CourseCourse
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCourses = `-- name: GetCourses :many
SELECT id, name, description, capacity, created_at, updated_at FROM course.courses
`

func (q *Queries) GetCourses(ctx context.Context) ([]CourseCourse, error) {
	rows, err := q.db.QueryContext(ctx, getCourses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CourseCourse
	for rows.Next() {
		var i CourseCourse
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
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

const updateCourse = `-- name: UpdateCourse :execrows
UPDATE course.courses
SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3
`

type UpdateCourseParams struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	ID          uuid.UUID      `json:"id"`
}

func (q *Queries) UpdateCourse(ctx context.Context, arg UpdateCourseParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateCourse, arg.Name, arg.Description, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
