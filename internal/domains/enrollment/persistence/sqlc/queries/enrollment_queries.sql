-- name: EnrollCustomerInEvent :exec
INSERT INTO events.customer_enrollment (customer_id, event_id)
VALUES ($1, $2);

-- name: EnrollCustomerInProgram :exec
WITH events AS (SELECT e.id
                         FROM events.events e
                         WHERE e.program_id = $1
                           AND e.start_at >= current_timestamp),
     event_inserts as (
         INSERT INTO events.customer_enrollment (customer_id, event_id)
             SELECT $2, id FROM events)
INSERT
INTO program.customer_enrollment(customer_id, program_id, is_cancelled)
VALUES ($2, $1, false);

-- name: EnrollCustomerInMembershipPlan :exec
INSERT INTO users.customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: UnEnrollCustomerFromEvent :execrows
UPDATE events.customer_enrollment
SET is_cancelled = true
WHERE customer_id = $1
  AND event_id = $2;

-- name: CheckProgramExists :one
SELECT EXISTS (SELECT 1
               FROM program.programs p
               WHERE p.id = $1);

-- name: CheckProgramCapacityExists :one
SELECT (capacity IS NOT NULL)::boolean
FROM program.programs p
WHERE p.id = $1;

-- name: CheckProgramIsFull :one
SELECT COUNT(ce.id) FILTER (
    WHERE ce.payment_status = 'paid'
        OR (ce.payment_status = 'pending' AND ce.payment_expired_at > CURRENT_TIMESTAMP))
           >= p.capacity
FROM program.programs p
         LEFT JOIN program.customer_enrollment ce ON p.id = ce.program_id
WHERE p.id = $1
GROUP BY p.capacity;

-- name: ReserveSeatInProgram :execrows
INSERT INTO program.customer_enrollment
    (customer_id, program_id, payment_expired_at, payment_status)
VALUES ($1, $2, CURRENT_TIMESTAMP + interval '10 minute', 'pending')
ON CONFLICT (customer_id, program_id)
    DO UPDATE SET payment_expired_at = EXCLUDED.payment_expired_at,
                  payment_status     = EXCLUDED.payment_status
WHERE program.customer_enrollment.payment_status != 'paid';

-- name: UpdateSeatReservationStatusInProgram :execrows
UPDATE program.customer_enrollment
SET payment_status = $1
WHERE customer_id = $2
  AND program_id = $3;

-- name: CheckEventIsFull :one
SELECT COUNT(ce.id) FILTER (
    WHERE ce.payment_status = 'paid'
        OR (ce.payment_status = 'pending' AND ce.payment_expired_at > NOW())
    ) >= COALESCE(e.capacity, p.capacity, t.capacity) AS is_full
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN athletic.teams t ON e.team_id = t.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
WHERE e.id = $1
GROUP BY e.capacity, p.capacity, t.capacity;

-- name: ReserveSeatInEvent :execrows
UPDATE events.customer_enrollment
SET payment_expired_at = CURRENT_TIMESTAMP + interval '10 minute',
    payment_status     = 'pending'
WHERE customer_id = $1
  AND event_id = $2;

-- name: GetTeamOfEvent :one
SELECT t.id
FROM events.events e
         LEFT JOIN athletic.teams t ON e.team_id = t.id
WHERE e.id = sqlc.arg('event_id');