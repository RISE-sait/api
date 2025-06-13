-- name: CreateSystem :one
INSERT INTO playground.systems (name)
VALUES ($1)
RETURNING *;

-- name: GetSystems :many
SELECT * FROM playground.systems ORDER BY name;

-- name: UpdateSystem :one
UPDATE playground.systems
SET name = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2
RETURNING *;

-- name: DeleteSystem :execrows
DELETE FROM playground.systems WHERE id = $1;