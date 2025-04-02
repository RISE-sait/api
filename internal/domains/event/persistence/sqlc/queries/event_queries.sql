-- name: CreateEvent :exec
INSERT INTO events.events (schedule_id, start_at, end_at, created_by, updated_by, capacity, location_id)
VALUES ($1, $2, $3,
        sqlc.arg('created_by')::uuid, sqlc.arg('created_by')::uuid, $4, $5);

-- name: GetEvents :many
SELECT distinct e.id,
                e.*,
                coalesce(e.capacity, p.capacity, t.capacity) AS capacity,

                p.id                                         AS program_id,
                p.name        AS program_name,
                p.description AS program_description,
                p."type"      AS program_type,

                l.name        AS location_name,
                l.address     AS location_address,

                t.id                                         as team_id,
                t.name                                       as team_name,

                creator.id                                   AS creator_id,
                creator.first_name                           AS creator_first_name,
                creator.last_name                            AS creator_last_name,
                creator.email                                AS creator_email,

                updater.id                                   AS updater_id,
                updater.first_name                           AS updater_first_name,
                updater.last_name                            AS updater_last_name,
                updater.email                                AS updater_email

FROM events.events e
         LEFT JOIN public.schedules s ON e.schedule_id = s.id
         LEFT JOIN program.programs p ON s.program_id = p.id
         INNER JOIN location.locations l ON coalesce(e.location_id, s.location_id) = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN athletic.teams t ON t.id = s.team_id
         JOIN users.users creator ON creator.id = e.created_by
         JOIN users.users updater ON updater.id = e.updated_by

WHERE (
          (sqlc.narg('program_id')::uuid = s.program_id OR sqlc.narg('program_id') IS NULL)
              AND (sqlc.narg('location_id')::uuid = s.location_id OR sqlc.narg('location_id') IS NULL)
              AND (e.start_at > sqlc.narg('after')::timestamp OR sqlc.narg('after') IS NULL)
              AND (e.end_at < sqlc.narg('before')::timestamp OR sqlc.narg('before') IS NULL)
              AND (sqlc.narg('type') = p.type OR sqlc.narg('type') IS NULL)
              AND (sqlc.narg('user_id')::uuid IS NULL OR ce.customer_id = sqlc.narg('user_id') OR
                   es.staff_id = sqlc.narg('user_id')::uuid)
              AND (sqlc.narg('team_id')::uuid IS NULL OR s.team_id = sqlc.narg('team_id'))
              AND (sqlc.narg('created_by')::uuid IS NULL OR e.created_by = sqlc.narg('created_by'))
              AND (sqlc.narg('updated_by')::uuid IS NULL OR e.updated_by = sqlc.narg('updated_by'))
          );

-- name: GetEventById :many
SELECT e.*,
       coalesce(e.capacity, p.capacity, t.capacity) AS capacity,

       p.id                                         AS program_id,
       p.name          AS program_name,
       p.description   AS program_description,
       p."type"        AS program_type,

       l.id                                         as location_id,
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

       -- Team field (added missing team reference)
       t.id            AS team_id,
       t.name                                       AS team_name,

       creator.id                                   AS creator_id,
       creator.first_name                           AS creator_first_name,
       creator.last_name                            AS creator_last_name,
       creator.email                                AS creator_email,

       updater.id                                   AS updater_id,
       updater.first_name                           AS updater_first_name,
       updater.last_name                            AS updater_last_name,
       updater.email                                AS updater_email
FROM events.events e
         LEFT JOIN public.schedules schedule ON e.schedule_id = schedule.id
         LEFT JOIN program.programs p ON schedule.program_id = p.id
         INNER JOIN location.locations l ON coalesce(e.location_id, schedule.location_id) = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN staff.staff s ON es.staff_id = s.id
         LEFT JOIN staff.staff_roles sr ON s.role_id = sr.id
         LEFT JOIN users.users us ON s.id = us.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN users.users uc ON ce.customer_id = uc.id
         LEFT JOIN athletic.teams t ON t.id = schedule.team_id
         JOIN users.users creator ON creator.id = e.created_by
         JOIN users.users updater ON updater.id = e.updated_by
WHERE e.id = $1
ORDER BY s.id, uc.id;

-- name: UpdateEvent :exec
UPDATE events.events
SET start_at    = $1,
    end_at      = $2,
    schedule_id = $3,
    capacity    = $4,
    location_id = $5,
    updated_by = sqlc.arg('updated_by')::uuid
WHERE id = $6;

-- name: GetEventCreatedBy :one
SELECT created_by
FROM events.events
WHERE id = $1;

-- name: DeleteEvent :exec
DELETE
FROM events.events
WHERE id = $1;