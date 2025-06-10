-- name: CreateSession :one
WITH inserted AS (
    INSERT INTO playground.sessions (system_id, customer_id, start_time, end_time)
    VALUES ($1, $2, $3, $4)
    RETURNING id, system_id, customer_id, start_time, end_time, created_at, updated_at
)
SELECT i.id,
       i.system_id,
       sys.name  AS system_name,
       i.customer_id,
       u.first_name AS customer_first_name,
       u.last_name  AS customer_last_name,
       i.start_time,
       i.end_time,
       i.created_at,
       i.updated_at
FROM inserted i
         JOIN playground.systems sys ON sys.id = i.system_id
         JOIN users.users u ON u.id = i.customer_id;

-- name: GetSessions :many
SELECT s.id,
       s.system_id,
       sys.name  AS system_name,
       s.customer_id,
       u.first_name AS customer_first_name,
       u.last_name  AS customer_last_name,
       s.start_time,
       s.end_time,
       s.created_at,
       s.updated_at
FROM playground.sessions s
         JOIN playground.systems sys ON sys.id = s.system_id
         JOIN users.users u ON u.id = s.customer_id
ORDER BY s.start_time;

-- name: GetSession :one
SELECT s.id,
       s.system_id,
       sys.name  AS system_name,
       s.customer_id,
       u.first_name AS customer_first_name,
       u.last_name  AS customer_last_name,
       s.start_time,
       s.end_time,
       s.created_at,
       s.updated_at
FROM playground.sessions s
         JOIN playground.systems sys ON sys.id = s.system_id
         JOIN users.users u ON u.id = s.customer_id
WHERE s.id = $1;

-- name: DeleteSession :execrows
DELETE FROM playground.sessions WHERE id = $1;