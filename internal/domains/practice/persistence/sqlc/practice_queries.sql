-- name: CreatePractice :exec
INSERT INTO practices (name, description, level, payg_price)
VALUES ($1, $2, $3, $4);

-- name: GetPractices :many
SELECT * FROM practices;

-- name: GetPracticeById :one
SELECT * FROM practices WHERE id = $1;

-- name: UpdatePractice :exec
UPDATE practices
SET
    name = $1,
    description = $2,
    level = $3,
    payg_price = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $5;

-- name: DeletePractice :execrows
DELETE FROM practices WHERE id = $1;