-- name: CreateLocation :one
INSERT INTO location.locations (name, address)
VALUES ($1, $2)
RETURNING *;

-- name: GetLocationById :one
SELECT *
from location.locations
WHERE id = $1;

-- name: GetLocations :many
SELECT *
from location.locations;

-- name: UpdateLocation :execrows
UPDATE location.locations
SET name    = $1,
    address = $2
WHERE id = $3
;

-- name: DeleteLocation :execrows
DELETE
FROM location.locations
WHERE id = $1;