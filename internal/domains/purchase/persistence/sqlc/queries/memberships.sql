-- name: CreateCustomerMembershipPlan :exec
INSERT INTO customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlanJoiningFee :one
SELECT joining_fee
FROM membership.membership_plans
WHERE id = $1;