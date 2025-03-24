-- name: EnrollCustomer :one
INSERT INTO events.customer_enrollment (customer_id, event_id, checked_in_at, is_cancelled)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCustomerEnrollments :many
SELECT customer_enrollment.*
FROM events.customer_enrollment
         JOIN users.users ON customer_enrollment.customer_id = users.id
WHERE (
          (customer_id = sqlc.narg('customer_id') OR sqlc.narg('customer_id') IS NULL)
              AND
          (event_id = sqlc.narg('event_id') OR sqlc.narg('event_id') IS NULL));

-- name: UnEnrollCustomer :execrows
DELETE
FROM events.customer_enrollment
WHERE id = $1;

-- name: GetEventIsFull :one
SELECT 
    (CASE 
        WHEN COALESCE(e.capacity, t.capacity) IS NULL THEN false
        ELSE COUNT(ce.customer_id) >= COALESCE(e.capacity, t.capacity)
    END)::boolean AS is_full
FROM events.events e 
LEFT JOIN athletic.teams t ON e.team_id = t.id
LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
WHERE e.id = @event_id
GROUP BY e.id, e.capacity, t.capacity;