-- name: CreateCourt :one
INSERT INTO location.courts (location_id, name)
VALUES ($1, $2)
RETURNING id, location_id, name;

-- name: GetCourtById :one
SELECT id, location_id, name
FROM location.courts
WHERE id = $1;

-- name: GetCourts :many
SELECT id, location_id, name
FROM location.courts;

-- name: UpdateCourt :execrows
UPDATE location.courts
SET location_id = $1,
    name = $2
WHERE id = $3;

-- name: DeleteCourt :execrows
DELETE FROM location.courts
WHERE id = $1;