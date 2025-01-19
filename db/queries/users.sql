-- name: CreateUser :execrows
INSERT INTO users (email) VALUES ($1); 

-- name: GetUserByEmail :one
SELECT id, email 
FROM users 
WHERE email = $1;

-- name: ListUsers :many
SELECT id, email 
FROM users;

-- name: DeleteUser :execrows
DELETE FROM users 
WHERE email = $1;

-- name: UpdateUserEmail :execrows
UPDATE users
SET email = $2
WHERE id = $1;