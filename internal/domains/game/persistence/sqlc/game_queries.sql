-- name: CreateGame :execrows
WITH program_insert AS (
    INSERT INTO program.programs (name, type)
    VALUES ($1, 'game')
    RETURNING id
)
INSERT INTO public.games (id, win_team, lose_team, win_score, lose_score)
VALUES ((SELECT id FROM program_insert), $2, $3, $4, $5);

-- name: GetGameById :one
SELECT p.*, wt.id as "win_team_id", lt.id as "lose_team_id",
wt.name as "win_team_name", wt.name as "lose_team_name", 
g.win_score, g.lose_score, p."type" FROM games g
JOIN athletic.teams wt ON g.win_team = wt.id
JOIN athletic.teams lt ON g.lose_team = lt.id
JOIN program.programs p ON g.id = p.id
WHERE g.id = $1;

-- name: GetGames :many
SELECT p.*, wt.id as "win_team_id", lt.id as "lose_team_id",
wt.name as "win_team_name", wt.name as "lose_team_name",
g.win_score, g.lose_score, p."type"
FROM games g
JOIN athletic.teams wt ON g.win_team = wt.id
JOIN athletic.teams lt ON g.lose_team = lt.id
JOIN program.programs p ON g.id = p.id;

-- name: UpdateGame :execrows
WITH program_update AS (
    UPDATE program.programs p
    SET name = $2
    WHERE p.id = $1
    RETURNING id
)
UPDATE games g
SET win_team = $3,
    lose_team = $4,
    win_score = $5,
    lose_score = $6
WHERE g.id = (SELECT id FROM program_update);

-- name: DeleteGame :execrows
DELETE FROM games WHERE id = $1;