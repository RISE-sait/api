-- name: CreateBarberEvent :one
INSERT INTO barber.barber_events (begin_date_time, end_date_time, barber_id, customer_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBarberEvents :many
SELECT be.*,
       (barbers.first_name || ' ' || barbers.last_name)::text     as barber_name,
       (customers.first_name || ' ' || customers.last_name)::text as customer_name
FROM barber.barber_events be
         JOIN users.users barbers
              ON barbers.id = be.barber_id
         JOIN users.users customers
              ON customers.id = be.customer_id
WHERE
    (barber_id = sqlc.narg('barber_id') OR sqlc.narg('barber_id') IS NULL) -- Filter by barber_id
  AND (customer_id = sqlc.narg('customer_id') OR sqlc.narg('customer_id') IS NULL)
  AND (sqlc.narg('before') >= begin_date_time OR sqlc.narg('before') IS NULL) -- within boundary
  AND (sqlc.narg('after') <= end_date_time OR sqlc.narg('after') IS NULL);


-- name: GetEventById :one
SELECT *,
       (barbers.first_name || ' ' || barbers.last_name)::text     as barber_name,
       (customers.first_name || ' ' || customers.last_name)::text as customer_name
FROM barber.barber_events e
         JOIN users.users barbers
              ON barbers.id = barber_id

         JOIN users.users customers
              ON customers.id = customer_id
WHERE e.id = $1;

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