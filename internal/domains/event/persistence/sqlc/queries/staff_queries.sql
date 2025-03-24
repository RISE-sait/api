-- name: AssignStaffToEvent :execrows
INSERT INTO events.staff (event_id, staff_id)
VALUES ($1, $2);

-- name: UnassignStaffFromEvent :execrows
DELETE
FROM events.staff
where staff_id = $1
and event_id = $2;