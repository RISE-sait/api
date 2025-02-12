-- name: CreateFacility :one
WITH inserted_facility AS (
    INSERT INTO facilities (name, location, facility_type_id)
    VALUES ($1, $2, $3)
    RETURNING *
)
SELECT f.*, ft.name AS facility_type_name
FROM inserted_facility f
JOIN facility_types ft ON f.facility_type_id = ft.id;

-- name: GetFacilityById :one
SELECT f.*, ft.name as facility_type 
FROM facilities f JOIN facility_types ft ON f.facility_type_id = ft.id WHERE f.id = $1;

-- name: GetFacilities :many
SELECT f.*,  ft.name as facility_type 
FROM facilities f JOIN facility_types ft ON f.facility_type_id = ft.id
WHERE (f.name ILIKE '%' || @name || '%' OR @name IS NULL);

-- name: UpdateFacility :execrows
WITH updated as (
    UPDATE facilities f
    SET name = $1, location = $2, facility_type_id = $3
    WHERE f.id = $4
    RETURNING *
)
SELECT f.*, ft.name as facility_type_name
FROM updated f
JOIN facility_types ft ON f.facility_type_id = ft.id;

-- name: DeleteFacility :execrows
DELETE FROM facilities WHERE id = $1;