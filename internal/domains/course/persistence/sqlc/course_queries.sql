-- name: CreateCourse :one
INSERT INTO courses (name, description, capacity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCourseById :one
SELECT * FROM courses WHERE id = $1;

-- name: GetCourses :many
SELECT * FROM courses
WHERE (name ILIKE '%' || @name || '%' OR @name IS NULL)
AND (description ILIKE '%' || sqlc.narg('description') || '%' OR sqlc.narg('description') IS NULL);

-- name: UpdateCourse :one
UPDATE courses
SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $3
RETURNING *;

-- name: DeleteCourse :execrows
DELETE FROM courses WHERE id = $1;