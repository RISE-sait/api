-- name: CreateFacilityType :execrows
INSERT INTO facility_types (name) VALUES ($1);

-- name: GetFacilityTypeById :one
SELECT name FROM facility_types WHERE id = $1;

-- name: GetAllFacilityTypes :many
SELECT * from facility_types;

-- name: UpdateFacilityType :execrows
UPDATE facility_types
SET name = $1
WHERE id = $2;

-- name: DeleteFacilityType :execrows
DELETE FROM facility_types WHERE id = $1;