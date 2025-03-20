-- name: AssignStaffToEvent :execrows
INSERT INTO event_staff (event_id, staff_id)
VALUES ($1, $2);

-- name: GetStaffsAssignedToEvent :many
SELECT s.*, sr.role_name, u.hubspot_id
FROM users.staff s
    JOIN users.staff_roles sr ON s.role_id = sr.id
    JOIN users.users u ON u.id = s.id
JOIN event_staff ON s.id = event_staff.staff_id
WHERE event_id = $1;

-- name: UnassignStaffFromEvent :execrows
DELETE FROM event_staff where staff_id = $1
and event_id = $2;