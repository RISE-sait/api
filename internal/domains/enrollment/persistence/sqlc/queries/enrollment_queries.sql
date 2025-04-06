-- name: EnrollCustomerInEvent :exec
INSERT INTO events.customer_enrollment (customer_id, event_id)
VALUES ($1, $2);

-- name: EnrollCustomerInProgramEvents :exec
WITH eligible_events AS (SELECT e.id
                         FROM events.events e
                         WHERE e.program_id = $1
                           AND e.start_at >= current_timestamp
                           AND NOT EXISTS (SELECT 1
                                           FROM events.customer_enrollment ce
                                           WHERE ce.customer_id = $2
                                             AND ce.event_id = e.id)),
     event_inserts as (
         INSERT INTO events.customer_enrollment (customer_id, event_id)
             SELECT $2, id FROM eligible_events)
INSERT
INTO program.customer_enrollment(customer_id, program_id, is_cancelled)
VALUES ($2, $1, false);

-- name: EnrollCustomerInMembershipPlan :exec
INSERT INTO users.customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: UnEnrollCustomer :execrows
UPDATE events.customer_enrollment
SET is_cancelled = true
WHERE customer_id = $1
  AND event_id = $2;

-- name: GetEventIsFull :one
SELECT COUNT(ce.customer_id) >= COALESCE(e.capacity, p.capacity, t.capacity)::boolean AS is_full
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN athletic.teams t ON e.team_id = t.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
WHERE e.id = @event_id
GROUP BY e.id, e.capacity, p.capacity, t.capacity;