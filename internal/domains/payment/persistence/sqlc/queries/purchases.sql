-- name: CreateCustomerMembershipPlan :exec
INSERT INTO customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlanJoiningRequirements :one
SELECT *
FROM membership.membership_plans
WHERE id = $1;

SELecT mp.id
FROM program_membership pm
         LEFT JOIN membership.memberships m ON m.id = pm.membership_id
         LEFT JOIN membership.membership_plans mp ON mp.membership_id = m.id
         LEFT JOIN customer_membership_plans cmp ON cmp.membership_plan_id = mp.id
WHERE pm.program_id = '9b8e816e-7b53-4586-a676-4777554af79c'
  AND pm.is_eligible = true
  AND cmp.customer_id = '70ad95a7-228e-4170-b363-3fe501cc5c08';

-- name: GetProgramRegisterInfoForCustomer :one
SELECT pm.is_eligible,
       pm.price_per_booking,
       p2.name                                                                       AS program_name,
       EXISTS (SELECT 1 FROM users.users u WHERE u.id = sqlc.arg('customer_id'))     AS customer_exists,
       EXISTS (SELECT 1 FROM program.programs p WHERE p.id = sqlc.arg('program_id')) AS program_exists,
       EXISTS (SELECT 1
               FROM public.customer_membership_plans
               WHERE customer_id = sqlc.arg('customer_id')
                 AND status = 'active')                                              AS customer_has_active_membership
FROM program.programs p2
         LEFT JOIN public.program_membership pm ON pm.program_id = p2.id
         LEFT JOIN membership.membership_plans mp ON mp.membership_id = pm.membership_id
         LEFT JOIN public.customer_membership_plans cmp_active
                   ON cmp_active.membership_plan_id = mp.id
WHERE p2.id = sqlc.arg('program_id')
  AND cmp_active.customer_id = sqlc.arg('customer_id')
  AND cmp_active.status = 'active'
GROUP BY pm.is_eligible, pm.price_per_booking, p2.name;