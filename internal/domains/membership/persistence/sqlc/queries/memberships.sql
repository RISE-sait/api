-- name: CreateMembership :one
INSERT INTO membership.memberships (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetMembershipById :one
SELECT * FROM membership.memberships WHERE id = $1;

-- name: GetMemberships :many
SELECT * FROM membership.memberships;

-- name: UpdateMembership :execrows
UPDATE membership.memberships
SET name        = $1,
    description = $2,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $3;

-- name: DeleteMembership :execrows
DELETE FROM membership.memberships WHERE id = $1;