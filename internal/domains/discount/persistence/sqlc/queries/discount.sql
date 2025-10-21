-- name: CreateDiscount :one
INSERT INTO discounts (
    name, description, discount_percent, discount_amount, discount_type,
    is_use_unlimited, use_per_client, is_active, valid_from, valid_to,
    duration_type, duration_months, applies_to, max_redemptions, stripe_coupon_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
) RETURNING *;

-- name: GetDiscountById :one
SELECT * FROM discounts WHERE id = $1;

-- name: GetDiscountByName :one
SELECT * FROM discounts WHERE name = $1;

-- name: GetDiscountByNameActive :one
SELECT * FROM discounts
WHERE name = $1
  AND is_active = true
  AND valid_from <= now()
  AND (valid_to IS NULL OR valid_to >= now());

-- name: ListDiscounts :many
SELECT * FROM discounts ORDER BY created_at DESC;

-- name: UpdateDiscount :one
UPDATE discounts
SET name = $1,
    description = $2,
    discount_percent = $3,
    discount_amount = $4,
    discount_type = $5,
    is_use_unlimited = $6,
    use_per_client = $7,
    is_active = $8,
    valid_from = $9,
    valid_to = $10,
    duration_type = $11,
    duration_months = $12,
    applies_to = $13,
    max_redemptions = $14,
    stripe_coupon_id = $15,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $16
RETURNING *;

-- name: DeleteDiscount :execrows
DELETE FROM discounts WHERE id = $1;

-- name: GetUsageCount :one
SELECT usage_count FROM users.customer_discount_usage WHERE customer_id = $1 AND discount_id = $2;

-- name: IncrementUsage :execrows
INSERT INTO users.customer_discount_usage (customer_id, discount_id)
VALUES ($1, $2)
ON CONFLICT (customer_id, discount_id) DO UPDATE
SET usage_count = users.customer_discount_usage.usage_count + 1,
    last_used_at = CURRENT_TIMESTAMP;

-- name: GetRestrictedPlans :many
SELECT membership_plan_id FROM membership.discount_restricted_membership_plans
WHERE discount_id = $1;