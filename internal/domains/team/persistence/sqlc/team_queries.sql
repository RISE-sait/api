-- name: CreateTeam :one
INSERT INTO athletic.teams (name, capacity, coach_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeams :many
SELECT *
FROM athletic.teams;

-- name: GetTeamById :one
SELECT *
FROM athletic.teams
WHERE id = $1;

-- name: UpdateTeam :one
UPDATE athletic.teams
SET name       = $1,
    coach_id   = $2,
    capacity   = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $4
RETURNING *;

-- name: DeleteTeam :execrows
DELETE
FROM athletic.teams
WHERE id = $1;