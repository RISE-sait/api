-- name: CreateCourse :one
INSERT INTO courses (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetCourseById :one
SELECT *
FROM courses
WHERE id = $1;

-- name: GetCourses :many
SELECT *
FROM courses;

-- name: UpdateCourse :execrows
UPDATE courses
SET name        = $1,
    description = $2,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $3;

-- name: DeleteCourse :execrows
DELETE
FROM courses
WHERE id = $1;