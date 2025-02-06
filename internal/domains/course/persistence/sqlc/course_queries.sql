-- name: CreateCourse :execrows
INSERT INTO courses (name, description, start_date, end_date)
VALUES ($1, $2, $3, $4);

-- name: GetCourseById :one
SELECT * FROM courses WHERE id = $1;

-- name: GetCourses :many
SELECT * FROM courses
WHERE (name ILIKE '%' || @name || '%' OR @name IS NULL)
AND (description ILIKE '%' || @description|| '%' OR @description IS NULL);

-- name: UpdateCourse :execrows
UPDATE courses
SET name = $1, description = $2, start_date = $3, end_date = $4
WHERE id = $5;

-- name: DeleteCourse :execrows
DELETE FROM courses WHERE id = $1;