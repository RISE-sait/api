-- name: CreateMembership :one
INSERT INTO membership.memberships (name, description, benefits)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMembershipById :one
SELECT * FROM membership.memberships WHERE id = $1;

-- name: GetMemberships :many
SELECT * FROM membership.memberships;

-- name: UpdateMembership :one
UPDATE membership.memberships
SET name        = $1,
    description = $2,
    benefits = $3,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $4
RETURNING *;

-- name: DeleteMembership :execrows
DELETE FROM membership.memberships WHERE id = $1;