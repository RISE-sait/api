-- name: GetUserByEmailPassword :one
SELECT * FROM user_optional_info WHERE id = (SELECT id FROM users WHERE email = $1) and hashed_password = $2;

-- name: CreateUserOptionalInfo :execrows
INSERT INTO user_optional_info (id, first_name, last_name, phone, hashed_password) 
VALUES (
    (SELECT id FROM users WHERE email = $1),
    $2,
    $3,
    $4,
    $5
    );
