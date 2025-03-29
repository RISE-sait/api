-- name: CreateCustomerMembershipPlan :exec
INSERT INTO customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlanJoiningRequirements :one
SELECT *
FROM membership.membership_plans
WHERE id = $1;

-- name: GetProgram :one
SELECT id, name
FROM program.programs
WHERE id = $1;

-- name: GetPaygPrice :one
SELECT payg_price
FROM program.programs
WHERE id = $1;

-- name: IsCustomerExist :one
SELECT EXISTS(SELECT 1 FROM users.users WHERE id = $1);

-- name: GetCustomerHasActiveMembershipPlan :one
SELECT EXISTS(SELECT 1
              FROM public.customer_membership_plans
              WHERE customer_id = sqlc.arg('customer_id')
                AND status = 'active');

-- name: GetProgramRegisterInfoForCustomer :one
SELECT pm.price_per_booking
FROM program.programs p
         LEFT JOIN public.program_membership pm ON pm.program_id = p.id
         LEFT JOIN membership.membership_plans mp ON mp.membership_id = pm.membership_id
         LEFT JOIN public.customer_membership_plans cmp_active
                   ON cmp_active.membership_plan_id = mp.id
WHERE p.id = sqlc.arg('program_id')
  AND cmp_active.customer_id = sqlc.arg('customer_id')
  AND cmp_active.status = 'active'
GROUP BY pm.price_per_booking, p.name;