-- name: GetStaffByEmail :one
SELECT oi.name, u.email, s.is_active, s.created_at, s.updated_at, sr.role_name FROM staff s
JOIN users u ON s.id = u.id
JOIN user_optional_info oi ON oi.id = u.id
JOIN staff_roles sr ON s.role_id = sr.id
WHERE u.email = $1;

-- name: GetStaffRoles :many
SELECT * FROM staff_roles;

-- name: CreateStaff :execrows
INSERT INTO staff (id, role_id, is_active) VALUES ((SELECT id from users WHERE email = $1), 
(SELECT id from staff_roles where role_name = $2), $3);
