-- name: CreateTeam :one
INSERT INTO athletic.teams (name, capacity, coach_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeams :many
SELECT t.*, u.email AS coach_email, (u.first_name || ' ' || u.last_name)::varchar AS coach_name
FROM athletic.teams t
         JOIN users.users u ON t.coach_id = u.id;

-- name: GetTeamById :one
SELECT t.*, u.email AS coach_email, (u.first_name || ' ' || u.last_name)::varchar AS coach_name
FROM athletic.teams t
         JOIN users.users u ON t.coach_id = u.id
WHERE t.id = $1;

-- name: GetTeamRoster :many
SELECT u.id,
       u.email,
       u.country_alpha2_code,
       (u.first_name || ' ' || u.last_name)::varchar AS name,
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals
FROM athletic.teams t
         JOIN athletic.athletes a ON t.id = a.team_id
         JOIN users.users u ON a.id = u.id
WHERE t.id = $1;

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