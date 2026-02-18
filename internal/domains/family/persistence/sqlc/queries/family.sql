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
  AND verified_at IS NULL
  AND completed_at IS NULL
  AND cancelled_at IS NULL
  AND expires_at > NOW()
RETURNING *;

-- name: MarkOldParentVerified :one
UPDATE users.parent_link_requests
SET old_parent_verified_at = NOW()
WHERE id = $1
  AND old_parent_verified_at IS NULL
  AND completed_at IS NULL
  AND cancelled_at IS NULL
  AND expires_at > NOW()
RETURNING *;

-- name: CompleteRequest :one
UPDATE users.parent_link_requests
SET completed_at = NOW()
WHERE id = $1
  AND completed_at IS NULL
  AND cancelled_at IS NULL
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
SELECT u.id, u.first_name, u.last_name, u.email,
       COALESCE(
           (SELECT plr.completed_at FROM users.parent_link_requests plr
            WHERE plr.child_id = u.id AND plr.new_parent_id = $1 AND plr.completed_at IS NOT NULL
            ORDER BY plr.completed_at DESC LIMIT 1),
           u.updated_at
       ) as linked_at
FROM users.users u
WHERE u.parent_id = $1 AND u.deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, email, parent_id
FROM users.users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserById :one
SELECT id, first_name, last_name, email, parent_id
FROM users.users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPendingRequestsForUser :many
SELECT plr.id, plr.child_id, plr.new_parent_id, plr.old_parent_id,
       plr.initiated_by, plr.verified_at, plr.old_parent_verified_at,
       plr.expires_at, plr.completed_at, plr.cancelled_at, plr.created_at,
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
JOIN users.users c ON c.id = plr.child_id AND c.deleted_at IS NULL
JOIN users.users np ON np.id = plr.new_parent_id AND np.deleted_at IS NULL
LEFT JOIN users.users op ON op.id = plr.old_parent_id AND op.deleted_at IS NULL
WHERE (plr.child_id = $1 OR plr.new_parent_id = $1 OR plr.old_parent_id = $1)
  AND plr.completed_at IS NULL
  AND plr.cancelled_at IS NULL
  AND plr.expires_at > NOW()
ORDER BY plr.created_at DESC;
