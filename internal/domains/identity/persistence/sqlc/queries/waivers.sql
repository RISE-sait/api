-- name: CreateWaiverSignedStatus :execrows
INSERT INTO waiver_signing (user_id, waiver_id, is_signed) 
VALUES ($1, $2, $3);

-- name: GetWaiver :one
SELECT * FROM waiver WHERE waiver_url = $1 LIMIT 1;

-- name: GetPendingChildAccountWaiverSigning :many
SELECT *, w.waiver_url FROM pending_accounts_waiver_signing ws 
JOIN waiver w ON w.id = ws.waiver_id
WHERE user_id = (SELECT id from pending_child_accounts WHERE user_email = $1);

-- name: CreatePendingChildAccountWaiverSigning :execrows
INSERT INTO pending_accounts_waiver_signing (user_id, waiver_id, is_signed) 
VALUES ($1, $2, $3);

-- name: DeletePendingChildAccountWaiverSigning :execrows
DELETE FROM pending_accounts_waiver_signing WHERE user_id = (SELECT id from pending_child_accounts WHERE user_email = $1);