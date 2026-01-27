-- name: InitiatePendingEmailChange :execrows
UPDATE users.users
SET pending_email = $2,
    pending_email_token = $3,
    pending_email_token_expires_at = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByPendingEmailToken :one
SELECT id, email, pending_email, first_name, pending_email_token_expires_at
FROM users.users
WHERE pending_email_token = $1 AND deleted_at IS NULL;

-- name: CompletePendingEmailChange :execrows
UPDATE users.users
SET email = pending_email,
    pending_email = NULL,
    pending_email_token = NULL,
    pending_email_token_expires_at = NULL,
    email_changed_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND pending_email IS NOT NULL AND deleted_at IS NULL;

-- name: CancelPendingEmailChange :execrows
UPDATE users.users
SET pending_email = NULL,
    pending_email_token = NULL,
    pending_email_token_expires_at = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users.users
    WHERE email = $1 AND deleted_at IS NULL
) AS exists;

-- name: GetUserPendingEmailInfo :one
SELECT id, email, pending_email, first_name, pending_email_token, pending_email_token_expires_at
FROM users.users
WHERE id = $1 AND deleted_at IS NULL;
