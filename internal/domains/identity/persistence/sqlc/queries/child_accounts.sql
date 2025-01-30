-- name: GetPendingChildAccountByChildEmail :one
SELECT * from pending_child_accounts WHERE user_email = $1;

-- name: CreatePendingChildAccount :execrows
INSERT INTO pending_child_accounts (user_email, parent_email, password ) VALUES ($1, $2, $3); 

-- name: DeletePendingChildAccount :execrows
DELETE FROM pending_child_accounts WHERE user_email = $1;