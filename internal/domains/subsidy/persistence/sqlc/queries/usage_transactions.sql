-- name: CreateUsageTransaction :one
INSERT INTO subsidies.usage_transactions (
    customer_subsidy_id,
    customer_id,
    transaction_type,
    membership_plan_id,
    original_amount,
    subsidy_applied,
    customer_paid,
    stripe_subscription_id,
    stripe_invoice_id,
    stripe_payment_intent_id,
    description
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetUsageTransaction :one
SELECT * FROM subsidies.usage_transactions
WHERE id = $1;

-- name: GetUsageTransactionByInvoice :one
SELECT * FROM subsidies.usage_transactions
WHERE stripe_invoice_id = $1
LIMIT 1;

-- name: ListUsageTransactions :many
SELECT
    ut.*,
    mp.name as membership_plan_name
FROM subsidies.usage_transactions ut
LEFT JOIN membership.membership_plans mp ON mp.id = ut.membership_plan_id
WHERE ut.customer_subsidy_id = $1
ORDER BY ut.applied_at DESC;

-- name: ListUsageTransactionsByCustomer :many
SELECT
    ut.*,
    mp.name as membership_plan_name,
    cs.provider_id
FROM subsidies.usage_transactions ut
LEFT JOIN membership.membership_plans mp ON mp.id = ut.membership_plan_id
LEFT JOIN subsidies.customer_subsidies cs ON cs.id = ut.customer_subsidy_id
WHERE ut.customer_id = $1
ORDER BY ut.applied_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsageTransactionsByCustomer :one
SELECT COUNT(*) FROM subsidies.usage_transactions
WHERE customer_id = $1;

-- name: GetSubsidyUsageStats :one
SELECT
    COUNT(*) as transaction_count,
    COALESCE(SUM(subsidy_applied), 0) as total_subsidy_used,
    COALESCE(SUM(customer_paid), 0) as total_customer_paid,
    COALESCE(SUM(original_amount), 0) as total_original_amount
FROM subsidies.usage_transactions
WHERE customer_subsidy_id = $1;
