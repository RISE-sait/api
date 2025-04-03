-- name: CreateCustomerMembershipPlan :exec
INSERT INTO users.customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlanJoiningRequirements :one
SELECT *
FROM membership.membership_plans
WHERE id = $1;

-- name: GetProgram :one
SELECT id, name
FROM program.programs
WHERE id = $1;

-- name: IsCustomerExist :one
SELECT EXISTS(SELECT 1 FROM users.users WHERE id = $1);

-- name: GetProgramRegistrationPriceIDForCustomer :one
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