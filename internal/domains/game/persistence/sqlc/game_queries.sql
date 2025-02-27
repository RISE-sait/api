-- name: CreateGame :one
INSERT INTO games (name, video_link)
VALUES ($1, $2)
RETURNING *;

-- name: GetGameById :one
SELECT * FROM games WHERE id = $1;

-- name: GetGames :many
SELECT * FROM games
WHERE (name ILIKE '%' || @name || '%' OR @name IS NULL)
LIMIT $1;

-- name: UpdateGame :one
UPDATE games
SET name = COALESCE(sqlc.narg('name'), name),
    video_link = COALESCE(sqlc.narg('video_link'), video_link)
WHERE id = $1
RETURNING *;

-- name: DeleteGame :execrows
DELETE FROM games WHERE id = $1;