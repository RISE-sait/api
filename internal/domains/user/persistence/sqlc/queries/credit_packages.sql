-- name: GetCreditPackageByID :one
SELECT * FROM users.credit_packages
WHERE id = $1;

-- name: GetCreditPackageByStripePriceID :one
SELECT * FROM users.credit_packages
WHERE stripe_price_id = $1;

-- name: GetAllCreditPackages :many
SELECT * FROM users.credit_packages
ORDER BY credit_allocation ASC;

-- name: CreateCreditPackage :one
INSERT INTO users.credit_packages (
    name,
    description,
    stripe_price_id,
    credit_allocation,
    weekly_credit_limit
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: UpdateCreditPackage :one
UPDATE users.credit_packages
SET name = $2,
    description = $3,
    stripe_price_id = $4,
    credit_allocation = $5,
    weekly_credit_limit = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteCreditPackage :exec
DELETE FROM users.credit_packages
WHERE id = $1;

-- name: GetCustomerActiveCreditPackage :one
SELECT cacp.*, cp.name as package_name, cp.credit_allocation
FROM users.customer_active_credit_package cacp
JOIN users.credit_packages cp ON cacp.credit_package_id = cp.id
WHERE cacp.customer_id = $1;

-- name: SetCustomerActiveCreditPackage :exec
INSERT INTO users.customer_active_credit_package (
    customer_id,
    credit_package_id,
    weekly_credit_limit,
    purchased_at
) VALUES (
    $1, $2, $3, CURRENT_TIMESTAMP
)
ON CONFLICT (customer_id) DO UPDATE SET
    credit_package_id = $2,
    weekly_credit_limit = $3,
    purchased_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteCustomerActiveCreditPackage :exec
DELETE FROM users.customer_active_credit_package
WHERE customer_id = $1;
