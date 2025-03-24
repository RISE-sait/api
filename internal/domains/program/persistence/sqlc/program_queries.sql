-- Active: 1739459832645@@127.0.0.1@5432@postgres
-- name: CreateProgram :exec
INSERT INTO program.programs (name, description, level, type)
VALUES ($1, $2, $3, $4);

-- name: GetPrograms :many
SELECT * FROM program.programs;

-- name: GetProgramById :one
SELECT * FROM program.programs WHERE id = $1;

-- name: UpdateProgram :exec
UPDATE program.programs
SET
    name = $1,
    description = $2,
    level = $3,
    type = $4,

    updated_at = CURRENT_TIMESTAMP
WHERE id = $5;

-- name: DeleteProgram :execrows
DELETE FROM program.programs WHERE id = $1;