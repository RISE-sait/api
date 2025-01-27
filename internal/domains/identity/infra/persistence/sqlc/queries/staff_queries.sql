-- name: GetStaffByEmail :one
SELECT oi.name, u.email, s.is_active, s.created_at, s.updated_at, s.role FROM staff s
JOIN users u ON s.id = u.id
JOIN user_optional_info oi ON oi.id = u.id WHERE u.email = $1;

-- name: CreateStaff :execrows
INSERT INTO staff (id, role, is_active) VALUES ((SELECT id from users WHERE email = $1), $2, $3);