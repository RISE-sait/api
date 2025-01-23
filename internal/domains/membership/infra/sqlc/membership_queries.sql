-- name: CreateMembership :execrows
INSERT INTO memberships (name, description, start_date, end_date)
VALUES ($1, $2, $3, $4);

-- name: GetMembershipById :one
SELECT * FROM memberships WHERE id = $1;

-- name: GetAllMemberships :many
SELECT * FROM memberships;

-- name: UpdateMembership :execrows
UPDATE memberships
SET name = $1, description = $2, start_date = $3, end_date = $4
WHERE id = $5;

-- name: DeleteMembership :execrows
DELETE FROM memberships WHERE id = $1;

-- name: IsMembershipIDExist :one
SELECT EXISTS (SELECT 1 FROM memberships WHERE id = $1) AS exists;