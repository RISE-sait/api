-- name: AssignStaffToEvent :execrows
INSERT INTO events.staff (event_id, staff_id)
VALUES ($1, $2);

-- name: GetStaffsAssignedToEvent :many
SELECT us.*, sr.role_name, u.hubspot_id
FROM users.staff us
         JOIN users.staff_roles sr ON us.role_id = sr.id
         JOIN users.users u ON u.id = us.id
         JOIN events.staff s ON us.id = s.staff_id
WHERE event_id = $1;

-- name: UnassignStaffFromEvent :execrows
DELETE
FROM events.staff
where staff_id = $1
and event_id = $2;