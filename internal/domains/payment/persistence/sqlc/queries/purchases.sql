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

-- name: IsCustomerExist :one
SELECT EXISTS(SELECT 1 FROM users.users WHERE id = $1);

-- name: GetProgramRegisterPricesForCustomer :one
WITH active_membership_id AS
         (SELECT mp.membership_id
          FROM public.customer_membership_plans cmp
                   LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                   LEFT JOIN membership.memberships m ON m.id = mp.membership_id
              WHERE customer_id = sqlc.arg('customer_id')
                AND status = 'active'
          ORDER BY cmp.start_date DESC
          LIMIT 1)

SELECT
    -- Member program price (if available)
    (SELECT ef.program_price
     FROM public.enrollment_fees ef
     WHERE ef.program_id = p.id
       AND ef.membership_id = (SELECT membership_id FROM active_membership_id))
           AS member_program_price,

    -- Member drop-in price (if available)
    (SELECT ef.drop_in_price
     FROM public.enrollment_fees ef
     WHERE ef.program_id = p.id
       AND ef.membership_id = (SELECT membership_id FROM active_membership_id))
           AS member_drop_in_price,

    -- Non-member program price
    (SELECT ef.program_price
     FROM public.enrollment_fees ef
     WHERE ef.program_id = p.id
       AND ef.membership_id IS NULL)
           AS non_member_program_price,

    -- Non-member drop-in price
    (SELECT ef.drop_in_price
     FROM public.enrollment_fees ef
     WHERE ef.program_id = p.id
       AND ef.membership_id IS NULL)
           AS non_member_drop_in_price,

    p.name AS program_name
FROM program.programs p
WHERE p.id = sqlc.arg('program_id');