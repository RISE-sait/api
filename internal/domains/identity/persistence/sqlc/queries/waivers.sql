-- name: CreateWaiverSignedStatus :execrows
WITH prepared_data as (SELECT unnest(@user_id_array::uuid[])    as user_id,
                              unnest(@waiver_url_array::text[]) as waiver_url,
                              unnest(@is_signed_array::bool[])  as is_signed)
INSERT INTO waiver.waiver_signing (user_id, waiver_id, is_signed)
SELECT p.user_id, w.id, p.is_signed
FROM prepared_data p
         LEFT JOIN waiver.waiver w ON w.waiver_url = p.waiver_url;

-- name: GetRequiredWaivers :many
SELECT *
FROM waiver.waiver;

-- name: CreateWaiverUpload :one
INSERT INTO waiver.waiver_uploads (user_id, file_url, file_name, file_type, file_size_bytes, uploaded_by, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetWaiverUploadsByUserId :many
SELECT wu.*,
       u.first_name as uploader_first_name,
       u.last_name as uploader_last_name
FROM waiver.waiver_uploads wu
LEFT JOIN users.users u ON wu.uploaded_by = u.id
WHERE wu.user_id = $1
ORDER BY wu.created_at DESC;

-- name: GetWaiverUploadById :one
SELECT * FROM waiver.waiver_uploads WHERE id = $1;

-- name: DeleteWaiverUpload :execrows
DELETE FROM waiver.waiver_uploads WHERE id = $1;