-- name: CreateEvents :execrows
WITH unnested_data AS (
    SELECT
        unnest(sqlc.arg('location_ids')::uuid[])                 AS location_id,
        unnest(sqlc.arg('program_ids')::uuid[])                  AS program_id,
        unnest(sqlc.arg('court_ids')::uuid[])                    AS court_id,
        unnest(sqlc.arg('team_ids')::uuid[])                     AS team_id,
        unnest(sqlc.arg('start_at_array')::timestamptz[])        AS start_at,
        unnest(sqlc.arg('end_at_array')::timestamptz[])          AS end_at,
        unnest(sqlc.arg('is_date_time_modified_array')::bool[])  AS is_date_time_modified,
        unnest(sqlc.arg('recurrence_ids')::uuid[])               AS recurrence_id,
        unnest(sqlc.arg('created_by_ids')::uuid[])               AS created_by,
        unnest(sqlc.arg('is_cancelled_array')::bool[])           AS is_cancelled,
        unnest(sqlc.arg('cancellation_reasons')::text[])         AS cancellation_reason,
        unnest(sqlc.arg('price_ids')::text[])                    AS price_id,
        unnest(sqlc.arg('credit_costs')::int[])                  AS credit_cost,
        unnest(sqlc.arg('registration_required_array')::bool[])  AS registration_required
)
INSERT INTO events.events (
    location_id,
    court_id,
    program_id,
    team_id,
    start_at,
    end_at,
    is_date_time_modified,
    recurrence_id,
    created_by,
    updated_by,
    is_cancelled,
    cancellation_reason,
    price_id,
    credit_cost,
    registration_required
)
SELECT
    location_id,
    NULLIF(court_id, '00000000-0000-0000-0000-000000000000'::uuid),
    program_id,
    NULLIF(team_id, '00000000-0000-0000-0000-000000000000'::uuid),
    start_at,
    end_at,
    is_date_time_modified,
    NULLIF(recurrence_id, '00000000-0000-0000-0000-000000000000'::uuid),
    created_by,
    created_by,
    is_cancelled,
    NULLIF(cancellation_reason, ''),
    NULLIF(price_id, ''),
    NULLIF(credit_cost, 0),
    registration_required
FROM unnested_data
ON CONFLICT ON CONSTRAINT no_overlapping_events DO NOTHING;



-- name: GetEvents :many
SELECT DISTINCT e.*,

                creator.first_name AS creator_first_name,
                creator.last_name  AS creator_last_name,

                updater.first_name AS updater_first_name,
                updater.last_name  AS updater_last_name,

                p.name             AS program_name,
                p.description      AS program_description,
                p."type"           AS program_type,
                p.photo_url        AS program_photo_url,
                l.name             AS location_name,
                l.address          AS location_address,
                c.name             AS court_name,
                t.name             as team_name
FROM events.events e
         JOIN program.programs p ON e.program_id = p.id
         JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN location.courts c ON e.court_id = c.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
         JOIN users.users creator ON creator.id = e.created_by
         JOIN users.users updater ON updater.id = e.updated_by
WHERE (
          (sqlc.narg('program_id')::uuid = e.program_id OR sqlc.narg('program_id') IS NULL)
              AND (sqlc.narg('location_id')::uuid = e.location_id OR sqlc.narg('location_id') IS NULL)
              AND (sqlc.narg('court_id')::uuid = e.court_id OR sqlc.narg('court_id') IS NULL)
              AND (sqlc.narg('after')::timestamptz <= e.start_at OR sqlc.narg('after') IS NULL)
              AND (sqlc.narg('before')::timestamptz >= e.end_at OR sqlc.narg('before') IS NULL)
              AND (sqlc.narg('type') = p.type OR sqlc.narg('type') IS NULL)
              AND (sqlc.narg('participant_id')::uuid IS NULL OR ce.customer_id = sqlc.narg('participant_id')::uuid OR
                   es.staff_id = sqlc.narg('participant_id')::uuid)
              AND (sqlc.narg('team_id')::uuid IS NULL OR e.team_id = sqlc.narg('team_id'))
              AND (sqlc.narg('created_by')::uuid IS NULL OR e.created_by = sqlc.narg('created_by'))
              AND (sqlc.narg('updated_by')::uuid IS NULL OR e.updated_by = sqlc.narg('updated_by'))
              AND (sqlc.narg('include_cancelled')::boolean IS NULL OR e.is_cancelled = sqlc.narg('include_cancelled'))
          )
          OFFSET sqlc.narg('offset') LIMIT sqlc.narg('limit');


