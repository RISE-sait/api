-- name: CreateCourse :one
INSERT INTO course.courses (name, description, capacity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCourseById :one
SELECT *
FROM course.courses
WHERE id = $1;

-- name: GetCourses :many
SELECT *
FROM course.courses;

-- name: UpdateCourse :execrows
UPDATE course.courses
SET name        = $1,
    description = $2,
    capacity = $3,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $4;

-- name: DeleteCourse :execrows
DELETE
FROM course.courses
WHERE id = $1;