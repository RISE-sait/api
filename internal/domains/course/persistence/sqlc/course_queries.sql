-- name: CreateCourse :one
INSERT INTO courses (name, description, capacity)
VALUES ($1, $2, $3)
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
    capacity    = $3,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $4;

-- name: DeleteCourse :execrows
DELETE
FROM courses
WHERE id = $1;