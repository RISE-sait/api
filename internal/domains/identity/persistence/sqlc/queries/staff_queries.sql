-- name: GetStaffById :one
SELECT s.*, sr.role_name, u.hubspot_id
FROM users.staff s
         JOIN users.users u ON s.id = u.id
         JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE u.id = $1;

-- name: GetStaffRoles :many
SELECT *
FROM users.staff_roles;

-- name: CreateApprovedStaff :execrows
INSERT INTO users.staff (id, role_id, is_active)
VALUES ($1,
        (SELECT id from users.staff_roles where role_name = $2), $3);