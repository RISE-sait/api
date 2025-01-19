-- name: CreateUserOptionalInfo :execrows
INSERT INTO user_optional_info (id, name, hashed_password)
VALUES ((SELECT id FROM users WHERE email = $1), $2, $3);

-- name: GetUserOptionalInfo :one
SELECT * FROM user_optional_info WHERE id = (SELECT id FROM users WHERE email = $1) and hashed_password = $2;

-- name: UpdateUsername :execrows
UPDATE user_optional_info
SET name = $1
WHERE id = (SELECT id FROM users WHERE email = $2);

-- name: UpdateUserPassword :execrows
UPDATE user_optional_info
SET hashed_password = $1
WHERE id = (SELECT id FROM users WHERE email = $2);

-- name: DeleteUserOptionalInfo :execrows
DELETE FROM user_optional_info WHERE id = (SELECT id FROM users WHERE email = $1);