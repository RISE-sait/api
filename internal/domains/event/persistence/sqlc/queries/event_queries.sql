-- name: CreateEvent :exec
INSERT INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, day, location_id,
                           program_id, capacity, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5,
        $6, $7, $8, sqlc.arg('created_by')::uuid, sqlc.arg('created_by')::uuid);

-- name: GetEvents :many
SELECT DISTINCT e.*,
                p.name        AS program_name,
                p.description AS program_description,
                p."type"      AS program_type,
                l.name        AS location_name,
                l.address     AS location_address,
                t.name        as team_name
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
WHERE (
          (sqlc.narg('program_id')::uuid = e.program_id OR sqlc.narg('program_id') IS NULL)
              AND (sqlc.narg('location_id')::uuid = e.location_id OR sqlc.narg('location_id') IS NULL)
              AND (sqlc.narg('after')::timestamp <= e.program_start_at OR sqlc.narg('after') IS NULL)
              AND (sqlc.narg('before')::timestamp >= e.program_end_at OR sqlc.narg('before') IS NULL OR
                   e.program_end_at IS NULL)
              AND (sqlc.narg('type') = p.type OR sqlc.narg('type') IS NULL)
              AND (sqlc.narg('user_id')::uuid IS NULL OR ce.customer_id = sqlc.narg('user_id')::uuid OR
                   es.staff_id = sqlc.narg('user_id')::uuid)
              AND (sqlc.narg('team_id')::uuid IS NULL OR e.team_id = sqlc.narg('team_id'))
              AND (sqlc.narg('created_by')::uuid IS NULL OR e.created_by = sqlc.narg('created_by'))
              AND (sqlc.narg('updated_by')::uuid IS NULL OR e.updated_by = sqlc.narg('updated_by'))
          );

-- name: GetEventById :many
SELECT e.*,
       p.name          AS program_name,
       p.description   AS program_description,
       p."type"        AS program_type,
       l.name          AS location_name,
       l.address       AS location_address,
       -- Staff fields
       s.id            AS staff_id,
       sr.role_name    AS staff_role_name,
       us.email        AS staff_email,
       us.first_name   AS staff_first_name,
       us.last_name    AS staff_last_name,
       us.gender       AS staff_gender,
       us.phone        AS staff_phone,
       -- Customer fields
       uc.id           AS customer_id,
       uc.first_name   AS customer_first_name,
       uc.last_name    AS customer_last_name,
       uc.email        AS customer_email,
       uc.phone        AS customer_phone,
       uc.gender       AS customer_gender,
       ce.is_cancelled AS customer_is_cancelled,
       -- Team field (added missing team reference)
       t.id            AS team_id,
       t.name          AS team_name
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN staff.staff s ON es.staff_id = s.id
         LEFT JOIN staff.staff_roles sr ON s.role_id = sr.id
         LEFT JOIN users.users us ON s.id = us.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN users.users uc ON ce.customer_id = uc.id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
WHERE e.id = $1
ORDER BY s.id, uc.id;

-- name: UpdateEvent :exec
UPDATE events.events
SET program_start_at = $1,
    program_end_at   = $2,
    location_id      = $3,
    program_id       = $4,
    event_start_time = $5,
    event_end_time   = $6,
    day              = $7,
    capacity         = $8,
    updated_at = current_timestamp,
    updated_by = sqlc.arg('updated_by')::uuid
WHERE id = $9;

-- name: GetEventCreatedBy :one
SELECT created_by
FROM events.events
WHERE id = $1;

-- name: DeleteEvent :exec
DELETE
FROM events.events
WHERE id = $1;