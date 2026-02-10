-- name: CreateLinkRequest :one
INSERT INTO users.parent_link_requests (
    child_id, new_parent_id, old_parent_id,
    initiated_by, verification_code, old_parent_code, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPendingRequestByChildId :one
SELECT * FROM users.parent_link_requests
WHERE child_id = $1
  AND completed_at IS NULL
  AND cancelled_at IS NULL
  AND expires_at > NOW();

-- name: GetRequestByVerificationCode :one
SELECT * FROM users.parent_link_requests
WHERE verification_code = $1
  AND completed_at IS NULL
  AND cancelled_at IS NULL
  AND expires_at > NOW();

-- name: GetRequestByOldParentCode :one
SELECT * FROM users.parent_link_requests
WHERE old_parent_code = $1
  AND completed_at IS NULL
  AND cancelled_at IS NULL
  AND expires_at > NOW();

-- name: GetRequestById :one
SELECT * FROM users.parent_link_requests
WHERE id = $1;

-- name: MarkVerified :one
UPDATE users.parent_link_requests
SET verified_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkOldParentVerified :one
UPDATE users.parent_link_requests
SET old_parent_verified_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CompleteRequest :one
UPDATE users.parent_link_requests
SET completed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelRequest :exec
UPDATE users.parent_link_requests
SET cancelled_at = NOW()
WHERE id = $1
  AND completed_at IS NULL
  AND cancelled_at IS NULL;

-- name: UpdateChildParent :exec
UPDATE users.users
SET parent_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: RemoveChildParent :exec
UPDATE users.users
SET parent_id = NULL, updated_at = NOW()
WHERE id = $1;

-- name: GetChildrenByParentId :many
SELECT id, first_name, last_name, email, created_at
FROM users.users
WHERE parent_id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email, parent_id
FROM users.users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserById :one
SELECT id, first_name, last_name, email, parent_id
FROM users.users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPendingRequestsForUser :many
SELECT plr.*,
       c.first_name as child_first_name,
       c.last_name as child_last_name,
       c.email as child_email,
       np.first_name as new_parent_first_name,
       np.last_name as new_parent_last_name,
       np.email as new_parent_email,
       op.first_name as old_parent_first_name,
       op.last_name as old_parent_last_name,
       op.email as old_parent_email
FROM users.parent_link_requests plr
JOIN users.users c ON c.id = plr.child_id
JOIN users.users np ON np.id = plr.new_parent_id
LEFT JOIN users.users op ON op.id = plr.old_parent_id
WHERE (plr.child_id = $1 OR plr.new_parent_id = $1 OR plr.old_parent_id = $1)
  AND plr.completed_at IS NULL
  AND plr.cancelled_at IS NULL
  AND plr.expires_at > NOW()
ORDER BY plr.created_at DESC;

-- name: IsUserChild :one
SELECT parent_id IS NOT NULL as is_child
FROM users.users
WHERE id = $1 AND deleted_at IS NULL;
