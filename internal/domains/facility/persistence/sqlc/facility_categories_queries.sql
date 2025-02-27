-- name: CreateFacilityCategory :one
INSERT INTO facility.facility_categories (name) VALUES ($1)
RETURNING *;

-- name: GetFacilityCategoryById :one
SELECT name FROM facility.facility_categories WHERE id = $1;

-- name: GetFacilityCategories :many
SELECT * from facility.facility_categories;

-- name: UpdateFacilityCategory :one
UPDATE facility.facility_categories
SET name = $1
WHERE id = $2
RETURNING *;

-- name: DeleteFacilityCategory :execrows
DELETE FROM facility.facility_categories WHERE id = $1;