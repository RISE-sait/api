-- name: CreatePaymentTransaction :one
INSERT INTO payments.payment_transactions (
    customer_id,
    customer_email,
    customer_name,
    transaction_type,
    transaction_date,
    original_amount,
    discount_amount,
    subsidy_amount,
    customer_paid,
    membership_plan_id,
    program_id,
    event_id,
    credit_package_id,
    subsidy_id,
    discount_code_id,
    stripe_customer_id,
    stripe_subscription_id,
    stripe_invoice_id,
    stripe_payment_intent_id,
    stripe_checkout_session_id,
    payment_status,
    payment_method,
    currency,
    description,
    metadata,
    receipt_url,
    invoice_url,
    invoice_pdf_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28
) RETURNING *;

-- name: GetPaymentTransaction :one
SELECT * FROM payments.payment_transactions
WHERE id = $1;

-- name: GetPaymentTransactionByStripeInvoice :one
SELECT * FROM payments.payment_transactions
WHERE stripe_invoice_id = $1
LIMIT 1;

-- name: GetPaymentTransactionByStripeSubscription :one
SELECT * FROM payments.payment_transactions
WHERE stripe_subscription_id = $1
ORDER BY transaction_date DESC
LIMIT 1;

-- name: GetPaymentTransactionByStripeCheckoutSession :one
SELECT * FROM payments.payment_transactions
WHERE stripe_checkout_session_id = $1
LIMIT 1;

-- name: ListPaymentTransactionsByCustomer :many
SELECT * FROM payments.payment_transactions
WHERE customer_id = $1
ORDER BY transaction_date DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPaymentTransactionsByCustomer :one
SELECT COUNT(*) FROM payments.payment_transactions
WHERE customer_id = $1;

-- name: ListPaymentTransactions :many
SELECT * FROM payments.payment_transactions
WHERE
    (sqlc.narg('customer_id')::uuid IS NULL OR customer_id = sqlc.narg('customer_id')) AND
    (sqlc.narg('transaction_type')::text IS NULL OR transaction_type = sqlc.narg('transaction_type')) AND
    (sqlc.narg('payment_status')::text IS NULL OR payment_status = sqlc.narg('payment_status')) AND
    (sqlc.narg('start_date')::timestamptz IS NULL OR transaction_date >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR transaction_date <= sqlc.narg('end_date')) AND
    (sqlc.narg('subsidy_id')::uuid IS NULL OR subsidy_id = sqlc.narg('subsidy_id'))
ORDER BY transaction_date DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPaymentTransactions :one
SELECT COUNT(*) FROM payments.payment_transactions
WHERE
    (sqlc.narg('customer_id')::uuid IS NULL OR customer_id = sqlc.narg('customer_id')) AND
    (sqlc.narg('transaction_type')::text IS NULL OR transaction_type = sqlc.narg('transaction_type')) AND
    (sqlc.narg('payment_status')::text IS NULL OR payment_status = sqlc.narg('payment_status')) AND
    (sqlc.narg('start_date')::timestamptz IS NULL OR transaction_date >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR transaction_date <= sqlc.narg('end_date')) AND
    (sqlc.narg('subsidy_id')::uuid IS NULL OR subsidy_id = sqlc.narg('subsidy_id'));

-- name: UpdatePaymentStatus :one
UPDATE payments.payment_transactions
SET payment_status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: RecordRefund :one
UPDATE payments.payment_transactions
SET payment_status = $2,
    refunded_amount = $3,
    refund_reason = $4,
    refunded_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetPaymentSummary :one
SELECT
    COUNT(*) as total_transactions,
    COALESCE(SUM(original_amount), 0) as total_original_amount,
    COALESCE(SUM(discount_amount), 0) as total_discount_amount,
    COALESCE(SUM(subsidy_amount), 0) as total_subsidy_amount,
    COALESCE(SUM(customer_paid), 0) as total_customer_paid,
    COALESCE(SUM(refunded_amount), 0) as total_refunded_amount
FROM payments.payment_transactions
WHERE
    (sqlc.narg('start_date')::timestamptz IS NULL OR transaction_date >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR transaction_date <= sqlc.narg('end_date')) AND
    (sqlc.narg('transaction_type')::text IS NULL OR transaction_type = sqlc.narg('transaction_type')) AND
    (sqlc.narg('payment_status')::text IS NULL OR payment_status = sqlc.narg('payment_status'));

-- name: GetPaymentSummaryByType :many
SELECT
    transaction_type,
    COUNT(*) as transaction_count,
    COALESCE(SUM(original_amount), 0) as total_original_amount,
    COALESCE(SUM(discount_amount), 0) as total_discount_amount,
    COALESCE(SUM(subsidy_amount), 0) as total_subsidy_amount,
    COALESCE(SUM(customer_paid), 0) as total_customer_paid
FROM payments.payment_transactions
WHERE
    (sqlc.narg('start_date')::timestamptz IS NULL OR transaction_date >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR transaction_date <= sqlc.narg('end_date')) AND
    payment_status = 'completed'
GROUP BY transaction_type
ORDER BY total_customer_paid DESC;

-- name: GetSubsidyUsageSummary :one
SELECT
    COUNT(*) as transactions_with_subsidy,
    COALESCE(SUM(subsidy_amount), 0) as total_subsidy_used
FROM payments.payment_transactions
WHERE
    subsidy_amount > 0 AND
    payment_status = 'completed' AND
    (sqlc.narg('start_date')::timestamptz IS NULL OR transaction_date >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR transaction_date <= sqlc.narg('end_date'));

-- name: ExportPaymentTransactions :many
SELECT
    id,
    customer_id,
    customer_email,
    customer_name,
    transaction_type,
    transaction_date,
    original_amount,
    discount_amount,
    subsidy_amount,
    customer_paid,
    stripe_invoice_id,
    payment_status,
    payment_method,
    currency,
    description,
    receipt_url,
    invoice_url,
    invoice_pdf_url,
    created_at
FROM payments.payment_transactions
WHERE
    (sqlc.narg('start_date')::timestamptz IS NULL OR transaction_date >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR transaction_date <= sqlc.narg('end_date')) AND
    (sqlc.narg('transaction_type')::text IS NULL OR transaction_type = sqlc.narg('transaction_type')) AND
    (sqlc.narg('payment_status')::text IS NULL OR payment_status = sqlc.narg('payment_status'))
ORDER BY transaction_date DESC;

-- name: UpdatePaymentUrls :exec
UPDATE payments.payment_transactions
SET receipt_url = COALESCE(sqlc.narg('receipt_url'), receipt_url),
    invoice_url = COALESCE(sqlc.narg('invoice_url'), invoice_url),
    invoice_pdf_url = COALESCE(sqlc.narg('invoice_pdf_url'), invoice_pdf_url),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetTransactionsForBackfill :many
SELECT id, stripe_checkout_session_id, stripe_payment_intent_id, stripe_invoice_id
FROM payments.payment_transactions
WHERE (stripe_checkout_session_id IS NOT NULL OR stripe_payment_intent_id IS NOT NULL OR stripe_invoice_id IS NOT NULL)
  AND receipt_url IS NULL
  AND invoice_url IS NULL;
