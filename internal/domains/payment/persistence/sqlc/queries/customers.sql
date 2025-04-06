-- name: IsCustomerExist :one
SELECT EXISTS(SELECT 1 FROM users.users WHERE id = $1);

-- name: GetCustomersTeam :one
SELECT t.*
FROM athletic.athletes a
         LEFT JOIN athletic.teams t ON a.team_id = t.id
WHERE a.id = sqlc.arg('customer_id');