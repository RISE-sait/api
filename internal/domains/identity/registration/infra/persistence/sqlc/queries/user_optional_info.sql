-- name: CreatePassword :execrows
INSERT INTO user_optional_info (id, hashed_password) VALUES ((SELECT id FROM users WHERE email = $1), $2);
