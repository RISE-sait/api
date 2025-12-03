-- name: CreateCourt :one
INSERT INTO location.courts (location_id, name)
VALUES ($1, $2)
RETURNING id, location_id, name;

-- name: GetCourtById :one
SELECT c.id, c.location_id, c.name, l.name as location_name
FROM location.courts c
JOIN location.locations l ON c.location_id = l.id
WHERE c.id = $1;

-- name: GetCourts :many
SELECT c.id, c.location_id, c.name, l.name as location_name
FROM location.courts c
JOIN location.locations l ON c.location_id = l.id;

-- name: UpdateCourt :execrows
UPDATE location.courts
SET location_id = $1,
    name = $2
WHERE id = $3;

-- name: DeleteCourt :execrows
DELETE FROM location.courts
WHERE id = $1;