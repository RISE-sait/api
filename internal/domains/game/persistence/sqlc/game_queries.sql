-- The following SQL functions provide full CRUD support for the game.games table.
-- This structure replaces the older approach of inserting game data via a WITH clause
-- that unwrapped parallel arrays (e.g., unnesting start_times, team_names, etc.).
-- 
-- - The new design promotes single-row transactional inserts, which are safer and easier to debug.
-- - Complex batch insertion with unnested arrays was moved into Go, giving more control over data preparation.
-- - This also simplifies SQL and avoids silent failures during multi-row joins.

-- name: CreateGame :exec
-- Inserts a single game into the game.games table using direct parameters.
INSERT INTO game.games (
  id, home_team_id, away_team_id, home_score, away_score, start_time,
  end_time,court_id, location_id, status
) VALUES (
  gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetGameById :one
-- Retrieves a specific game along with team names and location name.
SELECT 
    g.id,
    g.home_team_id,
    ht.name AS home_team_name,
    g.away_team_id,
    at.name AS away_team_name,
    g.home_score,
    g.away_score,
    g.start_time,
    g.end_time,
    g.location_id,
    loc.name AS location_name,
    g.court_id,
    c.name AS court_name,
    g.court_id,
    c.name AS court_name,
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.courts c ON g.court_id = c.id
JOIN location.locations loc ON g.location_id = loc.id
WHERE g.id = $1;

-- name: GetGames :many
-- Retrieves all games, with team and location names.
SELECT 
    g.id,
    g.home_team_id,
    ht.name AS home_team_name,
    ht.logo_url AS home_team_logo_url,
    g.away_team_id,
    at.name AS away_team_name,
    at.logo_url AS away_team_logo_url,
    g.home_score,
    g.away_score,
    g.start_time,
    g.end_time,
    g.location_id,
    loc.name AS location_name,
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN location.courts c ON g.court_id = c.id
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.locations loc ON g.location_id = loc.id
ORDER BY g.start_time ASC
LIMIT $1 OFFSET $2;

-- name: UpdateGame :execrows
-- Updates an existing game's scores, times, location, and status.
UPDATE game.games
SET home_score = $2,
    away_score = $3,
    start_time = $4,
    end_time = $5,
    location_id = $6,
    court_id = $7,
    status = $8,
    updated_at = now()
WHERE id = $1;

-- name: DeleteGame :execrows
-- Deletes a game by ID.
DELETE FROM game.games
WHERE id = $1;

-- name: GetUpcomingGames :many
-- Retrieves games that are upcoming and ongoing.
-- This includes games that have started but not yet ended.
SELECT 
    g.id,
    g.home_team_id,
    ht.name AS home_team_name,
    ht.logo_url AS home_team_logo_url,
    g.away_team_id,
    at.name AS away_team_name,
    at.logo_url AS away_team_logo_url,
    g.home_score,
    g.away_score,
    g.start_time,
    g.end_time,
    g.location_id,
    loc.name AS location_name,
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.locations loc ON g.location_id = loc.id
WHERE g.end_time >= NOW()
ORDER BY g.start_time ASC
LIMIT $1 OFFSET $2;

-- name: GetPastGames :many
-- Retrieves games that have already completed.
SELECT 
    g.id,
    g.home_team_id,
    ht.name AS home_team_name,
    ht.logo_url AS home_team_logo_url,
    g.away_team_id,
    at.name AS away_team_name,
    at.logo_url AS away_team_logo_url,
    g.home_score,
    g.away_score,
    g.start_time,
    g.end_time,
    g.location_id,
    loc.name AS location_name,
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.locations loc ON g.location_id = loc.id
WHERE g.start_time < NOW()
ORDER BY g.start_time DESC
LIMIT $1 OFFSET $2;

-- name: GetGamesByTeams :many
SELECT
    g.id,
    g.home_team_id,
    ht.name AS home_team_name,
    ht.logo_url AS home_team_logo_url,
    g.away_team_id,
    at.name AS away_team_name,
    at.logo_url AS away_team_logo_url,
    g.home_score,
    g.away_score,
    g.start_time,
    g.end_time,
    g.location_id,
    loc.name AS location_name,
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.locations loc ON g.location_id = loc.id
WHERE g.home_team_id = ANY(sqlc.arg('team_ids')::uuid[])
   OR g.away_team_id = ANY(sqlc.arg('team_ids')::uuid[])
ORDER BY g.start_time ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');