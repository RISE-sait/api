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