-- name: IsCustomerExist :one
SELECT EXISTS(SELECT 1 FROM users.users WHERE id = $1);

-- name: GetCustomersTeam :one
SELECT t.id
FROM athletic.athletes a
         LEFT JOIN athletic.teams t ON a.team_id = t.id
WHERE a.id = sqlc.arg('customer_id');

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