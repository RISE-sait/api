-- name: GetStaffById :one
SELECT s.is_active, s.created_at, s.updated_at, sr.role_name FROM staff s
JOIN users u ON s.id = u.id
JOIN staff_roles sr ON s.role_id = sr.id
WHERE u.id = $1;

-- name: GetStaffRoles :many
SELECT * FROM staff_roles;

-- name: CreateStaff :execrows
INSERT INTO staff (id, role_id, is_active) VALUES ($1,
(SELECT id from staff_roles where role_name = $2), $3);