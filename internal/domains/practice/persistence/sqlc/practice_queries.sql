-- name: CreatePractice :exec
INSERT INTO practice.practices (
    team_id, start_time, end_time, court_id, location_id, status
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: UpdatePractice :exec
UPDATE practice.practices
SET team_id=$1,
    start_time=$2,
    end_time=$3,
    court_id=$4,
    location_id=$5,
    status=$6,
    updated_at=now()
WHERE id=$7;

-- name: DeletePractice :exec
DELETE FROM practice.practices WHERE id=$1;

-- name: GetPracticeByID :one
SELECT p.id,
       p.team_id,
       t.name AS team_name,
       t.logo_url AS team_logo_url,
       p.start_time,
       p.end_time,
       p.location_id,
       l.name AS location_name,
       p.court_id,
       c.name AS court_name,
       p.status,
       p.created_at,
       p.updated_at
FROM practice.practices p
         JOIN athletic.teams t ON p.team_id = t.id
         JOIN location.locations l ON p.location_id = l.id
         JOIN location.courts c ON p.court_id = c.id
WHERE p.id = $1;

-- name: ListPractices :many
SELECT p.id,
       p.team_id,
       t.name AS team_name,
       t.logo_url AS team_logo_url,
       p.start_time,
       p.end_time,
       p.location_id,
       l.name AS location_name,
       p.court_id,
       c.name AS court_name,
       p.status,
       p.created_at,
       p.updated_at
FROM practice.practices p
         JOIN athletic.teams t ON p.team_id = t.id
         JOIN location.locations l ON p.location_id = l.id
         JOIN location.courts c ON p.court_id = c.id
WHERE (
    sqlc.narg('team_id')::uuid IS NULL
    OR p.team_id = sqlc.narg('team_id')::uuid
)
ORDER BY p.start_time ASC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');