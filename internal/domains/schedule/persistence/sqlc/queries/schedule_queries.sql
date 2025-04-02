-- name: CreateSchedule :exec
INSERT INTO public.schedules (recurrence_start_at, recurrence_end_at, event_start_time, event_end_time, day,
                              location_id,
                              program_id)
VALUES ($1, $2, $3, $4, $5,
        $6, $7);

-- name: GetSchedules :many
SELECT s.*,
       p.name        AS program_name,
       p.description AS program_description,
       p."type"      AS program_type,
       l.name        AS location_name,
       l.address     AS location_address,
       t.name        as team_name
FROM public.schedules s
         LEFT JOIN program.programs p ON s.program_id = p.id
         JOIN location.locations l ON s.location_id = l.id
         LEFT JOIN events.staff es ON s.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON s.id = ce.event_id
         LEFT JOIN athletic.teams t ON t.id = s.team_id
WHERE (
          (sqlc.narg('program_id')::uuid = s.program_id OR sqlc.narg('program_id') IS NULL)
              AND (sqlc.narg('location_id')::uuid = s.location_id OR sqlc.narg('location_id') IS NULL)
              AND (sqlc.narg('type') = p.type OR sqlc.narg('type') IS NULL)
              AND (sqlc.narg('user_id')::uuid IS NULL OR ce.customer_id = sqlc.narg('user_id')::uuid OR
                   es.staff_id = sqlc.narg('user_id')::uuid)
              AND (sqlc.narg('team_id')::uuid IS NULL OR s.team_id = sqlc.narg('team_id'))
          );

-- name: GetScheduleById :one
SELECT schedule.*,
       p.name        AS program_name,
       p.description AS program_description,
       p.type        AS program_type,

       l.name        AS location_name,
       l.address     AS location_address,

       t.name        AS team_name

FROM public.schedules schedule
         LEFT JOIN events.events e ON schedule.id = e.schedule_id
         LEFT JOIN program.programs p ON schedule.program_id = p.id
         JOIN location.locations l ON schedule.location_id = l.id
         LEFT JOIN athletic.teams t ON t.id = schedule.team_id
WHERE schedule.id = $1;

-- name: UpdateSchedule :execrows
UPDATE public.schedules
SET recurrence_start_at = $1,
    recurrence_end_at   = $2,
    location_id         = $3,
    program_id          = $4,
    event_start_time    = $5,
    event_end_time      = $6,
    day                 = $7,
    updated_at          = current_timestamp
WHERE id = $8;

-- name: DeleteSchedule :exec
DELETE
FROM events.events
WHERE id = $1;