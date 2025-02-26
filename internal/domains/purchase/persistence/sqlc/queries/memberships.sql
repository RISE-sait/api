-- name: PurchaseMembership :execrows
INSERT INTO customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);