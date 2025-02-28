-- name: CreatePendingUser :one
INSERT INTO users.pending_users (first_name, last_name, email, parent_hubspot_id, age)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeletePendingUser :execrows
DELETE FROM users.pending_users WHERE id = $1;

-- name: GetPendingUserByEmail :one
SELECT * FROM users.pending_users WHERE email = $1;