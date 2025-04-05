-- name: EnrollCustomerInEvent :exec
INSERT INTO events.customer_enrollment (customer_id, event_id)
VALUES ($1, $2);