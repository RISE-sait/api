-- name: CreatePendingUser :one
INSERT INTO users.pending_users (first_name, last_name, email, phone, parent_hubspot_id, age, has_sms_consent,
                                 has_marketing_email_consent, is_parent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: DeletePendingUser :execrows
DELETE FROM users.pending_users WHERE id = $1;

-- name: GetPendingUserByEmail :one
SELECT * FROM users.pending_users WHERE email = $1;