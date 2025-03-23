-- name: AssignStaffToEvent :execrows
INSERT INTO events.staff (event_id, staff_id)
VALUES ($1, $2);

-- name: GetStaffsAssignedToEvent :many
SELECT s.*, sr.role_name, u.hubspot_id
FROM staff.staff s
         JOIN staff.staff_roles sr ON s.role_id = sr.id
    JOIN users.users u ON u.id = s.id
         JOIN events.staff es ON s.id = es.staff_id
WHERE event_id = $1;

-- name: UnassignStaffFromEvent :execrows
DELETE
FROM events.staff
where staff_id = $1
and event_id = $2;