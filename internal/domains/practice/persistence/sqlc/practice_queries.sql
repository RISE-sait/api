-- name: CreatePractice :one
INSERT INTO practices (name, description, level, should_email_booking_notification, capacity,
                       start_date, end_date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
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
    start_date = $6,
    end_date = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $8;

-- name: DeletePractice :execrows
DELETE FROM practices WHERE id = $1;