-- name: GetEventCustomers :many
SELECT u.id            AS customer_id,
       u.first_name    AS customer_first_name,
       u.last_name     AS customer_last_name,
       u.email         AS customer_email,
       u.phone         AS customer_phone,
       u.gender        AS customer_gender,

       ce.is_cancelled AS customer_enrollment_is_cancelled

FROM events.customer_enrollment ce
         JOIN users.users u ON ce.customer_id = u.id
WHERE ce.event_id = $1;

-- name: GetEventStaffs :many
SELECT s.id         AS staff_id,
       sr.role_name AS staff_role_name,
       u.email      AS staff_email,
       u.first_name AS staff_first_name,
       u.last_name  AS staff_last_name,
       u.gender     AS staff_gender,
       u.phone      AS staff_phone
FROM events.staff es
         JOIN staff.staff s ON es.staff_id = s.id
         JOIN staff.staff_roles sr ON s.role_id = sr.id
         JOIN users.users u ON s.id = u.id
WHERE es.event_id = $1;

-- name: GetEventById :one
SELECT e.*,

       creator.first_name AS creator_first_name,
       creator.last_name  AS creator_last_name,

       updater.first_name AS updater_first_name,
       updater.last_name  AS updater_last_name,

       p.name             AS program_name,
       p.description      AS program_description,
       p."type"           AS program_type,
       p.photo_url        AS program_photo_url,
       l.name             AS location_name,
       l.address          AS location_address,
       c.name             AS court_name,
       t.name             AS team_name
FROM events.events e
         JOIN users.users creator ON creator.id = e.created_by
         JOIN users.users updater ON updater.id = e.updated_by
         JOIN program.programs p ON e.program_id = p.id
         JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN location.courts c ON e.court_id = c.id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
WHERE e.id = $1;

-- name: UpdateEvent :one
UPDATE events.events
SET start_at              = $1,
    end_at                = $2,
    location_id           = $3,
    program_id            = $4,
    court_id              = $5,
    team_id               = $6,
    is_cancelled          = $7,
    cancellation_reason   = $8,
    updated_at            = current_timestamp,
    updated_by            = sqlc.arg('updated_by')::uuid,
    is_date_time_modified = (recurrence_id IS NOT NULL),
    price_id              = $10,
    credit_cost           = $11,
    registration_required = $12
  WHERE id = $9
  RETURNING *;

-- name: DeleteEventsByIds :exec
DELETE
FROM events.events
WHERE id = ANY (sqlc.arg('ids')::uuid[]);

-- name: AddEventMembershipAccess :exec
INSERT INTO events.event_membership_access (event_id, membership_plan_id)
VALUES ($1, $2)
ON CONFLICT (event_id, membership_plan_id) DO NOTHING;

-- name: RemoveEventMembershipAccess :exec
DELETE FROM events.event_membership_access
WHERE event_id = $1 AND membership_plan_id = $2;

-- name: GetEventMembershipPlans :many
SELECT mp.id, mp.name
FROM events.event_membership_access ema
JOIN membership.membership_plans mp ON ema.membership_plan_id = mp.id
WHERE ema.event_id = $1;

-- name: CheckCustomerHasEventMembershipAccess :one
SELECT EXISTS(
    SELECT 1
    FROM events.event_membership_access ema
    JOIN users.customer_membership_plans cmp ON ema.membership_plan_id = cmp.membership_plan_id
    WHERE ema.event_id = $1
      AND cmp.customer_id = $2
      AND cmp.status = 'active'
) AS has_access;

-- name: ClearEventMembershipAccess :exec
DELETE FROM events.event_membership_access
WHERE event_id = $1;

-- name: SetEventMembershipAccess :exec
WITH deleted AS (
    DELETE FROM events.event_membership_access
    WHERE event_id = $1
)
INSERT INTO events.event_membership_access (event_id, membership_plan_id)
SELECT $1, unnest(sqlc.arg('membership_plan_ids')::uuid[])
ON CONFLICT (event_id, membership_plan_id) DO NOTHING;