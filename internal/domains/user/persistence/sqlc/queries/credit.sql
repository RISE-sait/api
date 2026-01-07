-- Customer Credits Queries

-- name: GetCustomerCredits :one
-- Get customer's current credit balance
SELECT credits
FROM users.customer_credits
WHERE customer_id = $1;

-- name: CreateCustomerCredits :exec
-- Create customer credits record with initial balance
INSERT INTO users.customer_credits (customer_id, credits)
VALUES ($1, $2)
ON CONFLICT (customer_id) DO NOTHING;

-- name: UpdateCustomerCredits :execrows
-- Update customer's credit balance directly
UPDATE users.customer_credits
SET credits = $2
WHERE customer_id = $1;

-- name: DeductCredits :execrows  
-- Deduct credits only if customer has sufficient balance
UPDATE users.customer_credits
SET credits = credits - $2
WHERE customer_id = $1 AND credits >= $2;

-- name: RefundCredits :execrows
-- Add credits back to customer's account
UPDATE users.customer_credits
SET credits = credits + $2
WHERE customer_id = $1;

-- name: CheckCustomerHasSufficientCredits :one
-- Check if customer has enough credits for a transaction
SELECT credits >= $2 as has_sufficient
FROM users.customer_credits
WHERE customer_id = $1;

-- Credit Transaction Queries

-- name: LogCreditTransaction :exec
-- Log a credit transaction for audit trail
INSERT INTO users.credit_transactions (customer_id, amount, transaction_type, event_id, description)
VALUES ($1, $2, $3, $4, $5);

-- name: GetCustomerCreditTransactions :many
-- Get customer's credit transaction history with pagination
SELECT id, customer_id, amount, transaction_type, event_id, description, created_at
FROM users.credit_transactions
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetEventCreditTransactions :many
-- Get all credit transactions for a specific event
SELECT ct.id, ct.customer_id, ct.amount, ct.transaction_type, ct.description, ct.created_at,
       u.first_name, u.last_name, u.email
FROM users.credit_transactions ct
JOIN users.users u ON ct.customer_id = u.id
WHERE ct.event_id = $1
ORDER BY ct.created_at DESC;

-- Event Credit Cost Queries

-- name: GetEventCreditCost :one
-- Get the credit cost for a specific event
SELECT credit_cost
FROM events.events
WHERE id = $1;

-- name: UpdateEventCreditCost :execrows
-- Update the credit cost for an event
UPDATE events.events
SET credit_cost = $2
WHERE id = $1;

-- Weekly Credit Usage Queries

-- name: GetWeeklyCreditsUsed :one
-- Get customer's credit usage for the current week
SELECT COALESCE(credits_used, 0) as credits_used
FROM users.weekly_credit_usage
WHERE customer_id = $1 
  AND week_start_date = $2;

