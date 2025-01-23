-- name: CreateFacility :execrows
INSERT INTO facilities (name, location, facility_type_id)
VALUES ($1, $2, $3);

-- name: GetFacilityById :one
SELECT * FROM facilities WHERE id = $1;

-- name: GetAllFacilities :many
SELECT * FROM facilities;

-- name: UpdateFacility :execrows
UPDATE facilities
SET name = $1, location = $2, facility_type_id = $3
WHERE id = $4;

-- name: DeleteFacility :execrows
DELETE FROM facilities WHERE id = $1;