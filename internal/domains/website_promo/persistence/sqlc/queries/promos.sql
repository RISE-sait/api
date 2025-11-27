-- name: CreateHeroPromo :one
INSERT INTO website.hero_promos (
    title, subtitle, description, image_url, button_text, button_link,
    display_order, duration_seconds, is_active, start_date, end_date,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12
) RETURNING *;

-- name: GetHeroPromoById :one
SELECT * FROM website.hero_promos WHERE id = $1;

-- name: GetAllHeroPromos :many
SELECT * FROM website.hero_promos ORDER BY display_order ASC, created_at DESC;

-- name: GetActiveHeroPromos :many
SELECT * FROM website.hero_promos
WHERE is_active = true
  AND (start_date IS NULL OR start_date <= CURRENT_TIMESTAMP)
  AND (end_date IS NULL OR end_date >= CURRENT_TIMESTAMP)
ORDER BY display_order ASC, created_at DESC;

-- name: UpdateHeroPromo :one
UPDATE website.hero_promos SET
    title = $2,
    subtitle = $3,
    description = $4,
    image_url = $5,
    button_text = $6,
    button_link = $7,
    display_order = $8,
    duration_seconds = $9,
    is_active = $10,
    start_date = $11,
    end_date = $12,
    updated_at = CURRENT_TIMESTAMP,
    updated_by = $13
WHERE id = $1
RETURNING *;

-- name: DeleteHeroPromo :execrows
DELETE FROM website.hero_promos WHERE id = $1;

-- name: CreateFeatureCard :one
INSERT INTO website.feature_cards (
    title, description, image_url, button_text, button_link,
    display_order, is_active, start_date, end_date,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10
) RETURNING *;

-- name: GetFeatureCardById :one
SELECT * FROM website.feature_cards WHERE id = $1;

-- name: GetAllFeatureCards :many
SELECT * FROM website.feature_cards ORDER BY display_order ASC, created_at DESC;

-- name: GetActiveFeatureCards :many
SELECT * FROM website.feature_cards
WHERE is_active = true
  AND (start_date IS NULL OR start_date <= CURRENT_TIMESTAMP)
  AND (end_date IS NULL OR end_date >= CURRENT_TIMESTAMP)
ORDER BY display_order ASC, created_at DESC;

-- name: UpdateFeatureCard :one
UPDATE website.feature_cards SET
    title = $2,
    description = $3,
    image_url = $4,
    button_text = $5,
    button_link = $6,
    display_order = $7,
    is_active = $8,
    start_date = $9,
    end_date = $10,
    updated_at = CURRENT_TIMESTAMP,
    updated_by = $11
WHERE id = $1
RETURNING *;

-- name: DeleteFeatureCard :execrows
DELETE FROM website.feature_cards WHERE id = $1;