-- name: UpdateWeeklyCreditsUsed :exec
-- Update (or insert) weekly credit usage for a customer
INSERT INTO users.weekly_credit_usage (customer_id, week_start_date, credits_used, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (customer_id, week_start_date)
DO UPDATE SET
    credits_used = users.weekly_credit_usage.credits_used + EXCLUDED.credits_used,
    updated_at = NOW();

-- name: ReduceWeeklyCreditsUsed :exec
-- Reduce weekly credit usage when credits are refunded (ensures it doesn't go below 0)
UPDATE users.weekly_credit_usage
SET credits_used = GREATEST(credits_used - $3, 0),
    updated_at = NOW()
WHERE customer_id = $1 AND week_start_date = $2;

-- name: GetActiveCustomerMembershipPlanID :one
-- Get customer's active membership plan ID
SELECT membership_plan_id
FROM users.customer_membership_plans
WHERE customer_id = $1 
  AND status = 'active'
ORDER BY created_at DESC
LIMIT 1;

-- name: GetCustomerMembershipPlan :one
-- Get customer's current membership plan with credit info
SELECT mp.credit_allocation, mp.weekly_credit_limit
FROM users.customer_membership_plans cmp
JOIN membership.membership_plans mp ON cmp.membership_plan_id = mp.id
WHERE cmp.customer_id = $1 
  AND cmp.status = 'active'
ORDER BY cmp.created_at DESC
LIMIT 1;

-- name: IsCustomerEnrolledInEvent :one
-- Check if customer is already enrolled in an event (to prevent duplicate credit payments)
SELECT EXISTS(
    SELECT 1 FROM events.customer_enrollment
    WHERE event_id = $1
    AND customer_id = $2
    AND payment_status IN ('pending', 'paid')
) as is_enrolled;

-- name: CheckWeeklyCreditLimit :one
-- Check if customer can use specified credits without exceeding weekly limit
-- Prioritizes active credit package over membership plan
SELECT
    COALESCE(wcu.credits_used, 0) as current_usage,
    COALESCE(cacp.weekly_credit_limit, mp.weekly_credit_limit, 0) as weekly_credit_limit,
    CASE
        -- First check if they have an active credit package
        WHEN cacp.weekly_credit_limit IS NOT NULL THEN
            CASE
                WHEN cacp.weekly_credit_limit = 0 THEN true  -- Unlimited credits
                WHEN COALESCE(wcu.credits_used, 0) + $3 <= cacp.weekly_credit_limit THEN true
                ELSE false
            END
        -- Otherwise check membership plan
        WHEN mp.weekly_credit_limit IS NOT NULL THEN
            CASE
                WHEN mp.weekly_credit_limit = 0 THEN true  -- Unlimited credits
                WHEN COALESCE(wcu.credits_used, 0) + $3 <= mp.weekly_credit_limit THEN true
                ELSE false
            END
        -- No active package or plan - allow usage (no limit)
        ELSE true
    END as can_use_credits
FROM (SELECT 1) AS dummy  -- Dummy table to ensure we always get a row
LEFT JOIN users.customer_active_credit_package cacp ON cacp.customer_id = $1
LEFT JOIN users.customer_membership_plans cmp ON (cmp.customer_id = $1 AND cmp.status = 'active')
LEFT JOIN membership.membership_plans mp ON cmp.membership_plan_id = mp.id
LEFT JOIN users.weekly_credit_usage wcu ON (
    wcu.customer_id = $1 AND
    wcu.week_start_date = $2
)
ORDER BY cmp.created_at DESC
LIMIT 1;

-- Credit Refund Queries

-- name: GetEnrollmentCreditTransaction :one
-- Get the credit transaction for a customer's event enrollment (to determine refund amount)
SELECT id, amount, created_at
FROM users.credit_transactions
WHERE customer_id = $1
  AND event_id = $2
  AND transaction_type = 'enrollment'
ORDER BY created_at DESC
LIMIT 1;

-- name: LogCreditRefundAudit :exec
-- Log a credit refund to the audit table with full event context
INSERT INTO audit.credit_refund_logs (
    customer_id, event_id, performed_by, credits_refunded,
    event_name, event_start_at, program_name, location_name,
    staff_role, reason, ip_address
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: GetCreditRefundLogs :many
-- Get credit refund audit logs with customer and staff names
SELECT
    crl.id,
    crl.customer_id,
    crl.event_id,
    crl.performed_by,
    crl.credits_refunded,
    crl.event_name,
    crl.event_start_at,
    crl.program_name,
    crl.location_name,
    crl.staff_role,
    crl.reason,
    crl.ip_address,
    crl.created_at,
    cu.first_name as customer_first_name,
    cu.last_name as customer_last_name,
    su.first_name as staff_first_name,
    su.last_name as staff_last_name
FROM audit.credit_refund_logs crl
JOIN users.users cu ON crl.customer_id = cu.id
JOIN users.users su ON crl.performed_by = su.id
WHERE (sqlc.narg(customer_id)::uuid IS NULL OR crl.customer_id = sqlc.narg(customer_id))
  AND (sqlc.narg(event_id)::uuid IS NULL OR crl.event_id = sqlc.narg(event_id))
ORDER BY crl.created_at DESC
LIMIT $1 OFFSET $2;