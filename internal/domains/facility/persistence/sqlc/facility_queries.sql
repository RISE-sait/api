-- name: CreateFacility :execrows
INSERT INTO facilities (name, location, facility_type_id)
VALUES ($1, $2, $3);

-- name: GetFacilityById :one
SELECT f.id, f.name, f.location, ft.name as facility_type FROM facilities f JOIN facility_types ft ON f.facility_type_id = ft.id WHERE f.id = $1;

-- name: GetAllFacilities :many
SELECT f.id, f.name, f.location, ft.name  as facility_type FROM facilities f JOIN facility_types ft ON f.facility_type_id = ft.id;

-- name: UpdateFacility :execrows
UPDATE facilities
SET name = $1, location = $2, facility_type_id = $3
WHERE id = $4;

-- name: DeleteFacility :execrows
DELETE FROM facilities WHERE id = $1;