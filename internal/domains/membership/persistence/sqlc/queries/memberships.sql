-- name: CreateMembership :execrows
INSERT INTO membership.memberships (name, description)
VALUES ($1, $2);

-- name: GetMembershipById :one
SELECT * FROM membership.memberships WHERE id = $1;

-- name: GetAllMemberships :many
SELECT * FROM membership.memberships;

-- name: UpdateMembership :execrows
UPDATE membership.memberships
SET name = $1, description = $2
WHERE id = $3;

-- name: DeleteMembership :execrows
DELETE FROM membership.memberships WHERE id = $1;