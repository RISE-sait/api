-- name: CreateFacilityCategory :one
INSERT INTO location.facility_categories (name) VALUES ($1)
RETURNING *;

-- name: GetFacilityCategoryById :one
SELECT name FROM location.facility_categories WHERE id = $1;

-- name: GetFacilityCategories :many
SELECT * from location.facility_categories;

-- name: UpdateFacilityCategory :one
UPDATE location.facility_categories
SET name = $1
WHERE id = $2
RETURNING *;

-- name: DeleteFacilityCategory :execrows
DELETE FROM location.facility_categories WHERE id = $1;