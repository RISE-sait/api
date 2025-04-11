-- name: GetProgram :one
SELECT id, name
FROM program.programs
WHERE id = $1;

-- name: GetEventIsExist :one
SELECT EXISTS(SELECT 1 FROM events.events WHERE id = $1);

-- name: GetProgramRegistrationPriceIdForCustomer :one
SELECT f.stripe_price_id
FROM program.fees f
WHERE f.membership_id = (SELECT mp.membership_id
                         FROM users.customer_membership_plans cmp
                                  LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                                  LEFT JOIN membership.memberships m ON m.id = mp.membership_id
                         WHERE customer_id = $1
                           AND status = 'active'
                         ORDER BY cmp.start_date DESC
                         LIMIT 1)
  AND pay_per_event = false
  AND f.program_id = $2;

-- name: GetEventRegistrationPriceIdForCustomer :one
SELECT f.stripe_price_id
FROM program.fees f
WHERE f.membership_id = (SELECT mp.membership_id
                         FROM users.customer_membership_plans cmp
                                  LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                                  LEFT JOIN membership.memberships m ON m.id = mp.membership_id
                         WHERE customer_id = $1
                           AND status = 'active'
                         ORDER BY cmp.start_date DESC
                         LIMIT 1)
  AND pay_per_event = true
  AND f.program_id = (SELECT p.id
                      FROM events.events e
                               LEFT JOIN program.programs p ON p.id = e.program_id
                      WHERE e.id = sqlc.arg('event_id')
                        AND e.start_at > current_timestamp);

-- name: GetProgramIdByStripePriceId :one
SELECT program_id
FROM program.fees
WHERE stripe_price_id = $1;