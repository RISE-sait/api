-- name: EnrollCustomer :one
INSERT INTO customer_enrollment (customer_id, event_id, checked_in_at, is_cancelled)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCustomerEnrollments :many
SELECT customer_enrollment.*, users.email FROM customer_enrollment
         JOIN users.users ON customer_enrollment.customer_id = users.id
         WHERE (customer_id = COALESCE(sqlc.narg('customer_id'), customer_id)
         OR
                event_id = COALESCE(sqlc.narg('event_id'), event_id)
                   )
;

-- name: UnEnrollCustomer :execrows
DELETE FROM customer_enrollment WHERE id = $1;