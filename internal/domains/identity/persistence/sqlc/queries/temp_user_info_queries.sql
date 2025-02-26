-- name: CreateTempUserInfo :one
INSERT INTO temp_users_info (id, first_name, last_name, email, parent_hubspot_id, age)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: DeleteTempUserInfo :execrows
DELETE FROM temp_users_info WHERE id = $1;

-- name: GetTempUserInfoByEmail :one
SELECT * FROM temp_users_info WHERE email = $1;