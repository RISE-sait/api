-- name: EnrollCustomer :one
INSERT INTO customer_enrollment (customer_id, event_id, checked_in_at, is_cancelled)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCustomerEnrollments :many
SELECT customer_enrollment.*
FROM customer_enrollment
         JOIN users.users ON customer_enrollment.customer_id = users.id
         WHERE (
                   (customer_id = sqlc.narg('customer_id') OR sqlc.narg('customer_id') IS NULL)
                       AND
                    (event_id = sqlc.narg('event_id') OR sqlc.narg('event_id') IS NULL));

-- name: UnEnrollCustomer :execrows
DELETE
FROM customer_enrollment
WHERE id = $1;

-- name: GetEventIsFull :one
SELECT COUNT(ce.customer_id) >= COALESCE(p.capacity, c.capacity) AS is_full
FROM events e
         LEFT JOIN customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN practices p ON e.practice_id = p.id
         LEFT JOIN course.courses c ON e.course_id = c.id
WHERE e.id = @event_id
GROUP BY e.id, e.practice_id, e.course_id, p.capacity, c.capacity;