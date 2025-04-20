-- name: InsertStaffActivity :exec
INSERT INTO audit.staff_activity_logs (staff_id, activity_description)
VALUES ($1, $2);

-- name: GetStaffActivityLogs :many
SELECT sal.id,
       sal.staff_id,
       sal.activity_description,
       sal.created_at,
       u.first_name,
       u.last_name,
       u.email
FROM audit.staff_activity_logs sal
         JOIN staff.staff s ON sal.staff_id = s.id
         JOIN users.users u ON s.id = u.id
WHERE (sqlc.narg('staff_id')::uuid IS NULL OR sal.staff_id = sqlc.narg('staff_id')::uuid)
  AND (
    sqlc.narg('search_description')::text IS NULL
        OR sal.activity_description ILIKE '%' || sqlc.narg('search_description')::text || '%'
    )
ORDER BY sal.created_at DESC
LIMIT $1 OFFSET $2;