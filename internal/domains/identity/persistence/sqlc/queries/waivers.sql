-- name: CreateWaiverSignedStatus :execrows
INSERT INTO waiver_signing (user_id, waiver_id, is_signed) 
VALUES ($1, $2, $3);

-- name: GetWaiver :one
SELECT * FROM waiver WHERE waiver_url = $1 LIMIT 1;
