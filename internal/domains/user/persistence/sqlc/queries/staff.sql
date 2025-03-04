-- name: GetStaffs :many
SELECT s.*, u.hubspot_id, sr.role_name FROM users.staff s
JOIN users.users u ON u.id = s.id
JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE
(sr.role_name = sqlc.narg('role') OR sqlc.narg('role') IS NULL)
    AND
    (hubspot_id = ANY(sqlc.narg('hubspot_ids')::text[]) OR sqlc.narg('hubspot_ids') IS NULL);

-- name: GetStaffByID :one
SELECT *, sr.role_name FROM users.staff s
JOIN users.staff_roles sr ON s.role_id = sr.id
WHERE s.id = $1;

-- name: UpdateStaff :one
WITH updated_staff AS (
    UPDATE users.staff s
    SET
        role_id = (SELECT id from users.staff_roles sr WHERE sr.role_name = $1),
        is_active = $2
    WHERE s.id = $3
    RETURNING *
)
SELECT us.*, sr.role_name
FROM updated_staff us
JOIN users.staff_roles sr ON us.role_id = sr.id;

-- name: DeleteStaff :execrows
DELETE FROM users.staff WHERE id = $1;