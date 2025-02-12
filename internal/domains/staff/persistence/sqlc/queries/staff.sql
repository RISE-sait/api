-- name: GetStaffs :many
SELECT s.*, sr.role_name FROM staff s
JOIN staff_roles sr ON s.role_id = sr.id
WHERE
(role_id = sqlc.narg('role_id') OR sqlc.narg('role_id') IS NULL);

-- name: GetStaffByID :one
SELECT *, sr.role_name FROM staff s 
JOIN staff_roles sr ON staff.role_id = staff_roles.id
WHERE s.id = $1;

-- name: UpdateStaff :one
WITH updated_staff AS (
    UPDATE staff s
    SET
        role_id = $1,
        is_active = $2
    WHERE s.id = $3
    RETURNING *
)
SELECT us.*, sr.role_name
FROM updated_staff us
JOIN staff_roles sr ON us.role_id = sr.id;

-- name: DeleteStaff :execrows
DELETE FROM staff WHERE id = $1;