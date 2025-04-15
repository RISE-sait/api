-- name: GetEventsSchedules :many
SELECT trim(to_char(start_at, 'Day')) AS day_of_week, -- More readable
       to_char(start_at, 'HH24:MI')   AS start_time,
       to_char(end_at, 'HH24:MI')     AS end_time,
       p.id                           AS program_id,
       p.name                         AS program_name,
       p.description                  AS program_description,
       p.type                         AS program_type,

       location_id,
       l.name                         AS location_name,
       l.address                      AS location_address,

       t.id                           AS team_id,
       t.name                         AS team_name,

       COUNT(*)                       AS event_count,
       MIN(start_at)::timestamp       AS first_occurrence,
       MAX(end_at)::timestamp         AS last_occurrence
FROM events.events e
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN athletic.teams t ON e.team_id = t.id
         JOIN location.locations l ON e.location_id = l.id
WHERE (sqlc.narg('program_id')::uuid IS NULL OR program_id = sqlc.narg('program_id')::uuid)
  AND (sqlc.narg('team_id')::uuid IS NULL OR e.team_id = sqlc.narg('team_id'))
  AND (sqlc.narg('location_id')::uuid IS NULL OR location_id = sqlc.narg('location_id')::uuid)
  AND (sqlc.narg('created_by')::uuid IS NULL OR e.created_by = sqlc.narg('created_by'))
  AND (sqlc.narg('updated_by')::uuid IS NULL OR e.updated_by = sqlc.narg('updated_by'))
  AND (sqlc.narg('after')::timestamp IS NULL OR start_at >= sqlc.narg('after')::timestamp)
  AND (sqlc.narg('before')::timestamp IS NULL OR end_at <= sqlc.narg('before')::timestamp)
  AND (sqlc.narg('type') = p.type OR sqlc.narg('type') IS NULL)
  AND (sqlc.narg('user_id')::uuid IS NULL OR ce.customer_id = sqlc.narg('user_id')::uuid OR
       es.staff_id = sqlc.narg('user_id')::uuid)
GROUP BY to_char(start_at, 'Day'),
         to_char(start_at, 'HH24:MI'),
         to_char(end_at, 'HH24:MI'),
         p.id,
         p.name,
         p.description,
         p.type,
         t.id,
         t.name,
         location_id,
         l.name,
         l.address;