-- name: CreateCourse :one
INSERT INTO courses (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetCourseById :one
SELECT * FROM courses WHERE id = $1;

-- name: GetCourses :many
SELECT * FROM courses
WHERE (name ILIKE '%' || @name || '%' OR @name IS NULL)
AND (description ILIKE '%' || sqlc.narg('description') || '%' OR sqlc.narg('description') IS NULL);

-- name: UpdateCourse :execrows
UPDATE courses
SET name = $1, description = $2
WHERE id = $3;

-- name: DeleteCourse :execrows
DELETE FROM courses WHERE id = $1;