-- name: GetStaffs :many
SELECT s.is_active, u.*, sr.role_name, cs.wins, cs.losses, s.photo_url
FROM staff.staff s
JOIN users.users u ON u.id = s.id
JOIN staff.staff_roles sr ON s.role_id = sr.id
LEFT JOIN athletic.coach_stats cs ON s.id = cs.coach_id
WHERE (sr.role_name = sqlc.narg('role_name') OR sqlc.narg('role_name') IS NULL);

-- name: UpdateCoachStats :execrows
UPDATE athletic.coach_stats
SET wins       = COALESCE(sqlc.narg('wins'), wins),
    losses     = COALESCE(sqlc.narg('losses'), losses),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id');

-- name: UpdateStaff :execrows
UPDATE staff.staff s
    SET role_id = (SELECT id from staff.staff_roles sr WHERE sr.role_name = $1),
        is_active  = $2,
        updated_at = CURRENT_TIMESTAMP
WHERE s.id = $3;

-- name: UpdateStaffProfile :execrows
UPDATE staff.staff
SET photo_url = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteStaff :execrows
DELETE
FROM staff.staff
WHERE id = $1;

-- name: CreateStaffRole :one
INSERT INTO staff.staff_roles (role_name)
VALUES ($1)
RETURNING *;

-- name: GetAvailableStaffRoles :many
SELECT *
FROM staff.staff_roles;