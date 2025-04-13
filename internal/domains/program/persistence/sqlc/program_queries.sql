-- Active: 1739459832645@@127.0.0.1@5432@postgres
-- name: CreateProgram :one
INSERT INTO program.programs (name, description, level, type, capacity)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPrograms :many
SELECT *
FROM program.programs
WHERE type = sqlc.narg('type')
   OR sqlc.narg('type') IS NULL;

-- name: GetProgramById :one
SELECT * FROM program.programs WHERE id = $1;

-- name: UpdateProgram :one
UPDATE program.programs
SET
    name = $1,
    description = $2,
    level = $3,
    type = $4,
    capacity = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $6
RETURNING *;

-- name: DeleteProgram :execrows
DELETE FROM program.programs WHERE id = $1;