-- name: CreatePractice :one
INSERT INTO practices (name, description, level, capacity)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPractices :many
SELECT * FROM practices;

-- name: GetPracticeById :one
SELECT * FROM practices WHERE id = $1;

-- name: GetPracticeByName :one
SELECT * FROM practices WHERE name = $1 LIMIT 1;

-- name: UpdatePractice :execrows
UPDATE practices
SET
    name = $1,
    description = $2,
    level = $3,
    capacity = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $5;

-- name: DeletePractice :execrows
DELETE FROM practices WHERE id = $1;