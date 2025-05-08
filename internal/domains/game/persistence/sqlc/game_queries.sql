-- name: CreateGame :exec
INSERT INTO game.games (
  id, home_team_id, away_team_id, home_score, away_score, start_time,
  end_time, location_id, status
) VALUES (
  gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetGameById :one
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
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.locations loc ON g.location_id = loc.id
WHERE g.id = $1;

-- name: GetGames :many
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
    g.status,
    g.created_at,
    g.updated_at
FROM game.games g
JOIN athletic.teams ht ON g.home_team_id = ht.id
JOIN athletic.teams at ON g.away_team_id = at.id
JOIN location.locations loc ON g.location_id = loc.id
ORDER BY g.start_time DESC;

-- name: UpdateGame :execrows
UPDATE game.games
SET home_score = $2,
    away_score = $3,
    start_time = $4,
    end_time = $5,
    location_id = $6,
    status = $7,
    updated_at = now()
WHERE id = $1;

-- name: DeleteGame :execrows
DELETE FROM game.games
WHERE id = $1;
