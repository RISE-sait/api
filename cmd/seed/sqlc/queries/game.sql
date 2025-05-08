-- name: InsertGames :many
WITH games_data AS (
    SELECT 
        unnest(@start_times::timestamptz[]) AS start_time,
        unnest(@end_times::timestamptz[])   AS end_time,
        unnest(@home_team_ids::uuid[])      AS home_team_id,
        unnest(@away_team_ids::uuid[])      AS away_team_id,
        unnest(@location_names::text[])     AS location_name,
        unnest(@home_scores::int[])         AS home_score,
        unnest(@away_scores::int[])         AS away_score,
        unnest(@statuses::text[])           AS status
)
INSERT INTO game.games (
    id, home_team_id, away_team_id, home_score, away_score, start_time, end_time, location_id, status
)
SELECT 
    gen_random_uuid(),
    g.home_team_id,
    g.away_team_id,
    g.home_score,
    g.away_score,
    g.start_time,
    g.end_time,
    loc.id,
    g.status
FROM games_data g
JOIN location.locations loc ON loc.name = g.location_name
RETURNING id;

-- name: CreateGame :one
INSERT INTO game.games (
  id, home_team_id, away_team_id, home_score, away_score,
  start_time, end_time, location_id, status
)
VALUES (
  gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;
