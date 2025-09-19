-- name: CreateHaircutEvent :one
INSERT INTO haircut.events (begin_date_time, end_date_time, barber_id, customer_id, service_type_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *,
    (SELECT first_name || ' ' || last_name FROM users.users WHERE id = customer_id)::varchar as customer_name,
    (SELECT first_name || ' ' || last_name FROM users.users WHERE id = barber_id)::varchar as barber_name;

-- name: GetHaircutEvents :many
SELECT e.*,
       (barbers.first_name || ' ' || barbers.last_name)::text     as barber_name,
       (customers.first_name || ' ' || customers.last_name)::text as customer_name
FROM haircut.events e
         JOIN users.users barbers
              ON barbers.id = e.barber_id
         JOIN users.users customers
              ON customers.id = e.customer_id
WHERE
    (barber_id = sqlc.narg('barber_id') OR sqlc.narg('barber_id') IS NULL) -- Filter by barber_id
  AND (customer_id = sqlc.narg('customer_id') OR sqlc.narg('customer_id') IS NULL)
  AND (sqlc.narg('before') >= begin_date_time OR sqlc.narg('before') IS NULL) -- within boundary
  AND (sqlc.narg('after') <= end_date_time OR sqlc.narg('after') IS NULL);


-- name: GetEventById :one
SELECT *,
       (barbers.first_name || ' ' || barbers.last_name)::text     as barber_name,
       (customers.first_name || ' ' || customers.last_name)::text as customer_name
FROM haircut.events e
         JOIN users.users barbers
              ON barbers.id = barber_id

         JOIN users.users customers
              ON customers.id = customer_id
WHERE e.id = $1;

-- name: UpdateEvent :one
UPDATE haircut.events
SET
    begin_date_time = $1,
    end_date_time = $2,
    barber_id = $3,
    customer_id = $4,
    updated_at  = current_timestamp
WHERE id = $5
RETURNING *;

-- name: DeleteEvent :execrows
DELETE
FROM haircut.events
WHERE id = $1;

-- name: GetBarberAvailability :many
SELECT day_of_week, start_time, end_time
FROM haircut.barber_availability
WHERE barber_id = $1 AND is_active = true
ORDER BY day_of_week, start_time;

-- name: GetBarberBookingsForDate :many
SELECT begin_date_time, end_date_time
FROM haircut.events
WHERE barber_id = $1 
  AND DATE(begin_date_time) = $2
ORDER BY begin_date_time;

-- name: CreateBarberAvailability :one
INSERT INTO haircut.barber_availability (barber_id, day_of_week, start_time, end_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBarberWorkingHoursForDay :many
SELECT start_time, end_time
FROM haircut.barber_availability
WHERE barber_id = $1 
  AND day_of_week = $2 
  AND is_active = true
ORDER BY start_time;

-- name: GetBarberAvailabilityByID :one
SELECT *
FROM haircut.barber_availability
WHERE id = $1;

-- name: GetBarberFullAvailability :many
SELECT *
FROM haircut.barber_availability
WHERE barber_id = $1
ORDER BY day_of_week, start_time;

-- name: UpsertBarberAvailability :one
INSERT INTO haircut.barber_availability (barber_id, day_of_week, start_time, end_time, is_active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (barber_id, day_of_week, start_time, end_time)
DO UPDATE SET 
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: UpdateBarberAvailability :one
UPDATE haircut.barber_availability
SET start_time = $2,
    end_time = $3,
    is_active = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteBarberAvailability :execrows
DELETE FROM haircut.barber_availability
WHERE id = $1;

-- name: DeleteBarberAvailabilityByDay :execrows
DELETE FROM haircut.barber_availability
WHERE barber_id = $1 AND day_of_week = $2;

-- name: InsertBarberAvailability :one
INSERT INTO haircut.barber_availability (barber_id, day_of_week, start_time, end_time, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;