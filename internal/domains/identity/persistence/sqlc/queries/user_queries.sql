-- name: CreateUser :execrows
INSERT INTO users (email) VALUES ($1);

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;