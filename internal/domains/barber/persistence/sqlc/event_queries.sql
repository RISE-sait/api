-- name: CreateBarberEvent :one
INSERT INTO barber.barber_events (begin_date_time, end_date_time, barber_id, customer_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBarberEvents :many
SELECT *
FROM barber.barber_events
WHERE
    (barber_id = sqlc.narg('barber_id') OR sqlc.narg('barber_id') IS NULL) -- Filter by barber_id
  AND (customer_id = sqlc.narg('customer_id') OR sqlc.narg('customer_id') IS NULL);


-- name: GetEventById :one
SELECT *
FROM barber.barber_events
WHERE id = $1;

-- name: UpdateEvent :one
UPDATE barber.barber_events
SET
    begin_date_time = $1,
    end_date_time = $2,
    barber_id = $3,
    customer_id = $4
WHERE id = $5
RETURNING *;

-- name: DeleteEvent :execrows
DELETE FROM barber.barber_events
WHERE id = $1;