-- name: EnrollCustomer :one
INSERT INTO events.customer_enrollment (customer_id, event_id, checked_in_at, is_cancelled)
VALUES ($1, $2, $3, false)
RETURNING *;

-- name: UnEnrollCustomer :execrows
UPDATE events.customer_enrollment
SET is_cancelled = true
WHERE customer_id = $1
  AND event_id = $2;

-- name: GetEventIsFull :one
SELECT COUNT(ce.customer_id) >= COALESCE(e.capacity, p.capacity, t.capacity)::boolean AS is_full
FROM events.events e
         JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN athletic.teams t ON e.team_id = t.id
LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
WHERE e.id = @event_id
GROUP BY e.id, e.capacity, p.capacity, t.capacity;