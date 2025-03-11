-- name: GetStaffs :many
SELECT s.is_active, u.*, sr.role_name
FROM users.staff s
JOIN users.users u ON u.id = s.id
JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE (sr.role_name = sqlc.narg('role') OR sqlc.narg('role') IS NULL);

-- name: GetStaffByID :one
SELECT u.*, s.is_active, sr.role_name
FROM users.staff s
         JOIN users.users u ON s.id = u.id
JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE s.id = $1;

-- name: UpdateStaff :execrows
UPDATE users.staff s
    SET
        role_id = (SELECT id from users.staff_roles sr WHERE sr.role_name = $1),
        is_active = $2
WHERE s.id = $3;

-- name: DeleteStaff :execrows
DELETE FROM users.staff WHERE id = $1;