-- name: GetProgram :one
SELECT id, name
FROM program.programs
WHERE id = $1;

-- name: GetProgramRegistrationPriceIdForCustomer :one
SELECT pm.stripe_program_price_id
FROM program.program_membership pm
WHERE pm.membership_id = (SELECT mp.membership_id
                          FROM users.customer_membership_plans cmp
                                   LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                                   LEFT JOIN membership.memberships m ON m.id = mp.membership_id
                          WHERE customer_id = $1
                            AND status = 'active'
                          ORDER BY cmp.start_date DESC
                          LIMIT 1)
  AND pm.program_id = $2;

-- name: GetProgramCapacityStatus :one
SELECT capacity,
       (SELECT COUNT(*) FROM program.customer_enrollment ce WHERE ce.program_id = $1) AS enrolled_count
FROM program.programs
WHERE id = $1;

-- name: GetProgramIdByStripePriceId :one
SELECT pm.program_id
FROM program.program_membership pm
WHERE pm.stripe_program_price_id = $1;