-- name: GetUserByEmailPassword :one
SELECT * FROM user_optional_info WHERE id = (SELECT id FROM users WHERE email = $1) and hashed_password = $2;

-- name: CreatePassword :execrows
INSERT INTO user_optional_info (id, hashed_password) VALUES ((SELECT id FROM users WHERE email = $1), $2);
