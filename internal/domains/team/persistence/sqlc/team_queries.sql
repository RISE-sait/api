-- name: CreateTeam :one
INSERT INTO athletic.teams (name, capacity, coach_id, logo_url, is_external)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateExternalTeam :one
INSERT INTO athletic.teams (name, capacity, logo_url, is_external)
VALUES ($1, $2, $3, TRUE)
RETURNING *;

-- name: GetTeams :many
SELECT t.*,
       COALESCE(u.email, '') AS coach_email,
       COALESCE(u.first_name || ' ' || u.last_name, '') AS coach_name
FROM athletic.teams t
LEFT JOIN users.users u ON t.coach_id = u.id
ORDER BY t.is_external ASC, t.name ASC;

-- name: GetInternalTeams :many
SELECT t.*,
       u.email AS coach_email,
       (u.first_name || ' ' || u.last_name)::varchar AS coach_name
FROM athletic.teams t
JOIN users.users u ON t.coach_id = u.id
WHERE t.is_external = FALSE
ORDER BY t.name ASC;

-- name: GetExternalTeams :many
SELECT t.*
FROM athletic.teams t
WHERE t.is_external = TRUE
ORDER BY t.name ASC;

-- name: GetTeamsByCoach :many
SELECT t.*,
       u.email AS coach_email,
       (u.first_name || ' ' || u.last_name)::varchar AS coach_name
FROM athletic.teams t
JOIN users.users u ON t.coach_id = u.id
WHERE t.coach_id = $1 AND t.is_external = FALSE
ORDER BY t.name ASC;

-- name: GetTeamById :one
SELECT t.*,
       COALESCE(u.email, '') AS coach_email,
       COALESCE(u.first_name || ' ' || u.last_name, '') AS coach_name
FROM athletic.teams t
LEFT JOIN users.users u ON t.coach_id = u.id
WHERE t.id = $1;

-- name: SearchTeamsByName :many
SELECT t.*,
       COALESCE(u.email, '') AS coach_email,
       COALESCE(u.first_name || ' ' || u.last_name, '') AS coach_name
FROM athletic.teams t
LEFT JOIN users.users u ON t.coach_id = u.id
WHERE LOWER(t.name) LIKE LOWER($1)
ORDER BY t.is_external ASC, t.name ASC
LIMIT $2;

-- name: CheckTeamNameExists :one
SELECT EXISTS(
    SELECT 1 FROM athletic.teams
    WHERE LOWER(TRIM(name)) = LOWER(TRIM($1))
) AS exists;

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
       a.steals,
       a.photo_url
FROM athletic.teams t
         JOIN athletic.athletes a ON t.id = a.team_id
         JOIN users.users u ON a.id = u.id
WHERE t.id = $1;

-- name: UpdateAthleteTeam :execrows
UPDATE athletic.athletes
SET team_id = $1
WHERE id = $2;

-- name: UpdateTeam :one
UPDATE athletic.teams
SET name       = $1,
    coach_id   = $2,
    capacity   = $3,
    logo_url   = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $5 AND is_external = FALSE
RETURNING *;

-- name: UpdateExternalTeam :one
UPDATE athletic.teams
SET name       = $1,
    capacity   = $2,
    logo_url   = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $4 AND is_external = TRUE
RETURNING *;

-- name: DeleteTeam :execrows
DELETE
FROM athletic.teams
WHERE id = $1;