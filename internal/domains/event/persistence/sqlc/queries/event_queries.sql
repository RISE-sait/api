-- name: CreateEvent :one
INSERT INTO events.events (location_id, program_id, team_id, start_at, end_at, created_by, updated_by, capacity)
VALUES ($1, $2, $3, $4, $5,
        sqlc.arg('created_by')::uuid, sqlc.arg('created_by')::uuid, $6)
RETURNING *;

-- name: CreateEvents :exec
INSERT INTO events.events
(location_id, program_id, team_id, start_at, end_at, created_by, updated_by, capacity, is_cancelled,
 cancellation_reason)
SELECT unnest($1::uuid[]),
       unnest($2::uuid[]),
       unnest($3::uuid[]),
       unnest($4::timestamptz[]),
       unnest($5::timestamptz[]),
       unnest($6::uuid[]),
       unnest($7::uuid[]),
       unnest($8::int[]),
       unnest($9::bool[]),
       unnest($10::text[]);

-- name: GetEvents :many
SELECT DISTINCT e.*,

                creator.first_name AS creator_first_name,
                creator.last_name  AS creator_last_name,

                updater.first_name AS updater_first_name,
                updater.last_name  AS updater_last_name,

                p.name             AS program_name,
                p.description      AS program_description,
                p."type"           AS program_type,
                l.name             AS location_name,
                l.address          AS location_address,
                t.name             as team_name
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
         JOIN users.users creator ON creator.id = e.created_by
         JOIN users.users updater ON updater.id = e.updated_by
WHERE (
          (sqlc.narg('program_id')::uuid = e.program_id OR sqlc.narg('program_id') IS NULL)
              AND (sqlc.narg('location_id')::uuid = e.location_id OR sqlc.narg('location_id') IS NULL)
              AND (sqlc.narg('after')::timestamp <= e.start_at OR sqlc.narg('after') IS NULL)
              AND (sqlc.narg('before')::timestamp >= e.end_at OR sqlc.narg('before') IS NULL)
              AND (sqlc.narg('type') = p.type OR sqlc.narg('type') IS NULL)
              AND (sqlc.narg('user_id')::uuid IS NULL OR ce.customer_id = sqlc.narg('user_id')::uuid OR
                   es.staff_id = sqlc.narg('user_id')::uuid)
              AND (sqlc.narg('team_id')::uuid IS NULL OR e.team_id = sqlc.narg('team_id'))
              AND (sqlc.narg('created_by')::uuid IS NULL OR e.created_by = sqlc.narg('created_by'))
              AND (sqlc.narg('updated_by')::uuid IS NULL OR e.updated_by = sqlc.narg('updated_by'))
              AND (sqlc.narg('include_cancelled')::boolean IS NULL OR e.is_cancelled = sqlc.narg('include_cancelled'))
          );

-- name: GetEventById :many
SELECT e.*,

       creator.first_name AS creator_first_name,
       creator.last_name  AS creator_last_name,

       updater.first_name AS updater_first_name,
       updater.last_name  AS updater_last_name,

       p.name             AS program_name,
       p.description      AS program_description,
       p."type"           AS program_type,
       l.name             AS location_name,
       l.address          AS location_address,
       -- Staff fields
       s.id               AS staff_id,
       sr.role_name       AS staff_role_name,
       us.email           AS staff_email,
       us.first_name      AS staff_first_name,
       us.last_name       AS staff_last_name,
       us.gender          AS staff_gender,
       us.phone           AS staff_phone,
       -- Customer fields
       uc.id              AS customer_id,
       uc.first_name      AS customer_first_name,
       uc.last_name       AS customer_last_name,
       uc.email           AS customer_email,
       uc.phone           AS customer_phone,
       uc.gender          AS customer_gender,

       ce.is_cancelled    AS customer_enrollment_is_cancelled,

       -- Team field (added missing team reference)
       t.id               AS team_id,
       t.name             AS team_name
FROM events.events e
         JOIN users.users creator ON creator.id = e.created_by
         JOIN users.users updater ON updater.id = e.updated_by
         LEFT JOIN program.programs p ON e.program_id = p.id
         JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN staff.staff s ON es.staff_id = s.id
         LEFT JOIN staff.staff_roles sr ON s.role_id = sr.id
         LEFT JOIN users.users us ON s.id = us.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN users.users uc ON ce.customer_id = uc.id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
WHERE e.id = $1
ORDER BY s.id, uc.id;

-- name: UpdateEvent :one
UPDATE events.events
SET start_at            = $1,
    end_at              = $2,
    location_id         = $3,
    program_id          = $4,
    team_id             = $5,
    is_cancelled        = $6,
    cancellation_reason = $7,
    capacity            = $8,
    updated_at          = current_timestamp,
    updated_by          = sqlc.arg('updated_by')::uuid
WHERE id = $9
RETURNING *;

-- name: GetEventCreatedBy :one
SELECT created_by
FROM events.events
WHERE id = $1;

-- name: DeleteEvent :exec
DELETE
FROM events.events
WHERE id = $1;