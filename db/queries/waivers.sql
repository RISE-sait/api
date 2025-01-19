-- name: GetWaiverByEmailAndDocLink :one
SELECT *
FROM waivers
WHERE email = $1
AND document_link = $2;

-- name: UpdateWaiverSignedStatusByEmail :execrows
UPDATE waivers
SET is_signed = $2, updated_at = CURRENT_TIMESTAMP
WHERE email = $1;

-- name: GetAllUniqueWaiverDocs :many
SELECT DISTINCT ON (document_link) *
FROM waivers;