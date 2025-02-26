-- name: CreateFacility :one
WITH inserted_facility AS (
    INSERT INTO facilities (name, address, facility_category_id)
    VALUES ($1, $2, $3)
    RETURNING *
)
SELECT f.*, fc.name AS facility_category_name
FROM inserted_facility f
JOIN facility_categories fc ON f.facility_category_id = fc.id;

-- name: GetFacilityById :one
SELECT f.*, fc.name as facility_category_name
FROM facilities f JOIN facility_categories fc ON f.facility_category_id = fc.id WHERE f.id = $1;

-- name: GetFacilities :many
SELECT f.*,  fc.name as facility_category_name
FROM facilities f JOIN facility_categories fc ON f.facility_category_id = fc.id
WHERE (f.name ILIKE '%' || @facility_name || '%' OR @facility_name IS NULL);

-- name: UpdateFacility :execrows
WITH updated as (
    UPDATE facilities f
    SET name = $1, address = $2, facility_category_id = $3
    WHERE f.id = $4
    RETURNING *
)
SELECT f.*, fc.name as facility_category_name
FROM updated f
JOIN facility_categories fc ON f.facility_category_id = fc.id;

-- name: DeleteFacility :execrows
DELETE FROM facilities WHERE id = $1;