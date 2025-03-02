-- name: GetStaffs :many
SELECT s.*, sr.role_name FROM users.staff s
JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE
(role_id = sqlc.narg('role_id') OR sqlc.narg('role_id') IS NULL);

-- name: GetStaffByID :one
SELECT *, sr.role_name FROM users.staff s
JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE s.id = $1;

-- name: UpdateStaff :one
WITH updated_staff AS (
    UPDATE users.staff s
    SET
        role_id = $1,
        is_active = $2
    WHERE s.id = $3
    RETURNING *
)
SELECT us.*, sr.role_name
FROM updated_staff us
JOIN users.staff_roles sr ON us.role_id = sr.id;

-- name: DeleteStaff :execrows
DELETE FROM users.staff WHERE id = $1;