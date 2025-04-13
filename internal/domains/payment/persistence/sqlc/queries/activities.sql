-- name: GetRegistrationPriceIdForCustomer :one
SELECT f.stripe_price_id, p.pay_per_event
FROM program.fees f
         JOIN program.programs p ON p.id = f.program_id
WHERE f.membership_id = (SELECT mp.membership_id
                         FROM users.customer_membership_plans cmp
                                  LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                                  LEFT JOIN membership.memberships m ON m.id = mp.membership_id
                         WHERE customer_id = $1
                           AND status = 'active'
                         ORDER BY cmp.start_date DESC
                         LIMIT 1)
  AND f.program_id = $2;

-- name: GetProgramIdByStripePriceId :one
SELECT program_id
FROM program.fees
WHERE stripe_price_id = $1;

-- name: GetEventIdByStripePriceId :one
SELECT e.id
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN program.fees f ON p.id = f.program_id
WHERE f.stripe_price_id = $1;