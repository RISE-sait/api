-- name: GetPendingChildAccountByChildEmail :one
SELECT * from pending_child_accounts WHERE user_email = $1;

-- name: CreatePendingChildAccount :one
INSERT INTO pending_child_accounts (user_email, parent_email, first_name, last_name, password ) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeletePendingChildAccount :execrows
DELETE FROM pending_child_accounts WHERE user_email = $1;