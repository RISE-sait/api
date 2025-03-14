-- name: CreateGame :one
INSERT INTO games (name)
VALUES ($1)
RETURNING *;

-- name: GetGameById :one
SELECT * FROM games WHERE id = $1;

-- name: GetGames :many
SELECT * FROM games
WHERE (name ILIKE '%' || @name || '%' OR @name IS NULL);

-- name: UpdateGame :one
UPDATE games
SET name = COALESCE(sqlc.narg('name'), name)
WHERE id = $1
RETURNING *;

-- name: DeleteGame :execrows
DELETE FROM games WHERE id = $1;