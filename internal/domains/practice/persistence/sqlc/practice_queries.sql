-- name: CreatePractice :one
INSERT INTO practices (name, description, level, should_email_booking_notification, capacity)
VALUES ($1, $2, $3, $4, $5)
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
    should_email_booking_notification = $4,
    capacity = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $6;

-- name: DeletePractice :execrows
DELETE FROM practices WHERE id = $1;