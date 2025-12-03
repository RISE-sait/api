-- name: EnrollCustomerInEvent :exec
INSERT INTO events.customer_enrollment (customer_id, event_id, payment_status)
VALUES ($1, $2, 'paid');

-- name: EnrollCustomerInProgram :exec
WITH events AS (SELECT e.id
                         FROM events.events e
                         WHERE e.program_id = $1
                           AND e.start_at >= current_timestamp),
     event_inserts as (
         INSERT INTO events.customer_enrollment (customer_id, event_id, payment_status)
             SELECT $2, id, 'paid' FROM events)
INSERT
INTO program.customer_enrollment(customer_id, program_id, is_cancelled, payment_status)
VALUES ($2, $1, false, 'paid');

-- name: EnrollCustomerInMembershipPlan :exec
INSERT INTO users.customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date, next_billing_date, subscription_source, stripe_subscription_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (customer_id, membership_plan_id)
DO UPDATE SET
    status = EXCLUDED.status,
    start_date = EXCLUDED.start_date,
    renewal_date = EXCLUDED.renewal_date,
    next_billing_date = EXCLUDED.next_billing_date,
    subscription_source = EXCLUDED.subscription_source,
    stripe_subscription_id = EXCLUDED.stripe_subscription_id,
    updated_at = CURRENT_TIMESTAMP;

-- name: UnEnrollCustomerFromEvent :execrows
UPDATE events.customer_enrollment
SET is_cancelled = true
WHERE customer_id = $1
  AND event_id = $2;

-- name: RemoveCustomerFromEvent :execrows
DELETE FROM events.customer_enrollment
WHERE customer_id = $1
  AND event_id = $2;

-- name: CheckProgramCapacityExists :one
SELECT (capacity IS NOT NULL)::boolean
FROM program.programs p
WHERE p.id = $1;

-- name: CheckEventCapacityExists :one
SELECT (coalesce(t.capacity, p.capacity) IS NOT NULL)::boolean
FROM events.events e
         LEFT JOIN athletic.teams t ON e.team_id = t.id
         LEFT JOIN program.programs p ON e.program_id = p.id
WHERE e.id = $1;

-- name: CheckProgramIsFull :one
SELECT COUNT(ce.id) FILTER (
    WHERE ce.payment_status = 'paid'
        OR (ce.payment_status = 'pending' AND ce.payment_expired_at > CURRENT_TIMESTAMP))
           >= p.capacity
FROM program.programs p
         LEFT JOIN program.customer_enrollment ce ON p.id = ce.program_id
WHERE p.id = $1
GROUP BY p.capacity;

-- name: GetCustomerIsEnrolledInProgram :one
SELECT (EXISTS (SELECT 1
                FROM program.customer_enrollment
                WHERE customer_id = $1
                  AND program_id = $2
                  AND payment_status = 'paid'));

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

-- name: UpdateSeatReservationStatusInEvent :execrows
UPDATE events.customer_enrollment
SET payment_status = $1
WHERE customer_id = $2
  AND event_id = $3;

-- name: CheckEventIsFull :one
SELECT COUNT(ce.id) FILTER (
    WHERE ce.payment_status = 'paid'
        OR (ce.payment_status = 'pending' AND ce.payment_expired_at > NOW())
    ) >= COALESCE(p.capacity, t.capacity) AS is_full
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN athletic.teams t ON e.team_id = t.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
WHERE e.id = $1
GROUP BY p.capacity, t.capacity;

-- name: ReserveSeatInEvent :execrows
INSERT INTO events.customer_enrollment
    (customer_id, event_id, payment_expired_at, payment_status)
VALUES ($1, $2, CURRENT_TIMESTAMP + interval '10 minute', 'pending')
ON CONFLICT (customer_id, event_id)
    DO UPDATE SET payment_expired_at = EXCLUDED.payment_expired_at,
                  payment_status     = EXCLUDED.payment_status
WHERE events.customer_enrollment.payment_status != 'paid';

-- name: GetTeamOfEvent :one
SELECT t.id
FROM events.events e
         LEFT JOIN athletic.teams t ON e.team_id = t.id
WHERE e.id = sqlc.arg('event_id');

-- name: UpdateMembershipPlanStatus :execrows
UPDATE users.customer_membership_plans 
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE customer_id = $2 AND membership_plan_id = $3;

-- name: UpdateMembershipPlanStatusByStripeSubscription :execrows
UPDATE users.customer_membership_plans 
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE customer_id = $2 AND subscription_source = 'stripe';

-- name: GetMembershipPlanByCustomerId :one
SELECT customer_id, membership_plan_id, status
FROM users.customer_membership_plans
WHERE customer_id = $1 AND subscription_source = 'stripe'
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateMembershipPlanByCustomerId :execrows
UPDATE users.customer_membership_plans
SET status = $1, updated_at = CURRENT_TIMESTAMP
WHERE customer_id = $2 AND subscription_source = 'stripe';

-- name: UpdateMembershipStatusAndNextBilling :execrows
UPDATE users.customer_membership_plans
SET status = $1, next_billing_date = $2, updated_at = CURRENT_TIMESTAMP
WHERE customer_id = $3 AND subscription_source = 'stripe';

-- name: GetCustomerActiveMembershipPlans :many
SELECT customer_id, membership_plan_id, status, start_date, renewal_date, next_billing_date, stripe_subscription_id
FROM users.customer_membership_plans
WHERE customer_id = $1 AND status = 'active'
ORDER BY created_at DESC;